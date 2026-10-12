package eval

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"mihan/internal/ast"
	"mihan/internal/diag"
	"mihan/internal/parser"
	rt "mihan/internal/runtime"
)

func writeFiles(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func runIn(t *testing.T, dir, src string) (string, error) {
	t.Helper()
	main := filepath.Join(dir, "main.mhn")
	if err := os.WriteFile(main, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	bag := &diag.Bag{}
	file := parser.Parse(main, []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read: %s", bag.Sorted()[0].EN)
	}
	var out strings.Builder
	in := New(&out, strings.NewReader(""))
	if err := in.RunFile(file); err != nil {
		return out.String(), err
	}
	return out.String(), nil
}

const rulesFile = "funct dar_khush_amad(name) -> matn\n" +
	"    return \"salam \" + name\n" +
	"number = 42\n"

func TestAnImportedFileGivesItsNames(t *testing.T) {
	dir := writeFiles(t, map[string]string{"rules.mhn": rulesFile})
	got, err := runIn(t, dir, "biar \"rules.mhn\"\n"+
		"chap_khat dar_khush_amad(\"mihan\")\n"+
		"chap_khat number\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "salam mihan\n42\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnAliasPutsTheNamesBehindOneName(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"hesab.mhn": "funct jam(a, b) -> arz\n    return a + b\nbenevis = \"from hesab\"\n",
		"rang.mhn":  "benevis = \"from rang\"\n",
	})
	got, err := runIn(t, dir, "biar \"hesab.mhn\" as h\n"+
		"biar \"rang.mhn\" as r\n"+
		"chap_khat h.benevis\n"+
		"chap_khat r.benevis\n"+
		"chap_khat h.jam(2, 3)\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "from hesab\nfrom rang\n5\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAFileImportedTwiceRunsOnce(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"rules.mhn": "chap_khat \"rules ran\"\njam = 7\n",
		"one.mhn":   "biar \"rules.mhn\"\n",
		"two.mhn":   "biar \"rules.mhn\"\n",
	})
	got, err := runIn(t, dir, "biar \"one.mhn\" as a\n"+
		"biar \"two.mhn\" as b\n"+
		"biar \"rules.mhn\"\n"+
		"chap_khat jam\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "rules ran\n7\n"; got != want {
		t.Errorf("got %q, want %q — a shared file should run once however many want it", got, want)
	}
}

func TestAFileThatImportsItselfIsStopped(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"a.mhn": "biar \"b.mhn\"\n",
		"b.mhn": "biar \"a.mhn\"\n",
	})
	_, err := runIn(t, dir, "biar \"a.mhn\"\n")
	if err == nil {
		t.Fatal("a circle of imports ran for ever instead of being stopped")
	}
	msg := err.Error()
	if !strings.Contains(msg, "دایره") && !strings.Contains(msg, "circle") {
		t.Errorf("the mistake did not say it was a circle: %v", err)
	}

	for _, want := range []string{"a.mhn", "b.mhn"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the mistake did not name %s, so the reader cannot see which file to open: %v", want, err)
		}
	}
}

func TestAFileThatImportsItselfDirectlyIsStopped(t *testing.T) {
	dir := writeFiles(t, map[string]string{"a.mhn": "biar \"a.mhn\"\n"})
	if _, err := runIn(t, dir, "biar \"a.mhn\"\n"); err == nil {
		t.Fatal("a file importing itself ran for ever instead of being stopped")
	}
}

func TestAMissingFileIsNamedAndSaysWhereItLooked(t *testing.T) {
	dir := writeFiles(t, nil)
	_, err := runIn(t, dir, "biar \"nowhere.mhn\"\n")
	if err == nil {
		t.Fatal("a missing file was not a mistake")
	}
	msg := err.Error()
	if !strings.Contains(msg, "nowhere.mhn") {
		t.Errorf("the mistake did not name the file it could not find: %v", err)
	}

	var p *rt.Panic
	if !errors.As(err, &p) {
		t.Fatalf("the mistake did not come back as a fault: %v", err)
	}
	if e, ok := p.Value.(*rt.ErrorValue); !ok || e.Name != "E4009" {
		t.Errorf("the mistake had the wrong code, so nothing could match on it: %v", err)
	}
}

func TestAMistakeInsideAnImportedFileNamesThatFile(t *testing.T) {
	dir := writeFiles(t, map[string]string{"broken.mhn": "x = \n"})
	_, err := runIn(t, dir, "biar \"broken.mhn\"\n")
	if err == nil {
		t.Fatal("a file that does not read was not a mistake")
	}
	if !strings.Contains(err.Error(), "broken.mhn") {
		t.Errorf("the mistake did not say which file it was in: %v", err)
	}
}

func TestAFileInAFolderIsFound(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		filepath.Join("lib", "rules.mhn"): "jam = 3\n",
	})
	got, err := runIn(t, dir, "biar \"lib/rules.mhn\"\nchap_khat jam\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "3\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTwoProgramsInOneProcessDoNotShareImports(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"rules.mhn": "chap_khat \"once\"\njam = 1\n",
	})
	src := "biar \"rules.mhn\"\nchap_khat jam\n"
	first, err := runIn(t, dir, src)
	if err != nil {
		t.Fatalf("the first program stopped: %v", err)
	}
	second, err := runIn(t, dir, src)
	if err != nil {
		t.Fatalf("the second program stopped: %v", err)
	}
	if first != second {
		t.Errorf("two runs of one program gave different answers: %q then %q", first, second)
	}
	if want := "once\n1\n"; first != want {
		t.Errorf("got %q, want %q", first, want)
	}
}

func TestAnImportDoesNotOverwriteANameAlreadyThere(t *testing.T) {
	dir := writeFiles(t, map[string]string{"rules.mhn": "jam = 99\n"})
	got, err := runIn(t, dir, "jam = 1\nbiar \"rules.mhn\"\nchap_khat jam\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "1\n"; got != want {
		t.Errorf("got %q, want %q — an import must not overwrite a name that is already there", got, want)
	}
}

func TestAnImportedFileMayImportAnother(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"base.mhn": "jam = 2\n",
		"mid.mhn":  "biar \"base.mhn\"\ndonble = jam * 10\n",
	})
	got, err := runIn(t, dir, "biar \"mid.mhn\"\nchap_khat donble\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "20\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAnImportedFilePrintsToTheSamePlace(t *testing.T) {
	dir := writeFiles(t, map[string]string{
		"hello.mhn": "chap_khat \"from the other file\"\n",
	})
	got, err := runIn(t, dir, "biar \"hello.mhn\"\nchap_khat \"and from this one\"\n")
	if err != nil {
		t.Fatalf("the program stopped: %v", err)
	}
	if want := "from the other file\nand from this one\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBiarNeedsAPath(t *testing.T) {
	bag := &diag.Bag{}
	parser.Parse("t.mhn", []byte("biar\n"), bag)
	if !bag.HasErrors() {
		t.Error("`biar` with no path was read, but a reader would not know what to bring in")
	}
	if got := importPathOf(parseOK(t, "biar \"a.mhn\"")); got != "a.mhn" {
		t.Errorf("the path was read as %q", got)
	}
	if got := importAliasOf(parseOK(t, "biar \"a.mhn\" as m")); got != "m" {
		t.Errorf("the short name was read as %q", got)
	}
	if got := importAliasOf(parseOK(t, "biar \"a.mhn\"")); got != "" {
		t.Errorf("a plain biar invented a short name: %q", got)
	}
}

func parseOK(t *testing.T, src string) *ast.File {
	t.Helper()
	bag := &diag.Bag{}
	f := parser.Parse("t.mhn", []byte(src), bag)
	if bag.HasErrors() {
		t.Fatalf("the program did not read: %s", bag.Sorted()[0].EN)
	}
	return f
}

func importPathOf(f *ast.File) string {
	for _, s := range f.Decls {
		if d, ok := s.(*ast.ImportDecl); ok {
			return d.Path
		}
	}
	return ""
}

func importAliasOf(f *ast.File) string {
	for _, s := range f.Decls {
		if d, ok := s.(*ast.ImportDecl); ok {
			return d.Alias
		}
	}
	return ""
}
