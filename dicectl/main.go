package main

import (
	"dicectl/cmd/version"
	"github.com/spf13/cobra"
)

func main() {
	var command = &cobra.Command{Use: "dicectl"}
	command.AddCommand(version.CmdVersion)
	command.Execute()
}