<div dir="rtl">

# درس ۱۱ — فرانت‌اند یک وب‌سایت

> **خلاصه:** دو راه برای ساختن صفحه هست: (۱) فایل‌های معمولی HTML/CSS/JS کنار برنامه که میهن آن‌ها را می‌فرستد،
> (۲) ساختن صفحه با کلمه‌های خودِ میهن مثل `onvan`، `karte`، `dokme` بدون نوشتن HTML.

## ۱۱.۱ راه ۱ — فایل‌های واقعی HTML، CSS و جاوااسکریپت

صفحه را در پوشه‌ای به اسم `safhe` کنار برنامه‌تان بگذارید. `parvande_pasokh` یک فایل را از
آن پوشه می‌فرستد (و `/` یعنی `index.html`). صفحه با `fetch` به API شما درخواست می‌دهد،
درست همان‌طور که به هر بک‌اند دیگری درخواست می‌داد.

</div>

```
mysite/
  server.mhn
  safhe/
    index.html
    style.css
    app.js
```

```html file=safhe/index.html
<!doctype html>
<html lang="fa" dir="rtl">
<head>
  <meta charset="utf-8">
  <title>Yaddasht</title>
  <link rel="stylesheet" href="/style.css">
</head>
<body>
  <h1>Yaddasht-ha</h1>
  <form id="f"><input id="m" placeholder="yaddasht..."><button>afzoodan</button></form>
  <ul id="list"></ul>
  <script src="/app.js"></script>
</body>
</html>
```

```css file=safhe/style.css
body { font-family: system-ui, sans-serif; max-width: 32rem; margin: 3rem auto; }
li { padding: .4rem 0; border-bottom: 1px solid #ddd; }
```

```js file=safhe/app.js
async function load() {
  const items = await (await fetch('/api/yaddasht')).json();
  document.getElementById('list').innerHTML = items.map(t => `<li>${t}</li>`).join('');
}
document.getElementById('f').onsubmit = async (e) => {
  e.preventDefault();
  const m = document.getElementById('m');
  await fetch('/api/yaddasht', { method: 'POST', body: JSON.stringify(m.value) });
  m.value = '';
  load();
};
load();
```

<div dir="rtl">

و خود برنامه، یعنی بک‌اند، کوتاه است:

</div>

```mihan serve port=8711 curl=/api/yaddasht,POST:/api/yaddasht|"salam",/api/yaddasht,/style.css
yaddasht_ha = []

funct pasokh(darakht)
    agar darakht["path"] == "/api/yaddasht"
        agar darakht["method"] == "POST"
            bast yaddasht_ha, json_bekhan(darakht["matn"])
        bargard {
            "vaziat": 200,
            "matn": json_benevis(yaddasht_ha),
            "sarlawah": {"Content-Type": "application/json"}
        }
    bargard parvande_pasokh("safhe", darakht["path"])

serva_bgo(8711, pasokh)
```

```output
GET /api/yaddasht -> []
POST /api/yaddasht -> ["salam"]
GET /api/yaddasht -> ["salam"]
GET /style.css -> body { font-family: system-ui, sans-serif; max-width: 32rem; margin: 3rem auto; }
li { padding: .4rem 0; border-bottom: 1px solid #ddd; }
```

<div dir="rtl">

`http://127.0.0.1:8711/` را باز کنید و یک برنامه‌ی کامل دارید: صفحه‌ی HTML، استایل CSS،
جاوااسکریپت در مرورگر و میهن پشت همه‌شان.

## ۱۱.۲ راه ۲ — صفحه‌هایی ساخته‌شده از تکه‌های میهن

برای ابزارهای سریع اصلاً لازم نیست HTML بنویسید. میهن *تکه‌ها* دارد:

| تکه | چیست |
|---|---|
| `onvan "x"`، `onvan2 "x"` | تیتر بزرگ و تیتر کوچک‌تر |
| `band "x"` | یک پاراگراف |
| `karte [...]` | یک کارت با گوشه‌های گرد |
| `ofoghi [...]`، `amudi [...]` | تکه‌ها کنار هم / زیر هم |
| `navar [...]` | نوار لینک‌ها در بالای صفحه |
| `peyvand "matn", "/masir"` | یک لینک |
| `dokme "matn", "kar"` | یک دکمه؛ زدنش `kar` را به برنامه می‌فرستد |
| `jadval sar, satr_ha` | یک جدول |
| `liste [...]` | یک فهرست نقطه‌دار |

هر تکه می‌تواند به‌عنوان آخرین آرگومان یک دیکشنری استایل با کلمه‌های میهن بگیرد:
`"rang"` (رنگ متن)، `"pas_zamine"` (پس‌زمینه)، `"khat_dor"` (گردی گوشه‌ها)، `"fasele"`
(فاصله‌ی داخلی).

`rah(darakht, routes)` صفحه‌ی مربوط به مسیر درخواست‌شده را انتخاب می‌کند و کل صفحه‌ی HTML
را برایتان می‌سازد.

</div>

```mihan
tedad = 0

funct menu()
    bargard navar([peyvand("khane", "/"), peyvand("darbare", "/darbare")])

funct khane(darakht)
    bargard amudi([
        menu(),
        karte([
            onvan("Shomarande"),
            band($"tedad: {tedad}"),
            dokme("+1", "ziad", {"pas_zamine": "sabz", "rang": "sefid", "khat_dor": 8})
        ], {"fasele": 24, "khat_dor": 16})
    ])

funct darbare(darakht)
    bargard amudi([menu(), onvan2("darbare"), band("sakhte shode ba Mihan")])

funct pasokh(darakht)
    agar darakht["method"] == "POST"
        form_ = form_bekhan(darakht["matn"])
        agar form_["kar"] == "ziad"
            tedad += 1
    bargard rah(darakht, [
        {"path": "/", "onvan": "khane", "tabe": khane},
        {"path": "/darbare", "onvan": "darbare", "tabe": darbare}
    ])

# Try the page without a browser:
javab = pasokh({"method": "POST", "path": "/", "matn": "kar=ziad", "sorok": {}})
chap_khat javab["vaziat"]
chap_khat "tedad: 1" dar javab["matn"]
chap_khat "<!doctype html>" dar javab["matn"].kiawtar()
javab = pasokh({"method": "GET", "path": "/nist", "matn": "", "sorok": {}})
chap_khat javab["vaziat"]
```

```output
200
dorost
dorost
404
```

<div dir="rtl">

برای این‌که واقعاً سرو شود، یک خط به آخرش اضافه کنید: `serva_bgo(8080, pasokh)`.

## ۱۱.۳ از کدام راه استفاده کنم؟

* **تکه‌ها**: ابزارهای مدیریتی، داشبوردهای سریع، پروژه‌های مدرسه؛ بدون نیاز به HTML.
* **فایل‌های HTML/CSS/JS**: هر چیزی که باید دقیقاً شبیه یک طرح باشد یا از یک کتابخانه‌ی
  جاوااسکریپت استفاده کند. در این حالت میهن بک‌اند و سرور فایل شماست.

پوشه‌ی `examples/pro-sites/` سه سایت کامل و تمیز دارد که با راه دوم ساخته شده‌اند.

## تمرین‌ها

۱. به شمارنده یک دکمه‌ی `"-1"` اضافه کنید که `kam` را بفرستد.
۲. صفحه‌ای به اسم `/list` اضافه کنید که `jadval(["naam", "sen"], [["Ali", "20"], ["Sara", "25"]])` را نشان بدهد.

</div>

<details><summary>جواب‌ها</summary>

```mihan
tedad = 5

funct khane(darakht)
    bargard amudi([band($"tedad: {tedad}"), dokme("+1", "ziad"), dokme("-1", "kam")])

funct list_(darakht)
    bargard jadval(["naam", "sen"], [["Ali", "20"], ["Sara", "25"]])

funct pasokh(darakht)
    agar darakht["method"] == "POST"
        kar = form_bekhan(darakht["matn"])["kar"]
        agar kar == "ziad"
            tedad += 1
        agar kar == "kam"
            tedad -= 1
    bargard rah(darakht, [
        {"path": "/", "onvan": "khane", "tabe": khane},
        {"path": "/list", "onvan": "list", "tabe": list_}
    ])

pasokh({"method": "POST", "path": "/", "matn": "kar=kam", "sorok": {}})
chap_khat tedad
chap_khat "Sara" dar pasokh({"method": "GET", "path": "/list", "matn": "", "sorok": {}})["matn"]
```

```output
4
dorost
```

</details>

<div dir="rtl">

**بعدی:** [درس ۱۲ — یک برنامه‌ی ویندوزی](12-windows-app.md)

</div>
