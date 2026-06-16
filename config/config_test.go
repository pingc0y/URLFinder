package config

import (
	"strings"
	"testing"

	"github.com/pingc0y/URLFinder/mode"
)

func TestValidateRegexConfigAcceptsDefaultPatterns(t *testing.T) {
	conf := mode.Config{
		JsFind:   JsFind,
		UrlFind:  UrlFind,
		JsFiler:  JsFiler,
		UrlFiler: UrlFiler,
		InfoFind: map[string][]string{
			"Phone":  Phone,
			"Email":  Email,
			"IDcard": IDcard,
			"Jwt":    Jwt,
			"Other":  Other,
		},
	}

	if err := ValidateRegexConfig(conf); err != nil {
		t.Fatalf("ValidateRegexConfig(defaults) error = %v, want nil", err)
	}
}

func TestValidateRegexConfigRejectsInvalidPatterns(t *testing.T) {
	conf := mode.Config{
		JsFind:  []string{"["},
		UrlFind: []string{"https://example\\.com"},
		InfoFind: map[string][]string{
			"Email": {"[a-z]+"},
		},
	}

	err := ValidateRegexConfig(conf)
	if err == nil {
		t.Fatal("ValidateRegexConfig() error = nil, want invalid regexp error")
	}
	if !strings.Contains(err.Error(), "jsFind[0]") {
		t.Fatalf("ValidateRegexConfig() error = %q, want field name jsFind[0]", err)
	}
}
