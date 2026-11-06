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
