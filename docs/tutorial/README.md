# Learn Mihan — the course

فارسی: [fa/README.md](fa/README.md)

> میهن را از صفر یاد بگیرید. هر درس توضیح، مثال قابل اجرا، خروجی واقعی و تمرین با جواب دارد.

Read the lessons in order. Every code sample is a real program: copy it into a
`.mhn` file, run it with `mihan run file.mhn`, and you will see exactly the output
shown under it. (`python3 tools/tutorial_check.py` re-runs every sample and fails
if any output differs, so the course cannot drift from the language.)

| # | Lesson | You will learn |
|---|---|---|
| 1 | [Getting started](01-shoru.md) | install, first program, printing, the REPL |
| 2 | [Values and variables](02-arzesh-ha.md) | numbers, text, true/false, types, converting |
| 3 | [Text](03-matn-ha.md) | `$"..."`, slicing, text methods, formatting |
| 4 | [Lists, dicts and sets](04-majmooe-ha.md) | collections, `filter`, `tabdil_kol`, ranges |
| 5 | [Decisions and loops](05-tasmim-va-tekrar.md) | `agar`, `warna`, `har`, `ta`, `shek`, `edame` |
| 6 | [Functions](06-tabe-ha.md) | `funct`, defaults, named args, lambdas, closures |
| 7 | [Classes](07-darghan-ha.md) | `darghan`, `khod`, constructors, inheritance |
| 8 | [Errors](08-khata-ha.md) | `koshish`, `giroft`, `portab`, `nemigir` |
| 9 | [Files, JSON and scripts](09-script-ha.md) | files, JSON, input, arguments, `biar` |
| 10 | [A website backend](10-backend.md) | `serva_bgo`, requests, a JSON REST API |
| 11 | [A website frontend](11-frontend.md) | HTML/CSS/JS files, or pages of Mihan pieces |
| 12 | [A Windows app](12-windows-app.md) | `baza_bgo`, shipping your app |
| 13 | [Tools](13-abzar.md) | `mihan test`, `check`, `fmt`, VS Code |
| 14 | [Sharing as one file](14-build.md) | `mihan build` for Windows and Linux |
| 15 | [On an Android phone](15-android.md) | `mihan build --os android`, an APK |
| 16 | [A real web API](16-web-api.md) | routes with `:id`, methods, CORS, cookies, `jalase`, uploads, `qaleb`, live events |
| 17 | [Trying an Android app](17-preview.md) | `mihan preview --android`, `mihan.toast/vibrate/share/onBack` |
| 18 | [Reference and more tools](18-marja.md) | every built-in word, maths, type checks |

The short `NN-*.mhn` files in this folder are the original one-page reference
programs; they are run by `mihan test` and complement the lessons.

## Keyword cheat-sheet

| Mihan | English | | Mihan | English |
|---|---|---|---|---|
| `chap_khat` | println | | `funct` | def / func |
| `agar` | if | | `bargard` | return |
| `warna` | else | | `darghan` | class |
| `har … dar` | for … in | | `khod` | self |
| `har i az a ta b` | for i in range(a, b) | | `koshish` / `giroft` | try / catch |
| `ta` | while | | `portab` | raise / throw |
| `shek` / `edame` | break / continue | | `biar` | import |
| `va` / `ya` / `na` | and / or / not | | `dorost` / `ghalat` / `hich` | true / false / null |
