# 搜索排除语法（`a !b`）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让搜索支持 `a !b`（包含 a 且不包含 b），SQLite 与 Postgres 语义一致。

**Architecture:** 新增共享解析包 `searchquery` 把查询串解析为带 `Negated`
标记的词；SQLite 用 FTS5 `not in (...)` 子查询或负向 `LIKE` 生成排除条件，
Postgres 在 `to_tsquery` 中生成 `!word:*`。不改 schema、前端和协议。

**Tech Stack:** Go、SQLite FTS5（`mattn/go-sqlite3`）、PostgreSQL `to_tsquery`。

**Spec:** `docs/search-negation-design.md`

## Global Constraints

- 只支持词首单个 `!`；单独 `!` 忽略，`! b` 中的 `!` 也忽略。
- 词内部 `!`（`!!b`、`a!b`）不属于支持语法，不写跨后端断言。
- 跨后端一致性只承诺 ASCII 字母/数字词；标点、URL 词类沿用既有差异。
- 不新增数据库迁移，不改前端，不改排序。
- 不自动提交（用户未要求 commit）。

## Review Focus

- 只有排除词或只有 `!`：不能生成空 MATCH/空 tsquery，`!` 必须等价于无过滤。
- 多个排除词：必须逐词取反后 AND，不能合并成一个 MATCH（否则语义变成“同时包含才排除”）。
- title/content 为空或 NULL：负向 LIKE 不能把这类条目错误排除，需 `coalesce`。
- 前缀边界：`!appl` 必须排除 `apple`（与正向 `appl` 命中 `apple` 对称）。
- 大小写：`APPLE !BANANA` 与 `apple !banana` 结果一致。

---

### Task 1: 共享解析包 `searchquery`

**Files:**
- Create: `src/storage/searchquery/searchquery.go`
- Test: `src/storage/searchquery/searchquery_test.go`

**Interfaces:**
- Produces:
  - `type Term struct { Text string; Negated bool }`
  - `func Parse(query string) []Term`

- [ ] **Step 1: 写失败测试**

```go
package searchquery

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  []Term
	}{
		{"empty", "", nil},
		{"whitespace", " \t\n ", nil},
		{"single", "apple", []Term{{Text: "apple"}}},
		{"multi", "apple banana", []Term{{Text: "apple"}, {Text: "banana"}}},
		{"negated", "apple !banana", []Term{{Text: "apple"}, {Text: "banana", Negated: true}}},
		{"all negated", "!apple !banana", []Term{{Text: "apple", Negated: true}, {Text: "banana", Negated: true}}},
		{"lone bang", "!", nil},
		{"trailing lone bang", "apple !", []Term{{Text: "apple"}}},
		{"lone bang spaced", "! b", []Term{{Text: "b"}}},
		{"double bang", "!!b", []Term{{Text: "!b", Negated: true}}},
		{"middle bang", "a!b", []Term{{Text: "a!b"}}},
		{"tabs and newlines", "a\t!b\nc", []Term{{Text: "a"}, {Text: "b", Negated: true}, {Text: "c"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(tc.query)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Parse(%q) = %#v, want %#v", tc.query, got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" ./src/storage/searchquery/`
Expected: FAIL / build error: `undefined: Parse`

- [ ] **Step 3: 实现解析器**

`src/storage/searchquery/searchquery.go`：

```go
package searchquery

import "strings"

type Term struct {
	Text    string
	Negated bool
}

// Parse splits query on whitespace. A single leading '!' marks a negated
// term; a lone "!" is ignored.
func Parse(query string) []Term {
	terms := make([]Term, 0)
	for _, word := range strings.Fields(query) {
		if word == "!" {
			continue
		}
		if strings.HasPrefix(word, "!") {
			terms = append(terms, Term{Text: word[1:], Negated: true})
			continue
		}
		terms = append(terms, Term{Text: word})
	}
	return terms
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" ./src/storage/searchquery/`
Expected: PASS

---

### Task 2: SQLite 搜索谓词支持排除

**Files:**
- Modify: `src/storage/sqlite/item.go`（`listQueryPredicate` 搜索块，约 122-153 行；imports）
- Test: `src/storage/tests/item_search_negation_test.go`

**Interfaces:**
- Consumes: `searchquery.Parse`, `searchquery.Term`（Task 1）
- Produces: 无新导出接口；行为变更仅作用于 `ListItems(ItemFilter{Search: ...})`

- [ ] **Step 1: 写失败测试**

创建 `src/storage/tests/item_search_negation_test.go`。跨后端用例走
`dbtest`；标点与 href 用例直接 `storage.New`（仅 SQLite）。

数据集（日期递增，`newestFirst=false` 时按 date asc）：

```go
func TestSearchNegation(t *testing.T) {
	dbtest(t, func(t *testing.T, db storage.Storage) {
		feed := db.CreateFeed(model.CreateFeedParams{Title: "f", FeedLink: "http://f.xml"})
		base := time.Now()
		db.CreateItems([]model.Item{
			{GUID: "i1", FeedId: feed.Id, Title: "apple banana", Content: "first", Date: base.Add(1 * time.Minute)},
			{GUID: "i2", FeedId: feed.Id, Title: "apple", Content: "second", Date: base.Add(2 * time.Minute)},
			{GUID: "i3", FeedId: feed.Id, Title: "banana", Content: "third", Date: base.Add(3 * time.Minute)},
			{GUID: "i4", FeedId: feed.Id, Title: "apple cherry", Content: "banana", Date: base.Add(4 * time.Minute)},
			{GUID: "i5", FeedId: feed.Id, Title: "苹果", Content: "水果", Date: base.Add(5 * time.Minute)},
			{GUID: "i6", FeedId: feed.Id, Title: "苹果 香蕉", Content: "水果", Date: base.Add(6 * time.Minute)},
			{GUID: "i7", FeedId: feed.Id, Title: "apple pie", Content: "plain", Date: base.Add(7 * time.Minute)},
			{GUID: "i8", FeedId: feed.Id, Title: "apple cherry", Content: "nothing", Date: base.Add(8 * time.Minute)},
			{GUID: "i9", FeedId: feed.Id, Title: "", Content: "", Date: base.Add(9 * time.Minute)},
		})

		tests := []struct {
			name    string
			search  string
			wantIDs []string
		}{
			{"positive and exclude", "apple !banana", []string{"i2", "i7", "i8"}},
			{"exclude only", "!banana", []string{"i2", "i5", "i6", "i7", "i8", "i9"}},
			{"multiple excludes", "apple !banana !cherry", []string{"i2", "i7"}},
			{"prefix exclude", "apple !appl", []string{}},
			{"cjk", "苹果 !香蕉", []string{"i5"}},
			{"cjk exclude only", "!香蕉", []string{"i1", "i2", "i3", "i4", "i5", "i7", "i8", "i9"}},
			{"missing exclude term", "apple !不存在词", []string{"i1", "i2", "i4", "i7", "i8"}},
			{"lone bang", "!", []string{"i1", "i2", "i3", "i4", "i5", "i6", "i7", "i8", "i9"}},
			{"case insensitive", "APPLE !BANANA", []string{"i2", "i7", "i8"}},
			{"exclude only fts", "!apple", []string{"i3", "i5", "i6", "i9"}},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				search := tc.search
				have := getItemGuids(db.ListItems(model.ItemFilter{Search: &search}, 100, false, false))
				if !reflect.DeepEqual(have, tc.wantIDs) {
					t.Errorf("search %q: want %v, have %v", tc.search, tc.wantIDs, have)
				}
			})
		}
	})
}

func TestSearchNegationSQLiteOnly(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	feed := db.CreateFeed(model.CreateFeedParams{Title: "f", FeedLink: "http://f.xml"})
	base := time.Now()
	db.CreateItems([]model.Item{
		{GUID: "p1", FeedId: feed.Id, Title: "Release 8-1 is out", Content: "see 8-1 for details", Date: base.Add(1 * time.Minute)},
		{GUID: "p2", FeedId: feed.Id, Title: "Release 8-2 is out", Content: "see 8-2 for details", Date: base.Add(2 * time.Minute)},
		{GUID: "h1", FeedId: feed.Id, Title: "Link Only", Content: `<p>See the announcement.</p><a href="https://github.com/dbcli/mycli/issues/2048">link</a>`, Date: base.Add(3 * time.Minute)},
		{GUID: "h2", FeedId: feed.Id, Title: "Other", Content: `<p>nothing here</p>`, Date: base.Add(4 * time.Minute)},
	})

	tests := []struct {
		name    string
		search  string
		wantIDs []string
	}{
		{"punctuated exclude", "release !8-1", []string{"p2"}},
		{"punctuated exclude only", "!8-1", []string{"p2", "h1", "h2"}},
		{"href exclude", "!mycli", []string{"p1", "p2", "h2"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			search := tc.search
			have := getItemGuids(db.ListItems(model.ItemFilter{Search: &search}, 100, false, false))
			if !reflect.DeepEqual(have, tc.wantIDs) {
				t.Errorf("search %q: want %v, have %v", tc.search, tc.wantIDs, have)
			}
		})
	}
}
```

imports：`path/filepath`、`reflect`、`testing`、`time`、`storage`、`model`。

- [ ] **Step 2: 运行测试确认失败**

Run: `go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" -short ./src/storage/tests/ -run 'TestSearchNegation' -v`
Expected: FAIL（`apple !banana` 等用例返回错误结果）

- [ ] **Step 3: 改造 `src/storage/sqlite/item.go`**

import 增加：

```go
"github.com/nkanaev/yarr/src/storage/searchquery"
```

将 `if filter.Search != nil { ... }` 块替换为：

```go
	if filter.Search != nil {
		var (
			positiveFTS []string
			negativeFTS []string
			positiveLike []string
			negativeLike []string
		)
		for _, term := range searchquery.Parse(*filter.Search) {
			switch {
			case isFTSTerm(term.Text) && term.Negated:
				negativeFTS = append(negativeFTS, term.Text+"*")
			case isFTSTerm(term.Text):
				positiveFTS = append(positiveFTS, term.Text+"*")
			case term.Negated:
				negativeLike = append(negativeLike, term.Text)
			default:
				positiveLike = append(positiveLike, term.Text)
			}
		}
		if len(positiveFTS) > 0 {
			cond = append(
				cond,
				"i.id in (select rowid as id from search where search match :search)",
			)
			args = append(args, sql.Named("search", strings.Join(positiveFTS, " ")))
		}
		for i, term := range negativeFTS {
			name := fmt.Sprintf("neg_fts%d", i)
			cond = append(cond, "i.id not in (select rowid as id from search where search match :"+name+")")
			args = append(args, sql.Named(name, term))
		}
		for i, term := range positiveLike {
			name := fmt.Sprintf("like%d", i)
			cond = append(cond, fmt.Sprintf(
				"(i.title LIKE '%%' || :%s || '%%' OR searchable(i.content) LIKE '%%' || :%s || '%%')",
				name, name,
			))
			args = append(args, sql.Named(name, term))
		}
		for i, term := range negativeLike {
			name := fmt.Sprintf("notlike%d", i)
			cond = append(cond, fmt.Sprintf(
				"NOT (coalesce(i.title, '') LIKE '%%' || :%s || '%%' OR searchable(coalesce(i.content, '')) LIKE '%%' || :%s || '%%')",
				name, name,
			))
			args = append(args, sql.Named(name, term))
		}
	}
```

（`gofmt` 会对齐字段，以 gofmt 结果为准。）

- [ ] **Step 4: 运行 SQLite 测试确认通过**

Run: `go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" -short ./src/storage/tests/ -run 'TestSearch' -v`
Expected: PASS（含原有 `TestSearch*` 与新增用例）

---

### Task 3: Postgres 搜索谓词支持排除

**Files:**
- Modify: `src/storage/postgres/item.go`（`listQueryPredicate` 搜索块，约 116-127 行；imports）

**Interfaces:**
- Consumes: `searchquery.Parse`, `searchquery.Term`（Task 1）
- Produces: 与 SQLite 相同的 `a !b` 行为

- [ ] **Step 1: 改造 `src/storage/postgres/item.go`**

import 增加：

```go
"github.com/nkanaev/yarr/src/storage/searchquery"
```

将 `if filter.Search != nil { ... }` 块替换为：

```go
	if filter.Search != nil {
		terms := make([]string, 0)
		for _, term := range searchquery.Parse(*filter.Search) {
			query := term.Text + ":*"
			if term.Negated {
				query = "!" + query
			}
			terms = append(terms, query)
		}
		if len(terms) > 0 {
			cond = append(cond, fmt.Sprintf(
				"i.search @@ to_tsquery('simple', $%d)", next(),
			))
			args = append(args, strings.Join(terms, " & "))
		}
	}
```

- [ ] **Step 2: 运行测试**

Run（SQLite 必跑）：
`go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" -short ./src/storage/...`
Expected: PASS

Run（有 Docker 时，跨后端）：
`YARR_POSTGRES_TEST_IMAGE=postgres:16 go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" ./src/storage/tests/ -run 'TestSearchNegation' -v`
Expected: PASS

若无 Docker：用临时 `package main` 程序调用 `storage.New("<本地 PG URL>")`
+ `ListItems` 验证 `apple !banana`、`!banana`、`苹果 !香蕉` 三个查询，
验证后删除临时文件（不留在仓库）。

---

### Task 4: 全量验证与文档状态

**Files:**
- Modify: `docs/search-negation-design.md`（状态改为已实现）
- 可能微调：按 gofmt/测试结果修正 Task 2/3 的代码

- [ ] **Step 1: 格式化与静态检查**

Run: `gofmt -l src/storage/searchquery src/storage/sqlite/item.go src/storage/postgres/item.go src/storage/tests/item_search_negation_test.go`
Expected: 无输出

Run: `go vet -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" ./src/storage/...`
Expected: 无错误

- [ ] **Step 2: 全量测试**

Run: `go test -tags "sqlite_foreign_keys sqlite_json sqlite_fts5" -short ./...`
Expected: PASS

- [ ] **Step 3: 更新设计文档状态**

将 `docs/search-negation-design.md` 头部状态改为：

```markdown
- 状态：已实现（2026-11-06）
```

- [ ] **Step 4: 汇报验证结果**

列出实际执行的命令与结果；若 Postgres 未实测（无 Docker/本地 PG），明确说明
未验证项。
