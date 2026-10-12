<div dir="rtl">

# درس ۳ — متن (matn)

> **خلاصه:** متن بین دو `"` نوشته می‌شود. با `$"..."` می‌توان مقدار را داخل متن گذاشت.

## ۳.۱ چسباندن و پر کردن

`+` متن‌ها را به هم می‌چسباند، حتی با عدد. یک `$` جلوی گیومه‌ها اجازه می‌دهد هر عبارتی را
داخل `{ }` بگذارید:

</div>

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

<div dir="rtl">

بدون `$`، آکولادها فقط چند حرف معمولی‌اند: `"salam {naam}"` دقیقاً همان را چاپ می‌کند.

## ۳.۲ حرف‌های ویژه

</div>

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

<div dir="rtl">

## ۳.۳ طول، جایگاه‌ها و تکه‌ها

</div>

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

<div dir="rtl">

جایگاه‌ها از `0` شروع می‌شوند و `-1` آخرین حرف است. هر حرف فارسی یک حرف حساب می‌شود.

## ۳.۴ متدها

متد با یک نقطه روی متنی که جلویش است صدا زده می‌شود.

</div>

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

<div dir="rtl">

| متد | معنی |
|---|---|
| `jabeja()` | حذف فاصله‌های دو طرف (trim) |
| `gooshband()` / `kiawtar()` | حروف بزرگ / حروف کوچک |
| `tab(",")` | شکستن به یک لیست |
| `jigarre(a, b)` | جایگزین کردن |
| `shamil(x)` | آیا شامل است؟ |
| `barmigefth(x)` / `bapakhsh(x)` | با ... شروع می‌شود / با ... تمام می‌شود |
| `por(list, sep)` | چسباندن یک لیست به‌صورت متن |

## ۳.۵ عددها در یک قالب

</div>

```mihan
chap_khat katoos(3.14159, "%0.2f")
chap_khat katoos(7, "%03d")
```

```output
3.14
007
```

<div dir="rtl">

## تمرین‌ها

۱. با داشتن `esm = "reza"`، `REZA` و تعداد حروف را چاپ کنید.
۲. `"sib,moz,porteghal"` را بشکنید و میوه‌ی دوم را چاپ کنید.
۳. از عدد `12.5`، متن `Gheymat: 12.50 toman` را چاپ کنید.

</div>

<details><summary>جواب‌ها</summary>

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

<div dir="rtl">

**بعدی:** [درس ۴ — لیست‌ها، دیکشنری‌ها و مجموعه‌ها](04-majmooe-ha.md)

</div>
