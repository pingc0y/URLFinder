package util

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/mode"
)

func TestGetProtocolReturnsEmptyWhenBothSchemesFail(t *testing.T) {
	got := GetProtocol("127.0.0.1:1")
	if got != "" {
		t.Fatalf("GetProtocol() = %q, want empty string for unreachable host", got)
	}
}

func TestRegexpMatchReturnsFalseForInvalidRegexp(t *testing.T) {
	if RegexpMatch("[", "https://example.com") {
		t.Fatal("RegexpMatch() returned true for invalid regexp")
	}
}

func TestDomainNameFilterReturnsEmptyForInvalidRegexp(t *testing.T) {
	oldD := cmd.D
	t.Cleanup(func() {
		cmd.D = oldD
	})

	cmd.D = "["

	got := domainNameFilter("https://example.com")
	if got != "" {
		t.Fatalf("domainNameFilter() = %q, want empty string for invalid regexp", got)
	}
}

func TestReadAllLimitedRejectsOversizedBody(t *testing.T) {
	_, err := readAllLimited(strings.NewReader("abcd"), 3)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("readAllLimited() error = %v, want ErrResponseTooLarge", err)
	}
}

func TestGetProtocolWithClientHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	}))
	defer server.Close()

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 20 * time.Millisecond}

	start := time.Now()
	got := getProtocol(parsed.Host, client)
	elapsed := time.Since(start)

	if got != "" {
		t.Fatalf("getProtocol() = %q, want empty string for timed out target", got)
	}
	if elapsed > 200*time.Millisecond {
		t.Fatalf("getProtocol() took %s, want it to honor client timeout", elapsed)
	}
}

func TestDel404KeepsResultsWhenNoDominantBodySize(t *testing.T) {
	links := []mode.Link{
		{Url: "https://example.com/a", Size: "1"},
		{Url: "https://example.com/b", Size: "20"},
		{Url: "https://example.com/c", Size: "300"},
	}

	got := Del404(links)

	if len(got) != len(links) {
		t.Fatalf("Del404() len = %d, want %d: %#v", len(got), len(links), got)
	}
}

func TestDel404DropsDominantBodySize(t *testing.T) {
	links := []mode.Link{
		{Url: "https://example.com/not-found-a", Size: "404"},
		{Url: "https://example.com/not-found-b", Size: "404"},
		{Url: "https://example.com/admin", Size: "20"},
	}

	got := Del404(links)

	if len(got) != 1 || got[0].Url != "https://example.com/admin" {
		t.Fatalf("Del404() = %#v, want only the non-dominant size result", got)
	}
}

func TestUpdateHTTPClientDoesNotDisableTLSVerification(t *testing.T) {
	client := updateHTTPClient()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("updateHTTPClient transport = %T, want *http.Transport", client.Transport)
	}
	if transport.TLSClientConfig != nil && transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("updateHTTPClient disables TLS certificate verification")
	}
}
