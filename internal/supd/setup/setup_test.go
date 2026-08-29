package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSupcScriptLiteral(t *testing.T) {
	want := "#/bin/sh\n\nsudo supd ctl \"$@\"\n"
	if supcScript != want {
		t.Fatalf("supcScript mismatch:\n got %q\nwant %q", supcScript, want)
	}
	if len(supcScript) != 29 {
		t.Fatalf("supcScript length = %d, want 29 (matches legacy bin/supc)", len(supcScript))
	}
}

func TestSupdIniLiteral(t *testing.T) {
	if len(supdIni) != 451 {
		t.Fatalf("supdIni length = %d, want 451 (matches legacy cmd/supd/etc/supd/supd.ini)", len(supdIni))
	}
}

func TestSetupEnvKind(t *testing.T) {
	got := envKind()
	switch got {
	case "darwin", "systemd", "":
		// expected
	default:
		t.Fatalf("unexpected envKind() = %q", got)
	}
}

func TestSetupHandlersRegistration(t *testing.T) {
	// The current build must register exactly one init-system handler
	// (linux -> systemd, darwin -> darwin). Windows / other builds have none.
	h := HandlerForTest()
	switch envKind() {
	case "systemd":
		if _, ok := h["systemd"]; !ok {
			t.Fatalf("expected handlers[\"systemd\"] to be registered on linux")
		}
		if _, ok := h["darwin"]; ok {
			t.Fatalf("handlers[\"darwin\"] must NOT be registered on linux builds")
		}
	case "darwin":
		if _, ok := h["darwin"]; !ok {
			t.Fatalf("expected handlers[\"darwin\"] to be registered on darwin")
		}
		if _, ok := h["systemd"]; ok {
			t.Fatalf("handlers[\"systemd\"] must NOT be registered on darwin builds")
		}
	default:
		if len(h) != 0 {
			t.Fatalf("expected zero handlers on unsupported platforms, got %d entries", len(h))
		}
	}
}

func TestLocateSourceBinaryOverride(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fake-supd")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(sourceEnvName, bin)

	got, err := locateSourceBinary()
	if err != nil {
		t.Fatalf("locateSourceBinary: %v", err)
	}
	if got != bin {
		t.Fatalf("locateSourceBinary = %q, want %q", got, bin)
	}
}

func TestLocateSourceBinaryMissingOverride(t *testing.T) {
	t.Setenv(sourceEnvName, "/nonexistent/supd-path-for-test")
	if _, err := locateSourceBinary(); err == nil {
		t.Fatal("expected error when SUPD_SETUP_SOURCE points at a missing file")
	}
}

func TestTrim(t *testing.T) {
	cases := map[string]string{
		"abc\n":      "abc",
		"abc\r\n":    "abc",
		"abc ":       "abc",
		"abc \r\n":   "abc",
		"":           "",
		"a":          "a",
		"\n\r\n \r ": "",
	}
	for in, want := range cases {
		if got := trim(in); got != want {
			t.Errorf("trim(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestHasTTY(t *testing.T) {
	// We can't usefully assert true/false here because tests are usually non-tty;
	// just make sure the function returns without panicking.
	_ = hasTTY()
}

func TestSetupExecuteUsage(t *testing.T) {
	// When invoked with no positional args, Command.Execute must print the
	// four-line usage and return nil regardless of host state.
	c := &Command{}
	if err := c.Execute(nil); err != nil {
		t.Fatalf("Execute(nil) = %v, want nil", err)
	}
	c.Op = "bogus"
	if err := c.Execute(nil); err != nil {
		t.Fatalf("Execute(Op=bogus) = %v, want nil", err)
	}
	c.Op = ""
	if err := c.Execute([]string{"also-bogus"}); err != nil {
		t.Fatalf("Execute(args=bogus) = %v, want nil", err)
	}
}

func TestSetupExecuteUnsupportedOS(t *testing.T) {
	if envKind() != "" {
		t.Skip("test only meaningful on hosts without systemd/darwin")
	}
	c := &Command{Op: "install"}
	if err := c.Execute(nil); err != nil {
		t.Fatalf("Execute on unsupported OS = %v, want nil", err)
	}
}

func TestResolveInstallPathsDefaults(t *testing.T) {
	c := &Command{}
	c.ResolveInstallPaths()
	if c.BinPath != BinPath {
		t.Errorf("default BinPath = %q, want %q", c.BinPath, BinPath)
	}
	if c.EtcPath != EtcPath {
		t.Errorf("default EtcPath = %q, want %q", c.EtcPath, EtcPath)
	}
}

func TestResolveInstallPathsOverrides(t *testing.T) {
	c := &Command{BinPath: "/opt/supd/bin", EtcPath: "/etc/supd-custom"}
	c.ResolveInstallPaths()
	if c.BinPath != "/opt/supd/bin" {
		t.Errorf("BinPath not preserved: got %q", c.BinPath)
	}
	if c.EtcPath != "/etc/supd-custom" {
		t.Errorf("EtcPath not preserved: got %q", c.EtcPath)
	}
}

func TestNeedsWriteMissing(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "nope")
	if !needsWrite(missing) {
		t.Fatalf("needsWrite should return true for a missing path")
	}
}

func TestNeedsWriteNonEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "filled")
	if err := os.WriteFile(p, []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	if needsWrite(p) {
		t.Fatalf("needsWrite should return false for a non-empty existing file")
	}
}

func TestNeedsWriteEmpty(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "empty")
	if err := os.WriteFile(p, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if !needsWrite(p) {
		t.Fatalf("needsWrite should return true for a zero-size existing file")
	}
}

func TestRunExtractDryRun(t *testing.T) {
	src := filepath.Join(t.TempDir(), "fake-supd")
	if err := os.WriteFile(src, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(dryRunEnvName, "1")
	binPath := "/opt/supd-test/bin"

	if err := RunExtract(src, binPath); err != nil {
		t.Fatalf("RunExtract dry-run: %v", err)
	}
}

func TestRunExtractMissingSource(t *testing.T) {
	if err := RunExtract("/no/such/supd", "/tmp/bin"); err == nil {
		t.Fatal("expected error when source binary is missing")
	}
}

func TestSetupExecuteExtractMissingSource(t *testing.T) {
	t.Setenv(sourceEnvName, "/no/such/supd-for-extract-test")
	c := &Command{Op: "extract"}
	if err := c.Execute(nil); err != nil {
		t.Fatalf("Execute(extract) = %v, want nil (missing source must surface as message, not error)", err)
	}
}

func TestDemoIniTplBytes(t *testing.T) {
	// DemoIniTpl must be byte-for-byte identical with the historical
	// cmd/supd/etc/supd/conf.d/demo.ini.tpl file. That on-disk file
	// was removed along with the cmd/ directory; verify against a
	// pinned literal that mirrors its documented byte sequence.
	want := "# rename this tpl to demo.ini for running\n" +
		"[program:demo]\n" +
		"#user=root # run user\n" +
		"#environment=GIN_MODE=\"release\"\n" +
		"#directory= # current run directory\n" +
		"command=ping 163.com # run program\n" +
		"autostart=true\n" +
		"startsecs=3\n" +
		"startretries=3\n" +
		"autorestart=true\n" +
		"exitcodes=0,2\n" +
		"stopsignal=TERM\n" +
		"stopwaitsecs=10\n" +
		"#stopasgroup=true\n" +
		"#killasgroup=true\n" +
		"stdout_logfile=$PRJ_ROOT/var/log/demo.logfile.stdout # if PRJ_ROOT not set, using the root directory\n" +
		"stdout_logfile_maxbytes=1MB\n" +
		"stdout_logfile_backups=10\n" +
		"stderr_logfile=$PRJ_ROOT/var/log/demo.logfile.stderr\n" +
		"stderr_logfile_maxbytes=1MB\n" +
		"stderr_logfile_backups=10\n"
	if want != DemoIniTpl {
		t.Fatalf("DemoIniTpl mismatch (pinned literal %d bytes vs const %d bytes)",
			len(want), len(DemoIniTpl))
	}
}