package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
)

type Conn struct {
	r *bufio.Reader
	w io.Writer

	mu sync.Mutex

	nextID int

	pending map[int]chan *Response

	handler Handler

	closed bool
}

type Handler interface {
	Initialize(params *InitializeParams) *InitializeResult

	DidOpen(uri string, text string)
	DidChange(uri string, text string)
	DidClose(uri string)

	Completion(uri string, pos Position) *CompletionList

	Hover(uri string, pos Position) *Hover

	SignatureHelp(uri string, pos Position) *SignatureHelp

	Definition(uri string, pos Position) []Location

	DocumentSymbol(uri string) []SymbolInformation

	Format(uri string, text string) []TextEdit

	Shutdown()
}

func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{
		r:       bufio.NewReaderSize(r, 64*1024),
		w:       w,
		nextID:  1,
		pending: map[int]chan *Response{},
	}
}

func (c *Conn) SetHandler(h Handler) { c.handler = h }

func (c *Conn) Serve() error {
	for {
		msg, err := c.read()
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			return err
		}
		c.dispatch(msg)
	}
}

type Message struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *ResponseError) Error() string { return e.Message }

func (c *Conn) read() (*Message, error) {

	length := -1
	for {
		line, err := c.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = trimLine(line)
		if line == "" {
			break
		}
		name, value, ok := splitHeader(line)
		if !ok {
			continue
		}
		if name == "Content-Length" {
			n, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("the editor sent a length I cannot read: %q", value)
			}
			length = n
		}
	}
	if length < 0 {
		return nil, fmt.Errorf("the editor sent a message with no length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(c.r, body); err != nil {
		return nil, err
	}
	var m Message
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("the editor sent something that is not a message: %w", err)
	}
	return &m, nil
}

func trimLine(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func splitHeader(s string) (name, value string, ok bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			name = s[:i]
			value = s[i+1:]
			for len(value) > 0 && value[0] == ' ' {
				value = value[1:]
			}
			return name, value, true
		}
	}
	return "", "", false
}

func (c *Conn) dispatch(m *Message) {
	if m.Method == "" {
		c.deliver(m)
		return
	}
	if c.handler == nil {

		c.replyError(m.ID, -32603, "the server is not ready yet")
		return
	}
	result, err := c.call(m)
	if m.ID == nil {
		return
	}
	if err != nil {
		c.replyError(m.ID, -32603, err.Error())
		return
	}
	c.reply(m.ID, result)
}

func (c *Conn) call(m *Message) (json.RawMessage, error) {
	h := c.handler
	switch m.Method {
	case "initialize":
		var p InitializeParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		out, err := json.Marshal(h.Initialize(&p))
		return out, err
	case "initialized":
		return nil, nil
	case "shutdown":
		h.Shutdown()
		return json.Marshal(nil)
	case "exit":
		return nil, errExit
	case "textDocument/didOpen":
		var p DidOpenParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		h.DidOpen(p.TextDocument.URI, p.TextDocument.Text)
		return nil, nil
	case "textDocument/didChange":
		var p DidChangeParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}

		if n := len(p.ContentChanges); n > 0 {
			h.DidChange(p.TextDocument.URI, p.ContentChanges[n-1].Text)
		}
		return nil, nil
	case "textDocument/didClose":
		var p DidCloseParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		h.DidClose(p.TextDocument.URI)
		return nil, nil
	case "textDocument/didSave":
		return nil, nil
	case "textDocument/completion":
		var p CompletionParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		return marshalOrNil(h.Completion(p.TextDocument.URI, p.Position))
	case "textDocument/hover":
		var p HoverParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		return marshalOrNil(h.Hover(p.TextDocument.URI, p.Position))
	case "textDocument/signatureHelp":
		var p HoverParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		return marshalOrNil(h.SignatureHelp(p.TextDocument.URI, p.Position))
	case "textDocument/definition":
		var p HoverParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		return marshalOrNil(h.Definition(p.TextDocument.URI, p.Position))
	case "textDocument/documentSymbol":
		var p TextDocumentParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		return marshalOrNil(h.DocumentSymbol(p.TextDocument.URI))
	case "textDocument/formatting":
		var p FormattingParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, err
		}
		return marshalOrNil(h.Format(p.TextDocument.URI, p.TextDocument.Text))
	}

	return json.Marshal(nil)
}

func marshalOrNil(v any) (json.RawMessage, error) {
	if v == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(v)
}

func (c *Conn) deliver(m *Message) {
	if len(m.ID) == 0 {
		return
	}
	id, err := strconv.Atoi(string(trimJSON(m.ID)))
	if err != nil {
		return
	}
	c.mu.Lock()
	ch, ok := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()
	if ok {
		ch <- &Response{JSONRPC: "2.0", ID: m.ID, Result: m.Result, Error: m.Error}
	}
}

func trimJSON(b []byte) []byte {
	b = []byte(string(b))
	for len(b) > 0 && (b[0] == ' ' || b[0] == '"') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '"') {
		b = b[:len(b)-1]
	}
	return b
}

func (c *Conn) reply(id json.RawMessage, result json.RawMessage) {
	if result == nil {
		result = json.RawMessage("null")
	}
	c.send(&Message{JSONRPC: "2.0", ID: id, Result: result})
}

func (c *Conn) replyError(id json.RawMessage, code int, message string) {
	if id == nil {
		return
	}
	c.send(&Message{JSONRPC: "2.0", ID: id, Error: &ResponseError{Code: code, Message: message}})
}

func (c *Conn) send(m *Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	body, err := json.Marshal(m)
	if err != nil {
		return
	}
	fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body))
	c.w.Write(body)
	if f, ok := c.w.(interface{ Flush() error }); ok {
		f.Flush()
	}
}

func (c *Conn) Notify(method string, params any) {
	body, err := json.Marshal(params)
	if err != nil {
		return
	}
	c.send(&Message{JSONRPC: "2.0", Method: method, Params: body})
}

var errExit = &ResponseError{Code: -1, Message: "the editor asked the server to stop"}
