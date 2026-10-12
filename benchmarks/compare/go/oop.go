//go:build ignore

package main

import "fmt"

type Hesab struct{ mojudi int }

func (h *Hesab) variz(m int) { h.mojudi += m }

func main() {
	h := &Hesab{}
	for i := 0; i < 1000000; i++ {
		h.variz(i % 7)
	}
	dobar := func(x int) int { return x * 2 }
	s := 0
	for i := 0; i < 300000; i++ {
		s += dobar(i)
	}
	fmt.Println(h.mojudi, s)
}
