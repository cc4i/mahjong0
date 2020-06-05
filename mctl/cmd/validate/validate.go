package validate

import (
	log "github.com/sirupsen/logrus"
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
		log.Fatal("Need deployment file to apply")
	} else {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("%s\n", err)
		}

		var dcode, tcode int
		dcode, _ = cmd.RunPost(addr, "deployment", buf)
		if dcode != 200 {
			tcode, _ = cmd.RunPost(addr, "tile", buf)
			if tcode != 200 {
				log.Printf("%s\n", "Supplied content were neither 'Deployment' nor 'Tile'")
			}
		}
	}

}
func init() {
	Validate.PersistentFlags().StringP("filename", "f", "", "that contains the configuration to be validated")
	Validate.MarkFlagRequired("filename")
}
