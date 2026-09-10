# v-show 与 Bootstrap d-flex 冲突导致元素无法隐藏

**日期**: 2026-06-13
**涉及文件**: `src/assets/index.html`, `src/assets/javascripts/app.js`, `src/assets/javascripts/key.js`

## 需求

按 `w` 键切换显示/隐藏左侧 feed 导航面板。

## 实现方式

| 文件 | 改动 |
|---|---|
| `app.js` | 新增 `feedListVisible: true` 数据属性 + `toggleFeedList()` 方法 |
| `key.js` | 新增 `toggleFeedList` 快捷键函数，绑定到 `w` / `KeyW` |
| `index.html` | 在 `#col-feed-list` 上添加 `v-show="feedListVisible"` |

## 问题现象

按 `w` 键没有任何效果，在浏览器控制台执行 `vm.feedListVisible = false` 也无法隐藏面板。

## 根因

`v-show` 和 Bootstrap 的 `d-flex` 存在 CSS 优先级冲突：

```
v-show 隐藏方式:  内联样式 display: none          （无 !important）
d-flex 类样式:    .d-flex { display: flex !important }
```

CSS 优先级规则：`!important` > 内联样式（无 `!important`）。

所以 `d-flex` 的 `display: flex !important` 覆盖了 `v-show` 设置的内联 `display: none`，元素永远无法被隐藏。

## 修复

将 `v-show` 改为 `v-if`：

```html
<!-- 修复前 -->
<div id="col-feed-list" class="... d-flex ..." v-show="feedListVisible">

<!-- 修复后 -->
<div id="col-feed-list" class="... d-flex ..." v-if="feedListVisible">
```

`v-if` 直接从 DOM 中移除/插入元素，不依赖 `display` 属性，完全绕开了 CSS 优先级问题。

## 经验总结

> 当元素同时使用了 Bootstrap 的 `d-flex`、`d-block`、`d-none` 等 display 工具类时，
> 不要用 `v-show` 控制显隐，应使用 `v-if`。
> 因为这些 Bootstrap 类都带有 `!important`，会覆盖 `v-show` 的内联 `display: none`。
