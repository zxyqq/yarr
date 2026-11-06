package tests

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/nkanaev/yarr/src/storage"
	"github.com/nkanaev/yarr/src/storage/model"
)

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
