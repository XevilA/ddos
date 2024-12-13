package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

func main() {
	url := flag.String("url", "http://bodin2.ac.th", "Target URL")
	requests := flag.Int("requests", 1000, "Total number of requests")
	concurrency := flag.Int("concurrency", 100, "Number of concurrent requests")
	payloadSize := flag.Int("size", 1024, "Payload size in bytes")
	timeout := flag.Int("timeout", 1000, "Request timeout in seconds")
	flag.Parse()

	payload := bytes.Repeat([]byte("a"), *payloadSize)

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	responseTimes := make(chan time.Duration, *requests)
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, *concurrency)
	for i := 0; i < *requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			semaphore <- struct{}{}
			start := time.Now()
			resp, err := client.Post(*url, "application/json", bytes.NewReader(payload))
			elapsed := time.Since(start)
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			responseTimes <- elapsed
			<-semaphore
		}()
	}

	wg.Wait()
	close(responseTimes)

	var fastest time.Duration = time.Hour
	for rt := range responseTimes {
		if rt < fastest {
			fastest = rt
		}
	}

	fmt.Printf("Fastest response time: %v\n", fastest)
}
