package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"zto-api-proxy/config"
	"zto-api-proxy/logger"
)

// ProxyRequest 代理请求结构
type ProxyRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

// ProxyResponse 代理响应结构
type ProxyResponse struct {
	Success     bool        `json:"success"`
	StatusCode  int         `json:"statusCode"`
	Data        interface{} `json:"data"`
	Error       string      `json:"error,omitempty"`
	RequestTime string      `json:"requestTime"`
	Duration    int64       `json:"duration"` // 毫秒
}

// Client HTTP 客户端
type Client struct {
	httpClient    *http.Client
	onNeedRefresh func() error
}

// NewClient 创建代理客户端
func NewClient(onNeedRefresh func() error) *Client {
	cfg := config.GetConfig()
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.RequestTimeout) * time.Second,
		},
		onNeedRefresh: onNeedRefresh,
	}
}

// DoRequest 执行代理请求（带重试）
func (c *Client) DoRequest(req *ProxyRequest) *ProxyResponse {
	cfg := config.GetConfig()
	startTime := time.Now()

	var lastErr error
	var resp *ProxyResponse

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			logger.API("重试请求 (%d/%d): %s", attempt, cfg.MaxRetries, req.URL)
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Millisecond)
		}

		resp, lastErr = c.doSingleRequest(req)

		if lastErr == nil && resp.Success {
			resp.Duration = time.Since(startTime).Milliseconds()
			logger.API("%s %s -> %d (%dms)", req.Method, truncateURL(req.URL), resp.StatusCode, resp.Duration)
			return resp
		}

		// 如果是 401/403，尝试刷新 Token
		if resp != nil && (resp.StatusCode == 401 || resp.StatusCode == 403) {
			logger.Token("检测到认证失败 (%d)，尝试刷新 Token", resp.StatusCode)
			if c.onNeedRefresh != nil {
				if err := c.onNeedRefresh(); err != nil {
					logger.Error("Token 刷新失败: %v", err)
				} else {
					logger.Token("Token 刷新成功，重新请求")
					continue
				}
			}
		}
	}

	// 所有重试都失败
	duration := time.Since(startTime).Milliseconds()
	if resp == nil {
		resp = &ProxyResponse{
			Success:     false,
			StatusCode:  0,
			Error:       fmt.Sprintf("请求失败: %v", lastErr),
			RequestTime: startTime.Format(time.RFC3339),
			Duration:    duration,
		}
	}

	logger.Error("请求最终失败: %s -> %s", req.URL, resp.Error)
	return resp
}

func (c *Client) doSingleRequest(req *ProxyRequest) (*ProxyResponse, error) {
	// 序列化请求体
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// 创建 HTTP 请求
	method := req.Method
	if method == "" {
		method = "GET"
	}

	httpReq, err := http.NewRequest(method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认 headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/plain, */*")
	httpReq.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	httpReq.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	httpReq.Header.Set("Origin", "https://www.zt-express.com")
	httpReq.Header.Set("Referer", "https://www.zt-express.com/")

	// 添加自定义 headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 添加 cookies
	cookieStr := config.GetCookieString()
	if cookieStr != "" {
		httpReq.Header.Set("Cookie", cookieStr)
	}

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析响应
	var data interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &data); err != nil {
			// 如果不是 JSON，返回原始字符串
			data = string(respBody)
		}
	}

	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return &ProxyResponse{
		Success:     success,
		StatusCode:  resp.StatusCode,
		Data:        data,
		RequestTime: time.Now().Format(time.RFC3339),
	}, nil
}

func truncateURL(url string) string {
	if len(url) > 80 {
		return url[:80] + "..."
	}
	return url
}
