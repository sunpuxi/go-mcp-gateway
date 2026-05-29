package domain

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPClient 封装 HTTP 请求调用
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DoRequest 执行一次 HTTP 请求，支持按请求设置超时（通过 context）
// 返回响应体和响应状态码，err 表示网络/传输层错误
func (c *HTTPClient) DoRequest(method, url string, header http.Header, body []byte, timeoutMs int) (int, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(strings.ToUpper(method), url, reqBody)
	if err != nil {
		return 0, nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 Header
	for k, vals := range header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 每个请求独立超时控制，不修改共享的 client.Timeout
	if timeoutMs > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
		defer cancel()
		req = req.WithContext(ctx)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("请求下游失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	return resp.StatusCode, respBody, nil
}
