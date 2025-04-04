package main

import (
	"go-tools/loadtester"
	"log"
	"time"
)

type (
	Input struct {
		Name string
	}

	Output struct {
		Result int
	}
)

const (
	url         = "http://localhost:8080"
	numRequests = 1000
)

func main() {
	lt := loadtester.NewLoadTester[Input, Output](
		url, numRequests,
		func(i int) map[string]string {
			return map[string]string{
				"Content-Type": "application/json",
			}
		},
		func(i int) Input {
			return Input{
				Name: "Justin",
			}
		},
	)

	go func() {
		time.Sleep(2 * time.Second)
		lt.Stop()
	}()

	results := lt.Start(time.Second)
	var successes, failures int
	for res := range results {
		if res.Status != 200 || res.Err != nil {
			log.Println(res.Status)
			failures++
		} else {
			log.Println("W", res.Status)
			successes++
		}
	}

	log.Printf("done. successes: %d, failures: %d", successes, failures)
}
