package main

import "testing"

func TestTheExpectationMarkerIsReadEitherWay(t *testing.T) {
	cases := []struct {
		name string
		src  string
		kind expectKind
		want string
	}{
		{"a block with no space", "#khoroji\n#| 8\n", expectOutput, "8\n"},
		{"a block with a space", "# khoroji\n#| 8\n", expectOutput, "8\n"},
		{"a block with spaces around it", "  #  khoroji  \n  #| 8  \n", expectOutput, "8\n"},
		{"several lines", "#khoroji\n#| 1\n#| 2\n#| 3\n", expectOutput, "1\n2\n3\n"},
		{"a blank expected line", "#khoroji\n#| 1\n#|\n#| 3\n", expectOutput, "1\n\n3\n"},
		{"the inline form", "#khoroji: \"8\\n\"\n", expectOutput, "8\n"},
		{"the inline form with a space", "# khoroji: \"8\\n\"\n", expectOutput, "8\n"},
		{"a failure is expected", "#khata:\n", expectError, ""},
		{"a failure is expected, spaced", "# khata:\n", expectError, ""},
		{"nothing is expected", "x = 1\n", expectNone, ""},
		{"a comment is not an expectation", "# just a note\nx = 1\n", expectNone, ""},

		{"a file saved with a mark", "\xEF\xBB\xBF# khoroji\n#| 8\n", expectOutput, "8\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kind, got := expectationOf([]byte(c.src))
			if kind != c.kind {
				t.Fatalf("the file said %v, want %v", kind, c.kind)
			}
			if got != c.want {
				t.Errorf("the file said it expects %q, want %q", got, c.want)
			}
		})
	}
}

func TestAStrayBlockIsNamed(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want int
	}{
		{"a block with no marker", "x = 1\n#| 8\n", 2},
		{"a block with no marker, spaced", "x = 1\n#| 8\n", 2},
		{"a block with a marker", "#khoroji\n#| 8\n", 0},
		{"a block with a marker, spaced", "# khoroji\n#| 8\n", 0},
		{"nothing at all", "x = 1\n", 0},
		{"a stray block after code", "#khoroji\n#| 8\nx = 2\n#| 9\n", 4},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := strayBlock([]byte(c.src)); got != c.want {
				t.Errorf("strayBlock(%q) = %d, want %d", c.src, got, c.want)
			}
		})
	}
}
