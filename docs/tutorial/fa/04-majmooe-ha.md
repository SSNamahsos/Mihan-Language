<div dir="rtl">

# درس ۴ — لیست‌ها، دیکشنری‌ها و مجموعه‌ها

> **خلاصه:** `list` یک صف از مقدارهاست، `dict` جدول کلید و مقدار، و `set` مجموعه‌ای بدون تکرار.

## ۴.۱ لیست‌ها

لیست مقدارها را به ترتیب نگه می‌دارد. جایگاه‌ها از `0` شروع می‌شوند.

</div>

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

<div dir="rtl">

| تابع آماده | کارش |
|---|---|
| `bast xs, x` | اضافه کردن به آخر (push / append) |
| `bardavar xs` | برداشتن آخرین عضو و برگرداندنش (pop) |
| `biya xs, i, x` | گذاشتن در یک جایگاه مشخص |
| `hazar xs, x` | جایگاه یک عضو، یا -1 |
| `barmadar xs, x` | آیا این عضو در لیست هست (در دیکشنری: آیا این کلید هست)؟ در متن: آیا این تکه‌متن در آن هست؛ `barmadar("salam donya", "donya")` می‌شود `dorost` |
| `taghib xs` | یک کپی مرتب‌شده |
| `forogh xs` | یک کپی برعکس‌شده |

</div>

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

<div dir="rtl">

## ۴.۲ انجام یک کار روی همه‌ی عضوها

`filter` عضوهایی را که از یک آزمون رد می‌شوند نگه می‌دارد، `tabdil_kol` همه‌ی عضوها را
تغییر می‌دهد و `kamil` کل لیست را به یک مقدار تبدیل می‌کند. `funct(x) -> ...` یک تابع کوچک
است که همان‌جا نوشته می‌شود (یک *لامبدا*).

</div>

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

<div dir="rtl">

### ساختن یک لیست در یک خط

`[kar har x dar list]` از یک لیست، لیست تازه‌ای می‌سازد و `ke` (یا `agar`) فقط بعضی
عضوها را نگه می‌دارد؛ مثل list comprehension در پایتون:

</div>

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

<div dir="rtl">

## ۴.۳ دیکشنری‌ها

دیکشنری کلیدها را به مقدارها وصل می‌کند و ترتیب اضافه شدن کلیدها را به خاطر می‌سپارد.

</div>

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

<div dir="rtl">

دیکشنری‌ها و لیست‌ها می‌توانند تو در تو باشند؛ این دقیقاً شکل JSON است:

</div>

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

<div dir="rtl">

## ۴.۴ مجموعه‌ها

مجموعه (set) هر مقدار را فقط یک بار نگه می‌دارد.

</div>

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

<div dir="rtl">

## ۴.۵ بازه‌ها

`a..b` یعنی عددهای `a` تا `b` (خود `b` شامل نمی‌شود)؛ `a..+b` خود `b` را هم شامل می‌شود.

</div>

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

<div dir="rtl">

## تمرین‌ها

۱. یک لیست از پنج عدد بسازید و با `taghib` بزرگ‌ترینش را چاپ کنید.
۲. از `["man", "ketab", "to", "daftar"]` فقط کلمه‌هایی را نگه دارید که بیشتر از ۳ حرف دارند.
۳. یک دیکشنری از سه دوست و سنشان بسازید و هر کدام را به شکل `name: age` چاپ کنید.

</div>

<details><summary>جواب‌ها</summary>

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

<div dir="rtl">

**بعدی:** [درس ۵ — تصمیم‌ها و حلقه‌ها](05-tasmim-va-tekrar.md)

</div>
