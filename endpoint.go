package mixin

import (
	"net/url"
	"os"
	"strings"
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
