package mixinnet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type (
	Client struct {
		http.Client
		safe  bool
		hosts []string
	}
)

var (
	defaultLegacyMixinNetClient = NewClient(false)
	defaultSafeMixinNetClient   = NewClient(true)
)

func DefaultClient(safe bool) *Client {
	if safe {
		return defaultSafeMixinNetClient
	} else {
		return defaultLegacyMixinNetClient
	}
}

func NewClient(safe bool, hosts ...string) *Client {
	if len(hosts) == 0 {
		if safe {
			hosts = safeHosts
		} else {
			hosts = legacyHosts
		}
	}
	return &Client{
		hosts: hosts,
		safe:  safe,
		Client: http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) SetHosts(hosts []string) {
	c.hosts = hosts
}

func (c *Client) CallMixinNetRPC(ctx context.Context, resp interface{}, method string, params ...interface{}) error {
	bts, err := json.Marshal(map[string]interface{}{
		"method": method,
		"params": params,
	})
	if err != nil {
		return err
	}

	r, err := c.Post(c.HostFromContext(ctx), "application/json", bytes.NewReader(bts))
	if err != nil {
		return err
	}

	return UnmarshalResponse(r, resp)
}

func DecodeResponse(resp *http.Response) ([]byte, error) {
	var body struct {
		Error string          `json:"error,omitempty"`
		Data  json.RawMessage `json:"data,omitempty"`
	}
	defer resp.Body.Close()
	if err := json.NewDecoder((resp.Body)).Decode(&body); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, createError(resp.StatusCode, resp.StatusCode, resp.Status)
		}
		return nil, createError(resp.StatusCode, resp.StatusCode, err.Error())
	}

	if body.Error != "" {
		return nil, parseError(body.Error)
	}

	return body.Data, nil
}

func UnmarshalResponse(resp *http.Response, v interface{}) error {
	data, err := DecodeResponse(resp)
	if err != nil {
		return err
	}

	if v != nil {
		return json.Unmarshal(data, v)
	}

	return nil
}
