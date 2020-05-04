package validate

import (
	"github.com/spf13/cobra"
)

var Validate = &cobra.Command{
	Use:   "validate",
	Short: "\tValidate Tile/Deployment as per semantic definition",
	Long:  "\tValidate Tile/Deployment as per semantic definition",
	Run: func(cmd *cobra.Command, args []string) {
		validateFunc(cmd, args)
	},
}

func validateFunc(cmd *cobra.Command, args []string) {

}
