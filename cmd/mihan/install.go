package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"mihan/internal/version"
)

func cmdInstall(args []string) error {
	dir, err := installDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fatal("cannot make %s: %v", dir, err)
	}
	built, err := buildBinaries(dir)
	if err != nil {
		return err
	}
	added, err := addToPath(dir)
	if err != nil {

		fmt.Fprintf(os.Stderr, "\nThe programs are in %s, but the path could not be changed:\n  %v\n", dir, err)
		fmt.Fprintf(os.Stderr, "Add it by hand and open a new terminal:\n  %s\n\n", pathAdvice(dir))
		return nil
	}
	fmt.Printf("Mihan %s is installed.\n\n", version.Short())
	for _, b := range built {
		fmt.Printf("  %-14s %s\n", b.name, b.path)
	}
	if added {
		fmt.Printf("\n%s is now on your path. Open a new terminal and try:\n  mihan version\n", dir)
	} else {
		fmt.Printf("\n%s was already on your path.\n", dir)
	}
	fmt.Printf("\nFor an editor, install the Mihan extension. It starts mihan-ls by\nname, so it finds what was just installed.\n")
	return nil
}

type builtBinary struct{ name, path string }

func buildBinaries(dir string) ([]builtBinary, error) {
	if _, err := os.Stat("go.mod"); err != nil {
		return nil, fatal("run this from the Mihan source folder, the one with go.mod in it")
	}
	var out []builtBinary
	for _, b := range []struct{ name, pkg string }{
		{"mihan.exe", "./cmd/mihan"},
		{"mihan-ls.exe", "./cmd/mihan-ls"},
	} {
		path := filepath.Join(dir, b.name)
		cmd := exec.Command("go", "build", "-trimpath", "-o", path, b.pkg)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fatal("cannot build %s: %v", b.name, err)
		}
		out = append(out, builtBinary{name: b.name, path: path})
	}
	return out, nil
}

func installDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fatal("cannot find your home folder: %v", err)
	}
	base := ".mihan/bin"
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {

			return filepath.Join(local, "Mihan", "bin"), nil
		}
	} else if data := os.Getenv("XDG_DATA_HOME"); data != "" {
		return filepath.Join(data, "mihan", "bin"), nil
	}
	return filepath.Join(home, base), nil
}

func addToPath(dir string) (bool, error) {
	key := "PATH"
	if runtime.GOOS == "windows" {
		key = "Path"
	}
	switch runtime.GOOS {
	case "windows":
		return addWindowsPath(dir, key)
	case "darwin":
		return addApplePath(dir, key)
	default:
		return false, fmt.Errorf("automatic path setting is only done on Windows and macOS; add %s yourself", dir)
	}
}

func addWindowsPath(dir, key string) (bool, error) {
	cmd := exec.Command("reg", "query",
		"HKCU\\Environment", "/v", key)
	cmd.Stdout = nil
	cmd.Stderr = nil
	existing, _ := cmd.Output()
	have := windowsPathValue(string(existing))
	if hasPathEntry(have, dir) {
		return false, nil
	}

	ps := `$p = [Environment]::GetEnvironmentVariable('` + key + `', 'User')
if ([string]::IsNullOrEmpty($p)) { $p = '` + dir + `' } else { $p = '` + dir + `;' + $p }
[Environment]::SetEnvironmentVariable('` + key + `', $p, 'User')`
	set := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", ps)
	set.Stdout = os.Stderr
	set.Stderr = os.Stderr
	if err := set.Run(); err != nil {
		return false, err
	}
	return true, nil
}

func windowsPathValue(regDump string) string {
	for _, line := range strings.Split(regDump, "\n") {
		if !strings.Contains(line, "Path") {
			continue
		}
		parts := strings.SplitN(line, "REG_", 2)
		if len(parts) != 2 {
			continue
		}
		if at := strings.Index(parts[1], "\t"); at >= 0 {
			return parts[1][at+1:]
		}
	}
	return ""
}

func hasPathEntry(pathList, dir string) bool {
	want := strings.ToLower(strings.TrimRight(dir, `\/`))
	for _, entry := range strings.Split(pathList, ";") {
		if strings.ToLower(strings.TrimRight(strings.TrimSpace(entry), `\/`)) == want {
			return true
		}
	}
	return false
}

func addApplePath(dir, key string) (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	profile := filepath.Join(home, ".zprofile")
	line := fmt.Sprintf("export %s=%s\n", key, dir)
	old, _ := os.ReadFile(profile)
	if strings.Contains(string(old), dir) {
		return false, nil
	}
	out := string(old)
	if out != "" && !strings.HasSuffix(out, "\n") {
		out += "\n"
	}
	out += "# added by `mihan install`\n" + line
	if err := os.WriteFile(profile, []byte(out), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

func pathAdvice(dir string) string {
	if runtime.GOOS == "windows" {
		return "set the Path environment variable to include " + dir
	}
	return "add `export PATH=\"" + dir + ":$PATH\"` to your shell profile"
}

func cmdEnv(args []string) error {
	dir, err := installDir()
	if err != nil {
		return err
	}
	fmt.Printf("mihan       %s\n", version.Short())
	fmt.Printf("go          %s\n", goVersion())
	fmt.Printf("system      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Printf("installed   %s\n", dir)
	fmt.Printf("on the path %v\n", onPath(dir))
	fmt.Printf("editor      install the Mihan extension; it starts mihan-ls by name\n")
	fmt.Println()
	fmt.Println("To use a fresh terminal, run:")
	fmt.Println("  mihan install")
	return nil
}

func goVersion() string {
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err != nil {
		return "not found on the path, which is needed to build Mihan"
	}
	return strings.TrimSpace(string(out))
}

func onPath(dir string) bool {
	for _, entry := range filepath.SplitList(os.Getenv("PATH")) {
		if strings.EqualFold(strings.TrimRight(entry, `\/`), strings.TrimRight(dir, `\/`)) {
			return true
		}
	}
	return false
}
