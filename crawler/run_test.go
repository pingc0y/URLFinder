package crawler

import (
	"testing"

	"github.com/pingc0y/URLFinder/cmd"
	"github.com/pingc0y/URLFinder/config"
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
