# Lesson 4 — Lists, dicts and sets

> **خلاصه:** `list` یک صف از مقدارهاست، `dict` جدول کلید و مقدار، و `set` مجموعه‌ای بدون تکرار.

## 4.1 Lists

A list keeps values in order. Positions start at `0`.

```mihan
miveh = ["sib", "moz", "anar"]
chap_khat miveh
chap_khat miveh[0]
chap_khat miveh[-1]
chap_khat tule miveh
bast miveh, "hoolu"
chap_khat miveh
miveh[1] = "gilas"
chap_khat miveh
chap_khat miveh[1:3]
chap_khat bardavar(miveh)
chap_khat miveh
```

```output
["sib", "moz", "anar"]
sib
anar
3
["sib", "moz", "anar", "hoolu"]
["sib", "gilas", "anar", "hoolu"]
["gilas", "anar"]
hoolu
["sib", "gilas", "anar"]
```

| builtin | what it does |
|---|---|
| `bast xs, x` | add to the end (push / append) |
| `bardavar xs` | take the last item off and return it (pop) |
| `biya xs, i, x` | insert at a position |
| `hazar xs, x` | position of an item, or -1 |
| `barmadar xs, x` | is the item in the list (a key, in a dict)? In a text, is this piece of text in it: `barmadar("salam donya", "donya")` is `dorost` |
| `taghib xs` | a sorted copy |
| `forogh xs` | a reversed copy |

```mihan
adad = [5, 3, 8, 1]
chap_khat taghib adad
chap_khat forogh adad
chap_khat hazar(adad, 8)
chap_khat barmadar(adad, 4)
chap_khat 3 dar adad
biya adad, 0, 100
chap_khat adad
chap_khat [1, 2] + [3, 4]
```

```output
[1, 3, 5, 8]
[1, 8, 3, 5]
2
ghalat
dorost
[100, 5, 3, 8, 1]
[1, 2, 3, 4]
```

## 4.2 Doing something to every item

`filter` keeps items that pass a test, `tabdil_kol` changes every item, and
`kamil` folds the list into one value. `funct(x) -> ...` is a small function
written in place (a *lambda*).

```mihan
adad = [1, 2, 3, 4, 5, 6]
chap_khat filter(adad, funct(x) -> x % 2 == 0)
chap_khat tabdil_kol(adad, funct(x) -> x * x)
chap_khat kamil(adad, funct(a, b) -> a + b)
chap_khat kamej(adad, funct(x) -> x > 5)
chap_khat hamej(adad, funct(x) -> x > 0)
```

```output
[2, 4, 6]
[1, 4, 9, 16, 25, 36]
21
dorost
dorost
```

### Building a list in one line

`[kar har x dar list]` makes a new list from another one, and `ke` (or `agar`)
keeps only some items — like Python's list comprehensions:

```mihan
adad = [1, 2, 3, 4, 5, 6]
chap_khat [x * x har x dar adad]
chap_khat [x har x dar adad ke x % 2 == 0]
chap_khat [i * 10 har i az 0 ta 4]
chap_khat [k + "!" har k dar ["salam", "dorood"]]
```

```output
[1, 4, 9, 16, 25, 36]
[2, 4, 6]
[0, 10, 20, 30]
["salam!", "dorood!"]
```

## 4.3 Dicts

A dict maps keys to values. It remembers the order keys were added.

```mihan
shakhs = {"naam": "Sara", "sen": 25}
chap_khat shakhs["naam"]
shakhs["shahr"] = "Shiraz"
shakhs["sen"] = 26
chap_khat shakhs
chap_khat dast shakhs
chap_khat meghdar shakhs
chap_khat "naam" dar shakhs
chap_khat tule shakhs
```

```output
Sara
{"naam": "Sara", "sen": 26, "shahr": "Shiraz"}
["naam", "sen", "shahr"]
["Sara", 26, "Shiraz"]
dorost
3
```

Dicts and lists nest — this is exactly the shape of JSON:

```mihan
dokan = {
    "naam": "Dokan-e Ali",
    "kala": [
        {"naam": "nan", "gheymat": 2},
        {"naam": "shir", "gheymat": 5}
    ]
}
chap_khat dokan["kala"][1]["naam"]
har k dar dokan["kala"]
    chap_khat $"{k["naam"]}: {k["gheymat"]}"
```

```output
shir
nan: 2
shir: 5
```

## 4.4 Sets

A set keeps each value only once.

```mihan
s = {1, 2, 2, 3, 3, 3}
chap_khat s
chap_khat 2 dar s
chap_khat setare([1, 1, 2, 5, 5])
```

```output
{1, 2, 3}
dorost
{1, 2, 5}
```

## 4.5 Ranges

`a..b` is the numbers from `a` up to (not including) `b`; `a..+b` includes `b`.

```mihan
chap_khat 1..5
chap_khat 1..+5
chap_khat rang(0, 10, 3)
chap_khat doreh(["a", "b"])
chap_khat zip([1, 2, 3], ["yek", "do", "se"])
```

```output
1..5
1..+5
0..3..10
[[0, "a"], [1, "b"]]
[[1, "yek"], [2, "do"], [3, "se"]]
```

## Exercises

1. Make a list of five numbers and print the biggest, using `taghib`.
2. Keep only the words longer than 3 letters from `["man", "ketab", "to", "daftar"]`.
3. Make a dict of three friends and their ages, then print each as `name: age`.

<details><summary>Solutions</summary>

```mihan
adad = [4, 9, 2, 7, 5]
chap_khat taghib(adad)[-1]
kalame = ["man", "ketab", "to", "daftar"]
chap_khat filter(kalame, funct(k) -> tule(k) > 3)
doostan = {"Ali": 20, "Neda": 22, "Reza": 19}
har naam dar doostan
    chap_khat $"{naam}: {doostan[naam]}"
```

```output
9
["ketab", "daftar"]
Ali: 20
Neda: 22
Reza: 19
```

</details>

**Next:** [Lesson 5 — Decisions and loops](05-tasmim-va-tekrar.md)
