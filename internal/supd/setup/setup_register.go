package setup

import (
	"github.com/jessevdk/go-flags"
)

// Register wires the setup subcommand into the provided go-flags parser.
// The supd root package calls this from its init() so that the global
// `parser` variable does not have to be exported into this package.
func Register(parser *flags.Parser) {
	parser.AddCommand("setup",
		"install/upgrade/clean/extract supd on the host system",
		"The setup subcommand replaces the legacy setup.sh wrapper. It registers supd "+
			"with systemd on linux or launchd on darwin, lays down /etc/supd/supd.ini and "+
			"the /usr/local/bin/supc wrapper, and removes everything on clean. "+
			"`setup extract` lays down supd+supc only without registering a service.",
		&Command{})
}