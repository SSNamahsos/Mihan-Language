package mihan

import (
	"bytes"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/parser"
)

func TestWhereTheTimeGoes(t *testing.T) {
	if os.Getenv("MIHAN_PROFILE") == "" {
		t.Skip("set MIHAN_PROFILE=1 to write loop.pprof")
	}
	src := `
n = 0
har i az 0 ta 300_000
    n = n + i * 2
chap_khat n`
	bag := &diag.Bag{}
	f := parser.Parse("p.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatal(bag.Sorted()[0].Message())
	}
	var out bytes.Buffer
	in := eval.New(&out, nil)
	if err := in.RunFile(f); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	runtime.GC()
	file, err := os.Create("loop.pprof")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := pprof.Lookup("allocs").WriteTo(file, 0); err != nil {
		t.Fatal(err)
	}
	t.Log("wrote loop.pprof")
}
