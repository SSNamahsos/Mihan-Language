package lsp

import "encoding/json"

type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type TextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidOpenParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeParams struct {
	TextDocument   TextDocumentIdentifier `json:"textDocument"`
	ContentChanges []TextChange           `json:"contentChanges"`
}

type TextChange struct {
	Range *Range `json:"range,omitempty"`
	Text  string `json:"text"`
}

type DidCloseParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type CompletionParams struct {
	TextDocumentPositionParams
}

type HoverParams struct {
	TextDocumentPositionParams
}

type FormattingParams struct {
	TextDocument TextDocumentItem  `json:"textDocument"`
	Options      FormattingOptions `json:"options"`
}

type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"`
	Code     string `json:"code,omitempty"`
	Source   string `json:"source"`
	Message  string `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type InitializeParams struct {
	ProcessID int    `json:"processId"`
	RootURI   string `json:"rootUri,omitempty"`

	Capabilities map[string]json.RawMessage `json:"capabilities"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`

	ServerInfo ServerInfo `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerCapabilities struct {
	TextDocumentSync           int                   `json:"textDocumentSync"`
	CompletionProvider         *CompletionOptions    `json:"completionProvider,omitempty"`
	HoverProvider              bool                  `json:"hoverProvider"`
	DefinitionProvider         bool                  `json:"definitionProvider"`
	DocumentSymbolProvider     bool                  `json:"documentSymbolProvider"`
	DocumentFormattingProvider bool                  `json:"documentFormattingProvider"`
	SignatureHelpProvider      *SignatureHelpOptions `json:"signatureHelpProvider,omitempty"`
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
	ResolveProvider   bool     `json:"resolveProvider"`
}

type CompletionItemKind int

const (
	KindText     CompletionItemKind = 1
	KindMethod   CompletionItemKind = 2
	KindFunction CompletionItemKind = 3
	KindField    CompletionItemKind = 5
	KindVariable CompletionItemKind = 6
	KindClass    CompletionItemKind = 7
	KindKeyword  CompletionItemKind = 14
	KindSnippet  CompletionItemKind = 15
	KindConstant CompletionItemKind = 21
)

const (
	InsertPlain   = 1
	InsertSnippet = 2
)

type CompletionItem struct {
	Label string `json:"label"`

	Kind CompletionItemKind `json:"kind"`

	Detail string `json:"detail,omitempty"`

	Documentation *MarkupContent `json:"documentation,omitempty"`

	InsertText string `json:"insertText,omitempty"`

	InsertTextFormat int `json:"insertTextFormat,omitempty"`

	FilterText string `json:"filterText,omitempty"`

	SortText string `json:"sortText,omitempty"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

const Markup = "markdown"

type Hover struct {
	Contents MarkupContent `json:"contents"`

	Range *Range `json:"range,omitempty"`
}

type SymbolKind int

const (
	SymbolFile      SymbolKind = 1
	SymbolClass     SymbolKind = 5
	SymbolMethod    SymbolKind = 6
	SymbolFunction  SymbolKind = 12
	SymbolEnum      SymbolKind = 10
	SymbolInterface SymbolKind = 11
)

type SymbolInformation struct {
	Name     string     `json:"name"`
	Kind     SymbolKind `json:"kind"`
	Location Location   `json:"location"`

	ContainerName string `json:"containerName,omitempty"`
}
