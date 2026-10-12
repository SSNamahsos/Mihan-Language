# Lesson 7 — Classes (darghan)

> **خلاصه:** `darghan` یک نوع تازه می‌سازد: داده‌ها (فیلدها) و کارهایی که روی آن انجام می‌شود (متدها).
> `khod` یعنی «خودِ همین شیء» (مثل `self` در پایتون).

## 7.1 A first class

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

* Fields are written at the top of the class with a starting value.
* `__sakht__` is the constructor: it runs when you write `Hesab(...)`.
* Inside a method, `khod.x` is this object's field `x`.

## 7.2 Every object has its own fields

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

## 7.3 Inheritance

`darghan Bache < Pedar` builds a class on top of another one. It gets every
field and method of the parent, and can replace (override) any of them.

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

## Exercises

1. Make a class `Mostatil` (rectangle) with `arz_` and `tool`, and a method
   `masahat()` that returns the area.
2. Make a class `Daneshjoo` with a name and a list of grades, a method
   `nomre(n)` that adds a grade, and `miangin()` that returns the average.

<details><summary>Solutions</summary>

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

**Next:** [Lesson 8 — Errors](08-khata-ha.md)
