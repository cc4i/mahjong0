package validate

import (
	"github.com/kris-nova/logger"
	"github.com/spf13/cobra"
	"io/ioutil"
	"mctl/cmd"
)

var Validate = &cobra.Command{
	Use:   "validate",
	Short: "\tValidate Tile/Deployment as per semantic definition",
	Long:  "\tValidate Tile/Deployment as per semantic definition",
	Run: func(cmd *cobra.Command, args []string) {
		validateFunc(cmd, args)
	},
}

func validateFunc(c *cobra.Command, args []string) {
	addr, _ := c.Flags().GetString("addr")
	filename, _ := c.Flags().GetString("filename")
	if filename == "" {
		logger.Info("Need deployment file to apply")
	} else {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			logger.Warning("%s\n", err)
		}

		var dcode, tcode int
		dcode, _ = cmd.RunPostByVersion(addr, "deployment", buf)
		if dcode != 200 {
			tcode, _ = cmd.RunPostByVersion(addr, "tile", buf)
			if tcode != 200 {
				logger.Info("%s\n", "Supplied content were neither 'Deployment' nor 'Tile'")
			}
		}
	}

}
func init() {
	Validate.PersistentFlags().StringP("filename", "f", "", "that contains the configuration to be validated")
	Validate.MarkFlagRequired("filename")
}
