# Mihan extension 0.7.8 — Finglish CSS

* 111 Finglish style words (`rang`, `zamine`, `gooshe_gerd`, `hashie`, `fasele_dakheli`, `fasele_birooni`, `pahna`, `ertefa`, `ghalam`, `andaze_matn`, `vazn_matn`, `tarazbandi`, `nemayesh`, `jahat`, `chidman`, `hamtarazi`, `sotun_ha`, `mogheiyat`, `laye`, `sarriz`, `saye`, `shafafiat`, `enteghal`, `tabdil`, `harekat`, `nashangar`, ...) and Finglish values (`siah`, `sefid`, `ghermez`, `markaz`, `por`, `hich`, `makhfi`, `charkhesh()`, `bozorgi()`, ...). English CSS names keep working.
* `sabk {...}` builds a style sheet (selectors with piece names, `:rooye`, `@harekat`, `@kochak_tar`, `@kam_harekat`); `sabk_khat {...}` builds one style text; `safhe_kon` takes a sheet as its third argument; pieces take `kelas` and `shenase`.
* Autocomplete offers style words and values (with colours) inside dict strings; style keys are highlighted.
* Error messages are Finglish only. `rang_ya_khata(a, b)` now counts from a up to b.

# Mihan extension 0.7.6 — Run and F5

* F5 / Ctrl+F5 run the open .mhn file through a built-in `mihan` debugger type (no launch.json needed); output goes to the Debug Console, programs that read input (khanak) run in a terminal, sites get `--open`.
* Run button (play) in the editor title, plus editor and explorer menu entries: `Mihan: Run this file` runs it in a terminal named Mihan.
* launch.json snippets: run this file, program with input, website.

# Mihan extension 0.7.5 — smarter autocomplete

* After `.` the list depends on the value in front: matn, list and dict methods
  (Finglish and English names), the fields and methods of your own darghan
  (including `khod.` and base classes), and the names of a `biar "x.mhn" as m` module.
  Chains such as `esm.tab(",").` follow the result type.
* Inside a funct, parameters and locals come first, then globals, then snippets,
  builtins and keywords. Names from imported files are offered too.
* Every builtin and keyword has a Persian and English description and its own kind.
* Block snippets with tab stops: agar, agar…warna, warna agar, har…dar, har…az…ta,
  ta, funct, darghan (with __sakht__), koshish, koshish…nemigir, biar.
* Nothing is offered inside text or notes; block keywords only at the start of a line.
* Signature help for builtins (named parameters), user functs, methods,
  module functs and constructors, across several lines.
* Editor defaults tuned for Tab/Enter acceptance.

# Mihan 0.7.1 — changes

* **Fix:** `barmadar(matn, x)` now looks for a piece of text:
  `barmadar("salam donya", "donya")` is `dorost` (it used to compare against
  single letters only, so any longer text gave `ghalat`). Lists, sets, ranges
  and dicts (keys) work as before. Regression test: `tests/barmadar-matn.mhn`.
* **`mihan run --open site.mhn`** and the new **`mihan serve site.mhn`**: when the
  server is listening, Mihan prints `Mihan site is ready: http://...` (in a
  terminal) and opens the address in the default browser (Windows: the URL
  handler via `rundll32 url.dll,FileProtocolHandler`; Linux: `xdg-open`;
  macOS: `open`).
* **`mihan build site.mhn --open-browser`**: a built `.exe`/binary opens its
  site in the default browser when double-clicked.
* Docs: lessons 4, 10 and 14 and the README updated.
* Checks: `go vet`, `go test ./...`, `mihan test` 33/33, course checker 90/90.

# Mihan 0.7.0 — changes

All checks pass (numbers in the report below are the real ones):
* `go vet` (Linux and Windows) and `gofmt`;
* `go test ./...`, with new tests for the web API, cookies, sessions, uploads,
  static caching, templates, live events and `mihan preview`;
* `go test -race`;
* `mihan test` and the course checker.

## `mihan preview app.mhn --android [--device pixel|small|tablet]`
* Runs the program as the APK does (`MIHAN_ANDROID=1`, the window address
  through `MIHAN_WINDOW_FILE`) and shows it in a phone frame in the system
  browser's app mode: status bar, back, home, rotate, three device sizes,
  restart. A program that only prints is shown as text, as on the phone.
* A live log: app lifecycle (onCreate, onPause, onResume, exit), prints,
  every request with status and time, program and JavaScript errors,
  `console.log`, and every phone feature used.
* **It is not an emulator.** The program runs on the computer and the page is
  drawn by the computer's browser, not Android's WebView. KVM on the build box
  is broken, so no APK was run in an emulator either.

## Android features for pages
* Every HTML page sent by an Android app (or a preview) carries
  `/_mihan/mihan.js`, which gives `window.mihan`: `toast`, `vibrate`, `share`,
  `onBack(f)` and the `mihan-pause` / `mihan-resume` events. The wrapper app
  has a `MihanAndroid` JavaScript bridge (Android toast, `Vibrator`, share
  sheet; the back button asks the page first). In a browser each call falls
  back to something sensible. `android_hast()` tells a program it is on a phone.

## Web
* `rah` routes: `/kala/:id` (into `darakht["params"]`), trailing `*`, and an
  optional `method` that answers 405 otherwise.
* `json_pasokh(x, vaziat)`; `"cors"` in an answer (`dorost` or one origin, with
  credentials); `darakht["cookie"]` and `"cookie"` in an answer; `jalase(darakht)`
  sessions; `darakht["form"]` and `darakht["parvande"]` (urlencoded and
  multipart uploads); dict `hazf` / `delete`.
* Static files: right MIME types, `ETag`, `Last-Modified`, `Cache-Control:
  no-cache`, and `304 Not Modified`; HEAD has no body.
* `qaleb(matn, dict)` templates (escaped `{{x}}`, raw `{{{x}}}`, `{{a.b}}`),
  `html_amn(x)`.
* Live server-sent events: pages listen on `/_mihan/rooydad/<kanal>`, the
  program sends with `rooydad_ferest(kanal, x)`. These are served beside the
  one-at-a-time request lock, so listeners never hold up other requests.
  WebSockets were not added; SSE covers program-to-page news, and pages send
  with `fetch`.
* Showcase: `examples/site/namayesh/` (CSS animations and transitions,
  JavaScript, a JSON API, live events, the phone bridge). Screenshots, frames
  and a gif are in `deliver/screens/`.
* New lessons: 16 (a real web API) and 17 (`mihan preview`).

# Mihan 0.6.0 — changes

All checks pass:
* `go vet` (Linux and Windows) and `gofmt`;
* `go test ./...`, with new tests: compiled-vs-walked programs for methods, try,
  closures, lambdas, block-local functs and `+=` on text; the quick JSON
  reader against encoding/json; and the Android build;
* `go test -race`;
* `mihan test` (31/31) and the course checker (81/81).

## Android: `mihan build app.mhn --os android`
* Makes an installable, signed `.apk`. No Android Studio, SDK, Java or NDK is
  needed: `mihan` writes the APK itself. It renames the package in the compiled
  manifest, keeps `resources.arsc` stored and 4-byte aligned, and signs with
  **APK Signature Scheme v2**. The signing key is a debug key that `mihan` makes
  once and keeps in the user's config folder (`android-debug.pem`), so updates
  install over the old app.
* Inside is a small wrapper app (`android/`, a full-screen WebView, minSdk 24),
  with the program as `lib/<abi>/libmihanapp.so`. That library is the Linux
  runtime for the phone's processor with the program appended, as in a Linux
  build. A program that opens a window (`baza_bgo`) or starts a server is shown
  in the WebView (through `MIHAN_WINDOW_FILE`). A program that only prints has
  its output shown.
* `--arch arm64` (the default), `amd64` (for the emulator) or `all`.
  `--name` sets the label and `--package` the app id.
* New lesson: `docs/tutorial/15-android.md`. The wrapper's source and
  `android/build-template.sh` (which needs the SDK, for Mihan's own developers
  only) are in `android/`.
* **What was tested, honestly:**
  * `apksigner verify` (v2) and `zipalign -c 4` pass, and `aapt2 dump badging`
    reads the package, label, SDK levels and permission.
  * The arm64 program inside the APK ran on the build box under
    `qemu-aarch64`, served its page and counted clicks. The x86_64 one ran
    natively.
  * **The APK has not been run on a phone or in an emulator.** The box's KVM is
    broken (the kernel faults on it), and the Android emulator needs KVM. So
    the Java wrapper has been compiled and dexed but never started.

## Speed
* The closure compiler now also covers:
  * **darghan methods**: `khod` gets a slot. A call `obj.method(...)` on a
    compiled method makes no bound-method object, and the class's method is
    remembered at each call site.
  * **lambdas** (one-line and block);
  * **functs defined inside blocks or inside other functs** (the inner funct is
    compiled, and reads outer names through its scope);
  * **koshish / giroft / nemigir**;
  * conditions of `agar` / `ta` on whole numbers, which no longer build a
    truth value.
* A funct that *contains* a lambda or a nested funct still runs in the tree
  walker, because its locals must be visible to the closure.
* **`t += "..."` is linear**, not quadratic: the place doing it keeps spare
  room after the text it built last, and only adds to it when the text is
  exactly that one (texts made earlier never change).
* `"x" + 5` and text + text take a direct path.
* **Dicts** find text and whole-number keys through typed maps
  (`map[string]` / `map[int64]`), with no key string built. Other keys work as
  before.
* **Sorting** whole numbers compares plain numbers, not interfaces.
* **JSON reading** uses a reader of its own for well-formed JSON. Anything it is
  unsure of goes to encoding/json, so error messages are unchanged.
* Not done: a cheaper value representation (a tagged value instead of Go
  `any`). It would touch every file of the runtime. Small whole numbers are
  still cached, as in 0.5.

### Numbers
Best of 5 runs, whole process, seconds, on the build box (Linux amd64).
`python3 benchmarks/compare/bench.py mihan` makes this table.

| test | Mihan 0.5.0 | **Mihan 0.6.0** | Python 3.13 | Go 1.26 |
|---|---|---|---|---|
| fib(30) | 0.118 | **0.117** | 0.094 | 0.006 |
| 5M-step loop | 0.200 | **0.209** | 0.303 | 0.003 |
| dict+list+sort | 0.260 | **0.193** | 0.129 | 0.046 |
| string building | 1.347 | **0.092** | 0.637 | 0.042 |
| JSON write+read | 0.348 | **0.167** | 0.107 | 0.114 |
| file write+read | 0.087 | **0.078** | 0.058 | 0.038 |
| objects + lambdas | 0.463 | **0.168** | 0.110 | 0.004 |
| start-up (hello) | — | **0.003** | 0.012 | 0.001 |
| HTTP requests/s (higher is better) | — | **68 256** | 10 003 | 86 614 |

The 0.5.0 timings were re-measured on the same box today. They are a little
lower than the numbers in the 0.5.0 notes, because the box was less busy.

Where Mihan wins, honestly:
* It beats Python on start-up (4× faster) and on HTTP (about 7× the requests,
  and about 80% of Go's net/http, because Mihan's server *is* Go's net/http).
  It also beats Python on the 5M-step loop.
* It beats Python on string building only because Python's `t += ...` is slow
  at the top of a file. Put inside a function, the Python version takes
  0.05 s and beats Mihan.
* It loses to Python on fib(30) (1.2×), dict+list+sort (1.5×), JSON (1.6×),
  file IO (1.3×) and objects+lambdas (1.5×).
* Go is 10–60× faster than Mihan on pure computation. Mihan is not close to Go
  there, and will not be without a compiler to machine code.

# Mihan 0.5.0 — changes

`go vet`, `gofmt`, `go test ./...` (+1 new test that runs 19 programs both
compiled and walked and checks that output and error messages match),
`go test -race`, `mihan test` (31/31) and the course checker (81/81) all pass.

## Speed: closure compilation
* Top-level functs and top-level `har`/`ta` loops are compiled once into Go
  closures (`internal/eval/compile.go`). Each local is a numbered slot in its
  call's own array: parameters, names the funct assigns, and `har`/
  comprehension names. Names from the file's scope are cached bindings.
* Call frames come from a per-interpreter free list. Arguments are passed
  without building a scope map.
* Whole numbers from -1024 to 32767 are boxed once and reused. `+ - * % < <= >
  >= == !=` on whole numbers take a direct path with the same 64-bit overflow
  error (E4126).
* Text, list and dict methods (`d.begir`, `xs.bast` ...) and builtins called
  with plain arguments skip the wrapper function.
* Rules are unchanged. Writing to a name the file already has still changes
  that name. Loop names still stay inside the loop. `sabit` is still refused,
  and the depth limit and traces are the same. Mistakes, method lookup and
  conversions all go through the interpreter's own code, so the messages are
  identical.
* Not compiled, and run by the tree walker as before:
  * functs that use `try`, lambdas, nested functs, typed or `sabit` names, `khod`,
    slices or sets;
  * methods of darghan;
  * functs defined inside blocks;
  * anything run with a step limit or hooks (the REPL, the debugger).

Best of 5, seconds (Linux amd64, process start included):

| test | 0.3.0 | 0.4.0 | 0.5.0 | Python 3 |
|---|---|---|---|---|
| fib(30) | 0.76 | 0.73 | 0.14 | 0.10 |
| 5M loop | 0.47 | 0.52 | 0.22 | 0.29 |
| dict+list+sort | 0.47 | 0.46 | 0.26 | 0.15 |
| fib+loop | 1.19 | 1.22 | 0.38 | 0.41 |

Honest summary: about 3–5× faster than 0.4.0. It beats Python on the loop
and on fib+loop, but is still about 1.4× slower on fib(30) and 1.7× slower on
dict+list+sort. A part of the gap is the dictionary and sort runtime, not
the interpreter.

# Mihan 0.4.0 — changes

`go vet`, `gofmt`, `go test ./...` (+3 new tests), `go test -race`, `mihan test`
(31/31) and the course checker (81/81 samples, now 14 lessons) all pass.

## mihan build: Windows and Linux
* `mihan build app.mhn [--os linux|windows] [--arch amd64|arm64] [--window] [-o name] [--base runtime]`.
  The default is the system you are on.
* **Cross builds need no Go.** Runtimes for linux/amd64, linux/arm64,
  windows/amd64 and windows/arm64 ship in a `runtimes/` folder beside `mihan`
  (also looked for in `../runtimes`, in `$MIHAN_RUNTIMES`, or given with `--base`).
  A runtime for the wrong system is refused by name.
* **`--window` hides the console** of a Windows app: it sets the executable's
  subsystem to GUI, the same header field `go build -H windowsgui` sets, so one
  runtime serves both. Without it a Windows build keeps its console (for
  scripts).
* Linux builds unpack to `~/.cache/mihan-apps/<name>/`.
* New lesson: `docs/tutorial/14-build.md`.

## Window apps on Linux
`baza_bgo` opens the window in Chrome, Chromium, Edge or Brave in app mode (or
Firefox, in a window with its own profile), falling back to `xdg-open`.
`MIHAN_BROWSER` picks a browser, and `MIHAN_WINDOW_NO_SANDBOX=1` adds
`--no-sandbox` for containers. Each window gets its own profile folder, so the
program stops when its window closes, on Linux as on Windows (not with
`xdg-open`, which cannot be followed). Tested end to end on the box: a
Linux-built `examples/windows/counter.mhn` opened its Chrome app window,
closing the window (Ctrl+W) ended the program with status 0, and its profile
folder was removed. The window also no longer offers Google Translate.
`mihan test` sets `MIHAN_NO_WINDOW=1`, so testing a window app no longer
opens windows (now that Linux has a real window opener).

## Speed: name resolution, first step
Before a file runs, a resolver finds the names read inside a funct that no
scope between the read and the file's own scope could ever bind (no parameter,
assignment, loop name, catch name or declaration with that spelling anywhere
on the way). Those reads go straight to the file's scope, and the binding found
there is remembered on the name, so later reads are one pointer load. It is
conservative on purpose: when in doubt a name is looked up the old way, and a
file that imports inside a block or funct is not touched. Top-level functions
and classes now live in the file's own scope (before, in a block scope under
it). Tests cover the tricky cases (shadowing, closures, loop-made functs,
comprehensions, catch names, methods).

**Honest result: no measurable gain on these benchmarks** (best of 5, same box):

| test | 0.2.0 | 0.3.0 | 0.4.0 | Python 3.13 |
|---|---|---|---|---|
| fib(30) | 1.39 s | 0.70 s | 0.71 s | 0.10 s |
| 5M-step loop | 0.64 s | 0.46 s | 0.49 s | 0.30 s |
| dict + list + sort | 1.00 s | 0.44 s | 0.47 s | 0.13 s |
| fib(30) + 5M loop | 2.18 s | 1.22 s | 1.25 s | 0.38 s |

The differences between 0.3.0 and 0.4.0 are within run-to-run noise (single
runs ranged ±5%). Finding a name was no longer the expensive part: the time goes
on walking the tree, boxing numbers into interfaces, and making a scope and a
trace frame for every call. Real gains need locals in numbered slots of a
per-call array (no scope maps at all) or a working bytecode VM. That is a larger
change to how scopes work and was not started.

## Known limitations
* Speed (above): still about 2–7× slower than Python.
* A Windows build (console or `--window`) has been built and its header and
  payload checked, but **not run on a real Windows machine**.
* A Linux window app opened with `xdg-open` keeps running after its tab closes.

# Mihan 0.3.0

All of `go vet`, `gofmt`, `go test ./...` (+9 new tests), `go test -race`,
`mihan test` (31/31) and the course checker (81/81 samples) pass.

## Language
1. **`na` / `!` bind tightly, like Python and Go.** `na a ya b` is now
   `(na a) ya b`, and `!dorost || dorost` is `dorost`. A comparison still counts
   as one value: `na x == y` is `na (x == y)`. Lesson 2 and the sites were
   updated (the `(na …)` brackets they needed are gone).
2. **Block scope like Python.** A name made inside `agar`, `har`, `ta` or
   `koshish` is still there after the block (`koshish: f = …` then use `f`).
   Names made in a `funct` still stay in the function, typed/`sabit` names
   (`arz x`, `sabit x`) still belong to their block, and each loop turn still
   has its own loop variable, so a lambda made in a loop keeps its own value.
3. **List comprehensions:** `[x * x har x dar xs]`, `[x har x dar xs ke x > 2]`
   (`agar` works in place of `ke`), `[i har i az 0 ta 10]`.
4. **Integer overflow is a mistake (E4126)**, not a silent wrap-around, for
   `+ - * **`.
5. **Text inside lists, dicts, sets and objects prints with quotes**, as in
   Python: `["a", 1]`, `{"naam": "Sara"}`. Text on its own still prints bare.
6. New builtins: `adad_ya_hich` / `to_number` (the number, or `hich`; pairs with
   `??`), `zaman()` / `now` (seconds as a float), `tarikh(pattern?, moment?)` /
   `date` (`"YYYY-MM-DD HH:mm:ss"`), `sabr s` / `sleep`.
7. A lambda can be passed to a verb without brackets: `bast fs, funct() -> 1`
   (before, only `bast(fs, funct() -> 1)` parsed).
8. Runtime errors now show the failing line under the message (before, only the
   message was printed).

## Tools
* **`mihan build app.mhn [-o App.exe] [--base mihan.exe]`** makes one executable
  holding the runtime and the program's folder (`safhe/`, other `.mhn` files;
  `dade/`, hidden folders and `.exe` files are left out). Run on Windows it makes
  a `.exe`; `--base` picks a different runtime (e.g. build a Windows `.exe` from
  Linux). On first run it unpacks to `%LOCALAPPDATA%\mihan-apps\<name>\` (once
  per build); data the program writes there survives rebuilds.
* **Windows apps stop when their window is closed.** Each program's window gets
  its own Edge profile folder, so the Edge process is the window's own; when it
  ends, the program ends and the profile is removed.
* VS Code extension 0.3.0: highlighting for the new builtins; completion, hover
  and signature help pick them up from the builtin table.

## Speed
The interpreter was profiled and the hot paths fixed: scopes keep a few names in
a small inline array instead of a map, blocks that declare nothing share the
function's scope, positional arguments are no longer wrapped, plain calls bind
parameters without building a map, whole-number arithmetic and comparisons have
a fast path, `taghib` of only numbers or only text uses a faster sort, a dict
method no longer builds a table on every call, and the garbage collector runs
less often (`GOGC=400` unless set).

Best of 5 runs, same Linux box (scripts in `benchmarks/compare/`):

| test | Mihan 0.2.0 | Mihan 0.3.0 | Python 3.13 | Go |
|---|---|---|---|---|
| fib(30) | 1.49 s | 0.76 s | 0.10 s | |
| 5M-step loop | 0.69 s | 0.49 s | 0.31 s | |
| dict + list + sort (300k) | 1.14 s | 0.48 s | 0.15 s | |
| fib(30) + 5M loop | 2.26 s | 1.29 s | 0.40 s | 0.009 s |

So 0.3.0 is **1.4–2.4× faster** than 0.2.0, but still **about 1.5–7× slower than
Python** and far from Go. The bytecode VM (`internal/vm`) was **not** wired into
`mihan run`: measured on its own benchmarks it is only ~25% faster than the old
interpreter, it lacks features the interpreter has (and does not know the new
block scope, comprehensions or overflow check), so switching would have risked
wrong answers for little gain. Real speed needs name resolution to slots at
parse time (no string lookups) — the next step.

## Known limitations
* Speed (above).
* A built `.exe` and the close-the-window behaviour were built and checked on
  Linux (`GOOS=windows` build, payload layout verified, a Linux build runs end to
  end) but **not run on a real Windows machine**. A built app still opens a
  console window next to its app window.
* `chap_khat a, b` joins its values with no space.

# Mihan 0.2.0

Everything below was found by building Mihan, running `go test ./...`,
`go test -race`, `mihan test`, every example, and by writing the new course and
three example sites with it. Every fix has a test (`internal/eval/fixes_test.go`,
`internal/lsp/signature_test.go`, or a checked sample in `docs/tutorial/*.md`).

## Test results

| | before | after |
|---|---|---|
| `go build`, `go vet ./...`, `gofmt -l` | clean | clean |
| `go test ./...` | all pass | all pass (+13 new tests) |
| `go test -race ./...` | **data races** in `internal/eval` (web server) | no races |
| `mihan test` (docs, tests, examples) | **hung forever** on `examples/site/khane.mhn` (intermittent), 26 files | 31/31 pass in ~30 s |
| course samples (`tools/tutorial_check.py`) | did not exist | 76/76 samples print exactly what the lessons show |
| `GOOS=windows go build/vet` | ok | ok |

## Bugs fixed

### Toolchain
1. **`mihan test` could hang forever.** It called `cmd.Wait()` twice on the same
   process (once in a goroutine, once in a defer), read the output buffers while
   the copying goroutines were still writing them (a data race), and had no
   `WaitDelay`, so a child holding the pipe kept the runner waiting. Now: one
   `Wait`, a locked buffer, `WaitDelay`, and on Linux/macOS the program runs in its
   own process group so the whole tree is killed (`cmd/mihan/test.go`,
   `proc_unix.go`, `proc_windows.go`).
2. **Every failed program printed a stray `mihan: ` line** with an empty message
   after the real diagnostic (`cmd/mihan/main.go`).

### Error messages
3. **Runtime errors were not bilingual.** `faultAt` threw the Persian half away
   and every runtime error said only "برنامه در همین‌جا ایستاد"; `faultFA` did the
   opposite and replaced the English half with the Persian one. Runtime errors
   now show the real Persian sentence, the English sentence, and their own code
   (`E4133` instead of a generic `E5001`).
4. **Parser messages printed `%!(EXTRA string=…)`** inside the Persian text when
   the Persian sentence had no placeholder (e.g. E2025).
5. A Persian message had a typo (`بast`).

### Language
6. **`f(1, b = 2)` silently passed `hich`.** Only `b: 2` named an argument; `b = 2`
   was read as an assignment expression. `name = value` now names an argument,
   as in Python.
7. **`tabdil x, arz` — the documented form — failed** ("arz is not a name I
   know"); only `tabdil(x, "arz")` worked. Type words now work as values.
   `tabdil` also learnt English names (`int`, `float`, `str`), `tabdil(7, matn)`
   (was an error), `tabdil("2.5", kasr)` (was an error), and no longer accepts
   `"12abc"` as 12.
8. **`zip` was a documented builtin that did not exist.**
9. **A subclass did not get its parent's field defaults** — `B < A` objects had
   `hich` where `A` said `v = 1`.
10. **List/dict/set field defaults were shared by every object** (`items = []` in a
    `darghan`: adding to one object's list added to all of them). Each object now
    gets its own copy.
11. **`darghan B < A` (without the closing `>`) was a parse error.** The `>` is
    now optional.
12. **`funct f(arz a, arz b)` gave seven errors.** The type-before-name form is
    now accepted, the same as `arz x = 1` everywhere else (`a: arz` still works).
13. **A variable named like a builtin could not be changed with `+=`**
    (`shomar += 1` → parse error) though `shomar = 0` worked.
14. **`a ?? b` silently dropped `b`** (it was read as two `?` marks). It is now
    "a, or b when a is hich", backed by a new builtin `pishfarz` / `coalesce`.
15. **The web server ran Mihan code on several goroutines at once** (one per
    request) — real data races on the program's variables. Requests are now
    answered one at a time; `darkhast` releases the lock while it waits, so a
    program can still call its own server.
16. `serva_bgo`'s window: `baza_bgo` now prefers `msedge.exe` over
    `msedgewebview2.exe` (the WebView2 engine is not meant to be started on its
    own with `--app`, and may show nothing).
17. `d.barmadar()` with no argument crashed into a Go error; dict methods now
    check how many values they were given.
18. The doc line of `adad_karda` promised `hich` for bad text, but it throws
    E4125 (which `examples/scripts/hesab.mhn` relies on). The doc now says so.
19. README promised `"salam {name}"` interpolates; only `$"salam {name}"` does.

### New (small, needed while building real sites)
* `d.begir(key, default)` / `d.get` — Python's `dict.get`.
* `poshe_barname()` / `script_dir` — the folder of the running program, so a site
  can keep its data beside itself wherever it is started from.
* `x ?? y`.
* A clearer note on "name not known" explaining block scope (see limitations).

## Editor (VS Code extension 0.2.0, publisher `mihani-circle` / "Mihani Circle")
* Language server: **signature help** (`(` and `,`) for builtins and for the
  file's own functs, with the active parameter; completions of your own functs
  show their full signature; hover on a funct shows its signature.
* Extension: signature-help provider; requests time out after 5 s instead of
  freezing suggestions when the server dies; **bundled `server/mihan.exe`**, so
  it works before `mihan install`; 5 project snippets (`m-api`, `m-page`,
  `m-window`, `m-script`, `m-khoroji`); publisher/version/description updated.
* Keywords, builtins with docs, user variables/functions/classes, methods after
  a dot, hover, diagnostics, go-to-definition, outline and format-on-save were
  already there and are unchanged.

## Documentation
* `docs/tutorial/README.md` + 13 lessons (`01-shoru.md` … `13-abzar.md`): basics →
  functions → classes → errors → files/JSON/scripts → backend → frontend →
  Windows app → tools. Each has explanation, runnable samples with their real
  output, and exercises with solutions. `tools/tutorial_check.py` runs every
  sample (including servers, which it calls over HTTP) and is part of
  `.\build.ps1 test` when Python is installed.
* `examples/pro-sites/`: `negar` (studio landing page + contact API),
  `ketabkhane` (book shop: search, cart, checkout, stock, orders on disk),
  `dashboard` (admin: KPIs, chart, orders, product CRUD). Screenshots in
  `examples/pro-sites/screens/`.

