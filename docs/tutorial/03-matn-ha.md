# Lesson 3 — Text (matn)

> **خلاصه:** متن بین دو `"` نوشته می‌شود. با `$"..."` می‌توان مقدار را داخل متن گذاشت.

## 3.1 Joining and filling in

`+` joins text — even with numbers. A `$` in front of the quotes lets you put any
expression inside `{ }`:

```mihan
naam = "Sara"
sen = 25
chap_khat "salam " + naam
chap_khat "sen: " + sen
chap_khat $"{naam} {sen} sale ast."
chap_khat $"sale dige {sen + 1} sale mishe."
```

```output
salam Sara
sen: 25
Sara 25 sale ast.
sale dige 26 sale mishe.
```

Without the `$`, braces are just characters: `"salam {naam}"` prints exactly that.

## 3.2 Special characters

```mihan
chap_khat "khat 1\nkhat 2"
chap_khat "a\tb"
chap_khat "u goft: \"salam\""
```

```output
khat 1
khat 2
a	b
u goft: "salam"
```

## 3.3 Length, positions and pieces

```mihan
s = "Mihan"
chap_khat tule s
chap_khat s[0]
chap_khat s[-1]
chap_khat s[1:3]
chap_khat "سلام"[0]
```

```output
5
M
n
ih
س
```

Positions start at `0`; `-1` is the last character. Persian letters count as one
character each.

## 3.4 Methods

A method is called with a dot on the text in front of it.

```mihan
s = "  Salam Donya  "
chap_khat s.jabeja()
chap_khat s.jabeja().gooshband()
chap_khat s.jabeja().kiawtar()
chap_khat "a,b,c".tab(",")
chap_khat "salam".jigarre("a", "o")
chap_khat "salam".shamil("lam")
chap_khat "salam".barmigefth("sa")
chap_khat "salam".bapakhsh("am")
chap_khat por(["x", "y", "z"], "-")
```

```output
Salam Donya
SALAM DONYA
salam donya
["a", "b", "c"]
solom
dorost
dorost
dorost
x-y-z
```

| method | meaning |
|---|---|
| `jabeja()` | remove spaces at both ends (trim) |
| `gooshband()` / `kiawtar()` | UPPER / lower |
| `tab(",")` | split into a list |
| `jigarre(a, b)` | replace |
| `shamil(x)` | contains? |
| `barmigefth(x)` / `bapakhsh(x)` | starts with / ends with |
| `por(list, sep)` | join a list into text |

## 3.5 Numbers in a shape

```mihan
chap_khat katoos(3.14159, "%0.2f")
chap_khat katoos(7, "%03d")
```

```output
3.14
007
```

## Exercises

1. Given `esm = "reza"`, print `REZA` and the number of letters.
2. Split `"sib,moz,porteghal"` and print the second fruit.
3. Print `Gheymat: 12.50 toman` from the number `12.5`.

<details><summary>Solutions</summary>

```mihan
esm = "reza"
chap_khat esm.gooshband()
chap_khat tule esm
miveh = "sib,moz,porteghal".tab(",")
chap_khat miveh[1]
chap_khat $"Gheymat: {katoos(12.5, "%0.2f")} toman"
```

```output
REZA
4
moz
Gheymat: 12.50 toman
```

</details>

**Next:** [Lesson 4 — Lists, dicts and sets](04-majmooe-ha.md)
