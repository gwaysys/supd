//go:build !windows

package svc

import (
	daemon "github.com/ochinchina/go-daemon"
	log "github.com/sirupsen/logrus"
)

// daemonize double-forks the current process and runs proc in the child.
// Equivalent to the legacy Deamonize() in the root package's
// daemonize.go. Unusual spelling is preserved so the public Daemonize
// symbol in svc.go stays backwards compatible with the original shell
// wrappers.
func daemonize(proc func()) {
	context := daemon.Context{LogFileName: "/dev/stdout"}

	child, err := context.Reborn()
	if err != nil {
		context := daemon.Context{}
		child, err = context.Reborn()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Fatal("Unable to run")
		}
	}
	if child != nil {
		return
	}
	defer context.Release()
	proc()
}