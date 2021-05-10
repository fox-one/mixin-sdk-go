package mixin

import (
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
	blazeHost = DefaultBlazeHost
	blazePath = "/"
)

func UseBlazeHost(host string) {
	blazeHost = host
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

	if path, ok := os.LookupEnv("MIXIN_SDK_BLAZE_PATH"); ok && path != "" {
		blazePath = path
	}

	if hosts, ok := os.LookupEnv("MIXIN_SDK_MIXINNET_HOSTS"); ok && hosts != "" {
		UseMixinNetHosts(strings.Split(hosts, ","))
	}
}
