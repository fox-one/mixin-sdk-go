package mixin

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
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

func UseAutoFasterRoute() {
	for {
		var w sync.WaitGroup
		var defaultTime time.Time
		var zeromeshTime time.Time
		w.Add(1)
		go func() {
			_, err := http.Get(DefaultApiHost)
			if err == nil {
				defaultTime = time.Now()
				w.Done()
			}
		}()
		go func() {
			_, err := http.Get(ZeromeshApiHost)
			if err == nil {
				zeromeshTime = time.Now()
				w.Done()
			}
		}()
		go func() {
			time.Sleep(time.Second * 30)
			w.Done()
		}()
		w.Wait()
		if defaultTime.IsZero() {
			UseApiHost(ZeromeshApiHost)
			UseBlazeHost(ZeromeshBlazeHost)
		} else if zeromeshTime.IsZero() {
			UseApiHost(DefaultApiHost)
			UseBlazeHost(DefaultBlazeHost)
		} else if defaultTime.After(zeromeshTime) {
			UseApiHost(ZeromeshApiHost)
			UseBlazeHost(ZeromeshBlazeHost)
		} else {
			UseApiHost(DefaultApiHost)
			UseBlazeHost(DefaultBlazeHost)
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

	if hosts, ok := os.LookupEnv("MIXIN_SDK_MIXINNET_HOSTS"); ok && hosts != "" {
		UseMixinNetHosts(strings.Split(hosts, ","))
	}
}
