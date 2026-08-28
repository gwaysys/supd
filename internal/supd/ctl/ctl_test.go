package ctl

import (
	"strings"
	"testing"
)

func TestRenderProgramIni(t *testing.T) {
	got := RenderProgramIni("myapp", "alice", "/opt/myapp", "/opt/myapp/bin --flag")
	if len(got) == 0 {
		t.Fatal("RenderProgramIni returned empty")
	}
	for _, want := range []string{
		"[program:myapp]",
		"user=alice",
		"directory=/opt/myapp",
		"command=/opt/myapp/bin --flag",
		"stdout_logfile=$PRJ_ROOT/var/log/myapp.logfile.stdout",
		"stderr_logfile_maxbytes=10MB",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("RenderProgramIni output missing %q\n---\n%s", want, got)
		}
	}
	// Substitution correctness: program name should appear in stdout_logfile
	// and stderr_logfile paths.
	if strings.Count(got, "myapp.logfile.stdout") != 2 {
		t.Errorf("expected stdout_logfile path to interpolate appName twice, got %q", got)
	}
}