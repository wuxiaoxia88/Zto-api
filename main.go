package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"zto-api-proxy/browser"
	"zto-api-proxy/config"
	"zto-api-proxy/logger"
	"zto-api-proxy/proxy"
	"zto-api-proxy/scheduler"
	"zto-api-proxy/server"
	"zto-api-proxy/tray"
)

var (
	version    = "1.0.0"
	noTray     bool
	testMode   bool
	refreshNow bool
)

func init() {
	flag.BoolVar(&noTray, "no-tray", false, "禁用系统托盘")
	flag.BoolVar(&testMode, "test", false, "测试模式（仅检查配置）")
	flag.BoolVar(&refreshNow, "refresh", false, "立即刷新 Token")
}

func main() {
	flag.Parse()

	fmt.Printf("ZTO API Proxy v%s\n", version)
	fmt.Println("================================")

	// 初始化配置
	cfg := config.GetConfig()

	// 确保数据目录存在
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		fmt.Printf("创建数据目录失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logDir := filepath.Join(cfg.DataDir, "logs")
	if err := logger.Init(logDir); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("ZTO API Proxy 启动中...")
	logger.Info("数据目录: %s", cfg.DataDir)
	logger.Info("监听端口: %d", cfg.Port)

	// 测试模式
	if testMode {
		runTestMode()
		return
	}

	// 创建浏览器实例
	browserInstance := browser.NewBrowser()

	// 创建服务器实例指针，以便在 refreshFunc 中闭包引用
	var srv *server.Server

	// 创建刷新函数
	refreshFunc := func() error {
		err := browserInstance.RefreshToken()
		if err == nil && srv != nil {
			srv.CheckZBox()
		}
		return err
	}

	// 立即刷新模式
	if refreshNow {
		logger.Info("执行立即刷新...")
		if err := refreshFunc(); err != nil {
			logger.Error("刷新失败: %v", err)
			os.Exit(1)
		}
		logger.Info("刷新成功")
		return
	}

	// 创建代理客户端
	proxyClient := proxy.NewClient(refreshFunc)

	// 创建调度器
	sched := scheduler.NewScheduler(refreshFunc)
	sched.Start()
	defer sched.Stop()

	// 创建服务器
	srv = server.NewServer(proxyClient, refreshFunc)

	// 停止函数
	stopFunc := func() {
		logger.Info("正在停止服务...")
		sched.Stop()
		srv.Stop()
		os.Exit(0)
	}

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		stopFunc()
	}()

	// 启动服务器（后台）
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()

	// 在非 -no-tray 模式下，尝试自动打开控制台
	if !noTray && !refreshNow {
		go func() {
			time.Sleep(1 * time.Second) // 等待服务器启动
			url := fmt.Sprintf("http://localhost:%d", cfg.Port)
			logger.Info("正在打开控制中心: %s", url)
			exec.Command("cmd", "/c", "start", url).Run()
		}()
	}

	// 检查 Token 状态
	if !config.IsTokenValid() {
		logger.Warn("Token 无效或已过期，建议刷新")
	} else {
		token := config.GetTokenData()
		logger.Info("Token 有效，过期时间: %s", token.ExpiresAt.Format("2006-01-02 15:04:05"))
	}

	// 启动托盘或等待
	if noTray {
		logger.Info("无托盘模式，按 Ctrl+C 退出")
		select {} // 永久等待
	} else {
		// 启动系统托盘（阻塞）
		t := tray.NewTray(refreshFunc, stopFunc)
		t.Run()
	}
}

func runTestMode() {
	fmt.Println("\n[测试模式]")

	// 检查配置
	cfg := config.GetConfig()
	fmt.Printf("✓ 配置加载成功\n")
	fmt.Printf("  - 端口: %d\n", cfg.Port)
	fmt.Printf("  - 数据目录: %s\n", cfg.DataDir)
	fmt.Printf("  - 刷新时间: %s\n", cfg.RefreshTime)
	fmt.Printf("  - 预防时间: %s\n", cfg.PreventTime)
	fmt.Printf("  - 最大重试: %d\n", cfg.MaxRetries)

	// 检查 Token
	if config.IsTokenValid() {
		token := config.GetTokenData()
		fmt.Printf("✓ Token 有效\n")
		fmt.Printf("  - 过期时间: %s\n", token.ExpiresAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  - 上次刷新: %s\n", token.LastRefresh.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("✗ Token 无效或已过期\n")
	}

	// 检查浏览器
	browserInstance := browser.NewBrowser()
	_ = browserInstance // 用于后续测试
	fmt.Printf("✓ 浏览器模块初始化成功\n")

	fmt.Println("\n测试完成")
}
