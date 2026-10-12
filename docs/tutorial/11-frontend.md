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

## 11.3 Which way should I use?

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
