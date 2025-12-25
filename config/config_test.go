package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 8765 {
		t.Errorf("期望端口 8765, 实际 %d", cfg.Port)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("期望最大重试 3, 实际 %d", cfg.MaxRetries)
	}

	if cfg.RefreshTime != "00:05" {
		t.Errorf("期望刷新时间 00:05, 实际 %s", cfg.RefreshTime)
	}
}

func TestTokenDataSaveLoad(t *testing.T) {
	// 使用临时目录
	tmpDir := filepath.Join(os.TempDir(), "zto-api-proxy-test")
	defer os.RemoveAll(tmpDir)

	// 修改配置使用临时目录
	cfg := GetConfig()
	cfg.DataDir = tmpDir

	// 创建测试 Token
	testToken := &TokenData{
		Cookies: map[string]string{
			"wyzdzjxhdnh": "test-token-1",
			"wyandyy":     "test-token-2",
		},
		LastRefresh: time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	// 保存
	if err := SetTokenData(testToken); err != nil {
		t.Fatalf("保存 Token 失败: %v", err)
	}

	// 重新加载
	loadTokenData()
	loaded := GetTokenData()

	if loaded.Cookies["wyzdzjxhdnh"] != "test-token-1" {
		t.Errorf("Token 加载不正确")
	}
}

func TestIsTokenValid(t *testing.T) {
	// 设置有效 Token
	validToken := &TokenData{
		Cookies: map[string]string{
			"wyzdzjxhdnh": "test-token",
		},
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	tokenData = validToken
	if !IsTokenValid() {
		t.Error("有效 Token 应返回 true")
	}

	// 设置过期 Token
	expiredToken := &TokenData{
		Cookies: map[string]string{
			"wyzdzjxhdnh": "test-token",
		},
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	tokenData = expiredToken
	if IsTokenValid() {
		t.Error("过期 Token 应返回 false")
	}
}

func TestGetCookieString(t *testing.T) {
	tokenData = &TokenData{
		Cookies: map[string]string{
			"cookie1": "value1",
			"cookie2": "value2",
		},
	}

	cookieStr := GetCookieString()

	if cookieStr == "" {
		t.Error("Cookie 字符串不应为空")
	}

	// 检查是否包含两个 cookie
	if len(cookieStr) < 10 {
		t.Error("Cookie 字符串长度不正确")
	}
}
