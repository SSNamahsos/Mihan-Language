package eval

import (
	"os"
	"path/filepath"
	"strings"

	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/parser"
	rt "mihan/internal/runtime"
)

type importState struct {
	stack  []string
	loaded map[string]*module
}

type module struct {
	names map[string]rt.Value

	order []string
}

func (in *Interp) child(file string) *Interp {
	c := *in
	c.File = file
	c.Globals = rt.NewEnv(nil)
	c.Trace = nil
	c.imps = in.imps
	return &c
}

func (in *Interp) execImport(s *ast.ImportDecl, env *rt.Env) error {

	if _, found := env.Lookup(s.Path); found != nil {
		return nil
	}
	if s.Alias != "" {
		if _, found := env.Lookup(s.Alias); found != nil {
			return nil
		}
	}

	path, err := in.findModule(s)
	if err != nil {
		return err
	}
	mod, err := in.loadModule(path, s)
	if err != nil {
		return err
	}

	if s.Alias == "" {

		for name, v := range mod.names {
			if b, _ := env.Lookup(name); b == nil {
				env.Declare(name, v)
			}
		}
		return nil
	}

	ns := rt.NewObject(&rt.Class{
		Name:     s.Alias,
		Fields:   moduleFields(mod),
		Order:    moduleOrder(mod),
		Methods:  map[string]*ast.FuncDecl{},
		Defaults: map[string]rt.Value{},
	})
	for name, v := range mod.names {
		ns.Fields[name] = v
	}
	env.Declare(s.Alias, ns)
	return nil
}

func moduleOrder(m *module) []string {
	return m.order
}

func moduleFields(m *module) []*ast.FieldDecl {
	out := make([]*ast.FieldDecl, 0, len(m.order))
	for _, name := range m.order {
		out = append(out, &ast.FieldDecl{Name: name})
	}
	return out
}

func (in *Interp) findModule(s *ast.ImportDecl) (string, error) {
	if filepath.IsAbs(s.Path) {
		if _, err := os.Stat(s.Path); err != nil {
			return "", in.moduleMissing(s)
		}
		return filepath.Clean(s.Path), nil
	}
	dir := filepath.Dir(in.File)
	candidates := []string{
		filepath.Join(dir, s.Path),
		s.Path,
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, s.Path))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return filepath.Clean(c), nil
		}
	}
	return "", in.moduleMissing(s)
}

func (in *Interp) moduleMissing(s *ast.ImportDecl) error {
	dir := filepath.Dir(in.File)
	return in.faultFA(s, "E4009",
		"parvande-i be esm-e '%s' peyda nashod; dar %s va kenar-e khod-e barname gashtam",
		[]any{s.Path, dir},
		"module %s was not found; looked beside the program and in the working directory", s.Path)
}

func (in *Interp) loadModule(path string, s *ast.ImportDecl) (*module, error) {
	clean := filepath.Clean(path)

	if i := indexOf(in.imps.stack, clean); i >= 0 {
		chain := make([]string, 0, len(in.imps.stack)-i+1)
		chain = append(chain, in.imps.stack[i:]...)
		chain = append(chain, clean)
		return nil, in.faultFA(s, "E4010",
			"in parvande-ha dar yek dayere hamdige ro seda mizanan va hich vaght tamoom nemishe: %s",
			[]any{strings.Join(chain, " ← ")},
			"these files import each other in a circle, which would never finish: %s",
			strings.Join(chain, " -> "))
	}
	if mod, ok := in.imps.loaded[clean]; ok {
		return mod, nil
	}

	src, err := os.ReadFile(clean)
	if err != nil {
		return nil, in.moduleMissing(s)
	}
	bag := &diag.Bag{}
	file := parser.Parse(clean, src, bag)
	if bag.HasErrors() {
		d := bag.Sorted()[0]
		return nil, in.faultFA(s, "E2001",
			"parvande-ye %s khoonde nashod: %s",
			[]any{clean, d.FA},
			"module %s did not read: %s", clean, d.EN)
	}

	in.imps.stack = append(in.imps.stack, clean)
	child := in.child(clean)

	err = child.RunTopLevel(file)
	in.imps.stack = in.imps.stack[:len(in.imps.stack)-1]
	if err != nil {
		return nil, err
	}

	mod := &module{names: map[string]rt.Value{}}
	for _, name := range child.Globals.Names() {
		if b, _ := child.Globals.Lookup(name); b != nil && b.Used {
			mod.names[name] = b.Value
			mod.order = append(mod.order, name)
		}
	}
	in.imps.loaded[clean] = mod
	return mod, nil
}

func indexOf(ss []string, s string) int {
	for i, x := range ss {
		if x == s {
			return i
		}
	}
	return -1
}

func lastSegment(p string) string {
	base := filepath.Base(p)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
