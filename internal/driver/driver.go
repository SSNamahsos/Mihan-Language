package driver

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/parser"
	rt "mihan/internal/runtime"
)

type Result struct {
	Diagnostics []*diag.Diagnostic

	Failed bool

	ExitCode int

	Ran bool

	Interrupted bool
}

type Config struct {
	Out io.Writer
	In  io.Reader

	Args []string

	MaxSteps int

	Name string
}

func Run(path string, src []byte, cfg Config) (*Result, error) {
	res := &Result{ExitCode: 0}
	name := cfg.Name
	if name == "" {
		name = path
	}

	bag := &diag.Bag{}
	file := parser.Parse(name, src, bag)
	res.Diagnostics = bag.Sorted()
	if bag.HasErrors() {
		res.Failed = true
		res.ExitCode = 1
		return res, nil
	}

	in := eval.New(cfg.Out, cfg.In)
	in.File = name
	in.Args = cfg.Args
	in.MaxSteps = cfg.MaxSteps
	if err := in.RunFile(file); err != nil {
		if code, ok := eval.ExitCode(err); ok {
			res.ExitCode = code
			res.Ran = true
			return res, nil
		}
		res.Failed = true
		res.ExitCode = 1
		res.Diagnostics = append(res.Diagnostics, runtimeDiag(in, err, src)...)
		return res, nil
	}
	res.Ran = true
	return res, nil
}

func RunFileOnDisk(path string, cfg Config) (*Result, []byte, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	res, err := Run(path, src, cfg)
	return res, src, err
}

func runtimeDiag(in *eval.Interp, err error, src []byte) []*diag.Diagnostic {
	p, ok := rt.AsPanic(err)
	if !ok {
		return []*diag.Diagnostic{{
			Severity: diag.Error,
			Code:     "E5000",
			FA:       "barname ba eshkal motevaghef shod: " + err.Error(),
		}}
	}
	pos := diag.Pos{}
	if len(p.Trace) > 0 {
		pos = p.Trace[0].Pos
	}
	if pos.Line == 0 {
		pos.Line = 1
	}
	d := &diag.Diagnostic{
		Severity: diag.Error,
		Code:     "E5001",
		FA:       "barname hamin ja istad",
		Span:     diag.Span{Start: pos, End: pos},
	}

	if len(p.Trace) > 0 && p.Trace[0].File == in.File && p.Trace[0].Pos.Line > 0 {
		d.File = in.File
	}
	if ev, ok := p.Value.(*rt.ErrorValue); ok {
		d.EN = string(ev.Message)
		d.FA = string(ev.Message)
		if ev.FA != "" {
			d.FA = ev.FA
		}
		if len(ev.Name) == 5 && ev.Name[0] == 'E' {
			d.Code = ev.Name
		}
		if ev.Note != "" {
			d.Notes = append(d.Notes, ev.Note)
		}
	}

	if len(p.Trace) > 1 {
		var b strings.Builder
		b.WriteString("az koji: ")
		for i, f := range p.Trace {
			if i == 0 {
				b.WriteString(f.String())
				continue
			}
			fmt.Fprintf(&b, " → %s", f.String())
		}
		d.Notes = append(d.Notes, b.String())
	}
	_ = in
	_ = src
	return []*diag.Diagnostic{d}
}

func Report(ds []*diag.Diagnostic, src map[string][]byte) string {
	var b strings.Builder
	for _, d := range ds {
		body := src[d.File]
		b.WriteString(d.Render(body))
		b.WriteByte('\n')
	}
	return b.String()
}

type Kind int

const (
	KindFile Kind = iota

	KindDir

	KindMissing
)

func ClassifyPath(path string) Kind {
	info, err := os.Stat(path)
	if err != nil {
		return KindMissing
	}
	if info.IsDir() {
		return KindDir
	}
	return KindFile
}

func SourceFiles(dir string) ([]string, error) {
	main := filepath.Join(dir, "main.mhn")
	if _, err := os.Stat(main); err == nil {
		return []string{main}, nil
	}
	var out []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name != "." && (strings.HasPrefix(name, ".") || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == "."+Extension {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

const Extension = "mhn"

func IsMihanFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), "."+Extension)
}

func FprintDiagnostics(w io.Writer, ds []*diag.Diagnostic, src map[string][]byte) {
	for _, d := range ds {
		fmt.Fprint(w, d.Render(src[d.File]))
		fmt.Fprintln(w)
	}
}

func CountBySeverity(ds []*diag.Diagnostic) string {
	errs, warns := 0, 0
	for _, d := range ds {
		switch d.Severity {
		case diag.Error:
			errs++
		case diag.Warning:
			warns++
		}
	}
	switch {
	case errs == 0 && warns == 0:
		return "0 eshkal"
	case errs == 0:
		return fmt.Sprintf("%d hoshdar", warns)
	default:
		return fmt.Sprintf("%d eshkal", errs)
	}
}
