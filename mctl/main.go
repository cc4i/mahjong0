package main

import (
	"github.com/kris-nova/logger"
	"github.com/spf13/cobra"
	"mctl/cmd/deploy"
	"mctl/cmd/initial"
	"mctl/cmd/list"
	"mctl/cmd/validate"
	"mctl/cmd/version"
)

func init() {
	// Control colored output
	logger.Color = true
	logger.Fabulous = true
	// Add timestamps
	logger.Timestamps = false
	logger.Level = 4
}

func main() {
	var cmd = &cobra.Command{
		Use:   "mctl",
		Short: "The official CLI for Mahjong",
	}
	// Root flag for Dice's address
	cmd.PersistentFlags().String("addr", "127.0.0.1:9090", "Dice's address & port, default : 127.0.0.1:9090")
	addr := cmd.PersistentFlags().Lookup("addr")
	addr.Shorthand = "s"

	// dry-run
	cmd.PersistentFlags().Bool("dry-run", false, "Only print out the yaml that would be executed")

	// parallel execution
	cmd.PersistentFlags().Bool("parallel", false, "Deployment would be executed with parallel manner")
	parallel := cmd.PersistentFlags().Lookup("parallel")
	parallel.Shorthand = "p"

	// Initial commands
	cmd.AddCommand(initial.Init,
		validate.Validate,
		deploy.Deploy,
		version.Version,
		list.Repo)
	cmd.TraverseChildren = true

	// Running mctl
	cmd.Execute()
}
