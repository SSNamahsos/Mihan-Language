# Lesson 11 — A website frontend

> **خلاصه:** دو راه برای ساختن صفحه هست: (۱) فایل‌های معمولی HTML/CSS/JS کنار برنامه که میهن آن‌ها را می‌فرستد،
> (۲) ساختن صفحه با کلمه‌های خودِ میهن مثل `onvan`، `karte`، `dokme` بدون نوشتن HTML.

## 11.1 Way 1 — real HTML, CSS and JavaScript files

Put the page in a folder called `safhe` next to your program. `parvande_pasokh`
sends a file from that folder (and `/` means `index.html`). The page calls your
API with `fetch`, exactly as it would call any backend.

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

And the program — the backend — is short:

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

Open `http://127.0.0.1:8711/` and you have a working app: HTML page, CSS style,
JavaScript in the browser, and Mihan behind it.

## 11.2 Way 2 — pages made of Mihan pieces

For quick tools you don't have to write HTML at all. Mihan has *pieces*:

| piece | what it is |
|---|---|
| `onvan "x"`, `onvan2 "x"` | big and smaller heading |
| `band "x"` | a paragraph |
| `karte [...]` | a rounded card |
| `ofoghi [...]`, `amudi [...]` | pieces side by side / one under another |
| `navar [...]` | the bar of links at the top |
| `peyvand "matn", "/masir"` | a link |
| `dokme "matn", "kar"` | a button; pressing it sends `kar` to the program |
| `jadval sar, satr_ha` | a table |
| `liste [...]` | a bulleted list |

Every piece takes an optional style dict as its last argument, with Mihan words:
`"rang"` (text colour), `"pas_zamine"` (background), `"khat_dor"` (corner
radius), `"fasele"` (padding).

`rah(darakht, routes)` chooses the page for the path that was asked, and builds
the whole HTML page for you.

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

To serve it for real, add one line at the end: `serva_bgo(8080, pasokh)`.

## 11.3 Every piece, and boxes you build yourself

Each piece function (`onvan`, `band`, `karte` …) only makes a **dict**. You can
write that dict yourself: `"gone"` says what kind of piece it is, `"matn"` is its
text and `"darun"` holds the pieces inside it. This is how you get pieces that
have no function of their own, such as a picture or a smaller heading.

| `"gone"` | becomes | extra keys |
|---|---|---|
| `jabe` | a box (`div`) | `darun` |
| `band`, `onvan`, `onvan2`, `onvan3` | paragraph, headings | `matn` |
| `peyvand` | a link | `be` = the address |
| `tasvir` | a picture | `be` = the picture's address, `matn` = its description |
| `khat` | a dividing line | |
| `fehrest`, `radif` | a list and one item of it | |
| `jadval`, `satr`, `sarnevesht`, `khane_jadval` | a table, a row, a heading cell, a cell | |
| `form`, `vorudi`, `dokme` | a form, a field, a button | see 11.4 |
| `navar` | the bar of links | |

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

### Style words

The last argument of a piece (or any key in a box dict) can be a style word
such as `rang`, `zamine`, `gooshe_gerd` or `fasele_dakheli`. Numbers for sizes
are pixels. Section 11.5 lists every word and every value.

## 11.4 Events: buttons and forms

Pages made of pieces talk back to your program. A `dokme` with a `kar` sends a
POST with `kar=…` to the same address when it is pressed, and the page is
replaced by your new answer: that is the counter in 11.2. A `form` sends every
`vorudi` inside it; `form_bekhan` turns what arrived into a dict.

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

`jalase(darakht)` gives each visitor a dict of their own, kept between requests
with a cookie, so two people do not share one counter (lesson 16).

## 11.5 Styling in Finglish: style words, `sabk_khat` and `sabk`

Every place that takes a style in Mihan takes the same Finglish words: the
style dict at the end of a piece, the keys of a box dict, `sabk_khat` (one
`style="..."` text) and `sabk` (a whole style sheet). Write `rang` instead of
`color`, `gooshe_gerd` instead of `border-radius`, `siah` instead of `black`.
A plain number for a size means pixels; any other unit (`%`, `rem`, `em`,
`vh`, `s`, `deg`) you write in the text exactly as in CSS.

```mihan
chap_khat sabk_khat({"rang": "siah", "gooshe_gerd": 30, "hashie": "1px saf khakestari"})
chap_khat sabk_khat({"zamine": "sormei", "fasele_dakheli": "8px 16px", "andaze_matn": 18, "vazn_matn": "porrang"})
```

```output
color:black;border-radius:30px;border:1px solid #6b7280;
background:#1e3a8a;padding:8px 16px;font-size:18px;font-weight:bold;
```

The English CSS names keep working, with `-` or `_`, so you can copy a style
from anywhere and mix the two:

```mihan
chap_khat sabk_khat({"color": "black", "border-radius": 30, "rang_zamine": "zard"})
```

```output
color:black;border-radius:30px;background-color:#eab308;
```

### Layout and position

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

### Effects and motion

Inside a value, Finglish function names work too: `charkhesh` (rotate),
`bozorgi` (scale), `jabejayi_y` (translateY), `tari` (blur), `shib`
(linear-gradient), `aks` (url) and the rest of the table below. In `enteghal`
you can name the style word that should move smoothly.

```mihan
chap_khat sabk_khat({"tabdil": "charkhesh(10deg) bozorgi(1.1)", "enteghal": "tabdil 0.3s narm", "safi": "tari(4px)", "shafafiat": 0.9, "nashangar": "dast"})
```

```output
transform:rotate(10deg) scale(1.1);transition:transform 0.3s ease;filter:blur(4px);opacity:0.9;cursor:pointer;
```

### A whole style sheet with `sabk`

`sabk` turns a dict of selectors into CSS text. Piece names work as selectors
(`onvan` is `h1`, `dokme` is `button`, `jabe` is `div`, `badane` is `body`),
and so do Finglish states: `:rooye` (hover), `:fokus`, `:faal` (active),
`:aval`, `:akhar`, `::ghabl`, `::baad`. Special keys:

- `"@harekat nam"`: an animation, with frames `az` (from), `ta` (to) or `"50%"`.
- `"@kochak_tar 600"` / `"@bozorg_tar 900"`: only on screens narrower / wider than that.
- `"@kam_harekat"`: for visitors who asked for less motion. `"@tarik"` / `"@roshan"`: dark or light system theme.

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

Give `safhe_kon` the sheet as its third argument, and give a piece a class
with `kelas` (or an id with `shenase`) so the sheet can find it:

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

A word Mihan does not know stops the program and says which word it thinks
you meant:

```mihan fails
chap_khat sabk_khat({"gooshe_gard": 12})
```

```output
khata: E4350
  in kalame-ye sabk ro nemishnasam: gooshe_gard; shayad manzooret 'gooshe_gerd' bood?
   1 | chap_khat sabk_khat({"gooshe_gard": 12})
                 ^
```

### The full list

**Colour and background**

| Mihan | Older name | CSS |
|---|---|---|
| `rang` |  | `color` |
| `zamine` | `pas_zamine` | `background` |
| `rang_zamine` |  | `background-color` |
| `tasvir_zamine` |  | `background-image` |
| `andaze_zamine` |  | `background-size` |
| `jaye_zamine` |  | `background-position` |

**Border and corners**

| Mihan | Older name | CSS |
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

**Spacing**

| Mihan | Older name | CSS |
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

**Size**

| Mihan | Older name | CSS |
|---|---|---|
| `pahna` |  | `width` |
| `ertefa` | `bala` | `height` |
| `kamtarin_pahna` |  | `min-width` |
| `bishtarin_pahna` |  | `max-width` |
| `kamtarin_ertefa` |  | `min-height` |
| `bishtarin_ertefa` |  | `max-height` |
| `andaze_jabe` |  | `box-sizing` |

**Text and font**

| Mihan | Older name | CSS |
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

**Layout: flex and grid**

| Mihan | Older name | CSS |
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

**Position and layers**

| Mihan | Older name | CSS |
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

**Effects and motion**

| Mihan | Older name | CSS |
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

**Colours**

| Mihan | CSS | Mihan | CSS | Mihan | CSS |
|---|---|---|---|---|---|
| `abi` | `#2563eb` | `abi_aseman` | `#38bdf8` | `banafsh` | `#7c3aed` |
| `banafshe` | `#7c3aed` | `firoozei` | `#06b6d4` | `ghahvei` | `#92400e` |
| `ghermez` | `#dc2626` | `ghermez_tire` | `#7f1d1d` | `hamrang` | `currentColor` |
| `kerem` | `#fef3c7` | `khakestari` | `#6b7280` | `meshki` | `black` |
| `narenji` | `#f97316` | `noghrei` | `#c0c0c0` | `sabz` | `#16a34a` |
| `sabz_tire` | `#14532d` | `sefid` | `white` | `shafaf` | `transparent` |
| `siah` | `black` | `soorati` | `#ec4899` | `sormei` | `#1e3a8a` |
| `talayi` | `#d4af37` | `zard` | `#eab308` | `zereshki` | `#9f1239` |

**Value words**

| Mihan | CSS | Mihan | CSS | Mihan | CSS |
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

**Functions inside values**

| Mihan | CSS | Mihan | CSS | Mihan | CSS |
|---|---|---|---|---|---|
| `aks()` | `url()` | `ashba()` | `saturate()` | `bishtarin()` | `max()` |
| `bozorgi()` | `scale()` | `charkhesh()` | `rotate()` | `jabejayi()` | `translate()` |
| `jabejayi_x()` | `translateX()` | `jabejayi_y()` | `translateY()` | `kaji()` | `skew()` |
| `kamtarin()` | `min()` | `khakestari()` | `grayscale()` | `mahdood()` | `clamp()` |
| `mohasebe()` | `calc()` | `motaghayer()` | `var()` | `roshanai()` | `brightness()` |
| `saye()` | `drop-shadow()` | `shib()` | `linear-gradient()` | `shib_dayere()` | `radial-gradient()` |
| `tari()` | `blur()` | `tekrar()` | `repeat()` |  |  |

## 11.6 Your own design and animations with `qaleb`

Pieces cover quick pages. For a page that moves, write the HTML and CSS as a
template and let Mihan fill it in: `qaleb` puts each `{{nam}}` from a dict into
the text, made safe, and `{{{nam}}}` exactly as it is. CSS animations, hover
effects and transitions all go in the template's `<style>`.

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

`html_amn` makes text safe to put inside HTML, so a visitor cannot inject a tag.

## 11.7 Calling an API

**From the page.** Answer JSON with `json_pasokh`, and the page's JavaScript
calls it with `fetch` (as in 11.1).

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

**From your program.** `darkhast url` asks another server and gives back a dict
with `vaziat`, `matn` and `sarlawah`; a second argument can set `method`,
`matn` and `sarlawah`.

```mihan norun
javab = darkhast("https://api.example.com/kala", {"method": "POST", "matn": json_benevis({"nam": "medad"}), "sarlawah": {"Content-Type": "application/json"}})
agar javab["vaziat"] == 200
    kala = json_bekhan(javab["matn"])
```

## 11.8 Running and sharing your site

| command | what it does |
|---|---|
| `mihan run site.mhn` | runs the server; open the address it prints |
| `mihan run --open site.mhn` | runs it and opens it in your default browser |
| `mihan serve site.mhn` | the same as `run --open` |
| `mihan build site.mhn --open-browser` | one `.exe` that starts the site and opens the browser when double-clicked |
| `mihan build site.mhn --window` | one `.exe` that shows the site in its own window (lesson 12) |

Inside the program, `serva_khod()` is the site's address and `serva_khorshod()`
its port, so you can print or open the right address even when the port was
chosen for you.

## 11.9 Which way should I use?

* **Pieces** — admin tools, quick dashboards, school projects: no HTML needed.
* **HTML/CSS/JS files** — anything that must look exactly like a design, or uses
  a JavaScript library. Mihan is then your backend and file server.

The folder `examples/pro-sites/` has three complete, polished sites built the
second way.

## Exercises

1. Add a `"-1"` button to the counter that sends `kam`.
2. Add a page `/list` that shows `jadval(["naam", "sen"], [["Ali", "20"], ["Sara", "25"]])`.

<details><summary>Solutions</summary>

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

**Next:** [Lesson 12 — A Windows app](12-windows-app.md)
