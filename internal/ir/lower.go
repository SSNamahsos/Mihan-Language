package ir

import (
	"mihan/internal/ast"
	"mihan/internal/diag"
)

type lowerer struct {
	p *Program

	code []Instr

	frame *frameInfo

	funcIdx   map[ast.Stmt]int32
	clsIdx    map[ast.Stmt]int32
	enumIdx   map[ast.Stmt]int32
	iflIdx    map[ast.Stmt]int32
	lambdaIdx map[*ast.Lambda]int32

	bag *diag.Bag

	pending []pending

	loops []*loopJumps

	depth int
}

func (l *lowerer) begin(f *Func, fr *frameInfo) func() {
	oldCode, oldFrame := l.code, l.frame
	l.code, l.frame = f.Code, fr
	return func() {
		f.Code, f.Locals = l.code, int(fr.slots)
		f.Upvals = len(fr.upvals)

		f.UpvalNames = make([]int32, 0, len(fr.upvals))
		for slot := 0; slot < len(fr.upvals); slot++ {
			for name, s := range fr.upvals {
				if int(s) == slot {
					f.UpvalNames = append(f.UpvalNames, l.nameOf(name))
					break
				}
			}
		}

		f.LocalNames = make([]int32, 0, fr.slots)
		for slot := int32(0); slot < fr.slots; slot++ {
			var found int32 = -1
			for name, s := range fr.locals {
				if s == slot {
					found = l.nameOf(name)
					break
				}
			}
			f.LocalNames = append(f.LocalNames, found)
		}
		l.code, l.frame = oldCode, oldFrame
	}
}

type frameInfo struct {
	locals map[string]int32

	upvals map[string]int32

	parent *frameInfo

	class *Class

	kind FuncKind

	slots int32
}

func Lower(f *ast.File, bag *diag.Bag) *Program {
	p := NewProgram(f.Path)
	p.ImportPaths = importPaths(f)
	if f.Module != "" {
		p.ModulePath = f.Module
	}
	if f.Program != "" {
		p.ProgramName = f.Program
	}
	l := &lowerer{
		p:       p,
		bag:     bag,
		funcIdx: map[ast.Stmt]int32{},
		clsIdx:  map[ast.Stmt]int32{},
		enumIdx: map[ast.Stmt]int32{},
		iflIdx:  map[ast.Stmt]int32{},
		frame:   &frameInfo{locals: map[string]int32{}, upvals: map[string]int32{}, kind: FuncTop},
	}

	for _, d := range f.Decls {
		l.build(d)
	}

	l.decls(f.Decls)
	l.emit(Instr{Op: OpReturnNil, Line: endLine(f)})
	p.Top.Code = l.code
	p.Top.Locals = int(l.frame.slots)
	p.Top.Upvals = len(l.frame.upvals)
	return p
}

func importPaths(f *ast.File) []string {
	out := make([]string, 0, len(f.Imports))
	for _, im := range f.Imports {
		out = append(out, im.Path)
	}
	return out
}

func endLine(f *ast.File) int32 {
	if len(f.Decls) == 0 {
		return 1
	}
	return int32(f.Decls[len(f.Decls)-1].Span().End.Line)
}

func (l *lowerer) emit(in Instr) int32 {
	pc := int32(len(l.code))
	l.code = append(l.code, in)
	return pc
}

func (l *lowerer) here() int32 { return int32(len(l.code)) }

func (l *lowerer) patch(at int32, target int32) {
	l.code[at].A = target
}

func (l *lowerer) line(n ast.Node) int32 { return int32(n.Span().Start.Line) }

func (l *lowerer) constOf(n ast.Node) int32 {
	switch x := n.(type) {
	case *ast.IntLit:
		return l.p.AddConst(Const{Kind: ConstInt, I: x.Value})
	case *ast.FloatLit:
		return l.p.AddConst(Const{Kind: ConstFloat, F: x.Value})
	case *ast.StrLit:
		return l.p.AddConst(Const{Kind: ConstStr, S: x.Value})
	case *ast.CharLit:
		return l.p.AddConst(Const{Kind: ConstRune, I: int64(x.Value)})
	case *ast.BoolLit:
		return l.p.AddConst(Const{Kind: ConstBool, B: x.Value})
	case *ast.HichLit:
		return l.p.AddConst(Const{Kind: ConstHich})
	}
	return l.p.AddConst(Const{Kind: ConstHich})
}

func (l *lowerer) nameOf(s string) int32 { return l.p.AddName(s) }

func (f *frameInfo) declareLocal(name string) int32 {
	if s, ok := f.locals[name]; ok {
		return s
	}
	s := f.slots
	f.slots++
	f.locals[name] = s
	return s
}

func (l *lowerer) resolve(name string) (Op, int32) {
	if s, ok := l.frame.locals[name]; ok {
		return OpLocalGet, s
	}
	for f := l.frame.parent; f != nil; f = f.parent {
		if _, ok := f.locals[name]; ok {
			return OpUpvalueGet, l.upvalueFor(name)
		}
	}
	return OpGlobalGet, l.nameOf(name)
}

func (l *lowerer) nameSlotOf(name string) int32 { return l.nameOf(name) }

func (l *lowerer) upvalueFor(name string) int32 {
	if s, ok := l.frame.upvals[name]; ok {
		return s
	}
	s := int32(len(l.frame.upvals))
	l.frame.upvals[name] = s
	return s
}

func (l *lowerer) load(name string, line int32) {
	op, slot := l.resolve(name)
	if op == OpUpvalueGet {

		l.emit(Instr{Op: op, A: slot, B: l.nameSlotOf(name), Line: line})
		return
	}
	l.emit(Instr{Op: op, A: slot, Line: line})
}

type pending struct {
	pc    int32
	decl  ast.Stmt
	frame *frameInfo
}

func (l *lowerer) declareNameFor(name string) bool {
	op, _ := l.resolve(name)
	if op != OpGlobalGet || l.frame.kind == FuncTop {
		return false
	}
	l.frame.declareLocal(name)
	return true
}

func (l *lowerer) flushPending() {
	for {
		i := -1
		for j, p := range l.pending {
			if p.frame == l.frame {
				i = j
				break
			}
		}
		if i < 0 {
			return
		}
		p := l.pending[i]
		l.pending = append(l.pending[:i], l.pending[i+1:]...)
		switch x := p.decl.(type) {
		case *ast.FuncDecl:
			l.funcIdx[x] = l.buildFunc(x.Name, FuncPlain, x.Params, x.Body, x.Span(), x.Doc, false)
			l.code[p.pc].B = l.funcIdx[x]
		case *ast.ClassDecl:
			l.clsIdx[x] = l.buildClass(x.Name, x.Fields, x.Methods, x.Bases, x.Span(), x.Doc, false, x.CtorName)
			l.code[p.pc].B = l.clsIdx[x]
		case *ast.StructDecl:
			l.clsIdx[x] = l.buildClass(x.Name, x.Fields, nil, x.Bases, x.Span(), x.Doc, true, "")
			l.code[p.pc].B = l.clsIdx[x]
		case *ast.EnumDecl:
			l.enumIdx[x] = l.enumRun(x)
			l.code[p.pc].B = l.enumIdx[x]
		case *ast.InterfaceDecl:
			l.iflIdx[x] = l.nameRun(x.Methods)
			l.code[p.pc].B = l.iflIdx[x]
		}
	}
}

func (l *lowerer) store(name string, line int32) {
	if l.declareNameFor(name) {
		op, slot := l.resolve(name)
		l.emit(Instr{Op: setOp(op), A: slot, Line: line})
		return
	}
	op, slot := l.resolve(name)
	if op == OpUpvalueGet {
		l.emit(Instr{Op: setOp(op), A: slot, B: l.nameSlotOf(name), Line: line})
		return
	}
	l.emit(Instr{Op: setOp(op), A: slot, Line: line})
}

func setOp(o Op) Op {
	switch o {
	case OpLocalGet:
		return OpLocalSet
	case OpUpvalueGet:
		return OpUpvalueSet
	default:
		return OpGlobalSet
	}
}

func defineOp(o Op) Op { return setOp(o) }
