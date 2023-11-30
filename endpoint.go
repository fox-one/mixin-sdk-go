package mixin

import (
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	DefaultApiHost   = "https://api.mixin.one"
	DefaultBlazeHost = "blaze.mixin.one"

	ZeromeshApiHost   = "https://mixin-api.zeromesh.net"
	ZeromeshBlazeHost = "mixin-blaze.zeromesh.net"

	EchoApiHost = "https://echo.yiplee.com"
)

func UseApiHost(host string) {
	httpClient.HostURL = host
}

var (
	blazeURL = buildBlazeURL(DefaultBlazeHost)
)

func buildBlazeURL(host string) string {
	u := url.URL{Scheme: "wss", Host: host}
	return u.String()
}

func UseBlazeHost(host string) {
	blazeURL = buildBlazeURL(host)
}

func UseBlazeURL(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	blazeURL = u.String()
}

func useApi(url string) <-chan string {
	r := make(chan string)
	go func() {
		defer close(r)
		_, err := http.Get(url)
		if err == nil {
			r <- url
		}
	}()
	return r
}

func timer() <-chan string {
	r := make(chan string)
	go func() {
		defer close(r)
		time.Sleep(time.Second * 30)
		r <- ""
	}()
	return r
}

func UseAutoFasterRoute() {
	for {
		var r string
		select {
		case r = <-useApi(DefaultApiHost):
		case r = <-useApi(ZeromeshApiHost):
		case r = <-timer():
		}
		if r == DefaultApiHost {
			UseApiHost(DefaultApiHost)
			UseBlazeHost(DefaultBlazeHost)
		} else if r == ZeromeshApiHost {
			UseApiHost(ZeromeshApiHost)
			UseBlazeHost(ZeromeshBlazeHost)
		}
		time.Sleep(time.Minute * 5)
	}
}

func init() {
	if _, ok := os.LookupEnv("MIXIN_SDK_USE_ZEROMESH"); ok {
		UseApiHost(ZeromeshApiHost)
		UseBlazeHost(ZeromeshBlazeHost)
	}

	if host, ok := os.LookupEnv("MIXIN_SDK_API_HOST"); ok && host != "" {
		UseApiHost(host)
	}

	if host, ok := os.LookupEnv("MIXIN_SDK_BLAZE_HOST"); ok && host != "" {
		UseBlazeHost(host)
	}

	if rawURL, ok := os.LookupEnv("MIXIN_SDK_BLAZE_URL"); ok && rawURL != "" {
		UseBlazeURL(rawURL)
	}
}
