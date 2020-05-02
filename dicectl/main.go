package main

import (
	"dicectl/cmd"
	"github.com/spf13/cobra"
)

func main() {
	 var command = &cobra.Command{Use: "dicectl"}
	command.AddCommand(cmd.CmdVersion)
	command.Execute()
}