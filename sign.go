package mixin

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/oxtoacart/bpool"
)

var bufferPool = bpool.NewSizedBufferPool(16, 256)

func SignRaw(method, uri string, body []byte) string {
	b := bufferPool.Get()
	defer bufferPool.Put(b)

	b.WriteString(method)
	b.WriteString(uri)
	b.Write(body)
	sum := sha256.Sum256(b.Bytes())
	return hex.EncodeToString(sum[:])
}

func SignRequest(r *http.Request) string {
	method := r.Method
	uri := trimURLHost(r.URL)

	var body []byte
	if r.GetBody != nil {
		if b, _ := r.GetBody(); b != nil {
			body, _ = io.ReadAll(b)
			_ = b.Close()
		}
	}

	return SignRaw(method, uri, body)
}

func SignResponse(r *resty.Response) string {
	method := r.Request.Method
	uri := trimURLHost(r.Request.RawRequest.URL)
	return SignRaw(method, uri, r.Body())
}

func trimURLHost(u *url.URL) string {
	path := u.Path
	if path == "" || path == "/" {
		return "/"
	}

	uri := u.String()

	if idx := strings.Index(uri, path); idx >= 0 {
		uri = uri[idx:]
	}

	return uri
}
