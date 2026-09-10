package htmlutil

import (
	"strings"
	"testing"
)

func TestSearchableText(t *testing.T) {
	content := `<div class="secret-class">hello</div>` +
		`<a href="https://example.com/dbcli/mycli">a</a>` +
		`<img src='/img/logo.png'>` +
		`<script src=app.js></script>`

	have := SearchableText(content)
	for _, want := range []string{"hello", "https://example.com/dbcli/mycli", "/img/logo.png", "app.js"} {
		if !strings.Contains(have, want) {
			t.Errorf("SearchableText() = %q, missing %q", have, want)
		}
	}
	if strings.Contains(have, "secret-class") {
		t.Errorf("SearchableText() = %q, should not include class names", have)
	}
}

func TestExtractText(t *testing.T) {
	testcases := [][2]string{
		{"hello", "<div>hello</div>"},
		{"hello world", "<div>hello</div> world"},
		{"helloworld", "<div>hello</div>world"},
		{"hello world", "hello <div>world</div>"},
		{"helloworld", "hello<div>world</div>"},
		{"hello world!", "hello <div>world</div>!"},
		{"hello world !", "hello <div>   world\r\n </div>!"},
	}
	for _, testcase := range testcases {
		want := testcase[0]
		base := testcase[1]
		have := ExtractText(base)
		if want != have {
			t.Logf("base: %#v\n", base)
			t.Logf("want: %#v\n", want)
			t.Logf("have: %#v\n", have)
			t.Fail()
		}
	}
}

func TestTruncateText(t *testing.T) {
	input := "Lorem ipsum — классический текст-«рыба»"

	size := 30
	want := "Lorem ipsum — классический ..."
	have := TruncateText(input, size)
	if want != have {
		t.Errorf("\nsize: %d\nwant: %#v\nhave: %#v", size, want, have)
	}

	size = 1000
	want = input
	have = TruncateText(input, size)
	if want != have {
		t.Errorf("\nsize: %d\nwant: %#v\nhave: %#v", size, want, have)
	}
}
