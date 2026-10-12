package eval

import (
	"bytes"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

func runBoth(t *testing.T, src string) {
	t.Helper()
	res := [2]string{}
	for i := range res {
		bag := &diag.Bag{}
		file := parser.Parse("test.mhn", []byte(src), bag)
		if bag.HasErrors() {
			t.Fatalf("did not read: %s", src)
		}
		var out bytes.Buffer
		in := New(&out, nil)
		if i == 1 {
			in.MaxSteps = 1 << 60
		}
		err := in.RunFile(file)
		res[i] = out.String()
		if err != nil {
			res[i] += "\nERR " + in.RenderError(err, []byte(src))
		}
	}
	if res[0] != res[1] {
		t.Fatalf("compiled and walked differ for\n%s\n--- compiled:\n%s\n--- walked:\n%s", src, res[0], res[1])
	}
}

func TestCompiledMatchesWalker(t *testing.T) {
	progs := []string{
		"funct fib(n)\n    agar n < 2\n        bargard n\n    bargard fib(n - 1) + fib(n - 2)\nchap_khat fib(20)\n",
		"s = 0\nhar i az 0 ta 1000\n    s += i\nchap_khat s\n",
		"total = 0\nfunct add(x)\n    total = total + x\nadd(3)\nadd(4)\nchap_khat total\n",
		"funct f()\n    y = 5\n    bargard y\nchap_khat f()\nchap_khat y\n",
		"har i az 0 ta 3\n    j = i * 2\nchap_khat j\nchap_khat i\n",
		"funct big(x)\n    bargard x * x * x * x * x\nchap_khat big(100000)\n",
		"funct inf(n)\n    bargard inf(n + 1)\ninf(0)\n",
		"funct bad(n)\n    agar n == 0\n        bargard 1 / 0\n    bargard bad(n - 1)\nbad(3)\n",
		"funct g(a, b = 10)\n    bargard a + b\nchap_khat g(1), g(1, 2), g(b = 3, a = 4)\n",
		"xs = []\nhar i az 0 ta 10\n    agar i % 2 == 0\n        edame\n    agar i > 7\n        beshkan\n    bast xs, i\nchap_khat xs\n",
		"d = {}\nhar w dar [\"a\", \"b\", \"a\"]\n    d[w] = d.begir(w, 0) + 1\nchap_khat d\n",
		"funct sq(xs)\n    bargard [x * x har x dar xs ke x > 1]\nchap_khat sq([1, 2, 3])\n",
		"funct t(n)\n    i = 0\n    ta i < n\n        i += 1\n    bargard \"n={i}\"\nchap_khat t(5)\n",
		"funct m(x)\n    bargard x.nist()\nm(3)\n",
		"funct u()\n    bargard nadide + 1\nu()\n",
		"sabit K = 3\nfunct w()\n    K = 4\nw()\n",
		"funct neg(x)\n    bargard -x\nchap_khat neg(5), neg(-9223372036854775807 - 1)\n",
		"chap_khat 9223372036854775807 + 1\n",
		"darghan Hesab\n    saheb = \"\"\n    mojudi = 0\n\n    funct __sakht__(saheb, mojudi)\n        khod.saheb = saheb\n        khod.mojudi = mojudi\n\n    funct variz(mablagh)\n        khod.mojudi += mablagh\n\n    funct bardasht(mablagh)\n        agar mablagh > khod.mojudi\n            portab khata(\"kam\")\n        khod.mojudi -= mablagh\n        bargard khod\n\nh = Hesab(\"Sara\", 10)\nhar i az 0 ta 5\n    h.variz(i)\nchap_khat h.mojudi\nkoshish\n    h.bardasht(100)\ngiroft khata e\n    chap_khat e.message\nh.bardasht(1).bardasht(100)\n",
		"funct kar(n)\n    koshish\n        agar n < 0\n            portab khata(\"manfi\")\n        chap_khat \"ok\"\n        bargard 1\n    giroft khata e\n        chap_khat \"giroftam: \" + e.message\n        e = 5\n        bargard e\n    nemigir\n        chap_khat \"tamiz\"\nchap_khat kar(1), kar(-1)\n",
		"funct sakhtShomarande()\n    n = 0\n    funct yeki()\n        n = n + 1\n        bargard n\n    bargard yeki\nc = sakhtShomarande()\nc()\nc()\nchap_khat c()\n",
		"xs = [3, 1, 2]\nf = funct(a, b) -> a * 10 + b\nchap_khat f(1, 2), kamil(xs, funct(a, b) -> a + b)\ng = funct(x) -> x.nist\ng(1)\n",
		"agar dorost\n    funct dar_block(x)\n        bargard x + 1\n    chap_khat dar_block(1)\nhar i az 0 ta 3\n    funct h()\n        bargard i\n    chap_khat h()\n",
		"funct bad()\n    koshish\n        bargard 1 / 0\n    nemigir\n        chap_khat \"fin\"\nbad()\n",
		"darghan A\n    funct f()\n        bargard khod.g() + 1\n    funct g()\n        bargard 41\nchap_khat A().f()\nchap_khat A.g()\n",
		"funct f()\n    t = \"\"\n    keep = []\n    har i az 0 ta 200\n        t += \"ab\"\n        agar i % 50 == 0\n            bast keep, t\n    u = keep[1]\n    u += \"ZZ\"\n    v = keep[1]\n    v += \"YY\"\n    bargard [tule(t), u.tab(\"Z\")[0] == keep[1], v.tab(\"Y\")[0] == keep[1], tule(keep[3]), \"a\" + 1, 1 + \"a\"]\nchap_khat f()\n",
		"funct s(x)\n    bargard \"a\" + x\nchap_khat s(1)\n",
	}
	for _, p := range progs {
		runBoth(t, p)
	}
}
