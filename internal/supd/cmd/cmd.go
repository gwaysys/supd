// Package cmd is the entry-point package for the supervisor daemon.
// After the cmd/ subdirectory was removed, all entry-point code lives
// here: it owns the root flag parser, the cross-subcommand Options
// struct (--configuration / --daemon / --env-file) and the Run() helper
// invoked by the repo-root main.go. All actual implementation lives in
// the setup/ctl/svc subpackages under internal/supd/.
package cmd

import (
	"fmt"
	"os"

	"github.com/gwaysys/supd/internal/supd/ctl"
	"github.com/gwaysys/supd/internal/supd/setup"
	"github.com/gwaysys/supd/internal/supd/svc"
	"github.com/gwaysys/supd/logger"

	"github.com/jessevdk/go-flags"
	log "github.com/sirupsen/logrus"
)

// Options mirrors the legacy root-package flag struct. Kept here so the
// global go-flags parser can bind flags at the root level before
// dispatching to subcommands.
type Options struct {
	Configuration string `short:"c" long:"configuration" description:"the configuration file"`
	Daemon        bool   `short:"d" long:"daemon" description:"run as daemon"`
	EnvFile       string `long:"env-file" description:"the environment file"`
}

var options Options
var parser = flags.NewParser(&options, flags.Default & ^flags.PrintErrors)

// GetOptions returns the root-package Options pointer so other packages
// (e.g. ctl) can read the daemon-mode flag values without depending on
// the unexported global directly.
func GetOptions() *Options {
	return &options
}

// init configures the default logrus logger and registers the three
// top-level subcommands: ctl, setup and (via svc) the daemon mode.
// version/init/encode/decode are registered from their respective files
// in this package (version.go, config_template.go, encode.go).
func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&logger.CustomFormatter{FullTimestamp: true})
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)

	ctl.Register(parser)
	setup.Register(parser)
}

// Run is the entry point invoked by the repo-root main.go. It reaps
// zombies (linux only), then parses CLI flags. When no subcommand is
// given it either daemonizes (--daemon) or runs the supervisor in the
// foreground via svc.NewServer().Run().
func Run() {
	ReapZombie()

	if _, err := parser.Parse(); err != nil {
		flagsErr, ok := err.(*flags.Error)
		if ok {
			switch flagsErr.Type {
			case flags.ErrHelp:
				fmt.Fprintln(os.Stdout, err)
				os.Exit(0)
			case flags.ErrCommandRequired:
				if options.Daemon {
					svc.Daemonize(func() {
						_ = svc.NewServer().Run()
					})
				} else if len(os.Args) == 2 {
					fmt.Println(err.Error())
				} else {
					_ = svc.NewServer().Run()
				}
			default:
				fmt.Fprintf(os.Stderr, "error when parsing command: %s\n", err)
				os.Exit(1)
			}
		}
	}
}