# ZTO API Proxy (中通 API 代理服务) 🚀

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D6?style=flat&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

一个专业、高效的中通快递 API 代理与自动化管理方案。本项目旨在解决中通业务 API 调用过程中的 Token 维护难、身份校验繁琐以及跨域限制等痛点，提供一站式的可视化控制中心。

---

## ✨ 核心亮点

### 🎨 极简控制中心 (Web GUI)
内置现代化的暗黑风格控制面板，基于原生 HTML/CSS/JS 构建，无外部依赖，极速响应。
- **双 Token 精确追踪**：同时解析并展示 `wyandyy` (应用授权, 10h) 与 `wyzdzjxhdnh` (会话授权, 14d) 的精确失效时间。
- **自愈式运行控制台**：动态滚动系统日志，支持跨天日志自动回溯，确保监控不留白。
- **智能 API 测试终端**：
    - **增强型 cURL 解析**：支持解析复杂、多行的浏览器 cURL 命令，**自动提取所有业务 Headers** 及 Body 载荷。
    - **自定义 Headers 支持**：新增专用 Headers 编辑区，支持 `siteinfo` 等关键业务头部的动态透传。
    - **跨域透传代理**：内置 `/proxy` 接口，自动通过后端中转远程 API，彻底规避浏览器 CORS 限制。

### 🤖 自动化 Token 体系
- **无人值守刷新**：基于 Playwright 自动化引擎，模拟真实浏览器行为进行一键登录。
- **智能预刷新策略**：
    - **动态过期监测**：每分钟自检 Token 状态，在**任意 Token** 即将过期前 30 分钟自动静默刷新。
    - **常规维护刷新**：每日 `00:05` 强制同步最新状态。
    - **失败自动重试**，确保 24/7 服务可用。

### 🖥️ Windows 原生体验
- **精美视觉识别**：全画幅自定义“Z”标识图标，完美融入 Windows 任务栏。
- **静默后台运行**：编译为 `windowsgui` 模式，无黑窗口干扰，用户全流程通过托盘和 Web 界面交互。
- **极低资源占用**：优化进程探测逻辑，由“暴力轮询”升级为“按需触发”。

---

## 🛠️ 安装与运行

### 系统要求
- Windows 10/11 (x64)
- **Google Chrome 浏览器** (自动化刷新必备)
- **宝盒 (ZBox/Lite)** 已在系统中登录

### 获取程序
您可以从 [Releases](https://github.com/wuxiaoxia88/Zto-api/releases) 下载最新的 `ZTO_API_Proxy_Final.exe`。

### 快速启动
1. 双击运行 `ZTO_API_Proxy_Final.exe`。
2. 程序将静默运行，并自动打开默认浏览器访问控制中心：`http://localhost:8765`。
3. 观察托盘区是否出现蓝色“Z”图标。

---

## 📡 API 调用指南

服务默认监听 `http://localhost:8765`，支持以下两种模式：

### 1. 便捷业务接口 (内置 Token 处理)
| 路径 | 方法 | 说明 |
| :--- | :--- | :--- |
| `/api/query/order_trace` | `POST` | 预约单轨迹查询 |
| `/orders/todo` | `POST` | 待办事项汇总 |
| `/province-report` | `POST` | 字节省市区数据报表 |

---

## 🔍 调试与测试指令

您可以直接使用 PowerShell 进行快速测试和服务验证：

### 🩺 系统诊断
```powershell
# 1. 健康状态检测 (返回 ok 表示服务存活)
Invoke-RestMethod -Uri "http://localhost:8765/health"

# 2. 查看完整运行状态 (Token 有效期、宝盒 PID 等)
Invoke-RestMethod -Uri "http://localhost:8765/status" | ConvertTo-Json

# 3. 手动触发一次 Token 自动刷新流程
Invoke-RestMethod -Method Post -Uri "http://localhost:8765/refresh"
```

### 📦 业务接口测试示例
```powershell
# 测试：预约单跟单查询
$body = @{
    traceQueryChannel = "ALL_PICK_CHANNEL"
    startTime = "2025-12-25 00:00:00"
    endTime = "2025-12-25 23:59:59"
    pageNum = 1
    pageSize = 10
} | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri "http://localhost:8765/api/query/order_trace" -Body $body -ContentType "application/json"
```

---

### 2. 通用透传代理 (`/proxy`)
如果您需要调用其他中通接口（如账单、网点明细等），直接将请求发送至此路由。支持全量透传 Headers：
```bash
POST http://localhost:8765/proxy
Content-Type: application/json

{
  "url": "https://szapi.zt-express.com/send-bills/...",
  "method": "POST",
  "headers": {
    "siteinfo": "eyJzaXRlQ29kZSI6IjUxMjA4Iiw... ",
    "x-zop-ns": "shenzhou-"
  },
  "body": { "pageSize": 100, "dateType": 1 }
}
```

---

## 🏗️ 开发者指南

### 编译项目
如果您想自行编译，请确保已安装 Go 1.20+ 环境：

```bash
# 1. 编译 Windows 资源文件 (图标和清单)
rsrc -manifest main.manifest -ico tray/icon.ico -o main.syso

# 2. 构建生产版可执行文件 (隐藏黑窗口)
go build -ldflags="-s -w -H windowsgui" -o "ZTO_API_Proxy.exe" .
```

### 项目结构
```text
├── browser/    # 自动化浏览器交互 (Chromedp)
├── server/     # 嵌入式控制中心 (Vanilla HTML/JS)
├── tray/       # 系统托盘交互逻辑
├── proxy/      # 核心 HTTP 透传引擎 (支持 Headers 解析)
├── config/     # 配置持久化与 Token 解析 (JWT Sync)
└── scheduler/  # 智能预刷新任务调度
```

---

## 📜 许可证
本项目采用 [MIT License](LICENSE) 开源。

---
**ZTO API Proxy** - 让中通 API 调用回归简单。
如果您觉得本项目对您有帮助，欢迎点一个 **Star** ⭐
