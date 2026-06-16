package util

import (
	"errors"
	"strings"
	"testing"

	"github.com/pingc0y/URLFinder/cmd"
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
