package lsp

import (
	"strings"

	"mihan/internal/eval"
)

const KindValue CompletionItemKind = 12
const KindColor CompletionItemKind = 16

var colorProps = map[string]bool{"color": true, "background": true, "background-color": true, "border-color": true, "outline-color": true, "caret-color": true, "accent-color": true}

func styleContext(text string, off int) (inStr bool, isValue bool, key string) {
	if off > len(text) {
		off = len(text)
	}
	ls := strings.LastIndexByte(text[:off], '\n') + 1
	line := text[ls:off]
	q := -1
	open := false
	for i := 0; i < len(line); i++ {
		if line[i] == '\\' {
			i++
			continue
		}
		if line[i] == '"' {
			open = !open
			if open {
				q = i
			}
		}
	}
	if !open || q < 0 {
		return false, false, ""
	}
	before := strings.TrimRight(text[:ls+q], " \t\r\n")
	if before == "" {
		return false, false, ""
	}
	switch before[len(before)-1] {
	case ':':
		k := strings.TrimRight(before[:len(before)-1], " \t")
		if strings.HasSuffix(k, "\"") {
			k = k[:len(k)-1]
			if j := strings.LastIndexByte(k, '"'); j >= 0 {
				return true, true, k[j+1:]
			}
		}
		return true, true, ""
	case '{', ',':
		return true, false, ""
	}
	return false, false, ""
}

func styleItems(isValue bool, key string) []CompletionItem {
	var out []CompletionItem
	if !isValue {
		for _, w := range eval.StyleWordList() {
			out = append(out, CompletionItem{
				Label:         w[0],
				Kind:          KindProperty,
				Detail:        "sabk → " + w[1],
				Documentation: &MarkupContent{Kind: Markup, Value: "**" + w[0] + "** · CSS `" + w[1] + "`\n\n" + w[2]},
				SortText:      "0" + w[0],
			})
		}
		return out
	}
	css := ""
	if key != "" {
		for _, w := range eval.StyleWordList() {
			if w[0] == key || w[1] == strings.ReplaceAll(key, "_", "-") {
				css = w[1]
				break
			}
		}
	}
	colors := colorProps[css]
	for _, v := range eval.StyleValueList() {
		it := CompletionItem{Label: v[0], Kind: KindValue, Detail: "sabk → " + v[1],
			Documentation: &MarkupContent{Kind: Markup, Value: "**" + v[0] + "** · CSS `" + v[1] + "`"}, SortText: "1" + v[0]}
		if strings.HasSuffix(v[0], "()") {
			name := strings.TrimSuffix(v[0], "()")
			it.Label = name
			it.InsertText = name + "($1)"
			it.InsertTextFormat = 2
			it.Kind = KindFunction
		}
		if strings.HasPrefix(v[1], "#") || v[1] == "black" || v[1] == "white" || v[1] == "transparent" {
			it.Kind = KindColor
			it.Documentation = &MarkupContent{Kind: "plaintext", Value: v[1]}
			if colors {
				it.SortText = "0" + v[0]
			}
		}
		out = append(out, it)
	}
	return out
}
