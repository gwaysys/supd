//go:build linux

package setup

import (
	"strings"
	"testing"
)

func TestSystemdServiceUnitDefaultPaths(t *testing.T) {
	got := SystemdServiceUnit("/usr/local/bin", "/etc/supd")
	if len(got) != 365 {
		t.Fatalf("SystemdServiceUnit (defaults) length = %d, want 365 (matches legacy script/supd.service)", len(got))
	}
	for _, want := range []string{
		"ExecStart=/usr/local/bin/supd -c /etc/supd/supd.ini",
		"ExecStop=/usr/local/bin/supd ctl shutdown",
		"ExecReload=/usr/local/bin/supd ctl reload",
		"WantedBy=multi-user.target",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("SystemdServiceUnit missing %q", want)
		}
	}
}

func TestSystemdServiceUnitCustomPaths(t *testing.T) {
	got := SystemdServiceUnit("/opt/supd/bin", "/etc/supd-custom")
	if !strings.Contains(got, "ExecStart=/opt/supd/bin/supd -c /etc/supd-custom/supd.ini") {
		t.Fatalf("SystemdServiceUnit did not honor --bin-path/--etc-path:\n%s", got)
	}
	if strings.Contains(got, "/usr/local/bin") {
		t.Fatalf("SystemdServiceUnit still contains default path:\n%s", got)
	}
}