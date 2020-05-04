package initial

import (
	"github.com/spf13/cobra"
)


var Init = &cobra.Command{
	Use:   "init",
	Short: "\tInitial Tile or Deployment.",
	Long:  "\tInitial Tile or Deployment with all basic needs.",
	TraverseChildren: true,
	//Run: func(cmd *cobra.Command, args []string) {
	//	fmt.Println("Tile")
	//},
}


var Tile = &cobra.Command{
	Use:   "tile",
	Short: "\tInitial Tile with basic template.",
	Long:  "\tInitial Tile to kick off building a nwe Tile, with all basic needs.",
	Run: func(cmd *cobra.Command, args []string) {
		tileFunc(cmd, args)
	},
}

var Deployment = &cobra.Command{
	Use:   "deployment",
	Short: "\tInitial Deployment with basic template.",
	Long:  "\tInitial Deployment to start a nwe deployment, with all basic needs.",
	Run: func(cmd *cobra.Command, args []string) {
		deploymentFunc(cmd, args)
	},
}


func init() {
	Init.PersistentFlags().String("name","", "The name of building Tile/Deployment")
	Init.PersistentFlags().String("version","0.0.1", "The version of building Tile/Deployment")
	Init.PersistentFlags().String("directory","", "Where to place the Tile/Deployment with templates")

	Init.AddCommand(Tile, Deployment)

}


func tileFunc(cmd *cobra.Command, args []string) {

}

func deploymentFunc(cmd *cobra.Command, args []string) {

}