package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/driver"
	"mihan/internal/eval"
	"mihan/internal/format"
	"mihan/internal/ir"
	"mihan/internal/parser"
)

func cmdRun(args []string) error {

	for len(args) > 0 && args[0] == "--open" {
		os.Setenv("MIHAN_OPEN_BROWSER", "1")
		args = args[1:]
	}
	if len(args) > 1 && args[len(args)-1] == "--open" {
		os.Setenv("MIHAN_OPEN_BROWSER", "1")
		args = args[:len(args)-1]
	}
	path := "main.mhn"
	if len(args) > 0 {
		path = args[0]
	}
	switch driver.ClassifyPath(path) {
	case driver.KindMissing:
		return fatal("%s is not here", path)
	case driver.KindDir:
		files, err := driver.SourceFiles(path)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return fatal("%s holds no .%s file, and no main.%s either", path, driver.Extension, driver.Extension)
		}
		path = files[0]
	}

	rest := args
	if len(rest) > 0 {
		rest = rest[1:]
	}
	res, src, err := driver.RunFileOnDisk(path, driver.Config{
		Out: os.Stdout, In: os.Stdin, Args: rest,
	})
	if err != nil {
		return err
	}
	if res.Failed {
		printDiagnostics(res.Diagnostics, map[string][]byte{path: src})
		return errSilent
	}
	if code := res.ExitCode; code != 0 {

		os.Exit(code)
	}
	return nil
}

var errSilent = &silentError{}

type silentError struct{}

func (e *silentError) Error() string { return "" }

func cmdCheck(args []string) error {
	if len(args) == 0 {
		return fatal("give me a file to check, like `mihan check salam.mhn`")
	}
	path := args[0]
	src, err := os.ReadFile(path)
	if err != nil {
		return fatal("%s cannot be read: %v", path, err)
	}
	bag := &diag.Bag{}
	file := parser.Parse(path, src, bag)
	if !bag.HasErrors() {

		ir.Lower(file, bag)
	}
	printDiagnostics(bag.Sorted(), map[string][]byte{path: src})
	if bag.HasErrors() {
		return errSilent
	}
	fmt.Printf("salam! %s is ready.\n", filepath.Base(path))
	return nil
}

func printDiagnostics(ds []*diag.Diagnostic, src map[string][]byte) {
	for _, d := range ds {
		fmt.Fprint(os.Stderr, d.Render(src[d.File]))
		fmt.Fprintln(os.Stderr)
	}
}

func cmdNew(args []string) error {
	if len(args) == 0 {
		return fatal("what should I call the new program?")
	}
	name := args[0]
	dir := name
	if filepath.Ext(name) == "."+driver.Extension {
		dir = strings.TrimSuffix(name, "."+driver.Extension)
	}
	if _, err := os.Stat(dir); err == nil {
		return fatal("%s is already there", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	main := fmt.Sprintf(`# %s

barname %s

chap $"salam, donya! man {Name} hastam."

Name = "%s"
`, dir, titleCase(dir), titleCase(dir))
	if err := os.WriteFile(filepath.Join(dir, "main."+driver.Extension), []byte(main), 0o644); err != nil {
		return err
	}
	fmt.Printf("sakhtam: %s\n", dir)
	fmt.Printf("alan benevis: mihan run %s\n", dir)
	return nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func cmdFmt(args []string) error {
	if len(args) == 0 {
		return fatal("give me a file to tidy, like `mihan fmt salam.mhn`")
	}
	var failed bool
	for _, path := range args {
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		bag := &diag.Bag{}
		file := parser.Parse(path, src, bag)
		if bag.HasErrors() {
			fmt.Fprintf(os.Stderr, "mihan fmt: %s has mistakes, so I left it alone\n", path)
			for _, d := range bag.Sorted() {
				fmt.Fprint(os.Stderr, d.Render(src))
				fmt.Fprintln(os.Stderr)
			}
			failed = true
			continue
		}
		out := format.File(file, src)
		if out == string(src) {
			continue
		}
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			return err
		}
		fmt.Println(path)
	}
	if failed {
		return errSilent
	}
	return nil
}

func cmdRepl(args []string) error {
	fmt.Println(replBanner())
	r := eval.New(os.Stdout, os.Stdin)

	r.MaxSteps = 0
	var pending strings.Builder
	for {
		prompt := "mihan> "
		if pending.Len() > 0 {
			prompt = "  ...> "
		}
		fmt.Print(prompt)

		line, err := r.ReadLine()
		if err != nil && line == "" {
			fmt.Println()
			return nil
		}
		switch strings.TrimSpace(strings.ToLower(line)) {
		case "exit", "quit", "khareh", "bafandeh":

			fmt.Println("khodahafzar! ")
			return nil
		}
		if pending.Len() == 0 {
			switch strings.TrimSpace(strings.ToLower(line)) {
			case "help", "rahnama":
				fmt.Println(replHelp())
				continue
			case "clear", "paki":
				fmt.Print("\033[H\033[2J")
				continue
			}
		}
		pending.WriteString(line)
		pending.WriteString("\n")
		if !looksComplete(pending.String()) {
			continue
		}
		src := pending.String()
		pending.Reset()
		runREPLLine(r, src)
	}
}

func echoLine(src string) string {
	bag := &diag.Bag{}
	file := parser.Parse("<repl>", []byte(src), bag)
	if bag.HasErrors() || len(file.Decls) == 0 {
		return src
	}
	last := file.Decls[len(file.Decls)-1]
	es, ok := last.(*ast.ExprStmt)
	if !ok {
		return src
	}

	switch es.X.(type) {
	case *ast.Call, *ast.Assign, *ast.ErrorExpr:
		return src
	}
	return src + "chap_khat neshan(" + format.Expr(es.X) + ")\n"
}

func runREPLLine(r *eval.Interp, src string) {
	bag := &diag.Bag{}
	src = echoLine(src)
	file := parser.Parse("<repl>", []byte(src), bag)
	if bag.HasErrors() {
		for _, d := range bag.Sorted() {
			fmt.Fprint(os.Stderr, d.Render([]byte(src)))
			fmt.Fprintln(os.Stderr)
		}
		return
	}
	if err := r.RunTopLevel(file); err != nil {
		if _, ok := eval.ExitCode(err); ok {
			return
		}
		fmt.Fprint(os.Stderr, r.RenderError(err, []byte(src)))
	}
}
