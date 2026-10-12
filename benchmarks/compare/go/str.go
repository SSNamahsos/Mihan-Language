//go:build ignore

package main

import (
	"fmt"
	"strconv"
	"strings"
)

func main() {
	parts := []string{}
	for i := 0; i < 300000; i++ {
		parts = append(parts, "x"+strconv.Itoa(i))
	}
	s := strings.Join(parts, ",")
	var t strings.Builder
	for i := 0; i < 100000; i++ {
		t.WriteString("ab")
	}
	fmt.Println(len(s), t.Len())
}
