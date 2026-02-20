# HandlerDirSearch

## 📖 项目简介

HandlerDirSearch 是一个基于 Wails + Vue 开发的桌面应用程序，用于批量扫描和分析 URL。该工具可以从包含 HTTP 状态码（200、301、403）的日志文件中提取 URL，并对这些 URL 进行并发扫描，提取页面标题、组件信息等关键数据，最终生成详细的 Markdown 格式报告。

## ✨ 功能特性

- 📁 **智能文件解析**：自动从日志文件中提取 HTTP 200、301、403 状态码对应的 URL
- 🚀 **高效并发扫描**：支持自定义并发数（默认 30，最大 100），大幅提升扫描效率
- ⏱️ **灵活超时控制**：可配置请求超时时间（默认 5 秒，最大 120 秒）
- 🔄 **重定向控制**：支持选择是否跟随 HTTP 重定向
- 📊 **详细报告生成**：自动生成 Markdown 格式的扫描报告，包含 URL、标题、组件信息和错误详情
- 🎯 **组件识别**：自动识别网页中使用的技术栈和组件（通过响应头和 HTML 内容检测）
- 🖥️ **跨平台支持**：支持 Windows、macOS 和 Linux 系统

## 📦 下载安装

### Release 版本下载

您可以从以下渠道下载最新的 Release 版本：

- **GitHub Releases**：访问项目的 GitHub Releases 页面下载对应操作系统的安装包
- **Windows**：下载 `.exe` 安装程序或便携版


## 🚀 使用方法

### 基本使用流程

1. **启动应用**：双击运行安装后的应用程序
2. **选择输入文件**：点击"选择文件"按钮，选择包含 URL 的日志文件（支持 .txt、.log、.md、.csv 格式）
3. **配置扫描参数**：
   - **并发数**：设置同时扫描的 URL 数量（建议 30-100）
   - **超时时间**：设置每个请求的超时时间（建议 5-30 秒）
   - **跟随重定向**：勾选是否跟随 HTTP 3xx 重定向
4. **开始扫描**：点击"开始扫描"按钮
5. **查看报告**：扫描完成后，报告将自动保存到输入文件所在目录下的 `scan_report.md`

### 输入文件格式

输入文件应包含 HTTP 状态码日志，程序会自动提取状态码为 200、301 或 403 的行中的 URL。

示例输入文件内容：
```
200 GET https://example.com/page1
301 https://example.com/old-path
403 POST https://example.com/admin
200 https://example.com/page3
```

### 报告格式

生成的 Markdown 报告包含以下信息：
- 扫描时间
- 输入文件路径
- 总匹配状态码行数（200、301、403）
- 提取的 URL 总数
- 成功扫描数
- 失败扫描数
- 详细的扫描结果表格（URL、标题、组件、错误信息）

## 🛠️ 开发说明

### 环境要求

- **Go**：1.18 或更高版本
- **Node.js**：16 或更高版本
- **Wails CLI**：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 安装依赖

```bash
# 安装 Go 依赖
go mod download

# 安装前端依赖
cd frontend
npm install
```

### 开发模式

运行以下命令启动开发模式：

```bash
wails dev
```

这将启动：
- Vite 开发服务器，提供快速的热重载
- Go 后端服务
- 浏览器开发服务器（http://localhost:34115），可以在浏览器中调试并调用 Go 方法

### 项目结构

```
handlerdirsearch/
├── app.go           # 应用主逻辑和 API
├── scanner.go       # URL 扫描核心功能
├── report.go        # 报告生成功能
├── main.go          # Wails 应用入口
├── frontend/        # Vue 前端代码
│   ├── src/
│   │   ├── App.vue
│   │   ├── components/
│   │   └── main.js
│   └── package.json
└── wails.json       # Wails 项目配置
```

## 📦 构建发布

### 构建生产版本

在项目根目录运行：

```bash
wails build
```

### 构建特定平台

```bash
# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/amd64
wails build -platform darwin/arm64

# Linux
wails build -platform linux/amd64
```

## ⚙️ 配置说明

项目配置文件为 `wails.json`，可以配置以下内容：

- `name`：应用名称
- `outputfilename`：输出文件名
- `author`：作者信息

更多配置选项请参考：https://wails.io/docs/reference/project-config

