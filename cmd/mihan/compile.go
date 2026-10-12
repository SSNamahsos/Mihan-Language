package main

import (
	"fmt"
	"os"

	"mihan/internal/diag"
	"mihan/internal/driver"
	"mihan/internal/ir"
	"mihan/internal/parser"
)

func cmdCompile(args []string) error {
	pools := false
	var path string
	for _, a := range args {
		switch a {
		case "--pools":
			pools = true
		case "--ir":

		default:
			path = a
		}
	}
	if path == "" {
		return fatal("give me a file to compile, like `mihan compile salam.mhn`")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return fatal("%s cannot be read: %v", path, err)
	}
	bag := &diag.Bag{}
	file := parser.Parse(path, src, bag)
	if bag.HasErrors() {
		printDiagnostics(bag.Sorted(), map[string][]byte{path: src})
		return errSilent
	}
	prog := ir.Lower(file, bag)
	if bag.HasErrors() {
		printDiagnostics(bag.Sorted(), map[string][]byte{path: src})
		return errSilent
	}
	if pools {
		printPools(prog)
		return nil
	}
	fmt.Print(ir.Disassemble(prog))
	return nil
}

func printPools(p *ir.Program) {
	fmt.Printf("constants (%d)\n", len(p.Consts))
	for i, c := range p.Consts {
		switch c.Kind {
		case ir.ConstInt:
			fmt.Printf("  %4d  adad   %d\n", i, c.I)
		case ir.ConstRune:
			fmt.Printf("  %4d  harf   %q\n", i, c.S)
		case ir.ConstFloat:
			fmt.Printf("  %4d  kesr   %g\n", i, c.F)
		case ir.ConstStr:
			fmt.Printf("  %4d  matn   %q\n", i, c.S)
		case ir.ConstBool:
			fmt.Printf("  %4d  dorost %v\n", i, c.B)
		case ir.ConstHich:
			fmt.Printf("  %4d  hich\n", i)
		}
	}
	fmt.Printf("\nnames (%d)\n", len(p.Names))
	for i, n := range p.Names {
		fmt.Printf("  %4d  %s\n", i, n)
	}
	if len(p.Nums) > 0 {
		fmt.Printf("\nnumbers (%d)\n", len(p.Nums))
		for i, n := range p.Nums {
			fmt.Printf("  %4d  %d\n", i, n)
		}
	}
	fmt.Printf("\nclasses (%d)\n", len(p.Classes))
	for i, c := range p.Classes {
		kind := "darghan"
		if c.IsStruct {
			kind = "sakht"
		}
		fmt.Printf("  %4d  %s %s, %d field(s), %d method(s)\n", i, kind, c.Name, c.FieldCount, len(c.Methods))
	}
}

var _ = driver.Extension
