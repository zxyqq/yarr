package tests

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/storage/model"
)

func TestSearchPunctuatedTerm(t *testing.T) {
	db, err := storage.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	feed := db.CreateFeed(model.CreateFeedParams{Title: "test", FeedLink: "http://test.com/feed.xml"})
	ok := db.CreateItems([]model.Item{
		{
			GUID:    "item1",
			FeedId:  feed.Id,
			Title:   "Release 8-1 is out",
			Content: "see 8-1 for details",
			Date:    time.Now(),
		},
	})
	if !ok {
		t.Fatal("failed to create item")
	}

	tests := []struct {
		name    string
		search  string
		wantIDs []string
	}{
		{"hyphen", "8-1", []string{"item1"}},
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
