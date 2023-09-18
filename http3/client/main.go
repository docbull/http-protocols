package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

var certPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Dir(filename)
}

// AddRootCA adds the root CA certificate to a cert pool
func AddRootCA(certPool *x509.CertPool) {
	caCertPath := path.Join(certPath, "../testdata/ca.pem")
	caCertRaw, err := os.ReadFile(caCertPath)
	if err != nil {
		panic(err)
	}
	if ok := certPool.AppendCertsFromPEM(caCertRaw); !ok {
		panic("Could not add root ceritificate to pool.")
	}
}

func main() {
	loop := flag.Bool("loop", false, "request loop")
	quiet := flag.Bool("q", false, "don't print the data")
	keyLogFile := flag.String("keylog", "", "key log file")
	insecure := flag.Bool("insecure", false, "skip certificate verification")
	flag.Parse()
	urls := flag.Args()

	var keyLog io.Writer
	if len(*keyLogFile) > 0 {
		f, err := os.Create(*keyLogFile)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		keyLog = f
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatal(err)
	}
	AddRootCA(pool)

	var qconf quic.Config
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs:            pool,
			InsecureSkipVerify: *insecure,
			KeyLogWriter:       keyLog,
		},
		QuicConfig: &qconf,
	}

	defer roundTripper.Close()
	hclient := &http.Client{
		Transport: roundTripper,
	}

	var wg sync.WaitGroup
	wg.Add(len(urls))

	if *loop {
		fmt.Printf("GET %s\n", urls[0])

		go func(addr string) {
			for {
				start := time.Now()
				rsp, err := hclient.Get(addr)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Got response for %s: %#v\n", addr, rsp)

				body := &bytes.Buffer{}
				_, err = io.Copy(body, rsp.Body)

				fmt.Println("Elaped:", time.Since(start))

				if err != nil {
					log.Fatal(err)
				}
				if *quiet {
					fmt.Printf("Response Body: %d bytes\n", body.Len())
				} else {
					fmt.Printf("Response Body:\n")
					fmt.Printf("%s\n", body.Bytes())
				}
			}
			wg.Done()
		}(urls[0])
	} else {
		for _, addr := range urls {
			fmt.Printf("GET %s\n", addr)

			go func(addr string) {
				start := time.Now()
				rsp, err := hclient.Get(addr)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Got response for %s: %#v\n", addr, rsp)

				body := &bytes.Buffer{}
				_, err = io.Copy(body, rsp.Body)

				fmt.Println("Elaped:", time.Since(start))

				if err != nil {
					log.Fatal(err)
				}
				if *quiet {
					fmt.Printf("Response Body: %d bytes\n", body.Len())
				} else {
					fmt.Printf("Response Body:\n")
					fmt.Printf("%s\n", body.Bytes())
				}
				wg.Done()
			}(addr)
		}
	}
	wg.Wait()
}
