//go:build ignore

package main

import (
	"fmt"
	"sort"
	"strconv"
)

func main() {
	d := map[string]int{}
	for i := 0; i < 300000; i++ {
		k := "k" + strconv.Itoa(i%1000)
		d[k] += i
	}
	xs := []int{}
	for i := 0; i < 300000; i++ {
		xs = append(xs, (i*7919)%100003)
	}
	sort.Ints(xs)
	fmt.Println(len(d), xs[1000])
}
