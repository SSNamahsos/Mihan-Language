package eval

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"mihan/internal/ast"
	rt "mihan/internal/runtime"
)

type builtinDef struct {
	Name string

	Alias string

	Doc string

	MinArgs, MaxArgs int

	Fn func(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error)

	Result string
}

var builtins = []builtinDef{
	{Name: "chap", Alias: "print", MinArgs: 0, MaxArgs: -1, Result: "batri",
		Doc: "chap x — write a value out", Fn: fnChap},
	{Name: "chap_khat", Alias: "println", MinArgs: 0, MaxArgs: -1, Result: "batri",
		Doc: "chap_khat x — write a value out, then start a new line", Fn: fnChapLine},
	{Name: "adad_karda", Alias: "atoi", MinArgs: 1, MaxArgs: 1, Result: "adad",
		Doc: "adad_karda x — the number a text says; a mistake (E4125) when it says none",
		Fn:  fnNumber},
	{Name: "adad_ya_hich", Alias: "to_number", MinArgs: 1, MaxArgs: 1, Result: "adad",
		Doc: "adad_ya_hich x — the number a text says, or hich when it says none (never a mistake)",
		Fn:  fnNumberOrNil},
	{Name: "zaman", Alias: "now", MinArgs: 0, MaxArgs: 0, Result: "ashari",
		Doc: "zaman() — seconds since 1970 as a float; subtract two to time something",
		Fn:  fnTime},
	{Name: "tarikh", Alias: "date", MinArgs: 0, MaxArgs: 2, Result: "matn",
		Doc: "tarikh(olgoo?, zaman?) — the date as text, e.g. tarikh(\"YYYY-MM-DD HH:mm:ss\")",
		Fn:  fnDate},
	{Name: "sabr", Alias: "sleep", MinArgs: 1, MaxArgs: 1, Result: "hich",
		Doc: "sabr s — wait s seconds",
		Fn:  fnSleep},

	{Name: "arghoman", Alias: "args", MinArgs: 0, MaxArgs: 1, Result: "list",
		Doc: "arghoman() / arghoman(i) - the words given after the program's name",
		Fn:  fnArgs}, {Name: "neshan", Alias: "repr", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "neshan x — the readable form of a value, as matn", Fn: fnRepr},
	{Name: "tule", Alias: "len", MinArgs: 1, MaxArgs: 1, Result: "arz",
		Doc: "tule x — how many items are in x", Fn: fnLen},
	{Name: "rang", Alias: "range", MinArgs: 1, MaxArgs: 3, Result: "rang",
		Doc: "rang(az, ta), rang(az, ta, ghadam) — a run of numbers from az up to just before ta", Fn: fnRange},
	{Name: "doreh", Alias: "enumerate", MinArgs: 1, MaxArgs: 2, Result: "list",
		Doc: "doreh rang(1, 4) — walk numbers alongside their positions", Fn: fnEnumerate},
	{Name: "poshe_barname", Alias: "script_dir", MinArgs: 0, MaxArgs: 0, Result: "matn",
		Doc: "poshe_barname() — the folder this program's file is in, so data can be kept beside it", Fn: fnScriptDir},
	{Name: "pishfarz", Alias: "coalesce", MinArgs: 2, MaxArgs: 2, Result: "any",
		Doc: "pishfarz x, y  /  x ?? y — x, or y when x is hich", Fn: fnCoalesce},
	{Name: "zip", MinArgs: 2, MaxArgs: 2, Result: "list",
		Doc: "zip a, b — pair up the items of two lists, as long as the shorter one", Fn: fnZip},
	{Name: "takfir", Alias: "abs", MinArgs: 1, MaxArgs: 1, Result: "kasr",
		Doc: "takfir x — how far x is from zero", Fn: fnAbs},
	{Name: "goosh", Alias: "round", MinArgs: 1, MaxArgs: 1, Result: "arz",
		Doc: "goosh x — the nearest whole number", Fn: fnRound},
	{Name: "dope", Alias: "min", MinArgs: 1, MaxArgs: -1, Result: "any",
		Doc: "dope a, b — the smaller of two values", Fn: fnMin},
	{Name: "balia", Alias: "max", MinArgs: 1, MaxArgs: -1, Result: "any",
		Doc: "balia a, b — the larger of two values", Fn: fnMax},
	{Name: "shomar", Alias: "count", MinArgs: 1, MaxArgs: 1, Result: "arz",
		Doc: "shomar xs — how many items hold, or how many are true", Fn: fnCount},
	{Name: "hast", Alias: "exists", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "hast x — is there a value here at all", Fn: fnExists},
	{Name: "tabdil", Alias: "cast", MinArgs: 2, MaxArgs: 2, Result: "any",
		Doc: "tabdil x, arz — change a value's type", Fn: fnCast},
	{Name: "banafsh", Alias: "hash", MinArgs: 1, MaxArgs: 1, Result: "arz",
		Doc: "banafsh x — a number derived from x", Fn: fnHash},
	{Name: "bast", Alias: "push", MinArgs: 2, MaxArgs: 2, Result: "list",
		Doc: "bast xs, x — add one item to the end of a list", Fn: fnPush},
	{Name: "bardavar", Alias: "pop", MinArgs: 1, MaxArgs: 1, Result: "any",
		Doc: "bardavar xs — take the last item off a list and give it back", Fn: fnPop},
	{Name: "biya", Alias: "insert", MinArgs: 3, MaxArgs: 3, Result: "list",
		Doc: "biya xs, 0, x — put an item at a chosen place", Fn: fnInsert},
	{Name: "hazar", Alias: "index", MinArgs: 2, MaxArgs: 2, Result: "arz",
		Doc: "hazar xs, x — where an item sits, or -1 when it is not there", Fn: fnIndexOf},
	{Name: "barmadar", Alias: "contains", MinArgs: 2, MaxArgs: 2, Result: "bool",
		Doc: "barmadar xs, x — is this item in there", Fn: fnContains},
	{Name: "taghib", Alias: "sort", MinArgs: 1, MaxArgs: 1, Result: "list",
		Doc: "taghib xs — a new list in order", Fn: fnSort},
	{Name: "forogh", Alias: "reverse", MinArgs: 1, MaxArgs: 1, Result: "list",
		Doc: "forogh xs — a new list back to front", Fn: fnReverse},
	{Name: "por", Alias: "join", MinArgs: 2, MaxArgs: 2, Result: "matn",
		Doc: "por xs, \"-\" — glue items together into text", Fn: fnJoin},
	{Name: "tab", Alias: "split", MinArgs: 2, MaxArgs: 2, Result: "list",
		Doc: "tab matn, \",\" — cut text into a list", Fn: fnSplit},
	{Name: "setare", Alias: "setof", MinArgs: 1, MaxArgs: 1, Result: "set",
		Doc: "setare xs — keep only the items that differ", Fn: fnSetOf},
	{Name: "filter", MinArgs: 2, MaxArgs: 2, Result: "list",
		Doc: "filter xs, funct(x) -> x > 2 — keep the items that pass a test", Fn: fnFilter},
	{Name: "tabdil_kol", Alias: "map", MinArgs: 2, MaxArgs: 2, Result: "list",
		Doc: "tabdil_kol xs, funct(x) -> x * 2 — rebuild every item", Fn: fnMap},
	{Name: "kamil", Alias: "reduce", MinArgs: 2, MaxArgs: 3, Result: "any",
		Doc: "kamil rang(1, 5), funct(i) -> i * i — fold a list into one value", Fn: fnReduce},
	{Name: "kamej", Alias: "any", MinArgs: 2, MaxArgs: 2, Result: "bool",
		Doc: "kamej xs, test — is any item true", Fn: fnAny},
	{Name: "hamej", Alias: "all", MinArgs: 2, MaxArgs: 2, Result: "bool",
		Doc: "hamej xs, test — is every item true", Fn: fnAll},
	{Name: "dast", Alias: "keys", MinArgs: 1, MaxArgs: 1, Result: "list",
		Doc: "dast d — the keys of a dict, as a list", Fn: fnKeys},
	{Name: "meghdar", Alias: "values", MinArgs: 1, MaxArgs: 1, Result: "list",
		Doc: "meghdar d — the values of a dict, as a list", Fn: fnValues},
	{Name: "khata", Alias: "error", MinArgs: 1, MaxArgs: 2, Result: "khata",
		Doc: "khata(\"peyham\") — build a failure to hand to giroft", Fn: fnError},
	{Name: "khanak", Alias: "input", MinArgs: 0, MaxArgs: 1, Result: "matn",
		Doc: "khanak \"soru: \" — read one line of typed input", Fn: fnInput},
	{Name: "serva_bgo", Alias: "serve", MinArgs: 1, MaxArgs: 3, Result: "batri",
		Doc: "serva_bgo [darport] tabe [va] — answer web requests, then run one funct", Fn: fnServeStart},
	{Name: "serva_khorshod", Alias: "server_port", MinArgs: 0, MaxArgs: 0, Result: "adad",
		Doc: "serva_khorshod — the port the server is listening on", Fn: fnServePort},
	{Name: "serva_khod", Alias: "server_url", MinArgs: 0, MaxArgs: 0, Result: "matn",
		Doc: "serva_khod — the address the server can be reached at", Fn: fnServeURL},
	{Name: "serva_band", Alias: "server_stop", MinArgs: 0, MaxArgs: 0, Result: "batri",
		Doc: "serva_band — stop the server and let the program finish", Fn: fnServeStop},
	{Name: "baza_bgo", Alias: "open_window", MinArgs: 1, MaxArgs: 3, Result: "batri",
		Doc: "baza_bgo address — open a window of your own program", Fn: fnOpenWindow},

	{Name: "bekhan", Alias: "readfile", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "bekhan \"notes.txt\" — read a file as text", Fn: fnReadFile},
	{Name: "benevis", Alias: "writefile", MinArgs: 2, MaxArgs: 2, Result: "batri",
		Doc: "benevis \"notes.txt\", matn — write text into a file", Fn: fnWriteFile},
	{Name: "fehrest", Alias: "listdir", MinArgs: 1, MaxArgs: 1, Result: "list",
		Doc: "fehrest \".\" — the names in a folder", Fn: fnListDir},
	{Name: "hazf", Alias: "deletefile", MinArgs: 1, MaxArgs: 1, Result: "batri",
		Doc: "hazf \"notes.txt\" — remove a file, or an empty folder", Fn: fnRemove},
	{Name: "pooshe", Alias: "mkdir", MinArgs: 1, MaxArgs: 1, Result: "batri",
		Doc: "pooshe \"notes\" — make a folder, and the folders above it", Fn: fnMkdir},
	{Name: "masir", Alias: "joinpath", MinArgs: 1, MaxArgs: -1, Result: "matn",
		Doc: "masir \"a\", \"b\" — join parts of a path", Fn: fnJoinPath},
	{Name: "masir_hast", Alias: "pathexists", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "masir_hast \"a\" — is there a file or a folder here", Fn: fnPathExists},
	{Name: "mohit", Alias: "env", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "mohit \"TEMP\" — an environment variable, or hich when it was never set", Fn: fnEnv},
	{Name: "json_bekhan", Alias: "fromjson", MinArgs: 1, MaxArgs: 1, Result: "any",
		Doc: "json_bekhan matn — a value read from JSON text", Fn: fnJSONRead},
	{Name: "json_benevis", Alias: "tojson", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "json_benevis x — JSON text for a value", Fn: fnJSONWrite},
	{Name: "parvande_pasokh", Alias: "sendfile", MinArgs: 2, MaxArgs: 2, Result: "dict",
		Doc: "parvande_pasokh \"safhe\", path — the answer that sends one file from a folder", Fn: fnSendFile},
	{Name: "safhe_kon", Alias: "page", MinArgs: 1, MaxArgs: 2, Result: "matn",
		Doc: "safhe_kon \"onvan\", jabe — a page from Mihan words such as rang and khat_dor", Fn: fnPage},
	{Name: "form_bekhan", Alias: "readform", MinArgs: 1, MaxArgs: 1, Result: "dict",
		Doc: "form_bekhan matn — the fields a form sent, as a dict", Fn: fnFormRead},
	{Name: "jabe", Alias: "box", MinArgs: 0, MaxArgs: -1, Result: "dict",
		Doc: "jabe [pieces], style — a box on the page", Fn: fnJabe},
	{Name: "band", Alias: "paragraph", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "band matn, style — a paragraph", Fn: fnBand},
	{Name: "onvan", Alias: "h1", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "onvan matn, style — a heading", Fn: fnOnvan},
	{Name: "onvan2", Alias: "h2", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "onvan2 matn, style — a smaller heading", Fn: fnOnvan2},
	{Name: "dokme", Alias: "button", MinArgs: 1, MaxArgs: 3, Result: "dict",
		Doc: "dokme matn, kar, style — a button; kar is what the program is told", Fn: fnDokme},
	{Name: "peyvand", Alias: "link", MinArgs: 2, MaxArgs: 3, Result: "dict",
		Doc: "peyvand matn, be, style — a link to another page", Fn: fnPeyvand},
	{Name: "vorudi", Alias: "field", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "vorudi nam, style — a field the person fills in", Fn: fnVorudi},
	{Name: "form", Alias: "forme", MinArgs: 2, MaxArgs: 3, Result: "dict",
		Doc: "form be, pieces, style — a form that sends those fields", Fn: fnForm},
	{Name: "karte", Alias: "card", MinArgs: 0, MaxArgs: -1, Result: "dict",
		Doc: "karte pieces, style — a rounded card", Fn: fnKarte},
	{Name: "navar", Alias: "navbar", MinArgs: 0, MaxArgs: -1, Result: "dict",
		Doc: "navar links, style — the bar of links at the top of a site", Fn: fnNavar},
	{Name: "ofoghi", Alias: "row", MinArgs: 0, MaxArgs: -1, Result: "dict",
		Doc: "ofoghi pieces, style — pieces side by side", Fn: fnOfoghi},
	{Name: "amudi", Alias: "column", MinArgs: 0, MaxArgs: -1, Result: "dict",
		Doc: "amudi pieces, style — pieces one under another", Fn: fnAmudi},
	{Name: "liste", Alias: "ul", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "liste items, style — a list of pieces", Fn: fnListe},
	{Name: "jadval", Alias: "table", MinArgs: 2, MaxArgs: 3, Result: "dict",
		Doc: "jadval headings, rows, style — a table", Fn: fnJadval},
	{Name: "rah", Alias: "route", MinArgs: 2, MaxArgs: 2, Result: "dict",
		Doc: "rah darakht, rahha — the page for the path that was asked", Fn: fnRah},
	{Name: "jalase", Alias: "session", MinArgs: 1, MaxArgs: 1, Result: "dict",
		Doc: "jalase darakht — this visitor's own dict, kept between requests (a cookie)", Fn: fnSession},
	{Name: "json_pasokh", Alias: "json_answer", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "json_pasokh x, vaziat — an answer whose body is x as JSON", Fn: fnJSONAnswer},
	{Name: "qaleb", Alias: "template", MinArgs: 2, MaxArgs: 2, Result: "matn",
		Doc: "qaleb matn, dict — {{nam}} filled in (made safe), {{{nam}}} as it is", Fn: fnTemplate},
	{Name: "html_amn", Alias: "html_escape", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "html_amn x — x as text that is safe to put on a page", Fn: fnHTMLSafe},
	{Name: "rooydad_ferest", Alias: "send_event", MinArgs: 2, MaxArgs: 3, Result: "adad",
		Doc: "rooydad_ferest kanal, x — send x to every page listening on /_mihan/rooydad/kanal", Fn: fnSendEvent},
	{Name: "android_hast", Alias: "is_android", MinArgs: 0, MaxArgs: 0, Result: "bool",
		Doc: "android_hast() — is this running in the Android app (or its preview)", Fn: fnAndroid},
	{Name: "darkhast", Alias: "request", MinArgs: 1, MaxArgs: 2, Result: "dict",
		Doc: "darkhast url — ask another server and take back what it said", Fn: fnRequest},
	{Name: "salam", Alias: "hello", MinArgs: 0, MaxArgs: 1, Result: "batri",
		Doc: "salam — a friendly word from the REPL", Fn: fnHello},
	{Name: "khoroj", Alias: "exit", MinArgs: 0, MaxArgs: 1, Result: "batri",
		Doc: "khoroj 0 — stop the program with a status code", Fn: fnExit},
	{Name: "koshish", MinArgs: 0, MaxArgs: 0, Result: "any",
		Doc: "koshish — run a block that may fail", Fn: fnReserved},
	{Name: "sin", MinArgs: 1, MaxArgs: 1, Result: "kasr",
		Doc: "sin x — sine, in radians", Fn: fnSin},
	{Name: "cos", MinArgs: 1, MaxArgs: 1, Result: "kasr", Doc: "cos x — cosine", Fn: fnCos},
	{Name: "tan", MinArgs: 1, MaxArgs: 1, Result: "kasr", Doc: "tan x — tangent", Fn: fnTan},
	{Name: "radians", MinArgs: 1, MaxArgs: 1, Result: "kasr",
		Doc: "radians d — degrees as radians", Fn: fnRadians},
	{Name: "dereche", MinArgs: 1, MaxArgs: 1, Result: "kasr",
		Doc: "dereche r — radians as degrees", Fn: fnDegrees},
	{Name: "jazr", MinArgs: 1, MaxArgs: 1, Result: "kasr",
		Doc: "jazr x — square root", Fn: fnSqrt},
	{Name: "tavan", MinArgs: 2, MaxArgs: 2, Result: "kasr",
		Doc: "tavan x, y — x to the power y", Fn: fnPow},
	{Name: "b logaritmi", MinArgs: 1, MaxArgs: 1, Result: "kasr",
		Doc: "logarithm x — natural logarithm", Fn: fnLog},
	{Name: "gooshbad", MinArgs: 1, MaxArgs: 1, Result: "int",
		Doc: "gooshbad x — the whole part, rounded down", Fn: fnFloor},
	{Name: "sher", MinArgs: 1, MaxArgs: 1, Result: "int",
		Doc: "sher x — the whole part, rounded away from zero", Fn: fnCeil},
	{Name: "goroosh", MinArgs: 1, MaxArgs: 1, Result: "int",
		Doc: "goroosh x — round to the nearest whole number", Fn: fnTrunc},
	{Name: "hast_rand", MinArgs: 0, MaxArgs: 1, Result: "kasr",
		Doc: "hast_rand — a number between 0 and 1", Fn: fnRandom},
	{Name: "bast_hosh", MinArgs: 1, MaxArgs: 1, Result: "list",
		Doc: "bast_hosh n — n whole numbers starting at 0", Fn: fnSeq},
	{Name: "katoos", MinArgs: 2, MaxArgs: 2, Result: "matn",
		Doc: "katoos x, \"%0.2f\" — turn a value into text in a set shape", Fn: fnFormat},
	{Name: "adad_hast", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "adad_hast x — is this value a number", Fn: fnIsNumber},
	{Name: "matn_hast", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "matn_hast x — is this value text", Fn: fnIsString},
	{Name: "list_hast", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "list_hast x — is this value a list", Fn: fnIsList},
	{Name: "dict_hast", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "dict_hast x — is this value a dict", Fn: fnIsDict},
	{Name: "darghan_hast", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "darghan_hast x — is this value an object", Fn: fnIsObject},
	{Name: "gozaresh", MinArgs: 1, MaxArgs: 1, Result: "bool",
		Doc: "gozaresh x — did this class promise the given methods", Fn: fnSatisfies},
	{Name: "rang_ya_khata", MinArgs: 0, MaxArgs: -1, Result: "list",
		Doc: "rang_ya_khata(a, b) — the numbers a, a+1, ... as a list", Fn: fnSeq},

	{Name: "ghesmat", Alias: "floordiv", MinArgs: 2, MaxArgs: 2, Result: "any",
		Doc: "ghesmat a, b  /  floordiv a, b — divide, dropping the remainder", Fn: fnFloorDiv},
	{Name: "baron", Alias: "intdiv", MinArgs: 2, MaxArgs: 2, Result: "arz",
		Doc: "baron a, b — a divided by b, cut off", Fn: fnFloorDiv},

	{Name: "gooshband", Alias: "upper", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "gooshband matn — the same letters, bigger", Fn: fnStringMethod("gooshband")},
	{Name: "kiawtar", Alias: "lower", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "kiawtar matn — the same letters, smaller", Fn: fnStringMethod("kiawtar")},
	{Name: "jabeja", Alias: "trim", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "jabeja matn — the same text without space at either end", Fn: fnStringMethod("jabeja")},
	{Name: "azadi", Alias: "strip", MinArgs: 1, MaxArgs: 1, Result: "matn",
		Doc: "azadi matn — the same text without space at either end", Fn: fnStringMethod("azadi")},
	{Name: "barmigefth", Alias: "startsWith", MinArgs: 2, MaxArgs: 2, Result: "bool",
		Doc: "barmigefth matn, \"a\" — does it begin this way", Fn: fnStringMethod("barmigefth")},
	{Name: "bapakhsh", Alias: "endsWith", MinArgs: 2, MaxArgs: 2, Result: "bool",
		Doc: "bapakhsh matn, \"a\" — does it finish this way", Fn: fnStringMethod("bapakhsh")},
	{Name: "shamil", Alias: "includes", MinArgs: 2, MaxArgs: 2, Result: "bool",
		Doc: "shamil matn, \"a\" — does it contain this", Fn: fnStringMethod("shamil")},
	{Name: "jigarre", Alias: "replace", MinArgs: 3, MaxArgs: 3, Result: "matn",
		Doc: "jigarre matn, \"a\", \"b\" — text with parts swapped out", Fn: fnStringMethod("jigarre")},
}

func fnStringMethod(name string) func(*Interp, []rt.Value, ast.Node) (rt.Value, error) {
	return func(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
		s, ok := args[0].(rt.Str)
		if !ok {
			return nil, in.faultAt(nil, "E4250",
				fmt.Sprintf("%s فقط روی matn کار می‌کند", name),
				"%s works on matn only, not on %s", name, rt.TypeName(args[0]))
		}
		return in.strMethod(s, name, args[1:])
	}
}

var stringMethods = map[string]bool{
	"gooshband": true, "upper": true, "kiawtar": true, "lower": true,
	"jabeja": true, "trim": true, "azadi": true, "strip": true,
	"barmigefth": true, "startsWith": true, "bapakhsh": true, "endsWith": true,
	"shamil": true, "includes": true, "jigarre": true, "replace": true,
	"tab": true, "split": true, "tule": true, "len": true,
	"por": true, "join": true, "banafsh": true, "hash": true,
	"adad_hast": true, "shamil_harf": true,
}

var listMethods = map[string]bool{
	"bast": true, "push": true, "bardavar": true, "pop": true,
	"biya": true, "insert": true, "hazar": true, "index": true,
	"barmadar": true, "contains": true, "taghib": true, "sort": true,
	"forogh": true, "reverse": true, "por": true, "join": true,
	"filter": true, "tabdil_kol": true, "map": true, "kamil": true, "reduce": true,
	"kamej": true, "any": true, "hamej": true, "all": true,
	"tule": true, "len": true, "shomar": true, "count": true,
	"setare": true, "setof": true, "taghib_kol": true,
}

var dictMethods = map[string]bool{
	"dast": true, "keys": true, "meghdar": true, "values": true,
	"tule": true, "len": true, "barmadar": true, "contains": true,
	"hazar": true, "index": true, "bast": true, "put": true,
	"bardavar": true, "pop": true, "has_key": true,
	"begir": true, "get": true, "hazf": true, "delete": true,
}

func (in *Interp) nativeMethod(kind, name string) bool {
	switch kind {
	case "matn":
		return stringMethods[name]
	case "list":
		return listMethods[name]
	case "dict":
		return dictMethods[name]
	}
	return false
}

func installBuiltins(in *Interp) {
	in.builtins = make(map[string]*rt.Function, len(builtins)*2)
	for i := range builtins {
		d := &builtins[i]
		fn := &rt.Function{
			Name:        d.Name,
			BuiltinName: d.Name,
			Doc:         d.Doc,
			Result:      d.Result,
			Native: func(host any, args []rt.Value, node ast.Node) (rt.Value, error) {
				ip := host.(*Interp)
				if err := ip.checkArity(node, d); err != nil {
					return nil, err
				}
				return d.Fn(ip, args, node)
			},
		}
		in.builtins[d.Name] = fn
		if d.Alias != "" {

			alias := *fn
			alias.Name = d.Alias
			alias.BuiltinName = d.Alias
			in.builtins[d.Alias] = &alias
		}
	}
}

func (in *Interp) BuiltinNames() []string {
	out := make([]string, 0, len(in.builtins))
	for n := range in.builtins {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

type BuiltinInfo struct {
	Name string

	Alias string

	Doc string

	Result string

	MinArgs, MaxArgs int
}

func Builtins() []BuiltinInfo {
	out := make([]BuiltinInfo, 0, len(builtins))
	for _, d := range builtins {
		out = append(out, BuiltinInfo{
			Name:    d.Name,
			Alias:   d.Alias,
			Doc:     d.Doc,
			Result:  d.Result,
			MinArgs: d.MinArgs,
			MaxArgs: d.MaxArgs,
		})
	}
	return out
}

func (in *Interp) BuiltinDocs() []string {
	out := make([]string, 0, len(builtins))
	for _, d := range builtins {
		out = append(out, d.Doc)
	}
	sort.Strings(out)
	return out
}

func (in *Interp) checkArity(node ast.Node, d *builtinDef) error {
	n := -1
	switch a := node.(type) {
	case *ast.Call:
		n = len(a.Args)
	case *ast.ExprStmt:
		if c, ok := a.X.(*ast.Call); ok {
			n = len(c.Args)
		}
	}
	if n < 0 {
		return nil
	}
	if n < d.MinArgs {
		return in.faultAt(node, "E4200",
			fmt.Sprintf("«%s» دست‌کم %d چیز می‌خواهد ولی %d دادی", d.Name, d.MinArgs, n),
			"%s needs at least %d values and you gave %d", d.Name, d.MinArgs, n)
	}
	if d.MaxArgs >= 0 && n > d.MaxArgs {
		return in.faultAt(node, "E4201",
			fmt.Sprintf("«%s» بیش از %d چیز نمی‌گیرد ولی %d دادی", d.Name, d.MaxArgs, n),
			"%s takes at most %d values and you gave %d", d.Name, d.MaxArgs, n)
	}
	return nil
}

func fnChap(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	for _, a := range args {
		if _, err := fmt.Fprint(in.Out, rt.Show(a)); err != nil {
			return rt.Nil, err
		}
	}
	return rt.Nil, nil
}

func fnChapLine(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	for _, a := range args {
		if _, err := fmt.Fprint(in.Out, rt.Show(a)); err != nil {
			return rt.Nil, err
		}
	}
	_, err := fmt.Fprintln(in.Out)
	return rt.Nil, err
}

func fnRepr(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	return rt.Str(rt.Repr(args[0])), nil
}

func fnHello(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	who := "donya"
	if len(args) > 0 {
		who = rt.Show(args[0])
	}
	_, err := fmt.Fprintf(in.Out, "salam %s! khosh amadid.\n", who)
	return rt.Nil, err
}

func fnInput(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	if len(args) > 0 {
		if _, err := fmt.Fprint(in.Out, rt.Show(args[0])); err != nil {
			return rt.Nil, err
		}
	}
	line, err := readLine(in)
	if err != nil {
		return rt.Str(""), nil
	}
	return rt.Str(line), nil
}

func fnExit(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	code := 0
	if len(args) > 0 {
		if i, ok := toInt(args[0]); ok {
			code = int(i)
		}
	}
	return rt.Nil, &exitRequest{code: code, out: in}
}

type exitRequest struct {
	code int
	out  *Interp
}

func (e *exitRequest) Error() string { return "khoroj " + strconv.Itoa(e.code) }

func ExitCode(err error) (int, bool) {
	if e, ok := err.(*exitRequest); ok {
		return e.code, true
	}
	return 0, false
}

func fnReserved(in *Interp, _ []rt.Value, node ast.Node) (rt.Value, error) {
	return nil, in.faultAt(node, "E4202",
		"این یک ساختار است، نه تابع", "this word starts a structure, it is not a funct you can call")
}

func num1(in *Interp, args []rt.Value, node ast.Node, what string) (float64, error) {
	f, ok := toFloat(args[0])
	if !ok {
		return 0, in.faultAt(node, "E4210",
			fmt.Sprintf("%s فقط با عدد کار می‌کند", what),
			"%s only works on numbers, not %s", what, rt.TypeName(args[0]))
	}
	return f, nil
}

func fnAbs(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	if i, ok := args[0].(rt.Int); ok {
		if i < 0 {
			return -i, nil
		}
		return i, nil
	}
	f, err := num1(in, args, node, "takfir")
	if err != nil {
		return nil, err
	}
	return rt.Float(math.Abs(f)), nil
}

func fnRound(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "goosh")
	if err != nil {
		return nil, err
	}
	return rt.Int(int64(math.Round(f))), nil
}

func fnFloor(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "gooshbad")
	if err != nil {
		return nil, err
	}
	return rt.Int(int64(math.Floor(f))), nil
}

func fnCeil(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "sher")
	if err != nil {
		return nil, err
	}
	return rt.Int(int64(math.Ceil(f))), nil
}

func fnTrunc(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "goroosh")
	if err != nil {
		return nil, err
	}
	return rt.Int(int64(math.Trunc(f))), nil
}

func fnSqrt(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "jazr")
	if err != nil {
		return nil, err
	}
	if f < 0 {
		return nil, in.faultAt(node, "E4211",
			"ریشه دومِ عدد منفی جوابِ واقعی ندارد", "the square root of a negative number is not a real number")
	}
	return rt.Float(math.Sqrt(f)), nil
}

func fnPow(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	a, err := num1(in, args[:1], node, "tavan")
	if err != nil {
		return nil, err
	}
	b, err := num1(in, args[1:], node, "tavan")
	if err != nil {
		return nil, err
	}
	return rt.Float(math.Pow(a, b)), nil
}

func fnLog(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "logarithm")
	if err != nil {
		return nil, err
	}
	if f <= 0 {
		return nil, in.faultAt(node, "E4212",
			"لگاریتمِ صفر یا منفی جوابی ندارد", "the logarithm of zero or a negative number has no answer")
	}
	return rt.Float(math.Log(f)), nil
}

func fnSin(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "sin")
	if err != nil {
		return nil, err
	}
	return rt.Float(math.Sin(f)), nil
}

func fnCos(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "cos")
	if err != nil {
		return nil, err
	}
	return rt.Float(math.Cos(f)), nil
}

func fnTan(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "tan")
	if err != nil {
		return nil, err
	}
	return rt.Float(math.Tan(f)), nil
}

func fnRadians(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "radians")
	if err != nil {
		return nil, err
	}
	return rt.Float(f * math.Pi / 180), nil
}

func fnDegrees(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	f, err := num1(in, args, node, "dereche")
	if err != nil {
		return nil, err
	}
	return rt.Float(f * 180 / math.Pi), nil
}

func fnRandom(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	if len(args) == 0 {
		return rt.Float(rand.Float64()), nil
	}
	f, ok := toFloat(args[0])
	if !ok {
		return rt.Float(rand.Float64()), nil
	}
	return rt.Float(rand.Float64() * f), nil
}

func fnMin(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return in.extreme(node, args, -1)
}

func fnMax(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return in.extreme(node, args, 1)
}

func fnFloorDiv(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return in.floorDiv(node, args[0], args[1])
}

func (in *Interp) extreme(node ast.Node, args []rt.Value, want int) (rt.Value, error) {
	best := args[0]
	for _, a := range args[1:] {
		c, ok := rt.Compare(a, best)
		if !ok {
			continue
		}
		if c == want {
			best = a
		}
	}
	_ = node
	return best, nil
}

func fnCount(in *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	switch t := args[0].(type) {
	case *rt.List:
		return rt.Int(len(t.Items)), nil
	case *rt.Set:
		return rt.Int(t.Len()), nil
	case *rt.Dict:
		return rt.Int(t.Len()), nil
	case rt.Str:
		return rt.Int(rt.RuneCount(string(t))), nil
	}
	return rt.Int(0), nil
}

func fnExists(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	if args[0] == nil {
		return rt.Bool(false), nil
	}
	if _, ok := args[0].(rt.Hich); ok {
		return rt.Bool(false), nil
	}
	return rt.Bool(true), nil
}

func fnHash(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	if _, ok := rt.HashKey(args[0]); !ok {
		return nil, in.faultAt(node, "E4213",
			fmt.Sprintf("از %s نمی‌شود banafsh گرفت", rt.TypeName(args[0])),
			"a %s cannot be hashed; only numbers, matn, bool and hich can", rt.TypeName(args[0]))
	}
	h, _ := rt.HashKey(args[0])

	var sum int64 = 1469598103934665603
	for i := 0; i < len(h); i++ {
		sum ^= int64(h[i])
		sum *= 1099511628211
	}
	return rt.Int(sum), nil
}

func fnFormat(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	shape, ok := args[1].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4214",
			"قالبِ katoos باید matn باشد", "the shape given to katoos has to be matn")
	}
	if i, ok := args[0].(rt.Int); ok {
		return rt.Str(fmt.Sprintf(string(shape), int64(i))), nil
	}
	f, err := num1(in, args[:1], node, "katoos")
	if err != nil {
		return nil, err
	}
	return rt.Str(fmt.Sprintf(string(shape), f)), nil
}

func fnLen(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	switch t := args[0].(type) {
	case rt.Str:
		return rt.Int(rt.RuneCount(string(t))), nil
	case *rt.List:
		return rt.Int(len(t.Items)), nil
	case *rt.Set:
		return rt.Int(t.Len()), nil
	case *rt.Dict:
		return rt.Int(t.Len()), nil
	case *rt.Range:
		return rt.Int(len(rangeItems(*t))), nil
	}
	return nil, in.faultAt(node, "E4220",
		fmt.Sprintf("tule را نمی‌شود از %s گرفت", rt.TypeName(args[0])),
		"you cannot ask the length of a %s", rt.TypeName(args[0]))
}

func fnRange(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	nums := make([]rt.Int, len(args))
	for i, a := range args {
		v, ok := toInt(a)
		if !ok {
			return nil, in.faultAt(node, "E4221",
				"سرِ rang باید عددِ کامل باشد", "every end of a rang must be a whole number")
		}
		nums[i] = v
	}
	switch len(nums) {
	case 1:
		return &rt.Range{Lo: 0, Hi: nums[0]}, nil
	case 2:
		return &rt.Range{Lo: nums[0], Hi: nums[1]}, nil
	case 3:

		step := nums[2]
		if step == 0 {
			return nil, in.faultAt(node, "E4222",
				"گامِ صفر جلو نمی‌رود", "a step of zero never moves, so the rang would never end")
		}
		return &rt.StepRange{Lo: nums[0], Hi: nums[1], Step: step}, nil
	}
	return &rt.Range{}, nil
}

func fnSeq(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	if len(args) == 0 {
		return rt.NewList(), nil
	}
	n, ok := toInt(args[0])
	if !ok || n < 0 {
		return rt.NewList(), nil
	}
	out := make([]rt.Value, 0, n)
	for i := rt.Int(0); i < n; i++ {
		out = append(out, i)
	}
	return rt.NewList(out...), nil
}

func fnEnumerate(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	out := make([]rt.Value, 0, len(items))
	for i, v := range items {
		out = append(out, rt.NewList(rt.Int(i), v))
	}
	return rt.NewList(out...), nil
}

func fnScriptDir(in *Interp, _ []rt.Value, _ ast.Node) (rt.Value, error) {
	if in.File == "" {
		wd, _ := os.Getwd()
		return rt.Str(wd), nil
	}
	full, err := filepath.Abs(in.File)
	if err != nil {
		return rt.Str(filepath.Dir(in.File)), nil
	}
	return rt.Str(filepath.Dir(full)), nil
}

func fnCoalesce(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	if args[0] == nil || args[0] == rt.Nil {
		return args[1], nil
	}
	return args[0], nil
}

func fnZip(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	a, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	b, err := in.itemsOf(node, args[1])
	if err != nil {
		return nil, err
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]rt.Value, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, rt.NewList(a[i], b[i]))
	}
	return rt.NewList(out...), nil
}

func fnPush(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	l, ok := args[0].(*rt.List)
	if !ok {
		return nil, in.faultAt(node, "E4223",
			fmt.Sprintf("bast فقط روی list کار می‌کند، نه %s", rt.TypeName(args[0])),
			"bast works on a list, not on a %s", rt.TypeName(args[0]))
	}
	l.Items = append(l.Items, args[1])
	return l, nil
}

func fnPop(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	l, ok := args[0].(*rt.List)
	if !ok {
		return nil, in.faultAt(node, "E4224",
			fmt.Sprintf("bardavar فقط روی list کار می‌کند، نه %s", rt.TypeName(args[0])),
			"bardavar works on a list, not on a %s", rt.TypeName(args[0]))
	}
	if len(l.Items) == 0 {
		return rt.Nil, in.faultAt(node, "E4225",
			"این list خالی است", "this list is empty, so there is nothing to take off it")
	}
	last := l.Items[len(l.Items)-1]
	l.Items = l.Items[:len(l.Items)-1]
	return last, nil
}

func fnInsert(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	l, ok := args[0].(*rt.List)
	if !ok {
		return nil, in.faultAt(node, "E4226",
			"biya فقط روی list کار می‌کند", "biya works on a list only")
	}
	at, ok := toInt(args[1])
	if !ok {
		return nil, in.faultAt(node, "E4227",
			"جای biya باید عدد باشد", "the position for biya must be a whole number")
	}
	idx := int(at)
	if idx < 0 {
		idx += len(l.Items)
	}
	if idx < 0 {
		idx = 0
	}
	if idx > len(l.Items) {
		idx = len(l.Items)
	}
	l.Items = append(l.Items, rt.Nil)
	copy(l.Items[idx+1:], l.Items[idx:])
	l.Items[idx] = args[2]
	return l, nil
}

func fnIndexOf(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	for i, v := range items {
		if rt.Equal(v, args[1]) {
			return rt.Int(i), nil
		}
	}
	return rt.Int(-1), nil
}

func fnContains(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {

	if s, ok := args[0].(rt.Str); ok {
		sub, ok := args[1].(rt.Str)
		if !ok {
			return nil, in.faultAt(node, "E4241",
				"در یک متن فقط متن را می‌شود جست",
				"barmadar on a text looks for a text, not %s", rt.TypeName(args[1]))
		}
		return rt.Bool(strings.Contains(string(s), string(sub))), nil
	}
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	for _, v := range items {
		if rt.Equal(v, args[1]) {
			return rt.Bool(true), nil
		}
	}
	return rt.Bool(false), nil
}

func fnSort(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	out := make([]rt.Value, len(items))
	copy(out, items)
	rt.SortValues(out)
	return rt.NewList(out...), nil
}

func fnReverse(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	out := make([]rt.Value, len(items))
	for i, v := range items {
		out[len(items)-1-i] = v
	}
	return rt.NewList(out...), nil
}

func fnJoin(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	sep := ""
	if s, ok := args[1].(rt.Str); ok {
		sep = string(s)
	}
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = rt.Show(v)
	}
	return rt.Str(strings.Join(parts, sep)), nil
}

func fnSplit(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	s, ok := args[0].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4228",
			"tab فقط روی matn کار می‌کند", "tab works on matn only")
	}
	sep, ok := args[1].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4229",
			"جداکننده باید matn باشد", "the separator must be matn")
	}
	parts := strings.Split(string(s), string(sep))
	out := make([]rt.Value, len(parts))
	for i, p := range parts {
		out[i] = rt.Str(p)
	}
	return rt.NewList(out...), nil
}

func fnSetOf(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	s := rt.NewSet()
	for _, v := range items {
		if _, ok := rt.HashKey(v); ok {
			s.Add(v)
		}
	}
	return s, nil
}

func fnKeys(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	d, ok := args[0].(*rt.Dict)
	if !ok {
		return nil, in.faultAt(node, "E4230",
			"dast فقط روی dict کار می‌کند", "dast works on a dict only")
	}
	out := make([]rt.Value, 0, d.Len())
	out = append(out, d.Keys()...)
	return rt.NewList(out...), nil
}

func fnValues(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	d, ok := args[0].(*rt.Dict)
	if !ok {
		return nil, in.faultAt(node, "E4231",
			"meghdar فقط روی dict کار می‌کند", "meghdar works on a dict only")
	}
	out := make([]rt.Value, 0, d.Len())
	for _, k := range d.Keys() {
		v, _ := d.Get(k)
		out = append(out, v)
	}
	return rt.NewList(out...), nil
}

func fnFilter(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	var out []rt.Value
	for _, v := range items {
		keep, err := in.callWith(node, args[1], []rt.Value{v})
		if err != nil {
			return nil, err
		}
		if rt.Truthy(keep) {
			out = append(out, v)
		}
	}
	if out == nil {
		out = []rt.Value{}
	}
	return rt.NewList(out...), nil
}

func fnMap(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	out := make([]rt.Value, 0, len(items))
	for _, v := range items {
		r, err := in.callWith(node, args[1], []rt.Value{v})
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return rt.NewList(out...), nil
}

func fnReduce(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		if len(args) > 2 {
			return args[2], nil
		}
		return nil, in.faultAt(node, "E4232",
			"kamil روی list خالی چیزی برای جمع کردن ندارد",
			"kamil has nothing to fold because the list is empty; give it a starting value")
	}
	acc := items[0]
	for _, v := range items[1:] {
		var err error
		acc, err = in.callWith(node, args[1], []rt.Value{acc, v})
		if err != nil {
			return nil, err
		}
	}
	return acc, nil
}

func fnAny(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return anyAll(in, node, args, true)
}

func fnAll(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	return anyAll(in, node, args, false)
}

func anyAll(in *Interp, node ast.Node, args []rt.Value, want bool) (rt.Value, error) {
	items, err := in.itemsOf(node, args[0])
	if err != nil {
		return nil, err
	}
	for _, v := range items {
		r, err := in.callWith(node, args[1], []rt.Value{v})
		if err != nil {
			return nil, err
		}
		if rt.Truthy(r) == want {
			return rt.Bool(want), nil
		}
	}
	return rt.Bool(!want), nil
}

func fnCast(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	typeName, ok := args[1].(rt.Str)
	if !ok {
		return nil, in.faultAt(node, "E4233",
			"نوعِ tabdil باید matn باشد", "the type given to tabdil must be matn")
	}
	return in.convertTo(node, args[0], string(typeName))
}

func fnError(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	e := &rt.ErrorValue{Message: rt.Str("khata"), Data: rt.Nil, Name: "khata"}
	if len(args) > 0 {
		e.Message = rt.Str(rt.Show(args[0]))
	}
	if len(args) > 1 {
		e.Data = args[1]
	}
	return e, nil
}

func fnSatisfies(in *Interp, args []rt.Value, node ast.Node) (rt.Value, error) {
	obj, ok := args[0].(*rt.Object)
	if !ok {
		return rt.Bool(false), nil
	}
	cv, ok := args[1].(*rt.ClassValue)
	if !ok {
		return rt.Bool(false), nil
	}
	for name := range cv.Class.Methods {
		if _, found := obj.Class.FindMethod(name); found == nil {
			return rt.Bool(false), nil
		}
	}
	return rt.Bool(obj.Class.IsA(cv.Class)), nil
}

func fnIsNumber(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	return rt.Bool(rt.IsNumber(args[0])), nil
}
func fnIsString(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	_, ok := args[0].(rt.Str)
	return rt.Bool(ok), nil
}
func fnIsList(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	_, ok := args[0].(*rt.List)
	return rt.Bool(ok), nil
}
func fnIsDict(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	_, ok := args[0].(*rt.Dict)
	return rt.Bool(ok), nil
}
func fnIsObject(_ *Interp, args []rt.Value, _ ast.Node) (rt.Value, error) {
	_, ok := args[0].(*rt.Object)
	return rt.Bool(ok), nil
}

func (in *Interp) itemsOf(node ast.Node, v rt.Value) ([]rt.Value, error) {
	switch t := v.(type) {
	case *rt.List:
		out := make([]rt.Value, len(t.Items))
		copy(out, t.Items)
		return out, nil
	case *rt.Set:
		return t.Items(), nil
	case rt.Str:
		s := string(t)
		n := rt.RuneCount(s)
		out := make([]rt.Value, n)
		for i := 0; i < n; i++ {
			r, _ := rt.RuneAt(s, i)
			out[i] = rt.Str(string(r))
		}
		return out, nil
	case *rt.Dict:
		return t.Keys(), nil
	case *rt.Range:
		return rangeItems(*t), nil
	case *rt.StepRange:
		return stepRangeItems(*t), nil
	}
	return nil, in.faultAt(node, "E4240",
		fmt.Sprintf("از %s نمی‌شود فهرست ساخت", rt.TypeName(v)),
		"you cannot walk a %s item by item", rt.TypeName(v))
}

func (in *Interp) ReadLine() (string, error) { return readLine(in) }

func (in *Interp) RunBuiltin(name string, args []rt.Value, node ast.Node) (rt.Value, bool, error) {
	fn, ok := in.builtins[name]
	if !ok {
		return nil, false, nil
	}
	v, err := fn.Native(in, args, node)
	return v, true, err
}

func (in *Interp) MethodNames(kind string) map[string]bool {
	out := map[string]bool{}
	var table map[string]bool
	switch kind {
	case "matn":
		table = stringMethods
	case "list":
		table = listMethods
	case "dict":
		table = dictMethods
	}
	for k := range table {
		out[k] = true
	}

	out["tule"] = true
	out["len"] = true
	return out
}

func (in *Interp) RunMethod(recv rt.Value, name string, args []rt.Value, node ast.Node) (rt.Value, bool, error) {
	switch t := recv.(type) {
	case rt.Str:
		if in.nativeMethod("matn", name) {
			v, err := in.strMethod(t, name, args)
			return v, true, err
		}
	case *rt.List:
		if in.nativeMethod("list", name) {
			v, err := in.listMethod(t, name, args)
			return v, true, err
		}
	case *rt.Dict:
		if in.nativeMethod("dict", name) {
			v, err := in.dictMethod(t, name, args)
			return v, true, err
		}
	}
	return nil, false, nil
}

func readLine(in *Interp) (string, error) {
	if in.reader == nil {
		in.reader = bufio.NewReader(in.In)
	}
	line, err := in.reader.ReadString('\n')
	if err != nil && line == "" {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
