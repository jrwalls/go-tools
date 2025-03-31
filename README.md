## Scaling Worker Pool

A simple generic worker pool that supports dynamic scaling of worker count at runtime.

### Features
- Generic type support
- Dynamically scale workers with `SetWorkerCount(n)`
- Graceful shutdown via `StopAllWorkers()`

### Usage
```go
package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	workerFunc := func(job string) {
		time.Sleep(1 * time.Second)
		log.Println(job)
	}

	queueSize := 100
	pool := scaling_worker_pool.NewWorkerPool[string](workerFunc, queueSize)
	pool.SetWorkerCount(1)

	go func() {
		for i := 0; i < 10; i++ {
			pool.Send("wee")
		}
	}()

	go func() {
		sc := bufio.NewScanner(os.Stdin)
		for sc.Scan() {
			if n, err := strconv.Atoi(strings.TrimSpace(sc.Text())); err == nil {
				pool.SetWorkerCount(n)
				log.Println("scaling to", n)
			}
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	pool.StopAllWorkers()
}
```