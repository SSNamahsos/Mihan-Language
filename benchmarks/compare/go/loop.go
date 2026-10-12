//go:build ignore

package main

import "fmt"

func main() {
	s := 0
	for i := 0; i < 5000000; i++ {
		s += i
	}
	fmt.Println(s)
}
