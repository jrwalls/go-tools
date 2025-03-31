package scaling_worker_pool

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
	pool := NewWorkerPool[string](workerFunc, queueSize)
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
				continue
			}
		}
		if err := sc.Err(); err != nil {
			log.Printf("read stdin: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	pool.StopAllWorkers()
}
