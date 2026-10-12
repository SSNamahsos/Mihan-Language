package ir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

func TestEveryProgramInTheRepositoryLowers(t *testing.T) {
	root := repoRoot(t)
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
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("cannot read %s: %v", path, err)
		}
		bag := &diag.Bag{}
		f := parser.Parse(path, src, bag)
		if bag.HasErrors() {

			continue
		}
		p := Lower(f, bag)
		if p == nil || p.Top == nil {
			t.Errorf("%s lowered to nothing", path)
			continue
		}
		reportProblems(t, path, p)
	}
}

func reportProblems(t *testing.T, path string, p *Program) {
	t.Helper()
	consts, names, funcs, classes := len(p.Consts), len(p.Names), len(p.Funcs), len(p.Classes)
	for _, f := range p.Funcs {
		for pc, in := range f.Code {
			if !in.Op.valid() {
				t.Errorf("%s: %s instruction %d has an operation that is not real (%d)",
					path, f.Name, pc, int(in.Op))
				continue
			}

			if in.Line < 1 {
				t.Errorf("%s: %s instruction %d (%s) has no source line", path, f.Name, pc, in.Op)
			}
			bad := func(what string, v int32, n int) {
				if v < 0 || int(v) >= n {
					t.Errorf("%s: %s instruction %d (%s) %s %d, which is outside %d",
						path, f.Name, pc, in.Op, what, v, n)
				}
			}
			switch in.Op {
			case OpConst:
				bad("names constant", in.A, consts)
			case OpName, OpGetField, OpSetField, OpIsType, OpCallBuiltin:
				bad("names", in.A, names)
			case OpLocalGet, OpLocalSet, OpLocalDefine:
				bad("names slot", in.A, f.Locals+1)
			case OpUpvalueGet, OpUpvalueSet, OpUpvalueDefine:
				if in.A < 0 || int(in.A) >= f.Upvals+1 {
					t.Errorf("%s: %s instruction %d (%s) names captured name %d, but the function captures %d",
						path, f.Name, pc, in.Op, in.A, f.Upvals)
				}
			case OpDefineFunc, OpPushHandler:
				bad("names function", in.B, funcs)
			case OpDefineClass, OpDefineStruct, OpNewClass:
				bad("names class", in.B, classes)
			case OpJump, OpJumpIfFalse, OpJumpIfTrue, OpJumpIfNotNil:
				if in.A < 0 || int(in.A) > len(f.Code) {
					t.Errorf("%s: %s instruction %d (%s) goes to %d, outside its %d instructions",
						path, f.Name, pc, in.Op, in.A, len(f.Code))
				}
			}
		}
	}

	for i, c := range p.Classes {
		if c.FieldStart < 0 || int(c.FieldStart)+int(c.FieldCount) > names {
			t.Errorf("%s: class %d names fields %d..%d but there are %d names",
				path, i, c.FieldStart, int(c.FieldStart)+int(c.FieldCount), names)
		}
		if c.CtorFn < -1 || int(c.CtorFn) >= funcs {
			t.Errorf("%s: class %d names a constructor %d but there are %d functions",
				path, i, c.CtorFn, funcs)
		}
	}
}

func TestPoolsAreSelfConsistent(t *testing.T) {
	p := lower(t, "darghan N\n    x = 0\n    __sakht__(x)\n        khod.x = x\n    funct kard()\n        chap_khat khod.x\nshenase Rang\n    zaban = 0\n    taj = 1\nn = N(1)\nn.kard()\nchap_khat Rang.zaban\n")
	if len(p.Names) == 0 {
		t.Fatal("names were resolved away entirely, which cannot be right")
	}
	for i, n := range p.Names {
		if n == "" && i == 0 {
			t.Errorf("name %d is empty, which is a hole in the pool", i)
		}
	}
	if len(p.Nums) == 0 {
		t.Error("an enum must have written its members down")
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot find where I am: %v", err)
	}
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("cannot find the repository root")
	return ""
}
