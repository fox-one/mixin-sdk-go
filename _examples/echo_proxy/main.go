package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/oxtoacart/bpool"
	"github.com/rs/cors"
)

var cfg struct {
	keystore string
	port     int
	getOnly  bool
	endpoint string
}

var (
	xAuthorization = http.CanonicalHeaderKey("Authorization")
)

func main() {
	flag.StringVar(&cfg.keystore, "key", "", "keystore file path")
	flag.IntVar(&cfg.port, "port", 9999, "server port")
	flag.BoolVar(&cfg.getOnly, "get", false, "only allow GET method")
	flag.StringVar(&cfg.endpoint, "endpoint", mixin.DefaultApiHost, "mixin api host")
	flag.Parse()

	f, err := os.Open(cfg.keystore)
	if err != nil {
		log.Panicln(err)
	}

	var keystore mixin.Keystore
	if err := json.NewDecoder(f).Decode(&keystore); err != nil {
		log.Panicln(err)
	}

	_ = f.Close()

	auth, err := mixin.AuthFromKeystore(&keystore)
	if err != nil {
		log.Panicln(err)
	}

	endpoint, _ := url.Parse(cfg.endpoint)

	proxy := &httputil.ReverseProxy{
		BufferPool: bpool.NewBytePool(16, 1024*8),
		Director: func(req *http.Request) {
			if token := req.Header.Get(xAuthorization); token == "" {
				var body []byte
				if req.Body != nil {
					body, _ = ioutil.ReadAll(req.Body)
					_ = req.Body.Close()
					req.Body = ioutil.NopCloser(bytes.NewReader(body))
				}

				sig := mixin.SignRaw(req.Method, req.URL.String(), body)
				token := auth.SignToken(sig, mixin.RandomTraceID(), time.Minute)
				req.Header.Set(xAuthorization, "Bearer "+token)
			}

			// mixin api server 屏蔽来自 proxy 的请求
			// 这里在转发请求的时候不带上 X-Forwarded-For
			// https://github.com/golang/go/issues/38079 go 1.15 上线
			req.Header["X-Forwarded-For"] = nil

			req.Host = endpoint.Host
			req.URL.Host = endpoint.Host
			req.URL.Scheme = endpoint.Scheme
		},
	}

	var handler http.Handler = proxy

	if cfg.getOnly {
		handler = allowMethod(http.MethodGet)(proxy)
	}

	// cors
	handler = cors.AllowAll().Handler(handler)

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: handler,
	}

	if err := svr.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func allowMethod(methods ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			for _, method := range methods {
				if r.Method == method {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusMethodNotAllowed)
		}

		return http.HandlerFunc(fn)
	}
}
