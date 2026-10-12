# Lesson 2 — Values and variables (arzesh-ha)

> **خلاصه:** متغیر یک اسم برای یک مقدار است. با `=` ساخته می‌شود.

## 2.1 Making a name

A variable is a name for a value. `=` gives a name a value — the first time it
*creates* the name, after that it *changes* it.

```mihan
sen = 20
naam = "Ali"
chap_khat naam
chap_khat sen
sen = sen + 1
chap_khat sen
sen += 1
chap_khat sen
```

```output
Ali
20
21
22
```

## 2.2 The kinds of values

| Mihan | English | Example |
|---|---|---|
| `arz` | whole number (int) | `42`, `-7` |
| `kasr` | number with a fraction (float) | `3.14` |
| `matn` | text (string) | `"salam"` |
| `bool` | true / false | `dorost`, `ghalat` |
| `hich` | nothing (None / null) | `hich` |
| `list`, `dict`, `set` | collections | lesson 4 |

```mihan
chap_khat 42
chap_khat 3.14
chap_khat "salam"
chap_khat dorost
chap_khat ghalat
chap_khat hich
chap_khat 10 / 2
```

```output
42
3.14
salam
dorost
ghalat
hich
5.0
```

`10 / 2` printed `5.0` and not `5`: division always gives a `kasr`.

## 2.3 Typed names (optional)

You may write the type in front of a name. It then starts at that type's zero and
documents what you intend to keep in it.

```mihan
arz shomar
matn esm
chap_khat shomar
chap_khat esm == ""
shomar = 5
chap_khat shomar
```

```output
0
dorost
5
```

## 2.4 Constants

`sabit` makes a name that can never change.

```mihan
sabit PI = 3.14159
chap_khat PI * 2
```

```output
6.28318
```

## 2.5 Changing a value's type

`tabdil x, type` converts. This is how you turn typed text into a number.

```mihan
chap_khat tabdil("42", arz) + 1
chap_khat tabdil("2.5", kasr) * 2
chap_khat tabdil(7, matn) + "!"
chap_khat tabdil(3.99, arz)
chap_khat adad_karda("12")
chap_khat adad_karda("2.5")
```

```output
43
5.0
7!
3
12
2.5
```

`adad_karda` reads a number from text and keeps it whole when it is whole. When
the text is not a number at all, the program stops and says so, in Persian and
English:

```mihan fails
chap_khat adad_karda("dah")
```

```output
error: E4125
  «dah» عدد نیست
  dah is not a number
   1 | chap_khat adad_karda("dah")
                 ^
```

Lesson 8 shows how to catch a mistake like this with `koshish` / `giroft`.

## 2.6 Comparing

```mihan
chap_khat 3 > 2
chap_khat 3 == 3.0
chap_khat "a" != "b"
chap_khat 1 < 2 va 2 < 3
chap_khat 1 > 2 ya dorost
chap_khat na dorost
```

```output
dorost
dorost
dorost
dorost
dorost
ghalat
```

`va` = and, `ya` = or, `na` = not.

`na` applies to the value right after it (a comparison counts as one value),
exactly like Python's `not`: `na a ya b` means `(na a) ya b`, and
`na x == y` means `na (x == y)`.

```mihan
chap_khat na dorost ya dorost
chap_khat na 1 == 2
```

```output
dorost
dorost
```

## Exercises

1. Make `gheymat = 12000` and `tedad = 3`, and print the total price.
2. Convert the text `"250"` to a number and add 50.
3. Is `17` bigger than `4 * 4`? Print the answer.

<details><summary>Solutions</summary>

```mihan
gheymat = 12000
tedad = 3
chap_khat gheymat * tedad
chap_khat tabdil("250", arz) + 50
chap_khat 17 > 4 * 4
```

```output
36000
300
dorost
```

</details>

**Next:** [Lesson 3 — Text](03-matn-ha.md)
