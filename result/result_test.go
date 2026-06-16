package result

import (
	stdhtml "html"
	"strings"
	"testing"

	"github.com/pingc0y/URLFinder/mode"
)

func TestOutHtmlStringEscapesLinkFields(t *testing.T) {
	maliciousURL := `https://example.com/"><script>alert(1)</script>`
	maliciousTitle := `<img src=x onerror=alert(1)>`
	link := mode.Link{
		Url:      maliciousURL,
		Status:   "200",
		Size:     "123",
		Title:    maliciousTitle,
		Redirect: maliciousURL,
		Source:   maliciousURL,
	}

	got := outHtmlString(link)

	for _, raw := range []string{maliciousURL, maliciousTitle} {
		if strings.Contains(got, raw) {
			t.Fatalf("outHtmlString() contained unescaped value %q in %q", raw, got)
		}
		if !strings.Contains(got, stdhtml.EscapeString(raw)) {
			t.Fatalf("outHtmlString() did not contain escaped value %q in %q", raw, got)
		}
	}
}

func TestOutHtmlInfoStringEscapesInfoFields(t *testing.T) {
	rawValue := `<script>alert(1)</script>`
	rawSource := `https://example.com/"><script>alert(1)</script>`

	got := outHtmlInfoString("Other", rawValue, rawSource)

	for _, raw := range []string{rawValue, rawSource} {
		if strings.Contains(got, raw) {
			t.Fatalf("outHtmlInfoString() contained unescaped value %q in %q", raw, got)
		}
		if !strings.Contains(got, stdhtml.EscapeString(raw)) {
			t.Fatalf("outHtmlInfoString() did not contain escaped value %q in %q", raw, got)
		}
	}
}
