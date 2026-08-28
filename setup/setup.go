// Package setup implements the `supd setup` subcommand, replacing the legacy
// setup.sh / script/boot_*.sh / script/install_bin.sh stack with a single
// Go-implemented registry of operations.
//
// The package owns all install/upgrade/clean/extract logic. Per-OS init system
// handlers (systemd on linux, launchd on darwin) are registered into
// `handlers` by setup_systemd.go / setup_darwin.go via build tags, so the
// compiled binary only links the handler matching the target GOOS.
package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Command is the entry of `supd setup` subcommand.
// It exposes install/upgrade/clean/extract as four Go-implemented operations.
type Command struct {
	Op string `positional-args:"yes" required:"0" description:"install, upgrade, clean or extract"`

	// Flags. Defaults align with the historical install_bin.sh paths.
	BinPath string `short:"b" long:"bin-path" description:"destination directory for supd + supc binaries (default /usr/local/bin)"`
	EtcPath string `short:"e" long:"etc-path" description:"destination directory for supd.ini and conf.d (default /etc/supd)"`
}

// supcScript is the literal content written to <bin-path>/supc by `supd setup install`.
// Keep this byte-for-byte identical with the historical bin/supc wrapper.
const supcScript = "#/bin/sh\n\nsudo supd ctl \"$@\"\n"

// supdIni is the default <etc-path>/supd.ini written by `supd setup install`.
// Kept byte-for-byte identical with cmd/supd/etc/supd/supd.ini.
const supdIni = `#
## Not need root auth
#[inet_http_server]
#port=127.0.0.1:9002
#
#[supervisorctl]
#serverurl=http://127.0.0.1:9002

# Need root auth when deploy in root
[unix_http_server]
file=/var/supd/supd.sock

[supervisorctl]
serverurl=unix:///var/supd/supd.sock

# for supd
[supervisord]
pidfile = /var/supd/supd.pid
logfile = /dev/stdout,/var/log/supd.log
logfile_maxbytes = 10MB
logfile_backups = 10
loglevel = info

[include]
files = /etc/supd/conf.d/*.ini
`

// handlers is populated by setup_systemd.go / setup_darwin.go via build tags.
// The key is the OS identifier returned by envKind(); the value is the function
// that performs install/upgrade/clean for that init system, given the resolved
// source-binary path, the user-requested bin-path and etc-path.
var handlers = map[string]func(op, src, binPath, etcPath string) error{}

// Default install paths, kept aligned with the original install_bin.sh.
const (
	BinPath  = "/usr/local/bin"
	EtcPath  = "/etc/supd"
	usrBin   = "/usr/bin"
	dryRunEnvName = "SUPD_SETUP_DRYRUN"
	sourceEnvName = "SUPD_SETUP_SOURCE"
	forceEnvName  = "SUPD_SETUP_FORCE_INSTALL"
)

// envKind returns "darwin" / "systemd" / "" depending on runtime OS and the
// availability of systemd on linux. Mirrors the bootName dispatch in setup.sh.
func envKind() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		if _, err := os.Stat("/lib/systemd/system"); err == nil {
			return "systemd"
		}
	}
	return ""
}

// ResolveInstallPaths fills in c.BinPath / c.EtcPath with defaults when the user
// did not provide them on the command line.
func (c *Command) ResolveInstallPaths() {
	if c.BinPath == "" {
		c.BinPath = BinPath
	}
	if c.EtcPath == "" {
		c.EtcPath = EtcPath
	}
}

// locateSourceBinary decides which supd binary should be copied into bin-path.
// Priority:
//  1. $SUPD_SETUP_SOURCE (explicit override for cross-builds / custom GOBIN)
//  2. $GOBIN/supd from `go env GOBIN` (the canonical go install target)
//  3. The currently running executable itself when invoked from a non-GOBIN location
//     (e.g. running ./supd from the repo root for development).
func locateSourceBinary() (string, error) {
	if v := os.Getenv(sourceEnvName); v != "" {
		if _, err := os.Stat(v); err == nil {
			return v, nil
		}
		return "", fmt.Errorf("SUPD_SETUP_SOURCE=%s not found", v)
	}

	if gobin, err := goBin(); err == nil && gobin != "" {
		candidate := filepath.Join(gobin, "supd")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		if _, err := os.Stat(exe); err == nil {
			return exe, nil
		}
	}

	return "", fmt.Errorf("supd binary not found; please run 'go install github.com/gwaysys/supd@latest' first")
}

// goBin returns the value of `go env GOBIN`; empty string + nil error when go is not on PATH.
func goBin() (string, error) {
	if !isDryRun() {
		out, err := exec.Command("go", "env", "GOBIN").Output()
		if err != nil {
			return "", err
		}
		return trim(string(out)), nil
	}
	// In dry-run mode, do not shell out to go; rely on the GOBIN env var if set.
	return trim(os.Getenv("GOBIN")), nil
}

func trim(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r' || s[len(s)-1] == ' ') {
		s = s[:len(s)-1]
	}
	return s
}

// IsDryRun returns true when the user opts into a non-mutating setup for verification / CI.
func IsDryRun() bool {
	return os.Getenv(dryRunEnvName) == "1"
}

// isDryRun internal alias to avoid stutter when calling from this package.
func isDryRun() bool { return IsDryRun() }

// isForceInstall returns true when the user wants to bypass the "Already installed"
// early return in the systemd / darwin handlers.
func isForceInstall() bool {
	return os.Getenv(forceEnvName) == "1"
}

// sudoExec runs a privileged command. As root it executes directly; otherwise it
// shells out to `sudo -E` with stdin/stdout/stderr attached so the password prompt
// works on a real tty. Non-tty + non-root callers are rejected up-front to mirror
// the failure mode of the original shell scripts in CI environments.
func sudoExec(name string, args ...string) error {
	if isDryRun() {
		log.Printf("[dry-run] would exec: %s %v", name, args)
		return nil
	}
	if os.Geteuid() == 0 {
		cmd := exec.Command(name, args...)
		return cmd.Run()
	}
	full := append([]string{"-E", name}, args...)
	cmd := exec.Command("sudo", full...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// hasTTY reports whether stdin is attached to an interactive terminal.
// Used to short-circuit sudo password prompts in CI / docker builds.
func hasTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// copyAsRoot copies src to dst by delegating to `sudo cp -f`. Used for paths under
// the bin-path / etc-path / systemd unit location that require root.
func copyAsRoot(src, dst string) error {
	if isDryRun() {
		log.Printf("[dry-run] would sudo cp %s %s", src, dst)
		return nil
	}
	return sudoExec("cp", "-f", src, dst)
}

// writeAsRoot writes data to dst with mode via a temp file + sudo cp. Used for
// inline unit files like /lib/systemd/system/supd.service.
func writeAsRoot(dst string, data []byte, mode os.FileMode) error {
	if isDryRun() {
		log.Printf("[dry-run] would sudo tee %s (%d bytes, mode=%v)", dst, len(data), mode)
		return nil
	}
	tmp, err := os.CreateTemp("", "supd-setup-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return sudoExec("cp", "-f", tmpName, dst)
}

// mkdirAsRoot creates a directory tree with the given mode via `sudo mkdir -p`.
func mkdirAsRoot(dir string, mode os.FileMode) error {
	if isDryRun() {
		log.Printf("[dry-run] would sudo mkdir -p %s (mode=%v)", dir, mode)
		return nil
	}
	return sudoExec("mkdir", "-p", dir)
}

// needsWrite reports whether dst is missing or empty, in which case a補齐式
// write should run. Stale / non-empty files are left untouched (補齐式 semantics).
func needsWrite(dst string) bool {
	st, err := os.Stat(dst)
	if err != nil {
		return true
	}
	return st.Size() == 0
}

// installBin copies supd + supc into bin-path and lays down etc-path/supd.ini + etc-path/conf.d.
// Each individual target is written only when missing/empty (補齐式 semantics):
// running setup install twice never overwrites an existing non-empty file.
// src is the absolute path to the supd binary that should be installed (typically $GOBIN/supd).
func installBin(src, installBinPath, installEtcPath string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source supd binary not found: %s", src)
	}
	// `/usr/bin` is created as a side-effect, matching install_bin.sh.
	if err := mkdirAsRoot(usrBin, 0755); err != nil {
		return err
	}

	supdDst := filepath.Join(installBinPath, "supd")
	if needsWrite(supdDst) {
		if err := copyAsRoot(src, supdDst); err != nil {
			return err
		}
	} else {
		log.Printf("supd already present at %s, skipping copy", supdDst)
	}

	supcDst := filepath.Join(installBinPath, "supc")
	if needsWrite(supcDst) {
		if err := writeAsRoot(supcDst, []byte(supcScript), 0755); err != nil {
			return err
		}
	} else {
		log.Printf("supc already present at %s, skipping write", supcDst)
	}

	if err := mkdirAsRoot(installEtcPath, 0755); err != nil {
		return err
	}
	if err := mkdirAsRoot(filepath.Join(installEtcPath, "conf.d"), 0755); err != nil {
		return err
	}
	iniDst := filepath.Join(installEtcPath, "supd.ini")
	if needsWrite(iniDst) {
		if err := writeAsRoot(iniDst, []byte(supdIni), 0644); err != nil {
			return err
		}
	} else {
		log.Printf("supd.ini already present at %s, skipping write", iniDst)
	}

	if !isDryRun() {
		fmt.Println("Install supd done.")
	}
	return nil
}

// upgradeBin refreshes supd + supc in bin-path without touching etc-path.
// Each binary target is only overwritten when missing/empty (補齐式).
func upgradeBin(src, installBinPath string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source supd binary not found: %s", src)
	}
	supdDst := filepath.Join(installBinPath, "supd")
	if needsWrite(supdDst) {
		if err := copyAsRoot(src, supdDst); err != nil {
			return err
		}
	} else {
		log.Printf("supd already present at %s, skipping copy", supdDst)
	}
	supcDst := filepath.Join(installBinPath, "supc")
	if needsWrite(supcDst) {
		if err := writeAsRoot(supcDst, []byte(supcScript), 0755); err != nil {
			return err
		}
	} else {
		log.Printf("supc already present at %s, skipping write", supcDst)
	}
	return nil
}

// cleanBin removes supd + supc from bin-path but leaves etc-path alone,
// matching the historical "need remove '$install_etc_path' by manully" behavior.
func cleanBin(installBinPath string) error {
	if isDryRun() {
		log.Printf("[dry-run] would sudo rm -f %s/supd %s/supc", installBinPath, installBinPath)
		log.Printf("[dry-run] /etc/supd must be removed manually")
		return nil
	}
	// Best-effort: ignore individual rm failures so partial installs can still be cleaned.
	_ = sudoExec("rm", "-f", filepath.Join(installBinPath, "supd"))
	_ = sudoExec("rm", "-f", filepath.Join(installBinPath, "supc"))
	fmt.Printf("Clean bin done, need remove '%s' by manully.\n", EtcPath)
	return nil
}

// RunExtract implements `supd setup extract --bin-path <dir>`:
// lay down supd + supc (補齐式) into bin-path only, no etc-path, no init system.
// Independent of init system — works on plain linux containers / FreeBSD / etc.
func RunExtract(src, binPath string) error {
	if binPath == "" {
		binPath = BinPath
	}
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("source supd binary not found: %s", src)
	}
	if err := mkdirAsRoot(binPath, 0755); err != nil {
		return err
	}
	if err := upgradeBin(src, binPath); err != nil {
		return err
	}
	if !isDryRun() {
		fmt.Printf("Extract supd+supc to %s done.\n", binPath)
	}
	return nil
}

// PrintUsage prints the four-line usage banner for `supd setup`.
func PrintUsage() {
	fmt.Println("install -- install to system.")
	fmt.Println("upgrade -- reinstall supd binary and restart supd.")
	fmt.Println("clean -- remove the installed.")
	fmt.Println("extract -- extract supd+supc to --bin-path only, no service registration.")
}

// Execute implements go-flags.Command; it dispatches on the positional Op field.
func (c *Command) Execute(args []string) error {
	op := c.Op
	if op == "" && len(args) > 0 {
		op = args[0]
	}
	if op == "" {
		PrintUsage()
		return nil
	}

	c.ResolveInstallPaths()

	// Non-root + no-tty => sudo cannot prompt; bail out cleanly instead of failing
	// mid-way through the install like the original shell scripts would.
	// extract is allowed to run even in this state as long as no privileged writes
	// are required (target dir writable by current user); sudoExec itself will
	// surface the failure otherwise.
	if op != "extract" && os.Geteuid() != 0 && !hasTTY() {
		fmt.Println("please run with sudo or as root")
		return nil
	}

	switch op {
	case "install", "upgrade", "clean":
		// fall through to handler dispatch below
	case "extract":
		src, err := locateSourceBinary()
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if err := RunExtract(src, c.BinPath); err != nil {
			fmt.Println(err)
		}
		return nil
	default:
		PrintUsage()
		return nil
	}

	kind := envKind()
	if kind == "" {
		fmt.Println("system not support, run supd by command './bin/supd -c ./etc/supd/supd.ini'")
		return nil
	}

	handler, ok := handlers[kind]
	if !ok {
		fmt.Printf("no setup handler registered for kind=%s (this binary was built without it)\n", kind)
		return nil
	}

	src, err := locateSourceBinary()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return handler(op, src, c.BinPath, c.EtcPath)
}

// HandlerForTest exposes the handlers map for tests; do not use from production code.
func HandlerForTest() map[string]func(op, src, binPath, etcPath string) error {
	return handlers
}