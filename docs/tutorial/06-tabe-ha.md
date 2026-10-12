# Lesson 6 — Functions (funct)

> **خلاصه:** تابع با `funct` ساخته می‌شود و با `bargard` مقدار برمی‌گرداند.

## 6.1 Defining and calling

```mihan
funct salam_kon(naam)
    chap_khat $"salam {naam}!"

salam_kon("Sara")
salam_kon("Ali")
```

```output
salam Sara!
salam Ali!
```

`bargard` (return) gives a value back:

```mihan
funct masahat(arz_, tool)
    bargard arz_ * tool

chap_khat masahat(3, 4)
x = masahat(2, 5) + 1
chap_khat x
```

```output
12
11
```

## 6.2 Default and named arguments

```mihan
funct moarefi(naam, shahr = "Tehran")
    bargard $"{naam} az {shahr}"

chap_khat moarefi("Sara")
chap_khat moarefi("Ali", "Tabriz")
chap_khat moarefi("Neda", shahr: "Yazd")
chap_khat moarefi(shahr = "Rasht", naam = "Omid")
```

```output
Sara az Tehran
Ali az Tabriz
Neda az Yazd
Omid az Rasht
```

## 6.3 Types (optional)

You can write the type of each parameter and of the result. It documents the
function and shows up in the editor's help.

```mihan
funct jam(arz a, arz b) -> arz
    bargard a + b
chap_khat jam(2, 3)
```

```output
5
```

## 6.4 Recursion

A function may call itself.

```mihan
funct factoriel(n)
    agar n <= 1
        bargard 1
    bargard n * factoriel(n - 1)

chap_khat factoriel(5)
chap_khat factoriel(20)
```

```output
120
2432902008176640000
```

## 6.5 Functions are values

A function can be stored in a variable, passed to another function, or created
in place with `funct(x) -> ...`.

```mihan
funct do_bar(f, x)
    bargard f(f(x))

zarb3 = funct(x) -> x * 3
chap_khat do_bar(zarb3, 2)
chap_khat do_bar(funct(s) -> s + "!", "salam")

funct sazande_jam(n)
    bargard funct(x) -> x + n
be_alave_10 = sazande_jam(10)
chap_khat be_alave_10(5)
```

```output
18
salam!!
15
```

The last example is a *closure*: the inner function remembers `n`.

## 6.6 Variables inside and outside

A function can read and change a name made at the top of the file:

```mihan
shomar = 0
funct ziad_kon()
    shomar += 1

ziad_kon()
ziad_kon()
chap_khat shomar
```

```output
2
```

## Exercises

1. Write `bozorgtar(a, b)` that returns the bigger of two numbers (without `balia`).
2. Write `tavan2(n)` that returns `2` to the power `n` using a loop.
3. Write `shomaresh(xs, x)` that counts how many times `x` is in a list.

<details><summary>Solutions</summary>

```mihan
funct bozorgtar(a, b)
    agar a > b
        bargard a
    bargard b

funct tavan2(n)
    javab = 1
    har i az 0 ta n
        javab *= 2
    bargard javab

funct shomaresh(xs, x)
    c = 0
    har y dar xs
        agar y == x
            c += 1
    bargard c

chap_khat bozorgtar(3, 9)
chap_khat tavan2(10)
chap_khat shomaresh([1, 2, 1, 3, 1], 1)
```

```output
9
1024
3
```

</details>

**Next:** [Lesson 7 — Classes](07-darghan-ha.md)
