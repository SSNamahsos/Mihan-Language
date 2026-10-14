package eval

import (
	"bytes"
	"strings"
	"testing"

	"mihan/internal/diag"
	"mihan/internal/parser"
	rt "mihan/internal/runtime"
)

func run(t *testing.T, src string) string {
	t.Helper()
	bag := &diag.Bag{}
	file := parser.Parse("test.mhn", []byte(src), bag)
	if bag.HasErrors() {
		var lines []string
		for _, d := range bag.Sorted() {
			lines = append(lines, d.Code+": "+d.EN)
		}
		t.Fatalf("the program did not read:\n%s\n  %s", src, strings.Join(lines, "\n  "))
	}
	var out bytes.Buffer
	in := New(&out, nil)
	if err := in.RunFile(file); err != nil {
		t.Fatalf("the program stopped: %s\n  in:\n%s", in.RenderError(err, []byte(src)), src)
	}
	return out.String()
}

func runFail(t *testing.T, src string) string {
	t.Helper()
	bag := &diag.Bag{}
	file := parser.Parse("test.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not even read:\n%s", src)
	}
	var out bytes.Buffer
	in := New(&out, nil)
	err := in.RunFile(file)
	if err == nil {
		t.Fatalf("I expected this to stop, but it printed:\n%s\n  in:\n%s", out.String(), src)
	}
	p, ok := rt.AsPanic(err)
	if !ok {
		return err.Error()
	}
	ev, ok := p.Value.(*rt.ErrorValue)
	if !ok {
		return rt.Show(p.Value)
	}
	return string(ev.Message)
}

func check(t *testing.T, src, want string) {
	t.Helper()
	if got := run(t, src); got != want {
		t.Errorf("%s\n  got:  %q\n  want: %q", src, got, want)
	}
}

func checkFail(t *testing.T, src, wantIn string) {
	t.Helper()
	got := runFail(t, src)
	if !strings.Contains(got, wantIn) {
		t.Errorf("%s\n  the message was %q\n  it should mention %q", src, got, wantIn)
	}
}

func TestHello(t *testing.T) {
	check(t, "chap \"salam\"\n", "salam")
}

func TestPrintWithoutBracketsMatchesBrackets(t *testing.T) {
	check(t, "chap \"a\"\nchap(\"a\")\n", "aa")
}

func TestArithmetic(t *testing.T) {
	cases := map[string]string{
		"1 + 2":         "3",
		"1 + 2 * 3":     "7",
		"(1 + 2) * 3":   "9",
		"7 / 2":         "3.5",
		"7 ghesmat 2":   "3",
		"ghesmat(7, 2)": "3",
		"7 % 3":         "1",
		"2 ** 10":       "1024",
		"-2 ** 2":       "-4",
		"10 - 2 - 3":    "5",
		"100 / 10 / 2":  "5.0",
		"3 * 4 % 5":     "2",
		"takfir(-5)":    "5",
		"balia(1, 2)":   "2",
		"dope(1, 2)":    "1",
		"goosh(2.4)":    "2",
		"goosh(2.5)":    "3",
		"gooshbad(2.9)": "2",
		"sher(2.1)":     "3",
		"goroosh(2.9)":  "2",
		"5 <=> 3":       "1",
		"3 <=> 3":       "0",
		"3 <=> 5":       "-1",
		"10 & 6":        "2",
		"10 | 3":        "11",
		"10 ^ 3":        "9",
		"1 << 4":        "16",
		"256 >> 4":      "16",
		"~0":            "-1",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestDivisionAndFloorDivisionMatchMaths(t *testing.T) {
	check(t, "chap_khat -7 ghesmat 2\n", "-4\n")
	check(t, "chap_khat -7 % 2\n", "1\n")
	check(t, "chap_khat 7 ghesmat -2\n", "-4\n")
}

func TestComparison(t *testing.T) {
	cases := map[string]string{
		"1 < 2":          "dorost",
		"2 < 1":          "ghalat",
		"1 <= 1":         "dorost",
		"1 > 2":          "ghalat",
		"1 >= 1":         "dorost",
		"1 == 1.0":       "dorost",
		"1 != 2":         "dorost",
		"'a' < 'b'":      "dorost",
		"[1,2] < [1,3]":  "dorost",
		"[1,2] == [1,2]": "dorost",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestLogic(t *testing.T) {
	cases := map[string]string{
		"dorost va dorost":    "dorost",
		"dorost va ghalat":    "ghalat",
		"ghalat ya dorost":    "dorost",
		"ghalat ya ghalat":    "ghalat",
		"na dorost":           "ghalat",
		"na ghalat":           "dorost",
		"!dorost":             "ghalat",
		"na dorost ya dorost": "dorost",
		"!dorost || dorost":   "dorost",
		"na 1 == 2":           "dorost",
		"!ghalat va dorost":   "dorost",
		"1 < 2 va 2 < 3":      "dorost",
		"1 > 2 ya 2 < 3":      "dorost",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestShortCircuit(t *testing.T) {

	check(t, "x = 0\nchap_khat x != 0 va 10 / x > 1\n", "ghalat\n")
	check(t, "x = 0\nchap_khat x == 0 ya 10 / x > 1\n", "dorost\n")
}

func TestText(t *testing.T) {
	cases := map[string]string{
		"\"a\" + \"b\"":                 "ab",
		"\"a\" + 1":                     "a1",
		"1 + \"a\"":                     "1a",
		"tule \"salam\"":                "5",
		"tule \"سلام\"":                 "4",
		"\"salam\".gooshband()":         "SALAM",
		"\"SALAM\".kiawtar()":           "salam",
		"\"  a  \".jabeja()":            "a",
		"\"a,b\".tab(\",\")":            `["a", "b"]`,
		"\"salam\".barmigefth(\"sa\")":  "dorost",
		"\"salam\".bapakhsh(\"am\")":    "dorost",
		"\"salam\".shamil(\"ala\")":     "dorost",
		"\"aXa\".jigarre(\"X\", \"-\")": "a-a",
		"neshan(\"a\")":                 "\"a\"",
		"neshan([1, 2])":                "[1, 2]",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestInterpolation(t *testing.T) {
	check(t, "naam = \"Mihan\"\nchap $\"salam {naam}! {1 + 1}\"\n", "salam Mihan! 2")
	check(t, "xs = [1, 2]\nchap $\"tule: {tule xs}\"\n", "tule: 2")

	check(t, "chap $\"nope\"\n", "nope")
	check(t, "chap_khat \"{x}\"\n", "{x}\n")
	check(t, "chap_khat $\"{{x}}\"\n", "{x}\n")
}

func TestLists(t *testing.T) {
	cases := map[string]string{
		"[1, 2, 3]":                            "[1, 2, 3]",
		"tule [1, 2, 3]":                       "3",
		"[1, 2, 3][0]":                         "1",
		"[1, 2, 3][-1]":                        "3",
		"[1, 2, 3][1:]":                        "[2, 3]",
		"[1, 2, 3][:2]":                        "[1, 2]",
		"[1, 2, 3][0..2]":                      "[1, 2, 3]",
		"[1, 2] + [3]":                         "[1, 2, 3]",
		"[1, 2] + 3":                           "[1, 2, 3]",
		"taghib([3, 1, 2])":                    "[1, 2, 3]",
		"forogh([1, 2, 3])":                    "[3, 2, 1]",
		"barmadar([1, 2], 2)":                  "dorost",
		"hazar([1, 2, 3], 3)":                  "2",
		"hazar([1, 2, 3], 9)":                  "-1",
		"por([1, 2, 3], \"-\")":                "1-2-3",
		"filter([1,2,3], funct(x) -> x > 1)":   "[2, 3]",
		"tabdil_kol([1,2], funct(x) -> x * 2)": "[2, 4]",
		"kamil([1,2,3], funct(a, b) -> a + b)": "6",
		"hamej([1,2], funct(x) -> x > 0)":      "dorost",
		"kamej([1,2], funct(x) -> x > 5)":      "ghalat",
		"setare([1, 1, 2])":                    "{1, 2}",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestListMutation(t *testing.T) {
	check(t, "xs = [1]\nbast xs, 2\nchap_khat xs\n", "[1, 2]\n")
	check(t, "xs = [1, 3]\nbiya xs, 1, 2\nchap_khat xs\n", "[1, 2, 3]\n")
	check(t, "xs = [1, 2, 3]\nchap_khat bardavar xs\nchap_khat xs\n", "3\n[1, 2]\n")

	check(t, "a = [1]\nb = a\nbast b, 2\nchap_khat a\n", "[1, 2]\n")
}

func TestDicts(t *testing.T) {
	cases := map[string]string{
		"{\"a\": 1}":                "{\"a\": 1}",
		"{\"a\": 1}[\"a\"]":         "1",
		"tule {\"a\": 1, \"b\": 2}": "2",
		"dast({\"a\": 1})":          `["a"]`,
		"meghdar({\"a\": 1})":       "[1]",
		"\"a\" dar {\"a\": 1}":      "dorost",
		"\"b\" dar {\"a\": 1}":      "ghalat",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestDictKeepsItsOrder(t *testing.T) {

	src := "d = {\"z\": 1, \"a\": 2, \"m\": 3}\nchap_khat d\nchap_khat d\n"
	check(t, src, "{\"z\": 1, \"a\": 2, \"m\": 3}\n{\"z\": 1, \"a\": 2, \"m\": 3}\n")
}

func TestWholeFloatIsOneKey(t *testing.T) {
	check(t, "chap_khat tule {1: \"a\", 1.0: \"b\"}\n", "1\n")
}

func TestSets(t *testing.T) {
	cases := map[string]string{
		"{1, 2, 2, 3}":   "{1, 2, 3}",
		"tule {1, 1, 2}": "2",
		"2 dar {1, 2}":   "dorost",
		"9 dar {1, 2}":   "ghalat",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestControlFlow(t *testing.T) {
	check(t, "har i az 0 ta 3\n    chap_khat i\n", "0\n1\n2\n")
	check(t, "har i az 0 ta 3+\n    chap_khat i\n", "0\n1\n2\n3\n")
	check(t, "har x dar [1, 2]\n    chap_khat x\n", "1\n2\n")
	check(t, "har i az 0 ta 5 ke i % 2 == 0\n    chap_khat i\n", "0\n2\n4\n")
	check(t, "i = 0\nta i < 3\n    chap_khat i\n    i = i + 1\n", "0\n1\n2\n")
	check(t, "har i az 0 ta 5\n    agar i == 2\n        shek\n    chap_khat i\n", "0\n1\n")
	check(t, "har i az 0 ta 4\n    agar i == 1\n        edame\n    chap_khat i\n", "0\n2\n3\n")
	check(t, "agar ghalat\n    chap_khat 1\nwarna agar dorost\n    chap_khat 2\nwarna\n    chap_khat 3\n", "2\n")
	check(t, "agar ghalat\n    chap_khat 1\n", "")
	check(t, "x = agar dorost 1 warna 2\nchap_khat x\n", "1\n")
}

func TestNestedLoopsAndBreak(t *testing.T) {

	src := "har i az 0 ta 3\n    har j az 0 ta 3\n        agar j == 1\n            shek\n        chap_khat $\"{i}{j} \"\nchap_khat \"\"\n"
	check(t, src, "00 \n10 \n20 \n\n")
}

func TestFunctions(t *testing.T) {
	check(t, "funct jam(a, b)\n    bargard a + b\nchap_khat jam(1, 2)\n", "3\n")
	check(t, "funct f(a, b = 10)\n    bargard a + b\nchap_khat f(1)\n", "11\n")
	check(t, "funct f(a, b = 10)\n    bargard a + b\nchap_khat f(1, 2)\n", "3\n")
	check(t, "funct f(a, b = 10)\n    bargard a + b\nchap_khat f(b: 2, a: 1)\n", "3\n")

	check(t, "funct f()\n    chap_khat 1\nchap_khat f()\n", "1\nhich\n")

	check(t, "x = 1\nfunct f(x)\n    bargard x\nchap_khat f(2)\nchap_khat x\n", "2\n1\n")

	check(t, "funct tule(x)\n    bargard 99\nchap_khat tule([1, 2])\n", "99\n")
}

func TestRecursion(t *testing.T) {
	check(t, "funct f(n)\n    agar n <= 1\n        bargard 1\n    bargard n * f(n - 1)\nchap_khat f(10)\n", "3628800\n")
}

func TestClosures(t *testing.T) {
	src := "funct shostar(add)\n    bargard funct(x) -> x + add\n" +
		"p = shostar(10)\n" +
		"chap_khat p(5)\n"
	check(t, src, "15\n")

	src2 := "funct shostar()\n    shomar = 0\n    bargard funct() -> shomar + 1\n" +
		"a = shostar()\nb = shostar()\n" +
		"chap_khat a()\nchap_khat a()\n"
	check(t, src2, "1\n1\n")
}

func TestClasses(t *testing.T) {
	src := `darghan Noghat
    baza x
    baza y

    __sakht__(x, y)
        khod.x = x
        khod.y = y

    vectoor()
        bargard goosh((khod.x ** 2 + khod.y ** 2) ** 0.5)

n = Noghat(3, 4)
chap_khat n.x
chap_khat n.vectoor()
chap_khat n
`
	check(t, src, "3\n5\nNoghat{x: 3, y: 4}\n")
}

func TestClassFieldsGetTheirDefaults(t *testing.T) {
	src := `darghan S
    a = 1
    b = ""

s = S()
chap_khat s.a
chap_khat s.b
`
	check(t, src, "1\n\n")
}

func TestInheritance(t *testing.T) {
	src := `darghan A
    funct hey()
        chap_khat "A"

darghan B < A>
    funct hey()
        chap_khat "B"

B().hey()
A().hey()
`
	check(t, src, "B\nA\n")
}

func TestKhodIsTheCurrentObject(t *testing.T) {
	src := `darghan Shomarande
    adad = 0

    __sakht__(adad)
        khod.adad = adad

    ezafe_kon(n)
        khod.adad = khod.adad + n
        bargard khod.adad

s = Shomarande(1)
chap_khat s.ezafe_kon(2)
chap_khat s.ezafe_kon(3)
`
	check(t, src, "3\n6\n")
}

func TestStructs(t *testing.T) {
	src := `sakht Noghat
    x: kasr
    y: kasr

n = Noghat()
n.x = 1
n.y = 2
chap_khat n.x
`
	check(t, src, "1\n")
}

func TestEnums(t *testing.T) {
	src := `shenase Rang
    sefid = 0
    siah = 1

chap_khat Rang.siah
chap_khat Rang.sefid
`
	check(t, src, "1\n0\n")
}

func TestErrors(t *testing.T) {
	src := `funct risky(n)
    agar n < 0
        portab khata("manfi nemidavad")
    chap_khat "majeb"

koshish
    risky(-1)
giroft khata e
    chap_khat $"dar gayefte: {e.message}"
`
	check(t, src, "dar gayefte: manfi nemidavad\n")
}

func TestFinallyRunsEitherWay(t *testing.T) {
	src := `funct f()
    portab khata("x")

koshish
    f()
giroft khata e
    chap_khat "giry"
nemigir
    chap_khat "nazir"
`
	check(t, src, "giry\nnazir\n")
}

func TestErrorCarriesData(t *testing.T) {
	src := `koshish
    portab khata("bad", 42)
giroft khata e
    chap_khat e.data
`
	check(t, src, "42\n")
}

func TestErrorStopsUncaught(t *testing.T) {
	checkFail(t, "portab khata(\"oops\")\n", "oops")

	got := runFail(t, "har i az 0 ta 3\n    chap_khat 1 / (i - 1)\n")
	if !strings.Contains(got, "taghsim") {
		t.Errorf("the message should name the problem, got %q", got)
	}
}

func TestNamesThatAreNotThere(t *testing.T) {
	checkFail(t, "chap_khat nada\n", "nada")
	checkFail(t, "f(1)\n", "hich ja nasakhtam")
	checkFail(t, "darghan A\n    x = 1\n\na = A()\nchap_khat a.nope\n", "nist")
	checkFail(t, "x = 1\nchap_khat x.nope\n", "az arz nemishe ozv khoond")
}

func TestMistakesAreNamed(t *testing.T) {
	checkFail(t, "1 / 0\n", "bar sefr taghsim")
	checkFail(t, "ghesmat(1, 0)\n", "taghsim")
	checkFail(t, "xs = [1]\nxs[5]\n", "dar list-i be boland-i")
	checkFail(t, "\"abc\"[9]\n", "dar matn-i be boland-i")
	checkFail(t, "d = {}\nd[\"nope\"]\n", "kelid-e 'nope' dar dict nist")
	checkFail(t, "sabit x = 1\nx = 2\n", "sabit")
	checkFail(t, "tule 5\n", "tule-ye arz")
	checkFail(t, "5 + \"a\" - 1\n", "")
	checkFail(t, "har i dar 5\n    chap 1\n", "too halghe charkhoond")
	checkFail(t, "1 < \"a\"\n", "moghayese kard")
}

func TestKhorojiUnlocked(t *testing.T) {

	check(t, "sabit A = 5\nchap_khat A\n", "5\n")
}

func TestTruthiness(t *testing.T) {

	cases := map[string]string{
		"0":          "ghalat",
		"1":          "dorost",
		"-1":         "dorost",
		"0.0":        "ghalat",
		"\"\"":       "ghalat",
		"\"a\"":      "dorost",
		"hich":       "ghalat",
		"[]":         "ghalat",
		"[1]":        "dorost",
		"{}":         "ghalat",
		"{\"a\": 1}": "dorost",
		"{} == {}":   "",
	}
	for src, want := range cases {
		if want == "" {
			continue
		}
		check(t, "agar "+src+"\n    chap_khat \"dorost\"\nwarna\n    chap_khat \"ghalat\"\n", want+"\n")
	}
}

func TestTypedDeclarations(t *testing.T) {
	check(t, "arz x = 1\nchap_khat x\n", "1\n")
	check(t, "arz x\nchap_khat x\n", "0\n")
	check(t, "kasr x\nchap_khat x\n", "0.0\n")
	check(t, "matn x\nchap_khat neshan(x)\n", "\"\"\n")
	check(t, "bool x\nchap_khat x\n", "ghalat\n")

	check(t, "kasr x = 1\nchap_khat x\n", "1.0\n")
	check(t, "arz x = 3.9\nchap_khat x\n", "3\n")

	check(t, "kasr x = 1.5\nchap_khat x * 2\n", "3.0\n")
}

func TestSabit(t *testing.T) {
	check(t, "sabit A = 1\nchap_khat A\n", "1\n")
	checkFail(t, "sabit A = 1\nA = 2\n", "sabit")
}

func TestRanges(t *testing.T) {

	cases := map[string]string{
		"rang(4)":                "0..4",
		"rang(1, 4)":             "1..4",
		"rang(1, 5, 2)":          "1..2..5",
		"0..3":                   "0..3",
		"0..+3":                  "0..+3",
		"tule rang(1, 5)":        "4",
		"por(rang(1, 5), \"-\")": "1-2-3-4",
		"taghib(rang(1, 5))":     "[1, 2, 3, 4]",
		"filter(rang(0, 6), funct(x) -> x % 2 == 0)": "[0, 2, 4]",
	}
	for src, want := range cases {
		check(t, "chap_khat "+src+"\n", want+"\n")
	}
}

func TestWalkEveryKindOfCollection(t *testing.T) {
	cases := map[string]string{
		"[1, 2]":               "12",
		"\"ab\"":               "ab",
		"rang(1,3)":            "12",
		"{\"a\": 1, \"b\": 2}": "ab",
	}
	for src, want := range cases {
		check(t, "out = \"\"\nhar x dar "+src+"\n    out = out + x\nchap_khat out\n", want+"\n")
	}
}

func TestMethodOnValueIsFoundByName(t *testing.T) {

	checkFail(t, "xs = [1]\nxs.nope\n", "chizi be esm-e 'nope' nist")
	checkFail(t, "\"a\".nope()\n", "chizi be esm-e 'nope' nist")
}

func TestCallingANumberIsExplained(t *testing.T) {
	got := runFail(t, "\"salam\".tule()\n")
	if !strings.Contains(got, "na funct-i ke beshe seda zad") {
		t.Errorf("the message should say it is not something to call, got %q", got)
	}
}

func TestBuiltinNamesAreValues(t *testing.T) {

	check(t, "funct f(g, x)\n    bargard g(x)\nchap_khat f(tule, [1, 2, 3])\n", "3\n")
}

func TestReadingTextDoesNotCopyTheWholeThing(t *testing.T) {
	allocs := testing.AllocsPerRun(200, func() {

		if _, ok := rt.RuneAt(strings.Repeat("abcdefghij", 500), 2500); !ok {
			t.Fatal("the character should be there")
		}
	})
	if allocs > 1 {
		t.Errorf("reading one character allocated %v times; it should be at most once", allocs)
	}
	check(t, "s = \"abcdefghij\" * 200\nchap_khat s[5]\nchap_khat tule s\n", "f\n2000\n")
}

func TestTextIndexingHandlesPersian(t *testing.T) {

	check(t, "s = \"سلام\"\nchap_khat s[0]\n", "س\n")
	check(t, "s = \"سلام\"\nchap_khat s[1]\n", "ل\n")
	check(t, "s = \"سلام\"\nchap_khat s[2]\n", "ا\n")
	check(t, "s = \"سلام\"\nchap_khat s[-1]\n", "م\n")
	check(t, "s = \"abcd\"\nchap_khat s[1:3]\n", "bc\n")

	check(t, "s = \"abcd\"\nchap_khat s[..2]\n", "abc\n")
	check(t, "s = \"abcd\"\nchap_khat s[1..3]\n", "bcd\n")
	check(t, "s = \"abcd\"\nchap_khat s[:2]\n", "ab\n")
	check(t, "har c dar \"سلام\"\n    chap_khat c\n", "س\nل\nا\nم\n")
}

func TestASliceBeyondTheEndIsEmptyNotACrash(t *testing.T) {
	check(t, "chap_khat [1, 2, 3][5:]\n", "[]\n")
	check(t, "chap_khat [1, 2, 3][3:]\n", "[]\n")
	check(t, "chap_khat [1, 2, 3][5:9]\n", "[]\n")
	check(t, "chap_khat [1, 2, 3][:9]\n", "[1, 2, 3]\n")
	check(t, "chap_khat \"abc\"[5:]\n", "\n")
	check(t, "chap_khat \"abc\"[:9]\n", "abc\n")
	check(t, "chap_khat {1, 2}[5:]\n", "[]\n")
	check(t, "chap_khat (1..3)[5:]\n", "[]\n")
}

func TestNameThatIsATypeWordIsStillAVariable(t *testing.T) {

	check(t, "matn = \"abc\"\nchap_khat matn[1]\n", "b\n")
	check(t, "list = [1, 2]\nchap_khat list[0]\n", "1\n")
}

func TestStringRepetition(t *testing.T) {
	check(t, "chap_khat \"ab\" * 3\n", "ababab\n")
	check(t, "chap_khat tule \"ab\" * 3\n", "6\n")
	check(t, "chap_khat 3 * \"-\"\n", "---\n")
}

func TestOutputOrderIsStable(t *testing.T) {
	src := "for i az 0 ta 3\n    chap_khat i\n"
	want := "0\n1\n2\n"
	for attempt := 0; attempt < 20; attempt++ {
		if got := run(t, src); got != want {
			t.Fatalf("attempt %d gave %q", attempt, got)
		}
	}
}

func TestStepLimitStopsARunaway(t *testing.T) {
	bag := &diag.Bag{}
	file := parser.Parse("t.mhn", []byte("ta dorost\n    chap 1\n"), bag)
	if bag.HasErrors() {
		t.Fatal(bag.Sorted()[0].Message())
	}
	var out bytes.Buffer
	in := New(&out, nil)
	in.MaxSteps = 5000
	err := in.RunFile(file)
	if err == nil {
		t.Fatal("a loop that never ends should have been stopped")
	}
	if !strings.Contains(runFailText(t, err), "bish az had kar kard") {
		t.Errorf("the message should say the program ran too long, got %q", err)
	}
}

func runFailText(t *testing.T, err error) string {
	t.Helper()
	p, ok := rt.AsPanic(err)
	if !ok {
		return err.Error()
	}
	if ev, ok := p.Value.(*rt.ErrorValue); ok {
		return string(ev.Message)
	}
	return rt.Show(p.Value)
}

func TestDeepRecursionIsStoppedNotCrashed(t *testing.T) {
	checkFail(t, "funct f()\n    f()\nf()\n", "bish az had hamdige ro seda zadan")
}

func TestNoNameIsCalledBeforeItIsGiven(t *testing.T) {
	checkFail(t, "x = y + 1\ny = 2\n", "y")
}
