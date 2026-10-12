package eval

import "testing"

func TestResolvedNamesStillFindTheNearestScope(t *testing.T) {
	cases := map[string]string{

		"x = 1\nfunct f()\n    bargard x\nchap_khat f()\nx = 2\nchap_khat f()\n": "1\n2\n",

		"x = 1\nfunct f(a)\n    agar a\n        x = 5\n    bargard x\nchap_khat f(dorost)\n": "5\n",

		"funct bar()\n    y = 3\n    funct dakhel()\n        bargard y\n    bargard dakhel()\nchap_khat bar()\n": "3\n",

		"fs = []\nhar i az 0 ta 2\n    funct g()\n        bargard i\n    bast fs, g\nchap_khat fs[0](), fs[1]()\n": "01\n",

		"fs = [funct() -> k har k dar [7, 8]]\nchap_khat fs[1]()\n": "8\n",

		"koshish\n    portab khata(\"a\")\ngiroft khata e\n    funct m()\n        bargard e.message\n    chap_khat m()\n": "a\n",

		"funct f(n)\n    agar n < 2\n        bargard n\n    bargard f(n - 1) + f(n - 2)\nchap_khat f(15)\n": "610\n",

		"payam = \"salam\"\ndarghan A\n    funct goft()\n        bargard payam\nchap_khat A().goft()\n": "salam\n",
	}
	for src, want := range cases {
		check(t, src, want)
	}
}
