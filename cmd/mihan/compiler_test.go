package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

var (
	compilerOnce sync.Once
	compilerExe  string
	compilerTmp  string
	compilerErr  error
)

func compilerForTest(t *testing.T) string {
	t.Helper()
	compilerOnce.Do(func() {
		compilerTmp, compilerErr = os.MkdirTemp("", "mihan-test-")
		if compilerErr != nil {
			return
		}

		name := "mihan.exe"
		if os.PathSeparator == '/' {
			name = "mihan"
		}
		compilerExe = filepath.Join(compilerTmp, name)
		cmd := exec.Command("go", "build", "-o", compilerExe, ".")
		cmd.Stderr = os.Stderr
		compilerErr = cmd.Run()
	})
	if compilerErr != nil {
		t.Fatalf("cannot build the compiler: %v", compilerErr)
	}
	return compilerExe
}

func TestMain(m *testing.M) {
	code := m.Run()
	if compilerTmp != "" {
		os.RemoveAll(compilerTmp)
	}
	os.Exit(code)
}
