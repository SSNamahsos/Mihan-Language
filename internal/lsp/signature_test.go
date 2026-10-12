package lsp

import (
	"encoding/json"
	"strings"
	"testing"
)

func (c *client) signature(line, char int) *SignatureHelp {
	c.t.Helper()
	res := c.ask("textDocument/signatureHelp", &HoverParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: "file:///t.mhn"},
			Position:     Position{Line: line, Character: char},
		},
	})
	var out *SignatureHelp
	if err := json.Unmarshal(res.Result, &out); err != nil {
		c.t.Fatalf("the signature could not be read: %v", err)
	}
	return out
}

func TestSignatureHelpForAFunctInTheFile(t *testing.T) {
	_, c := newClient(t)
	c.open("funct jam(a, b = 10)\n    bargard a + b\nchap_khat jam(1, \n")
	h := c.signature(2, 17)
	if h == nil || len(h.Signatures) != 1 {
		t.Fatalf("expected one signature, got %+v", h)
	}
	if h.Signatures[0].Label != "funct jam(a, b = 10)" {
		t.Errorf("label is %q", h.Signatures[0].Label)
	}
	if h.ActiveParameter != 1 {
		t.Errorf("active parameter is %d, want 1", h.ActiveParameter)
	}
}

func TestSignatureHelpForABuiltin(t *testing.T) {
	_, c := newClient(t)
	c.open("x = taghib(\n")
	h := c.signature(0, 11)
	if h == nil || len(h.Signatures) != 1 || h.Signatures[0].Documentation == nil ||
		!strings.Contains(h.Signatures[0].Documentation.Value, "taghib") {
		t.Fatalf("expected taghib's help, got %+v", h)
	}
}

func TestCompletionShowsAFunctsParameters(t *testing.T) {
	_, c := newClient(t)
	c.open("funct salam_kon(naam = \"doust\")\n    chap_khat naam\n\n")
	list := c.complete(2, 0)
	for _, it := range list.Items {
		if it.Label == "salam_kon" {
			if it.Detail != `funct salam_kon(naam = "doust")` {
				t.Errorf("detail is %q", it.Detail)
			}
			return
		}
	}
	t.Fatal("salam_kon was not offered")
}
