# ZTO API Proxy (中通 API 代理服务)

一个轻量级的中通快递 API 代理服务，自动管理登录状态和 Token 刷新。

## 功能特性

- ✅ **自动登录**: 通过 Headless Chrome 模拟一键登录，自动选择账号
- ✅ **Token 管理**: 自动检测 Token 有效期，定时刷新
- ✅ **API 代理**: 支持透传模式和便捷模式两种 API 调用方式
- ✅ **失败重试**: 请求失败自动重试 3 次
- ✅ **系统托盘**: 提供状态显示和手动操作入口
- ✅ **日志记录**: 完整的操作日志

## 系统要求

- Windows 10/11
- Google Chrome 浏览器
- 宝盒 (zbox-lite) 已登录运行

## 快速开始

### 1. 启动服务

```bash
# 正常启动（带托盘图标）
zto-api-proxy.exe

# 无托盘模式启动
zto-api-proxy.exe -no-tray

# 测试配置
zto-api-proxy.exe -test

# 立即刷新 Token
zto-api-proxy.exe -refresh
```

### 2. 使用 API

服务启动后监听 `http://0.0.0.0:8765`

#### 健康检查
```bash
GET http://localhost:8765/health
```

#### 查看状态
```bash
GET http://localhost:8765/status
```

#### 手动刷新 Token
```bash
POST http://localhost:8765/refresh
```

## API 接口

### 透传模式

完全自定义请求，适合灵活调用：

```bash
POST http://localhost:8765/proxy
Content-Type: application/json

{
  "url": "https://xxx.gw.zt-express.com/...",
  "method": "POST",
  "body": {...}
}
```

### 便捷模式

封装常用查询：

```bash
# 订单列表查询
GET http://localhost:8765/orders?start=2025-12-24 00:00:00&end=2025-12-24 23:59:59&page=1&size=50

# 待办事项
GET http://localhost:8765/orders/todo

# 省市区报表
GET http://localhost:8765/report/province?date=2025-12-24
```

## 配置

配置文件位置: `%APPDATA%\zto-api-proxy\config.json`

```json
{
  "port": 8765,
  "chromePath": "",
  "refreshTime": "00:05",
  "preventTime": "19:30",
  "maxRetries": 3,
  "requestTimeout": 30
}
```

## Token 刷新策略

| 时机 | 说明 |
|------|------|
| 每日 00:05 | 凌晨定时刷新 |
| 每日 19:30 | 预防性刷新（避免 20:00 过期） |
| 请求失败时 | 检测到 401/403 时自动刷新 |

## 日志

日志文件位置: `%APPDATA%\zto-api-proxy\logs\`

```
service_2025-12-24.log  # 按日期滚动
```

## 开发

```bash
# 编译
go build -o zto-api-proxy.exe .

# 编译优化版本（更小体积）
go build -ldflags="-s -w" -o zto-api-proxy.exe .

# 运行测试
go test ./... -v
```

## 项目结构

```
zto-api-proxy/
├── main.go           # 入口
├── browser/          # 浏览器自动化
├── config/           # 配置管理
├── logger/           # 日志模块
├── proxy/            # 请求转发
├── scheduler/        # 定时任务
├── server/           # HTTP 服务器
└── tray/             # 系统托盘
```

## 注意事项

1. 使用前请确保宝盒已登录
2. 首次使用建议先运行 `zto-api-proxy.exe -refresh` 获取 Token
3. Token 有效期至当天 20:00，服务会自动刷新

## License

MIT
