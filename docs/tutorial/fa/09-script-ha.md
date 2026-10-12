<div dir="rtl">

# درس ۹ — فایل‌ها، JSON و اسکریپت‌ها

> **خلاصه:** با `bekhan` فایل را می‌خوانیم، با `benevis` می‌نویسیم، با `json_bekhan` / `json_benevis` با JSON کار می‌کنیم،
> و با `arghoman()` ورودی‌های خط فرمان را می‌گیریم.

## ۹.۱ نوشتن و خواندن فایل‌ها

</div>

```mihan
benevis "yaddasht.txt", "khat aval\nkhat dovom\n"
matn_ = bekhan("yaddasht.txt")
chap_khat matn_.jabeja()
khatha = matn_.jabeja().tab("\n")
chap_khat tule khatha
chap_khat masir_hast("yaddasht.txt")
hazf "yaddasht.txt"
chap_khat masir_hast("yaddasht.txt")
```

```output
khat aval
khat dovom
2
dorost
ghalat
```

<div dir="rtl">

پوشه‌ها:

</div>

```mihan
pooshe "dade/aks"
benevis masir("dade", "a.txt"), "salam"
chap_khat fehrest("dade")
chap_khat bekhan(masir("dade", "a.txt"))
```

```output
["a.txt", "aks"]
salam
```

<div dir="rtl">

## ۹.۲ JSON

`json_benevis` هر مقداری را به متن JSON تبدیل می‌کند و `json_bekhan` متن JSON را دوباره به
لیست و دیکشنری برمی‌گرداند. برنامه‌ها این‌طوری داده ذخیره می‌کنند و با سایت‌ها حرف می‌زنند.

</div>

```mihan
dade = {"naam": "Ali", "nomreha": [18, 19.5], "fa'al": dorost, "address": hich}
matn_ = json_benevis(dade)
chap_khat matn_
benevis "dade.json", matn_

khande = json_bekhan(bekhan("dade.json"))
chap_khat khande["nomreha"][1]
chap_khat khande["fa'al"]
```

```output
{"naam":"Ali","nomreha":[18,19.5],"fa'al":true,"address":null}
19.5
dorost
```

<div dir="rtl">

## ۹.۳ خواندن چیزی که کاربر تایپ می‌کند

`khanak "question: "` یک سؤال نشان می‌دهد و منتظر می‌ماند تا کاربر یک خط بنویسد.

</div>

```mihan stdin=Sara|25
naam = khanak "esm-et chie? "
sen = adad_karda(khanak "chand salete? ")
chap_khat ""
chap_khat $"salam {naam}! sale dige {sen + 1} sale mishi."
```

```output
esm-et chie? chand salete? 
salam Sara! sale dige 26 sale mishi.
```

<div dir="rtl">

## ۹.۴ آرگومان‌های خط فرمان

`arghoman()` لیست کلمه‌هایی است که بعد از اسم فایل نوشته شده‌اند:
`mihan run jam.mhn 3 4 5`.

</div>

```mihan args=3,4,5
adad = tabdil_kol(arghoman(), funct(k) -> adad_karda(k))
chap_khat $"adad: {adad}"
chap_khat $"jam: {kamil(adad, funct(a, b) -> a + b)}"
```

```output
adad: [3, 4, 5]
jam: 12
```

<div dir="rtl">

## ۹.۵ یک اسکریپت کامل: شمارنده‌ی کلمه‌ها

با اسم `kalame.mhn` ذخیره‌اش کنید و `mihan run kalame.mhn matn.txt` را اجرا کنید.

</div>

```text file=matn.txt
salam donya salam Mihan
Mihan zaban-e farsi ast
```

```mihan args=matn.txt
agar tule(arghoman()) == 0
    chap_khat "estefade: mihan run kalame.mhn <file>"
    khoroj 1

shomaresh = {}
har khat dar bekhan(arghoman(0)).tab("\n")
    har k dar khat.tab(" ")
        k = k.jabeja().kiawtar()
        agar k == ""
            edame
        agar k dar shomaresh
            shomaresh[k] += 1
        warna
            shomaresh[k] = 1

har k dar dast(shomaresh)
    chap_khat $"{k}: {shomaresh[k]}"
```

```output
salam: 2
donya: 1
mihan: 2
zaban-e: 1
farsi: 1
ast: 1
```

<div dir="rtl">

## ۹.۶ تقسیم برنامه به چند فایل (biar)

`biar "file.mhn"` فایل دیگری را اجرا می‌کند و اسم‌هایش را در دسترس می‌گذارد.

</div>

```mihan file=abzar.mhn
funct dobrabar(x)
    bargard x * 2
```

```mihan
biar "abzar.mhn"
chap_khat dobrabar(21)
```

```output
42
```

<div dir="rtl">

## ۹.۷ تاریخ و زمان

`zaman()` زمان بر حسب ثانیه است (یک عدد اعشاری)؛ دو تا از آن را از هم کم کنید تا مدت یک
کار را اندازه بگیرید. `tarikh()` تاریخ را به‌صورت متن می‌دهد؛ می‌توانید یک الگو از
`YYYY MM DD HH mm ss` و در صورت تمایل یک لحظه از `zaman()` به آن بدهید. `sabr s` به اندازه‌ی
`s` ثانیه صبر می‌کند.

</div>

```mihan norun
chap_khat tarikh()                  # e.g. 2026-10-12 19:30:05
chap_khat tarikh("DD/MM/YYYY")      # e.g. 12/10/2026
```

```mihan
shoru = zaman()
sabr 0.2
chap_khat zaman() - shoru >= 0.2
chap_khat tarikh("YYYY-MM-DD", 31579200)
```

```output
dorost
1971-01-01
```

<div dir="rtl">

## تمرین‌ها

۱. اسکریپتی بنویسید که یک اسم را به‌عنوان آرگومان بگیرد و `salam <name>` را چاپ کند.
۲. یک لیست از سه کار را به‌صورت JSON در `kar.json` ذخیره کنید، دوباره بخوانیدش و هر کار را چاپ کنید.

</div>

<details><summary>جواب‌ها</summary>

```mihan args=Reza
chap_khat "salam " + arghoman(0)
```

```output
salam Reza
```

```mihan
kar_ha = ["kharid", "varzesh", "dars"]
benevis "kar.json", json_benevis(kar_ha)
har k dar json_bekhan(bekhan("kar.json"))
    chap_khat "- " + k
```

```output
- kharid
- varzesh
- dars
```

</details>

<div dir="rtl">

**بعدی:** [درس ۱۰ — بک‌اند یک وب‌سایت](10-backend.md)

</div>
