package parser

import (
	"strings"
	"testing"

	"mihan/internal/diag"
)

func TestEachHalfOfAMessageIsInItsOwnLanguage(t *testing.T) {
	cases := []struct {
		src  string
		code string
		fa   string
		en   string
	}{
		{"darghan N\n  x = 1\n  chap kard()\n", "E2001", "kalame-ye kelidi-ye 'chap'", "keyword \"chap\""},
		{"xs = [3, 1, 2\n", "E2001", "']' baste", "a closing bracket"},
		{"2 + * 3\n", "E2023", "'*'", "\"*\""},
	}
	for _, c := range cases {
		bag := &diag.Bag{}
		Parse("t.mhn", []byte(c.src), bag)
		var d *diag.Diagnostic
		for _, x := range bag.Sorted() {
			if x.Code == c.code {
				d = x
				break
			}
		}
		if d == nil {
			t.Errorf("%q: no %s, so the two halves were never checked", c.src, c.code)
			continue
		}
		if !strings.Contains(d.FA, c.fa) {
			t.Errorf("%q: %s Persian half is %q, which does not say %q", c.src, c.code, d.FA, c.fa)
		}
		if !strings.Contains(d.EN, c.en) {
			t.Errorf("%q: %s English half is %q, which does not say %q", c.src, c.code, d.EN, c.en)
		}

		if hasLatinWord(d.FA, "block", "brace", "keyword", "here", "expected") {
			t.Errorf("%q: %s Persian half has an English word in it: %q", c.src, c.code, d.FA)
		}
		if strings.Contains(d.EN, "می‌") || strings.Contains(d.EN, "اینجا") {
			t.Errorf("%q: %s English half has a Persian word in it: %q", c.src, c.code, d.EN)
		}
	}
}

func hasLatinWord(s string, words ...string) bool {
	for _, w := range words {
		i := 0
		for {
			at := strings.Index(s[i:], w)
			if at < 0 {
				break
			}
			at += i
			beforeOK := at == 0 || !isLatinLetter(byte(s[at-1]))
			end := at + len(w)
			afterOK := end == len(s) || !isLatinLetter(byte(s[end]))
			if beforeOK && afterOK {
				return true
			}
			i = at + 1
		}
	}
	return false
}

func isLatinLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
