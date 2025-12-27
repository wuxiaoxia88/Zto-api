package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Config 应用配置
type Config struct {
	Port           int    `json:"port"`
	ChromePath     string `json:"chromePath"`
	DataDir        string `json:"dataDir"`
	LoginURL       string `json:"loginUrl"`
	RefreshTime    string `json:"refreshTime"` // 凌晨刷新时间 "00:05"
	PreventTime    string `json:"preventTime"` // 预防刷新时间 "19:30"
	MaxRetries     int    `json:"maxRetries"`
	RetryDelay     int    `json:"retryDelay"`     // 毫秒
	RequestTimeout int    `json:"requestTimeout"` // 秒
}

// TokenData Token 存储结构
type TokenData struct {
	Cookies     map[string]string `json:"cookies"`
	LastRefresh time.Time         `json:"lastRefresh"`
	ExpiresAt   time.Time         `json:"expiresAt"`  // 综合失效时间 (最早的那个)
	AppExpire   time.Time         `json:"appExpire"`  // wyandyy 失效时间 (10h)
	SessExpire  time.Time         `json:"sessExpire"` // wyzdzjxhdnh 失效时间 (14d)
}

var (
	cfg     *Config
	cfgOnce sync.Once
	cfgLock sync.RWMutex

	tokenData *TokenData
	tokenLock sync.RWMutex
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:           8765,
		ChromePath:     "", // 自动检测
		DataDir:        getDefaultDataDir(),
		LoginURL:       "https://www.zt-express.com",
		RefreshTime:    "00:05",
		PreventTime:    "19:30",
		MaxRetries:     3,
		RetryDelay:     1000,
		RequestTimeout: 30,
	}
}

func getDefaultDataDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		appData = "."
	}
	return filepath.Join(appData, "zto-api-proxy")
}

// GetConfig 获取配置（单例）
func GetConfig() *Config {
	cfgOnce.Do(func() {
		cfg = DefaultConfig()
		loadConfig()
	})
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return cfg
}

func loadConfig() {
	configPath := filepath.Join(cfg.DataDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return // 使用默认配置
	}
	json.Unmarshal(data, cfg)
}

// SaveConfig 保存配置
func SaveConfig() error {
	cfgLock.RLock()
	defer cfgLock.RUnlock()

	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(cfg.DataDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// SetCustomConfig 更新并保存配置
func SetCustomConfig(newCfg *Config) error {
	cfgLock.Lock()
	defer cfgLock.Unlock()

	cfg.Port = newCfg.Port
	cfg.RefreshTime = newCfg.RefreshTime
	cfg.PreventTime = newCfg.PreventTime
	cfg.MaxRetries = newCfg.MaxRetries

	// 保存到文件
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return err
	}
	configPath := filepath.Join(cfg.DataDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

// GetTokenData 获取 Token 数据
func GetTokenData() *TokenData {
	tokenLock.RLock()
	defer tokenLock.RUnlock()

	if tokenData == nil {
		loadTokenData()
	}
	return tokenData
}

// SetTokenData 设置并保存 Token 数据
func SetTokenData(data *TokenData) error {
	tokenLock.Lock()
	defer tokenLock.Unlock()

	tokenData = data
	return saveTokenData()
}

func loadTokenData() {
	cfg := GetConfig()
	tokenPath := filepath.Join(cfg.DataDir, "token.json")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		tokenData = &TokenData{
			Cookies: make(map[string]string),
		}
		return
	}
	tokenData = &TokenData{}
	json.Unmarshal(data, tokenData)
	if tokenData.Cookies == nil {
		tokenData.Cookies = make(map[string]string)
	}
}

func saveTokenData() error {
	cfg := GetConfig()
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return err
	}

	tokenPath := filepath.Join(cfg.DataDir, "token.json")
	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenPath, data, 0644)
}

// IsTokenValid 检查 Token 是否有效
func IsTokenValid() bool {
	token := GetTokenData()
	if token == nil || len(token.Cookies) == 0 {
		return false
	}

	// 检查是否有必需的 cookies
	if _, ok := token.Cookies["wyzdzjxhdnh"]; !ok {
		return false
	}

	// 检查是否过期
	if token.ExpiresAt.IsZero() {
		return false
	}

	return time.Now().Before(token.ExpiresAt)
}

// GetCookieString 获取 Cookie 字符串
func GetCookieString() string {
	token := GetTokenData()
	if token == nil {
		return ""
	}

	result := ""
	for name, value := range token.Cookies {
		if result != "" {
			result += "; "
		}
		result += name + "=" + value
	}
	return result
}
