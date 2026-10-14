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

</div>


<div dir="rtl">

## ۱۱.۳ همه‌ی تکه‌ها، و جعبه‌هایی که خودتان می‌سازید

هر تابعِ تکه (`onvan`، `band`، `karte` و …) فقط یک **dict** می‌سازد. می‌توانید آن dict را خودتان
بنویسید: `"gone"` می‌گوید تکه از چه نوعی است، `"matn"` نوشته‌ی آن است و `"darun"` تکه‌های داخلش را
نگه می‌دارد. این‌طوری به تکه‌هایی می‌رسید که تابعِ جدا ندارند، مثل عکس یا یک عنوانِ کوچک‌تر.

| `"gone"` | تبدیل می‌شود به | کلیدهای اضافه |
|---|---|---|
| `jabe` | یک جعبه (`div`) | `darun` |
| `band`، `onvan`، `onvan2`، `onvan3` | پاراگراف و عنوان‌ها | `matn` |
| `peyvand` | پیوند | `be` = نشانی |
| `tasvir` | عکس | `be` = نشانیِ عکس، `matn` = توضیحِ آن |
| `khat` | خطِ جداکننده | |
| `fehrest`، `radif` | فهرست و یک ردیفِ آن | |
| `jadval`، `satr`، `sarnevesht`، `khane_jadval` | جدول، ردیف، خانه‌ی سرستون، خانه | |
| `form`، `vorudi`، `dokme` | فرم، ورودی، دکمه | بخش ۱۱.۴ |
| `navar` | نوارِ پیوندها | |

</div>

```mihan
safhe = safhe_kon("Gallery", amudi([
    {"gone": "onvan3", "matn": "aks-ha"},
    {"gone": "tasvir", "be": "/aks/gol.png", "matn": "yek gol", "pahna": 200, "khat_dor": 12},
    {"gone": "khat"},
    band("vasat", {"chid": "vasat", "rang": "abi", "andaze": 20, "vazn": "bold"})
], {"fasele": 16, "fasele_khane": 8}))
har khat_ dar safhe.tab("\n")
    agar "<img" dar khat_ ya "<h3" dar khat_ ya "<hr" dar khat_ ya "<p " dar khat_
        chap_khat khat_
```

```output
<div style="display:flex;flex-direction:column;gap:8px;padding:16px;"><h3>aks-ha</h3>
<img style="width:200px;border-radius:12px;" src="/aks/gol.png" alt="yek gol">
<hr>
<p style="text-align:center;color:#2563eb;font-size:20px;font-weight:bold;">vasat</p>
```

<div dir="rtl">

### واژه‌های سبک

آخرین ورودیِ هر تکه (یا هر کلید در dictِ یک جعبه) می‌تواند یک واژه‌ی سبک باشد، مثل `rang`، `zamine`، `gooshe_gerd` یا `fasele_dakheli`. عددها برای اندازه پیکسل هستند. بخشِ ۱۱.۵ همه‌ی واژه‌ها و مقدارها را آورده است.

## ۱۱.۴ رویدادها: دکمه‌ها و فرم‌ها

صفحه‌هایی که از تکه ساخته شده‌اند با برنامه‌ی شما حرف می‌زنند. یک `dokme` که `kar` دارد، وقتی فشرده
شود یک POST با `kar=…` به همان نشانی می‌فرستد و صفحه با پاسخِ تازه‌ی شما عوض می‌شود؛ همان شمارنده‌ی
بخش ۱۱.۲. یک `form` همه‌ی `vorudi`های داخلش را می‌فرستد و `form_bekhan` آنچه رسیده را به dict تبدیل می‌کند.

</div>

```mihan
salam_be = "doost"

funct khane(darakht)
    bargard karte([
        onvan2($"salam, {salam_be}!"),
        form("/", [
            vorudi("naam", {"pahna": 220}),
            dokme("ferestadan")
        ])
    ], {"fasele": 20})

funct pasokh(darakht)
    agar darakht["method"] == "POST"
        salam_be = form_bekhan(darakht["matn"])["naam"]
    bargard rah(darakht, [{"path": "/", "onvan": "salam", "tabe": khane}])

javab = pasokh({"method": "POST", "path": "/", "matn": "naam=Sara", "sorok": {}})
chap_khat "salam, Sara!" dar javab["matn"]
chap_khat "<input" dar javab["matn"]
```

```output
dorost
dorost
```

<div dir="rtl">

`jalase(darakht)` به هر بازدیدکننده یک dictِ مخصوصِ خودش می‌دهد که با یک کوکی میان درخواست‌ها
نگه داشته می‌شود، تا دو نفر یک شمارنده را با هم شریک نشوند (درس ۱۶).

## ۱۱.۵ سبک به فینگلیش: واژه‌های سبک، `sabk_khat` و `sabk`

هر جایی در میهن که سبک می‌گیرد، همین واژه‌های فینگلیش را می‌گیرد: دیکشنری سبکِ آخرِ یک تکه، کلیدهای دیکشنریِ یک جعبه، `sabk_khat` (یک متنِ `style="..."`) و `sabk` (یک برگه‌ی سبکِ کامل). به جای `color` بنویسید `rang`، به جای `border-radius` بنویسید `gooshe_gerd` و به جای `black` بنویسید `siah`. عددِ تنها برای اندازه یعنی پیکسل؛ واحدهای دیگر (`%`، `rem`، `em`، `vh`، `s`، `deg`) را درست مثل CSS در متن بنویسید.

```mihan
chap_khat sabk_khat({"rang": "siah", "gooshe_gerd": 30, "hashie": "1px saf khakestari"})
chap_khat sabk_khat({"zamine": "sormei", "fasele_dakheli": "8px 16px", "andaze_matn": 18, "vazn_matn": "porrang"})
```

```output
color:black;border-radius:30px;border:1px solid #6b7280;
background:#1e3a8a;padding:8px 16px;font-size:18px;font-weight:bold;
```

نام‌های انگلیسیِ CSS هم کار می‌کنند، با `-` یا `_`؛ پس می‌توانید سبکی را از هر جا کپی کنید و دو زبان را با هم بیاورید:

```mihan
chap_khat sabk_khat({"color": "black", "border-radius": 30, "rang_zamine": "zard"})
```

```output
color:black;border-radius:30px;background-color:#eab308;
```

### چیدمان و جای‌گذاری

```mihan
chap_khat sabk_khat({"nemayesh": "felex", "jahat": "sotun", "chidman": "faseledar", "hamtarazi": "markaz", "fasele_beyn": 12})
chap_khat sabk_khat({"nemayesh": "shabake", "sotun_ha": "tekrar(3, 1fr)", "fasele_beyn": 16})
chap_khat sabk_khat({"mogheiyat": "sabet", "az_payin": 20, "az_rast": 20, "laye": 10})
```

```output
display:flex;flex-direction:column;justify-content:space-between;align-items:center;gap:12px;
display:grid;grid-template-columns:repeat(3, 1fr);gap:16px;
position:fixed;bottom:20px;right:20px;z-index:10;
```

### جلوه و حرکت

داخلِ مقدار، نامِ تابع‌ها هم فینگلیش است: `charkhesh` (rotate)، `bozorgi` (scale)، `jabejayi_y` (translateY)، `tari` (blur)، `shib` (linear-gradient)، `aks` (url) و بقیه‌ی جدولِ پایین. در `enteghal` می‌توانید نامِ واژه‌ی سبکی را بیاورید که باید نرم تغییر کند.

```mihan
chap_khat sabk_khat({"tabdil": "charkhesh(10deg) bozorgi(1.1)", "enteghal": "tabdil 0.3s narm", "safi": "tari(4px)", "shafafiat": 0.9, "nashangar": "dast"})
```

```output
transform:rotate(10deg) scale(1.1);transition:transform 0.3s ease;filter:blur(4px);opacity:0.9;cursor:pointer;
```

### یک برگه‌ی سبکِ کامل با `sabk`

`sabk` یک دیکشنری از انتخابگرها را به متنِ CSS تبدیل می‌کند. نامِ تکه‌ها انتخابگر هستند (`onvan` یعنی `h1`، `dokme` یعنی `button`، `jabe` یعنی `div`، `badane` یعنی `body`) و حالت‌های فینگلیش هم: `:rooye` (hover)، `:fokus`، `:faal` (active)، `:aval`، `:akhar`، `::ghabl`، `::baad`. کلیدهای ویژه:

- `"@harekat nam"`: یک انیمیشن، با قاب‌های `az` (from)، `ta` (to) یا `"50%"`.
- `"@kochak_tar 600"` / `"@bozorg_tar 900"`: فقط روی صفحه‌های باریک‌تر / پهن‌تر از آن.
- `"@kam_harekat"`: برای بازدیدکننده‌ای که حرکتِ کمتر خواسته. `"@tarik"` / `"@roshan"`: پوسته‌ی تیره یا روشنِ سیستم.

```mihan
zaher = sabk({
    "badane": {"zamine": "kerem", "ghalam": "Vazirmatn, sans-serif"},
    "@harekat biya": {
        "az": {"shafafiat": 0, "tabdil": "jabejayi_y(20px)"},
        "ta": {"shafafiat": 1, "tabdil": "hich"}
    },
    "onvan": {"rang": "sormei", "harekat": "biya 0.8s narm har_do"},
    ".kart": {"zamine": "sefid", "gooshe_gerd": 12, "fasele_dakheli": 16, "enteghal": "tabdil 0.2s"},
    ".kart:rooye": {"tabdil": "bozorgi(1.05)"},
    "@kochak_tar 600": {".kart": {"fasele_dakheli": 8}},
    "@kam_harekat": {"onvan": {"harekat": "hich"}}
})
chap zaher
```

```output
body {background:#fef3c7;font-family:Vazirmatn, sans-serif;}
@keyframes biya {
from {opacity:0;transform:translateY(20px);}
to {opacity:1;transform:none;}
}
h1 {color:#1e3a8a;animation:biya 0.8s ease both;}
.kart {background:white;border-radius:12px;padding:16px;transition:transform 0.2s;}
.kart:hover {transform:scale(1.05);}
@media (max-width:600px) {
.kart {padding:8px;}
}
@media (prefers-reduced-motion:reduce) {
h1 {animation:none;}
}
```

برگه را به عنوانِ آرگومانِ سوم به `safhe_kon` بدهید، و با `kelas` به تکه کلاس بدهید (یا با `shenase` شناسه) تا برگه پیدایش کند:

```mihan
zaher = {
    ".kart": {"zamine": "sefid", "gooshe_gerd": 12, "fasele_dakheli": 16, "saye": "0 2px 8px khakestari"},
    ".kart:rooye": {"tabdil": "jabejayi_y(-4px)"}
}
safhe = safhe_kon("mive-ha", jabe([
    onvan("mive-ha", {"rang": "sabz"}),
    band("sib", {"kelas": "kart"}),
    band("anar", {"kelas": "kart", "rang": "ghermez"})
]), zaher)
chap_khat ".kart:hover" dar safhe
chap_khat "class=\"kart\"" dar safhe
chap_khat "color:#dc2626" dar safhe
```

```output
dorost
dorost
dorost
```

واژه‌ای که میهن نمی‌شناسد برنامه را نگه می‌دارد و می‌گوید احتمالاً منظورتان کدام واژه بوده:

```mihan fails
chap_khat sabk_khat({"gooshe_gard": 12})
```

```output
khata: E4350
  in kalame-ye sabk ro nemishnasam: gooshe_gard; shayad manzooret 'gooshe_gerd' bood?
   1 | chap_khat sabk_khat({"gooshe_gard": 12})
                 ^
```

### فهرستِ کامل

**رنگ و زمینه**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `rang` |  | `color` |
| `zamine` | `pas_zamine` | `background` |
| `rang_zamine` |  | `background-color` |
| `tasvir_zamine` |  | `background-image` |
| `andaze_zamine` |  | `background-size` |
| `jaye_zamine` |  | `background-position` |

**حاشیه و گوشه**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `hashie` | `khat` | `border` |
| `rang_hashie` | `khat_rang` | `border-color` |
| `pahnaye_hashie` |  | `border-width` |
| `sabk_hashie` |  | `border-style` |
| `hashie_bala` |  | `border-top` |
| `hashie_payin` |  | `border-bottom` |
| `hashie_chap` |  | `border-left` |
| `hashie_rast` |  | `border-right` |
| `gooshe_gerd` | `khat_dor` | `border-radius` |
| `khat_birooni` |  | `outline` |

**فاصله**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `fasele_dakheli` | `fasele` | `padding` |
| `fasele_dakheli_bala` |  | `padding-top` |
| `fasele_dakheli_payin` |  | `padding-bottom` |
| `fasele_dakheli_chap` |  | `padding-left` |
| `fasele_dakheli_rast` |  | `padding-right` |
| `fasele_birooni` | `fasele_birun` | `margin` |
| `fasele_birooni_bala` |  | `margin-top` |
| `fasele_birooni_payin` |  | `margin-bottom` |
| `fasele_birooni_chap` |  | `margin-left` |
| `fasele_birooni_rast` |  | `margin-right` |
| `fasele_beyn` | `fasele_khane` | `gap` |
| `fasele_radif` |  | `row-gap` |
| `fasele_sotun` |  | `column-gap` |

**اندازه**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `pahna` |  | `width` |
| `ertefa` | `bala` | `height` |
| `kamtarin_pahna` |  | `min-width` |
| `bishtarin_pahna` |  | `max-width` |
| `kamtarin_ertefa` |  | `min-height` |
| `bishtarin_ertefa` |  | `max-height` |
| `andaze_jabe` |  | `box-sizing` |

**متن و قلم**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `ghalam` | `khane` | `font-family` |
| `andaze_matn` | `andaze` | `font-size` |
| `vazn_matn` | `vazn` | `font-weight` |
| `sabk_matn` |  | `font-style` |
| `tarazbandi` | `chid` | `text-align` |
| `ertefa_khat` |  | `line-height` |
| `fasele_harf` |  | `letter-spacing` |
| `tazin_matn` |  | `text-decoration` |
| `halat_harf` |  | `text-transform` |
| `saye_matn` |  | `text-shadow` |
| `shekastan_kalame` |  | `word-break` |
| `faza_khali` |  | `white-space` |
| `jahat_matn` |  | `direction` |

**چیدمان: فلکس و گرید**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `nemayesh` |  | `display` |
| `jahat` |  | `flex-direction` |
| `pichesh` |  | `flex-wrap` |
| `chidman` |  | `justify-content` |
| `hamtarazi` |  | `align-items` |
| `hamtarazi_khod` |  | `align-self` |
| `hamtarazi_mohtava` |  | `align-content` |
| `felex` |  | `flex` |
| `roshd` |  | `flex-grow` |
| `kahesh` |  | `flex-shrink` |
| `paye` |  | `flex-basis` |
| `tartib` |  | `order` |
| `sotun_ha` |  | `grid-template-columns` |
| `radif_ha` |  | `grid-template-rows` |
| `nahie_ha` |  | `grid-template-areas` |
| `nahie` |  | `grid-area` |
| `sotun_shabake` |  | `grid-column` |
| `radif_shabake` |  | `grid-row` |
| `jaygozari` |  | `place-items` |

**جای‌گذاری و لایه**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `mogheiyat` |  | `position` |
| `az_bala` |  | `top` |
| `az_payin` |  | `bottom` |
| `az_chap` |  | `left` |
| `az_rast` |  | `right` |
| `laye` |  | `z-index` |
| `sarriz` |  | `overflow` |
| `sarriz_ofoghi` |  | `overflow-x` |
| `sarriz_amudi` |  | `overflow-y` |
| `shenavar` |  | `float` |

**جلوه و حرکت**

| میهن | نام قدیمی‌تر | CSS |
|---|---|---|
| `saye` |  | `box-shadow` |
| `shafafiat` | `shafaf` | `opacity` |
| `didani` |  | `visibility` |
| `safi` |  | `filter` |
| `safi_posht` |  | `backdrop-filter` |
| `enteghal` |  | `transition` |
| `tabdil` |  | `transform` |
| `mabda_tabdil` |  | `transform-origin` |
| `harekat` |  | `animation` |
| `nam_harekat` |  | `animation-name` |
| `modat_harekat` |  | `animation-duration` |
| `takhir_harekat` |  | `animation-delay` |
| `tekrar_harekat` |  | `animation-iteration-count` |
| `nashangar` |  | `cursor` |
| `lams` |  | `pointer-events` |
| `entekhab` |  | `user-select` |
| `jaygiri_tasvir` |  | `object-fit` |
| `nesbat` |  | `aspect-ratio` |
| `neshane_fehrest` |  | `list-style` |
| `mohtava` |  | `content` |

**رنگ‌ها**

| میهن | CSS | میهن | CSS | میهن | CSS |
|---|---|---|---|---|---|
| `abi` | `#2563eb` | `abi_aseman` | `#38bdf8` | `banafsh` | `#7c3aed` |
| `banafshe` | `#7c3aed` | `firoozei` | `#06b6d4` | `ghahvei` | `#92400e` |
| `ghermez` | `#dc2626` | `ghermez_tire` | `#7f1d1d` | `hamrang` | `currentColor` |
| `kerem` | `#fef3c7` | `khakestari` | `#6b7280` | `meshki` | `black` |
| `narenji` | `#f97316` | `noghrei` | `#c0c0c0` | `sabz` | `#16a34a` |
| `sabz_tire` | `#14532d` | `sefid` | `white` | `shafaf` | `transparent` |
| `siah` | `black` | `soorati` | `#ec4899` | `sormei` | `#1e3a8a` |
| `talayi` | `#d4af37` | `zard` | `#eab308` | `zereshki` | `#9f1239` |

**واژه‌های مقدار**

| میهن | CSS | میهن | CSS | میهن | CSS |
|---|---|---|---|---|---|
| `adi` | `normal` | `bala` | `top` | `baraks` | `reverse` |
| `bi_payan` | `infinite` | `bi_pich` | `nowrap` | `boluk` | `block` |
| `boluk_khati` | `inline-block` | `bozorg_harf` | `uppercase` | `chap` | `left` |
| `chap_be_rast` | `ltr` | `chasban` | `sticky` | `dast` | `pointer` |
| `dayere` | `circle` | `dor_ta_dor` | `space-around` | `dotayi` | `double` |
| `entezar` | `wait` | `ers` | `inherit` | `faseledar` | `space-between` |
| `felex` | `flex` | `hame` | `all` | `har_do` | `both` |
| `hich` | `none` | `istade` | `static` | `jabeja` | `move` |
| `jash_bede` | `contain` | `jolo` | `forwards` | `kaj` | `italic` |
| `keshide` | `stretch` | `khat_chin` | `dashed` | `khati` | `inline` |
| `khodkar` | `auto` | `kochak_harf` | `lowercase` | `lagzan` | `scroll` |
| `makhfi` | `hidden` | `mamnoo` | `not-allowed` | `markaz` | `center` |
| `mojood` | `border-box` | `mosavi` | `space-evenly` | `motlagh` | `absolute` |
| `narm` | `ease` | `narm_dotarafe` | `ease-in-out` | `narm_khoruj` | `ease-out` |
| `narm_vorud` | `ease-in` | `nazok` | `300` | `nesbi` | `relative` |
| `neshangar_matn` | `text` | `noghte` | `disc` | `noghte_chin` | `dotted` |
| `payan` | `flex-end` | `payin` | `bottom` | `peida` | `visible` |
| `pichide` | `wrap` | `pishfarz` | `default` | `por` | `100%` |
| `por_kon` | `cover` | `porrang` | `bold` | `radif` | `row` |
| `radif_baraks` | `row-reverse` | `raft_bargasht` | `alternate` | `rast` | `right` |
| `rast_be_chap` | `rtl` | `roo_khat` | `line-through` | `sabet` | `fixed` |
| `saf` | `solid` | `shabake` | `grid` | `shoru` | `flex-start` |
| `sotun` | `column` | `sotun_baraks` | `column-reverse` | `vasat` | `center` |
| `yeknavakht` | `linear` | `zakhim` | `bold` | `zir_khat` | `underline` |

**تابع‌های داخل مقدار**

| میهن | CSS | میهن | CSS | میهن | CSS |
|---|---|---|---|---|---|
| `aks()` | `url()` | `ashba()` | `saturate()` | `bishtarin()` | `max()` |
| `bozorgi()` | `scale()` | `charkhesh()` | `rotate()` | `jabejayi()` | `translate()` |
| `jabejayi_x()` | `translateX()` | `jabejayi_y()` | `translateY()` | `kaji()` | `skew()` |
| `kamtarin()` | `min()` | `khakestari()` | `grayscale()` | `mahdood()` | `clamp()` |
| `mohasebe()` | `calc()` | `motaghayer()` | `var()` | `roshanai()` | `brightness()` |
| `saye()` | `drop-shadow()` | `shib()` | `linear-gradient()` | `shib_dayere()` | `radial-gradient()` |
| `tari()` | `blur()` | `tekrar()` | `repeat()` |  |  |

## ۱۱.۶ طراحی و انیمیشنِ خودتان با `qaleb`

تکه‌ها برای صفحه‌های سریع‌اند. برای صفحه‌ای که حرکت دارد، HTML و CSS را به شکل قالب بنویسید و
بگذارید میهن پرش کند: `qaleb` هر `{{nam}}` را از dict امن‌شده در متن می‌گذارد و `{{{nam}}}` را
همان‌طور که هست. انیمیشن‌های CSS، افکت‌های hover و transitionها همه در `<style>`ِ قالب می‌آیند.

</div>

```mihan
zaher = sabk({
    "@harekat biya": {"az": {"shafafiat": 0, "tabdil": "jabejayi_y(20px)"}, "ta": {"shafafiat": 1}},
    "onvan": {"harekat": "biya 0.8s narm har_do"},
    ".karte": {"enteghal": "tabdil 0.2s"},
    ".karte:rooye": {"tabdil": "bozorgi(1.05)"}
})
qaleb_safhe = por([
    "<!doctype html><html lang=\"fa\" dir=\"rtl\"><head><meta charset=\"utf-8\"><title>{{onvan}}</title><style>{{{zaher}}}</style></head>",
    "<body><h1>{{onvan}}</h1>{{{kart_ha}}}</body></html>"
], "\n")

kart_ha = ""
har mive dar ["sib", "moz", "<anar>"]
    kart_ha += $"<div class=\"karte\">{html_amn(mive)}</div>"

safhe = qaleb(qaleb_safhe, {"onvan": "mive-ha", "kart_ha": kart_ha, "zaher": zaher})
chap_khat "animation:biya" dar safhe
chap_khat "&lt;anar&gt;" dar safhe
```

```output
dorost
dorost
```

<div dir="rtl">

`html_amn` متن را برای گذاشتن در HTML امن می‌کند تا بازدیدکننده نتواند تگ تزریق کند.

## ۱۱.۷ صدا زدنِ API

**از صفحه.** با `json_pasokh` پاسخِ JSON بدهید و جاوااسکریپتِ صفحه با `fetch` صدایش بزند (مثل ۱۱.۱).

</div>

```mihan
funct pasokh(darakht)
    agar darakht["path"] == "/api/saat"
        bargard json_pasokh({"saat": "12:00", "ok": dorost})
    bargard json_pasokh({"khata": "nist"}, 404)

javab = pasokh({"method": "GET", "path": "/api/saat", "matn": "", "sorok": {}})
chap_khat javab["vaziat"]
chap_khat javab["matn"]
chap_khat pasokh({"method": "GET", "path": "/x", "matn": "", "sorok": {}})["vaziat"]
```

```output
200
{"saat":"12:00","ok":true}
404
```

<div dir="rtl">

**از برنامه‌ی خودتان.** `darkhast url` از یک سرورِ دیگر می‌پرسد و یک dict با `vaziat`، `matn` و
`sarlawah` برمی‌گرداند؛ ورودیِ دوم می‌تواند `method`، `matn` و `sarlawah` را تعیین کند.

</div>

```mihan norun
javab = darkhast("https://api.example.com/kala", {"method": "POST", "matn": json_benevis({"nam": "medad"}), "sarlawah": {"Content-Type": "application/json"}})
agar javab["vaziat"] == 200
    kala = json_bekhan(javab["matn"])
```

<div dir="rtl">

## ۱۱.۸ اجرا و اشتراکِ سایت

| دستور | چه می‌کند |
|---|---|
| `mihan run site.mhn` | سرور را اجرا می‌کند؛ نشانی‌ای را که چاپ می‌کند باز کنید |
| `mihan run --open site.mhn` | اجرا می‌کند و در مرورگرِ پیش‌فرض باز می‌کند |
| `mihan serve site.mhn` | همان `run --open` |
| `mihan build site.mhn --open-browser` | یک `.exe` که با دوبار کلیک سایت را راه می‌اندازد و مرورگر را باز می‌کند |
| `mihan build site.mhn --window` | یک `.exe` که سایت را در پنجره‌ی خودش نشان می‌دهد (درس ۱۲) |

درونِ برنامه، `serva_khod()` نشانیِ سایت است و `serva_khorshod()` درگاهِ آن، تا حتی وقتی درگاه
برایتان انتخاب شده، نشانیِ درست را چاپ یا باز کنید.

</div>


<div dir="rtl">

## ۱۱.۹ از کدام راه استفاده کنم؟

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
