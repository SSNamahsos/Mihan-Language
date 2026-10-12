package eval

import (
	"reflect"

	"mihan/internal/ast"
)

func canShareScope(b *ast.Block) bool {
	if b == nil {
		return true
	}
	return !makesAClosure(b)
}

func makesAClosure(n ast.Node) bool {
	found := false
	walkNodes(reflect.ValueOf(n), func(v reflect.Value) bool {
		if found {
			return false
		}

		switch v.Type() {
		case reflect.TypeOf(diagSpan{}):
			return false
		}
		if !v.IsValid() {
			return false
		}
		switch v.Interface().(type) {
		case *ast.Lambda, *ast.FuncDecl, *ast.ClassDecl:
			found = true
			return false
		}
		return true
	})
	return found
}

type diagSpan struct{}

func walkNodes(v reflect.Value, visit func(reflect.Value) bool) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return
		}
		if v.Kind() == reflect.Interface {

			if _, ok := v.Interface().(ast.Node); ok {
				walkNodes(v.Elem(), visit)
			}
			return
		}
		if _, ok := v.Interface().(ast.Node); ok {
			if !visit(v) {
				return
			}
		}
		walkNodes(v.Elem(), visit)
		return
	case reflect.Struct:
		if _, ok := v.Interface().(ast.Node); !ok && v.Type() == reflect.TypeOf(diagSpan{}) {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.Name == "Loc" || !f.IsExported() {
				continue
			}
			walkNodes(v.Field(i), visit)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			walkNodes(v.Index(i), visit)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			walkNodes(v.MapIndex(k), visit)
		}
	}
}
