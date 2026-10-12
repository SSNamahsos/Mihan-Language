<div dir="rtl">

# درس ۸ — خطاها (khata)

> **خلاصه:** `portab` یک خطا پرتاب می‌کند، `koshish` / `giroft` آن را می‌گیرد، و `nemigir` همیشه اجرا می‌شود.

## ۸.۱ وقتی میهن یک اشتباه پیدا می‌کند

اگر برنامه کار غیرممکنی بکند، میهن متوقف می‌شود و می‌گوید کجا و چرا؛ به فارسی و انگلیسی،
همراه با کدی که می‌توانید جست‌وجویش کنید:

</div>

```mihan fails
x = 10
chap_khat x / 0
```

```output
error: E4133
  نمی‌شود بر صفر تقسیم کرد
  division by zero has no answer
   2 | chap_khat x / 0
                 ^
```

<div dir="rtl">

## ۸.۲ گرفتن یک خطا

`koshish` («تلاش») یک بلوک را اجرا می‌کند. اگر داخل آن چیزی خراب شود، به‌جایش
`giroft khata e` («بگیر») اجرا می‌شود و خطا در `e` قرار می‌گیرد.

</div>

```mihan
koshish
    chap_khat 10 / 0
giroft khata e
    chap_khat "khata: " + e.message
chap_khat "barname edame dad"
```

```output
khata: division by zero has no answer
barname edame dad
```

<div dir="rtl">

## ۸.۳ پرتاب کردن خطای خودتان

`portab khata("...")` کار فعلی را متوقف می‌کند و خطا را به نزدیک‌ترین `giroft` می‌دهد.
می‌توانید به‌عنوان آرگومان دوم داده هم همراهش بفرستید.

</div>

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

<div dir="rtl">

## ۸.۴ nemigir — همیشه اجرا می‌شود

</div>

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

<div dir="rtl">

## ۸.۵ خواندن امن عددها

یک کاربرد خیلی رایج: کاربر چیزی تایپ کرده که عدد نیست.

</div>

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

<div dir="rtl">

کوتاه‌تر: `adad_ya_hich` عدد را می‌دهد، یا اگر متن عدد نباشد `hich` می‌دهد، و `??` یک
مقدار جایگزین انتخاب می‌کند:

</div>

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

<div dir="rtl">

## ۸.۶ عددهایی که خیلی بزرگ می‌شوند

عدد صحیح ۶۴ بیت دارد. رد شدن از آن یک خطاست (E4126)، نه این‌که بی‌صدا دور بزند؛ برای
مقدارهای خیلی بزرگ از عدد اعشاری (`1.0 * x`) استفاده کنید:

</div>

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

<div dir="rtl">

## تمرین‌ها

۱. تابع `taghsim(a, b)` را بنویسید که وقتی `b` صفر است `khata("taghsim bar sefr")` پرتاب
   کند، و آن را بگیرید.
۲. `["1", "x", "3"]` را به عدد تبدیل کنید و آن‌هایی را که تبدیل نمی‌شوند رد کنید.

</div>

<details><summary>جواب‌ها</summary>

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

<div dir="rtl">

**بعدی:** [درس ۹ — فایل‌ها، JSON و اسکریپت‌ها](09-script-ha.md)

</div>
