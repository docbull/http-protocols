package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
}

func main() {
	H2CServerUpgrade()
}

func ConnHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Hello, %v, TLS: %v %v\n", r.URL.Path, r.TLS == nil, r.Proto)

	seg, err := ioutil.ReadFile("../http3/seg/master0.ts")
	if err != nil {
		log.Fatalln(err)
	}
	// data0 := Segment{
	// 	Number: "seg0",
	// 	Data:   seg0,
	// }
	// request := Segments{}
	// request.Requests = append(request.Requests, data0)
	// mRequest, _ := json.Marshal(request)

	// w.Write(mRequest)

	w.Write(seg)

	// w.Write([]byte{
	// 	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	// 	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x28,
	// 	0x01, 0x03, 0x00, 0x00, 0x00, 0xb6, 0x30, 0x2a, 0x2e, 0x00, 0x00, 0x00,
	// 	0x03, 0x50, 0x4c, 0x54, 0x45, 0x5a, 0xc3, 0x5a, 0xad, 0x38, 0xaa, 0xdb,
	// 	0x00, 0x00, 0x00, 0x0b, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0x63, 0x18,
	// 	0x61, 0x00, 0x00, 0x00, 0xf0, 0x00, 0x01, 0xe2, 0xb8, 0x75, 0x22, 0x00,
	// 	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
	// })
}

// This server supports "H2C upgrade" and "H2C prior knowledge" along with
// standard HTTP/2 and HTTP/1.1 that golang natively supports.
func H2CServerUpgrade() {
	h2s := &http2.Server{}

	handler := http.HandlerFunc(ConnHandler)

	server := &http.Server{
		Addr:    "0.0.0.0:6121",
		Handler: h2c.NewHandler(handler, h2s),
	}

	fmt.Printf("Listening [0.0.0.0:6121]...\n")
	checkErr(server.ListenAndServeTLS("../http3/testdata/cert.pem", "../http3/testdata/priv.key"), "while listening")

	// h2s := &http2.Server{}

	// handler := http.HandlerFunc(ConnHandler)

	// serverTLSCert, err := tls.LoadX509KeyPair("../http3/testdata/cert.pem", "../http3/testdata/priv.key")
	// if err != nil {
	// 	log.Fatalf("Error loading certificate and key file: %v", err)
	// }
	// certPool := x509.NewCertPool()
	// if caCertPEM, err := ioutil.ReadFile("../http3/testdata/ca.pem"); err != nil {
	// 	panic(err)
	// } else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
	// 	panic("invalid cert in CA PEM")
	// }

	// tlsConfig := &tls.Config{
	// 	ClientAuth:   tls.RequireAndVerifyClientCert,
	// 	ClientCAs:    certPool,
	// 	Certificates: []tls.Certificate{serverTLSCert},
	// }
	// server := http.Server{
	// 	Addr:      "0.0.0.0:6121",
	// 	Handler:   h2c.NewHandler(handler, h2s),
	// 	TLSConfig: tlsConfig,
	// }
	// defer server.Close()
	// log.Fatal(server.ListenAndServeTLS("", ""))
}

// This server only supports "H2C prior knowledge".
// You can add standard HTTP/2 support by adding a TLS config.
func H2CServerPrior() {
	server := http2.Server{}

	l, err := net.Listen("tcp", "0.0.0.0:6121")
	checkErr(err, "while listening")

	fmt.Printf("Listening [0.0.0.0:6121]...\n")
	for {
		conn, err := l.Accept()
		checkErr(err, "during accept")

		server.ServeConn(conn, &http2.ServeConnOpts{
			Handler: http.HandlerFunc(ConnHandler),
		})
	}
}
