package main

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func startPreview(t *testing.T, src string, extra ...string) (string, *bufio.Reader) {
	t.Helper()
	bin := compilerForTest(t)
	dir := t.TempDir()
	file := filepath.Join(dir, "app.mhn")
	if err := os.WriteFile(file, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	args := append([]string{"preview", file, "--android", "--no-open", "--port", strconv.Itoa(port)}, extra...)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cmd.Process.Kill(); cmd.Wait() })
	base := "http://127.0.0.1:" + strconv.Itoa(port)
	var resp *http.Response
	for i := 0; i < 100; i++ {
		if resp, err = http.Get(base + "/_preview/events"); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatal("the preview never answered:", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return base, bufio.NewReader(resp.Body)
}

func waitFor(t *testing.T, ev *bufio.Reader, kind, text string) {
	t.Helper()
	done := make(chan bool, 1)
	go func() {
		for {
			line, err := ev.ReadString('\n')
			if err != nil {
				done <- false
				return
			}
			var e struct{ Kind, Text string }
			if strings.HasPrefix(line, "data: ") && json.Unmarshal([]byte(line[6:]), &e) == nil &&
				e.Kind == kind && strings.Contains(e.Text, text) {
				done <- true
				return
			}
		}
	}()
	select {
	case ok := <-done:
		if !ok {
			t.Fatalf("the events ended before %s %q", kind, text)
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("no %s event with %q", kind, text)
	}
}

func state(t *testing.T, base string) map[string]any {
	t.Helper()
	resp, err := http.Get(base + "/_preview/state")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var st map[string]any
	json.NewDecoder(resp.Body).Decode(&st)
	return st
}

func TestPreviewOfAProgramThatPrints(t *testing.T) {
	base, ev := startPreview(t, "chap_khat(\"salam\")\nagar android_hast()\n    chap_khat(\"android\")\n", "--device", "small")
	waitFor(t, ev, "exit", "ended")
	st := state(t, base)
	if st["output"] != "salam\nandroid\n" || st["device"] != "small" || st["url"] != "" {
		t.Fatalf("state: %v", st)
	}
	resp, _ := http.Get(base + "/")
	page, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(page), "360px") || strings.Contains(string(page), "{{w}}") {
		t.Fatal("the phone page does not carry the small device's size")
	}
}

func TestPreviewOfAServerLogsRequestsAndThePhone(t *testing.T) {
	base, ev := startPreview(t, "funct pasokh(d)\n    bargard \"<html><head></head><body>hi</body></html>\"\nserva_bgo(0, pasokh)\n")
	waitFor(t, ev, "url", "http://127.0.0.1:")
	url, _ := state(t, base)["url"].(string)
	resp, err := http.Get(url + "/")
	if err != nil {
		t.Fatal(err)
	}
	page, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(page), `<script src="/_mihan/mihan.js"></script>`) {
		t.Fatalf("the Android page has no phone script: %q", page)
	}
	waitFor(t, ev, "req", "GET / 200")
	resp, _ = http.Get(url + "/_mihan/mihan.js")
	js, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(js), "MihanAndroid") {
		t.Fatal("mihan.js is not the bridge")
	}
	http.Post(base+"/_preview/do?what=toast&text=salam", "", nil)
	waitFor(t, ev, "bridge", "toast: salam")
	http.Post(base+"/_preview/do?what=finish", "", nil)
	waitFor(t, ev, "life", "")
}
