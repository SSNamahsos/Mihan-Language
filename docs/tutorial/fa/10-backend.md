<div dir="rtl">

# درس ۱۰ — بک‌اند یک وب‌سایت (serva_bgo)

> **خلاصه:** یک سایت در میهن فقط یک تابع است: درخواست (`darakht`) را می‌گیرد و پاسخ برمی‌گرداند.
> `serva_bgo(port, tabe)` سرور را روشن می‌کند.

## ۱۰.۱ کوچک‌ترین سرور

درخواست به‌شکل یک دیکشنری به اسم `darakht` با این کلیدها می‌رسد:

| کلید | معنی |
|---|---|
| `"method"` | `"GET"`، `"POST"`، … |
| `"path"` | مسیر، مثلاً `"/api/kala"` |
| `"sorok"` | پرس‌وجوی `?a=1&b=2` به‌صورت دیکشنری |
| `"matn"` | بدنه‌ی درخواست به‌صورت متن |
| `"sarlawah"` | هدرهای درخواست به‌صورت دیکشنری |

تابع شما یک دیکشنری برمی‌گرداند با `"vaziat"` (کد وضعیت)، `"matn"` (بدنه) و در صورت نیاز
`"sarlawah"` (هدرهای پاسخ).

</div>

```mihan serve port=8701 curl=/,/salam?naam=Sara
funct pasokh(darakht)
    agar darakht["path"] == "/salam"
        naam = darakht["sorok"]["naam"]
        bargard {"vaziat": 200, "matn": $"salam {naam}!"}
    bargard {"vaziat": 200, "matn": "khosh amadid"}

chap_khat "server roshan shod"
serva_bgo(8701, pasokh)
```

```output
server roshan shod
GET / -> khosh amadid
GET /salam?naam=Sara -> salam Sara!
```

<div dir="rtl">

با `mihan run server.mhn` اجرایش کنید و `http://127.0.0.1:8701/salam?naam=Sara` را در مرورگر
باز کنید. برای متوقف کردنش `Ctrl+C` را بزنید. (خط‌هایی که بالا با `GET` شروع می‌شوند همان
چیزی است که بررسی‌کننده‌ی آموزش هنگام باز کردن آن آدرس‌ها دریافت کرد.)

### باز شدن خودکار در مرورگر

</div>

```
mihan run --open server.mhn      # or:  mihan serve server.mhn
```

<div dir="rtl">

وقتی سرور آماده‌ی شنیدن شد، میهن (در ترمینال) `Mihan site is ready: http://...` را چاپ می‌کند
و با `--open` (یا `mihan serve`) آن آدرس را در مرورگر پیش‌فرض شما باز می‌کند: در ویندوز از
طریق باز‌کننده‌ی آدرس سیستم، در لینوکس با `xdg-open` و در macOS با `open`. `--open` را
می‌شود بعد از اسم فایل هم نوشت.

## ۱۰.۲ آزمایش منطق بدون مرورگر

چون سایت فقط یک تابع است، می‌توانید مستقیم صدایش بزنید؛ سریع‌ترین راه برای بررسی بک‌اند:

</div>

```mihan
funct pasokh(darakht)
    bargard {"vaziat": 200, "matn": "path: " + darakht["path"]}

javab = pasokh({"method": "GET", "path": "/test", "sorok": {}, "matn": ""})
chap_khat javab["vaziat"]
chap_khat javab["matn"]
```

```output
200
path: /test
```

<div dir="rtl">

## ۱۰.۳ یک API از نوع JSON همراه با داده

این یک API واقعی به سبک REST برای فهرست کارهاست. به این‌ها جواب می‌دهد:

* `GET /api/kar` → کل فهرست به‌صورت JSON
* `POST /api/kar` با `{"onvan": "..."}` → یک مورد اضافه می‌کند
* `GET /api/kar/2` → یک مورد، یا `404`

و همه چیز را در `kar.json` ذخیره می‌کند تا با راه‌اندازی دوباره‌ی سرور چیزی از دست نرود.

</div>

```mihan serve port=8702 curl=/api/kar,POST:/api/kar|{"onvan":"kharid"},POST:/api/kar|{"onvan":"varzesh"},/api/kar/2,/api/kar/9,/api/kar
parvande = "kar.json"
kar_ha = []
agar masir_hast(parvande)
    kar_ha = json_bekhan(bekhan(parvande))

funct json_pasokh(vaziat, dade)
    bargard {
        "vaziat": vaziat,
        "matn": json_benevis(dade),
        "sarlawah": {"Content-Type": "application/json"}
    }

funct pasokh(darakht)
    path = darakht["path"]
    agar path == "/api/kar" va darakht["method"] == "POST"
        jadid = json_bekhan(darakht["matn"])
        yek = {"id": tule(kar_ha) + 1, "onvan": jadid["onvan"], "anjam": ghalat}
        bast kar_ha, yek
        benevis parvande, json_benevis(kar_ha)
        bargard json_pasokh(201, yek)
    agar path == "/api/kar"
        bargard json_pasokh(200, kar_ha)
    agar path.barmigefth("/api/kar/")
        id = adad_karda(path.tab("/")[-1])
        har k dar kar_ha
            agar k["id"] == id
                bargard json_pasokh(200, k)
        bargard json_pasokh(404, {"khata": "peyda nashod"})
    bargard json_pasokh(404, {"khata": "in masir vojood nadarad"})

serva_bgo(8702, pasokh)
```

```output
GET /api/kar -> []
POST /api/kar -> {"id":1,"onvan":"kharid","anjam":false}
POST /api/kar -> {"id":2,"onvan":"varzesh","anjam":false}
GET /api/kar/2 -> {"id":2,"onvan":"varzesh","anjam":false}
GET /api/kar/9 -> {"khata":"peyda nashod"}
GET /api/kar -> [{"id":1,"onvan":"kharid","anjam":false},{"id":2,"onvan":"varzesh","anjam":false}]
```

<div dir="rtl">

می‌توانید از یک ترمینال دیگر با `curl` امتحانش کنید:

</div>

```
curl http://127.0.0.1:8702/api/kar
curl -X POST -d "{\"onvan\":\"kharid\"}" http://127.0.0.1:8702/api/kar
```

<div dir="rtl">

## ۱۰.۴ پورت 0 و آدرس

`serva_bgo(0, pasokh, bargoz)` اجازه می‌دهد کامپیوتر خودش یک پورت آزاد انتخاب کند. تابع سوم
وقتی سرور آماده‌ی شنیدن شد یک بار اجرا می‌شود و `serva_khod()` آدرس آن است:

</div>

```mihan serve
funct pasokh(darakht)
    bargard {"vaziat": 200, "matn": "ok"}

funct bargoz()
    chap_khat "amade ast"
    chap_khat serva_khod().barmigefth("http://127.0.0.1:")

serva_bgo(0, pasokh, bargoz)
```

```output
amade ast
dorost
```

<div dir="rtl">

## تمرین‌ها

۱. به سرور فهرست کارها مسیر `/api/saat` را اضافه کنید که جواب بدهد `{"vaziat": "ok"}`.
۲. سروری بنویسید که در آن `/jam?a=3&b=4` جواب بدهد `7`.

</div>

<details><summary>جواب‌ها</summary>

```mihan serve port=8703 curl=/jam?a=3&b=4,/api/saat
funct pasokh(darakht)
    agar darakht["path"] == "/jam"
        a = adad_karda(darakht["sorok"]["a"])
        b = adad_karda(darakht["sorok"]["b"])
        bargard {"vaziat": 200, "matn": tabdil(a + b, matn)}
    agar darakht["path"] == "/api/saat"
        bargard {"vaziat": 200, "matn": json_benevis({"vaziat": "ok"})}
    bargard {"vaziat": 404, "matn": "nist"}

serva_bgo(8703, pasokh)
```

```output
GET /jam?a=3&b=4 -> 7
GET /api/saat -> {"vaziat":"ok"}
```

</details>

<div dir="rtl">

**بعدی:** [درس ۱۱ — فرانت‌اند یک وب‌سایت](11-frontend.md)

</div>
