package list

import (
	"fmt"
	"github.com/spf13/cobra"
)

var TilesInRepo = &cobra.Command{
	Use:   "tile",
	Short: "\tList Tiles in the Repo.",
	Long:  "\tList Tiles in the Repo.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Tile")
	},
}

var HuInRepo = &cobra.Command{
	Use:   "tile",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hu")
	},
}

var SchemaInRepo = &cobra.Command{
	Use:   "schema",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("schema")
	},
}

var ListRepo = &cobra.Command{
	Use:   "list",
	Short: "\tList content (Tiles, Hu, etc.) in the Repo.",
	Long:  "\tList content (Tiles, Hu, etc.) in the Repo.",
}

func init() {
	ListRepo.AddCommand(TilesInRepo, HuInRepo, SchemaInRepo)
}
