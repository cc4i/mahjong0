package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

var CmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Dice",
	Long:  `All software has versions. This is Dice's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Dice Static Site Generator v0.1 -- HEAD")
	},
}