package mihan

import (
	"bytes"
	"testing"

	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/eval"
	"mihan/internal/parser"
)

var programs = map[string]string{

	"fib": `funct fib(n)
    agar n < 2
        bargard n
    bargard fib(n - 1) + fib(n - 2)
chap_khat fib(22)`,

	"loop": `
total = 0
har i az 0 ta 3_000_000
    total = total + i * 2
chap_khat total`,

	"list": `
xs = []
har i az 0 ta 200_000
    bast xs, i * i
shomar = 0
har x dar xs
    shomar = shomar + x
chap_khat shomar`,

	"text": `
matn = "abcdefghij" * 400
n = 0
har i az 0 ta 400_000
    agar matn[i % 4000] == "a"
        n = n + 1
chap_khat n`,

	"class": `
darghan Shomarande
    adad = 0
    __sakht__(a)
        khod.adad = a
    ziad_kon(n)
        khod.adad = khod.adad + n
s = Shomarande(0)
har i az 0 ta 1_000_000
    s.ziad_kon(i)
chap_khat s.adad`,

	"closure": `
dahan = funct(x) -> x * 2
jam = 0
har i az 0 ta 1_000_000
    jam = jam + dahan(i)
chap_khat jam`,
}

func compile(tb testing.TB, src string) *ast.File {
	tb.Helper()
	bag := &diag.Bag{}
	f := parser.Parse("bench.mhn", []byte(src), bag)
	if bag.HasErrors() {
		tb.Fatalf("the benchmark program did not read: %s", bag.Sorted()[0].Message())
	}
	return f
}

func BenchmarkMihan(b *testing.B) {
	for _, name := range sortedKeys(programs) {
		src := programs[name]
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				f := compile(b, src)
				var out bytes.Buffer
				in := eval.New(&out, nil)
				b.StartTimer()
				if err := in.RunFile(f); err != nil {
					b.Fatalf("the program stopped: %v", err)
				}
			}
		})
	}
}

func BenchmarkMihanCompile(b *testing.B) {
	for _, name := range sortedKeys(programs) {
		src := programs[name]
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				bag := &diag.Bag{}
				parser.Parse("bench.mhn", []byte(src), bag)
				if bag.HasErrors() {
					b.Fatal("did not read")
				}
			}
		})
	}
}

func BenchmarkGoDirect(b *testing.B) {
	b.Run("fib", func(b *testing.B) {
		b.ReportAllocs()
		var goFib func(int) int64
		goFib = func(n int) int64 {
			if n < 2 {
				return int64(n)
			}
			return goFib(n-1) + goFib(n-2)
		}
		for i := 0; i < b.N; i++ {
			_ = goFib(22)
		}
	})
	b.Run("loop", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			var total int64
			for j := 0; j < 3_000_000; j++ {
				total += int64(j) * 2
			}
			_ = total
		}
	})
}

func TestBenchmarkProgramsRun(t *testing.T) {
	for _, name := range sortedKeys(programs) {
		t.Run(name, func(t *testing.T) {
			f := compile(t, programs[name])
			var out bytes.Buffer
			in := eval.New(&out, nil)
			if err := in.RunFile(f); err != nil {
				t.Fatalf("the benchmark program stopped: %s", in.RenderError(err, nil))
			}
			if out.Len() == 0 {
				t.Fatal("the benchmark program printed nothing, so it is not measuring anything")
			}
		})
	}
}

func TestProgramsArePortable(t *testing.T) {
	for name, src := range programs {
		if bytes.Contains([]byte(src), []byte("khanak")) {
			t.Errorf("the %q benchmark asks for input, so its number would not be comparable", name)
		}
	}
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
