# Lesson 14 — Sharing your program as one file (mihan build)

You wrote a program; now you want to give it to someone who has never heard of
Mihan. `mihan build` makes **one file** that holds the Mihan runtime *and* your
program. The other person just runs it.

## 14.1 The simplest build

```mihan file=salam/salam.mhn
naam = khanak("esmet chie? ")
chap_khat "salam " + naam + "!"
```

```
mihan build salam/salam.mhn
```

On Windows that writes `salam.exe`; on Linux it writes `salam`. Run it:

```
./salam            # Linux
salam.exe          # Windows
```

Everything after the name goes to your program, so `arghoman()` works as in
lesson 9.

## 14.2 Choosing the system: --os and --arch

You can build for the other system from either one:

| you want | command |
|---|---|
| Windows, 64-bit PC | `mihan build app.mhn --os windows` |
| Windows on ARM | `mihan build app.mhn --os windows --arch arm64` |
| Linux, 64-bit PC | `mihan build app.mhn --os linux` |
| Linux on ARM (Raspberry Pi 4/5, ARM servers) | `mihan build app.mhn --os linux --arch arm64` |

`-o name` picks the output name. Building for another system needs that
system's runtime, which comes with Mihan in the `runtimes` folder beside
`mihan` (`mihan-linux-amd64`, `mihan-linux-arm64`, `mihan-windows-amd64.exe`,
`mihan-windows-arm64.exe`). You do **not** need Go. If you keep the runtimes
somewhere else, set `MIHAN_RUNTIMES` to that folder, or pass one with
`--base path/to/runtime`.

## 14.3 Window apps: --window

A program that opens a window with `baza_bgo` (lesson 12) is an *app*, and on
Windows an app should not have a black console window next to it:

```
mihan build counter.mhn --os windows --window     # no console window
mihan build tool.mhn --os windows                  # a script: keeps its console
```

Use `--window` only for programs that show a window; a script built with it
would have nowhere to print. On Linux there is no console to hide, so the flag
changes nothing there.

When the person closes the app's window, the program stops too: no server is
left running in the background.

On Linux the window opens in Chrome, Chromium, Edge or Brave in *app mode* (a
window with no address bar), or in Firefox in a window of its own. If none of
those is installed, the address opens in the default browser with `xdg-open`;
then Mihan cannot tell when the window closes, so the program keeps running
until you stop it. `MIHAN_BROWSER=/path/to/browser` picks a browser yourself.

## 14.3b Sites that open the browser: --open-browser

A site (a program that calls `serva_bgo`) built into an `.exe` would
otherwise start quietly and wait for someone to type its address. With
`--open-browser`, double-clicking the file opens the site in the default
browser as soon as it is listening:

```
mihan build examples/pro-sites/negar/negar.mhn --os windows --open-browser -o Negar.exe
```

Keep the console (no `--window`) if you want a place to see the address and
stop the site with Ctrl+C; add `--window` for a site that should run with no
console at all (stop it from Task Manager).

## 14.4 What goes inside

The whole folder of the file you build goes in: other `.mhn` files you `biar`,
the `safhe/` folder of pages, pictures, CSS. Left out: `dade/` folders (data
your program writes), hidden folders such as `.git`, and other `.exe` files.
So keep a program in a folder of its own.

```
mihan build examples/pro-sites/negar/negar.mhn --os windows --window -o Negar.exe
```

On its first run the file unpacks your program into a folder of its own:

* Windows: `%LOCALAPPDATA%\mihan-apps\<name>\`
* Linux: `~/.cache/mihan-apps/<name>/`

It does that once per build. Data your program writes beside itself (for
example `dade/`, using `poshe_barname()`) lives there and survives a rebuild.

## Exercises

1. Build lesson 9's word counter for your own system and run it with a file
   name after it.
2. Build the counter app from `examples/windows/counter.mhn` with `--window`
   for Windows and without it for Linux, and compare the two file sizes.

**Back to the start:** [Course contents](README.md)
