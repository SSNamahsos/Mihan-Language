<div dir="rtl">

# درس ۱ — شروع (shoru)

> **خلاصه:** میهن زبانی است که کلمه‌هایش فارسی هستند ولی با حروف لاتین نوشته می‌شوند (فینگلیش).
> در این درس برنامه‌ی اول را می‌نویسیم و اجرا می‌کنیم.

میهن یک زبان برنامه‌نویسی است که کلمه‌های کلیدی‌اش واژه‌های فارسی‌اند که با حروف لاتین نوشته شده‌اند.
`chap` یعنی *چاپ کن*، `agar` یعنی *اگر* و `har` یعنی *برای هر*. هر کلمه‌ی کلیدی یک املای انگلیسی هم دارد
(`print`، `if`، `for`)، پس وقتی دارید یاد می‌گیرید می‌توانید آن‌ها را با هم قاطی کنید.

## ۱.۱ نصب

در ویندوز، `mihan.exe` را در جایی ثابت بگذارید (مثلاً `C:\Mihan\`) و این را اجرا کنید:

</div>

```
C:\Mihan\mihan.exe install
```

<div dir="rtl">

این دستور میهن را در پوشه‌ی کاربری شما کپی می‌کند و به `PATH` اضافه‌اش می‌کند. یک ترمینال **جدید**
باز کنید و بررسی کنید:

</div>

```
mihan version
```

<div dir="rtl">

## ۱.۲ اولین برنامه‌ی شما

فایلی به اسم `salam.mhn` بسازید و یک خط در آن بنویسید:

</div>

```mihan
chap_khat "salam, donya!"
```

```output
salam, donya!
```

<div dir="rtl">

اجرایش کنید:

</div>

```
mihan run salam.mhn
```

<div dir="rtl">

`chap_khat` («چاپ خط») یک مقدار را می‌نویسد و بعد به خط بعد می‌رود. `chap` بدون رفتن به خط
بعد می‌نویسد:

</div>

```mihan
chap "yek "
chap "do "
chap_khat "se"
chap_khat "tamam"
```

```output
yek do se
tamam
```

<div dir="rtl">

## ۱.۳ توضیح‌ها (کامنت)

هر چیزی که در یک خط بعد از `#` بیاید توضیح است و میهن آن را نادیده می‌گیرد.

</div>

```mihan
# in yek tozih ast (this is a comment)
chap_khat 2 + 3   # the answer is printed, the comment is not
```

```output
5
```

<div dir="rtl">

## ۱.۴ میهن به‌عنوان ماشین‌حساب

</div>

```mihan
chap_khat 2 + 3 * 4
chap_khat (2 + 3) * 4
chap_khat 7 / 2
chap_khat 7 ghesmat 2
chap_khat 7 % 2
chap_khat 2 ** 10
```

```output
14
20
3.5
3
1
1024
```

<div dir="rtl">

به دو قانون دقت کنید:

* `/` **همیشه** عدد اعشاری می‌دهد: `7 / 2` می‌شود `3.5`.
* `ghesmat` تقسیم صحیح است (مثل `//` در پایتون) و `%` باقی‌مانده است.

## ۱.۵ محیط REPL

`mihan repl` به شما اجازه می‌دهد خط‌به‌خط تایپ کنید و جواب را همان لحظه ببینید؛ راهی عالی
برای آزمایش کردن. برای خروج `exit` را بنویسید.

</div>

```
mihan> 2 + 2
4
mihan> naam = "Sara"
naam = "Sara"
mihan> exit
```

<div dir="rtl">

## تمرین‌ها

۱. اسم خودتان را در یک خط و اسم شهرتان را در خط بعد چاپ کنید.
۲. چاپ کنید که یک هفته چند دقیقه است.
۳. حاصل تقسیم `7` بر `2` را هم به‌صورت اعشاری و هم به‌صورت عدد صحیح چاپ کنید.

</div>

<details><summary>جواب‌ها</summary>

```mihan
chap_khat "Sara"
chap_khat "Tehran"
chap_khat 7 * 24 * 60
chap_khat 7 / 2
chap_khat 7 ghesmat 2
```

```output
Sara
Tehran
10080
3.5
3
```

</details>

<div dir="rtl">

**بعدی:** [درس ۲ — مقدارها و متغیرها](02-arzesh-ha.md)

</div>
