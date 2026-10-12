# Mihan for VS Code

Mihan — a programming language that reads the way Persian is spoken.

## What you get

- **The words, as you type.** Type `def` or `funct` and press Enter. The rest of
  the function appears, and the name is selected, so the next thing you type is
  the name. The same happens for `if` / `agar`, `for` / `har`, `class` /
  `darghan`, `try` / `koshish` and `serve` / `serva_bgo`. `def` writes `funct`,
  because the file stays in Finglish. After a dot, only the methods of the value
  in front of the dot.
- **Mistakes, while the file is open.** Underlined in place, in Persian and in
  English, each with a code you can look up.
- **Go to definition.** Hold `Ctrl` and click a name to jump to the line that
  gave it a value — including a name inside a function, which jumps to the
  function's own copy of it.
- **What a name means.** Hover over anything to see whether it is a variable, a
  function, a class, or one of the language's builtins, and what that builtin
  does.
- **The outline.** Every `darghan`, `funct`, `shenase` and `dastoor` in the file,
  in the order you wrote them.
- **Tidy on save.** The same tidying `mihan fmt` does. Turn it off with
  `"editor.formatOnSave": false`.
- **A new program in one command.** **Mihan: Start a new program** writes a first
  file that already runs, so the next thing you do is run it.

Everything above is decided by Mihan's own language server, in Mihan's own code,
reading the same tables the language runs from. A word added to the language
appears in this editor at the same moment, with the same explanation.

## Learning Mihan in this editor

You do not need to install anything else to *learn*: the editor reads the
tutorial and tells you what each word means as you go.

1. **Open the tutorial and read it where it lives.**
   `docs/tutorial/` is the course, in order, and every page is a real program. Open
   a page and the editor explains every word in it — hover over `agar` or
   `bargard` and the one-line meaning appears, because the language server reads
   the same table the language runs from.

2. **Run the page you are reading.** With the file open, either use your own
   terminal:

   ```
   mihan run docs/tutorial/02-arzhad.mhn
   ```

   or install the [Code Runner](https://marketplace.visualstudio.com/items?itemName=formulahendry.code-runner)
   extension and press `Ctrl+Alt+N`. The file runs, and you see exactly what the
   page claims it prints.

3. **Change it.** Learning to program is mostly changing things and seeing what
   happens. Delete a bracket, misspell a name, divide by zero — the mistake
   appears underlined in the editor before you run anything, in Persian and in
   English.

4. **Start your own.** Press `Ctrl+Shift+P`, choose **Mihan: Start a new program**,
   give it a name, and a folder appears with a `main.mhn` that already runs. Then
   `mihan run <folder>`.

5. **Check your own work.** Put what you expect your program to print at the top
   of the file, like this:

   ```mihan
   #khoroji
   #| 42
   #| tamam shod
   ```

   and `mihan test yourfile.mhn` tells you whether it still prints that. This is
   how every page of the tutorial is checked, and page 9 (`09-ayaz.mhn`)
   explains the shape. A program that can check itself is a program you can
   change without fear.

6. **Ask the language what it has.** Open the Command Palette and run
   **Mihan: Show the language server's messages** to see what the editor is
   asking behind the scenes, or type `mihan doc` in a terminal for the whole
   language in one place.

Two things worth knowing while you learn:

- **The REPL is faster than a file for a single question.** `mihan repl`, then
  type `2 + 2` and press Enter. It answers `4`. Names you give it stay for the
  next line.
- **The editor and the command line are the same language.** Nothing you see in
  the editor is a simplified version of Mihan; it is the same language server the
  compiler runs, so what the editor tells you is what will happen when you run it.

## Installing

1. Install the extension from the store: [Mihan Language on the VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=MihaniCircle.mihan-language), or search **Mihan Language** in the Extensions panel of VS Code.
2. Open a file ending in `.mhn`. Type `def` and press Enter. A function appears, with the name selected.

The extension already includes Mihan for Windows, so it works right away. To run programs from a terminal, download `mihan.exe` from the [Mihan Language releases](https://github.com/SSNamahsos/Mihan-Language/releases/latest) and run `mihan.exe install`.

## What the editor gets, and where it comes from

Suggestions, signature help, hover, mistakes, jumps and formatting all come from
`mihan lsp`, Mihan's own language server (`internal/lsp`). Keywords come from
`token.Keywords` and builtins from `eval.Builtins`, so a word added to the language
shows up in the editor at once. The snippets in `snippets/mihan.json` are the
larger project shapes (`m-api`, `m-page`, `m-window`, `m-script`, `m-khoroji`).
