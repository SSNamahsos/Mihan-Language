# Lesson 9 — Files, JSON and scripts

> **خلاصه:** با `bekhan` فایل را می‌خوانیم، با `benevis` می‌نویسیم، با `json_bekhan` / `json_benevis` با JSON کار می‌کنیم،
> و با `arghoman()` ورودی‌های خط فرمان را می‌گیریم.

## 9.1 Writing and reading files

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

Folders:

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

## 9.2 JSON

`json_benevis` turns any value into JSON text; `json_bekhan` turns JSON text back
into lists and dicts. This is how programs save data and talk to websites.

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

## 9.3 Reading what the user types

`khanak "question: "` shows a question and waits for a line of input.

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

## 9.4 Command-line arguments

`arghoman()` is the list of words written after the file name:
`mihan run jam.mhn 3 4 5`.

```mihan args=3,4,5
adad = tabdil_kol(arghoman(), funct(k) -> adad_karda(k))
chap_khat $"adad: {adad}"
chap_khat $"jam: {kamil(adad, funct(a, b) -> a + b)}"
```

```output
adad: [3, 4, 5]
jam: 12
```

## 9.5 A complete script: a word counter

Save as `kalame.mhn` and run `mihan run kalame.mhn matn.txt`.

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

## 9.6 Splitting a program into files (biar)

`biar "file.mhn"` runs another file and makes its names available.

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

## 9.7 Date and time

`zaman()` is the time in seconds (a float) — subtract two to time something.
`tarikh()` is the date as text; give it a pattern of `YYYY MM DD HH mm ss`, and
optionally a moment from `zaman()`. `sabr s` waits `s` seconds.

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

## Exercises

1. Write a script that takes a name as an argument and prints `salam <name>`.
2. Save a list of three tasks as JSON in `kar.json`, read it back and print each task.

<details><summary>Solutions</summary>

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

**Next:** [Lesson 10 — A website backend](10-backend.md)
