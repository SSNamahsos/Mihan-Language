package main

import (
	"strings"
	"testing"
)

func TestLooksCompleteWaitsOnlyForSomethingUnfinished(t *testing.T) {
	cases := []struct {
		src  string
		want bool
		why  string
	}{
		{"x = 1\n", true, "one whole statement is a whole program"},
		{"xs = [3, 1, 2]\n", true, "a closed list needs nothing more"},
		{"xs = [3, 1, 2\n", false, "a list left open is waiting for its partner"},
		{"agar x > 10\n", false, "a block left with no body is waiting for its body"},
		{"funct f(a)\n", false, "a funct left with no body is waiting for its body"},
		{"matn = \"salam\n", false, "a text left unclosed is waiting for its last mark"},
		{"2 + * 3\n", true, "a mistake in the middle is a mistake, not a half-typed line"},
		{"xs[0] = \n", false, "an assignment with nothing to its right is unfinished"},
	}
	for _, c := range cases {
		if got := looksComplete(c.src); got != c.want {
			t.Errorf("looksComplete(%q) = %v, want %v (%s)", c.src, got, c.want, c.why)
		}
	}
}

func TestEchoLineShowsTheValueOfAnExpression(t *testing.T) {
	cases := []struct {
		src  string
		want string
		why  string
	}{
		{"2 + 2 * 3\n", "2 + 2 * 3\nchap_khat neshan(2 + 2 * 3)\n", "a number is worth an answer"},
		{"xs\n", "xs\nchap_khat neshan(xs)\n", "a name is worth its value"},
		{"x = 1\n", "x = 1\n", "an assignment does something, so it shows nothing"},
		{"chap \"salam\"\n", "chap \"salam\"\n", "a call already says what it did"},
		{"taghib(forogh(xs))\n", "taghib(forogh(xs))\n", "a call is not asking a question"},
		{"agar x > 10\n", "agar x > 10\n", "half a line is not an expression to answer"},
	}
	for _, c := range cases {
		if got := echoLine(c.src); got != c.want {
			t.Errorf("echoLine(%q) =\n  %q\nwant\n  %q\n(%s)", c.src, got, c.want, c.why)
		}
	}
}

func TestEchoLineDoesNotBreakInterpolatedText(t *testing.T) {

	src := "$\"salam {naam}\"\n"
	got := echoLine(src)
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("echoLine did not end with a newline: %q", got)
	}
	if strings.Count(got, "neshan(") != 1 {
		t.Errorf("echoLine(%q) = %q, which does not echo the value once", src, got)
	}
}
