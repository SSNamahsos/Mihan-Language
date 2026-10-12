package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTheBuiltProgramAnswersAnEditor(t *testing.T) {
	if testing.Short() {
		t.Skip("this builds the compiler, which is slow")
	}
	bin := compilerForTest(t)
	cmd := exec.Command(bin, "lsp")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	})

	c := newProcClient(t, stdin, stdout)

	res := c.askMethod("initialize", map[string]any{
		"processId":    1,
		"capabilities": map[string]any{},
	})
	caps, _ := res["capabilities"].(map[string]any)
	if caps == nil {
		t.Fatalf("the program did not say what it can do, it said %v", res)
	}
	if caps["hoverProvider"] != true {
		t.Errorf("the program says it cannot hover, so an editor would never ask:\n%v", caps)
	}
	c.notify("initialized", map[string]any{})

	c.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri": "file:///t.mhn", "languageId": "mihan", "version": 1,
			"text": "x = \n",
		},
	})
	c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": "file:///t.mhn"},
		"contentChanges": []any{map[string]any{"text": "x = \n"}},
	})
	d := c.awaitMethod("textDocument/publishDiagnostics")
	if d == nil {
		t.Fatal("a file that assigns nothing was never reported, so the editor would show no squiggle")
	}
	diags, _ := d["diagnostics"].([]any)
	if len(diags) == 0 {
		t.Fatalf("the notice about the file carried no mistakes:\n%v", d)
	}

	c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": "file:///t.mhn"},
		"contentChanges": []any{map[string]any{"text": "chap_khat 1\n"}},
	})
	c.awaitMethod("textDocument/publishDiagnostics")
	list := c.askMethod("textDocument/completion", map[string]any{
		"textDocument": map[string]any{"uri": "file:///t.mhn"},
		"position":     map[string]any{"line": 1, "character": 0},
	})
	items, _ := list["items"].([]any)
	if len(items) == 0 {
		t.Errorf("the program offered nothing at all, which is not a useful editor:\n%v", list)
	}
	found := false
	for _, it := range items {
		m, _ := it.(map[string]any)
		if m["label"] == "chap" {
			found = true
		}
	}
	if !found {
		t.Error("the program did not offer chap, which is the first word anyone writes")
	}
}

type procClient struct {
	t    *testing.T
	w    io.Writer
	r    *bufio.Reader
	id   int
	msgs chan map[string]any
	done chan struct{}
}

func newProcClient(t *testing.T, w io.Writer, r io.Reader) *procClient {
	c := &procClient{
		t: t, w: w, r: bufio.NewReader(r),
		msgs: make(chan map[string]any, 64),
		done: make(chan struct{}),
	}
	go c.read()
	return c
}

func (c *procClient) read() {
	defer close(c.msgs)
	for {
		length := -1
		for {
			line, err := c.r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			if name, value, ok := splitColon(line); ok && strings.EqualFold(name, "Content-Length") {
				n, err := strconv.Atoi(value)
				if err == nil {
					length = n
				}
			}
		}
		if length < 0 {
			return
		}
		body := make([]byte, length)
		if _, err := io.ReadFull(c.r, body); err != nil {
			return
		}
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			return
		}
		c.msgs <- m
	}
}

func (c *procClient) awaitMethod(method string) map[string]any {
	c.t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		select {
		case m, ok := <-c.msgs:
			if !ok {
				c.t.Fatalf("the program closed its side while %s was expected", method)
			}
			if m["method"] == method {
				params, _ := m["params"].(map[string]any)
				return params
			}
		case <-deadline:
			c.t.Fatalf("the program never sent %s", method)
			return nil
		}
	}
}

func (c *procClient) awaitID(id int) map[string]any {
	c.t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		select {
		case m, ok := <-c.msgs:
			if !ok {
				c.t.Fatalf("the program closed its side while the answer to %d was expected", id)
			}
			if got, _ := m["id"].(float64); int(got) == id {
				res, _ := m["result"].(map[string]any)
				return res
			}
		case <-deadline:
			c.t.Fatalf("the program never answered the question numbered %d", id)
			return nil
		}
	}
}

func (c *procClient) askMethod(method string, params any) map[string]any {
	c.t.Helper()
	c.id++
	id := c.id
	c.send(id, method, params)
	return c.awaitID(id)
}

func (c *procClient) notify(method string, params any) {
	c.t.Helper()
	c.send(nil, method, params)
}

func (c *procClient) send(id any, method string, params any) {
	c.t.Helper()
	msg := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	if id != nil {
		msg["id"] = id
	}
	body, err := json.Marshal(msg)
	if err != nil {
		c.t.Fatal(err)
	}
	if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		c.t.Fatalf("cannot write to the program: %v", err)
	}
	if _, err := c.w.Write(body); err != nil {
		c.t.Fatalf("cannot write to the program: %v", err)
	}
}

func splitColon(s string) (string, string, bool) {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			v := strings.TrimLeft(s[i+1:], " ")
			return s[:i], v, true
		}
	}
	return "", "", false
}
