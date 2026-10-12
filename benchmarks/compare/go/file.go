//go:build ignore

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	lines := []string{}
	for i := 0; i < 200000; i++ {
		lines = append(lines, "khat "+strconv.Itoa(i))
	}
	_ = os.WriteFile("/tmp/mihan-bench.txt", []byte(strings.Join(lines, "\n")), 0o644)
	b, _ := os.ReadFile("/tmp/mihan-bench.txt")
	fmt.Println(len(strings.Split(string(b), "\n")))
}
