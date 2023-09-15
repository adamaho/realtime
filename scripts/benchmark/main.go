package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const NUM_CLIENTS = 2

func main() {
	var wg sync.WaitGroup

	for i := 0; i < NUM_CLIENTS; i++ {
		wg.Add(1)

		i := i
		go func() {
			defer wg.Done()

			tls := &tls.Config{
				InsecureSkipVerify: true,
			}

			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig:   tls,
					IdleConnTimeout:   60 * time.Second,
					ForceAttemptHTTP2: true,
				},
			}

			req, err := http.NewRequest("GET", "https://localhost:3000/count", nil)
			if err != nil {
				panic("Failed to get request for client")
			}

			req.Header.Set("x-stream", "true")

			res, err := client.Do(req)
			if err != nil {
				fmt.Println("Oops, something went wrong with the request", err)
			}

			defer res.Body.Close()

			reader := bufio.NewReader(res.Body)

			for {
				str, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				fmt.Print(i, str)
			}
		}()
	}

	wg.Wait()
}
