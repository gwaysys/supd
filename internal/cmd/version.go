package cmd

import (
	"fmt"
)

var (
	Version   = "0.0.0"
	GitCommit = "dev"
)

type VersionCommand struct {
}

var versionCommand VersionCommand

func (v VersionCommand) Execute(args []string) error {
	fmt.Printf("%s-%s\n", Version, GitCommit)
	return nil
}

func init() {
	parser.AddCommand("version",
		"show the version of supervisor",
		"display the supervisor version",
		&versionCommand)
}
