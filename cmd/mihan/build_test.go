package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildMakesOneExecutableThatRunsTheProgram(t *testing.T) {
	mihan := compilerForTest(t)
	dir := t.TempDir()
	app := filepath.Join(dir, "app")
	if err := os.MkdirAll(filepath.Join(app, "safhe"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(app, "safhe", "i.html"), []byte("<h1>hi</h1>"), 0o644)
	os.WriteFile(filepath.Join(app, "main.mhn"), []byte(
		"chap_khat arghoman()\nchap_khat bekhan(masir(poshe_barname(), \"safhe\", \"i.html\"))\n"), 0o644)
	out := filepath.Join(dir, "built-app-test")
	if b, err := exec.Command(mihan, "build", app, "-o", out).CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, b)
	}
	cache := filepath.Join(dir, "cache")
	cmd := exec.Command(out, "a", "b")
	cmd.Env = append(os.Environ(), "XDG_CACHE_HOME="+cache, "LOCALAPPDATA="+cache, "HOME="+dir)
	b, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, b)
	}
	if got, want := strings.ReplaceAll(string(b), "\r\n", "\n"), "[\"a\", \"b\"]\n<h1>hi</h1>\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestBuildForWindowsSetsTheSubsystem(t *testing.T) {
	mihan := compilerForTest(t)
	dir := t.TempDir()
	winRT := filepath.Join(dir, "mihan-windows-amd64.exe")
	gb := exec.Command("go", "build", "-o", winRT, ".")
	gb.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64")
	if b, err := gb.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, b)
	}
	app := filepath.Join(dir, "w.mhn")
	os.WriteFile(app, []byte("chap_khat 1\n"), 0o644)
	for _, tc := range []struct {
		flag string
		want uint16
	}{{"--window", 2}, {"", 3}} {
		out := filepath.Join(dir, "w"+tc.flag+".exe")
		args := []string{"build", app, "--os", "windows", "-o", out}
		if tc.flag != "" {
			args = append(args, tc.flag)
		}
		cmd := exec.Command(mihan, args...)
		cmd.Env = append(os.Environ(), "MIHAN_RUNTIMES="+dir)
		if b, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %v: %v\n%s", args, err, b)
		}
		b, _ := os.ReadFile(out)
		off := int(b[0x3c]) | int(b[0x3d])<<8 | int(b[0x3e])<<16 | int(b[0x3f])<<24
		if got := uint16(b[off+92]) | uint16(b[off+93])<<8; got != tc.want {
			t.Errorf("%q: subsystem %d, want %d", tc.flag, got, tc.want)
		}
	}

	b, err := exec.Command(mihan, "build", app, "--os", "linux", "--arch", "amd64", "--base", winRT, "-o", filepath.Join(dir, "x")).CombinedOutput()
	if err == nil || !strings.Contains(string(b), "windows/amd64") {
		t.Errorf("wrong runtime: %v %s", err, b)
	}
}
