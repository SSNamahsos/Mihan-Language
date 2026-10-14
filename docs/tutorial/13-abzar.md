# Lesson 13 — Tools: test, check, fmt and VS Code

> **خلاصه:** `mihan test` بررسی می‌کند برنامه همان چیزی را چاپ کند که گفته‌ای، `mihan check` اشتباه‌ها را بدون اجرا پیدا می‌کند،
> و افزونه‌ی VS Code تکمیل خودکار و راهنما می‌دهد.

## 13.1 Saying what a program should print

Put a `#khoroji` block at the top of a file: one `#|` line per line of output.
`mihan test file.mhn` runs it and compares.

```mihan
#khoroji
#| 42
#| salam

chap_khat 6 * 7
chap_khat "salam"
```

```output
42
salam
```

```
> mihan test javab.mhn
salam  javab.mhn                                      kharch

1/1 sabet. Mihan 0.2.0
```

`mihan test` with no file runs every `.mhn` under `docs/`, `tests/` and `examples/`.

## 13.2 Other commands

| command | what it does |
|---|---|
| `mihan run x.mhn` | run a program |
| `mihan check x.mhn` | find mistakes without running |
| `mihan fmt x.mhn` | tidy the file into the standard shape |
| `mihan repl` | try one line at a time |
| `mihan doc` | the whole language on one page |
| `mihan new myapp` | start a new project folder |
| `mihan build app.mhn` | make **one `.exe`** that carries the program (see below) |

Lesson 14 shows `mihan build` in full.

## 13.3 VS Code

Install the extension (`mihan-0.7.8.vsix`: *Extensions → … → Install from VSIX*).
Then, in any `.mhn` file you get:

* **Autocomplete** for every keyword (Finglish and English), every builtin with
  its one-line help, and your own variables, functions and classes. Your own
  functions show their parameters.
* **Snippets**: type `funct`, `agar`, `har`, `darghan`, `koshish` and press
  Enter for the whole shape. `m-api`, `m-page`, `m-window`, `m-script` and
  `m-khoroji` write a complete server, page, window app, script or test block.
* **Signature help**: type `(` after a function to see its parameters.
* **Hover** over any word to see what it means.
* **Mistakes** underlined as you type, in Finglish.
* **Ctrl+click** to jump to where a name was made, and **format on save**.

The extension contains `mihan.exe`, so it works even before you run
`mihan install`.

## 13.4 Where to go next

* `mihan doc` — every word and builtin.
* `examples/` — scripts, sites and windows apps that are tested on every build.
* `examples/pro-sites/` — three complete, good-looking websites.

**Next:** [Lesson 14 — Sharing your program as one file](14-build.md)
