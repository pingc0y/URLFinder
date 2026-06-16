package crawler

import (
	"testing"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/result"
)

func TestUrlFilterDropsNonFetchableReferences(t *testing.T) {
	matches := [][]string{
		{"", "data:text/plain"},
		{"", "mailto:security@example.com"},
		{"", "javascript:alert(1)"},
		{"", "#section"},
		{"", "tel:123456"},
		{"", "https://target.test/api"},
		{"", "/api/users"},
	}

	got := urlFilter(matches)

	for i, want := range []string{"", "", "", "", "", "https://target.test/api", "/api/users"} {
		if got[i][0] != want {
			t.Fatalf("urlFilter()[%d] = %q, want %q", i, got[i][0], want)
		}
	}
}

func TestUrlFindDoesNotJoinNonFetchableReferencesToHost(t *testing.T) {
	oldB, oldD, oldM := cmd.B, cmd.D, cmd.M
	oldUrlSteps := config.UrlSteps
	t.Cleanup(func() {
		cmd.B, cmd.D, cmd.M = oldB, oldD, oldM
		config.UrlSteps = oldUrlSteps
	})

	cmd.B = ""
	cmd.D = ""
	cmd.M = 1
	config.UrlSteps = 1
	Initialization()

	html := `<a href="data:text/plain">bad</a><a href="#section">anchor</a><a href="/api/users">good</a>`
	urlFind(html, "example.com", "https", "/", "https://example.com", 1)

	if len(result.ResultUrl) != 1 {
		t.Fatalf("len(result.ResultUrl) = %d, want 1: %#v", len(result.ResultUrl), result.ResultUrl)
	}
	if result.ResultUrl[0].Url != "https://example.com/api/users" {
		t.Fatalf("result url = %q, want %q", result.ResultUrl[0].Url, "https://example.com/api/users")
	}
}
