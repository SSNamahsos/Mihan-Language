# Lesson 8 — Errors (khata)

> **خلاصه:** `portab` یک خطا پرتاب می‌کند، `koshish` / `giroft` آن را می‌گیرد، و `nemigir` همیشه اجرا می‌شود.

## 8.1 When Mihan finds a mistake

If a program does something impossible, Mihan stops and says where and why — in
Finglish, with a code you can search for:

```mihan fails
x = 10
chap_khat x / 0
```

```output
khata: E4133
  nemishe bar sefr taghsim kard
   2 | chap_khat x / 0
                 ^
```

## 8.2 Catching a mistake

`koshish` ("try") runs a block. If something goes wrong inside it, `giroft khata e`
("catch") runs instead, with the error in `e`.

```mihan
koshish
    chap_khat 10 / 0
giroft khata e
    chap_khat "khata: " + e.message
chap_khat "barname edame dad"
```

```output
khata: nemishe bar sefr taghsim kard
barname edame dad
```

## 8.3 Throwing your own error

`portab khata("...")` stops the current work and hands the error to the nearest
`giroft`. You can attach data as a second argument.

```mihan
funct bardasht(mojudi, mablagh)
    agar mablagh > mojudi
        portab khata("mojudi kafi nist", mablagh - mojudi)
    bargard mojudi - mablagh

koshish
    chap_khat bardasht(100, 30)
    chap_khat bardasht(100, 250)
    chap_khat "in khat chap nemishavad"
giroft khata e
    chap_khat $"{e.message} (kam: {e.data})"
```

```output
70
mojudi kafi nist (kam: 150)
```

## 8.4 nemigir — always runs

```mihan
funct kar(n)
    koshish
        agar n < 0
            portab khata("manfi")
        chap_khat "ok"
    giroft khata e
        chap_khat "giroftam: " + e.message
    nemigir
        chap_khat "tamiz kari"

kar(1)
kar(-1)
```

```output
ok
tamiz kari
giroftam: manfi
tamiz kari
```

## 8.5 Reading numbers safely

A very common use: the user typed something that is not a number.

```mihan
funct adad_ya_sefr(matn_)
    koshish
        bargard adad_karda(matn_)
    giroft khata e
        bargard 0

chap_khat adad_ya_sefr("42")
chap_khat adad_ya_sefr("chehel")
```

```output
42
0
```

Shorter: `adad_ya_hich` gives the number, or `hich` when the text is not a
number, and `??` picks a fallback:

```mihan
chap_khat adad_ya_hich("42")
chap_khat adad_ya_hich("chehel")
chap_khat adad_ya_hich("chehel") ?? 0
```

```output
42
hich
0
```

## 8.6 Numbers that get too big

A whole number has 64 bits. Going past that is a mistake (E4126), not a silent
wrap-around; use a float (`1.0 * x`) for huge values:

```mihan
koshish
    chap_khat 9223372036854775807 + 1
giroft khata e
    chap_khat "kheili bozorg!"
chap_khat 2.0 ** 70
```

```output
kheili bozorg!
1.1805916207174113e+21
```

## Exercises

1. Write `taghsim(a, b)` that throws `khata("taghsim bar sefr")` when `b` is 0,
   and catch it.
2. Turn `["1", "x", "3"]` into numbers, skipping the ones that fail.

<details><summary>Solutions</summary>

```mihan
funct taghsim(a, b)
    agar b == 0
        portab khata("taghsim bar sefr")
    bargard a / b

koshish
    chap_khat taghsim(9, 3)
    chap_khat taghsim(1, 0)
giroft khata e
    chap_khat e.message

javab = []
har m dar ["1", "x", "3"]
    koshish
        bast javab, adad_karda(m)
    giroft khata e
        edame
chap_khat javab
```

```output
3.0
taghsim bar sefr
[1, 3]
```

</details>

**Next:** [Lesson 9 — Files, JSON and scripts](09-script-ha.md)
