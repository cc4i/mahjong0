package deploy

import (
	"github.com/kris-nova/logger"
	"github.com/spf13/cobra"
	"io/ioutil"
	"mctl/cmd"
)

var Deploy = &cobra.Command{
	Use:   "deploy",
	Short: "\tDeploy Tile/Deployment to target platform.",
	Long:  "\tDeploy Tile/Deployment to target platform as per definition",
	Run: func(cmd *cobra.Command, args []string) {
		deployFunc(cmd, args)
	},
}

func init() {
	Deploy.PersistentFlags().StringP("filename", "f", "", "that contains the configuration to apply")
	Deploy.MarkFlagRequired("filename")
}

func deployFunc(c *cobra.Command, args []string) {

	filename, _ := c.Flags().GetString("filename")
	addr, _ := c.Flags().GetString("addr")
	dryRun, _ := c.Flags().GetBool("dry-run")
	if filename == "" {
		if len(args) == 1 {
			cmd.Run(addr, dryRun, []byte(args[0]))
		} else {
			logger.Info("Need deployment file to apply")
		}
	} else {
		buf, err := ioutil.ReadFile(filename)
		if err != nil {
			logger.Info("%s\n", err)
		}
		cmd.Run(addr, dryRun, buf)
	}

}
