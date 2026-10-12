package vm

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/ir"
	"mihan/internal/parser"
)

func TestEveryDocumentedProgramEitherAgreesOrRefusesByName(t *testing.T) {
	root := repoRootForVM(t)
	var files []string
	for _, dir := range []string{"docs", "tests"} {
		base := filepath.Join(root, dir)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".mhn") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("cannot read %s: %v", base, err)
		}
	}
	if len(files) == 0 {
		t.Skip("no .mhn files were found, so this test is checking nothing")
	}

	var agreed, refused int
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			bag := &diag.Bag{}
			file := parser.Parse(path, src, bag)
			if bag.HasErrors() {
				t.Skip("a file that does not read is the parser's business")
			}
			var want bytes.Buffer
			in := eval.New(&want, nil)
			if err := in.RunFile(file); err != nil {
				t.Skipf("the interpreter stopped, which is that test's business: %v", err)
			}
			lbag := &diag.Bag{}
			p := ir.Lower(parser.Parse(path, src, lbag), lbag)
			if p == nil || p.Top == nil {
				t.Fatal("the program did not lower")
			}
			var got bytes.Buffer
			err = New(p, &got).Run()
			if err != nil {

				msg := err.Error()
				if !namesAnOperation(msg) {
					t.Errorf("the machine refused without naming what it cannot do: %v", err)
				}
				refused++
				return
			}
			if got.String() != want.String() {
				t.Errorf("the machine disagreed with the interpreter.\n  want %q\n  got  %q",
					want.String(), got.String())
			}
			agreed++
		})
	}
	t.Logf("%d programs agree, %d are refused by name, of %d", agreed, refused, len(files))
}

func namesAnOperation(msg string) bool {
	return strings.Contains(msg, "this machine does not do")
}

func repoRootForVM(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("cannot find the repository root")
	return ""
}
