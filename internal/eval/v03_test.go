package eval

import "testing"

func TestBlockNamesStayVisibleAfterTheBlock(t *testing.T) {
	check(t, "agar dorost\n    a = 1\nchap_khat a\n", "1\n")
	check(t, "koshish\n    f = 5\ngiroft khata e\n    chap 0\nchap_khat f\n", "5\n")
	check(t, "har i az 0 ta 3\n    akhar = i\nchap_khat akhar\n", "2\n")
	check(t, "i = 0\nta i < 3\n    i += 1\n    dide = i * 2\nchap_khat dide\n", "6\n")

	check(t, "funct f()\n    dakhel = 1\nf()\nkoshish\n    chap_khat dakhel\ngiroft khata e\n    chap_khat \"na\"\n", "na\n")

	check(t, "fs = []\nhar i az 0 ta 3\n    bast(fs, funct(z) -> i)\nchap_khat tabdil_kol(fs, funct(g) -> g(0))\n", "[0, 1, 2]\n")
}

func TestLambdaAsABareVerbArgument(t *testing.T) {
	check(t, "fs = []\nbast fs, funct() -> 7\nchap_khat fs[0]()\n", "7\n")
}

func TestListComprehensions(t *testing.T) {
	check(t, "chap_khat [x * x har x dar [1, 2, 3]]\n", "[1, 4, 9]\n")
	check(t, "chap_khat [x har x dar [1, 2, 3, 4] ke x % 2 == 0]\n", "[2, 4]\n")
	check(t, "chap_khat [i har i az 0 ta 3]\n", "[0, 1, 2]\n")
	check(t, "chap_khat [k har k dar \"abc\" agar k != \"b\"]\n", "[\"a\", \"c\"]\n")
}

func TestNumberOrNilAndTime(t *testing.T) {
	check(t, "chap_khat adad_ya_hich(\"12\"), \" \", adad_ya_hich(\"x\"), \" \", adad_ya_hich(\"2.5\")\n", "12 hich 2.5\n")
	check(t, "chap_khat adad_ya_hich(\"x\") ?? -1\n", "-1\n")
	check(t, "t = zaman()\nsabr 0.05\nchap_khat zaman() - t >= 0.04\n", "dorost\n")
	check(t, "chap_khat tule(tarikh(\"YYYY-MM-DD\"))\n", "10\n")
}

func TestIntegerOverflowIsAMistake(t *testing.T) {
	for _, src := range []string{
		"chap_khat 9223372036854775807 + 1\n",
		"chap_khat -9223372036854775807 - 2\n",
		"chap_khat 4294967296 * 4294967296\n",
		"chap_khat 2 ** 64\n",
	} {
		checkFail(t, src, "64 bits")
	}
	check(t, "chap_khat 2 ** 62\n", "4611686018427387904\n")
}

func TestTextIsQuotedInsideLists(t *testing.T) {
	check(t, "chap_khat [\"a\", 1]\n", "[\"a\", 1]\n")
	check(t, "chap_khat \"a\"\n", "a\n")
	check(t, "chap_khat {\"k\": \"v\"}\n", "{\"k\": \"v\"}\n")
}
