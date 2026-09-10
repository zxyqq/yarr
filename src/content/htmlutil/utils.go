package htmlutil

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

var whitespaceRegex = regexp.MustCompile(`[\s]+`)

var searchableAttrRe = regexp.MustCompile(`(?i)\b(?:href|src)\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+))`)

func HTML(node *html.Node) string {
	writer := strings.Builder{}
	html.Render(&writer, node)
	return writer.String()
}

func InnerHTML(node *html.Node) string {
	writer := strings.Builder{}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		html.Render(&writer, c)
	}
	return writer.String()
}

func Attr(node *html.Node, key string) string {
	for _, a := range node.Attr {
		if strings.EqualFold(a.Key, key) {
			return a.Val
		}
	}
	return ""
}

func Text(node *html.Node) string {
	text := make([]string, 0)
	isTextNode := func(n *html.Node) bool {
		return n.Type == html.TextNode
	}
	for _, n := range FindNodes(node, isTextNode) {
		text = append(text, strings.TrimSpace(n.Data))
	}
	return strings.Join(text, " ")
}

func ExtractText(content string) string {
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	buffer := bytes.Buffer{}
	for {
		token := tokenizer.Next()
		if token == html.ErrorToken {
			break
		}
		if token == html.TextToken {
			buffer.WriteString(html.UnescapeString(string(tokenizer.Text())))
		}
	}
	text := buffer.String()
	text = strings.TrimSpace(text)
	text = whitespaceRegex.ReplaceAllLiteralString(text, " ")
	return text
}

// SearchableText returns the visible text of the content plus the values of
// href/src attributes, so that URLs embedded in HTML stay searchable, while
// other attributes (e.g. class names) are ignored.
func SearchableText(content string) string {
	text := ExtractText(content)
	seen := make(map[string]bool)
	for _, match := range searchableAttrRe.FindAllStringSubmatch(content, -1) {
		val := match[1]
		if val == "" {
			val = match[2]
		}
		if val == "" {
			val = match[3]
		}
		if val == "" {
			continue
		}
		val = html.UnescapeString(val)
		if seen[val] {
			continue
		}
		seen[val] = true
		text += " " + val
	}
	return text
}

func TruncateText(input string, size int) string {
	runes := []rune(input)
	if len(runes) <= size {
		return input
	}
	for i := size - 1; i > 0; i-- {
		if unicode.IsSpace(runes[i]) {
			return string(runes[:i]) + " ..."
		}
	}
	return input
}
