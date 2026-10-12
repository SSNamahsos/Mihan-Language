package token

import "sort"

type Group string

func (g Group) String() string { return string(g) }

const (
	GroupControl     Group = "control"
	GroupDeclaration Group = "declaration"
	GroupOperator    Group = "operator"
	GroupType        Group = "type"
	GroupModifier    Group = "modifier"
	GroupBuiltin     Group = "builtin"
	GroupMisc        Group = "misc"
)

type Assoc uint8

const (
	Left Assoc = iota
	Right
)

type Keyword struct {
	Name    string
	Aliases []string
	Group   Group
	Prec    int
	Assoc   Assoc

	NoBrackets bool

	Method bool
	Doc    string

	Snippet string
}

func (k *Keyword) IsVerb() bool { return k.Group == GroupBuiltin && !k.NoBrackets }

var KeywordGroups = []Group{
	GroupControl, GroupDeclaration, GroupOperator,
	GroupType, GroupModifier, GroupBuiltin, GroupMisc,
}

var Keywords = []Keyword{

	{Name: "agar", Aliases: []string{"if"}, Group: GroupControl, Doc: "agar x > 10  /  if x > 10 — start a conditional block",
		Snippet: "agar ${1:shart}\n    ${0}"},
	{Name: "warna", Aliases: []string{"else"}, Group: GroupControl, Doc: "warna  /  else — the alternative branch; `warna agar` chains another condition",
		Snippet: "warna\n    ${0}"},
	{Name: "ta", Aliases: []string{"while"}, Group: GroupControl, Doc: "ta shart  /  while cond — repeat while a condition holds",
		Snippet: "ta ${1:shart}\n    ${0}"},
	{Name: "har", Aliases: []string{"for"}, Group: GroupControl, Doc: "har i az 0 ta 10  /  for i from 0 to 10 — counted loop",
		Snippet: "har ${1:i} az ${2:0} ta ${3:10}\n    ${0}"},
	{Name: "shek", Aliases: []string{"break"}, Group: GroupControl, Doc: "shek  /  break — leave the nearest loop"},
	{Name: "edame", Aliases: []string{"continue"}, Group: GroupControl, Doc: "edame  /  continue — skip to the next iteration"},
	{Name: "bargard", Aliases: []string{"return"}, Group: GroupControl, Doc: "bargard x  /  return x — send a value back and leave"},
	{Name: "koshish", Aliases: []string{"try"}, Group: GroupControl, Doc: "koshish  /  try — run a block that may fail",
		Snippet: "koshish\n    ${1}\ngiroft khata ${2:e}\n    ${0}"},
	{Name: "giroft", Aliases: []string{"catch"}, Group: GroupControl, Doc: "giroft khata e  /  catch error e — handle a failure"},
	{Name: "nemigir", Aliases: []string{"finally"}, Group: GroupControl, Doc: "nemigir  /  finally — always runs, even after a failure"},
	{Name: "portab", Aliases: []string{"throw", "raise"}, Group: GroupControl, Doc: "portab khata  /  raise err — stop and report a failure"},

	{Name: "funct", Aliases: []string{"func", "fn", "def", "tabe"}, Group: GroupDeclaration, Doc: "funct add(a, b)  /  fn add(a, b) — declare a function",
		Snippet: "funct ${1:name}(${2:a})\n    ${0}"},
	{Name: "darghan", Aliases: []string{"class"}, Group: GroupDeclaration, Doc: "darghan Name  /  class Name — declare a class",
		Snippet: "darghan ${1:Name}\n    ${0}"},
	{Name: "sakht", Aliases: []string{"struct"}, Group: GroupDeclaration, Doc: "sakht Point  /  struct Point — declare a fixed-shape record",
		Snippet: "sakht ${1:Name}\n    ${0}"},
	{Name: "dastoor", Aliases: []string{"interface", "trait"}, Group: GroupDeclaration, Doc: "dastoor Name  /  interface Name — declare required behaviour"},
	{Name: "shenase", Aliases: []string{"enum"}, Group: GroupDeclaration, Doc: "shenase Color  /  enum Color — declare a closed set of named values"},
	{Name: "biar", Aliases: []string{"import", "use"}, Group: GroupDeclaration, Doc: `biar "std/fmt"  /  import "std/fmt" — bring in another module`,
		Snippet: "biar \"${1:rules.mhn}\""},
	{Name: "modul", Aliases: []string{"module"}, Group: GroupDeclaration, Doc: `modul "nama/mihan-example"  /  module "name/example"`},
	{Name: "barname", Aliases: []string{"program"}, Group: GroupDeclaration, Doc: "barname MyApp  /  program MyApp — name the program this file belongs to"},

	{Name: "va", Aliases: []string{"and"}, Group: GroupOperator, Prec: 2, Doc: "va  /  and — logical and"},
	{Name: "ya", Aliases: []string{"or"}, Group: GroupOperator, Prec: 1, Doc: "ya  /  or — logical or"},
	{Name: "na", Aliases: []string{"not"}, Group: GroupOperator, Prec: 3, Doc: "na x  /  not x — logical not"},
	{Name: "dar", Aliases: []string{"in"}, Group: GroupOperator, Prec: 5, Doc: "x dar list  /  x in list — membership or comprehension filter"},
	{Name: "ghesmat", Aliases: []string{"floordiv", "idiv"}, Group: GroupOperator, Prec: 12, Doc: "a ghesmat b — divide, dropping the remainder"},
	{Name: "az", Aliases: []string{"from"}, Group: GroupOperator, Doc: "har i az 0 ta 10  /  for i from 0 to 10 — the start of a counted loop"},
	{Name: "ke", Aliases: []string{"where", "suchthat"}, Group: GroupOperator, Doc: "har x dar xs ke x > 3  /  for x in xs where x > 3 — keep only the items that pass a test"},

	{Name: "arz", Aliases: []string{"int"}, Group: GroupType, Doc: "arz — 64-bit signed integer"},
	{Name: "kasr", Aliases: []string{"float", "double"}, Group: GroupType, Doc: "kasr — double-precision number"},
	{Name: "matn", Aliases: []string{"string", "str"}, Group: GroupType, Doc: "matn — text"},
	{Name: "harf", Aliases: []string{"char", "rune"}, Group: GroupType, Doc: "harf — one character"},
	{Name: "bool", Aliases: []string{"boolean"}, Group: GroupType, Doc: "bool — dorost / ghalat"},
	{Name: "araye", Aliases: []string{"array"}, Group: GroupType, Doc: "araye — fixed-length sequence"},
	{Name: "list", Group: GroupType, Doc: "list — growable sequence"},
	{Name: "dict", Aliases: []string{"map"}, Group: GroupType, Doc: "dict — key to value table"},
	{Name: "set", Group: GroupType, Doc: "set — unordered unique values"},
	{Name: "hich", Aliases: []string{"nil", "null", "none"}, Group: GroupType, Doc: "hich — the absence of a value"},
	{Name: "batri", Aliases: []string{"void"}, Group: GroupType, Doc: "batri — no value at all"},
	{Name: "gong", Aliases: []string{"type"}, Group: GroupType, Doc: "gong — a value's type"},
	{Name: "khata", Aliases: []string{"error"}, Group: GroupType, Doc: "khata — a recoverable failure carrying a matn mesaj"},
	{Name: "amane", Aliases: []string{"Promise", "Task"}, Group: GroupType, Doc: "amane — a value that will arrive later"},

	{Name: "dorost", Aliases: []string{"true"}, Group: GroupMisc, Doc: "dorost — the true value"},
	{Name: "ghalat", Aliases: []string{"false"}, Group: GroupMisc, Doc: "ghalat — the false value"},

	{Name: "baza", Aliases: []string{"public"}, Group: GroupModifier, Doc: "baza — visible everywhere (the default)"},
	{Name: "mahram", Aliases: []string{"private"}, Group: GroupModifier, Doc: "mahram — visible only inside its own class"},
	{Name: "sabit", Aliases: []string{"const", "final"}, Group: GroupModifier, Doc: "sabit — cannot change after the first assignment"},
	{Name: "osta", Aliases: []string{"static"}, Group: GroupModifier, Doc: "osta — one shared copy for the whole program"},
	{Name: "talign", Aliases: []string{"defer"}, Group: GroupModifier, Doc: "talign f() — run f when the enclosing block exits"},
	{Name: "hamvazan", Aliases: []string{"async"}, Group: GroupModifier, Doc: "hamvazan — run concurrently"},
	{Name: "intazar", Aliases: []string{"await"}, Group: GroupModifier, Doc: "intazar — wait for a result"},
	{Name: "go", Group: GroupModifier, Doc: "go f() — start f on a new thread"},

	{Name: "chap", Aliases: []string{"print"}, Group: GroupBuiltin, Doc: `chap x  /  print x — write a value out`},
	{Name: "chap_khat", Aliases: []string{"println"}, Group: GroupBuiltin, Doc: "chap_khat x  /  println x — print, then a line break",
		Snippet: "chap_khat ${0}"},
	{Name: "khanak", Aliases: []string{"input", "read"}, Group: GroupBuiltin, Doc: `khanak "soru: "  /  input "prompt: " — read one line of typed input`},
	{Name: "serva_bgo", Aliases: []string{"serve"}, Group: GroupBuiltin, Doc: "serva_bgo [darport] tabe [va]  /  serve — answer web requests, then run one funct",
		Snippet: "funct ${1:pasokh}(darakht)\n    bargard { \"vaziat\": 200, \"matn\": \"${2:salam}\" }\n\nserva_bgo(0, ${1:pasokh})\n${0}"},
	{Name: "serva_khorshod", Aliases: []string{"server_port"}, Group: GroupBuiltin, Doc: "serva_khorshod  /  server_port — the port the server is listening on"},
	{Name: "serva_khod", Aliases: []string{"server_url"}, Group: GroupBuiltin, Doc: "serva_khod  /  server_url — the address the server can be reached at"},
	{Name: "serva_band", Aliases: []string{"server_stop"}, Group: GroupBuiltin, Doc: "serva_band  /  server_stop — stop the server and let the program finish"},
	{Name: "baza_bgo", Aliases: []string{"open_window", "window"}, Group: GroupBuiltin, Doc: "baza_bgo address  /  open_window — open a window of your own program"},
	{Name: "bekhan", Aliases: []string{"readfile"}, Group: GroupBuiltin, Doc: "bekhan \"notes.txt\"  /  readfile — read a file as text"},
	{Name: "benevis", Aliases: []string{"writefile"}, Group: GroupBuiltin, Doc: "benevis \"notes.txt\", matn  /  writefile — write text into a file"},
	{Name: "fehrest", Aliases: []string{"listdir"}, Group: GroupBuiltin, Doc: "fehrest \".\"  /  listdir — the names in a folder"},
	{Name: "hazf", Aliases: []string{"deletefile"}, Group: GroupBuiltin, Doc: "hazf \"notes.txt\"  /  deletefile — remove a file, or an empty folder"},
	{Name: "pooshe", Aliases: []string{"mkdir"}, Group: GroupBuiltin, Doc: "pooshe \"notes\"  /  mkdir — make a folder"},
	{Name: "masir", Aliases: []string{"joinpath"}, Group: GroupBuiltin, Doc: "masir \"a\", \"b\"  /  joinpath — join parts of a path"},
	{Name: "masir_hast", Aliases: []string{"pathexists"}, Group: GroupBuiltin, Doc: "masir_hast \"a\"  /  pathexists — is there a file or a folder here"},
	{Name: "sabr", Aliases: []string{"sleep"}, Group: GroupBuiltin, Doc: "sabr 0.5  /  sleep — wait this many seconds"},
	{Name: "mohit", Aliases: []string{"env"}, Group: GroupBuiltin, Doc: "mohit \"TEMP\"  /  env — an environment variable, or hich"},
	{Name: "json_bekhan", Aliases: []string{"fromjson"}, Group: GroupBuiltin, Doc: "json_bekhan matn  /  fromjson — a value read from JSON text"},
	{Name: "json_benevis", Aliases: []string{"tojson"}, Group: GroupBuiltin, Doc: "json_benevis x  /  tojson — JSON text for a value"},
	{Name: "parvande_pasokh", Aliases: []string{"sendfile"}, Group: GroupBuiltin, Doc: "parvande_pasokh \"safhe\", path  /  sendfile — the answer that sends one file"},
	{Name: "safhe_kon", Aliases: []string{"page"}, Group: GroupBuiltin,
		Doc:     "safhe_kon \"onvan\", jabe  /  page — a page from Mihan words such as rang and khat_dor",
		Snippet: "safhe_kon \"${1:onvan}\", {\n    \"gone\": \"jabe\",\n    \"rang\": \"${2:meshki}\",\n    \"khat_dor\": ${3:20},\n    \"matn\": \"${4:salam}\"\n}"},
	{Name: "form_bekhan", Aliases: []string{"readform"}, Group: GroupBuiltin, Doc: "form_bekhan matn  /  readform — the fields a form sent, as a dict"},
	{Name: "jabe", Aliases: []string{"box"}, Group: GroupBuiltin, Doc: "jabe pieces, style  /  box — a box on the page",
		Snippet: "jabe([\n    ${0}\n])"},
	{Name: "band", Aliases: []string{"paragraph"}, Group: GroupBuiltin, Doc: "band matn  /  paragraph — a paragraph"},
	{Name: "onvan", Aliases: []string{"h1"}, Group: GroupBuiltin, Doc: "onvan matn  /  h1 — a heading",
		Snippet: "onvan(\"${1:salam}\")"},
	{Name: "onvan2", Aliases: []string{"h2"}, Group: GroupBuiltin, Doc: "onvan2 matn  /  h2 — a smaller heading"},
	{Name: "dokme", Aliases: []string{"button"}, Group: GroupBuiltin, Doc: "dokme matn, kar  /  button — a button; kar is what the program is told",
		Snippet: "dokme(\"${1:beza}\", \"${2:kar}\")"},
	{Name: "peyvand", Aliases: []string{"link"}, Group: GroupBuiltin, Doc: "peyvand matn, be  /  link — a link to another page"},
	{Name: "vorudi", Aliases: []string{"field"}, Group: GroupBuiltin, Doc: "vorudi nam  /  field — a field the person fills in"},
	{Name: "form", Aliases: []string{"forme"}, Group: GroupBuiltin, Doc: "form be, pieces  /  forme — a form that sends those fields"},
	{Name: "karte", Aliases: []string{"card"}, Group: GroupBuiltin, Doc: "karte pieces  /  card — a rounded card"},
	{Name: "navar", Aliases: []string{"navbar"}, Group: GroupBuiltin, Doc: "navar links  /  navbar — the bar of links at the top of a site"},
	{Name: "ofoghi", Aliases: []string{"row"}, Group: GroupBuiltin, Doc: "ofoghi pieces  /  row — pieces side by side"},
	{Name: "amudi", Aliases: []string{"column"}, Group: GroupBuiltin, Doc: "amudi pieces  /  column — pieces one under another"},
	{Name: "liste", Aliases: []string{"ul"}, Group: GroupBuiltin, Doc: "liste items  /  ul — a list of pieces"},
	{Name: "jadval", Aliases: []string{"table"}, Group: GroupBuiltin, Doc: "jadval headings, rows  /  table — a table"},
	{Name: "rah", Aliases: []string{"route"}, Group: GroupBuiltin, Doc: "rah darakht, rahha  /  route — the page for the path that was asked",
		Snippet: "rah(darakht, [\n    {\"path\": \"/\", \"onvan\": \"${1:khane}\", \"tabe\": ${2:khane}}\n])"},
	{Name: "darkhast", Aliases: []string{"request"}, Group: GroupBuiltin, Doc: "darkhast url  /  request — ask another server and take back what it said"},
	{Name: "khoroj", Aliases: []string{"exit"}, Group: GroupBuiltin, Doc: "khoroj 0  /  exit 0 — stop the program with a status code"},
	{Name: "khod", Aliases: []string{"self", "this"}, Group: GroupBuiltin, NoBrackets: true, Doc: "khod — the current object"},
	{Name: "tule", Aliases: []string{"len"}, Group: GroupBuiltin, Doc: "tule x  /  len x — how many items are in x"},
	{Name: "rang", Aliases: []string{"range"}, Group: GroupBuiltin, Doc: "rang(a, b) — a sequence of numbers from a up to just before b"},
	{Name: "salam", Aliases: []string{"hello", "hi"}, Group: GroupBuiltin, Doc: "salam  /  hello — say hello from the REPL"},
	{Name: "neshan", Aliases: []string{"repr", "toString"}, Group: GroupBuiltin, Doc: "neshan x  /  repr x — the readable form of a value, as matn"},
	{Name: "takfir", Aliases: []string{"abs"}, Group: GroupBuiltin, Doc: "takfir x  /  abs x — how far x is from zero"},
	{"dope", []string{"min"}, GroupBuiltin, 0, Left, false, false, "dope a, b  /  min a, b — the smaller of two values", ""},
	{"balia", []string{"max"}, GroupBuiltin, 0, Left, false, false, "balia a, b  /  max a, b — the larger of two values", ""},
	{"goosh", []string{"round"}, GroupBuiltin, 0, Left, false, false, "goosh x  /  round x — the nearest whole number", ""},
	{"shomar", []string{"count"}, GroupBuiltin, 0, Left, false, false, "shomar x  /  count x — how many items there are", ""},
	{"hast", []string{"exists"}, GroupBuiltin, 0, Left, false, false, "hast x  /  exists x — is there a value here at all", ""},
	{"tabdil", []string{"cast", "convert"}, GroupBuiltin, 0, Left, false, false, "tabdil x, arz  /  cast x, int — change a value's type", ""},
	{"banafsh", []string{"hash"}, GroupBuiltin, 0, Left, false, false, "banafsh x  /  hash x — a number derived from x", ""},
	{"bast", []string{"push"}, GroupBuiltin, 0, Left, false, false, "bast xs, x  /  push xs, x — add one item to a list", ""},
	{"bardavar", []string{"pop"}, GroupBuiltin, 0, Left, false, false, "bardavar xs  /  pop xs — take the last item off a list", ""},
	{"biya", []string{"insert"}, GroupBuiltin, 0, Left, false, false, "biya xs, 0, x  /  insert xs, 0, x — put an item at a chosen place", ""},
	{"hazar", []string{"index"}, GroupBuiltin, 0, Left, false, false, "hazar xs, x  /  index xs, x — where an item sits, or -1", ""},
	{"barmadar", []string{"contains"}, GroupBuiltin, 0, Left, false, false, "barmadar xs, x  /  contains xs, x — is this item in there", ""},
	{"taghib", []string{"sort"}, GroupBuiltin, 0, Left, false, false, "taghib xs  /  sort xs — a new list in order", ""},
	{"forogh", []string{"reverse"}, GroupBuiltin, 0, Left, false, false, "forogh xs  /  reverse xs — a new list back to front", ""},
	{"por", []string{"join"}, GroupBuiltin, 0, Left, false, false, "por xs, \"-\"  /  join xs, \"-\" — glue items into text", ""},
	{"filter", nil, GroupBuiltin, 0, Left, false, false, "filter xs, funct(x) -> x > 2  — keep the items that pass a test", ""},
	{"tabdil_kol", []string{"tabdil_koli"}, GroupBuiltin, 0, Left, false, false, "tabdil_kol xs, funct(x) -> x * 2  /  map xs, fn — rebuild every item", ""},
	{"kamil", []string{"reduce"}, GroupBuiltin, 0, Left, false, false, "kamil rang(1, 5), funct(i) -> i * i  /  reduce — fold a list into one value", ""},
	{"kamej", []string{"any"}, GroupBuiltin, 0, Left, false, false, "kamej xs, test  /  any xs, test — is any item true", ""},
	{"hamej", []string{"all"}, GroupBuiltin, 0, Left, false, false, "hamej xs, test  /  all xs, test — is every item true", ""},
	{"setare", []string{"setof"}, GroupBuiltin, 0, Left, false, false, "setare xs  /  setof xs — keep only the items that differ", ""},
	{"dast", []string{"keys"}, GroupBuiltin, 0, Left, false, false, "dast d  /  keys d — the keys of a dict, as a list", ""},
	{"meghdar", []string{"values"}, GroupBuiltin, 0, Left, false, false, "meghdar d  /  values d — the values of a dict, as a list", ""},
	{"bast_hosh", nil, GroupBuiltin, 0, Left, false, false, "bast_hosh n  /  seq n — n whole numbers starting at 0", ""},
	{"doreh", []string{"enumerate"}, GroupBuiltin, 0, Left, false, false, "doreh rang(1, 4)  /  enumerate — walk items with their positions", ""},
	{"zip", nil, GroupBuiltin, 0, Left, false, false, "zip a, b  /  zip a, b — pair up items of two lists", ""},

	{Name: "gooshband", Aliases: []string{"upper"}, Group: GroupBuiltin, Method: true, Doc: `"salam".gooshband()  /  "salam".upper() — the same letters, bigger`},
	{Name: "kiawtar", Aliases: []string{"lower"}, Group: GroupBuiltin, Method: true, Doc: `"Salam".kiawtar()  /  lower — the same letters, smaller`},
	{Name: "jabeja", Aliases: []string{"trim"}, Group: GroupBuiltin, Method: true, Doc: `"  a  ".jabeja()  /  trim — the same text without space at either end`},
	{Name: "azadi", Aliases: []string{"strip"}, Group: GroupBuiltin, Method: true, Doc: `"  a  ".azadi()  /  strip — the same text without space at either end`},
	{Name: "barmigefth", Aliases: []string{"startsWith"}, Group: GroupBuiltin, Method: true, Doc: `"salam".barmigefth("sa")  /  startsWith — does it begin this way`},
	{Name: "bapakhsh", Aliases: []string{"endsWith"}, Group: GroupBuiltin, Method: true, Doc: `"salam".bapakhsh("am")  /  endsWith — does it finish this way`},
	{Name: "shamil", Aliases: []string{"includes"}, Group: GroupBuiltin, Method: true, Doc: `"salam".shamil("ala")  /  includes — does it contain this`},
	{Name: "jigarre", Aliases: []string{"replace"}, Group: GroupBuiltin, Method: true, Doc: `"salam".jigarre("a", "o")  /  replace — text with parts swapped out`},
	{Name: "tab", Aliases: []string{"split"}, Group: GroupBuiltin, Method: true, Doc: `"a,b".tab(",")  /  split — cut text into a list`},
}

type Operator struct {
	Sym    string
	Prec   int
	Assoc  Assoc
	Punct  bool
	Assign bool
	Doc    string
}

var Operators = []Operator{

	{Sym: "**=", Assign: true, Doc: "replace with x to the power of y"},
	{Sym: "<<=", Assign: true, Doc: "replace with bits shifted left"},
	{Sym: ">>=", Assign: true, Doc: "replace with bits shifted right"},
	{Sym: "+=", Assign: true, Doc: "add and keep"},
	{Sym: "-=", Assign: true, Doc: "subtract and keep"},
	{Sym: "*=", Assign: true, Doc: "multiply and keep"},
	{Sym: "/=", Assign: true, Doc: "divide and keep"},
	{Sym: "%=", Assign: true, Doc: "remainder and keep"},
	{Sym: "&=", Assign: true, Doc: "bitwise and and keep"},
	{Sym: "|=", Assign: true, Doc: "bitwise or and keep"},
	{Sym: "^=", Assign: true, Doc: "bitwise xor and keep"},
	{Sym: ":=", Assign: true, Doc: "assign, also accepted for Go muscle memory"},
	{Sym: "=", Assign: true, Doc: "assign; declaring and assigning are the same act"},

	{Sym: "..+", Prec: 6, Doc: "a range that includes its last number"},
	{Sym: "..", Prec: 6, Doc: "a range up to but not including its last number"},

	{Sym: "<=>", Prec: 5, Doc: "-1, 0 or 1 — compare two orderable values"},
	{Sym: "==", Prec: 5, Doc: "are these the same value"},
	{Sym: "!=", Prec: 5, Doc: "are these different values"},
	{Sym: "<=", Prec: 5, Doc: "is the left side not larger"},
	{Sym: ">=", Prec: 5, Doc: "is the left side not smaller"},
	{Sym: "<", Prec: 5, Doc: "is the left side smaller"},
	{Sym: ">", Prec: 5, Doc: "is the left side larger"},

	{Sym: "&&", Prec: 2, Doc: "logical and"},
	{Sym: "||", Prec: 1, Doc: "logical or"},
	{Sym: "!", Prec: 3, Doc: "logical not"},

	{Sym: "<<", Prec: 7, Doc: "shift the bits left"},
	{Sym: ">>", Prec: 7, Doc: "shift the bits right"},
	{Sym: "&", Prec: 8, Doc: "bitwise and"},
	{Sym: "^", Prec: 9, Doc: "bitwise xor"},
	{Sym: "|", Prec: 10, Doc: "bitwise or"},
	{Sym: "~", Prec: 0, Doc: "flip every bit"},

	{Sym: "+", Prec: 11, Doc: "add"},
	{Sym: "-", Prec: 11, Doc: "subtract"},
	{Sym: "*", Prec: 12, Doc: "multiply"},
	{Sym: "/", Prec: 12, Doc: "divide"},
	{Sym: "%", Prec: 12, Doc: "remainder"},
	{Sym: "**", Prec: 14, Assoc: Right, Doc: "raise to a power"},

	{Sym: "->", Prec: 0, Punct: true, Doc: "the type a function gives back"},
	{Sym: "=>", Prec: 0, Punct: true, Doc: "separates a lambda's arguments from its body"},
	{Sym: "::", Prec: 0, Punct: true, Doc: "the name a value uses in a class"},
	{Sym: "...", Prec: 0, Punct: true, Doc: "spread a list into separate arguments"},
	{Sym: "?", Prec: 0, Punct: true, Doc: "marks a value that may be hich"},
	{Sym: "(", Prec: 0, Punct: true, Doc: "grouping and call"},
	{Sym: ")", Prec: 0, Punct: true, Doc: ""},
	{Sym: "[", Prec: 0, Punct: true, Doc: "index, slice, or list literal"},
	{Sym: "]", Prec: 0, Punct: true, Doc: ""},
	{Sym: "{", Prec: 0, Punct: true, Doc: "open a block"},
	{Sym: "}", Prec: 0, Punct: true, Doc: "close a block"},
	{Sym: ",", Prec: 0, Punct: true, Doc: "separator"},
	{Sym: ";", Prec: 0, Punct: true, Doc: "explicit end of statement"},
	{Sym: ":", Prec: 0, Punct: true, Doc: "block marker, slice bound, or label"},
	{Sym: ".", Prec: 0, Punct: true, Doc: "member access"},
	{Sym: "@", Prec: 0, Punct: true, Doc: "marks the receiver of a method"},
}

var (
	keywordByName      = map[string]*Keyword{}
	keywordByAlias     = map[string]*Keyword{}
	isKeywordKind      = map[Kind]bool{}
	operatorBySym      = map[string]*Operator{}
	operatorsByLen     []string
	maxSymbolLen       int
	operatorPrecedence = map[string]int{}

	ReservedWords = map[string]bool{}
)

func init() {
	for i := range Keywords {
		kw := &Keywords[i]
		isKeywordKind[Kind(kw.Name)] = true
		keywordByName[kw.Name] = kw
		keywordByAlias[kw.Name] = kw
		for _, a := range kw.Aliases {
			keywordByAlias[a] = kw
			ReservedWords[a] = true
		}
		ReservedWords[kw.Name] = true
		if kw.Prec > 0 {
			operatorPrecedence[kw.Name] = kw.Prec
		}
	}

	claimed := map[string]bool{}
	for _, kw := range Keywords {
		for _, s := range append([]string{kw.Name}, kw.Aliases...) {
			if claimed[s] {
				panic("token: " + s + " is claimed by more than one keyword")
			}
			claimed[s] = true
		}
	}
	seen := map[string]bool{}
	for i := range Operators {
		op := &Operators[i]
		if seen[op.Sym] {
			panic("token: duplicate operator symbol " + op.Sym)
		}
		seen[op.Sym] = true
		operatorBySym[op.Sym] = op
		operatorsByLen = append(operatorsByLen, op.Sym)
		if len(op.Sym) > maxSymbolLen {
			maxSymbolLen = len(op.Sym)
		}
	}
	for _, kw := range Keywords {
		for _, a := range kw.Aliases {
			if _, clash := operatorBySym[a]; clash {
				panic("token: alias " + a + " collides with an operator")
			}
		}
	}
	sort.Slice(operatorsByLen, func(i, j int) bool {
		return len(operatorsByLen[i]) > len(operatorsByLen[j])
	})
}

func LookupKeyword(text string) *Keyword { return keywordByAlias[text] }

func IsKeywordKind(k Kind) bool { return isKeywordKind[k] }

func LookupOperator(sym string) *Operator { return operatorBySym[sym] }

func (k Kind) IsOperator() bool { _, ok := operatorBySym[string(k)]; return ok }

func (k Kind) IsPunct() bool {
	op, ok := operatorBySym[string(k)]
	return ok && op.Punct
}

func (k Kind) IsAssignment() bool {
	if op, ok := operatorBySym[string(k)]; ok {
		return op.Assign
	}
	return false
}

func (k Kind) IsDeclaration() bool { return k == "=" || k == ":=" }

func Precedence(k Kind) int {
	if p, ok := operatorPrecedence[string(k)]; ok {
		return p
	}
	if op, ok := operatorBySym[string(k)]; ok {
		if op.Prec > 0 {
			return op.Prec
		}
		return 0
	}
	return 15
}

func AssocOf(k Kind) Assoc {
	if kw, ok := keywordByAlias[string(k)]; ok && kw.Prec > 0 {
		return kw.Assoc
	}
	if op, ok := operatorBySym[string(k)]; ok {
		return op.Assoc
	}
	return Left
}

func MaxSymbolLen() int { return maxSymbolLen }

func SymbolsLongestFirst() []string { return operatorsByLen }

func KeywordList() []Keyword { return Keywords }
