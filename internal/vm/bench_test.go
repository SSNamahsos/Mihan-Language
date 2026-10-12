package vm

import (
	"bytes"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/ir"
	"mihan/internal/parser"
)

const benchLoop = `
n = 0
har i az 0 ta 20_000_000
    n = n + i * 2 - 1
chap_khat n
`

func lower(tb testing.TB, src string) *ir.Program {
	tb.Helper()
	bag := &diag.Bag{}
	p := ir.Lower(parser.Parse("b.mhn", []byte(src), bag), bag)
	if p == nil {
		tb.Fatalf("the program did not lower: %s", bag.Sorted()[0].EN)
	}
	return p
}

func BenchmarkMihanLoopOnTheMachine(b *testing.B) {
	p := lower(b, benchLoop)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := New(p, &bytes.Buffer{})
		if err := m.Run(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMihanLoopOnTheInterpreter(b *testing.B) {
	p := lower(b, benchLoop)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bag := &diag.Bag{}
		f := parser.Parse("b.mhn", []byte(benchLoop), bag)
		_ = p
		if err := eval.New(&bytes.Buffer{}, nil).RunFile(f); err != nil {
			b.Fatal(err)
		}
	}
}

const benchSum = `
n = 0
har i az 0 ta 2_000_000
    n += i
chap_khat n
`

func BenchmarkMihanSumOnTheMachine(b *testing.B) {
	p := lower(b, benchSum)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := New(p, &bytes.Buffer{})
		if err := m.Run(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMihanSumOnTheInterpreter(b *testing.B) {
	p := lower(b, benchSum)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bag := &diag.Bag{}
		f := parser.Parse("b.mhn", []byte(benchSum), bag)
		_ = p
		if err := eval.New(&bytes.Buffer{}, nil).RunFile(f); err != nil {
			b.Fatal(err)
		}
	}
}
