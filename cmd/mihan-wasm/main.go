//go:build js && wasm

package main

import (
	"bytes"
	"io"
	"strings"
	"syscall/js"

	"mihan/internal/driver"
)

type promptReader struct {
	buf bytes.Buffer
	out *bytes.Buffer
}

func (p *promptReader) Read(b []byte) (int, error) {
	if p.buf.Len() == 0 {
		v := js.Global().Call("mihanInput", p.out.String())
		if v.IsNull() || v.IsUndefined() {
			return 0, io.EOF
		}
		p.buf.WriteString(v.String() + "\n")
	}
	return p.buf.Read(b)
}

func run(this js.Value, args []js.Value) any {
	src := args[0].String()
	var argv []string
	if len(args) > 1 && args[1].String() != "" {
		argv = strings.Split(args[1].String(), ",")
	}
	handler := js.FuncOf(func(_ js.Value, pa []js.Value) any {
		resolve := pa[0]
		go func() { resolve.Invoke(js.ValueOf(execute(src, argv))) }()
		return nil
	})
	return js.Global().Get("Promise").New(handler)
}

func execute(src string, argv []string) map[string]any {
	var out bytes.Buffer
	in := &promptReader{out: &out}
	res, err := driver.Run("barname.mhn", []byte(src), driver.Config{Out: &out, In: in, Args: argv, MaxSteps: 50000000, Name: "barname.mhn"})
	errText := ""
	if res != nil && len(res.Diagnostics) > 0 {
		errText = driver.Report(res.Diagnostics, map[string][]byte{"barname.mhn": []byte(src)})
	}
	if err != nil && errText == "" {
		errText = err.Error()
	}
	return map[string]any{"out": out.String(), "err": strings.TrimSpace(errText)}
}

func main() {
	js.Global().Set("mihanRun", js.FuncOf(run))
	select {}
}
