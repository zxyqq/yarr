package tests

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/storage/model"
)

func TestSearchChinese(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}

	feed := db.CreateFeed(model.CreateFeedParams{Title: "test", FeedLink: "http://test.com/feed.xml"})
	now := time.Now()

	// The exact title from the user's example
	title := "Prompt、Agent、Skill、MCP 到底是啥？用一家饭馆的后厨给你讲透 - it排球君 - 博客园"
	ok := db.CreateItems([]model.Item{
		{GUID: "item1", FeedId: feed.Id, Title: title, Content: "测试内容", Date: now},
	})
	if !ok {
		t.Fatal("failed to create item")
	}

	tests := []struct {
		name    string
		search  string
		wantIDs []string
	}{
		// English words - should work via FTS
		{"english: prompt", "prompt", []string{"item1"}},
		{"english: agent", "agent", []string{"item1"}},
		{"english: skill", "skill", []string{"item1"}},
		{"english: mcp", "mcp", []string{"item1"}},
		{"english: it", "it", []string{"item1"}},
		// Chinese words - should work via LIKE
		{"chinese: 饭馆", "饭馆", []string{"item1"}},
		{"chinese: 后厨", "后厨", []string{"item1"}},
		{"chinese: 到底", "到底", []string{"item1"}},
		{"chinese: 博客园", "博客园", []string{"item1"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			search := tc.search
			have := getItemGuids(db.ListItems(model.ItemFilter{Search: &search}, 10, false, false))
			if !reflect.DeepEqual(have, tc.wantIDs) {
				t.Errorf("search %q: want %v, have %v", tc.search, tc.wantIDs, have)
			}
		})
	}
}
