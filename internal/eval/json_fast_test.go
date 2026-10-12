package eval

import (
	"encoding/json"
	"strings"
	"testing"

	rt "mihan/internal/runtime"
)

func TestFastJSONAgrees(t *testing.T) {
	cases := []string{
		`{"a": 1, "b": [1, 2.5, -3, 0, 1e3, 2.0, -0.5E-2], "c": {"d": null, "e": true, "f": false}}`,
		`[]`, `{}`, `"salam"`, `"سلام \u0633 \n \"q\" \\ \/"`, `123456789012345678901`, `-0`,
		`[1, 2,]`, `{"a" 1}`, `01`, `1.`, `.5`, `tru`, `[1] [2]`, "\"a\x01\"", `{"a": 1, "a": 2}`,
		`[[[[[]]]]]`, `"\ud83d\ude00"`, `{"k": "v"`, `  [ 1 , 2 ]  `, `nul`, `"\x"`,
	}
	for _, c := range cases {
		text := strings.TrimSpace(c)
		fast, ok := fastJSON(text)
		dec := json.NewDecoder(strings.NewReader(text))
		dec.UseNumber()
		slow, err := readJSON(dec)
		if err == nil && dec.More() {
			err = errMore
		}
		if !ok {
			continue
		}
		if err != nil {
			t.Fatalf("%q: quick reader accepted what encoding/json refuses (%v)", c, err)
		}
		if rt.Repr(fast) != rt.Repr(slow) || rt.TypeName(fast) != rt.TypeName(slow) {
			t.Fatalf("%q: quick %s, slow %s", c, rt.Repr(fast), rt.Repr(slow))
		}
	}
}

var errMore = json.Unmarshal([]byte("x"), new(any))
