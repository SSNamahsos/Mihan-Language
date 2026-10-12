package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildAndroid(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MIHAN_ANDROID_KEY_DIR", filepath.Join(dir, "key"))
	app := filepath.Join(dir, "app")
	os.MkdirAll(app, 0o755)
	os.WriteFile(filepath.Join(app, "salam.mhn"), []byte("chap_khat \"salam\"\n"), 0o644)
	self, _ := os.Executable()
	out := filepath.Join(dir, "salam.apk")
	if err := buildAndroid(filepath.Join(app, "salam.mhn"), out, "amd64", self, "", "Salam"); err != nil {

		if strings.Contains(err.Error(), "runtime for") {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	z, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]*zip.File{}
	for _, f := range z.File {
		names[f.Name] = f
	}
	for _, n := range []string{"AndroidManifest.xml", "classes.dex", "resources.arsc", "lib/x86_64/libmihanapp.so"} {
		if names[n] == nil {
			t.Fatalf("%s is missing from the APK", n)
		}
	}
	rc, _ := names["AndroidManifest.xml"].Open()
	var m bytes.Buffer
	m.ReadFrom(rc)
	if !bytes.Contains(m.Bytes(), utf16le("org.mihan.app.salam")) && !bytes.Contains(m.Bytes(), []byte("org.mihan.app.salam")) {
		t.Fatal("the manifest was not given the app's package name")
	}
	if bytes.Contains(m.Bytes(), []byte("placeholder")) || bytes.Contains(m.Bytes(), utf16le("placeholder")) {
		t.Fatal("a placeholder is left in the manifest")
	}

	ar := names["resources.arsc"]
	if ar.Method != zip.Store {
		t.Fatal("resources.arsc is compressed")
	}
	if off, _ := ar.DataOffset(); off%4 != 0 {
		t.Fatalf("resources.arsc data is at %d, not 4-byte aligned", off)
	}

	eocd := len(b) - 22
	cd := int(binary.LittleEndian.Uint32(b[eocd+16:]))
	if string(b[cd-16:cd]) != "APK Sig Block 42" {
		t.Fatal("no APK signing block before the central directory")
	}

	k1, c1, _ := androidDebugKey()
	k2, c2, _ := androidDebugKey()
	if !k1.Equal(k2) || !bytes.Equal(c1, c2) {
		t.Fatal("the debug key changed between builds")
	}
}

func utf16le(s string) []byte {
	var b []byte
	for _, r := range s {
		b = append(b, byte(r), 0)
	}
	return b
}

func TestPackageWord(t *testing.T) {
	for in, want := range map[string]string{"Counter": "counter", "my-app": "my_app", "2048": "app2048", "سلام": "app"} {
		if got := packageWord(in); got != want {
			t.Fatalf("packageWord(%q) = %q, want %q", in, got, want)
		}
	}
	if validPackage("org.mihan.app.2x") || !validPackage("org.example.app") {
		t.Fatal("validPackage")
	}
}
