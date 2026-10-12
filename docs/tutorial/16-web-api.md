# Lesson 16 — A real web API: routes, methods, cookies, uploads, live events

> **خلاصه:** این درس همهٔ چیزهایی را که یک API واقعی لازم دارد نشان می‌دهد:
> `json_pasokh` برای جوابِ JSON با کدِ وضعیت، `rah` با پارامترِ مسیر (`/kala/:id`)
> و فیلترِ `method`، `sorok` و `sarlawah`، کوکی و `jalase`، آپلودِ فایل، CORS،
> قالب با `qaleb`، و رویدادهای زنده با `rooydad_ferest`.

Lesson 10 answered requests by hand. This lesson uses the pieces that make a
bigger API short and correct. Every sample here is started as a server and
asked real questions; the answers below each sample are what it really said.

## 16.1 JSON answers with a status: `json_pasokh`

`json_pasokh(x)` is an answer whose body is `x` written as JSON, with the type
`application/json`. A second value is the status code.

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

## 16.2 Routes with parameters, and methods

`rah(darakht, [...])` picks the first route whose `path` fits. A part written
`:nam` fits any one part of the address and lands in `darakht["params"]`; a
path that ends in `*` fits everything under it. A route may say which
`method` it answers; asked with another method, it answers 405 for you.

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

A full REST API is the same idea with one funct per address and a look at
`darakht["method"]` (GET, POST, PUT, DELETE). Dicts have `hazf(kelid)` for
the DELETE side:

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

`darakht["sorok"]` is the query (`?q=med`), and `darakht["sarlawah"]` is the
request's headers. An answer can carry its own headers under `"sarlawah"`.

## 16.3 CORS

A page from another address may only read your answers if you say so. Put
`"cors": dorost` in an answer to allow every origin, or an origin's address
to allow just that one (with cookies). Answer the browser's `OPTIONS`
question the same way:

```mihan norun
funct pasokh(darakht)
    agar darakht["method"] == "OPTIONS"
        bargard {"vaziat": 204, "matn": "", "cors": "https://example.com"}
    bargard {"vaziat": 200, "matn": "{}", "cors": "https://example.com"}

serva_bgo(8080, pasokh)
```

## 16.4 Cookies and sessions

`darakht["cookie"]` holds the cookies the browser sent. To set one, an answer
gives `"cookie": {"nam": "meghdar"}`, or a dict with `meghdar`, `umr`
(seconds), `masir` and `httponly`; `hich` removes it.

`jalase(darakht)` is a session: a dict that belongs to this one visitor and is
kept between requests. Mihan gives the visitor a `mihan_jalase` cookie the
first time, and nothing else is needed:

```mihan norun
funct pasokh(darakht)
    j = jalase(darakht)
    j["bar"] = j.begir("bar", 0) + 1
    bargard {"matn": $"you have been here {j["bar"]} times", "cookie": {"akharin": darakht["path"]}}

serva_bgo(8080, pasokh)
```

## 16.5 Forms and file uploads

A form sent as `application/x-www-form-urlencoded` or `multipart/form-data`
lands in `darakht["form"]`. Files sent with it are in `darakht["parvande"]`,
each a dict with `nam` (the form field), `esm` (the file's name), `noe` (its
type), `andaze` (its size) and `matn` (its content):

```mihan norun
funct pasokh(darakht)
    agar darakht["method"] == "POST"
        har f dar darakht["parvande"]
            benevis_file("upload-" + f["esm"], f["matn"])
        bargard json_pasokh({"name": darakht["form"].begir("nam", ""), "files": tule(darakht["parvande"])}, 201)
    bargard "<form method=post enctype=multipart/form-data><input name=nam><input type=file name=f><button>send</button></form>"

serva_bgo(8080, pasokh)
```

## 16.6 Static files: types and caching

`parvande_pasokh("safhe", darakht["path"])` sends a file with the right type
(`text/css`, `text/javascript`, `image/png`, …) and an `ETag`. A browser that
already has the file asks with `If-None-Match` and gets `304 Not Modified` with
no body, so pages with many files load fast the second time and never show an
old file after you change it.

## 16.7 Templates: `qaleb`

`qaleb(matn, dict)` fills `{{nam}}` in a text. What it puts in is made safe for
HTML (`<` becomes `&lt;`), so a visitor's words cannot become a script;
`{{{nam}}}` puts it in as it is, and `{{a.b}}` reaches inside a dict.
`html_amn(x)` makes one text safe by itself.

```mihan
safhe = "<h1>{{onvan}}</h1><p>{{karbar.nam}}</p>{{{tah}}}"
chap_khat(qaleb(safhe, {"onvan": "<salam>", "karbar": {"nam": "Sara"}, "tah": "<hr>"}))
chap_khat(html_amn("a < b & c"))
```

```output
<h1>&lt;salam&gt;</h1><p>Sara</p><hr>
a &lt; b &amp; c
```

## 16.8 Live events: `rooydad_ferest`

A page can listen for news from the program without asking again and again:

```html
<script>
  new EventSource("/_mihan/rooydad/khabar").onmessage = (m) => console.log(JSON.parse(m.data));
</script>
```

and the program sends to everyone listening on that channel with
`rooydad_ferest("khabar", x)`, which gives back how many were listening.
These are server-sent events: they go one way, from the program to the
page, which is what news, chat messages and progress need. The page sends with
`fetch`, as before.

```mihan norun
funct pasokh(darakht)
    agar darakht["method"] == "POST"
        n = rooydad_ferest("khabar", {"matn": darakht["matn"]})
        bargard json_pasokh({"shenavande": n})
    bargard parvande_pasokh("safhe", darakht["path"])

serva_bgo(8080, pasokh)
```

`examples/site/namayesh/` puts all of this together on one animated page.
