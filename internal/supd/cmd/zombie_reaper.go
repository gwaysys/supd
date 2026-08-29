// +build !windows

package cmd

import (
	reaper "github.com/ochinchina/go-reaper"
)

func ReapZombie() {
	go reaper.Reap()
}
