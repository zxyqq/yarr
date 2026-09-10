# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

yarr (yet another rss reader) 是一个基于 Web 的 RSS 订阅聚合器，编译为单一二进制文件（内嵌 SQLite 数据库）。支持桌面应用（带系统托盘）和自托管服务器两种模式。

**Project structure (`src/`):**
| Dir | Purpose |
|---|---|
| `assets/` | Frontend — HTML, JS, CSS, graphic arts |
| `content/` | Content fetching/processing |
| `parser/` | Feed parser (RSS/Atom) |
| `platform/` | OS-specific code |
| `server/` | HTTP server, routing, auth, Fever API, OPML |
| `storage/` | SQLite storage layer (feeds, folders, items, migrations) |
| `systray/` | System tray GUI |
| `worker/` | Background feed refresh worker |

## 常用命令

```bash
# 构建 CLI 版本（当前平台）
make host

# 开发模式运行（使用 debug 标签，热加载前端资源）
make serve

# 运行所有测试
make test

# 运行单个包的测试
go test -tags "sqlite_foreign_keys sqlite_json" -v ./src/parser/

# 运行单个测试函数
go test -tags "sqlite_foreign_keys sqlite_json" -v ./src/parser/ -run TestRSSParse

# 交叉编译（需要 Zig >= 0.14.0）
make linux_amd64
make windows_amd64
```

构建依赖：Go >= 1.23、C 编译器（GCC/Clang）、CGO_ENABLED=1（SQLite 需要）。

## 架构概览

### 后端（Go）

- **入口**: `cmd/yarr/main.go` — 解析命令行参数，初始化 storage 和 server
- **HTTP 服务**: `src/server/` — 路由定义在 `routes.go`，handler 在同目录各文件
  - 自定义路由器: `src/server/router/`（非标准库，支持路径参数 `:id`）
  - 中间件: gzip (`src/server/gzip/`)、认证 (`src/server/auth/`)、OPML (`src/server/opml/`)
- **数据层**: `src/storage/` — SQLite 操作，所有数据库 schema 变更在 `migrate()` 函数中管理
- **后台任务**: `src/worker/` — Feed 刷新、Favicon 获取、Feed 清理
- **Feed 解析**: `src/parser/` — 支持 RSS、Atom、RDF、JSON Feed 格式
- **内容处理**: `src/content/` — HTML 清洗 (`sanitizer/`)、正文提取 (`readability/`)、抓取 (`scraper/`)、HTML 工具 (`htmlutil/`)
- **平台适配**: `src/platform/` — 系统托盘、浏览器打开、版本信息等平台特定代码
  - `gui` build tag 启用桌面 GUI 模式

### 前端（内嵌资源）

- **位置**: `src/assets/` — 通过 `//go:embed` 嵌入二进制
- **技术栈**: 原生 JavaScript + Vue.js 2 + Bootstrap CSS，无构建工具
- **关键文件**: `index.html`（主页面）、`javascripts/app.js`（应用逻辑）、`javascripts/api.js`（API 调用）、`javascripts/key.js`（键盘快捷键）
- **模板分隔符**: `{%` `%}`（Go template，避免与 Vue.js `{{ }}` 冲突）
- **Debug 模式**: `debug` build tag 下直接从 `src/assets/` 目录读取文件（无需重新编译即可看到前端改动）

### 数据流

1. `cmd/yarr/main.go` → 初始化 `storage.Storage`（SQLite）和 `server.Server`
2. Server 启动时初始化 `worker.Worker`，定时刷新 feeds
3. 前端通过 `/api/*` REST 端点与后端交互
4. Fever API 兼容层在 `src/server/fever.go`

## 关键约定

- 构建标签: `sqlite_foreign_keys sqlite_json` 是必须的，所有 `go build` 和 `go test` 命令都需要
- API 路由遵循 RESTful 风格: GET 列表/详情、POST 创建、PUT 更新、DELETE 删除
- 前端无构建步骤 — 直接编辑 JS/CSS/HTML 文件，`debug` 模式下刷新浏览器即可看到变更
- 数据库迁移在 `src/storage/migration.go` 的 `migrate()` 中线性执行
