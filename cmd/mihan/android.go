package main

import (
	"archive/zip"
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"hash/crc32"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"
)

//go:embed android/template.apk
var androidTemplate []byte

const (
	pkgPlaceholder   = "org.mihan.placeholder.app.aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	labelPlaceholder = "MihanAppLabelPlaceholder.................................."
	androidLib       = "libmihanapp.so"
)

var androidABIs = map[string]string{"arm64": "arm64-v8a", "amd64": "x86_64"}

func buildAndroid(path, out, arch, base, pkg, label string) error {
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if name == "main" {
		if abs, err := filepath.Abs(filepath.Dir(path)); err == nil {
			name = filepath.Base(abs)
		}
	}
	if out == "" {
		out = name + ".apk"
	}
	if label == "" {
		label = name
	}
	if pkg == "" {
		pkg = "org.mihan.app." + packageWord(name)
	}
	if !validPackage(pkg) {
		return fatal("%s is not a usable Android package name: use words of letters, digits and _ joined by dots, like org.example.app", pkg)
	}
	archs := []string{arch}
	if arch == "all" {
		archs = []string{"arm64", "amd64"}
	}
	if base != "" && len(archs) > 1 {
		return fatal("--base names one runtime, so pick one --arch with it")
	}
	payload, files, err := packFolder(filepath.Dir(path), filepath.Base(path), out, false)
	if err != nil {
		return err
	}
	libs := map[string][]byte{}
	for _, a := range archs {
		rtPath := base
		if rtPath == "" {

			if rtPath, err = findRuntime("linux", a); err != nil {
				return err
			}
		}
		exe, err := os.ReadFile(rtPath)
		if err != nil {
			return fatal("%s cannot be read: %v", rtPath, err)
		}
		exe = stripPayload(exe)
		if gotOS, gotArch := binaryTarget(exe); gotOS != "linux" || gotArch != a {
			return fatal("%s is a runtime for %s/%s, not linux/%s, which an Android %s build needs", rtPath, gotOS, gotArch, a, a)
		}
		libs["lib/"+androidABIs[a]+"/"+androidLib] = withPayload(exe, payload)
	}
	key, cert, err := androidDebugKey()
	if err != nil {
		return fatal("cannot make the signing key: %v", err)
	}
	apk, err := makeAPK(pkg, label, libs, key, cert)
	if err != nil {
		return fatal("cannot make the APK: %v", err)
	}
	if err := os.WriteFile(out, apk, 0o644); err != nil {
		return fatal("cannot write %s: %v", out, err)
	}
	fmt.Printf("built %s (android/%s, package %s, %d files, %.1f MB, signed with a debug key)\n",
		out, arch, pkg, files, float64(len(apk))/1e6)
	return nil
}

func packageWord(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-' || r == ' ' || r == '.':
			b.WriteByte('_')
		}
	}
	w := b.String()
	if w == "" || w[0] < 'a' || w[0] > 'z' {
		w = "app" + w
	}
	return w
}

func validPackage(p string) bool {
	parts := strings.Split(p, ".")
	if len(parts) < 2 || len(p) > len(pkgPlaceholder) {
		return false
	}
	for _, s := range parts {
		if s == "" || !(s[0] >= 'a' && s[0] <= 'z' || s[0] >= 'A' && s[0] <= 'Z') {
			return false
		}
		for _, r := range s {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_') {
				return false
			}
		}
	}
	return true
}

func rewriteManifest(xml []byte, repl map[string]string) ([]byte, error) {
	if len(xml) < 8+28 || binary.LittleEndian.Uint16(xml) != 0x0003 {
		return nil, fmt.Errorf("the template manifest is not binary XML")
	}
	sp := xml[8:]
	if binary.LittleEndian.Uint16(sp) != 0x0001 {
		return nil, fmt.Errorf("the template manifest has no string pool first")
	}
	hdr := int(binary.LittleEndian.Uint16(sp[2:]))
	size := int(binary.LittleEndian.Uint32(sp[4:]))
	count := int(binary.LittleEndian.Uint32(sp[8:]))
	styles := int(binary.LittleEndian.Uint32(sp[12:]))
	flags := binary.LittleEndian.Uint32(sp[16:])
	start := int(binary.LittleEndian.Uint32(sp[20:]))
	if styles != 0 {
		return nil, fmt.Errorf("the template manifest has styled strings")
	}
	utf8 := flags&0x100 != 0
	strs := make([]string, count)
	for i := 0; i < count; i++ {
		off := start + int(binary.LittleEndian.Uint32(sp[hdr+4*i:]))
		strs[i] = decodePoolString(sp[off:], utf8)
	}
	found := 0
	for i, s := range strs {
		if r, ok := repl[s]; ok {
			strs[i] = r
			found++
		}
	}
	if found != len(repl) {
		return nil, fmt.Errorf("the template manifest is missing its placeholders")
	}
	var data bytes.Buffer
	offs := make([]uint32, count)
	for i, s := range strs {
		offs[i] = uint32(data.Len())
		data.Write(encodePoolString(s, utf8))
	}
	for data.Len()%4 != 0 {
		data.WriteByte(0)
	}
	var pool bytes.Buffer
	pool.Write(sp[:hdr])
	for _, o := range offs {
		binary.Write(&pool, binary.LittleEndian, o)
	}
	newStart := pool.Len()
	pool.Write(data.Bytes())
	p := pool.Bytes()
	binary.LittleEndian.PutUint32(p[4:], uint32(len(p)))
	binary.LittleEndian.PutUint32(p[20:], uint32(newStart))
	binary.LittleEndian.PutUint32(p[24:], 0)

	var outb bytes.Buffer
	outb.Write(xml[:8])
	outb.Write(p)
	outb.Write(xml[8+size:])
	o := outb.Bytes()
	binary.LittleEndian.PutUint32(o[4:], uint32(len(o)))
	return o, nil
}

func decodePoolString(b []byte, utf8 bool) string {
	if utf8 {
		i := 0
		if b[i]&0x80 != 0 {
			i += 2
		} else {
			i++
		}
		n := int(b[i])
		if b[i]&0x80 != 0 {
			n = int(b[i]&0x7f)<<8 | int(b[i+1])
			i += 2
		} else {
			i++
		}
		return string(b[i : i+n])
	}
	n := int(binary.LittleEndian.Uint16(b))
	i := 2
	if n&0x8000 != 0 {
		n = (n&0x7fff)<<16 | int(binary.LittleEndian.Uint16(b[2:]))
		i = 4
	}
	u := make([]uint16, n)
	for k := range u {
		u[k] = binary.LittleEndian.Uint16(b[i+2*k:])
	}
	return string(utf16.Decode(u))
}

func encodePoolString(s string, utf8 bool) []byte {
	var b bytes.Buffer
	if utf8 {
		lenField := func(n int) {
			if n > 0x7f {
				b.WriteByte(byte(n>>8) | 0x80)
			}
			b.WriteByte(byte(n))
		}
		lenField(len(utf16.Encode([]rune(s))))
		lenField(len(s))
		b.WriteString(s)
		b.WriteByte(0)
		return b.Bytes()
	}
	u := utf16.Encode([]rune(s))
	if len(u) > 0x7fff {
		binary.Write(&b, binary.LittleEndian, uint16(len(u)>>16|0x8000))
	}
	binary.Write(&b, binary.LittleEndian, uint16(len(u)))
	for _, c := range u {
		binary.Write(&b, binary.LittleEndian, c)
	}
	b.Write([]byte{0, 0})
	return b.Bytes()
}

type apkEntry struct {
	name  string
	data  []byte
	store bool
}

func makeAPK(pkg, label string, libs map[string][]byte, key *rsa.PrivateKey, cert []byte) ([]byte, error) {
	tz, err := zip.NewReader(bytes.NewReader(androidTemplate), int64(len(androidTemplate)))
	if err != nil {
		return nil, err
	}
	var entries []apkEntry
	for _, f := range tz.File {
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		if f.Name == "AndroidManifest.xml" {
			if b, err = rewriteManifest(b, map[string]string{pkgPlaceholder: pkg, labelPlaceholder: label}); err != nil {
				return nil, err
			}
		}
		e := apkEntry{name: f.Name, data: b, store: f.Name == "resources.arsc"}
		if e.store {

			entries = append([]apkEntry{e}, entries...)
			continue
		}
		entries = append(entries, e)
	}
	names := make([]string, 0, len(libs))
	for n := range libs {
		names = append(names, n)
	}
	sortStrings(names)
	for _, n := range names {
		entries = append(entries, apkEntry{name: n, data: libs[n]})
	}
	unsigned, err := writeAlignedZip(entries)
	if err != nil {
		return nil, err
	}
	return signV2(unsigned, key, cert)
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func writeAlignedZip(entries []apkEntry) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	stamp := time.Date(2008, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, e := range entries {
		h := &zip.FileHeader{Name: e.name, Method: zip.Deflate, Modified: stamp}
		if e.store {
			if err := zw.Flush(); err != nil {
				return nil, err
			}
			h.Method = zip.Store
			h.CRC32 = crc32.ChecksumIEEE(e.data)
			h.CompressedSize64 = uint64(len(e.data))
			h.UncompressedSize64 = uint64(len(e.data))

			at := buf.Len() + 30 + len(e.name)
			pad := (4 - (at+4)%4) % 4
			extra := make([]byte, 4+pad)
			binary.LittleEndian.PutUint16(extra, 0xd935)
			binary.LittleEndian.PutUint16(extra[2:], uint16(pad))
			h.Extra = extra
			w, err := zw.CreateRaw(h)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write(e.data); err != nil {
				return nil, err
			}
			continue
		}
		w, err := zw.CreateHeader(h)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(e.data); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const sigRSAPKCS1SHA256 = 0x0103

func lp(b []byte) []byte {
	out := make([]byte, 4+len(b))
	binary.LittleEndian.PutUint32(out, uint32(len(b)))
	copy(out[4:], b)
	return out
}

func signV2(apk []byte, key *rsa.PrivateKey, cert []byte) ([]byte, error) {
	eocd := len(apk) - 22
	if eocd < 0 || binary.LittleEndian.Uint32(apk[eocd:]) != 0x06054b50 {
		return nil, fmt.Errorf("the zip has an unexpected end")
	}
	cdOff := int(binary.LittleEndian.Uint32(apk[eocd+16:]))
	entries, cd, end := apk[:cdOff], apk[cdOff:eocd], apk[eocd:]

	digest := apkDigest(entries, cd, end)
	digests := lp(append(u32(sigRSAPKCS1SHA256), lp(digest)...))
	signed := append(append(lp(digests), lp(lp(cert))...), lp(nil)...)
	h := sha256.Sum256(signed)
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		return nil, err
	}
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	signer := append(append(lp(signed), lp(lp(append(u32(sigRSAPKCS1SHA256), lp(sig)...)))...), lp(pub)...)
	value := lp(lp(signer))

	var pair bytes.Buffer
	binary.Write(&pair, binary.LittleEndian, uint64(4+len(value)))
	binary.Write(&pair, binary.LittleEndian, uint32(0x7109871a))
	pair.Write(value)
	size := uint64(pair.Len() + 8 + 16)
	var block bytes.Buffer
	binary.Write(&block, binary.LittleEndian, size)
	block.Write(pair.Bytes())
	binary.Write(&block, binary.LittleEndian, size)
	block.WriteString("APK Sig Block 42")

	out := make([]byte, 0, len(apk)+block.Len())
	out = append(out, entries...)
	out = append(out, block.Bytes()...)
	out = append(out, cd...)
	newEnd := append([]byte(nil), end...)
	binary.LittleEndian.PutUint32(newEnd[16:], uint32(cdOff+block.Len()))
	out = append(out, newEnd...)
	return out, nil
}

func u32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}

func apkDigest(sections ...[]byte) []byte {
	var chunks [][]byte
	for _, s := range sections {
		for len(s) > 0 {
			n := min(len(s), 1<<20)
			h := sha256.New()
			h.Write([]byte{0xa5})
			h.Write(u32(uint32(n)))
			h.Write(s[:n])
			chunks = append(chunks, h.Sum(nil))
			s = s[n:]
		}
	}
	h := sha256.New()
	h.Write([]byte{0x5a})
	h.Write(u32(uint32(len(chunks))))
	for _, c := range chunks {
		h.Write(c)
	}
	return h.Sum(nil)
}

func androidDebugKey() (*rsa.PrivateKey, []byte, error) {
	dir := os.Getenv("MIHAN_ANDROID_KEY_DIR")
	if dir == "" {
		cfg, err := os.UserConfigDir()
		if err != nil {
			cfg = os.TempDir()
		}
		dir = filepath.Join(cfg, "mihan")
	}
	path := filepath.Join(dir, "android-debug.pem")
	if b, err := os.ReadFile(path); err == nil {
		var key *rsa.PrivateKey
		var cert []byte
		for {
			var blk *pem.Block
			blk, b = pem.Decode(b)
			if blk == nil {
				break
			}
			switch blk.Type {
			case "RSA PRIVATE KEY":
				key, _ = x509.ParsePKCS1PrivateKey(blk.Bytes)
			case "CERTIFICATE":
				cert = blk.Bytes
			}
		}
		if key != nil && cert != nil {
			return key, cert, nil
		}
	}
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "Mihan Debug", Organization: []string{"Mihan"}},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(30, 0, 0),
	}
	cert, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	var b bytes.Buffer
	pem.Encode(&b, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: cert})
	if err := os.MkdirAll(dir, 0o700); err == nil {
		_ = os.WriteFile(path, b.Bytes(), 0o600)
	}
	return key, cert, nil
}
