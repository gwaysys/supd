//go:build linux

package setup

import (
	"fmt"
	"os"
)

// SystemdServiceUnit renders /lib/systemd/system/supd.service for the given
// bin-path + etc-path. Matches the legacy script/supd.service byte-for-byte
// when the paths match the defaults (/usr/local/bin and /etc/supd).
func SystemdServiceUnit(binPath, etcPath string) string {
	return fmt.Sprintf(`[Unit]
Description=Supervisor process control system for UNIX
Documentation=https://github.com/gwaysys/supd
After=network.target

[Service]
ExecStart=%s/supd -c %s/supd.ini
ExecStop=%s/supd ctl shutdown
ExecReload=%s/supd ctl reload
KillMode=process
Restart=on-failure
RestartSec=50s

[Install]
WantedBy=multi-user.target
`, binPath, etcPath, binPath, binPath)
}

const systemdUnitPath = "/lib/systemd/system/supd.service"

// setupSystemd implements install/upgrade/clean for linux hosts running systemd.
// The semantics match the historical script/boot_systemd.sh wrapper, with the
// addition of補齐式 binary writes and user-configurable bin/etc paths.
func setupSystemd(op, src, binPath, etcPath string) error {
	switch op {
	case "install":
		if _, err := os.Stat(systemdUnitPath); err == nil && !isForceInstall() {
			fmt.Println("Already installed")
			//补齐式: even when already installed, ensure supd + supc binaries
			//and /etc/supd/supd.ini are present. This is safe because each
			//individual installBin call only writes files that are missing.
			if err := installBin(src, binPath, etcPath); err != nil {
				return err
			}
			return nil
		}
		if err := installBin(src, binPath, etcPath); err != nil {
			return err
		}
		if err := writeAsRoot(systemdUnitPath, []byte(SystemdServiceUnit(binPath, etcPath)), 0644); err != nil {
			return err
		}
		if err := sudoExec("systemctl", "daemon-reload"); err != nil {
			return err
		}
		if err := sudoExec("systemctl", "enable", "supd"); err != nil {
			return err
		}
		return sudoExec("systemctl", "start", "supd")

	case "upgrade":
		if _, err := os.Stat(systemdUnitPath); err == nil {
			if err := upgradeBin(src, binPath); err != nil {
				return err
			}
			fmt.Println("Upgrade bin done, need do 'systemctl restart supd' by manually.")
			// Intentionally NOT calling `systemctl restart supd` here:
			// matches the historical `#sudo systemctl restart supd` comment.
			return nil
		}
		fmt.Println("Not installed")

	case "clean":
		if _, err := os.Stat(systemdUnitPath); err == nil {
			if err := sudoExec("systemctl", "stop", "supd"); err != nil {
				return err
			}
			if err := sudoExec("systemctl", "disable", "supd"); err != nil {
				return err
			}
			if err := sudoExec("rm", "-f", systemdUnitPath); err != nil {
				return err
			}
			if err := sudoExec("systemctl", "daemon-reload"); err != nil {
				return err
			}
			return cleanBin(binPath)
		}
		fmt.Println("Not installed")
	}
	return nil
}

func init() {
	handlers["systemd"] = setupSystemd
}