# Lesson 10 — A website backend (serva_bgo)

> **خلاصه:** یک سایت در میهن فقط یک تابع است: درخواست (`darakht`) را می‌گیرد و پاسخ برمی‌گرداند.
> `serva_bgo(port, tabe)` سرور را روشن می‌کند.

## 10.1 The smallest server

A request comes in as a dict called `darakht` with these keys:

| key | meaning |
|---|---|
| `"method"` | `"GET"`, `"POST"`, … |
| `"path"` | the path, e.g. `"/api/kala"` |
| `"sorok"` | the `?a=1&b=2` query, as a dict |
| `"matn"` | the request body, as text |
| `"sarlawah"` | the request headers, as a dict |

Your function returns a dict with `"vaziat"` (status code), `"matn"` (the body)
and optionally `"sarlawah"` (response headers).

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

Run it with `mihan run server.mhn` and open `http://127.0.0.1:8701/salam?naam=Sara`
in a browser. Press `Ctrl+C` to stop it. (The lines starting with `GET` above are
what the tutorial's checker received when it visited those addresses.)

### Open it in the browser by itself

```
mihan run --open server.mhn      # or:  mihan serve server.mhn
```

When the server is listening, Mihan prints (in a terminal) `Mihan site is ready: http://...`
and, with `--open` (or `mihan serve`), opens that address in your default
browser: on Windows through the system's URL handler, on Linux with
`xdg-open`, on macOS with `open`. `--open` may also come after the file name.

## 10.2 Test the logic without a browser

Because a site is just a function, you can call it directly — the fastest way to
check your backend:

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

## 10.3 A JSON API with data

This is a real REST-style API for a to-do list. It answers:

* `GET /api/kar` → the whole list as JSON
* `POST /api/kar` with `{"onvan": "..."}` → adds an item
* `GET /api/kar/2` → one item, or `404`

and saves everything to `kar.json` so nothing is lost when the server restarts.

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

You can try it from a second terminal with `curl`:

```
curl http://127.0.0.1:8702/api/kar
curl -X POST -d "{\"onvan\":\"kharid\"}" http://127.0.0.1:8702/api/kar
```

## 10.4 Port 0 and the address

`serva_bgo(0, pasokh, bargoz)` lets the computer choose a free port. The third
function runs once the server is listening, and `serva_khod()` is its address:

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

## Exercises

1. Add a path `/api/saat` to the to-do server that answers `{"vaziat": "ok"}`.
2. Write a server where `/jam?a=3&b=4` answers `7`.

<details><summary>Solutions</summary>

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

**Next:** [Lesson 11 — A website frontend](11-frontend.md)
