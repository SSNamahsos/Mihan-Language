<div dir="rtl">

# درس ۱۸ — مرجع و ابزارهای بیشتر (marja)

> **خلاصه:** همه‌ی کلمه‌های داخلی میهن در یک جا، با چند ابزار ریاضی و بررسی نوع که در درس‌های قبل نیامدند.

## ۱۸.۱ عددها و بررسی‌ها

`jazr` ریشه‌ی دوم است، `tavan` توان، `takfir` فاصله از صفر، `goosh` گرد کردن، `gooshbad` گرد کردن به پایین،
`sher` گرد کردن به بالا، `goroosh` بریدنِ اعشار، `baron` تقسیمِ صحیح و `dope` کوچک‌ترین. `pishfarz x, y`
اگر `x` برابر `hich` باشد `y` را می‌دهد. `adad_hast`، `matn_hast`، `list_hast`، `dict_hast` و `hast` نوعِ یک مقدار را
بررسی می‌کنند، `hast_rand` یک عددِ تصادفی میان ۰ و ۱ است و `mohit` یک متغیرِ محیطی را می‌خواند.

</div>

```mihan
chap_khat $"jazr(16) = {jazr(16)}"
chap_khat $"tavan(2, 10) = {tavan(2, 10)}"
chap_khat $"takfir(-3) = {takfir(-3)}"
chap_khat $"goosh(2.6) = {goosh(2.6)}, gooshbad(2.6) = {gooshbad(2.6)}, sher(2.1) = {sher(2.1)}, goroosh(-2.5) = {goroosh(-2.5)}"
chap_khat $"baron(7, 2) = {baron(7, 2)}, dope(4, 9) = {dope(4, 9)}"
chap_khat $"dereche(radians(180)) = {dereche(radians(180))}, cos(0) = {cos(0)}"
chap_khat pishfarz(hich, "pishfarz")
chap_khat adad_hast(3), " ", matn_hast("a"), " ", list_hast([1]), " ", dict_hast({}), " ", hast(hich)
chap_khat bast_hosh(4)
chap_khat azadi("  salam  ")
r = hast_rand()
chap_khat r >= 0 va r < 1
chap_khat mohit("MIHAN_NIST")
```

```output
jazr(16) = 4.0
tavan(2, 10) = 1024.0
takfir(-3) = 3
goosh(2.6) = 3, gooshbad(2.6) = 2, sher(2.1) = 3, goroosh(-2.5) = -2
baron(7, 2) = 3, dope(4, 9) = 4
dereche(radians(180)) = 180.0, cos(0) = 1.0
pishfarz
dorost dorost dorost dorost ghalat
[0, 1, 2, 3]
salam
dorost
hich
```

<div dir="rtl">

## ۱۸.۲ همه‌ی کلمه‌های داخلی

`mihan doc` همین مرجع را در ترمینال چاپ می‌کند و VS Code هم وقتی موس را روی یک کلمه نگه دارید همین
متن را نشان می‌دهد.

</div>

| میهن | انگلیسی | |
|---|---|---|
| `chap` | `print` | chap x — write a value out |
| `chap_khat` | `println` | chap_khat x — write a value out, then start a new line |
| `adad_karda` | `atoi` | adad_karda x — the number a text says; a mistake (E4125) when it says none |
| `adad_ya_hich` | `to_number` | adad_ya_hich x — the number a text says, or hich when it says none (never a mistake) |
| `zaman` | `now` | zaman() — seconds since 1970 as a float; subtract two to time something |
| `tarikh` | `date` | tarikh(olgoo?, zaman?) — the date as text, e.g. tarikh("YYYY-MM-DD HH:mm:ss") |
| `sabr` | `sleep` | sabr s — wait s seconds |
| `arghoman` | `args` | arghoman() / arghoman(i) - the words given after the program's name |
| `neshan` | `repr` | neshan x — the readable form of a value, as matn |
| `tule` | `len` | tule x — how many items are in x |
| `rang` | `range` | rang(az, ta), rang(az, ta, ghadam) — a run of numbers from az up to just before ta |
| `doreh` | `enumerate` | doreh rang(1, 4) — walk numbers alongside their positions |
| `poshe_barname` | `script_dir` | poshe_barname() — the folder this program's file is in, so data can be kept beside it |
| `pishfarz` | `coalesce` | pishfarz x, y  /  x ?? y — x, or y when x is hich |
| `zip` |  | zip a, b — pair up the items of two lists, as long as the shorter one |
| `takfir` | `abs` | takfir x — how far x is from zero |
| `goosh` | `round` | goosh x — the nearest whole number |
| `dope` | `min` | dope a, b — the smaller of two values |
| `balia` | `max` | balia a, b — the larger of two values |
| `shomar` | `count` | shomar xs — how many items hold, or how many are true |
| `hast` | `exists` | hast x — is there a value here at all |
| `tabdil` | `cast` | tabdil x, arz — change a value's type |
| `banafsh` | `hash` | banafsh x — a number derived from x |
| `bast` | `push` | bast xs, x — add one item to the end of a list |
| `bardavar` | `pop` | bardavar xs — take the last item off a list and give it back |
| `biya` | `insert` | biya xs, 0, x — put an item at a chosen place |
| `hazar` | `index` | hazar xs, x — where an item sits, or -1 when it is not there |
| `barmadar` | `contains` | barmadar xs, x — is this item in there |
| `taghib` | `sort` | taghib xs — a new list in order |
| `forogh` | `reverse` | forogh xs — a new list back to front |
| `por` | `join` | por xs, "-" — glue items together into text |
| `tab` | `split` | tab matn, "," — cut text into a list |
| `setare` | `setof` | setare xs — keep only the items that differ |
| `filter` |  | filter xs, funct(x) -> x > 2 — keep the items that pass a test |
| `tabdil_kol` | `map` | tabdil_kol xs, funct(x) -> x * 2 — rebuild every item |
| `kamil` | `reduce` | kamil rang(1, 5), funct(i) -> i * i — fold a list into one value |
| `kamej` | `any` | kamej xs, test — is any item true |
| `hamej` | `all` | hamej xs, test — is every item true |
| `dast` | `keys` | dast d — the keys of a dict, as a list |
| `meghdar` | `values` | meghdar d — the values of a dict, as a list |
| `khata` | `error` | khata("peyham") — build a failure to hand to giroft |
| `khanak` | `input` | khanak "soru: " — read one line of typed input |
| `serva_bgo` | `serve` | serva_bgo [darport] tabe [va] — answer web requests, then run one funct |
| `serva_khorshod` | `server_port` | serva_khorshod — the port the server is listening on |
| `serva_khod` | `server_url` | serva_khod — the address the server can be reached at |
| `serva_band` | `server_stop` | serva_band — stop the server and let the program finish |
| `baza_bgo` | `open_window` | baza_bgo address — open a window of your own program |
| `bekhan` | `readfile` | bekhan "notes.txt" — read a file as text |
| `benevis` | `writefile` | benevis "notes.txt", matn — write text into a file |
| `fehrest` | `listdir` | fehrest "." — the names in a folder |
| `hazf` | `deletefile` | hazf "notes.txt" — remove a file, or an empty folder |
| `pooshe` | `mkdir` | pooshe "notes" — make a folder, and the folders above it |
| `masir` | `joinpath` | masir "a", "b" — join parts of a path |
| `masir_hast` | `pathexists` | masir_hast "a" — is there a file or a folder here |
| `mohit` | `env` | mohit "TEMP" — an environment variable, or hich when it was never set |
| `json_bekhan` | `fromjson` | json_bekhan matn — a value read from JSON text |
| `json_benevis` | `tojson` | json_benevis x — JSON text for a value |
| `parvande_pasokh` | `sendfile` | parvande_pasokh "safhe", path — the answer that sends one file from a folder |
| `safhe_kon` | `page` | safhe_kon "onvan", jabe, sabk? — a page from Mihan words such as rang and gooshe_gerd |
| `form_bekhan` | `readform` | form_bekhan matn — the fields a form sent, as a dict |
| `jabe` | `box` | jabe [pieces], style — a box on the page |
| `band` | `paragraph` | band matn, style — a paragraph |
| `onvan` | `h1` | onvan matn, style — a heading |
| `onvan2` | `h2` | onvan2 matn, style — a smaller heading |
| `dokme` | `button` | dokme matn, kar, style — a button; kar is what the program is told |
| `peyvand` | `link` | peyvand matn, be, style — a link to another page |
| `vorudi` | `field` | vorudi nam, style — a field the person fills in |
| `form` | `forme` | form be, pieces, style — a form that sends those fields |
| `karte` | `card` | karte pieces, style — a rounded card |
| `navar` | `navbar` | navar links, style — the bar of links at the top of a site |
| `ofoghi` | `row` | ofoghi pieces, style — pieces side by side |
| `amudi` | `column` | amudi pieces, style — pieces one under another |
| `liste` | `ul` | liste items, style — a list of pieces |
| `jadval` | `table` | jadval headings, rows, style — a table |
| `rah` | `route` | rah darakht, rahha — the page for the path that was asked |
| `jalase` | `session` | jalase darakht — this visitor's own dict, kept between requests (a cookie) |
| `json_pasokh` | `json_answer` | json_pasokh x, vaziat — an answer whose body is x as JSON |
| `sabk` | `css` | sabk {"onvan": {"rang": "ghermez"}} — CSS text from Finglish style words, with @harekat and @kochak_tar (lesson 11.5) |
| `sabk_khat` | `inline_style` | sabk_khat {"rang": "abi"} — one style="..." text from Finglish style words |
| `qaleb` | `template` | qaleb matn, dict — {{nam}} filled in (made safe), {{{nam}}} as it is |
| `html_amn` | `html_escape` | html_amn x — x as text that is safe to put on a page |
| `rooydad_ferest` | `send_event` | rooydad_ferest kanal, x — send x to every page listening on /_mihan/rooydad/kanal |
| `android_hast` | `is_android` | android_hast() — is this running in the Android app (or its preview) |
| `darkhast` | `request` | darkhast url — ask another server and take back what it said |
| `salam` | `hello` | salam — a friendly word from the REPL |
| `khoroj` | `exit` | khoroj 0 — stop the program with a status code |
| `koshish` |  | koshish — run a block that may fail |
| `sin` |  | sin x — sine, in radians |
| `cos` |  | cos x — cosine |
| `tan` |  | tan x — tangent |
| `radians` |  | radians d — degrees as radians |
| `dereche` |  | dereche r — radians as degrees |
| `jazr` |  | jazr x — square root |
| `tavan` |  | tavan x, y — x to the power y |
| `gooshbad` |  | gooshbad x — the whole part, rounded down |
| `sher` |  | sher x — the whole part, rounded away from zero |
| `goroosh` |  | goroosh x — round to the nearest whole number |
| `hast_rand` |  | hast_rand — a number between 0 and 1 |
| `bast_hosh` |  | bast_hosh n — n whole numbers starting at 0 |
| `katoos` |  | katoos x, "%0.2f" — turn a value into text in a set shape |
| `adad_hast` |  | adad_hast x — is this value a number |
| `matn_hast` |  | matn_hast x — is this value text |
| `list_hast` |  | list_hast x — is this value a list |
| `dict_hast` |  | dict_hast x — is this value a dict |
| `darghan_hast` |  | darghan_hast x — is this value an object |
| `gozaresh` |  | gozaresh x — did this class promise the given methods |
| `rang_ya_khata` |  | rang_ya_khata(a, b, gam?) — the numbers a, a+1, ... up to b as a list |
| `ghesmat` | `floordiv` | ghesmat a, b  /  floordiv a, b — divide, dropping the remainder |
| `baron` | `intdiv` | baron a, b — a divided by b, cut off |
| `gooshband` | `upper` | gooshband matn — the same letters, bigger |
| `kiawtar` | `lower` | kiawtar matn — the same letters, smaller |
| `jabeja` | `trim` | jabeja matn — the same text without space at either end |
| `azadi` | `strip` | azadi matn — the same text without space at either end |
| `barmigefth` | `startsWith` | barmigefth matn, "a" — does it begin this way |
| `bapakhsh` | `endsWith` | bapakhsh matn, "a" — does it finish this way |
| `shamil` | `includes` | shamil matn, "a" — does it contain this |
| `jigarre` | `replace` | jigarre matn, "a", "b" — text with parts swapped out |
