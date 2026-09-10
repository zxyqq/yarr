# Git 同步与版本分支操作规范

维护本地功能增强版 yarr 时，与 upstream 同步的标准流程。

## 分支模型

```
upstream/master          上游仓库（nkanaev/yarr），只读
master                   [已冻结] 基于 v2.7 的旧开发线，仅作历史参考
master-v2.9              v2.9 + 本地全部修改（当前活跃主线）
master-v2.10             下一版本 ∈ master-v2.9 + v2.10 差异
...
```

原则：

- 每个 `master-vX.Y` 分支固定对应一个上游 tag，不跨版本漂移。
- **不要**再 `merge master`（旧 v2.7 线）进任何 `master-v2.9` 之后的分支：
  旧 master 的 `src/assets/javascripts/*` 文件已被上游删除，重复合并会复活
  modify/delete 冲突。老 master 到此退役。
- 本地新功能开发应基于当前最新的 `master-vX.Y` 分支（新前端结构 `src/frontend/`）。

## 升级到新版本（如 upstream 发布 v2.10）

```bash
git fetch upstream
git checkout -b master-v2.10 v2.10   # 从上游 tag 直接拉新分支
git merge master-v2.9                # 把上一版成果合进来
```

冲突解决原则：

| 冲突类型 | 取谁 |
|---|---|
| `src/frontend/*` 前端冲突 | 两者融合：上游新代码为基底，保住本地已移植的功能（快捷键等） |
| Go 后端（server/storage/worker 等） | 语义合并，保留上游重构 + 本地功能点 |
| delete/modify（上游删了我们改过在旧路径的文件） | 取删除（上游结构迁移完成后的此类冲突应不再出现） |

本地已移植到新结构的改动清单（这些地方冲突时保留本地）：

- `src/storage/sqlite/item.go` — search 混合策略（ASCII 走 FTS，非 ASCII 走 LIKE）
- `src/worker/client.go` — 部分域名走固定代理
- `src/server/routes.go` — 空标题 item 补全、截图截断 280

## 测试与构建

```bash
# 全部测试（内存 sqlite 需要 fts5 tag；-short 跳过 postgres docker 测试）
go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" -short ./...

# 构建前端（上游 v2.8 起前端源码在 src/frontend，产物 bundle.js 不入库）
npm install
npm run build

# 编译 Windows 版
make windows_amd64_gui
```

注意：本机 Go 版本切换后，子 shell 里可能出现 stale `GOROOT`，报
`compile: version "go1.25.6" does not match go tool "go1.26.0"` 时执行：

```bash
export GOROOT=/home/yd/.gvm/gos/go1.26.0
```

## 验证编译产物来源

`make` 会把编译时 HEAD 写进二进制，反查：

```bash
go version -m out/windows_amd64_gui/yarr.exe | head
# -ldflags="-X 'main.GitHash=<short-hash>' ..." 对应即可定位 commit
```
