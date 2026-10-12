//go:build ignore

package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(map[string]any{"salam": r.URL.Path, "n": 1})
		w.Write(b)
	})
	http.ListenAndServe("127.0.0.1:8813", nil)
}
