package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"

	_ "net/http/pprof"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

type binds []string

type Segment struct {
	Number string `json:"segment"`
	Data   []byte `json:"data"`
}

type Segments struct {
	Requests []Segment
}

func (b binds) String() string {
	return strings.Join(b, ",")
}

func (b *binds) Set(v string) error {
	*b = strings.Split(v, ",")
	return nil
}

// Size is needed by the /demo/upload handler to determine the size of the uploaded file
type Size interface {
	Size() int64
}

var certPath string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current frame")
	}

	certPath = path.Dir(filename)
}

// GetCertificatePaths returns the paths to certificate and key
func GetCertificatePaths() (string, string) {
	return path.Join(certPath, "testdata/cert.pem"), path.Join(certPath, "testdata/priv.key")
}

// See https://en.wikipedia.org/wiki/Lehmer_random_number_generator
func generatePRData(l int) []byte {
	res := make([]byte, l)
	seed := uint64(1)
	for i := 0; i < l; i++ {
		seed = seed * 48271 % 2147483647
		res[i] = byte(seed)
	}
	return res
}

func setupHandler(www string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hellow!", r.Proto)

		// Small 40x40 png
		w.Write([]byte{
			0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
			0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x28, 0x00, 0x00, 0x00, 0x28,
			0x01, 0x03, 0x00, 0x00, 0x00, 0xb6, 0x30, 0x2a, 0x2e, 0x00, 0x00, 0x00,
			0x03, 0x50, 0x4c, 0x54, 0x45, 0x5a, 0xc3, 0x5a, 0xad, 0x38, 0xaa, 0xdb,
			0x00, 0x00, 0x00, 0x0b, 0x49, 0x44, 0x41, 0x54, 0x78, 0x01, 0x63, 0x18,
			0x61, 0x00, 0x00, 0x00, 0xf0, 0x00, 0x01, 0xe2, 0xb8, 0x75, 0x22, 0x00,
			0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
		})
	})

	mux.HandleFunc("/segment", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path, r.Proto)

		seg, err := ioutil.ReadFile("seg/master0.ts")
		if err != nil {
			log.Fatalln(err)
		}

		// data := Segment{
		// 	Number: "seg",
		// 	Data:   seg,
		// }

		// request := Segments{}
		// request.Requests = append(request.Requests, data)

		// mRequest, _ := json.Marshal(request)

		w.Write(seg)
	})

	return mux
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	bs := binds{}
	flag.Var(&bs, "bind", "bind to")
	www := flag.String("www", "", "www data")
	tcp := flag.Bool("tcp", false, "also listen on TCP")
	flag.Parse()

	if len(bs) == 0 {
		bs = binds{"localhost:6121"}
	}

	handler := setupHandler(*www)
	quicConf := &quic.Config{}

	var wg sync.WaitGroup
	wg.Add(len(bs))
	for _, b := range bs {
		bCap := b
		go func() {
			var err error
			if *tcp {
				certFile, keyFile := GetCertificatePaths()
				err = http3.ListenAndServe(bCap, certFile, keyFile, handler)
			} else {
				server := http3.Server{
					Handler:    handler,
					Addr:       bCap,
					QuicConfig: quicConf,
				}
				err = server.ListenAndServeTLS(GetCertificatePaths())
			}
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
