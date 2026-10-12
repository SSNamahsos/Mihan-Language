package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type client struct {
	t    *testing.T
	conn net.Conn
	msgs chan *Message
	mu   sync.Mutex
	id   int

	asked []string
}

func newClient(t *testing.T) (*Server, *client) {
	t.Helper()
	mine, theirs := net.Pipe()
	conn := NewConn(theirs, theirs)
	srv := NewServer(conn, "test")
	go func() { _ = conn.Serve() }()
	c := &client{t: t, conn: mine, msgs: make(chan *Message, 256), id: 1}
	go c.read()
	t.Cleanup(func() { mine.Close(); theirs.Close() })
	c.initialize()
	return srv, c
}

func (c *client) read() {
	defer close(c.msgs)
	br := bufio.NewReader(c.conn)
	for {
		m, err := readFramed(br)
		if err != nil {
			return
		}
		c.msgs <- m
	}
}

func readFramed(br *bufio.Reader) (*Message, error) {
	length := -1
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := splitHeader(line)
		if ok && name == "Content-Length" {
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
			length = n
		}
	}
	if length < 0 {
		return nil, fmt.Errorf("a message arrived with no length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(br, body); err != nil {
		return nil, err
	}
	var m Message
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *client) initialize() {
	res := c.ask("initialize", &InitializeParams{ProcessID: 1})
	var out InitializeResult
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the reply to initialize could not be read: %v", err)
	}
	if out.ServerInfo.Name != "mihan-ls" {
		c.t.Errorf("the server calls itself %q, which is not what an editor should be told", out.ServerInfo.Name)
	}

	caps := out.Capabilities
	switch {
	case caps.TextDocumentSync != 1:
		c.t.Errorf("the server asks for sync %d, so a change would not carry the whole file", caps.TextDocumentSync)
	case !caps.HoverProvider:
		c.t.Error("the server does not offer hover, so an editor would not ask")
	case !caps.DefinitionProvider:
		c.t.Error("the server does not offer go-to-definition")
	case !caps.DocumentSymbolProvider:
		c.t.Error("the server does not offer an outline")
	case !caps.DocumentFormattingProvider:
		c.t.Error("the server does not offer formatting")
	case caps.CompletionProvider == nil:
		c.t.Error("the server does not offer completion")
	}
	c.notify("initialized", struct{}{})
}

func (c *client) ask(method string, params any) *Message {
	c.t.Helper()
	c.mu.Lock()
	c.asked = append(c.asked, method)
	id := c.id
	c.id++
	c.mu.Unlock()

	body, err := json.Marshal(params)
	if err != nil {
		c.t.Fatal(err)
	}
	msg := Message{JSONRPC: "2.0", ID: json.RawMessage(strconv.Itoa(id)), Method: method, Params: body}
	if err := c.write(&msg); err != nil {
		c.t.Fatalf("cannot ask %s: %v", method, err)
	}
	want := strconv.Itoa(id)
	deadline := time.After(5 * time.Second)
	for {
		select {
		case m, ok := <-c.msgs:
			if !ok {
				c.t.Fatalf("the server closed the conversation while answering %s", method)
			}
			if m.Method == "" && string(m.ID) == want {
				if m.Error != nil {
					c.t.Fatalf("asking %s failed: %s", method, m.Error.Message)
				}
				return m
			}
		case <-deadline:
			c.t.Fatalf("the server never answered %s", method)
			return nil
		}
	}
}

func (c *client) notify(method string, params any) {
	c.t.Helper()
	body, err := json.Marshal(params)
	if err != nil {
		c.t.Fatal(err)
	}
	if err := c.write(&Message{JSONRPC: "2.0", Method: method, Params: body}); err != nil {
		c.t.Fatalf("cannot send %s: %v", method, err)
	}
}

func (c *client) write(m *Message) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.conn, "Content-Length: %d\r\n\r\n", len(b)); err != nil {
		return err
	}
	_, err = c.conn.Write(b)
	return err
}

func (c *client) open(text string) []Diagnostic {
	c.t.Helper()

	c.notify("textDocument/didOpen", &DidOpenParams{
		TextDocument: TextDocumentItem{URI: "file:///t.mhn", LanguageID: "mihan", Version: 1, Text: text},
	})
	m := c.await("textDocument/publishDiagnostics")
	var out PublishDiagnosticsParams
	if err := json.Unmarshal(m.Params, &out); err != nil {
		c.t.Fatalf("the mistakes could not be read: %v", err)
	}
	return out.Diagnostics
}

func (c *client) await(method string) *Message {
	c.t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case m, ok := <-c.msgs:
			if !ok {
				c.t.Fatalf("the server closed the conversation while sending %s", method)
			}
			if m.Method == method {
				return m
			}
		case <-deadline:
			c.t.Fatalf("the server never sent %s", method)
			return nil
		}
	}
}

func (c *client) complete(line, char int) *CompletionList {
	c.t.Helper()
	res := c.ask("textDocument/completion", &CompletionParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///t.mhn"},
			Position:     Position{Line: line, Character: char},
		},
	})
	var out CompletionList
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the suggestions could not be read: %v", err)
	}
	return &out
}

func (c *client) hover(line, char int) *Hover {
	c.t.Helper()
	res := c.ask("textDocument/hover", &HoverParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///t.mhn"},
			Position:     Position{Line: line, Character: char},
		},
	})
	if string(res.Result) == "null" {
		return nil
	}
	var out Hover
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the note could not be read: %v", err)
	}
	return &out
}

func (c *client) definition(line, char int) []Location {
	c.t.Helper()
	res := c.ask("textDocument/definition", &HoverParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///t.mhn"},
			Position:     Position{Line: line, Character: char},
		},
	})
	if string(res.Result) == "null" {
		return nil
	}
	var out []Location
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the jump could not be read: %v", err)
	}
	return out
}

func (c *client) symbols() []SymbolInformation {
	c.t.Helper()
	res := c.ask("textDocument/documentSymbol", &TextDocumentParams{
		TextDocument: TextDocumentIdentifier{URI: "file:///t.mhn"},
	})
	if string(res.Result) == "null" {
		return nil
	}
	var out []SymbolInformation
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the outline could not be read: %v", err)
	}
	return out
}

func (c *client) format(text string) []TextEdit {
	c.t.Helper()
	res := c.ask("textDocument/formatting", &FormattingParams{
		TextDocument: TextDocumentItem{URI: "file:///t.mhn", Text: text},
		Options:      FormattingOptions{TabSize: 4, InsertSpaces: true},
	})
	if string(res.Result) == "null" {
		return nil
	}
	var out []TextEdit
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the tidying could not be read: %v", err)
	}
	return out
}

func labels(list *CompletionList) map[string]CompletionItem {
	out := map[string]CompletionItem{}
	for _, it := range list.Items {
		out[it.Label] = it
	}
	return out
}
