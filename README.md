# Mihan Language

فارسی: [README.fa.md](README.fa.md)

> Repository: `Mihan-Language` · by Mihani Circle · MIT licensed

A programming language whose keywords are Persian words written in Latin letters,
so that a program reads the way a Persian speaker talks. Today `mihan run` is
an interpreter that compiles functs, methods, lambdas and top-level loops to Go
closures (near Python's speed, faster on start-up, loops and HTTP; see the
table below), and a beginner can learn it by reading its documentation once through.

```
chap "salam, donya!"
chap $"Miyan 2 + 2 = {2 + 2} ast."
```

That is a whole program. There is no `main`, no class, and nothing to set up.

Documentation and an in-browser playground: **[docs.mihan.mihani-circle.ir](https://docs.mihan.mihani-circle.ir)**

## Download

| File | For |
|---|---|
| [mihan.exe](https://github.com/SSNamahsos/Mihan-Language/releases/latest/download/mihan.exe) | Windows (64-bit) |
| [mihan-linux-amd64](https://github.com/SSNamahsos/Mihan-Language/releases/latest/download/mihan-linux-amd64) | Linux (64-bit) |
| [mihan-0.7.8.vsix](https://github.com/SSNamahsos/Mihan-Language/releases/latest/download/mihan-0.7.8.vsix) | VS Code extension |
- VS Code store: [Mihan Language on the Marketplace](https://marketplace.visualstudio.com/items?itemName=MihaniCircle.mihan-language) (or search "Mihan Language" in VS Code Extensions)

1. Put `mihan.exe` in a folder that is on your `PATH` (or run `mihan.exe install`,
   which does it for you). On Linux, rename the file to `mihan`, run
   `chmod +x mihan`, and move it into a folder on your `PATH`.
2. Open a new terminal and run `mihan version`.
3. Install the VS Code extension: *Extensions → … → Install from VSIX*, or
   `code --install-extension mihan-0.7.8.vsix`.

## Why it looks the way it looks

**Every keyword is a Persian word, and every keyword is also an English word.**

`agar` is `if`. `warna` is `else`. `ta` is `while`. `har` is `for`. `chap` is
`print`. `bargard` is `return`. `darghan` is `class`. A Persian speaker reads the
first column and an English speaker reads the second, and neither has to learn a
third thing.

```mihan
naam = "Mihan"

agar naam == "Mihan"
    chap $"salam {naam}!"
warna
    chap " salam nemikonam"
```

## Compared to other Persian languages

Mihan is not the first programming language for Persian speakers. These are
real projects we know of (described from their own pages, October 2026):

| Project | Script | Built on | What it does |
|---|---|---|---|
| [Mr. Hashemi](https://github.com/mr-hashemi/mr-hashemi) | Finglish (`bebin azinja`) | Java, a fork of GraalVM SimpleLanguage | C-like syntax, `.hashem` files, has a VS Code extension; last updated 2022 |
| [Salam](https://github.com/SalamLang/Salam) | Persian (Arabic planned) | C | scripting language for Persian/Arabic speakers |
| [XorLang](https://github.com/Mr-Ali-Jafari/xorlang) | Persian | compiles to C | typed, with HTTP/network modules |
| [FarsiScript](https://pypi.org/project/farsiscript-lang/) | Persian | Python | interpreter plus compile-to-C |
| [ParsLang](https://github.com/SAhmadrezaAnaami/ParsLang) | Persian and English keywords | Python | educational interpreter |
| [Modhej Script](https://github.com/mohsenmodhej/modhej-script) | Latin syntax | Python | Python-powered scripting with web support |
| [Persian-Compiler](https://github.com/mahi97/Persian-Compiler) | Persian | Flex/Bison → C | compiler course project |
| [ParsPL](https://github.com/yoosofan/parsPL) | Persian, right-to-left | Python (PLY) | aimed at school students |

What Mihan tries to do differently:

* **Finglish with English aliases.** Keywords are Persian words in Latin
  letters, so they type on any keyboard with no right-to-left mixing, and every
  keyword also has an English spelling. (Mr. Hashemi is also Finglish; it uses a
  C-like syntax, while Mihan reads like Python, with indentation.)
* **A Go implementation.** One static binary, no runtime to install, fast
  start-up; a closure compiler for speed (see the benchmarks).
* **`mihan build`:** one `.exe`/Linux binary that carries the program, and
  Android APKs.
* **Sites built in:** servers, JSON APIs, pages, sessions, uploads, live events,
  and `mihan serve` to open them in the browser.
* **Editor support:** a language server (completion, mistakes as you type,
  go-to-definition) with a VS Code extension.
* **A 17-lesson tutorial that is tested:** every sample in it is run by
  `tools/tutorial_check.py`.

Mihan is young, and several of the projects above are older or have more users.
This table is about how the projects differ, not a claim that Mihan is better
at everything.

## Install and run

Download the compiler, or build it from source with Go 1.26 or newer:

```
git clone <this repository>
cd CodeLanguage
.\build.ps1
.\build.ps1 run docs\tutorial\01-salam.mhn
```

| Command | What it does |
|---|---|
| `mihan run x.mhn` | run a program |
| `mihan run .` | run the folder, using its `main.mhn` |
| `mihan repl` | try a line at a time, with the names you already made |
| `mihan check x.mhn` | look for mistakes without running anything |
| `mihan test` | run every documented example as a test |
| `mihan fmt x.mhn` | put a file in Mihan's usual shape |
| `mihan new myapp` | start a new program in its own folder |
| `mihan build app.mhn` | make one file (Windows `.exe` or Linux binary, `--os`, `--arch`, `--window`, `--open-browser`) that carries the program |
| `mihan serve site.mhn` | run a site and open it in the default browser (same as `mihan run --open`) |
| `mihan doc` | the whole language reference, printed |

## Learning Mihan, in order

**Start with the course: [`docs/tutorial/README.md`](docs/tutorial/README.md).**
Seventeen lessons, from the first `chap_khat` to a JSON backend, a frontend, a
Windows app, an Android app and a full web API, each with runnable examples, their real output, and exercises with
solutions. `python3 tools/tutorial_check.py` re-runs every sample.

Three complete, good-looking websites built with Mihan are in
[`examples/pro-sites/`](examples/pro-sites/README.md).

The one-page reference programs below are the original short version:

`docs/tutorial/` is a course, and every page is a real program that runs. Read
them in order and run each one — the page is the lesson and the program is the
proof that the lesson is true:

| Page | What it teaches | Run it |
|---|---|---|
| `01-salam.mhn` | a whole program, and `$"…{…}…"` | `mihan run docs/tutorial/01-salam.mhn` |
| `02-arzhad.mhn` | numbers, text, `dorost`, `hich`, and typed names | …`/02-arzhad.mhn` |
| `03-matn.mhn` | text with holes, and the methods a text has | …`/03-matn.mhn` |
| `04-yeh.mhn` | lists, dicts, sets, and the builtins that work on them | …`/04-yeh.mhn` |
| `05-agar-va-ta.mhn` | `agar`, `ta`, `har`, `shek`, `edame`, and `ke` | …`/05-agar-va-ta.mhn` |
| `06-funct.mhn` | functions, defaults, recursion, and functions as values | …`/06-funct.mhn` |
| `07-darghan.mhn` | classes, `khod`, constructors and inheritance | …`/07-darghan.mhn` |
| `08-khata.mhn` | `portab`, `giroft`, `nemigir`, and what a `khata` carries | …`/08-khata.mhn` |
| `09-ayaz.mhn` | saying what your program should print, and checking it | …`/09-ayaz.mhn` |
| `10-dastoor-ha.mhn` | what a program becomes before it runs | …`/10-dastoor-ha.mhn` |
| `11-json.mhn` | JSON, and how a file is named | …`/11-json.mhn` |
| `12-safhe.mhn` | the shape of an answer a site gives back | …`/12-safhe.mhn` |
| `13-rang.mhn` | a page written with `rang` and `khat_dor` | …`/13-rang.mhn` |
| `14-rah.mhn` | pieces, buttons and which page answers a path | …`/14-rah.mhn` |

Three habits make this much faster:

- **Read a page, then change it.** The files are meant to be edited. Change a
  number, break something on purpose, and see what the message says.
- **Try lines in `mihan repl`.** It remembers what you typed, so you can build a
  value up a line at a time. `2 + 2` prints its answer without `chap`.
- **Check your own programs.** Put a `#khoroji` block at the top of a file with
  the lines you expect, and `mihan test yourfile.mhn` will tell you whether it
  still prints them. That is how the tutorial checks itself, and page 9 shows
  the shape.

```
> mihan repl
Mihan 0.7.8

  baraye khorod: mihan > 2 + 2
  baraye khanodan: chap "salam"
  baraye khoroj: exit

mihan> 2 + 2
4
mihan> xs = [3, 1, 2]
xs = [3, 1, 2]
mihan> taghib xs
[1, 2, 3]
mihan> exit
khodahafzar!
```

`mihan doc` prints the whole language in one place, every word and every builtin,
each with the one line that says what it does. It is written from the same table
the compiler runs from, so it cannot be out of date.

## What is in the language today

- **Values.** Whole numbers, numbers with a fraction, text, `dorost`/`ghalat`,
  `hich`, lists, dicts, sets and ranges.
- **Text.** `$"salam {name}"` (with the `$` in front) weaves values into a sentence. Methods with a dot:
  `"salam".gooshband()`, `"a,b".tab(",")`, `"salam".shamil("ala")`.
- **Choosing and repeating.** `agar`/`warna`/`warna agar`, `ta`,
  `har i az 0 ta 10`, `har i az 0 ta 10+`, `har x dar xs`, and a `ke` tail that
  keeps only the items that pass a test.
- **Functions.** `funct add(a, b) -> arz`, defaults, named arguments, recursion,
  closures, and `funct(x) -> x > 2` written where a value is expected.
- **Classes.** `darghan`, `sakht`, `shenase`, `dastoor`, inheritance, `khod`
  for the current object.
- **Failures.** `portab` throws, `giroft` catches, `nemigir` always runs, and a
  `khata` carries a message and any data you want it to.
- **Files and the world around the program.** `bekhan` reads a file, `benevis`
  writes one, `fehrest` lists a folder, `mohit` reads an environment variable,
  and `darkhast` asks another server. `json_bekhan` and `json_benevis` turn
  JSON into ordinary values and back, which is how a page talks to the program.
- **A site in pieces.** `onvan`, `karte`, `navar`, `dokme`, `jadval`, `form`
  and `ofoghi` / `amudi` are the pieces. `rah` picks the page for the path.
  A `dokme` carries a `kar`; pressing it tells the program, and the page comes
  back with the new answer already in it. No HTML, no stylesheet and no script
  is written in the program. `examples/site/barname/barname.mhn` is three pages
  of one program. `rang` and `khat_dor` are still how a piece looks.
- **A site, and a window of the same program.** `serva_bgo` answers requests.
  `parvande_pasokh` sends a file from a folder beside the program, so the
  frontend is HTML, CSS and JS and the backend is the funct that answers
  `/api`. `baza_bgo` opens that same address as a window.
  `examples/scripts/yaddasht.mhn` is a script. `examples/site/hesab` is a
  calculator, `examples/site/yaddasht` keeps notes, and `examples/site/dokan`
  is a shop. `examples/windows/counter.mhn` is a window.
- **Builtins**, listed by `mihan doc`.
- **Speed.** Best of 5 runs of `benchmarks/compare` (whole process, seconds,
  Linux amd64; `python3 benchmarks/compare/bench.py mihan`):

  | test | Mihan 0.6.0 | Python 3 | Go |
  |---|---|---|---|
  | fib(30) | 0.117 | 0.094 | 0.006 |
  | 5M-step loop | 0.209 | 0.303 | 0.003 |
  | dict+list+sort | 0.193 | 0.129 | 0.046 |
  | string building | 0.092 | 0.637 | 0.042 |
  | JSON write+read | 0.167 | 0.107 | 0.114 |
  | file write+read | 0.078 | 0.058 | 0.038 |
  | objects + lambdas | 0.168 | 0.110 | 0.004 |
  | start-up | 0.003 | 0.012 | 0.001 |
  | HTTP requests/s | 68 256 | 10 003 | 86 614 |

  Mihan wins on start-up, HTTP and plain loops. Its string-building win holds
  only against top-level Python code; inside a function Python takes 0.05 s.
  Python wins on the rest by 1.2–1.6×, and Go is far ahead on computation.

- **Android.** `mihan build app.mhn --os android` writes a signed APK
  (lesson 15). `mihan preview app.mhn --android` shows the app in a phone
  frame on your computer, with back/home/rotate and a live log (lesson 17).
  Pages get `mihan.toast`, `mihan.vibrate`, `mihan.share` and `mihan.onBack`.

- **Web APIs.** Routes with `:id` and `*`, method filters, `json_pasokh`,
  CORS, cookies, sessions (`jalase`), uploads, ETag caching, `qaleb`
  templates and live server-sent events (lesson 16,
  [`examples/site/namayesh/`](examples/site/namayesh/namayesh.mhn)).

- **Speed on demand.** `arz x = 5` starts a name at a known type with a known
  zero. Nothing else has to change, and a beginner never has to write one.

## Two engines, one language

There are two ways a Mihan program runs, and they must never disagree about what
a program prints:

- the **interpreter** (`internal/eval`), which reads the tree the parser built.
  This is the one `mihan run` uses, and it is the definition of what the language
  does.
- the **machine** (`internal/vm`), which reads instructions instead. A number
  lives in a stack slot rather than in a box, which is why a loop there allocates
  nothing and runs faster.

Every documented program is run through both, and the test fails if they print
different things — and also fails if the machine prints something the interpreter
does not, which is the kind of mistake a test has to be built to see. The machine
is finished feature by feature, and anything it has not been taught yet it
**refuses by name** rather than guessing, so a reader is never given a wrong
answer. `internal/vm/corpus_test.go` is the list of what it can do today.


## The documentation is executable

Every example in `docs/tutorial/` is a real `.mhn` program that really runs and
says what it should print. `mihan test` runs all of them, so a page whose
example has stopped working fails the build. A reader never sees a promise the
language does not keep.

```
> mihan test
salam  docs\tutorial\01-salam.mhn        kharch
salam  docs\tutorial\02-arzhad.mhn       kharch
...
8/8 sabet. Mihan 0.6.0
```

## Mistakes are written in Finglish

Every message says what went wrong in Finglish, the same Persian-in-Latin-letters
the language itself uses, with a code a search engine can find:

```
khata: E4133
  nemishe bar sefr taghsim kard
   4 | chap 10 / (x - 2)
       ^^^^^^^^^^^^^
```

## Where it is going

1. Benchmarks, so "as fast as C" is a number rather than a wish. Done.
2. A stable intermediate representation, then a bytecode virtual machine behind
   it, then a native backend. The representation and the machine are done as far
   as the language goes: the machine runs every documented program and agrees with
   the interpreter on every one. What is left is the rest of the standard library
   on the fast path, and then a backend that turns the same instructions into a
   program the operating system can run by itself.
3. A static type checker over the existing tree, so `arz x = 5` is checked rather
   than merely zeroed.
4. Windows programs on WebView2, and a site whose page is files and whose
   answers are JSON. The window and the site are the same program: one serves,
   the other opens a window onto it. Done as far as a page, an API and a
   window go. What is not done is hosting the window inside the process.

`AGENTS.md` has the architecture, the testing strategy, and the decisions that
are already settled.



## License

MIT © Mihani Circle. See [LICENSE](LICENSE).
