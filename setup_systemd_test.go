//go:build linux

package supd

import (
	"strings"
	"testing"
)

func TestSystemdServiceUnitDefaultPaths(t *testing.T) {
	got := systemdServiceUnit("/usr/local/bin", "/etc/supd")
	if len(got) != 365 {
		t.Fatalf("systemdServiceUnit (defaults) length = %d, want 365 (matches legacy script/supd.service)", len(got))
	}
	for _, want := range []string{
		"ExecStart=/usr/local/bin/supd -c /etc/supd/supd.ini",
		"ExecStop=/usr/local/bin/supd ctl shutdown",
		"ExecReload=/usr/local/bin/supd ctl reload",
		"WorkingDirectory not applicable on systemd", // sentinel to silence linter; replaced below
	} {
		_ = want
	}
	for _, want := range []string{
		"ExecStart=/usr/local/bin/supd -c /etc/supd/supd.ini",
		"ExecStop=/usr/local/bin/supd ctl shutdown",
		"ExecReload=/usr/local/bin/supd ctl reload",
		"WantedBy=multi-user.target",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("systemdServiceUnit missing %q", want)
		}
	}
}

func TestSystemdServiceUnitCustomPaths(t *testing.T) {
	got := systemdServiceUnit("/opt/supd/bin", "/etc/supd-custom")
	if !strings.Contains(got, "ExecStart=/opt/supd/bin/supd -c /etc/supd-custom/supd.ini") {
		t.Fatalf("systemdServiceUnit did not honor --bin-path/--etc-path:\n%s", got)
	}
	if strings.Contains(got, "/usr/local/bin") {
		t.Fatalf("systemdServiceUnit still contains default path:\n%s", got)
	}
}