package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"zto-api-proxy/config"
	"zto-api-proxy/logger"
)

// Browser 浏览器自动化
type Browser struct {
	chromeDataDir string
}

// NewBrowser 创建浏览器实例
func NewBrowser() *Browser {
	cfg := config.GetConfig()
	return &Browser{
		chromeDataDir: filepath.Join(cfg.DataDir, "chrome-data"),
	}
}

// RefreshToken 通过自动登录刷新 Token
func (b *Browser) RefreshToken() error {
	logger.Token("开始自动刷新 Token...")

	if !b.isZBoxRunning() {
		return fmt.Errorf("宝盒未运行，无法自动登录")
	}

	ctx, cancel := b.createContext()
	defer cancel()

	ctx, timeoutCancel := context.WithTimeout(ctx, 150*time.Second)
	defer timeoutCancel()

	var cookies []*network.Cookie
	var screenshot []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.zt-express.com"),

		// 1. 等待网页充分加载（用户建议 10 秒左右）
		chromedp.Sleep(12*time.Second),

		// 2. 模拟人工点击一键登录
		b.tryClickLogin(),

		// 3. 点击后等待流程完成（用户建议再等 5-10 秒，这里给 15 秒确保充分加载）
		chromedp.Sleep(15*time.Second),

		// 4. 轮询获取 Cookie（增加轮询次数）
		chromedp.ActionFunc(func(ctx context.Context) error {
			for i := 0; i < 15; i++ {
				var err error
				cookies, err = network.GetCookies().Do(ctx)
				if err != nil {
					return err
				}

				found := false
				for _, c := range cookies {
					if c.Name == "wyzdzjxhdnh" {
						found = true
						break
					}
				}
				if found {
					logger.Token("轮询检测：获取到核心 Token Cookie")
					return nil
				}
				logger.Token("轮询检测：核心 Cookie 尚未出现 (%d/15)", i+1)
				chromedp.Sleep(2 * time.Second).Do(ctx)
			}

			// 失败时截屏
			chromedp.CaptureScreenshot(&screenshot).Do(ctx)
			debugDir := filepath.Join(config.GetConfig().DataDir, "debug")
			os.MkdirAll(debugDir, 0755)
			screenshotPath := filepath.Join(debugDir, "login_failed.png")
			os.WriteFile(screenshotPath, screenshot, 0644)
			logger.Warn("Token 刷新失败，截图已保存: %s", screenshotPath)

			return fmt.Errorf("未能获取 wyzdzjxhdnh cookie (最终共 %d 个)", len(cookies))
		}),
	)

	if err != nil {
		return err
	}

	tokenData := &config.TokenData{
		Cookies:     make(map[string]string),
		LastRefresh: time.Now(),
	}

	for _, cookie := range cookies {
		if strings.Contains(cookie.Domain, "zt-express.com") {
			tokenData.Cookies[cookie.Name] = cookie.Value
			if cookie.Name == "wyandyy" {
				tokenData.ExpiresAt = extractExpireTime(cookie.Value)
			}
		}
	}

	if _, ok := tokenData.Cookies["wyzdzjxhdnh"]; !ok {
		return fmt.Errorf("解析结果中缺失核心 Token")
	}

	if err := config.SetTokenData(tokenData); err != nil {
		return fmt.Errorf("保存 Token 失败: %w", err)
	}

	logger.Token("Token 刷新成功，有效期至 %s", tokenData.ExpiresAt.Format("2006-01-02 15:04:05"))
	return nil
}

func (b *Browser) createContext() (context.Context, context.CancelFunc) {
	// 确保数据目录存在
	os.MkdirAll(b.chromeDataDir, 0755)

	// 配置 Chrome 选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true), // 生产环境使用 headless 模式
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", false),
		chromedp.UserDataDir(b.chromeDataDir),
		chromedp.WindowSize(1920, 1080),
	)

	// 检查是否有自定义 Chrome 路径
	cfg := config.GetConfig()
	if cfg.ChromePath != "" {
		opts = append(opts, chromedp.ExecPath(cfg.ChromePath))
	}

	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx,
		chromedp.WithLogf(func(format string, args ...interface{}) {
			logger.Debug("[Chrome] "+format, args...)
		}),
	)

	return ctx, cancel
}

func (b *Browser) tryClickLogin() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		logger.Token("尝试定位一键登录按钮坐标...")

		var coords []float64
		// 获取按钮中心坐标
		err := chromedp.Evaluate(`
			(function() {
				const findBtn = () => {
					const all = document.querySelectorAll('button, a, div, span');
					// 优先找蓝色的按钮（一键登录通常是一个大按钮）
					return Array.from(all).find(el => 
						el.textContent.trim() === '一键登录' && 
						el.offsetWidth > 50 && 
						el.offsetHeight > 20
					);
				};
				const btn = findBtn();
				if (btn) {
					const rect = btn.getBoundingClientRect();
					return [rect.left + rect.width / 2, rect.top + rect.height / 2];
				}
				return null;
			})()
		`, &coords).Do(ctx)

		if err != nil || len(coords) != 2 {
			logger.Warn("脚本定位按钮坐标失败，尝试使用备用文本搜索点击")
			return chromedp.Click("//*[contains(text(), '一键登录')]", chromedp.BySearch).Do(ctx)
		}

		logger.Token("按钮定位成功: (%.2f, %.2f)，执行模拟点击", coords[0], coords[1])

		// 模拟真实的鼠标点击
		return chromedp.MouseClickXY(coords[0], coords[1]).Do(ctx)
	}
}

func (b *Browser) isZBoxRunning() bool {
	// 使用 wmic 命令检测进程，避免中文编码问题
	cmd := exec.Command("wmic", "process", "where", "name like '%zbox%' or name like '%lite%'", "get", "processid")
	output, err := cmd.Output()
	if err != nil {
		// 备用方案：检查进程列表
		cmd2 := exec.Command("powershell", "-Command", "Get-Process | Where-Object {$_.ProcessName -like '*zbox*' -or $_.ProcessName -like '*lite*'} | Select-Object -First 1")
		output2, err2 := cmd2.Output()
		if err2 != nil {
			logger.Debug("进程检测失败: %v", err2)
			return false
		}
		return len(strings.TrimSpace(string(output2))) > 0
	}
	// 检查输出是否包含进程ID（数字）
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "ProcessId" {
			// 找到了进程ID
			return true
		}
	}
	return false
}

func extractExpireTime(jwtStr string) time.Time {
	// 默认 20:00 (兜底)
	now := time.Now()
	defaultExp := time.Date(now.Year(), now.Month(), now.Day(), 20, 0, 0, 0, now.Location())
	if now.After(defaultExp) {
		defaultExp = defaultExp.Add(24 * time.Hour)
	}

	parts := strings.Split(jwtStr, ".")
	if len(parts) < 2 {
		return defaultExp
	}

	payload, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return defaultExp
	}

	var data struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &data); err != nil || data.Exp == 0 {
		return defaultExp
	}

	return time.Unix(data.Exp, 0)
}

func truncateString(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
