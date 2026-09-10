# Windows 本机构建（windows_amd64_gui）

**日期**: 2026-09-11
**目标**: `make windows_amd64_gui`，产物为 `out/windows_amd64_gui/yarr.exe`（GUI 模式，无控制台窗口）

## 前置依赖

| 工具 | 用途 | 安装方式（scoop） |
|---|---|---|
| Go | 编译主程序 | `scoop install go` |
| Node.js + npm | 打包前端资源（`assets` 目标） | `scoop install nodejs-lts` |
| make | 执行 Makefile | `scoop install make` |
| MinGW-w64 | 提供 `gcc`（CGO 编译 sqlite）和 `windres`（编译版本/图标资源） | `scoop install mingw` |

> 原 Makefile 使用 `zig cc` 交叉编译，Windows 本机构建时已改为直接调用 `gcc`。

### 注意：scoop 的 mingw 包不生成 shim

scoop 的 `mingw` 包通过 `env_add_path` 把 `D:\Apps\scoop\apps\mingw\current\bin` 直接写入用户 PATH 注册表，
**不会**在 `scoop\shims` 下生成 `gcc.exe`/`windres.exe` 链接。

因此安装完成后，**已开启的终端读不到新 PATH**，需要重开终端窗口；或临时刷新：

```powershell
$env:Path = [Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User')
```

验证：

```powershell
gcc --version      # 16.2.0 (MinGW-Builds)
windres --version  # GNU Binutils 2.47
```

## node_modules 平台问题

如果 `node_modules` 是在 WSL/Linux 下安装的，esbuild 的平台二进制为 `@esbuild/linux-x64`，
在 Windows 上构建前端资源会报错：

```
Error: You installed esbuild for another platform than the one you're currently using.
```

解决：删除后在 Windows 下重装。

```powershell
Remove-Item -Recurse -Force node_modules
npm install
```

> 若 npm 提示 esbuild 的 postinstall 脚本未被 allowScripts 覆盖，可忽略——
> esbuild 的平台二进制通过 optionalDependencies（`@esbuild/win32-x64`）安装，不依赖 postinstall。

## 构建命令

```powershell
make windows_amd64_gui
```

等价的手动步骤：

```powershell
# 1. 前端资源（esbuild 压缩）
npm run build -- --minify

# 2. 版本/图标资源（syso 已存在时 make 会跳过 windres）
./etc/windows_versioninfo.sh -version "bleeding" -outfile src/platform/versioninfo.rc
windres -i src/platform/versioninfo.rc -O coff -o src/platform/versioninfo.syso

# 3. 编译
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build `
  -tags "sqlite_foreign_keys sqlite_json sqlite_fts5 gui" `
  -ldflags "-s -w -X 'main.Version=bleeding' -H windowsgui" `
  -o out/windows_amd64_gui/yarr.exe ./cmd/yarr
```

关键参数说明：

| 参数 | 作用 |
|---|---|
| `-tags gui` | 启用 GUI 入口（systray 等），对应 `src/platform` 下的 Windows 实现 |
| `-H windowsgui` | PE 子系统设为 GUI，运行时不弹出控制台窗口 |
| `-tags sqlite_*` | 启用 sqlite 外键、JSON1、FTS5 |
| `CGO_ENABLED=1` | sqlite 驱动需要 CGO，由 mingw 的 gcc 编译 |
| `versioninfo.syso` | go build 自动链接同目录下的 `.syso`，嵌入版本号和图标 |

## 产物

```
out/windows_amd64_gui/yarr.exe   # 约 17 MB，双击即可运行
```
