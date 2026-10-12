package vm

import (
	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

type classRuntime struct {
	class   *rt.Class
	methods map[string]int32
}

func (m *Machine) defineClass(nameIdx, classIdx int32, isStruct bool) (Value, *fault) {
	if int(classIdx) < 0 || int(classIdx) >= len(m.p.Classes) {
		return Nil(), &fault{code: "E4004",
			fa: "کلاسی که این دستور می‌سازد پیدا نشد",
			en: "the class named here was not found"}
	}
	ic := m.p.Classes[classIdx]
	c := &rt.Class{
		Name:     ic.Name,
		Doc:      ic.Doc,
		Methods:  map[string]*ast.FuncDecl{},
		Defaults: map[string]rt.Value{},
	}

	for _, base := range ic.Bases {
		b, ok := m.classByName(base)
		if !ok {
			return Nil(), &fault{code: "E4008",
				fa: "پایه‌ی «" + base + "» پیش از این ساخته نشده",
				en: base + " must be declared before it is used as a base"}
		}
		c.Parent = b
		break
	}

	for i := 0; i < int(ic.FieldCount); i++ {
		fieldName := m.nameText(ic.FieldStart + int32(i))
		var v rt.Value = rt.Nil
		if i < len(ic.Fields) {
			f := ic.Fields[i]
			switch {
			case f.Default >= 0:
				v = m.constant(f.Default).ToRT()
			case f.TypeName != "":
				v = zeroOfType(f.TypeName)
			}
		}
		c.Order = append(c.Order, fieldName)
		c.Fields = append(c.Fields, &ast.FieldDecl{Name: fieldName})
		c.Defaults[rt.HashKeyMust(fieldName)] = v
	}
	cr := &classRuntime{class: c, methods: map[string]int32{}}
	for _, method := range ic.Methods {
		cr.methods[method.Name] = method.Fn
	}
	m.classes[c.Name] = c
	m.classInfo[c] = cr
	return ClassValue(&rt.ClassValue{Class: c}), nil
}

func (m *Machine) classByName(name string) (*rt.Class, bool) {
	if m.classes == nil {
		return nil, false
	}
	c, ok := m.classes[name]
	return c, ok
}

func (m *Machine) defineEnum(nameIdx, run int32) (Value, *fault) {
	c := &rt.Class{
		Name:       m.nameText(nameIdx),
		Methods:    map[string]*ast.FuncDecl{},
		Defaults:   map[string]rt.Value{},
		Enumerated: true,
	}

	at := int(run)
	if at < 0 || at >= len(m.p.Nums) {
		return Nil(), &fault{code: "E4005",
			fa: "اعضای این shenase پیدا نشد",
			en: "the members of this shenase were not found"}
	}
	n := int(m.p.Nums[at])
	at++
	for i := 0; i < n && at+1 < len(m.p.Nums); i++ {
		member := m.nameText(m.p.Nums[at])
		val := m.constant(m.p.Nums[at+1]).ToRT()
		at += 2
		c.Order = append(c.Order, member)
		c.Fields = append(c.Fields, &ast.FieldDecl{Name: member, TypeName: c.Name})
		c.Defaults[rt.HashKeyMust(member)] = val
	}
	m.classes[c.Name] = c
	m.classInfo[c] = &classRuntime{class: c, methods: map[string]int32{}}
	return ClassValue(&rt.ClassValue{Class: c}), nil
}

func (m *Machine) defineInterface(_, _ int32) (Value, *fault) {
	return Nil(), nil
}

func zeroOfType(typeName string) rt.Value {
	base := typeName
	for i := 0; i < len(base); i++ {
		if base[i] == '<' {
			base = base[:i]
			break
		}
	}
	switch base {
	case "arz":
		return rt.Int(0)
	case "kasr":
		return rt.Float(0)
	case "matn", "harf":
		return rt.Str("")
	case "bool":
		return rt.Bool(false)
	case "araye", "list":
		return rt.NewList()
	case "dict":
		return rt.NewDict()
	case "set":
		return rt.NewSet()
	}
	return rt.Nil
}

func (m *Machine) construct(c *rt.Class, args []Value) (Value, bool, *fault, error) {
	if c.Enumerated {
		return Object(rt.NewObject(c)), false, nil, nil
	}
	cr := m.classInfo[c]
	if cr != nil {

		if _, f, found := m.methodOf(c, "__sakht__"); found {
			if f != nil {
				return Nil(), false, f, nil
			}
			fn, _, _ := m.methodOf(c, "__sakht__")
			obj := rt.NewObject(c)
			args = append([]Value{Object(obj)}, args...)
			if _, f, err := m.enterFunc(fn, args); f != nil || err != nil {
				return Nil(), false, f, err
			}

			m.f.ctor = obj
			return Nil(), true, nil, nil
		}
	}
	if len(args) > 0 {
		return Nil(), false, &fault{code: "E4157",
			fa: "«" + c.Name + "» ساخت‌دهنده ندارد، پس این مقدارها جایی نداشتند",
			en: c.Name + " has no constructor, so these values had nowhere to go"}, nil
	}
	return Object(rt.NewObject(c)), false, nil, nil
}

func (m *Machine) methodOf(c *rt.Class, name string) (*funcVal, *fault, bool) {
	for k := c; k != nil; k = k.Parent {
		if cr := m.classInfo[k]; cr != nil {
			if fnIdx, ok := cr.methods[name]; ok && fnIdx >= 0 && int(fnIdx) < len(m.fns) {
				fn := m.fns[fnIdx]
				fv := &funcVal{ir: fn, name: k.Name + "." + name}
				if len(fn.UpvalNames) > 0 {
					fv.upvals = make([]*Value, len(fn.UpvalNames))
					for i, n := range fn.UpvalNames {
						cell := m.capture(n)
						if cell == nil {
							return nil, &fault{code: "E4003",
								fa: "این متد نامی را می‌خواهد که بیرون از آن پیدا نمی‌شود: " + m.nameText(n),
								en: "this method wants a name that is not in sight: " + m.nameText(n)}, true
						}
						fv.upvals[i] = cell
					}
				}
				return fv, nil, true
			}
		}
	}
	return nil, nil, false
}

func (m *Machine) getField(recv Value, nameIdx int32) (Value, *fault) {
	n := m.nameText(nameIdx)
	switch t := recv.p.(type) {
	case *rt.Object:
		if v, ok := t.Get(n); ok {
			return FromRT(v), nil
		}
		return Nil(), &fault{code: "E4160",
			fa: "«" + n + "» در «" + t.Class.Name + "» نیست",
			en: n + " is not a field or a method of " + t.Class.Name}
	case *rt.ClassValue:
		if t.Class.Enumerated {
			if v, ok := t.Class.Defaults[rt.HashKeyMust(n)]; ok {
				return FromRT(v), nil
			}
		}
		return Nil(), &fault{code: "E4161",
			fa: "«" + n + "» روی «" + t.Class.Name + "» نیست",
			en: n + " is not a field or a method of " + t.Class.Name}
	case *rt.ErrorValue:

		switch n {
		case "message":
			return FromRT(t.Message), nil
		case "data":
			return FromRT(t.Data), nil
		case "name":
			return Str(t.Name), nil
		}
		return Nil(), &fault{code: "E4162",
			fa: "روی khata چیزی به نام «" + n + "» نیست",
			en: "a khata has no " + n + "; it carries message, data and name"}
	case *rt.List:

		if n == "tule" || n == "len" {
			return Int(int64(len(t.Items))), nil
		}
	case *rt.Dict:
		if n == "tule" || n == "len" {
			return Int(int64(t.Len())), nil
		}
	case *rt.Set:
		if n == "tule" || n == "len" {
			return Int(int64(t.Len())), nil
		}
	}
	if recv.tag == TagStr && (n == "tule" || n == "len") {
		return Int(int64(rt.RuneCount(recv.s))), nil
	}
	if it, ok := recv.p.(*iter); ok && (n == "tule" || n == "len") {
		return length(Iter(it)), nil
	}
	return Nil(), &fault{code: "E4164",
		fa: "از " + recv.TypeName() + " نمی‌شود عضو خواند",
		en: "you cannot read a field from a " + recv.TypeName() + "; only a darghan, a sakht or a khata has fields"}
}

func (m *Machine) setField(recv Value, nameIdx int32, val Value) *fault {
	n := m.nameText(nameIdx)
	switch t := recv.p.(type) {
	case *rt.Object:
		t.Set(n, val.ToRT())
		return nil
	case *rt.ClassValue:
		if t.Class.Enumerated {
			t.Class.Defaults[rt.HashKeyMust(n)] = val.ToRT()
			return nil
		}
	}
	return &fault{code: "E4161",
		fa: "روی این مقدار نمی‌شود «" + n + "» را نوشت",
		en: n + " cannot be written on this value"}
}

func (m *Machine) storeDefined(name int32, v Value) {
	if f := m.f; f != nil {
		if _, isLocal := f.localSlot(name); isLocal {
			m.push(v)
			return
		}
	}
	m.setGlobal(name, v)
}

func (m *Machine) initClasses() {
	if m.classes == nil {
		m.classes = map[string]*rt.Class{}
		m.classInfo = map[*rt.Class]*classRuntime{}
		m.structs = map[string]bool{}
	}
}
