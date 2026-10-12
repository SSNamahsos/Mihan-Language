# AGENTS.md — how to work on Mihan

Mihan is a programming language whose keywords are Persian words written in Latin
letters, so that a program reads the way a Persian speaker talks. It is written
in Go. The compiler and everything it needs live under `internal/`.

## The commands

```
.\build.ps1            build, run every test, run vet    ← start here
.\build.ps1 build      just build bin\mihan.exe
.\build.ps1 test       the Go tests and the documentation's examples
.\build.ps1 docs       only the documentation's examples
.\build.ps1 fmt        tidy the Go source and the Mihan programs
.\build.ps1 check      go vet, and confirm gofmt is happy
.\build.ps1 bench      the benchmarks
.\build.ps1 run x.mhn  build then run a program
.\build.ps1 repl       the line-at-a-time reader
```

`gofmt -w .` before you commit anything in Go. `go vet ./...` must stay quiet.

## How the language fits together

Each package does one job and knows nothing about the others, which is what makes
it possible to swap the interpreter out for a virtual machine later without
touching anything a program can see.

```
internal/token     every word, every operator, every alias. THE source of truth.
internal/lexer     text -> tokens, with indentation turned into INDENT/DEDENT
internal/ast       the syntax tree
internal/parser    tokens -> tree, driven by the precedence in the token table
internal/ir        tree -> instructions. The boundary a machine runs behind.
internal/analysis  what a name means and where it was written
internal/format    the printer, shared by `mihan fmt` and the editor
internal/lsp       the language server: the words, the mistakes, the jumps
internal/runtime   what a value is, what it prints as, how two compare
internal/eval      tree -> running program, and the 93 builtins
internal/driver    the order of the phases; the only place that knows it all
cmd/mihan          run, repl, check, compile, test, fmt, new, install, lsp
cmd/mihan-ls       the language server as a program of its own
extension/         the editor extension for VS Code
```

Two rules about the tables, because they are what keeps this from turning into
hard-coded spaghetti:

1. **A new keyword is one row of `token.Keywords`.** The lexer, the parser, the
   formatter, the REPL's completion and the reference documentation all read
   that table. Nothing else needs to learn the word. The table refuses to start
   up if two rows claim the same spelling, so a clash cannot be introduced
   quietly.
2. **A new builtin is one row of `eval.builtins`.** The interpreter installs
   from it and `BuiltinDocs` reads it for the documentation.

A third rule came later and is the same argument: **the editor reads those tables
too.** `internal/lsp` asks `eval.Builtins` for the builtins and `token.Keywords`
for the words, so a word added to the language appears in the editor at the same
moment with the same one line of help. There is no keyword list in the extension,
in the reference documentation, or anywhere else, because a second list is a
second thing to forget.

Precedence also lives in the table. `parseBinary` is a plain precedence-climbing
loop over `token.Precedence`, so reordering the operators in `table.go` reorders
the grammar for free.

## The test layers, and what each one is for

| Layer | Where | What it protects |
|---|---|---|
| Lexer | `internal/lexer/lexer_test.go` | the token stream, indentation, interpolation, real source positions |
| Parser | `internal/parser/parser_test.go` | the tree, operator precedence, every keyword, error recovery |
| Interpreter | `internal/eval/eval_test.go` | what each program prints, and that a mistake is named properly |
| Web | `internal/eval/server_test.go` | that a server answers a real request over a real socket |
| Analysis | `internal/analysis/analysis_test.go` | what a name means, and which of the same spelling it means |
| Language server | `internal/lsp/*_test.go` | what an editor is told, over a real JSON-RPC connection |
| Instructions | `internal/ir/ir_test.go`, `corpus_test.go` | that every documented program lowers, and every jump lands somewhere real |
| Documentation | `docs/tutorial/*.mhn`, `tests/`, `examples/` | that every promise the docs and examples make is kept |

The documentation is the one that matters most. **Every `.mhn` file under
`docs/`, `tests/` and `examples/` is a real program that really runs.** It says
what it should print with a `#khoroji` block (one `#|` line per line of expected
output), and `mihan test` runs it. A page whose example has stopped working fails
the build, so a reader never sees a promise the language does not keep.

Two things the block format has to cope with, both learned the hard way. A line
may contain `*`, which stands for any text, because a server prints the port the
machine chose and a test that could not say so would be wrong the moment somebody
else ran it. And a program that answers web requests does not finish, so it is run
as a process of its own, given a moment, and then ended and reported as *serves* —
with whatever it wrote before then still checked. A test runner that could not
handle a server would be saying that a whole class of programs may not be tested.

When you add a feature, add a line to `docs/tutorial/` in the same change. If
you cannot express it in a documented example, that is worth knowing.

The one thing that cannot be a `.mhn` file is the toolchain — installing, and
using an editor — because those are not programs Mihan can run. Those are
documented in `README.md` and `extension/README.md` instead, and they are checked
by `cmd/mihan/lsp_test.go`, which builds the real program and talks to it over a
pipe the way an editor does.

## Writing diagnostics

Every message is bilingual: a Persian half that makes the fix obvious to a
Persian speaker, and an English half that a search engine or a C++ developer
will match against. Both come from one call, so they cannot drift:

```go
in.faultAt(node, "E4133",
    "نمی‌شود بر صفر تقسیم کرد", "division by zero has no answer")
```

`faultAt` takes the same arguments for both halves. When the Persian sentence
needs a value the English one does not, use `faultFA`, which takes the two
argument lists separately rather than forcing a pointless `%s` into the Persian.

Give every diagnostic a code in the `E` followed by four digits, grouped by the
phase that found it: `E1xxx` the lexer, `E2xxx` the parser, `E4xxx` the
interpreter, `E5xxx` a program that stopped. A code is part of the interface; do
not reuse one for a different problem.

## The things that are decided, and why

These were argued about, and the answer is settled. Changing one is a language
change, not a refactor.

- **Colons are optional.** `agar x > 10` opens a block; the colon is accepted
  and ignored. Mihan has no single-line block keyword, so a colon earns nothing.
- **`/` always gives a fraction.** `7 / 2` is 3.5. `ghesmat` (and `ghesmat(7,2)`)
  is how you ask for the whole part. Nobody should be surprised by 7 / 2.
- **`bargard` gives back, `portab` throws.** Returning a `khata` and throwing one
  are different acts, and a beginner who mixes them up gets one clear error that
  names both words.
- **A builtin takes brackets or not, but never both ways ambiguously.**
  `tule xs` and `tule(xs)` are the same call. A *method* — a name that works on
  the value in front of it — is called with a dot, as in `"salam".gooshband()`.
- **Whitespace decides one thing.** `takfir-3` is a call with a negative number;
  `takfir - 3` is a subtraction. The space the programmer typed is what gets to
  say which was meant, so do not "normalise" it away.
- **A bracket-free call's arguments run to the end of the line or the next
  comma.** So `chap_khat 1 + 2 * 3` prints 7, and `chap_khat tule xs + 1` means
  `chap_khat(tule(xs + 1))`. Brackets say the other thing when that is what you
  meant. This is the same rule a shell uses when it splits a line into words.
- **Declaring and assigning are the same act.** `x = 1` introduces the name;
  `x += 1` changes one that exists. That is why the parser only treats `=` and
  `:=` as declarations, and why `sabit` cannot be given a second value.
- **A dict keeps the order its keys arrived in.** Printing the same dict twice
  gives the same answer every run, which is what a beginner needs to be able to
  trust what they see.

## Where the work goes next

The order matters, because each step depends on the one before it.

1. **Benchmarks** (`benchmarks/`). Without numbers, "as fast as C" is a wish.
   Done. They found the loop-scope problem, the text problem, and proved that
   what is left is `interface{}` boxing.
2. **A stable IR** in `internal/ir`, designed for a native and a WebAssembly
   backend. Done. `mihan compile x.mhn` prints the instructions, and
   `internal/ir/corpus_test.go` lowers every documented program so the
   instructions can be read rather than trusted.
3. **A bytecode virtual machine** over that IR, with a value that is a slot on a
   stack rather than a boxed interface. This is the next piece, and it is what
   the profile says the remaining cost is.
4. **A static type checker** over the existing AST. `arz x = 5` already parses
   and already zeroes correctly; what is missing is checking, not syntax. It can
   hang off `internal/ir` now that names are resolved to numbers, and off
   `internal/analysis` now that names are resolved to where they were written.
5. **The toolchain** — `mihan install`, the language server, the editor
   extension. Done. This is the part that decides whether the language is usable
   at all, and it works the way Go's does: one download, one install, and then
   the editor finds the rest by name.
6. **An HTTP server in the standard library**, so a Mihan program can answer a
   request without leaving the language. This is the missing half of "write a
   website", and it belongs in the language's own table of builtins rather than in
   a separate framework. Done: `serva_bgo`, `serva_khod`, `serva_band`, and
   `baza_bgo` for a window of the program's own. `examples/site/khane.mhn` is a
   website and `examples/windows/counter.mhn` is a windowed program, and both are
   checked by the build.
7. **Windows programs** on WebView2. Done as far as the runtime allows: `baza_bgo`
   opens a real chromeless window with the WebView2 runtime that is already
   installed. Embedding it with COM needs cgo, and this Go has `CGO_ENABLED=0`
   and no C compiler, so the window is the runtime opened in application mode
   rather than hosted in the process. That is a real window and a real
   application; it is not the same as hosting WebView2 in-process, and the
   difference is recorded here rather than hidden.

### What is not done, and why

**The bytecode virtual machine.** `internal/ir` is written, tested against every
documented program, and printable with `mihan compile`, so the boundary a machine
will run behind exists and can be read. The machine itself is not written. The
attempt was started and abandoned part-way, because half a machine is a package
that does not compile and a `.\build.ps1` that fails, which is worse than no
machine at all. What the profile says is waiting for it is unchanged: after the
loop scopes were fixed, one hundred per cent of the remaining allocations are a
number being put in a heap box, and only a value that is a slot on a stack rather
than an `any` removes them.

### The machine's value layer, and the four things it found

`internal/vm/value.go` and `internal/vm/arith.go` hold a value that is a struct
with a tag and unboxed fields rather than an `any`, and the arithmetic that works
on it. It is done, it is tested, and the tests are the interesting part: they ask
**the language itself** — by running a program through the interpreter and reading
what it printed — rather than asking a function standing next to the machine,
because two copies of the same code agreeing proves nothing.

Asking the language that way found four things, none of which were written down
anywhere, and all of which a machine written by someone who trusted their own
arithmetic would have got wrong:

- **`%` follows the divisor, not the dividend.** Go gives `-3 % 7 == -3`; Mihan
  gives `1`, because a remainder runs the way the division did. The machine uses
  `flooredMod`.
- **A whole `**` answer stays whole.** `2 ** 10` is `1024`, not `1024.0`.
- **A whole fraction prints with its mark.** `6 / 3` prints `2.0` and never `2`,
  so the machine's text of a number is the runtime's, not its own.
- **`true` is not `1`.** `true == 1` is false. A truth is only equal to a truth.

The last one is the kind of difference that a hand-written table would have got
wrong in whichever direction the author happened to believe, and the first three
would each have been a visible change in what every program printed. This is the
argument for the differential test being part of the machine, not a nicety: **a
machine written by someone who assumes their arithmetic is the language's
arithmetic is a second language wearing the first one's name.**

The claim the whole machine rests on is now measured rather than hoped for:
`TestAMillionArithmeticOperationsAllocateNothing` counts allocations for a million
multiplies, adds and subtracts and the count is **zero**, where the interpreter
needs one box per operation.

Also added: `ir.Func.UpvalNames`, so a closure can be built from the IR alone
instead of from a tree walk, and `Ir.UpvalueGet`/`Set` carry the name-pool index
in `B`.

### The dispatch loop, and what it actually bought

`internal/vm/machine.go` reads one instruction at a time out of an array. It has
the frames, the operand stack, the flat slot stack that a call's locals are carved
out of, and the name-to-slot table that decides once where a name lives and then
remembers. It runs loops, arithmetic, conditions, jumps, text, lists, dicts,
ranges, and the language's own words. It is tested by `runOnBoth`, which runs a
program through the interpreter **and** the machine and fails if what they printed
is not identical — not by asking two functions standing next to each other whether
they agree, which proves nothing.

**The allocation claim is met, and by a very large margin.** For a loop of twenty
million iterations:

| | allocations | memory |
|---|---|---|
| the interpreter | 79,999,935 | 640 MB |
| the machine | **19** | 78 KB |

Nineteen allocations for the whole program, and all nineteen happen before the
first instruction runs. A program that does nothing but count no longer grows.

**The speed claim is met, and the margin is real but smaller than one run suggested.**
The machine is **1.2 to 1.5 times faster** than the tree walker, depending on the run:

| 2-million sum | time | allocations |
|---|---|---|
| the interpreter | 0.276 s | 3,999,996 |
| **the machine** | **0.185 s** | **20** |

| 20-million loop | time | allocations | memory |
|---|---|---|---|
| the interpreter | 4.56 s | 79,999,912 | 640 MB |
| **the machine** | **3.85 s** | **20** | **65 KB** |

The sum is the number to trust and the loop is not: across runs the tree walker
measured anywhere from 4.5 s to 5.5 s on the same machine while the machine held
within 3.7–4.0 s. One run of a big loop is a number, not a measurement, which is
why `runOnBoth` and the allocation test exist as well. Twenty million iterations
for twenty allocations is the part that does not move.

Getting there took three changes, and the first one made it *slower*, which is
the part worth remembering:Getting there took three changes, and the first one made it *slower*, which is
the part worth remembering:

1. **The frame and the program counter moved into the machine.** Every instruction
   had been reading itself out of `m.frames[len(m.frames)-1].fn.Code[pc]` — three
   pointer hops before any work could start. The loop now reads
   `m.code[m.pc]` out of the machine it is already in.
2. **`Value` went from seven words to six.** A fraction's bits now sit in the same
   eight bytes as a whole number instead of in a field of their own, so a push or a
   pop copies 48 bytes instead of 56 — and a loop of arithmetic copies them several
   times a turn. Reading the bits back cannot change the number: not a negative
   zero, not a number too large to be whole, both of which a trip through `int64`
   would quietly turn into something else. `TestAValueFitsInSixWords` now fails if
   a field is added, because a value that grows by eight bytes slows down every
   arithmetic program in the language and nothing else would say so.
3. **The arithmetic writes its answer over the top two stack slots**, rather than
   popping twice and pushing once.

`+` joining text with anything was the fifth thing the differential test found, and
it is the same lesson as the other four: **a machine written by someone who assumes
their own arithmetic is the language's arithmetic is a second language wearing the
first one's name.**

One thing found while doing this, and worth keeping: `m.builts` is a table of what
each name means, worked out once and kept. A builtin called inside a loop used to
rebuild the same value on every turn. That cost was invisible in a benchmark that
only prints at the end — which is exactly why the allocation claim needs its own
test rather than being read off a timer.
**A fifth thing the differential test found.** `+` joins text with *anything* under
the language's rules, not only with other text: `"even " + 3` is `"even 3"`, and
the check has to happen before the numbers are pulled out of the values, or the
text is refused before the rule is ever read. The machine refused it for a while
and a test caught it. Five out of five, and every one of them would have been a
visible difference in what a program printed.
### Functions and closures, and the box a captured name needs

Functions and closures run. A call pushes a frame, the arguments are written into
that frame's own slots, and a return gives them back — 200,000 calls in a row leave
the machine's slot stack where it started, which is what a tree walker cannot say.

A captured name is held as **a pointer to a slot, not a copy of what is in the
slot**, and that is the whole difference between a closure and a note about one. A
closure that counts has to move the number its caller can also see; a copy would
leave the caller looking at the value from before the closure existed. The
interpreter gets this by handing a function its whole `Env`, so the machine has to
work the same way: a local that something captures is given a box, the first time
anything wants to hold on to it, and both the closure and the function that made it
read and write through that box.

Four things this needed that were not obvious:

- **The instructions now carry the name of every local slot** (`ir.Func.LocalNames`).
  `UpvalNames` said what a function wanted; a machine could build the closure but had
  no way to find *which slot of an enclosing function* a name meant, and a closure
  that cannot be connected is just a list of copies.
- **The call boundary must not box.** `popN` was handing back the runtime's values,
  which silently turned a machine's function into the interpreter's kind — a tree
  with an environment and no instructions. Passing a function to another function is
  ordinary code, not an edge case, so the conversion now happens at the builtins,
  which are written in Go and cannot read a machine value, and nowhere else.
- **A call that went into a function pushes nothing.** It used to push a placeholder,
  which sat between the caller and the answer it was waiting for, so the *next* call
  out of that caller picked up the placeholder instead of the thing it meant to call.
  A second call out of a function failed for this reason and nothing else.
- **A local access costs one nil check.** A frame's boxes are a nil slice until
  something is captured, so a program that uses no closures — which is nearly every
  program, and every benchmark shape — pays one predictable branch and nothing more.
  `TestAFrameIsSmall` holds the frame at 80 bytes, because there is one of these per
  call and a recursive function makes one per turn.

**Not done: a machine function inside a collection.** A list is a runtime value, so
a machine's function put into one becomes the interpreter's kind and has no
instructions. Taking it out again gives a value the machine cannot run, and it
**refuses** rather than guessing. `TestAFunctionInACollectionIsRefusedNotGuessed`
says so, and it changes from "refused" to "agrees" when `rt` learns to hold a
machine's function — which is the right fix, since `ir` does not import `rt`.
### What the machine still has to be taught

It runs loops, arithmetic, conditions, jumps, text, lists, dicts, ranges,
functions, closures and the words in its own table. It does **not** yet run classes,
`biar` (import), the exception handler (`koshish`/`giroft`), string interpolation,
slicing, or the other 90-odd builtins. Every one of those is **reported by name**
rather than guessed, and `TestAnOperationStillMissingIsRefusedNotGuessed` holds that
line: a program needing a class gets `this machine does not do DefineClass yet`, and
prints nothing. That test is what will tell us the machine is finished — when it
fails because the machine *can* do it, the feature is in.

**Loading a file from a program.** Done. `biar "rules.mhn"` runs that file and its
top-level names become the program's names; `biar "rules.mhn" as r` puts them behind
`r.` instead, so two files may both declare the same word without either one winning.
Four things the first version got wrong, each found by a test rather than by reading:

- **An imported file has to keep its names.** `RunFile` puts a program's top level in a
  scope that is thrown away when the run ends, because a program has nothing left to do
  afterwards. An imported file has a caller who still wants its names, so it is run
  through `RunTopLevel`, which is the same call the line-at-a-time reader needs and for
  the same reason.
- **A circle has to be stopped, and named.** Two files that import each other would run
  for ever saying nothing, which is the worst kind of mistake to hand a reader. The
  chain is kept and the whole thing is printed — `a.mhn ← b.mhn ← a.mhn` — so a reader
  can see which file to open. That same chain is what makes a file wanted by several
  others run once rather than several times, which is why it is a stack and not a set.
- **The state belongs to a program, not to the package.** The test runner runs many
  programs in one process and the reader runs many more; package-level state would let
  one program's imports be another program's answer. It is a field on the interpreter
  and an imported file inherits it by sharing, which is also how a circle is found at
  any depth.
- **`as` gives an object, not a dict.** A dict has no `r.name`, and a module is a thing
  with named parts — which is what the language already means by an object. So
  `h.jam(2, 3)` calls a function the way it calls a method and `h.benevis` reads a value
  the way it reads a field, with no second rule for the reader to learn.

`examples/windows/counter.mhn` now reads its rules from `shomar.mhn` with `biar`, which
is the duplication this file used to have to apologise for. The checking moved to
`azmoon.mhn`, because a file that is brought into another program *runs*, and a demo at
the top level of a module would happen once per importer.

**Command-line arguments.** A script cannot read them, so
`examples/scripts/hesab.mhn` writes its numbers in the file and says so.
`arghoman()` gives them as text, and `adad_karda` is the one place that turns one
into a number.

### The test runner, and the two ways it could take the machine down

`mihan test` runs each documented program as a process of its own, gives it a
moment, and ends it. Two things about that were wrong, and both were found by
running the tests rather than by reading them.

**The test binary is not the compiler.** `cmd/mihan/args_test.go` used to ask
`os.Executable()` for the program to run the compiler with. Under `go test` that
is the test binary, and a Go test binary given words like `run x.mhn` does not run
them — the testing package ignores leftover arguments and runs the whole suite
again. Every copy did it too, so one `go test ./cmd/mihan` became a tree of
processes until the machine ran out of memory. New files that never spawn anything
kept passing while the suite as a whole was un-runnable, which is why the tests
here now **build the real compiler once** (`compiler_test.go`) and hand the path
around. A test must never exec itself.

**A killed program is not a killed window.** A program that opened a window leaves
it running, because the window belongs to the WebView2 runtime and not to the
program. Killing the one process left a browser behind every run, and a day of
running the tests left a machine full of windows nobody opened. `killTree` ends
the whole tree with `taskkill /T` on Windows and a plain kill elsewhere, so the
window goes when the program does.

### The one rule the profile taught

Mihan's remaining slowness is not in the tree, the lexer or the parser. It is
`type Value any`, which puts every number in a heap box. Measure before
believing that, though: every optimisation in this list was found by looking at
where the allocations actually were, and two of the three were somewhere no one
had guessed. `internal/runtime.NewEnv` was 47% of them.

### The machine learns to interpolate and to slice, and the six things that found

`$"..."` and `xs[lo:hi]` now run on the machine, and so do the operators written
as words. Teaching it those found six differences between the two engines, and
every one of them was a program that answered wrongly rather than a program that
stopped — which is the class of mistake a differential test exists for, because
nothing else can see it.

- **A keyword's kind is its Persian spelling.** The lexer writes `dar`, `na` and
  `ghesmat` into the token's kind, but the lowering's operator table knew only
  `in` and `not` — and `ghesmat` was in no table at all. So `2 dar s` and
  `7 ghesmat 2` lowered to a **Nop**: the machine pushed a value and computed
  nothing, and the stack was one value short for the rest of the program. A word
  the lowering does not know must never become a Nop; `TestOperatorWordsLowerToRealInstructions`
  now walks each word and insists on the instruction it has to become.
- **A dict's operand was half a dict.** `OpMakeDict`'s operand is the number of
  *entries*, and an entry is two things on the stack, so the count had to be
  doubled. Popping only A took `"b", 2` out of `{"a": 1, "b": 2}` and left the
  rest as garbage. This was invisible until the next one was fixed: while every
  dict was empty, no dict test could see that it was also short.
- **A machine value is not a runtime key.** `Dict.Set` and `Set.Add` hash their
  keys, and `rt.HashKey` has never seen the machine's `Value` struct, so it
  returned "not a key" and dropped every one. `{"a": 1}` built an empty dict and
  `{1, 2, 3}` lost every member. The collections are the interpreter's, so the
  values going into them have to cross into the interpreter's shape at that
  boundary — which is now written down where it happens.
- **The interpreter crashed on a slice past the end.** `[1, 2, 3][5:]` reached
  the Go slice with `start=5, end=3` and panicked, which is not a Mihan mistake
  and so was not caught by the recover that turns `portab` into a message. The
  start is now clamped the way the end already was, and a bound past the end is an
  empty answer rather than a crash.
- **`dar` accepted fewer things on the machine.** The interpreter asks a list, a
  set, a dict, a range or text; the machine only knew list, dict and text-that-
  was-text. `2 dar s` refused with "you cannot look for an item" while the
  interpreter said `dorost`. A machine that accepts fewer values is not a faster
  language, it is a smaller one.
- **A range value and a loop's run are different things.** A run of numbers that
  a program holds (`r = 1..6`) is a value that can be sliced; the iterator a `har`
  loop walks is not. Slicing a run copies it before walking it, so the loop that
  walks the same range afterwards still sees all of it.

The last one is the shape of the whole exercise: teaching the machine one feature
is what makes the next difference visible, and the fix is never to make the machine
guess. It is to find out what the language does and do exactly that.

