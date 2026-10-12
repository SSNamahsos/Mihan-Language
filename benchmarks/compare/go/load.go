//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := os.Args[1]
	var n atomic.Int64
	stop := time.Now().Add(3 * time.Second)
	var wg sync.WaitGroup
	tr := &http.Transport{MaxIdleConnsPerHost: 16}
	c := &http.Client{Transport: tr}
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(stop) {
				r, err := c.Get(url)
				if err != nil {
					continue
				}
				io.Copy(io.Discard, r.Body)
				r.Body.Close()
				if r.StatusCode == 200 {
					n.Add(1)
				}
			}
		}()
	}
	wg.Wait()
	fmt.Printf("%.0f\n", float64(n.Load())/3)
}
