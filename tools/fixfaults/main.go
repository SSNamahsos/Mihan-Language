package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var strayEnglish = regexp.MustCompile(`fmt\.Sprintf\(("(?:[^"\\]|\\.)*"), ("(?:[^"\\]|\\.)*"), `)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: fixfaults <dir>")
		os.Exit(2)
	}
	dir := os.Args[1]
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	changed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		src := string(raw)
		out := strayEnglish.ReplaceAllString(src, `fmt.Sprintf($1, `)
		if out == src {
			continue
		}
		changed++
		if err := os.WriteFile(path, []byte(out), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("rewrote", path)
	}
	fmt.Println("files changed:", changed)
}
