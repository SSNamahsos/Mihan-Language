package eval

import "testing"

func TestNamedArgumentWithEquals(t *testing.T) {
	src := "funct f(a, b = 10)\n    bargard a + b\nchap_khat f(1, b = 2)\nchap_khat f(b = 5, a = 1)\nchap_khat f(1, b: 3)\n"
	check(t, src, "3\n6\n4\n")
}

func TestCastTakesATypeWord(t *testing.T) {
	src := "chap_khat tabdil(\"42\", arz) + 1\nchap_khat tabdil(\"2.5\", kasr) * 2\nchap_khat tabdil(7, matn) + \"!\"\nchap_khat tabdil(3.9, int)\n"
	check(t, src, "43\n5.0\n7!\n3\n")
}

func TestCastRefusesTextThatIsNotAWholeNumber(t *testing.T) {
	runFail(t, "chap_khat tabdil(\"12abc\", arz)\n")
}

func TestZip(t *testing.T) {
	check(t, "chap_khat zip([1, 2, 3], [\"a\", \"b\"])\n", "[[1, \"a\"], [2, \"b\"]]\n")
}

func TestInheritedFieldDefaults(t *testing.T) {
	src := "darghan A\n    v = 1\ndarghan B < A\n    funct get()\n        bargard 100 + khod.v\nchap_khat B().get()\n"
	check(t, src, "101\n")
}

func TestEachObjectHasItsOwnListField(t *testing.T) {
	src := "darghan S\n    items = []\na = S()\nb = S()\nbast a.items, 1\nchap_khat b.items\nchap_khat a.items\n"
	check(t, src, "[]\n[1]\n")
}

func TestTypeBeforeParameterName(t *testing.T) {
	check(t, "funct jam(arz a, arz b) -> arz\n    bargard a + b\nchap_khat jam(2, 3)\n", "5\n")
}

func TestABuiltinsNameAsAVariableCanBeChanged(t *testing.T) {
	src := "shomar = 0\nfunct z()\n    shomar += 1\nz()\nz()\nchap_khat shomar\n"
	check(t, src, "2\n")
}

func TestCoalesce(t *testing.T) {
	check(t, "chap_khat mohit(\"MIHAN_SURELY_NOT_SET\") ?? \"y\"\nchap_khat 5 ?? 6\n", "y\n5\n")
}

func TestDictGet(t *testing.T) {
	check(t, "d = {\"a\": 1}\nchap_khat d.begir(\"a\")\nchap_khat d.begir(\"b\")\nchap_khat d.begir(\"b\", 7)\n", "1\nhich\n7\n")
}
