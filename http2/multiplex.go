package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
)

// transport is an http.RoundTripper that keeps track of the in-flight
// request and implements hooks to report HTTP tracing events.
type transport struct {
	current *http.Request
}

// RoundTrip wraps http.DefaultTransport.RoundTrip to keep track
// of the current request.
func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.current = req
	return http.DefaultTransport.RoundTrip(req)
}

// GotConn prints whether the connection has been used previously
// for the current request.
func (t *transport) GotConn(info httptrace.GotConnInfo) {
	fmt.Printf("Connection reused? %v\n", info.Reused)
}

func test(reqNum int, url string, tr transport, client *http.Client) {
	fmt.Println("Beginning request #", reqNum)

	// Fetch the URL.
	req, _ := http.NewRequest("GET", url, nil)
	trace := &httptrace.ClientTrace{
		GotConn: tr.GotConn,
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	if resp, err := client.Do(req); err != nil {
		log.Fatal(err)
	} else {
		// fmt.Printf("Got response for #%v using %v\n", reqNum, resp.Proto)
		fmt.Printf("Got response for using %v\n", resp.Proto)
		resp.Body.Close()
	}
}

func main() {
	tr := &transport{}
	client := &http.Client{Transport: tr}
	// wg := sync.WaitGroup{}
	url := "https://www.google.com/"

	num := 3
	var wait sync.WaitGroup
	wait.Add(num)

	go func(num int, url string, tr transport, client *http.Client) {
		defer wait.Done()
		test(num, url, tr, client)
	}(num, url, *tr, client)
	go func(num int, url string, tr transport, client *http.Client) {
		defer wait.Done()
		test(num, url, tr, client)
	}(num, url, *tr, client)
	go func(num int, url string, tr transport, client *http.Client) {
		defer wait.Done()
		test(num, url, tr, client)
	}(num, url, *tr, client)

	wait.Wait()

	// for i := 1; i <= 5; i++ {
	// 	// Increment the WaitGroup counter.
	// 	wg.Add(1)
	// 	// Launch a goroutine to fetch the URL.
	// 	go func(reqNum int) {
	// 		fmt.Println("Beginning request #", reqNum)
	// 		// Decrement the counter when the goroutine completes.
	// 		defer wg.Done()
	// 		// Fetch the URL.
	// 		req, _ := http.NewRequest("GET", url, nil)
	// 		trace := &httptrace.ClientTrace{
	// 			GotConn: tr.GotConn,
	// 		}
	// 		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// 		if resp, err := client.Do(req); err != nil {
	// 			log.Fatal(err)
	// 		} else {
	// 			fmt.Printf("Got response for #%v using %v\n", reqNum, resp.Proto)
	// 			resp.Body.Close()
	// 		}
	// 	}(i)
	// }
	// wg.Wait()
}
