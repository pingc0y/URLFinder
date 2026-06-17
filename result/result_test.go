package result

import (
	"encoding/csv"
	stdhtml "html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pingc0y/URLFinder/cmd"
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

func TestOutFileHtmlHandlesOpenErrorWithoutPanic(t *testing.T) {
	oldU, oldS, oldO := cmd.U, cmd.S, cmd.O
	t.Cleanup(func() {
		cmd.U, cmd.S, cmd.O = oldU, oldS, oldO
	})

	cmd.U = "https://target.test"
	cmd.S = ""
	cmd.O = ""
	ResultJs = nil
	ResultUrl = []mode.Link{{Url: "https://target.test/api"}}
	Infos = nil
	Fuzzs = nil
	Domains = nil

	assertNotPanic(t, func() {
		OutFileHtml(filepath.Join(t.TempDir(), "missing", "out.html"))
	})
}

func TestCreateOutputFileReturnsOpenError(t *testing.T) {
	file, err := createOutputFile(filepath.Join(t.TempDir(), "missing", "out.html"))
	if err == nil {
		if file != nil {
			file.Close()
		}
		t.Fatal("createOutputFile() error = nil, want open error")
	}
	if file != nil {
		t.Fatal("createOutputFile() file != nil when open fails")
	}
}

func TestOutFileCsvUsesIDcardSectionTitle(t *testing.T) {
	oldU, oldS, oldD := cmd.U, cmd.S, cmd.D
	t.Cleanup(func() {
		cmd.U, cmd.S, cmd.D = oldU, oldS, oldD
		ResultJs = nil
		ResultUrl = nil
		Fuzzs = nil
		Infos = nil
		Domains = nil
	})

	cmd.U = "https://target.test"
	cmd.S = ""
	cmd.D = ""
	ResultJs = nil
	ResultUrl = nil
	Fuzzs = nil
	Infos = []mode.Info{{IDcard: []string{"110101199003074316"}, Source: "https://target.test/app.js"}}

	out := filepath.Join(t.TempDir(), "result.csv")
	OutFileCsv(out)

	file, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, record := range records {
		if len(record) > 0 && record[0] == "IDcard" {
			return
		}
	}
	t.Fatalf("CSV records did not contain IDcard section title: %#v", records)
}

func assertNotPanic(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("function panicked: %v", r)
		}
	}()

	fn()
}
