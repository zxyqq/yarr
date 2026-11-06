# 搜索排除语法（`a !b`）设计方案

- 日期：2026-11-06
- 状态：已实现（2026-11-06）
- 方案：A（最小跨后端支持）

## 1. 背景与现状

搜索串由前端原样透传：`App.vue:990` 写入 `query.search`，`routes.go:375-376`
写入 `ItemFilter.Search`，后端按空白切词后全部按**正向**处理。

- SQLite（默认后端）：ASCII 字母/数字词走 FTS5 前缀匹配，其余词走
  `LIKE` 子串匹配（`src/storage/sqlite/item.go:91-105,122-153`）。
  FTS 索引为 `fts5(title, content, content='items', tokenize='unicode61')`，
  content 使用 `searchable()`，包含 href/src 的值（`src/storage/sqlite/migration.go:437-466`）。
- PostgreSQL：所有词拼成 `to_tsquery('simple', 'a:* & b:*')`
  （`src/storage/postgres/item.go:116-127`）。`!b` 恰好被 tsquery 当作
  NOT，属于未定义、未测试的偶然行为。
- 结果：`a !b` 在默认 SQLite 上不会排除 `b`，而是把 `!b` 当作字面量
  `LIKE '%!b%'`；跨后端行为不一致。

## 2. 目标

1. `a !b` 语义为“包含 a 且不包含 b”。
2. 支持只有排除项的查询（如 `!b`）。
3. 排除词复用该词作为正向词时的匹配方式：ASCII 字母/数字词的
   `a !b` 在 SQLite 与 Postgres 上结果一致；标点、CJK、URL 等词类
   沿用各后端现有的正向匹配差异，本次不新增也不消除这些差异。
4. 现有纯正向查询的行为、排序、性能不退化。
5. 不新增数据库迁移，不改前端协议。

## 3. 非目标

- 不支持 `-b`、`OR`、括号、引号短语。
- 不定义词内部 `!`（含 `!!b`、`a!b`）的跨后端语义：SQLite 会当作
  字面量走 LIKE，Postgres 可能解释为 tsquery 的 NOT 或直接报语法错误
  （已验证 `a!b:*` 报错、`!!b:*` 等于双重否定）。这类输入不属于支持
  的语法，只保证不会比现状更差。
- 不修复 Postgres 既有的标点词、URL 内部路径词匹配问题（独立问题）。
- 不做搜索排名/相关性调整。
- 本期不改 UI 文案（可选后续在占位符或文档中提示语法）。

## 4. 语法与语义

对 `filter.Search` 使用 `strings.Fields`（Unicode 空白）切词，逐词解析：

- 以 `!` 开头且长度大于 1：排除词，去掉恰好一个 `!`。
- 其余：正向词。
- 去掉 `!` 后为空的词（即单独的 `!`）忽略。
- 正向词之间为 AND；所有排除词都不得命中；正负条件整体 AND。
- 若解析后没有任何有效词，则视为空搜索（返回全部，不加条件）。
- `!` 只在词首生效；词中 `!` 不属于支持语法：SQLite 按字面量 LIKE，
  Postgres 可能报 tsquery 语法错误。
- 只剥离一个 `!`：`!!b` 解析为“排除文本 `!b`”，但 Postgres 侧
  `!!b:*` 是双重否定，因此该输入不纳入跨后端约定。
- `! b`（`!` 后有空格）中 `!` 被忽略，`b` 是正向词。
- 每个词的匹配方式与其作为正向词时完全一致：
  - FTS 词（ASCII 字母/数字）：token 前缀匹配；
  - 非 FTS 词（标点、CJK 等）：SQLite 用 `LIKE` 子串匹配；
    Postgres 仍走 tsquery，属既有差异。
- 大小写不敏感，与现状一致。

示例：

| 查询 | 含义 |
|---|---|
| `apple !banana` | 含 apple*，且不含 banana* |
| `!banana` | 不含 banana* |
| `apple !8-1` | 含 apple*，且标题/正文/链接中不含子串 `8-1`（SQLite 语义） |
| `!香蕉` | 不含香蕉（SQLite 子串匹配 / Postgres token 匹配） |
| `apple !` | 等价于 `apple` |
| `!` | 等价于空搜索 |

## 5. 设计

### 5.1 新增共享解析器 `src/storage/searchquery`

新建小包，供两个后端共用，避免在 `sqlite`/`postgres` 各写一份解析逻辑，
也避免 `sqlite` 反向依赖 `storage` 造成循环引用。

```go
package searchquery

type Term struct {
    Text    string
    Negated bool
}

// Parse 按空白切词并解析前导 '!'。
func Parse(query string) []Term
```

行为由第 4 节定义；单独提供单元测试。

### 5.2 SQLite predicate 改造

改造 `src/storage/sqlite/item.go` 的 `listQueryPredicate` 搜索块：

1. 用 `searchquery.Parse` 得到词列表。
2. 正向 FTS 词：保持现状，合并为一个条件
   `i.id in (select rowid as id from search where search match :search)`，
   参数为各词加 `*` 后以空格连接（FTS5 隐式 AND）。
3. 正向非 FTS 词：保持现状
   `(i.title like '%' || :likeN || '%' or searchable(i.content) like '%' || :likeN || '%')`。
4. 排除 FTS 词：**每个词一个条件**，不能合并（合并会变成“同时包含才排除”）：
   `i.id not in (select rowid as id from search where search match :negN)`
   参数为该词加 `*`。
5. 排除非 FTS 词：
   `not (i.title like '%' || :likeN || '%' or searchable(i.content) like '%' || :likeN || '%')`。
6. 所有条件以 `and` 连接；无有效词时不加条件。

`not in` 子查询与现有 `in` 子查询对称，仍走 FTS 索引。

已验证：FTS5 中 `a* NOT b*` 只在有正向词时可用，单独 `NOT b*` 是语法
错误，因此排除统一用 SQL 层 `not in`，不拼进 FTS 查询串。

### 5.3 PostgreSQL predicate 改造

改造 `src/storage/postgres/item.go:116-127`：

- 用 `searchquery.Parse` 得到词列表。
- 正向词生成 `word:*`，排除词生成 `!word:*`，以 ` & ` 连接。
- 无有效词时不加条件（避免空 tsquery 报错）。
- 保留现有标点处理方式不变；本次只把 `!` 行为显式化并纳入测试。

生成的查询示例：`apple !banana` → `to_tsquery('simple', 'apple:* & !banana:*')`。

### 5.4 覆盖字段

排除条件必须覆盖与正向一致的字段：

- SQLite：FTS 索引覆盖 title + `searchable(content)`（含 href/src）；
  LIKE 排除同时查 `title` 与 `searchable(content)`。
- Postgres：`search` tsvector 在写入时由 title + `SearchableText(content)`
  生成（`src/storage/postgres/item.go:50,263`）。

## 6. 行为矩阵

| 查询 | SQLite | Postgres |
|---|---|---|
| `apple` | FTS 前缀 | tsquery 前缀 |
| `8-1` | LIKE 子串 | 解析为短语 `'8':* <-> '-1':*`（既有行为） |
| `apple !banana` | apple* 且 not banana* | `apple:* & !banana:*` |
| `!banana` | not banana* | `!banana:*` |
| `!香蕉` | not LIKE `%香蕉%` | `!香蕉:*` |
| `!8-1` | not LIKE `%8-1%` | `!('8':* <-> '-1':*)`（非子串语义，既有差异） |
| `!` | 无过滤 | 无过滤 |
| `!!b` / `a!b` | 字面量 LIKE | 双重否定 / 语法错误（不支持） |

跨后端一致性只承诺表中的 ASCII 词查询（含 `apple !banana`、`!banana`）；
带标点、URL 等词类沿用各后端既有差异。

## 7. 测试计划

### 7.1 解析器单元测试

`src/storage/searchquery/searchquery_test.go`：

- 空串、纯空白、单正向、多正向
- `!a`、`a !b`、`!a !b`
- `!` 单独、`!` 后空格、`!!a`（排除 `!a`）
- 词中 `!`（`a!b` 不排除）
- 制表符/换行分隔

### 7.2 存储集成测试

新增 `src/storage/tests/item_search_negation_test.go`，分两层：

**A. 跨后端用例（`dbtest`，SQLite 必跑；设置
`YARR_POSTGRES_TEST_IMAGE` 时同跑 Postgres）**

- `apple !banana` 命中含 apple 但不含 banana 的项
- `!banana` 只返回不含 banana 的项
- `apple !banana !cherry` 多排除
- 排除词命中 title / content 两种位置
- 前缀语义：`apple !appl` 排除所有 apple 项
- CJK：`苹果 !香蕉`、`!香蕉`
- `apple !不存在词` 不误伤
- `!` 返回全部

**B. 仅 SQLite 用例（直接 `storage.New`，与现有
`item_search_punct_test.go` 同模式）**

- 非 FTS 排除词：`!8-1` 按子串排除
- href/src 内的排除词：`!mycli` 能排除链接 URL 中的词
  （Postgres 现有正向搜索本就匹配不到 URL 内部路径，见第 8 节）

**C. 回归**

- 现有 `TestSearch`、`TestSearchPunctuatedTerm`、`TestSearchChinese`
  不改；其中 href 正向用例在 Postgres 下不通过属既有问题，不在本方案
  修复范围（新增跨后端用例刻意避开该场景）。

### 7.3 验证命令

```bash
go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" \
  -short ./src/storage/...

YARR_POSTGRES_TEST_IMAGE=postgres:16 \
go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" \
  ./src/storage/tests/
```

## 8. 性能与兼容性

- 性能：排除 FTS 词使用 `not in (select rowid ... match ...)`，与现有
  正向 `in` 同一代价级别；排除 LIKE 词与现有正向 LIKE 词同级。
- 无 schema/migration 变更，无索引重建。
- 行为变化：查询中带前导 `!` 的用户会从“字面量/偶然 NOT”变为明确的
  排除语义，这是本方案目标；`!` 在词中不受影响。
- 已知限制：Postgres 对含特殊符号的词仍可能触发 tsquery 语法差异
  （如 `a!b` 报错、`!!b` 变双重否定）；本次不扩大处理范围，只对
  词首单个 `!` 的约定语法负责。
- 既有差异（不属本方案）：Postgres 无法匹配 URL 内部路径词
  （`github.com/dbcli/mycli/...` 搜索 `mycli` 在 PG 上无结果，SQLite
  的 `searchable()` 可以）；因此现有 `TestSearch` 的 href 正向用例在
  Postgres 下预期失败，新增跨后端用例已避开。

### 验证记录

设计前已用实际代码/数据库验证（PostgreSQL 16.15、SQLite FTS5）：

- SQLite：`i.id not in (select rowid from search where search match 'b*')`
  正确排除；单独 `NOT b*` 拼进 FTS 查询会语法错误，故采用 SQL 层
  `not in`。
- Postgres：`apple:* & !banana:*`、`!banana:*`、`!香蕉:*` 均可正常
  匹配；`!8-1:*` 解析为非子串的短语否定；`!!b:*` 为双重否定；
  `a!b:*` 直接报语法错误。

## 9. 改动文件与工作量

| 文件 | 改动 |
|---|---|
| `src/storage/searchquery/searchquery.go` | 新增解析器（约 40 行） |
| `src/storage/searchquery/searchquery_test.go` | 单元测试（约 60 行） |
| `src/storage/sqlite/item.go` | 搜索块改造（约 25 行） |
| `src/storage/postgres/item.go` | 搜索块改造（约 10 行） |
| `src/storage/tests/item_search_negation_test.go` | 集成测试（约 120 行） |
| `docs/search-negation-design.md` | 本文档 |

预计 0.5 人日（含测试），前端零改动。

## 10. 决策记录（实现前可推翻）

- 只支持 `!` 前缀，不支持 `-`，避免与合法搜索词冲突。
- 只剥离一个前导 `!`；单独 `!` 忽略。
- 词内部 `!`（`!!b`、`a!b`）不纳入支持语法，只记录各后端现状。
- 跨后端一致性只承诺 ASCII 字母/数字词；标点、URL 词类沿用既有差异。
- 解析后无有效词时不过滤。
- 共享解析器放新包 `searchquery`，不放进 `model`。
- 测试分两层：ASCII/CJK 否定进 `dbtest`，标点与 href 否定仅 SQLite。
- UI 语法提示列为后续可选项，不阻塞本期。
