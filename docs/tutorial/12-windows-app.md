# Lesson 12 — A Windows app (baza_bgo)

> **خلاصه:** یک برنامه‌ی ویندوزی در میهن همان سایت است که در یک پنجره‌ی جدا (بدون نوار آدرس) باز می‌شود.

A Mihan desktop app is a small web server plus a window pointed at it.
`baza_bgo(address, width, height)` opens a frameless window (Microsoft Edge in
*app mode*, which is installed on every Windows 10/11 PC). Everything you learned
in lessons 10 and 11 works unchanged inside the window.

## 12.1 A counter app

The app's logic is kept in plain functions so it can be tested without a window:

```mihan
funct safhe(tedad)
    bargard por([
        "<!doctype html><html><head><meta charset='utf-8'><title>Shomarande</title>",
        "<style>body{font-family:system-ui;display:grid;place-items:center;height:100vh;margin:0;background:#14161a;color:#eee}",
        "h1{font-size:5rem;margin:0}button{font-size:1.2rem;margin:.3rem;padding:.6rem 1.2rem;border:0;border-radius:8px}</style></head>",
        "<body><div><h1>" + tedad + "</h1>",
        "<button onclick=\"location='/?kar=ziad'\">+1</button>",
        "<button onclick=\"location='/?kar=sefr'\">0</button></div></body></html>"
    ], "")

funct enjam(tedad, kar)
    agar kar == "ziad"
        bargard tedad + 1
    agar kar == "sefr"
        bargard 0
    bargard tedad

chap_khat enjam(4, "ziad")
chap_khat enjam(4, "sefr")
chap_khat "<h1>7</h1>" dar safhe(7)
```

```output
5
0
dorost
```

And this is the whole app. It cannot run inside this tutorial's checker (there is
no screen), so it is only checked for mistakes — run it yourself on Windows:

```mihan norun
tedad = 0

funct safhe(tedad)
    bargard "<!doctype html><meta charset='utf-8'><body style='font-family:system-ui;text-align:center'><h1>" + tedad + "</h1><button onclick=\"location='/?kar=ziad'\">+1</button></body>"

funct pasokh(darakht)
    agar darakht["sorok"].barmadar("kar")
        tedad += 1
    bargard {"vaziat": 200, "matn": safhe(tedad), "sarlawah": {"Content-Type": "text/html; charset=utf-8"}}

funct bargoz()
    baza_bgo(serva_khod(), 420, 520)

serva_bgo(0, pasokh, bargoz)
```

```
mihan run shomarande.mhn
```

A window opens with your app in it. Close the terminal (or press `Ctrl+C`) to quit.

## 12.2 Giving it to someone else

Copy `mihan.exe` and your `.mhn` files (and the `safhe` folder, if you have one)
into one folder. Make a file `start.bat` beside them:

```
@echo off
"%~dp0mihan.exe" run "%~dp0shomarande.mhn"
```

Double-clicking `start.bat` starts your app.

## 12.3 Honest limits (today)

* The window is Edge in app mode, not a native Win32 window. It needs Microsoft
  Edge, which Windows 10 and 11 include.
* Closing the window does not stop the program; stop it from the terminal, or add
  a "khoroj" button whose handler calls `serva_band()`.
* There is no single-file `.exe` builder for your program yet: you ship
  `mihan.exe` + your files.

## Exercise

Add a "khoroj" button to the counter app that stops the program.

<details><summary>Solution</summary>

```mihan norun
funct pasokh(darakht)
    agar darakht["path"] == "/khoroj"
        serva_band()
        bargard {"vaziat": 200, "matn": "khodahafez"}
    bargard {"vaziat": 200, "matn": "<a href='/khoroj'>khoroj</a>", "sarlawah": {"Content-Type": "text/html; charset=utf-8"}}

funct bargoz()
    baza_bgo(serva_khod(), 300, 200)

serva_bgo(0, pasokh, bargoz)
```

</details>

**Next:** [Lesson 13 — Tools: test, fmt, check and the editor](13-abzar.md)
