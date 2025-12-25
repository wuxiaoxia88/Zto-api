package tray

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/getlantern/systray"

	"zto-api-proxy/config"
	"zto-api-proxy/logger"
)

//go:embed icon.ico
var trayIcon []byte

// Tray 系统托盘
type Tray struct {
	refreshFunc func() error
	stopFunc    func()
}

// NewTray 创建托盘
func NewTray(refreshFunc func() error, stopFunc func()) *Tray {
	return &Tray{
		refreshFunc: refreshFunc,
		stopFunc:    stopFunc,
	}
}

// Run 运行托盘（阻塞）
func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

func (t *Tray) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("ZTO API Proxy")
	systray.SetTooltip("中通 API 代理服务")

	// 状态
	mStatus := systray.AddMenuItem("状态: 运行中", "服务状态")
	mStatus.Disable()

	systray.AddSeparator()

	// 刷新 Token
	mRefresh := systray.AddMenuItem("刷新 Token", "手动刷新认证 Token")

	// 打开日志
	mLogs := systray.AddMenuItem("打开日志", "打开日志文件夹")

	// 打开配置
	mConfig := systray.AddMenuItem("打开配置", "打开配置文件夹")

	systray.AddSeparator()

	// 退出
	mQuit := systray.AddMenuItem("退出", "退出程序")

	// 事件处理
	go func() {
		for {
			select {
			case <-mRefresh.ClickedCh:
				logger.Info("用户触发手动刷新")
				if t.refreshFunc != nil {
					if err := t.refreshFunc(); err != nil {
						logger.Error("手动刷新失败: %v", err)
					}
				}

			case <-mLogs.ClickedCh:
				cfg := config.GetConfig()
				logsDir := filepath.Join(cfg.DataDir, "logs")
				os.MkdirAll(logsDir, 0755)
				exec.Command("explorer", logsDir).Start()

			case <-mConfig.ClickedCh:
				cfg := config.GetConfig()
				cmd := exec.Command("explorer", cfg.DataDir)
				if runtime.GOOS == "windows" {
					cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				}
				cmd.Start()

			case <-mQuit.ClickedCh:
				logger.Info("用户退出程序")
				if t.stopFunc != nil {
					t.stopFunc()
				}
				systray.Quit()
				return
			}
		}
	}()

	logger.Info("系统托盘已启动")
}

func (t *Tray) onExit() {
	logger.Info("系统托盘已退出")
}

// UpdateStatus 更新状态显示
func (t *Tray) UpdateStatus(status string) {
	systray.SetTooltip(fmt.Sprintf("中通 API 代理服务 - %s", status))
}
