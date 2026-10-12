package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"mihan/internal/eval"
	"mihan/internal/token"
)

func cmdDoc(args []string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "Mihan — the whole language on one page\n\n")
	fmt.Fprintf(&b, "Every word below is one row of the compiler's own table, so this\n")
	fmt.Fprintf(&b, "page and the language cannot disagree. A Persian word is always the\n")
	fmt.Fprintf(&b, "name, and the English name after it always works too.\n\n")

	b.WriteString("VALUES\n")
	for _, row := range [][2]string{
		{"arz", "a whole number"},
		{"kasr", "a number with a fraction"},
		{"matn", "text"},
		{"harf", "one character"},
		{"bool", "dorost or ghalat"},
		{"list", "a growable sequence"},
		{"dict", "a table of names against values, kept in the order it arrived"},
		{"set", "values that cannot repeat"},
		{"hich", "the absence of a value"},
		{"batri", "no value at all"},
		{"khata", "a failure carrying a message"},
	} {
		fmt.Fprintf(&b, "  %-10s %s\n", row[0], row[1])
	}
	b.WriteString("\n  A name with a type in front of it starts at that type's zero, and\n")
	b.WriteString("  the type is a promise about what may go into it: `arz shomar` is 0.\n\n")

	groupTitles := map[token.Group]string{
		token.GroupControl:     "CHOOSING AND REPEATING",
		token.GroupDeclaration: "DECLARING",
		token.GroupOperator:    "OPERATORS THAT ARE WORDS",
		token.GroupType:        "TYPES",
		token.GroupModifier:    "MODIFIERS",
		token.GroupBuiltin:     "BUILTIN WORDS",
		token.GroupMisc:        "TRUTH VALUES",
	}
	for _, g := range token.KeywordGroups {
		title, ok := groupTitles[g]
		if !ok {
			continue
		}
		rows := keywordsIn(g)
		if len(rows) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s\n", title)
		for _, kw := range rows {
			names := kw.Name
			if len(kw.Aliases) > 0 {
				names += "  (also " + strings.Join(kw.Aliases, ", ") + ")"
			}
			fmt.Fprintf(&b, "  %-26s %s\n", names, kw.Doc)
		}
		b.WriteByte('\n')
	}

	b.WriteString("SYMBOLS\n")
	ops := make([]*token.Operator, 0, len(token.SymbolsLongestFirst()))
	for _, sym := range token.SymbolsLongestFirst() {
		if op := token.LookupOperator(sym); op != nil {
			ops = append(ops, op)
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].Sym < ops[j].Sym })
	for _, op := range ops {
		if op.Doc == "" {
			continue
		}
		fmt.Fprintf(&b, "  %-6s %s\n", op.Sym, op.Doc)
	}
	b.WriteByte('\n')

	b.WriteString("BUILTINS\n")
	in := eval.New(os.Stdout, os.Stdin)
	for _, line := range in.BuiltinDocs() {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	b.WriteByte('\n')

	b.WriteString(`RULES WORTH KNOWING

  A colon is optional.        agar x > 10      and      agar x > 10:   are the same.
  Braces work too.           if x > 10 { chap 1 }  is accepted, for the habits
                             C++, C# and Go leave behind.

  / always gives a fraction. 7 / 2 is 3.5. Use ghesmat for the whole part.

  A method is called with a dot, because it works on the value in front of it:
  "salam".gooshband(). The same name also works as a function with brackets:
  gooshband("salam").

  Whitespace decides one thing. takfir-3 calls takfir with -3; takfir - 3
  subtracts 3. The space you typed says which one you meant.

  A call without brackets runs to the end of the line, or to the next comma:
  chap_khat 1 + 2 * 3 prints 7, and chap_khat tule xs + 1 means
  chap_khat(tule(xs + 1)). Brackets say the other thing when that is what you
  meant. It is the rule a shell uses when it splits a line into words.

  Declaring and assigning are the same act. x = 1 introduces the name; x += 1
  changes one that already exists; sabit may never change at all.

  portab throws, bargard gives back. Returning a khata and throwing one are
  different acts.

  A mistake Mihan finds by itself stops the program and says where. A situation
  a program should handle is a khata you portab and someone else catches.
`)
	fmt.Print(b.String())
	return nil
}

func keywordsIn(g token.Group) []token.Keyword {
	var out []token.Keyword
	for _, kw := range token.KeywordList() {
		if kw.Group == g {
			out = append(out, kw)
		}
	}
	return out
}
