package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

// const url = "https://3.39.193.212"
const url = "https://127.0.0.1"

var (
	loop         = flag.Bool("loop", false, "request loop")
	httpVersion  = flag.Int("version", 2, "HTTP version")
	num          = flag.Int("num", 1, "requests")
	totalLatency = time.Duration(0)
)

type Segment struct {
	Number string `json:"segment"`
	Data   []byte `json:"data"`
}

type Segments struct {
	Requests []Segment
}

func checkErr(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Printf("ERROR: %s: %s\n", msg, err)
	os.Exit(1)
}

func main() {
	flag.Parse()

	var wait sync.WaitGroup
	wait.Add(*num)

	go func() {
		for i := 0; i < *num; i++ {
			defer wait.Done()
			HttpClientExample(":6121")
		}
	}()

	wait.Wait()

	fmt.Printf("latency: %dms\n", int64((totalLatency/time.Duration(*num))/time.Millisecond))
}

func HttpClientExample(port string) {
	client := &http.Client{}

	// Create a pool with the server certificate since it is not signed
	// by a known CA
	caCert, err := ioutil.ReadFile("../http3/testdata/cert.pem")
	if err != nil {
		log.Fatalf("Reading server certificate: %s", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS configuration with the certificate of the server
	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
		ServerName:         "127.0.0.1",
	}

	switch *httpVersion {
	case 1:
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	case 2:
		client.Transport = &http2.Transport{
			TLSClientConfig: tlsConfig,
			// AllowHTTP:       true,
			// DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			// 	return net.Dial(network, addr)
			// },
		}
	}

	if *loop {
		for {
			start := time.Now()
			resp, err := client.Get(url + port)
			checkErr(err, "during get")

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			fmt.Printf("Client Proto: %d\n", resp.ProtoMajor)
			if err != nil {
				log.Fatalf("Failed reading response body: %s", err)
			}
			if body == nil {
				fmt.Println("null")
			}

			elapsed := time.Since(start)
			totalLatency += elapsed
		}
	} else {
		start := time.Now()
		resp, err := client.Get(url + port)
		checkErr(err, "during get")

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Printf("Client Proto: %d\n", resp.ProtoMajor)
		if err != nil {
			log.Fatalf("Failed reading response body: %s", err)
		}
		if body == nil {
			fmt.Println("null")
		}

		elapsed := time.Since(start)
		totalLatency += elapsed
	}
}
