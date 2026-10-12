# Lesson 1 — Getting started (shoru)

> **خلاصه:** میهن زبانی است که کلمه‌هایش فارسی هستند ولی با حروف لاتین نوشته می‌شوند (فینگلیش).
> در این درس برنامه‌ی اول را می‌نویسیم و اجرا می‌کنیم.

Mihan is a programming language whose keywords are Persian words written in Latin
letters. `chap` means *print*, `agar` means *if*, `har` means *for*. Every keyword
also has an English spelling (`print`, `if`, `for`), so you can mix while you learn.

## 1.1 Install

On Windows, put `mihan.exe` somewhere permanent (for example `C:\Mihan\`) and run:

```
C:\Mihan\mihan.exe install
```

That copies Mihan into your user folder and adds it to your `PATH`. Open a **new**
terminal and check:

```
mihan version
```

## 1.2 Your first program

Make a file called `salam.mhn` and write one line in it:

```mihan
chap_khat "salam, donya!"
```

```output
salam, donya!
```

Run it:

```
mihan run salam.mhn
```

`chap_khat` ("print line") writes a value and then starts a new line. `chap` writes
without the new line:

```mihan
chap "yek "
chap "do "
chap_khat "se"
chap_khat "tamam"
```

```output
yek do se
tamam
```

## 1.3 Comments

Everything after `#` on a line is a comment — Mihan ignores it.

```mihan
# in yek tozih ast (this is a comment)
chap_khat 2 + 3   # the answer is printed, the comment is not
```

```output
5
```

## 1.4 Mihan as a calculator

```mihan
chap_khat 2 + 3 * 4
chap_khat (2 + 3) * 4
chap_khat 7 / 2
chap_khat 7 ghesmat 2
chap_khat 7 % 2
chap_khat 2 ** 10
```

```output
14
20
3.5
3
1
1024
```

Notice two rules:

* `/` **always** gives a fraction: `7 / 2` is `3.5`.
* `ghesmat` is whole-number division (Python's `//`), `%` is the remainder.

## 1.5 The REPL

`mihan repl` lets you type one line at a time and see the answer immediately — a
great way to experiment. Type `exit` to leave.

```
mihan> 2 + 2
4
mihan> naam = "Sara"
naam = "Sara"
mihan> exit
```

## Exercises

1. Print your own name on one line, and your city on the next.
2. Print how many minutes there are in a week.
3. Print `7` divided by `2` both as a fraction and as a whole number.

<details><summary>Solutions</summary>

```mihan
chap_khat "Sara"
chap_khat "Tehran"
chap_khat 7 * 24 * 60
chap_khat 7 / 2
chap_khat 7 ghesmat 2
```

```output
Sara
Tehran
10080
3.5
3
```

</details>

**Next:** [Lesson 2 — Values and variables](02-arzesh-ha.md)
