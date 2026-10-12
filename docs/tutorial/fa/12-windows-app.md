<div dir="rtl">

# درس ۱۲ — یک برنامه‌ی ویندوزی (baza_bgo)

> **خلاصه:** یک برنامه‌ی ویندوزی در میهن همان سایت است که در یک پنجره‌ی جدا (بدون نوار آدرس) باز می‌شود.

برنامه‌ی دسکتاپ میهن یک وب‌سرور کوچک است به‌اضافه‌ی یک پنجره که به آن اشاره می‌کند.
`baza_bgo(address, width, height)` یک پنجره‌ی بدون قاب باز می‌کند (Microsoft Edge در *حالت
برنامه*، که روی همه‌ی کامپیوترهای ویندوز ۱۰ و ۱۱ نصب است). هر چیزی که در درس‌های ۱۰ و ۱۱
یاد گرفتید بدون تغییر داخل پنجره کار می‌کند.

## ۱۲.۱ یک برنامه‌ی شمارنده

منطق برنامه در تابع‌های ساده نگه داشته می‌شود تا بشود بدون پنجره آزمایشش کرد:

</div>

```mihan
funct safhe(tedad)
    bargard por([
        "<!doctype html><html><head><meta charset='utf-8'><title>Shomarande</title>",
        "<style>body{font-family:system-ui;display:grid;place-items:center;height:100vh;margin:0;background:#14161a;color:#eee}",
        "h1{font-size:5rem;margin:0}button{font-size:1.2rem;margin:.3rem;padding:.6rem 1.2rem;border:0;border-radius:8px}</style></head>",
        "<body><div><h1>" + tedad + "</h1>",
        "<button onclick=\"location='/?kar=ziad'\">+1</button>",
        "<button onclick=\"location='/?kar=sefr'\">0</button></div></body></html>"
    ], "")

funct enjam(tedad, kar)
    agar kar == "ziad"
        bargard tedad + 1
    agar kar == "sefr"
        bargard 0
    bargard tedad

chap_khat enjam(4, "ziad")
chap_khat enjam(4, "sefr")
chap_khat "<h1>7</h1>" dar safhe(7)
```

```output
5
0
dorost
```

<div dir="rtl">

و این کل برنامه است. داخل بررسی‌کننده‌ی این آموزش اجرا نمی‌شود (صفحه‌نمایشی وجود ندارد)،
پس فقط از نظر اشتباه بررسی می‌شود؛ خودتان روی ویندوز اجرایش کنید:

</div>

```mihan norun
tedad = 0

funct safhe(tedad)
    bargard "<!doctype html><meta charset='utf-8'><body style='font-family:system-ui;text-align:center'><h1>" + tedad + "</h1><button onclick=\"location='/?kar=ziad'\">+1</button></body>"

funct pasokh(darakht)
    agar darakht["sorok"].barmadar("kar")
        tedad += 1
    bargard {"vaziat": 200, "matn": safhe(tedad), "sarlawah": {"Content-Type": "text/html; charset=utf-8"}}

funct bargoz()
    baza_bgo(serva_khod(), 420, 520)

serva_bgo(0, pasokh, bargoz)
```

```
mihan run shomarande.mhn
```

<div dir="rtl">

یک پنجره با برنامه‌ی شما باز می‌شود. برای خروج ترمینال را ببندید (یا `Ctrl+C` را بزنید).

## ۱۲.۲ دادن برنامه به کس دیگر

`mihan.exe` و فایل‌های `.mhn` خودتان (و پوشه‌ی `safhe` اگر دارید) را در یک پوشه کپی کنید.
کنارشان فایلی به اسم `start.bat` بسازید:

</div>

```
@echo off
"%~dp0mihan.exe" run "%~dp0shomarande.mhn"
```

<div dir="rtl">

با دوبار کلیک روی `start.bat` برنامه‌تان اجرا می‌شود.

## ۱۲.۳ محدودیت‌ها، صادقانه (فعلاً)

* پنجره همان Edge در حالت برنامه است، نه یک پنجره‌ی بومی Win32. به Microsoft Edge نیاز
  دارد که در ویندوز ۱۰ و ۱۱ هست.
* بستن پنجره برنامه را متوقف نمی‌کند؛ از ترمینال متوقفش کنید، یا یک دکمه‌ی «khoroj» اضافه
  کنید که `serva_band()` را صدا بزند.
* در این روش برنامه به‌صورت یک فایل `.exe` تکی ساخته نمی‌شود: `mihan.exe` را همراه
  فایل‌هایتان می‌فرستید. (برای یک فایل تکی، درس ۱۴ را ببینید.)

## تمرین

به برنامه‌ی شمارنده یک دکمه‌ی «khoroj» اضافه کنید که برنامه را متوقف کند.

</div>

<details><summary>جواب</summary>

```mihan norun
funct pasokh(darakht)
    agar darakht["path"] == "/khoroj"
        serva_band()
        bargard {"vaziat": 200, "matn": "khodahafez"}
    bargard {"vaziat": 200, "matn": "<a href='/khoroj'>khoroj</a>", "sarlawah": {"Content-Type": "text/html; charset=utf-8"}}

funct bargoz()
    baza_bgo(serva_khod(), 300, 200)

serva_bgo(0, pasokh, bargoz)
```

</details>

<div dir="rtl">

**بعدی:** [درس ۱۳ — ابزارها: test، fmt، check و ویرایشگر](13-abzar.md)

</div>
