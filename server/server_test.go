package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"zto-api-proxy/proxy"
)

func TestHandleHealth(t *testing.T) {
	srv := NewServer(nil, nil)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200, 实际 %d", w.Code)
	}

	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != "ok" {
		t.Errorf("期望 status=ok")
	}
}

func TestHandleStatus(t *testing.T) {
	srv := NewServer(nil, nil)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	srv.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200, 实际 %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["service"] != "running" {
		t.Errorf("期望 service=running")
	}
}

func TestHandleProxy_InvalidMethod(t *testing.T) {
	srv := NewServer(nil, nil)

	req := httptest.NewRequest("GET", "/proxy", nil)
	w := httptest.NewRecorder()

	srv.handleProxy(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("期望 405, 实际 %d", w.Code)
	}
}

func TestHandleProxy_MissingURL(t *testing.T) {
	srv := NewServer(nil, nil)

	body := strings.NewReader(`{"method": "GET"}`)
	req := httptest.NewRequest("POST", "/proxy", body)
	w := httptest.NewRecorder()

	srv.handleProxy(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("期望 400, 实际 %d", w.Code)
	}
}

func TestHandleProxy_Success(t *testing.T) {
	// 创建模拟目标服务器
	targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer targetServer.Close()

	proxyClient := proxy.NewClient(nil)
	srv := NewServer(proxyClient, nil)

	body := strings.NewReader(`{"url": "` + targetServer.URL + `", "method": "GET"}`)
	req := httptest.NewRequest("POST", "/proxy", body)
	w := httptest.NewRecorder()

	srv.handleProxy(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("期望 200, 实际 %d", w.Code)
	}
}

func TestHandleRefresh_InvalidMethod(t *testing.T) {
	srv := NewServer(nil, nil)

	req := httptest.NewRequest("GET", "/refresh", nil)
	w := httptest.NewRecorder()

	srv.handleRefresh(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("期望 405, 实际 %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	srv := NewServer(nil, nil)

	handler := srv.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("缺少 CORS header")
	}
}
