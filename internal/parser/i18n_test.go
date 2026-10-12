package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestEveryExpectPhraseHasAPersianHalf(t *testing.T) {
	used := expectPhrases(t)
	if len(used) == 0 {
		t.Fatal("found no call to expect, so this test is checking nothing")
	}
	for _, phrase := range used {
		if _, ok := farsiPhrases[phrase]; !ok {
			t.Errorf("expect(%q) has no Persian half; add it to farsiPhrases", phrase)
		}
	}

	for phrase := range farsiPhrases {
		if !contains(used, phrase) {
			t.Errorf("farsiPhrases has %q, which no call site passes to expect", phrase)
		}
	}
}

func expectPhrases(t *testing.T) []string {
	t.Helper()
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", nil, 0)
	if err != nil {
		t.Fatalf("cannot read the package: %v", err)
	}
	var used []string
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "expect" || len(call.Args) < 2 {
					return true
				}
				lit, ok := call.Args[1].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				phrase, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatalf("%s: cannot read %s: %v", fset.Position(lit.Pos()), lit.Value, err)
				}
				used = append(used, phrase)
				return true
			})
		}
	}
	return used
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestEveryTokenKindHasAPersianDescription(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "token", "token.go"))
	if err != nil {
		t.Fatalf("cannot read token.go: %v", err)
	}
	src := string(data)
	for _, kind := range []struct{ name, method string }{
		{"EOF", "return \"پایان پرونده\""},
		{"NEWLINE", "return \"پایان سطر\""},
		{"INDENT", "return \"بلوکِ تورفتگی\""},
		{"DEDENT", "return \"پایانِ بلوکِ تورفتگی\""},
	} {
		if !contains2(src, kind.method) {
			t.Errorf("DescribeFarsi has no Persian word for %s", kind.name)
		}
	}
}

func contains2(haystack, needle string) bool {
	return len(needle) > 0 && len(haystack) >= len(needle) &&
		func() bool {
			for i := 0; i+len(needle) <= len(haystack); i++ {
				if haystack[i:i+len(needle)] == needle {
					return true
				}
			}
			return false
		}()
}
