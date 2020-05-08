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
		fmt.Printf("Mahjong Server %s. \nMahjong Client %s.\n", "v0.1.0", "v0.1.0")
	},
}
