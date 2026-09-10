# 前端快捷键移植计划（master → master-v2.9）

把 v2.7 旧前端（`master:src/assets/javascripts/key.js` 之外的本地增强）移植到 v2.9 新结构（`src/frontend/js/`）。移植完成后删除本文件。

## 现状对照

v2.9 官方 `key.ts` bindings:

```
o/i/r/R/s / j k l h f b q ? 1 2 3
```

旧 master 增强版 `key.js` bindings（要移植的增量）:

| 键 | 功能 | v2.9 现状 | 移植工作量 |
|---|---|---|---|
| `w` | toggleFeedList 隐藏/显示 feed 列表 | 无（仅有拖动调宽） | 中：需新增 state |
| `a` | markAllRead 全部已读（不只 unread 筛选） | 无（R 只在 unread filter 下生效） | 小 |
| `A` | showAllFeeds（回到"全部文章"视图） | 无 | 小 |
| `x` / `c` | toggleCurrentFolder 展开/收起**当前**文件夹 | 有 toggleFolderExpanded，无按键 | 小：补按键 |
| `C` / `X` | toggleAllFolders 全部收起/展开 | 无 | 中：新增批量函数 |
| `[` `]` | resizeFeedList -50/+50 | 无按键（拖动有），函数有 | 小 |
| `{` `}` | resizeItemList -50/+50 | 同上 | 小 |
| `e` | scrollBackward 重绑 | 上游用 `b`，可选保留 `e` | 小 |
| `q` | showUnread 重绑 | 上游 `q`=closeItem，注意冲突 | 需决策 |
| `u` | toggleItemRead | 上游 `r`，作为别名无害 | 小 |
| `v` | openItemLink 别名 | 无 | 小 |

## 目标文件与改动点

### 1. `src/frontend/js/key.ts`

- 追加 `shortcutFunctions`：
  - `toggleFeedList()` → `vm.toggleFeedList()`
  - `showAllFeeds()` → `vm.feedSelected = null; vm.filterSelected = ''`
  - `markAllReadAll()`（a 键，无视 filter） → `vm.markItemsRead()` 的全量版（下一步 App.vue 提供）
  - `toggleCurrentFolder()` → `vm.toggleCurrentFolder()`
  - `toggleAllFolders(expand: bool)` → `vm.toggleAllFolders(expand)`
  - `resizeFeedList(dx)` → `vm.resizeFeedList(vm.feedListWidth + dx)`
  - `resizeItemList(dx)` → `vm.resizeItemList(vm.itemListWidth + dx)`
- `keybindings` 表新增：
  ```
  w: toggleFeedList, a: markAllReadAll, A: showAllFeeds,
  x/c: toggleCurrentFolder(expand/收起按当前状态),
  C: collapseAll, X: expandAll,
  "[": resizeFeedList(-50), "]": resizeFeedList(+50),
  "{": resizeItemList(-50), "}": resizeItemList(+50),
  u: toggleItemRead, e: scrollBackward（别名）, v: openItemLink（别名）
  ```
- `q` 保持上游语义（closeItem）不改，避免和 1/2/3 修改筛选冲突；
  旧 `q=showUnread` 不移植（与一键"显示未读"语义重复）。

### 2. `src/frontend/js/pages/App.vue`

- 新增状态：
  ```ts
  feedListVisible: s.feed_list_visible !== false,
  ```
  并持久化到 settings（`api.settings.update({ feed_list_visible: ... })`，
  与 `feedListWidth` 的 debounce 写法一致）。
- 模板 feed-list 容器 `<div id="col-feed-list">` 加 `v-show="feedListVisible"`。
- methods：
  - `toggleFeedList()` — 翻转并持久化
  - `toggleCurrentFolder()` — 找到 `folderSelected`/`feedSelected` 所属 folder,
    调 `toggleFolderExpanded(folder)`
  - `toggleAllFolders(expand: boolean)` — 遍历 `this.folders`，
    批量 `api.folders.update(id, {is_expanded})`（与 toggleFolderExpanded 相同调用）
- `markItemsRead` 泛化：v2.9 现版本要求 `filterSelected == 'unread'`，
  增加一个 `markAllItemsRead()` 供 `a` 键使用。

### 3. 帮助弹窗

`showModal = "shortcuts"` 的键盘帮助列表（模板在 App.vue shortcuts 区块）
同步加上新键说明，与 keybindings 保持一致（上游注释有硬性要求）。

## 验收

```bash
npm install && npm run build          # 产出 src/assets/static/bundle.js
go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" ./src/server
make windows_amd64_gui
go version -m out/windows_amd64_gui/yarr.exe | grep GitHash
```

手工核对表：w 隐藏/恢复、a 全部已读、A 回全视图、x/c/C/X 行为、四向宽度键、
`?` 帮助内容与实际一致。
