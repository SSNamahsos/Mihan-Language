package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"debug/elf"
	"debug/pe"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"mihan/internal/diag"
	"mihan/internal/parser"
)

const appMagic = "MIHANAPP"
const trailerLen = 8 + len(appMagic)
const mainEntry = ".mihan-main"

const openEntry = ".mihan-open-browser"

func cmdBuild(args []string) error {
	var path, out, base string
	goos, goarch := runtime.GOOS, runtime.GOARCH
	window := false
	openBrowser := false
	archGiven := false
	pkg, label := "", ""
	for i := 0; i < len(args); i++ {
		switch a := args[i]; {
		case (a == "-o" || a == "--out") && i+1 < len(args):
			i++
			out = args[i]
		case a == "--base" && i+1 < len(args):
			i++
			base = args[i]
		case a == "--os" && i+1 < len(args):
			i++
			goos = args[i]
		case a == "--arch" && i+1 < len(args):
			i++
			goarch = args[i]
			archGiven = true
		case a == "--package" && i+1 < len(args):
			i++
			pkg = args[i]
		case a == "--name" && i+1 < len(args):
			i++
			label = args[i]
		case a == "--window" || a == "--gui":
			window = true
		case a == "--open-browser" || a == "--open":
			openBrowser = true
		case strings.HasPrefix(a, "-"):
			return fatal("I do not know the option %s\n  usage: %s", a, buildUsage)
		default:
			path = a
		}
	}
	if path == "" {
		return fatal("give me the program to build, like `mihan build app.mhn`\n  usage: %s", buildUsage)
	}
	if goos != "linux" && goos != "windows" && goos != "android" {
		return fatal("--os must be linux, windows or android, not %s", goos)
	}
	if goos == "android" && !archGiven {

		goarch = "arm64"
	}
	if goarch != "amd64" && goarch != "arm64" && !(goos == "android" && goarch == "all") {
		return fatal("--arch must be amd64 or arm64, not %s", goarch)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fatal("%s is not here", path)
	}
	if info.IsDir() {
		path = filepath.Join(path, "main.mhn")
		if _, err := os.Stat(path); err != nil {
			return fatal("%s has no main.mhn; name the file to build", filepath.Dir(path))
		}
	}

	if err := checkOnly(path); err != nil {
		return err
	}
	if goos == "android" {
		return buildAndroid(path, out, goarch, base, pkg, label)
	}
	if base == "" {
		if base, err = findRuntime(goos, goarch); err != nil {
			return err
		}
	}
	exe, err := os.ReadFile(base)
	if err != nil {
		return fatal("%s cannot be read: %v", base, err)
	}
	exe = stripPayload(exe)
	if gotOS, gotArch := binaryTarget(exe); gotOS != goos || gotArch != goarch {
		return fatal("%s is a runtime for %s/%s, not %s/%s", base, gotOS, gotArch, goos, goarch)
	}
	if goos == "windows" {

		sub := uint16(3)
		if window {
			sub = 2
		}
		if exe, err = setPESubsystem(exe, sub); err != nil {
			return fatal("%s: %v", base, err)
		}
	}
	if out == "" {
		name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if name == "main" {
			if abs, err := filepath.Abs(filepath.Dir(path)); err == nil {
				name = filepath.Base(abs)
			}
		}
		if goos == "windows" {
			name += ".exe"
		}
		out = name
	}
	payload, files, err := packFolder(filepath.Dir(path), filepath.Base(path), out, openBrowser)
	if err != nil {
		return err
	}
	built := withPayload(exe, payload)
	if err := os.WriteFile(out, built, 0o755); err != nil {
		return fatal("cannot write %s: %v", out, err)
	}
	kind := goos + "/" + goarch
	if openBrowser {
		kind += ", opens the browser"
	}
	if goos == "windows" && window {
		kind += ", window app (no console)"
	} else if goos == "windows" {
		kind += ", console"
	}
	fmt.Printf("built %s (%s, %d files, %.1f MB)\n", out, kind, files, float64(len(built))/1e6)
	return nil
}

func withPayload(exe, payload []byte) []byte {
	var buf bytes.Buffer
	buf.Write(exe)
	buf.Write(payload)
	var n [8]byte
	binary.LittleEndian.PutUint64(n[:], uint64(len(payload)))
	buf.Write(n[:])
	buf.WriteString(appMagic)
	return buf.Bytes()
}

const buildUsage = "mihan build app.mhn [-o name] [--os linux|windows|android] [--arch amd64|arm64|all] [--window] [--open-browser] [--base runtime] [--package id] [--name label]"

func findRuntime(goos, goarch string) (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fatal("cannot find my own executable: %v", err)
	}
	if goos == runtime.GOOS && goarch == runtime.GOARCH {
		return self, nil
	}
	name := "mihan-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	var dirs []string
	if d := os.Getenv("MIHAN_RUNTIMES"); d != "" {
		dirs = append(dirs, d)
	}
	if real, err := filepath.EvalSymlinks(self); err == nil {
		self = real
	}
	dir := filepath.Dir(self)
	dirs = append(dirs, filepath.Join(dir, "runtimes"), filepath.Join(dir, "..", "runtimes"), dir)
	for _, d := range dirs {
		p := filepath.Join(d, name)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fatal("I have no runtime for %s/%s: put %s in a `runtimes` folder beside mihan (it comes with Mihan), set MIHAN_RUNTIMES, or give --base", goos, goarch, name)
}

func binaryTarget(b []byte) (string, string) {
	r := bytes.NewReader(b)
	if f, err := pe.NewFile(r); err == nil {
		switch f.Machine {
		case pe.IMAGE_FILE_MACHINE_AMD64:
			return "windows", "amd64"
		case pe.IMAGE_FILE_MACHINE_ARM64:
			return "windows", "arm64"
		}
		return "windows", "unknown"
	}
	if f, err := elf.NewFile(r); err == nil {
		arch := "unknown"
		switch f.Machine {
		case elf.EM_X86_64:
			arch = "amd64"
		case elf.EM_AARCH64:
			arch = "arm64"
		}
		return "linux", arch
	}
	return "unknown", "unknown"
}

func setPESubsystem(exe []byte, sub uint16) ([]byte, error) {
	if len(exe) < 0x40 {
		return nil, fmt.Errorf("not a Windows executable")
	}
	off := int(binary.LittleEndian.Uint32(exe[0x3c:]))
	at := off + 4 + 20 + 68
	if off <= 0 || at+2 > len(exe) || string(exe[off:off+4]) != "PE\x00\x00" {
		return nil, fmt.Errorf("not a Windows executable")
	}
	out := append([]byte(nil), exe...)
	binary.LittleEndian.PutUint16(out[at:], sub)
	return out, nil
}

func isPE(b []byte) bool { return len(b) > 2 && b[0] == 'M' && b[1] == 'Z' }

func stripPayload(exe []byte) []byte {
	if n, ok := payloadLen(exe); ok {
		return exe[:len(exe)-trailerLen-n]
	}
	return exe
}

func payloadLen(b []byte) (int, bool) {
	if len(b) < trailerLen || string(b[len(b)-len(appMagic):]) != appMagic {
		return 0, false
	}
	n := binary.LittleEndian.Uint64(b[len(b)-trailerLen:])
	if n == 0 || n > uint64(len(b)-trailerLen) {
		return 0, false
	}
	return int(n), true
}

func skipDir(name string) bool {
	return name == "dade" || name == "node_modules" || strings.HasPrefix(name, ".")
}

func packFolder(dir, mainFile, out string, openBrowser bool) ([]byte, int, error) {
	outAbs, _ := filepath.Abs(out)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	files := 0
	var total int64
	err := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		if info.IsDir() {
			if rel != "." && skipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if abs, _ := filepath.Abs(p); abs == outAbs || strings.HasSuffix(strings.ToLower(p), ".exe") || strings.HasSuffix(strings.ToLower(p), ".apk") || !info.Mode().IsRegular() {
			return nil
		}
		total += info.Size()
		if total > 200<<20 {
			return fmt.Errorf("the folder %s holds more than 200 MB; build from a folder with only the program in it", dir)
		}
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		files++
		return err
	})
	if err != nil {
		return nil, 0, fatal("cannot pack the program: %v", err)
	}
	w, _ := zw.Create(mainEntry)
	_, _ = w.Write([]byte(mainFile))
	if openBrowser {
		_, _ = zw.Create(openEntry)
	}
	if err := zw.Close(); err != nil {
		return nil, 0, err
	}
	return buf.Bytes(), files, nil
}

func embeddedProgram() (string, bool, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", false, nil
	}
	f, err := os.Open(exe)
	if err != nil {
		return "", false, nil
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || st.Size() < int64(trailerLen) {
		return "", false, nil
	}
	tr := make([]byte, trailerLen)
	if _, err := f.ReadAt(tr, st.Size()-int64(trailerLen)); err != nil || string(tr[8:]) != appMagic {
		return "", false, nil
	}
	n := int64(binary.LittleEndian.Uint64(tr))
	if n <= 0 || n > st.Size()-int64(trailerLen) {
		return "", false, nil
	}
	payload := make([]byte, n)
	if _, err := f.ReadAt(payload, st.Size()-int64(trailerLen)-n); err != nil {
		return "", true, err
	}
	sum := sha256.Sum256(payload)
	stamp := hex.EncodeToString(sum[:8])

	root, err := os.UserCacheDir()
	if err != nil {
		root = os.TempDir()
	}
	name := strings.TrimSuffix(filepath.Base(exe), filepath.Ext(exe))
	dir := filepath.Join(root, "mihan-apps", name)
	zr, err := zip.NewReader(bytes.NewReader(payload), n)
	if err != nil {
		return "", true, err
	}
	mainFile := ""
	for _, zf := range zr.File {
		if zf.Name == openEntry && os.Getenv("MIHAN_OPEN_BROWSER") == "" {
			os.Setenv("MIHAN_OPEN_BROWSER", "1")
		}
		if zf.Name == mainEntry {
			rc, err := zf.Open()
			if err != nil {
				return "", true, err
			}
			b, _ := io.ReadAll(rc)
			rc.Close()
			mainFile = string(b)
		}
	}
	if mainFile == "" {
		return "", true, fmt.Errorf("the program inside this executable has no main file")
	}
	marker := filepath.Join(dir, ".mihan-build")
	if old, err := os.ReadFile(marker); err != nil || string(old) != stamp {
		for _, zf := range zr.File {
			if zf.Name == mainEntry || zf.Name == openEntry {
				continue
			}
			target := filepath.Join(dir, filepath.FromSlash(zf.Name))
			if rel, err := filepath.Rel(dir, target); err != nil || strings.HasPrefix(rel, "..") {
				continue
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return "", true, err
			}
			rc, err := zf.Open()
			if err != nil {
				return "", true, err
			}
			b, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				return "", true, err
			}
			if err := os.WriteFile(target, b, 0o644); err != nil {
				return "", true, err
			}
		}
		if err := os.WriteFile(marker, []byte(stamp), 0o644); err != nil {
			return "", true, err
		}
	}
	return filepath.Join(dir, filepath.FromSlash(mainFile)), true, nil
}

func checkOnly(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return fatal("%s cannot be read: %v", path, err)
	}
	bag := &diag.Bag{}
	parser.Parse(path, src, bag)
	if bag.HasErrors() {
		printDiagnostics(bag.Sorted(), map[string][]byte{path: src})
		return errSilent
	}
	return nil
}
