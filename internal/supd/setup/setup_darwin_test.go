//go:build darwin

package setup

import (
	"strings"
	"testing"
)

func TestDarwinPlistDefaultPaths(t *testing.T) {
	got := DarwinPlist("/usr/local/bin", "/etc/supd")
	if len(got) != 499 {
		t.Fatalf("DarwinPlist (defaults) length = %d, want 499 (matches legacy script/supd.plist)", len(got))
	}
	for _, want := range []string{
		"<string>/usr/local/bin/supd</string>",
		"<string>/etc/supd/supd.ini</string>",
		"<string>/usr/local/bin</string>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("DarwinPlist missing %q", want)
		}
	}
}

func TestDarwinPlistCustomPaths(t *testing.T) {
	got := DarwinPlist("/opt/supd/bin", "/etc/supd-custom")
	if !strings.Contains(got, "<string>/opt/supd/bin/supd</string>") {
		t.Fatalf("DarwinPlist did not honor --bin-path:\n%s", got)
	}
	if !strings.Contains(got, "<string>/etc/supd-custom/supd.ini</string>") {
		t.Fatalf("DarwinPlist did not honor --etc-path:\n%s", got)
	}
	if strings.Contains(got, "/usr/local/bin") {
		t.Fatalf("DarwinPlist still contains default path:\n%s", got)
	}
}