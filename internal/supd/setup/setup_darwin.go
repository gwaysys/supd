//go:build darwin

package setup

import (
	"fmt"
	"os"
)

// DarwinPlist renders /Library/LaunchDaemons/supd.plist for the given
// bin-path + etc-path. Matches the legacy script/supd.plist byte-for-byte
// when the paths match the defaults (/usr/local/bin and /etc/supd).
func DarwinPlist(binPath, etcPath string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>KeepAlive</key>
    <true/>

	<key>Label</key>
	<string>supd</string>

	<key>ProgramArguments</key>
	<array>
        <string>%s/supd</string>
        <string>-c</string>
        <string>%s/supd.ini</string>
	</array>

	<key>WorkingDirectory</key>
    <string>%s</string>
</dict>
</plist>
`, binPath, etcPath, binPath)
}

const darwinPlistPath = "/Library/LaunchDaemons/supd.plist"

// setupDarwin implements install/upgrade/clean for macOS hosts using launchd.
// The semantics match the historical script/boot_darwin.sh wrapper, with the
// addition of補齐式 binary writes and user-configurable bin/etc paths.
func setupDarwin(op, src, binPath, etcPath string) error {
	switch op {
	case "install":
		if _, err := os.Stat(darwinPlistPath); err == nil && !isForceInstall() {
			fmt.Println("Already installed")
			if err := installBin(src, binPath, etcPath); err != nil {
				return err
			}
			return nil
		}
		if err := installBin(src, binPath, etcPath); err != nil {
			return err
		}
		if err := writeAsRoot(darwinPlistPath, []byte(DarwinPlist(binPath, etcPath)), 0644); err != nil {
			return err
		}
		return sudoExec("launchctl", "load", darwinPlistPath)

	case "upgrade":
		if _, err := os.Stat(darwinPlistPath); err == nil {
			if err := upgradeBin(src, binPath); err != nil {
				return err
			}
			fmt.Println("Upgrade the binary done. You need use 'launchctl' tools to restart supd by manually.")
			return nil
		}
		fmt.Println("Not installed")

	case "clean":
		if _, err := os.Stat(darwinPlistPath); err == nil {
			if err := sudoExec("launchctl", "unload", darwinPlistPath); err != nil {
				return err
			}
			if err := sudoExec("rm", "-f", darwinPlistPath); err != nil {
				return err
			}
			return cleanBin(binPath)
		}
		fmt.Println("Not installed")
	}
	return nil
}

func init() {
	handlers["darwin"] = setupDarwin
}