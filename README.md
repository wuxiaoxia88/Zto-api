# ZTO API Proxy (中通 API 代理服务) 🚀

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Windows-0078D6?style=flat&logo=windows)](https://www.microsoft.com/windows)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

一个专业、高效的中通快递 API 代理与自动化管理方案。本项目旨在解决中通业务 API 调用过程中的 Token 维护难、身份校验繁琐以及跨域限制等痛点，提供一站式的可视化控制中心。

---

## ✨ 核心亮点

### 🎨 极简控制中心 (Web GUI)
内置现代化的暗黑风格控制面板，基于原生 HTML/CSS/JS 构建，无外部依赖，极速响应。
- **实时仪表盘**：一目了然的授权状态、Token 过期预测、宝盒 (ZBox) 环境检测。
- **实时控制台**：动态滚动系统运行日志，随时掌握后端自动化逻辑。
- **智能 API 测试终端**：
    - **cURL 智能解析**：直接粘贴浏览器复制的 cURL 命令，系统自动提取 URL、Method 和 JSON 载荷，并智能映射到代理路径。
    - **内置业务模版**：预设“预约单跟单”、“字节省市区报表”等常用业务指令，一键填充。
    - **跨域透传代理**：内置 `/proxy` 接口，自动通过后端中转远程 API，彻底规避浏览器 CORS 限制。

### 🤖 自动化 Token 体系
- **无人值守刷新**：基于 Playwright 自动化引擎，模拟真实浏览器行为进行一键登录。
- **智能调度策略**：
    - 每日 `00:05` 定时更新。
    - 每日 `19:30` 预防性刷新（对抗官方 20:00 的过期机制）。
    - 失败自动重试，确保 24/7 服务可用。

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

### 2. 通用透传代理 (`/proxy`)
如果您需要调用其他中通接口，直接将请求发送至此路由：
```bash
POST http://localhost:8765/proxy
Content-Type: application/json

{
  "url": "https://szapi.zt-express.com/siteJournal/...",
  "method": "POST",
  "body": { ... }
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
├── browser/    # 自动化浏览器交互 (Playwright)
├── server/     # 嵌入式控制中心与路由处理
├── tray/       # 系统托盘交互逻辑
├── proxy/      # 核心 HTTP 请求引擎
├── config/     # 配置持久化 (%APPDATA%)
└── scheduler/  # 自动化任务调度
```

---

## 📜 许可证
本项目采用 [MIT License](LICENSE) 开源。

---
**ZTO API Proxy** - 让中通 API 调用回归简单。
如果您觉得本项目对您有帮助，欢迎点一个 **Star** ⭐
