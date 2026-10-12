<div dir="rtl">

# درس ۱۶ — یک API واقعی وب: مسیرها، متدها، کوکی‌ها، آپلود، رویدادهای زنده

> **خلاصه:** این درس همهٔ چیزهایی را که یک API واقعی لازم دارد نشان می‌دهد:
> `json_pasokh` برای جوابِ JSON با کدِ وضعیت، `rah` با پارامترِ مسیر (`/kala/:id`)
> و فیلترِ `method`، `sorok` و `sarlawah`، کوکی و `jalase`، آپلودِ فایل، CORS،
> قالب با `qaleb`، و رویدادهای زنده با `rooydad_ferest`.

درس ۱۰ درخواست‌ها را دستی جواب می‌داد. این درس از ابزارهایی استفاده می‌کند که یک API بزرگ‌تر
را کوتاه و درست می‌کنند. هر نمونه‌ی این درس به‌صورت سرور اجرا شده و از آن سؤال‌های واقعی
پرسیده شده؛ جواب‌های زیر هر نمونه همان چیزی است که واقعاً گفته است.

## ۱۶.۱ جواب JSON همراه با وضعیت: `json_pasokh`

`json_pasokh(x)` جوابی است که بدنه‌اش `x` نوشته‌شده به‌صورت JSON است، با نوع
`application/json`. مقدار دوم کد وضعیت است.

</div>

```mihan serve port=8461 curl=/,/nist
funct pasokh(darakht)
    agar darakht["path"] == "/"
        bargard json_pasokh({"salam": "donya", "adad": [1, 2, 3]})
    bargard json_pasokh({"khata": "peyda nashod"}, 404)

serva_bgo(8461, pasokh)
```

```output
GET / -> {"salam":"donya","adad":[1,2,3]}
GET /nist -> {"khata":"peyda nashod"}
```

<div dir="rtl">

## ۱۶.۲ مسیرهای پارامتردار و متدها

`rah(darakht, [...])` اولین مسیری را انتخاب می‌کند که `path` آن جور دربیاید. بخشی که به‌شکل
`:nam` نوشته شود با هر یک بخش از آدرس جور می‌شود و در `darakht["params"]` قرار می‌گیرد؛
مسیری که با `*` تمام شود با همه‌ی چیزهای زیرش جور می‌شود. یک مسیر می‌تواند بگوید به کدام
`method` جواب می‌دهد؛ اگر با متد دیگری درخواست شود، خودش برایتان 405 جواب می‌دهد.

</div>

```mihan serve port=8462 curl=/kala/7,/kala/7/nazar/2,POST:/kala/7|x,/hame/a/b
funct kala(darakht)
    bargard json_pasokh({"id": darakht["params"]["id"]})

funct nazar(darakht)
    p = darakht["params"]
    bargard json_pasokh({"kala": p["id"], "nazar": p["n"]})

funct baghi(darakht)
    bargard "the rest: " + darakht["path"]

funct pasokh(darakht)
    bargard rah(darakht, [
        {"path": "/kala/:id", "tabe": kala, "method": "GET"},
        {"path": "/kala/:id/nazar/:n", "tabe": nazar},
        {"path": "/hame/*", "tabe": baghi},
    ])

serva_bgo(8462, pasokh)
```

```output
GET /kala/7 -> {"id":"7"}
GET /kala/7/nazar/2 -> {"kala":"7","nazar":"2"}
POST /kala/7 -> this path does not take POST
GET /hame/a/b -> the rest: /hame/a/b
```

<div dir="rtl">

یک REST API کامل هم همین ایده است: برای هر آدرس یک funct و نگاهی به `darakht["method"]`
(GET، POST، PUT، DELETE). دیکشنری‌ها برای سمت DELETE متد `hazf(kelid)` دارند:

</div>

```mihan serve port=8463 curl=/api/kala,POST:/api/kala|{"nam":"medad"},/api/kala/2,/api/kala?q=med
kala = {"1": {"id": 1, "nam": "ketab"}}

funct hame(darakht)
    agar darakht["method"] == "POST"
        x = json_bekhan(darakht["matn"])
        x["id"] = tule(kala) + 1
        kala[neshan(x["id"])] = x
        bargard json_pasokh(x, 201)
    q = darakht["sorok"].begir("q", "")
    peyda = []
    har k dar kala.dast()
        agar kala[k]["nam"].shamil(q)
            peyda = bast(peyda, kala[k])
    bargard json_pasokh(peyda)

funct yeki(darakht)
    id = darakht["params"]["id"]
    agar darakht["method"] == "DELETE"
        kala.hazf(id)
        bargard {"vaziat": 204, "matn": ""}
    bargard json_pasokh(kala.begir(id, hich))

funct pasokh(darakht)
    bargard rah(darakht, [
        {"path": "/api/kala", "tabe": hame},
        {"path": "/api/kala/:id", "tabe": yeki},
    ])

serva_bgo(8463, pasokh)
```

```output
GET /api/kala -> [{"id":1,"nam":"ketab"}]
POST /api/kala -> {"nam":"medad","id":2}
GET /api/kala/2 -> {"nam":"medad","id":2}
GET /api/kala?q=med -> [{"nam":"medad","id":2}]
```

<div dir="rtl">

`darakht["sorok"]` پرس‌وجو (`?q=med`) است و `darakht["sarlawah"]` هدرهای درخواست. یک جواب
می‌تواند هدرهای خودش را زیر `"sarlawah"` داشته باشد.

## ۱۶.۳ CORS

صفحه‌ای از یک آدرس دیگر فقط وقتی می‌تواند جواب‌های شما را بخواند که خودتان اجازه بدهید.
برای اجازه به همه‌ی مبداها `"cors": dorost` را در جواب بگذارید، یا آدرس یک مبدا را تا فقط
همان مجاز باشد (همراه با کوکی). به سؤال `OPTIONS` مرورگر هم همین‌طور جواب بدهید:

</div>

```mihan norun
funct pasokh(darakht)
    agar darakht["method"] == "OPTIONS"
        bargard {"vaziat": 204, "matn": "", "cors": "https://example.com"}
    bargard {"vaziat": 200, "matn": "{}", "cors": "https://example.com"}

serva_bgo(8080, pasokh)
```

<div dir="rtl">

## ۱۶.۴ کوکی‌ها و جلسه‌ها

`darakht["cookie"]` کوکی‌هایی را که مرورگر فرستاده نگه می‌دارد. برای گذاشتن یک کوکی، جواب
`"cookie": {"nam": "meghdar"}` را می‌دهد، یا یک دیکشنری با `meghdar`، `umr` (ثانیه)،
`masir` و `httponly`؛ `hich` آن را پاک می‌کند.

`jalase(darakht)` یک جلسه است: دیکشنری‌ای که مال همین یک بازدیدکننده است و بین درخواست‌ها
نگه داشته می‌شود. میهن بار اول یک کوکی `mihan_jalase` به بازدیدکننده می‌دهد و به هیچ چیز
دیگری نیاز نیست:

</div>

```mihan norun
funct pasokh(darakht)
    j = jalase(darakht)
    j["bar"] = j.begir("bar", 0) + 1
    bargard {"matn": $"you have been here {j["bar"]} times", "cookie": {"akharin": darakht["path"]}}

serva_bgo(8080, pasokh)
```

<div dir="rtl">

## ۱۶.۵ فرم‌ها و آپلود فایل

فرمی که به‌صورت `application/x-www-form-urlencoded` یا `multipart/form-data` فرستاده شود
در `darakht["form"]` قرار می‌گیرد. فایل‌هایی که همراهش فرستاده شده‌اند در `darakht["parvande"]`
هستند، هر کدام یک دیکشنری با `nam` (فیلد فرم)، `esm` (اسم فایل)، `noe` (نوعش)، `andaze`
(اندازه‌اش) و `matn` (محتوایش):

</div>

```mihan norun
funct pasokh(darakht)
    agar darakht["method"] == "POST"
        har f dar darakht["parvande"]
            benevis_file("upload-" + f["esm"], f["matn"])
        bargard json_pasokh({"name": darakht["form"].begir("nam", ""), "files": tule(darakht["parvande"])}, 201)
    bargard "<form method=post enctype=multipart/form-data><input name=nam><input type=file name=f><button>send</button></form>"

serva_bgo(8080, pasokh)
```

<div dir="rtl">

## ۱۶.۶ فایل‌های ثابت: نوع‌ها و کش

`parvande_pasokh("safhe", darakht["path"])` یک فایل را با نوع درست (`text/css`،
`text/javascript`، `image/png`، …) و یک `ETag` می‌فرستد. مرورگری که فایل را از قبل دارد با
`If-None-Match` می‌پرسد و `304 Not Modified` را بدون بدنه می‌گیرد، پس صفحه‌هایی با فایل‌های
زیاد بار دوم سریع باز می‌شوند و بعد از این‌که فایلی را عوض کردید هیچ‌وقت نسخه‌ی قدیمی را نشان
نمی‌دهند.

## ۱۶.۷ قالب‌ها: `qaleb`

`qaleb(matn, dict)` جای `{{nam}}` را در یک متن پر می‌کند. چیزی که می‌گذارد برای HTML امن
می‌شود (`<` تبدیل به `&lt;` می‌شود)، پس حرف‌های یک بازدیدکننده نمی‌تواند به اسکریپت تبدیل
شود؛ `{{{nam}}}` آن را همان‌طور که هست می‌گذارد، و `{{a.b}}` به داخل یک دیکشنری دسترسی
دارد. `html_amn(x)` یک متن را به‌تنهایی امن می‌کند.

</div>

```mihan
safhe = "<h1>{{onvan}}</h1><p>{{karbar.nam}}</p>{{{tah}}}"
chap_khat(qaleb(safhe, {"onvan": "<salam>", "karbar": {"nam": "Sara"}, "tah": "<hr>"}))
chap_khat(html_amn("a < b & c"))
```

```output
<h1>&lt;salam&gt;</h1><p>Sara</p><hr>
a &lt; b &amp; c
```

<div dir="rtl">

## ۱۶.۸ رویدادهای زنده: `rooydad_ferest`

یک صفحه می‌تواند بدون این‌که مدام دوباره بپرسد، به خبرهای برنامه گوش بدهد:

</div>

```html
<script>
  new EventSource("/_mihan/rooydad/khabar").onmessage = (m) => console.log(JSON.parse(m.data));
</script>
```

<div dir="rtl">

و برنامه با `rooydad_ferest("khabar", x)` برای همه‌ی کسانی که به آن کانال گوش می‌دهند
می‌فرستد و برمی‌گرداند که چند نفر گوش می‌دادند. این‌ها server-sent events هستند: یک‌طرفه‌اند،
از برنامه به صفحه، که همان چیزی است که خبرها، پیام‌های چت و نوار پیشرفت لازم دارند. صفحه
مثل قبل با `fetch` می‌فرستد.

</div>

```mihan norun
funct pasokh(darakht)
    agar darakht["method"] == "POST"
        n = rooydad_ferest("khabar", {"matn": darakht["matn"]})
        bargard json_pasokh({"shenavande": n})
    bargard parvande_pasokh("safhe", darakht["path"])

serva_bgo(8080, pasokh)
```

<div dir="rtl">

`examples/site/namayesh/` همه‌ی این‌ها را در یک صفحه‌ی متحرک کنار هم گذاشته است.

</div>
