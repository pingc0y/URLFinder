package crawler

import (
	"testing"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
	"github.com/pingc0y/URLFinder/result"
)

func TestLoadKeepsValidationChannelsUsableForSmallThreadCounts(t *testing.T) {
	oldU, oldF, oldFF := cmd.U, cmd.F, cmd.FF
	oldI, oldH := cmd.I, cmd.H
	oldT, oldTI := cmd.T, cmd.TI
	oldX := cmd.X
	oldCh, oldJsch, oldUrlch := config.Ch, config.Jsch, config.Urlch

	t.Cleanup(func() {
		cmd.U, cmd.F, cmd.FF = oldU, oldF, oldFF
		cmd.I, cmd.H = oldI, oldH
		cmd.T, cmd.TI = oldT, oldTI
		cmd.X = oldX
		config.Ch, config.Jsch, config.Urlch = oldCh, oldJsch, oldUrlch
	})

	cmd.U = "https://example.com"
	cmd.F = ""
	cmd.FF = ""
	cmd.I = false
	cmd.H = false
	cmd.T = 1
	cmd.TI = 1
	cmd.X = ""

	load()

	if cap(config.Jsch) < 1 {
		t.Fatalf("cap(config.Jsch) = %d, want at least 1", cap(config.Jsch))
	}
	if cap(config.Urlch) < 1 {
		t.Fatalf("cap(config.Urlch) = %d, want at least 1", cap(config.Urlch))
	}
}

func TestAppendJsHandlesSourceWithoutURL(t *testing.T) {
	Initialization()

	assertNotPanic(t, func() {
		got := AppendJs("https://example.com/app.js", "inline-script")
		if got != 0 {
			t.Fatalf("AppendJs() = %d, want 0", got)
		}
	})

	if len(result.ResultJs) != 1 {
		t.Fatalf("len(result.ResultJs) = %d, want 1", len(result.ResultJs))
	}
	if result.Jsinurl["https://example.com/app.js"] != "inline-script" {
		t.Fatalf("Jsinurl fallback = %q, want source fallback", result.Jsinurl["https://example.com/app.js"])
	}
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
