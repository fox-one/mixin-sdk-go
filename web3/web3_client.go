package web3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/go-resty/resty/v2"
)

const (
	MixinRouteClientID    = "61cb8dd4-16b1-4744-ba0c-7b2d2e52fc59"
	HeaderAccessTimestamp = "MR-ACCESS-TIMESTAMP"
	HeaderAccessSign      = "MR-ACCESS-SIGN"
	HeaderContentType     = "Content-Type"
	ContentTypeJSON       = "application/json"
)

// Web3Client 接口定义
type Web3Client interface {
	DoRequest(ctx context.Context, method, path string, query string, body interface{}, result interface{}) error
	Get(ctx context.Context, path string, query string, result interface{}) error
	Post(ctx context.Context, path string, body interface{}, result interface{}) error
}

// web3ClientImpl 实现
type web3ClientImpl struct {
	baseURL   string
	botClient *bot.BotAuthClient
	clientID  string
	client    *resty.Client
}

// Web3ClientOption 定义客户端选项
type Web3ClientOption func(*web3ClientImpl)

// NewWeb3Client 创建客户端
func NewWeb3Client(botClient *bot.BotAuthClient, opts ...Web3ClientOption) Web3Client {
	client := &web3ClientImpl{
		baseURL:   MixinRouteApiPrefix,
		botClient: botClient,
		clientID:  MixinRouteClientID,
		client: resty.New().
			SetBaseURL(MixinRouteApiPrefix).
			SetHeader(HeaderContentType, ContentTypeJSON).
			SetTimeout(10 * time.Second),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// WithBaseURL 设置基础URL
func WithBaseURL(url string) Web3ClientOption {
	return func(c *web3ClientImpl) {
		c.baseURL = url
		c.client.SetBaseURL(url)
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) Web3ClientOption {
	return func(c *web3ClientImpl) {
		c.client.SetTimeout(timeout)
	}
}

// WithRetry 设置重试策略
func WithRetry(count int, waitTime time.Duration) Web3ClientOption {
	return func(c *web3ClientImpl) {
		c.client.SetRetryCount(count).
			SetRetryWaitTime(waitTime).
			SetRetryMaxWaitTime(waitTime * 2)
	}
}

func (c *web3ClientImpl) DoRequest(ctx context.Context, method, path string, query string, body interface{}, result interface{}) (err error) {
	// 准备请求
	req := c.client.R().
		SetContext(ctx).
		SetResult(result)
	if body != nil {
		req.SetBody(body)
	}

	url := path
	if len(query) > 0 {
		url = fmt.Sprintf("%s?%s", path, query)
	}

	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body for signing: %w", err)
		}
	}

	ts := time.Now().Unix()
	httpReq, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request for signing: %w", err)
	}
	httpReq.Header.Set(HeaderContentType, ContentTypeJSON)

	signature, err := c.botClient.SignRequest(ctx, ts, c.clientID, httpReq)
	if err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	req.SetHeaders(map[string]string{
		HeaderAccessTimestamp: fmt.Sprintf("%d", ts),
		HeaderAccessSign:      signature,
	})
	resp, err := req.Execute(method, url)
	if err != nil {
		return fmt.Errorf("execute request: %w", err)
	}

	if !resp.IsSuccess() || resp.StatusCode() == http.StatusAccepted {
		var errResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errResp); err != nil {
			return &MixinOracleAPIError{
				StatusCode:  resp.StatusCode(),
				Description: resp.String(),
				RawBody:     resp.String(),
			}
		}

		return &MixinOracleAPIError{
			StatusCode:  resp.StatusCode(),
			Code:        errResp.Error.Code,
			Description: errResp.Error.Description,
			RawBody:     resp.String(),
		}
	}
	return nil
}

func (c *web3ClientImpl) Get(ctx context.Context, path string, query string, result interface{}) error {
	return c.DoRequest(ctx, "GET", path, query, nil, result)
}

func (c *web3ClientImpl) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.DoRequest(ctx, "POST", path, "", body, result)
}
