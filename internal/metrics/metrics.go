package metrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

func startStatsPrinter(count *atomic.Int64) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var lastCount int64 = 0
	for range ticker.C {
		c := count.Load()
		fmt.Println("posts seen:", c)
		fmt.Println("post rate: ", c-lastCount, " posts/s")
		lastCount = c
	}
}
