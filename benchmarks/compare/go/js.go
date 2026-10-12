//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func main() {
	rows := []map[string]any{}
	for i := 0; i < 50000; i++ {
		rows = append(rows, map[string]any{"id": i, "nam": "kar" + strconv.Itoa(i), "dorost": i%2 == 0, "vazn": float64(i) * 0.5})
	}
	b, _ := json.Marshal(rows)
	var back []any
	_ = json.Unmarshal(b, &back)
	fmt.Println(len(b), len(back))
}
