<div dir="rtl">

# درس ۶ — تابع‌ها (funct)

> **خلاصه:** تابع با `funct` ساخته می‌شود و با `bargard` مقدار برمی‌گرداند.

## ۶.۱ تعریف کردن و صدا زدن

</div>

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

<div dir="rtl">

`bargard` (return) یک مقدار را برمی‌گرداند:

</div>

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

<div dir="rtl">

## ۶.۲ آرگومان‌های پیش‌فرض و آرگومان‌های اسم‌دار

</div>

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

<div dir="rtl">

## ۶.۳ نوع‌ها (اختیاری)

می‌توانید نوع هر پارامتر و نوع نتیجه را بنویسید. این کار تابع را مستند می‌کند و در راهنمای
ویرایشگر هم نشان داده می‌شود.

</div>

```mihan
funct jam(arz a, arz b) -> arz
    bargard a + b
chap_khat jam(2, 3)
```

```output
5
```

<div dir="rtl">

## ۶.۴ بازگشت (recursion)

یک تابع می‌تواند خودش را صدا بزند.

</div>

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

<div dir="rtl">

## ۶.۵ تابع‌ها هم مقدارند

یک تابع را می‌شود در یک متغیر نگه داشت، به تابع دیگری داد، یا با `funct(x) -> ...` همان‌جا
ساخت.

</div>

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

<div dir="rtl">

مثال آخر یک *closure* است: تابع داخلی `n` را به خاطر می‌سپارد.

## ۶.۶ متغیرهای داخل و بیرون

یک تابع می‌تواند اسمی را که بالای فایل ساخته شده بخواند و تغییر دهد:

</div>

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

<div dir="rtl">

## تمرین‌ها

۱. تابع `bozorgtar(a, b)` را بنویسید که بزرگ‌ترِ دو عدد را برگرداند (بدون `balia`).
۲. تابع `tavan2(n)` را بنویسید که با یک حلقه `2` به توان `n` را برگرداند.
۳. تابع `shomaresh(xs, x)` را بنویسید که بشمارد `x` چند بار در یک لیست آمده است.

</div>

<details><summary>جواب‌ها</summary>

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

<div dir="rtl">

**بعدی:** [درس ۷ — کلاس‌ها](07-darghan-ha.md)

</div>
