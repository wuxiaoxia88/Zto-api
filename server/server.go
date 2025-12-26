package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"zto-api-proxy/config"
	"zto-api-proxy/logger"
	"zto-api-proxy/proxy"
)

//go:embed static/*
var staticFS embed.FS

// ProxyRecord 代理请求记录
type ProxyRecord struct {
	ID         string    `json:"id"`
	Time       time.Time `json:"time"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	StatusCode int       `json:"statusCode"`
	Duration   int64     `json:"duration"` // 毫秒
}

// Server HTTP 服务器
type Server struct {
	proxyClient *proxy.Client
	refreshFunc func() error
	httpServer  *http.Server
	history     []ProxyRecord
	historyLock sync.RWMutex
	lastFetch   time.Time
	zboxStatus  string
	zboxPid     string
	zboxLock    sync.Mutex
}

// NewServer 创建服务器
func NewServer(proxyClient *proxy.Client, refreshFunc func() error) *Server {
	s := &Server{
		proxyClient: proxyClient,
		refreshFunc: refreshFunc,
		zboxStatus:  "检测中...",
	}
	s.CheckZBox() // 启动时检查一次
	return s
}

// Start 启动服务器
func (s *Server) Start() error {
	cfg := config.GetConfig()

	mux := http.NewServeMux()

	// 静态资源控制面板
	fSys, _ := fs.Sub(staticFS, "static")
	mux.Handle("/", http.FileServer(http.FS(fSys)))

	// 透传模式 API
	mux.HandleFunc("/proxy", s.handleProxy)

	// 便捷模式 API
	mux.HandleFunc("/orders", s.handleOrders)
	mux.HandleFunc("/orders/todo", s.handleOrdersTodo)
	mux.HandleFunc("/province-report", s.handleProvinceReport)

	// 管理 API
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/refresh", s.handleRefresh)
	mux.HandleFunc("/health", s.handleHealth)

	// GUI 专用控制 API
	mux.HandleFunc("/admin/open-logs", s.handleOpenLogs)
	mux.HandleFunc("/admin/open-debug", s.handleOpenDebug)
	mux.HandleFunc("/admin/recent-logs", s.handleRecentLogs)
	mux.HandleFunc("/admin/request-logs", s.handleRequestLogs)
	mux.HandleFunc("/admin/config", s.handleGetConfig)
	mux.HandleFunc("/admin/save-config", s.handleSaveConfig)
	mux.HandleFunc("/admin/clear-logs", s.handleClearLogs)

	// 兼容性/自定义 API 路径
	mux.HandleFunc("/api/query/order_trace", s.handleLegacyOrders)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler:      s.corsMiddleware(s.logMiddleware(mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	logger.Info("HTTP 服务器启动在 http://0.0.0.0:%d", cfg.Port)
	return s.httpServer.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if s.httpServer != nil {
		return s.httpServer.Close()
	}
	return nil
}

// 中间件：CORS
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// 中间件：日志
func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Debug("HTTP %s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

// 透传代理
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, http.StatusMethodNotAllowed, "只支持 POST 方法")
		return
	}

	var req proxy.ProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, http.StatusBadRequest, "无效的请求格式: "+err.Error())
		return
	}

	if req.URL == "" {
		s.jsonError(w, http.StatusBadRequest, "url 是必需的")
		return
	}

	startTime := time.Now()
	resp := s.proxyClient.DoRequest(&req)
	duration := time.Since(startTime).Milliseconds()
	s.addHistory(r.Method, req.URL, resp.StatusCode, duration)
	if resp.StatusCode == 200 {
		s.lastFetch = time.Now()
	}
	s.jsonResponse(w, resp)
}

// 订单查询（便捷模式）
func (s *Server) handleOrders(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	start := query.Get("start")
	end := query.Get("end")
	page, _ := strconv.Atoi(query.Get("page"))
	size, _ := strconv.Atoi(query.Get("size"))
	siteCode := query.Get("siteCode")
	empCode := query.Get("empCode")

	if start == "" {
		start = time.Now().Format("2006-01-02") + " 00:00:00"
	}
	if end == "" {
		end = time.Now().Format("2006-01-02") + " 23:59:59"
	}
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 50
	}

	searchSiteCodeList := []string{}
	if siteCode != "" {
		searchSiteCodeList = append(searchSiteCodeList, siteCode)
	}

	searchEmpCodeList := []string{}
	if empCode != "" {
		searchEmpCodeList = append(searchEmpCodeList, empCode)
	}

	body := map[string]interface{}{
		"traceQueryChannel":      "ALL_PICK_CHANNEL",
		"traceAbnormalMarkQuery": "",
		"traceTimeRequireList":   []interface{}{},
		"baseList":               []interface{}{},
		"orderStatusList":        []interface{}{},
		"searchEmpCodeList":      searchEmpCodeList,
		"searchSiteCodeList":     searchSiteCodeList,
		"orderTypeList":          []interface{}{},
		"partnerIds":             []interface{}{},
		"traceQueryTime":         "ORDER_CREATE_TIME",
		"querySendAddress":       "",
		"queryReceiveAddress":    "",
		"pickUpCodeStatus":       "",
		"payStatus":              "",
		"appealStatusList":       []interface{}{},
		"orderType":              0,
		"startTime":              start,
		"endTime":                end,
		"querySendProv":          "",
		"querySendCity":          "",
		"querySendCounty":        "",
		"queryReceiveProv":       "",
		"queryReceiveCity":       "",
		"queryReceiveCounty":     "",
		"sortField":              "",
		"sortType":               0,
		"pageNum":                page,
		"pageSize":               size,
		"pageIndex":              page,
	}

	req := &proxy.ProxyRequest{
		URL:    "https://preorder-query-center.gw.zt-express.com/preOrderQuery/getSiteOrderTraceList",
		Method: "POST",
		Body:   body,
	}

	startTime := time.Now()
	resp := s.proxyClient.DoRequest(req)
	duration := time.Since(startTime).Milliseconds()
	s.addHistory(r.Method, req.URL, resp.StatusCode, duration)
	s.jsonResponse(w, resp)
}

// 兼容旧版/自定义路径的订单查询
func (s *Server) handleLegacyOrders(w http.ResponseWriter, r *http.Request) {
	token := config.GetTokenData()
	var errStr string
	if token == nil || len(token.Cookies) == 0 {
		errStr = "Token数据为空"
	} else if _, ok := token.Cookies["wyzdzjxhdnh"]; !ok {
		errStr = "缺少核心Cookie(wyzdzjxhdnh)"
	} else if token.ExpiresAt.IsZero() {
		errStr = "Token过期时间缺失"
	} else if !time.Now().Before(token.ExpiresAt) {
		errStr = fmt.Sprintf("Token已于 %s 过期", token.ExpiresAt.Format("15:04:05"))
	}

	if errStr != "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"api_id":    "order_trace",
			"success":   false,
			"data":      nil,
			"error":     "Token无效: " + errStr,
			"timestamp": time.Now().Format("2006-01-02T15:04:05.000000"),
		})
		return
	}

	// 尝试解析 body 中的 pageSize
	size := 50
	var bodyParams map[string]interface{}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&bodyParams)
		if s, ok := bodyParams["pageSize"].(float64); ok {
			size = int(s)
		}
	}

	// 复用 handleOrders 的逻辑但自定义返回
	body := map[string]interface{}{
		"traceQueryChannel": "ALL_PICK_CHANNEL",
		"startTime":         time.Now().Format("2006-01-02") + " 00:00:00",
		"endTime":           time.Now().Format("2006-01-02") + " 23:59:59",
		"pageNum":           1,
		"pageSize":          size,
		"pageIndex":         1,
	}
	req := &proxy.ProxyRequest{
		URL:    "https://preorder-query-center.gw.zt-express.com/preOrderQuery/getSiteOrderTraceList",
		Method: "POST",
		Body:   body,
	}

	startTime := time.Now()
	resp := s.proxyClient.DoRequest(req)
	duration := time.Since(startTime).Milliseconds()
	s.addHistory("POST", req.URL, resp.StatusCode, duration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"api_id":        "order_trace",
		"success":       resp.StatusCode == 200,
		"data":          resp.Data,
		"error":         "",
		"timestamp":     time.Now().Format("2006-01-02T15:04:05.000000"),
		"duration_ms":   duration,
		"custom_params": bodyParams,
	})
}

// 待办事项
func (s *Server) handleOrdersTodo(w http.ResponseWriter, r *http.Request) {
	body := map[string]interface{}{
		"traceQueryChannel": "ALL_PICK_CHANNEL",
	}

	// 如果请求带了 Body，则透传 Body
	if r.Method == "POST" && r.Body != nil {
		var customBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&customBody); err == nil && len(customBody) > 0 {
			body = customBody
		}
	}

	req := &proxy.ProxyRequest{
		URL:    "https://preorder-query-center.gw.zt-express.com/preOrderQuery/getTodoCenterList",
		Method: "POST",
		Body:   body,
	}

	startTime := time.Now()
	resp := s.proxyClient.DoRequest(req)
	duration := time.Since(startTime).Milliseconds()
	s.addHistory(r.Method, req.URL, resp.StatusCode, duration)
	s.jsonResponse(w, resp)
}

// 省市区报表
func (s *Server) handleProvinceReport(w http.ResponseWriter, r *http.Request) {
	// 默认参数 (从 Query 获取)
	query := r.URL.Query()
	date := query.Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	page, _ := strconv.Atoi(query.Get("page"))
	size, _ := strconv.Atoi(query.Get("size"))
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 100
	}

	body := map[string]interface{}{
		"empCode":                   "",
		"complianceResult":          "",
		"complianceResultQueryList": []interface{}{},
		"orderServiceType":          []interface{}{},
		"startTime":                 date,
		"endTime":                   date,
		"batchList":                 []interface{}{},
		"comparisonQueryCode":       "",
		"ddzlCompare":               "",
		"ddzlCompareType":           1,
		"ddzlCompareCount":          nil,
		"provinceName":              query.Get("province"),
		"cityName":                  query.Get("city"),
		"tiktokArea":                "",
		"streetName":                "",
		"siteCode":                  query.Get("siteCode"),
		"siteName":                  "",
		"sortType":                  1,
		"sortField":                 "",
		"whetherPreDepart":          nil,
		"whetherNewAreaBatch":       nil,
		"queryDateRangeType":        1,
		"pageSize":                  size,
		"pageIndex":                 page,
	}

	// 如果是 POST 且有 Body，则完全使用 Body
	if r.Method == "POST" && r.Body != nil {
		var customBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&customBody); err == nil && len(customBody) > 0 {
			body = customBody
		}
	}

	req := &proxy.ProxyRequest{
		URL:    "https://orderapi.zt-express.com/opsApi/zjProvinceReport/queryZjPreOrderReport",
		Method: "POST",
		Body:   body,
	}

	startTime := time.Now()
	resp := s.proxyClient.DoRequest(req)
	duration := time.Since(startTime).Milliseconds()
	s.addHistory(r.Method, req.URL, resp.StatusCode, duration)
	s.jsonResponse(w, resp)
}

// 状态查询
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	token := config.GetTokenData()
	cfg := config.GetConfig()

	s.zboxLock.Lock()
	zboxStatus := s.zboxStatus
	zboxPid := s.zboxPid
	s.zboxLock.Unlock()

	status := map[string]interface{}{
		"service":     "running",
		"port":        cfg.Port,
		"tokenValid":  config.IsTokenValid(),
		"expiresAt":   "",
		"lastRefresh": "",
		"lastFetch":   "",
		"zboxStatus":  zboxStatus,
		"zboxPid":     zboxPid,
	}

	if !s.lastFetch.IsZero() {
		status["lastFetch"] = s.lastFetch.Format(time.RFC3339)
	}

	if token != nil {
		if !token.ExpiresAt.IsZero() {
			status["expiresAt"] = token.ExpiresAt.Format(time.RFC3339)
		}
		if !token.LastRefresh.IsZero() {
			status["lastRefresh"] = token.LastRefresh.Format(time.RFC3339)
		}
	}

	s.jsonResponse(w, status)
}

// 手动刷新
func (s *Server) handleRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, http.StatusMethodNotAllowed, "只支持 POST 方法")
		return
	}

	s.CheckZBox() // 刷新前检查一下宝盒环境
	err := s.refreshFunc()
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "刷新失败: "+err.Error())
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"success": true,
		"message": "Token 刷新成功",
	})
}

// CheckZBox 探测宝盒状态
func (s *Server) CheckZBox() {
	s.zboxLock.Lock()
	defer s.zboxLock.Unlock()

	status := "未运行"
	pid := ""

	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	if out, err := cmd.Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(strings.ToLower(line), "zbox.exe") ||
				strings.Contains(strings.ToLower(line), "lite.exe") {
				parts := strings.Split(line, ",")
				if len(parts) > 1 {
					pid = strings.Trim(parts[1], "\"")
					status = "运行中"
					break
				}
			}
		}
	}

	s.zboxStatus = status
	s.zboxPid = pid
	logger.Info("宝盒进程探测结果: %s (PID: %s)", status, pid)
}

// 健康检查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, map[string]string{"status": "ok"})
}

// 打开日志目录
func (s *Server) handleOpenLogs(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	logsDir := filepath.Join(cfg.DataDir, "logs")
	os.MkdirAll(logsDir, 0755)
	cmd := exec.Command("explorer", logsDir)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Start()
	s.jsonResponse(w, map[string]bool{"success": true})
}

// 打开调试截图目录
func (s *Server) handleOpenDebug(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	debugDir := filepath.Join(cfg.DataDir, "debug")
	os.MkdirAll(debugDir, 0755)
	cmd := exec.Command("explorer", debugDir)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}
	cmd.Start()
	s.jsonResponse(w, map[string]bool{"success": true})
}

// 获取最近的日志内容
func (s *Server) handleRecentLogs(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetConfig()
	logsDir := filepath.Join(cfg.DataDir, "logs")

	// 1. 尝试读取当天的
	today := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logsDir, fmt.Sprintf("service_%s.log", today))

	data, err := os.ReadFile(logFile)
	if err != nil {
		// 2. 如果今天还没有日志，尝试找最新的一个文件
		files, _ := os.ReadDir(logsDir)
		var latestFile string
		for i := len(files) - 1; i >= 0; i-- {
			if !files[i].IsDir() && strings.HasPrefix(files[i].Name(), "service_") {
				latestFile = filepath.Join(logsDir, files[i].Name())
				break
			}
		}

		if latestFile == "" {
			s.jsonResponse(w, []string{"[SYSTEM] [INFO] 暂无日志记录"})
			return
		}
		data, err = os.ReadFile(latestFile)
		if err != nil {
			s.jsonResponse(w, []string{"[SYSTEM] [ERROR] 无法读取历史日志文件"})
			return
		}
	}

	lines := strings.Split(string(data), "\n")
	// 过滤空行
	var validLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			validLines = append(validLines, l)
		}
	}

	if len(validLines) > 50 {
		validLines = validLines[len(validLines)-50:]
	}

	s.jsonResponse(w, validLines)
}

// 获取请求日志记录
func (s *Server) handleRequestLogs(w http.ResponseWriter, r *http.Request) {
	s.historyLock.RLock()
	defer s.historyLock.RUnlock()
	s.jsonResponse(w, s.history)
}

// 清空历史记录
func (s *Server) handleClearLogs(w http.ResponseWriter, r *http.Request) {
	s.historyLock.Lock()
	s.history = []ProxyRecord{}
	s.historyLock.Unlock()
	s.jsonResponse(w, map[string]bool{"success": true})
}

// 获取配置
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	s.jsonResponse(w, config.GetConfig())
}

// 保存配置
func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.jsonError(w, http.StatusMethodNotAllowed, "只支持 POST")
		return
	}
	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		s.jsonError(w, http.StatusBadRequest, "无效的配置参数")
		return
	}
	// 这里目前只允许保存部分安全参数，且保存后需要重启生效
	config.SetCustomConfig(&newCfg)
	s.jsonResponse(w, map[string]interface{}{"success": true, "message": "配置已保存，部分设置需重启后生效"})
}

func (s *Server) addHistory(method, url string, statusCode int, duration int64) {
	s.historyLock.Lock()
	defer s.historyLock.Unlock()

	// 截取 URL 长度
	displayURL := url
	if len(displayURL) > 80 {
		displayURL = displayURL[:77] + "..."
	}

	record := ProxyRecord{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Time:       time.Now(),
		Method:     method,
		URL:        displayURL,
		StatusCode: statusCode,
		Duration:   duration,
	}

	s.history = append([]ProxyRecord{record}, s.history...)
	if len(s.history) > 100 {
		s.history = s.history[:100]
	}
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

func (s *Server) jsonError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}
