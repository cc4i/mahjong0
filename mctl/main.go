package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mctl/cmd/deploy"
	"mctl/cmd/initial"
	"mctl/cmd/list"
	"mctl/cmd/validate"
	"mctl/cmd/version"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
}

func main() {
	var cmd = &cobra.Command{Use: "mctl"}
	// Root flag for Dice's address
	cmd.PersistentFlags().String("addr", "127.0.0.1:9090", "Dice's address & port, default : 127.0.0.1:9090")
	addr := cmd.PersistentFlags().Lookup("addr")
	addr.Shorthand = "s"

	cmd.PersistentFlags().Bool("dry-run", false, "Only print out the yaml that would be executed")

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
