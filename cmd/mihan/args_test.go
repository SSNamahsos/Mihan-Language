package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func mihan(t *testing.T, args ...string) (string, error) {
	t.Helper()

	self := compilerForTest(t)
	cmd := exec.Command(self, args...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func writeScript(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "s.mhn")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestAProgramReadsTheWordsAfterItsName(t *testing.T) {
	p := writeScript(t, "chap_khat arghoman()\n")
	got, err := mihan(t, "run", p, "alpha", "beta")
	if err != nil {
		t.Fatal(err)
	}
	if want := "[\"alpha\", \"beta\"]\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAProgramWithNoWordsIsToldThereAreNone(t *testing.T) {
	p := writeScript(t, "chap_khat tule(arghoman())\n")
	got, err := mihan(t, "run", p)
	if err != nil {
		t.Fatal(err)
	}
	if want := "0\n"; got != want {
		t.Errorf("got %q, want %q — a program has to be able to tell no words from empty ones", got, want)
	}
}

func TestOneArgumentAtATime(t *testing.T) {
	p := writeScript(t, "chap_khat arghoman(0)\nchap_khat arghoman(2)\n")
	got, err := mihan(t, "run", p, "one", "two", "three")
	if err != nil {
		t.Fatal(err)
	}
	if want := "one\nthree\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAskingForAnArgumentThatIsNotThereIsAMistake(t *testing.T) {
	p := writeScript(t, "chap_khat arghoman(5)\n")
	_, err := mihan(t, "run", p, "one", "two")
	if err == nil {
		t.Fatal("asking for a fifth of two arguments was not a mistake")
	}
	out, _ := mihan(t, "run", p, "one", "two")
	if !strings.Contains(out, "E4123") && !strings.Contains(out, "2") {
		t.Errorf("the mistake did not say how many there were: %s", out)
	}
}

func TestTheCommandsOwnWordsAreNotThePrograms(t *testing.T) {
	p := writeScript(t, "chap_khat arghoman()\n")
	got, err := mihan(t, "run", p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "run") || strings.Contains(got, ".mhn") {
		t.Errorf("the program was handed the command's own words: %q", got)
	}
}

func TestANumberFromTheCommandLineIsAWholeNumber(t *testing.T) {
	p := writeScript(t, "chap_khat adad_karda(\"42\") + 1\nchap_khat adad_karda(\"1.5\")\n")
	got, err := mihan(t, "run", p)
	if err != nil {
		t.Fatal(err)
	}
	if want := "43\n1.5\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAWordThatIsNotANumberSaysWhich(t *testing.T) {
	p := writeScript(t, "chap_khat adad_karda(arghoman(0))\n")
	out, _ := mihan(t, "run", p, "12x")
	if !strings.Contains(out, "12x") {
		t.Errorf("the mistake did not quote the word it could not read: %s", out)
	}
}

func TestATestFileSaysHowItShouldBeRun(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "s.mhn")
	body := "#barghoman: 7\n#khoroji\n#| 7\n#| 8\n\nchap_khat adad_karda(arghoman(0))\nchap_khat adad_karda(arghoman(0)) + 1\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := mihan(t, "test", dir)
	if err != nil {
		t.Fatalf("the script did not pass its own test: %v\n%s", err, got)
	}
	if !strings.Contains(got, "sabet") {
		t.Errorf("the runner did not report the script as passing:\n%s", got)
	}
}

func TestTheArgumentMarkerIsNotAnExpectation(t *testing.T) {

	dir := t.TempDir()
	p := filepath.Join(dir, "s.mhn")
	body := "#barghoman: 7\n#khoroji\n#| seven\n\nchap_khat \"seven\"\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := mihan(t, "test", dir)
	if err != nil {
		t.Fatalf("the script did not pass: %v\n%s", err, got)
	}
	if strings.Contains(got, `wanted "7"`) {
		t.Errorf("the arguments were read as the expected output:\n%s", got)
	}
}
