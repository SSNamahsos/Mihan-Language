package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"mihan/internal/diag"
	"mihan/internal/driver"
	"mihan/internal/parser"
	"mihan/internal/version"
)

func cmdTest(args []string) error {
	dirs := args
	if len(dirs) == 0 {

		dirs = []string{"docs", "tests", "examples"}
	}
	type result struct {
		path string
		ok   bool
		note string
	}
	var results []result
	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			return fatal("%s is not here", dir)
		}

		if !info.IsDir() {
			if !driver.IsMihanFile(dir) {
				return fatal("%s is not a Mihan file", dir)
			}
			ok, note := runOneTest(dir)
			results = append(results, result{path: dir, ok: ok, note: note})
			continue
		}
		var files []string
		err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil || fi.IsDir() || !driver.IsMihanFile(path) {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return err
		}
		sort.Strings(files)
		for _, path := range files {
			ok, note := runOneTest(path)
			r := result{path: path, ok: ok, note: note}
			results = append(results, r)
		}
	}

	if len(results) == 0 {
		return fatal("I found no .%s files to run", driver.Extension)
	}
	failed := 0
	for _, r := range results {
		mark := "salam"
		if !r.ok {
			mark = "KHATA"
			failed++
		}
		fmt.Printf("%-6s %-46s %s\n", mark, r.path, r.note)
	}
	fmt.Println()
	if failed == 0 {
		fmt.Printf("%d/%d sabet. %s\n", len(results), len(results), version.Short())
		return nil
	}
	fmt.Printf("%d/%d shekht. %d bar sai konid.\n", failed, len(results), failed)
	return errSilent
}

const serveGrace = 3 * time.Second

func runOneTest(path string) (ok bool, note string) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err.Error()
	}
	kind, expect := expectationOf(src)
	if kind == expectNone {

		if block := strayBlock(src); block != 0 {
			return false, fmt.Sprintf(
				"line %d starts a `#|` block, but no `#khoroji` marker is above it; the lines after it are code, not expected output",
				block)
		}
	}

	self, err := os.Executable()
	if err != nil {
		return false, fmt.Sprintf("cannot find the compiler to run the program with: %v", err)
	}

	cmd := exec.Command(self, append([]string{"run", path}, argumentsOf(src)...)...)

	cmd.Env = append(os.Environ(), "MIHAN_NO_WINDOW=1")

	var out, errs lockedBuffer
	cmd.Stdout = &out
	cmd.Stderr = &errs
	cmd.Stdin = strings.NewReader("")

	cmd.WaitDelay = 2 * time.Second
	setProcessGroup(cmd)
	if err := cmd.Start(); err != nil {
		return false, err.Error()
	}
	stopped := false
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	defer func() {
		if !stopped {
			killTree(cmd.Process)

			select {
			case <-finished:
			case <-time.After(5 * time.Second):
			}
		}
	}()

	select {
	case err := <-finished:
		stopped = true
		got := out.String()
		failed := err != nil
		if kind == expectError || kind == expectPanic {
			if !failed {
				return false, "I expected this to fail and it did not"
			}
			return true, firstCodeIn(errs.String())
		}
		if failed {
			return false, summaryOf(errs.String())
		}
		if expect == "" {
			return true, strings.TrimSpace(got)
		}
		if matches(expect, got) {
			return true, "kharch"
		}
		return false, fmt.Sprintf("wanted %q, got %q", expect, got)
	case <-time.After(serveGrace):
		got := out.String()
		switch kind {
		case expectNone:

			note := strings.TrimSpace(got)
			if note != "" {
				return true, "serves: " + note
			}
			return true, "serves"
		case expectOutput:
			if matches(expect, got) {
				return true, "serves"
			}
			return false, fmt.Sprintf("wanted %q before it started serving, got %q", expect, got)
		case expectError:
			return false, "I expected this to fail and it did not; it is still running"
		default:
			return false, fmt.Sprintf("this program is still running after %s and says what it should do, so it should have finished", serveGrace)
		}
	}
}

func matches(expect, got string) bool {
	expLines := splitKeepEmpty(expect)
	gotLines := splitKeepEmpty(got)
	return matchLines(expLines, gotLines)
}

func splitKeepEmpty(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")

	if n := len(lines); n > 0 && lines[n-1] == "" {
		lines = lines[:n-1]
	}
	return lines
}

func matchLines(exp, got []string) bool {
	if len(exp) == 0 {
		return len(got) == 0
	}

	if exp[0] == "*" {
		for i := 0; i <= len(got); i++ {
			if matchLines(exp[1:], got[i:]) {
				return true
			}
		}
		return false
	}
	if len(got) == 0 || !matchLine(exp[0], got[0]) {
		return false
	}
	return matchLines(exp[1:], got[1:])
}

func matchLine(exp, got string) bool {
	parts := strings.Split(exp, "*")
	if len(parts) == 1 {
		return exp == got
	}
	rest := got
	for i, part := range parts {
		if i == 0 {
			if !strings.HasPrefix(rest, part) {
				return false
			}
			rest = rest[len(part):]
			continue
		}
		if i == len(parts)-1 {
			return strings.HasSuffix(rest, part)
		}
		at := strings.Index(rest, part)
		if at < 0 {
			return false
		}
		rest = rest[at+len(part):]
	}
	return true
}

func firstCodeIn(text string) string {
	i := strings.Index(text, "error: E")
	if i < 0 {
		return "KHATA"
	}
	rest := text[i+len("error: "):]
	end := 0
	for end < len(rest) && (rest[end] == 'E' || rest[end] >= '0' && rest[end] <= '9') {
		end++
	}
	if end == 0 {
		return "KHATA"
	}
	return rest[:end]
}

func summaryOf(text string) string {
	var codes []string
	for _, d := range strings.Split(text, "\n") {
		if c := firstCodeIn(d); c != "KHATA" {
			codes = append(codes, c)
		}
	}
	if len(codes) == 0 {
		return "the program failed and said nothing about why"
	}
	return strings.Join(codes, "; ")
}

func killTree(p *os.Process) {
	if p == nil {
		return
	}
	if runtime.GOOS == "windows" {

		_ = exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(p.Pid)).Run()
		return
	}
	killGroup(p)
	_ = p.Kill()
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type expectKind int

const (
	expectNone expectKind = iota
	expectOutput
	expectError
	expectPanic
)

func strayBlock(src []byte) int {
	marked := false
	for i, line := range strings.Split(string(stripBOM(src)), "\n") {
		t := strings.TrimSpace(line)
		if kind, name, rest, ok := markerKind(t); ok {

			if kind == expectOutput && name == "khoroji" && rest == "" {
				marked = true
			}
			continue
		}
		if _, isBlock := blockLine(t); isBlock {
			if !marked {
				return i + 1
			}
			continue
		}
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}

		marked = false
	}
	return 0
}

func argumentsOf(src []byte) []string {
	var out []string
	for _, raw := range strings.Split(string(stripBOM(src)), "\n") {
		line := strings.TrimSpace(raw)
		_, name, rest, ok := markerKind(line)
		if !ok || name != "barghoman" {
			continue
		}
		out = strings.Fields(rest)
	}
	return out
}

func markerKind(line string) (kind expectKind, name, rest string, ok bool) {
	body, has := strings.CutPrefix(line, "#")
	if !has {
		return expectNone, "", "", false
	}
	body = strings.TrimSpace(body)
	switch {
	case body == "khoroji":
		return expectOutput, "khoroji", "", true
	case strings.HasPrefix(body, "khoroji:"):
		return expectOutput, "khoroji", strings.TrimSpace(strings.TrimPrefix(body, "khoroji:")), true
	case strings.HasPrefix(body, "khata:"):
		return expectError, "khata", strings.TrimSpace(strings.TrimPrefix(body, "khata:")), true
	case strings.HasPrefix(body, "barghoman:") || strings.HasPrefix(body, "args:"):

		return expectNone, "barghoman", strings.TrimSpace(rest2(body)), true
	}
	return expectNone, "", "", false
}

func rest2(body string) string {
	if _, r, ok := strings.Cut(body, ":"); ok {
		return r
	}
	return ""
}

func blockLine(line string) (string, bool) {
	rest, ok := strings.CutPrefix(line, "#|")
	if !ok {
		return "", false
	}

	return strings.TrimPrefix(rest, " "), true
}

func stripBOM(src []byte) []byte {
	if len(src) >= 3 && src[0] == 0xEF && src[1] == 0xBB && src[2] == 0xBF {
		return src[3:]
	}
	return src
}

func expectationOf(src []byte) (expectKind, string) {
	var lines []string
	inBlock := false
	for _, raw := range strings.Split(string(stripBOM(src)), "\n") {
		line := strings.TrimSpace(raw)
		if inBlock {

			want, ok := blockLine(line)
			if !ok {
				break
			}
			lines = append(lines, want)
			continue
		}
		kind, name, rest, ok := markerKind(line)
		if !ok {
			continue
		}
		if kind == expectError {
			return expectError, ""
		}

		if name == "barghoman" {
			continue
		}
		if name == "khoroji" && rest == "" {
			inBlock = true
			continue
		}
		return expectOutput, unquote(rest)
	}
	if inBlock {
		return expectOutput, strings.Join(lines, "\n") + "\n"
	}
	return expectNone, ""
}

func unquote(s string) string {
	if len(s) >= 2 && s[0] == '`' && s[len(s)-1] == '`' {
		return s[1 : len(s)-1]
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		inner := s[1 : len(s)-1]
		inner = strings.ReplaceAll(inner, `\"`, `"`)
		inner = strings.ReplaceAll(inner, `\\`, `\`)
		inner = strings.ReplaceAll(inner, `\n`, "\n")
		inner = strings.ReplaceAll(inner, `\t`, "\t")
		return inner
	}
	return s
}

func firstCode(ds []*diag.Diagnostic) string {
	for _, d := range ds {
		if d.Severity == diag.Error {
			return d.Code
		}
	}
	return ""
}

func looksComplete(src string) bool {
	bag := &diag.Bag{}
	parser.Parse("<repl>", []byte(src), bag)
	if !bag.HasErrors() {
		return true
	}
	end := len(strings.TrimRight(src, " \t\r\n"))
	for _, d := range bag.Sorted() {
		if d.Span.End.Off < end {
			return true
		}
	}
	return false
}

func replBanner() string {
	return fmt.Sprintf(`%s

  baraye khorod: mihan > 2 + 2
  baraye khanodan: chap "salam"
  baraye khoroj: exit

`, version.String())
}

func replHelp() string {
	return `Mihan dar haal-e khatere:

  <har chizi>      any expression or statement
  help            in hamin matn
  clear           saf kon inSafhe
  exit            khoroj

  Ctrl+D ya exit  khoruj
  A line ending in an open bracket waits for the rest, so a long
  expression can be spread over several lines.
`
}
