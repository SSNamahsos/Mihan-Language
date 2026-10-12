<div dir="rtl">

# درس ۲ — مقدارها و متغیرها (arzesh-ha)

> **خلاصه:** متغیر یک اسم برای یک مقدار است. با `=` ساخته می‌شود.

## ۲.۱ ساختن یک اسم

متغیر اسمی است برای یک مقدار. `=` به یک اسم مقدار می‌دهد؛ بار اول اسم را *می‌سازد* و
از آن به بعد مقدارش را *عوض می‌کند*.

</div>

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

<div dir="rtl">

## ۲.۲ انواع مقدارها

| میهن | معنی | مثال |
|---|---|---|
| `arz` | عدد صحیح (int) | `42`، `-7` |
| `kasr` | عدد اعشاری (float) | `3.14` |
| `matn` | متن (string) | `"salam"` |
| `bool` | درست / غلط | `dorost`، `ghalat` |
| `hich` | هیچ (None / null) | `hich` |
| `list`، `dict`، `set` | مجموعه‌ها | درس ۴ |

</div>

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

<div dir="rtl">

`10 / 2` به‌جای `5` مقدار `5.0` را چاپ کرد: تقسیم همیشه یک `kasr` می‌دهد.

## ۲.۳ اسم‌های نوع‌دار (اختیاری)

می‌توانید نوع را جلوی اسم بنویسید. در این صورت اسم با مقدار صفرِ آن نوع شروع می‌شود و
نشان می‌دهد قصد دارید چه چیزی در آن نگه دارید.

</div>

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

<div dir="rtl">

## ۲.۴ ثابت‌ها

`sabit` اسمی می‌سازد که هرگز عوض نمی‌شود.

</div>

```mihan
sabit PI = 3.14159
chap_khat PI * 2
```

```output
6.28318
```

<div dir="rtl">

## ۲.۵ عوض کردن نوع یک مقدار

`tabdil x, type` تبدیل می‌کند. این‌طوری متنی را که کاربر تایپ کرده به عدد تبدیل می‌کنید.

</div>

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

<div dir="rtl">

`adad_karda` یک عدد را از متن می‌خواند و اگر عدد صحیح باشد، صحیح نگهش می‌دارد. اگر متن
اصلاً عدد نباشد، برنامه متوقف می‌شود و این را به فارسی و انگلیسی می‌گوید:

</div>

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

<div dir="rtl">

درس ۸ نشان می‌دهد چطور چنین خطایی را با `koshish` / `giroft` بگیرید.

## ۲.۶ مقایسه

</div>

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

<div dir="rtl">

`va` = و، `ya` = یا، `na` = نه.

`na` روی مقداری که درست بعد از آن می‌آید اثر می‌کند (یک مقایسه یک مقدار حساب می‌شود)،
دقیقاً مثل `not` در پایتون: `na a ya b` یعنی `(na a) ya b` و
`na x == y` یعنی `na (x == y)`.

</div>

```mihan
chap_khat na dorost ya dorost
chap_khat na 1 == 2
```

```output
dorost
dorost
```

<div dir="rtl">

## تمرین‌ها

۱. `gheymat = 12000` و `tedad = 3` را بسازید و قیمت کل را چاپ کنید.
۲. متن `"250"` را به عدد تبدیل کنید و ۵۰ به آن اضافه کنید.
۳. آیا `17` از `4 * 4` بزرگ‌تر است؟ جواب را چاپ کنید.

</div>

<details><summary>جواب‌ها</summary>

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

<div dir="rtl">

**بعدی:** [درس ۳ — متن](03-matn-ha.md)

</div>
