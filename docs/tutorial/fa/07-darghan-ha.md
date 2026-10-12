<div dir="rtl">

# درس ۷ — کلاس‌ها (darghan)

> **خلاصه:** `darghan` یک نوع تازه می‌سازد: داده‌ها (فیلدها) و کارهایی که روی آن انجام می‌شود (متدها).
> `khod` یعنی «خودِ همین شیء» (مثل `self` در پایتون).

## ۷.۱ اولین کلاس

</div>

```mihan
darghan Hesab
    saheb = ""
    mojudi = 0

    funct __sakht__(saheb, mojudi)
        khod.saheb = saheb
        khod.mojudi = mojudi

    funct variz(mablagh)
        khod.mojudi += mablagh

    funct bardasht(mablagh)
        agar mablagh > khod.mojudi
            chap_khat "mojudi kafi nist"
            bargard ghalat
        khod.mojudi -= mablagh
        bargard dorost

    funct gozaresh()
        chap_khat $"{khod.saheb}: {khod.mojudi} toman"

h = Hesab("Sara", 1000)
h.variz(500)
h.gozaresh()
h.bardasht(2000)
h.bardasht(300)
h.gozaresh()
chap_khat h.mojudi
```

```output
Sara: 1500 toman
mojudi kafi nist
Sara: 1200 toman
1200
```

<div dir="rtl">

* فیلدها بالای کلاس همراه با یک مقدار اولیه نوشته می‌شوند.
* `__sakht__` سازنده است: وقتی `Hesab(...)` می‌نویسید اجرا می‌شود.
* داخل یک متد، `khod.x` فیلد `x` همین شیء است.

## ۷.۲ هر شیء فیلدهای خودش را دارد

</div>

```mihan
darghan Sabad
    kala = []
    funct ezafe(x)
        bast khod.kala, x

a = Sabad()
b = Sabad()
a.ezafe("nan")
a.ezafe("shir")
b.ezafe("sib")
chap_khat a.kala
chap_khat b.kala
```

```output
["nan", "shir"]
["sib"]
```

<div dir="rtl">

## ۷.۳ ارث‌بری

`darghan Bache < Pedar` یک کلاس را روی کلاس دیگری می‌سازد. همه‌ی فیلدها و متدهای کلاس
پدر را می‌گیرد و می‌تواند هر کدام را جایگزین (override) کند.

</div>

```mihan
darghan Heyvan
    naam = ""
    funct __sakht__(naam)
        khod.naam = naam
    funct seda()
        bargard "..."
    funct moarefi()
        chap_khat $"{khod.naam} migooyad {khod.seda()}"

darghan Gorbe < Heyvan
    funct seda()
        bargard "miyoo"

darghan Sag < Heyvan
    funct seda()
        bargard "hap hap"

har h dar [Gorbe("Pishi"), Sag("Jessi"), Heyvan("Mahi")]
    h.moarefi()
```

```output
Pishi migooyad miyoo
Jessi migooyad hap hap
Mahi migooyad ...
```

<div dir="rtl">

## تمرین‌ها

۱. کلاسی به اسم `Mostatil` (مستطیل) با `arz_` و `tool` بسازید، با متدی به اسم
   `masahat()` که مساحت را برگرداند.
۲. کلاسی به اسم `Daneshjoo` با یک اسم و یک لیست نمره بسازید، با متد `nomre(n)` که یک نمره
   اضافه کند و `miangin()` که میانگین را برگرداند.

</div>

<details><summary>جواب‌ها</summary>

```mihan
darghan Mostatil
    arz_ = 0
    tool = 0
    funct __sakht__(arz_, tool)
        khod.arz_ = arz_
        khod.tool = tool
    funct masahat()
        bargard khod.arz_ * khod.tool

chap_khat Mostatil(3, 4).masahat()

darghan Daneshjoo
    naam = ""
    nomreha = []
    funct __sakht__(naam)
        khod.naam = naam
    funct nomre(n)
        bast khod.nomreha, n
    funct miangin()
        bargard kamil(khod.nomreha, funct(a, b) -> a + b) / tule(khod.nomreha)

d = Daneshjoo("Neda")
d.nomre(18)
d.nomre(15)
d.nomre(20)
chap_khat $"{d.naam}: {d.miangin()}"
```

```output
12
Neda: 17.666666666666668
```

</details>

<div dir="rtl">

**بعدی:** [درس ۸ — خطاها](08-khata-ha.md)

</div>
