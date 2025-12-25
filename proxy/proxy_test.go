package proxy

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoRequest_Success(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(nil)

	req := &ProxyRequest{
		URL:    server.URL,
		Method: "GET",
	}

	resp := client.DoRequest(req)

	if !resp.Success {
		t.Errorf("请求应该成功: %s", resp.Error)
	}

	if resp.StatusCode != 200 {
		t.Errorf("期望状态码 200, 实际 %d", resp.StatusCode)
	}
}

func TestDoRequest_Retry(t *testing.T) {
	attempts := 0

	// 创建模拟服务器，前两次失败，第三次成功
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient(nil)

	req := &ProxyRequest{
		URL:    server.URL,
		Method: "GET",
	}

	resp := client.DoRequest(req)

	if attempts < 3 {
		t.Errorf("应该重试至少3次, 实际 %d 次", attempts)
	}

	if !resp.Success {
		t.Errorf("最终请求应该成功")
	}
}

func TestDoRequest_WithBody(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"received": "ok"})
	}))
	defer server.Close()

	client := NewClient(nil)

	req := &ProxyRequest{
		URL:    server.URL,
		Method: "POST",
		Body: map[string]interface{}{
			"startTime": "2025-12-24 00:00:00",
			"endTime":   "2025-12-24 23:59:59",
			"pageSize":  50,
		},
	}

	resp := client.DoRequest(req)

	if !resp.Success {
		t.Errorf("请求应该成功: %s", resp.Error)
	}

	if receivedBody["pageSize"] != float64(50) {
		t.Errorf("请求体未正确传递")
	}
}

func TestTruncateURL(t *testing.T) {
	shortURL := "https://example.com/api"
	if truncateURL(shortURL) != shortURL {
		t.Error("短 URL 不应被截断")
	}

	longURL := "https://example.com/api/very/long/path/that/exceeds/eighty/characters/in/length/for/testing"
	truncated := truncateURL(longURL)
	if len(truncated) > 83 { // 80 + "..."
		t.Errorf("长 URL 应被截断, 实际长度 %d", len(truncated))
	}
}
