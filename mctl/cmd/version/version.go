package version

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "\tPrint the version number of Dice",
	Long:  "\tAll software has versions. This is Dice's",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Mahjong Client v0.1")
	},
}