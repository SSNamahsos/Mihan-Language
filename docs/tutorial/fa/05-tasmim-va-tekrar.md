<div dir="rtl">

# درس ۵ — تصمیم‌ها و حلقه‌ها

> **خلاصه:** `agar` یعنی «اگر»، `warna` یعنی «وگرنه»، `ta` یعنی «تا وقتی که»، `har` یعنی «برای هر».
> بلوک‌ها با تورفتگی (۴ فاصله) مشخص می‌شوند، مثل پایتون.

## ۵.۱ agar / warna

بلوک یعنی خط‌هایی که زیر یک خط با تورفتگی آمده‌اند. از ۴ فاصله استفاده کنید. دونقطه‌ی آخر
خط اختیاری است.

</div>

```mihan
nomre = 17
agar nomre >= 18
    chap_khat "aali"
warna agar nomre >= 12
    chap_khat "ghabool"
warna
    chap_khat "mardood"
```

```output
ghabool
```

<div dir="rtl">

## ۵.۲ har — شمردن

`har i az a ta b` از `a` می‌شمارد تا `b`، ولی خود `b` را **شامل نمی‌شود**. برای این‌که `b` هم
شامل شود، `ta b+` بنویسید.

</div>

```mihan
har i az 0 ta 3
    chap_khat i
har i az 1 ta 3+
    chap_khat $"dor {i}"
```

```output
0
1
2
dor 1
dor 2
dor 3
```

<div dir="rtl">

## ۵.۳ har — گشتن روی یک لیست

</div>

```mihan
rangha = ["ghermez", "sabz", "abi"]
har r dar rangha
    chap_khat r.gooshband()
har jofte dar doreh(rangha)
    chap_khat $"{jofte[0]}: {jofte[1]}"
```

```output
GHERMEZ
SABZ
ABI
0: ghermez
1: sabz
2: abi
```

<div dir="rtl">

`ke` فقط عضوهایی را نگه می‌دارد که از یک آزمون رد شوند:

</div>

```mihan
har n az 1 ta 10 ke n % 3 == 0
    chap_khat n
```

```output
3
6
9
```

<div dir="rtl">

## ۵.۴ ta — تکرار تا وقتی که

</div>

```mihan
pool = 100
sal = 0
ta pool < 200
    pool = pool * 1.25
    sal += 1
chap_khat $"{sal} sal, pool: {pool}"
```

```output
4 sal, pool: 244.140625
```

<div dir="rtl">

## ۵.۵ shek و edame

`shek` (break) از حلقه بیرون می‌رود. `edame` (continue) به دور بعدی می‌پرد.

</div>

```mihan
har i az 1 ta 10
    agar i == 3
        edame
    agar i == 6
        shek
    chap_khat i
```

```output
1
2
4
5
```

<div dir="rtl">

## ۵.۶ اسم‌هایی که داخل بلوک ساخته می‌شوند بعد از آن هم می‌مانند

مثل پایتون، اسمی که داخل یک بلوک تورفته (`agar`، `har`، `ta`، `koshish`) ساخته شود مال
تابع (یا فایلی) است که دورش است، پس بعد از بلوک هم می‌توانید از آن استفاده کنید:

</div>

```mihan
har i az 0 ta 3
    akhar = i
chap_khat akhar

agar dorost
    payam = "salam"
chap_khat payam
```

```output
2
salam
```

<div dir="rtl">

اسم‌هایی که داخل یک `funct` ساخته می‌شوند همچنان فقط داخل همان تابع می‌مانند.

## ۵.۷ یک برنامه‌ی کوچک: FizzBuzz

</div>

```mihan
har i az 1 ta 15+
    agar i % 15 == 0
        chap_khat "FizzBuzz"
    warna agar i % 3 == 0
        chap_khat "Fizz"
    warna agar i % 5 == 0
        chap_khat "Buzz"
    warna
        chap_khat i
```

```output
1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz
```

<div dir="rtl">

## تمرین‌ها

۱. جدول ضرب ۷ را چاپ کنید (7 x 1 … 7 x 10).
۲. همه‌ی عددهای ۱ تا ۱۰۰ را با هم جمع کنید.
۳. اولین عدد بزرگ‌تر از ۱۰۰۰ را پیدا کنید که بر ۳۷ بخش‌پذیر باشد.

</div>

<details><summary>جواب‌ها</summary>

```mihan
har i az 1 ta 10+
    chap_khat $"7 x {i} = {7 * i}"
jam = 0
har i az 1 ta 100+
    jam += i
chap_khat jam
n = 1001
ta n % 37 != 0
    n += 1
chap_khat n
```

```output
7 x 1 = 7
7 x 2 = 14
7 x 3 = 21
7 x 4 = 28
7 x 5 = 35
7 x 6 = 42
7 x 7 = 49
7 x 8 = 56
7 x 9 = 63
7 x 10 = 70
5050
1036
```

</details>

<div dir="rtl">

**بعدی:** [درس ۶ — تابع‌ها](06-tabe-ha.md)

</div>
