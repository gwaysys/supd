package cmd

import (
	"fmt"
)

const Version = "v1.2.0"

type VersionCommand struct {
}

var versionCommand VersionCommand

func (v VersionCommand) Execute(args []string) error {
	fmt.Println(Version)
	return nil
}

func init() {
	parser.AddCommand("version",
		"show the version of supervisor",
		"display the supervisor version",
		&versionCommand)
}
