package searchquery

import "strings"

type Term struct {
	Text    string
	Negated bool
}

// Parse splits query on whitespace. A single leading '!' marks a negated
// term; a lone "!" is ignored.
func Parse(query string) []Term {
	var terms []Term
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
