package eval

import (
	"reflect"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

func resolveFile(f *ast.Block) {
	if f == nil {
		return
	}
	top := map[string]bool{}
	nestedImport := false
	for _, st := range f.Stmts {
		walkOutsideFuncs(reflect.ValueOf(st), top, &nestedImport, true)
	}
	if nestedImport {
		return
	}
	markFuncs(reflect.ValueOf(f), top)
}

func walkOutsideFuncs(v reflect.Value, bound map[string]bool, nestedImport *bool, topStmt bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return
		}
		switch n := v.Interface().(type) {
		case *ast.FuncDecl:
			if !topStmt {
				bound[n.Name] = true
			}
			return
		case *ast.Lambda:
			return
		case *ast.ImportDecl:
			if !topStmt {
				*nestedImport = true
			}
		case *ast.VarDecl:
			if !topStmt || n.Sabit || n.TypeName != "" {
				bound[n.Name] = true
			}
		case *ast.ClassDecl, *ast.StructDecl, *ast.EnumDecl, *ast.InterfaceDecl:
			if !topStmt {
				collectNames(v, bound)
			}

			return
		}

		walkOutsideFuncs(v.Elem(), bound, nestedImport, false)
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			f := v.Field(i)
			name := t.Field(i).Name
			if f.Kind() == reflect.String && (name == "VarName" || (name == "Name" && t.Name() == "Catch")) {
				bound[f.String()] = true
				continue
			}
			walkOutsideFuncs(f, bound, nestedImport, false)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			walkOutsideFuncs(v.Index(i), bound, nestedImport, false)
		}
	}
}

func collectNames(v reflect.Value, out map[string]bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return
		}
		if _, ok := v.Interface().(*ast.Ident); ok {

			return
		}
		if a, ok := v.Interface().(*ast.Assign); ok {
			if id, ok := a.Target.(*ast.Ident); ok {
				out[id.Name] = true
			}
		}
		if _, ok := v.Interface().(*ast.ImportDecl); ok {
			out["\x00import"] = true
		}
		collectNames(v.Elem(), out)
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			if !t.Field(i).IsExported() {
				continue
			}
			f := v.Field(i)
			n := t.Field(i).Name
			if f.Kind() == reflect.String && (n == "Name" || n == "VarName" || n == "Alias") {
				out[f.String()] = true
				continue
			}
			collectNames(f, out)
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			collectNames(v.Index(i), out)
		}
	}
}

func markFuncs(v reflect.Value, outer map[string]bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return
		}
		switch v.Interface().(type) {
		case *ast.FuncDecl, *ast.Lambda:
			local := make(map[string]bool, len(outer)+8)
			for k := range outer {
				local[k] = true
			}
			if fd, ok := v.Interface().(*ast.FuncDecl); ok {

				collectNames(reflect.ValueOf(fd.Params), local)
				collectNames(reflect.ValueOf(fd.Body), local)
			} else {
				collectNames(v, local)
			}
			if local["\x00import"] {
				return
			}
			markReads(v.Elem(), local)
			return
		}
		markFuncs(v.Elem(), outer)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				markFuncs(v.Field(i), outer)
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			markFuncs(v.Index(i), outer)
		}
	}
}

func markReads(v reflect.Value, bound map[string]bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return
		}
		switch n := v.Interface().(type) {
		case *ast.Ident:
			if !bound[n.Name] && n.Name != "khod" {
				n.FileScope = true
			}
			return
		case *ast.FuncDecl, *ast.Lambda:
			markFuncs(v, bound)
			return
		}
		markReads(v.Elem(), bound)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).IsExported() {
				markReads(v.Field(i), bound)
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			markReads(v.Index(i), bound)
		}
	}
}

func lookupIdent(x *ast.Ident, env *rt.Env) *rt.Binding {
	if !x.FileScope {
		b, _ := env.Lookup(x.Name)
		return b
	}
	root := env.Root()
	if s := x.LoadSlot(); s != nil && s.Scope == any(root) {
		return s.Binding.(*rt.Binding)
	}
	b, _ := root.Lookup(x.Name)
	if b != nil {
		x.StoreSlot(&ast.Slot{Scope: root, Binding: b})
	}
	return b
}
