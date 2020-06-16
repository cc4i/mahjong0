package version

import (
	"github.com/kris-nova/logger"
	"github.com/spf13/cobra"
	"mctl/cmd"
	"runtime"
)

var (
	ClientVersion string
	GitCommit     string
	Built         string
)
var Version = &cobra.Command{
	Use:   "version",
	Short: "\tPrint the version number of Dice",
	Long:  "\tAll software has versions. This is Dice's",
	Run: func(c *cobra.Command, args []string) {
		clientVersion(c, args)
		serverVersion(c, args)
	},
}

func clientVersion(c *cobra.Command, args []string) {
	logger.Info("\nClient Version:\n"+
		"\tVersion:\t%s\n"+
		"\tGo version:\t%s\n"+
		"\tGit commit:\t%s\n"+
		"\tBuilt:\t%s\n"+
		"\tOS/Arch:\t%s/%s\n", ClientVersion, runtime.Version(), GitCommit, Built, runtime.GOOS, runtime.GOARCH)
}

func serverVersion(c *cobra.Command, args []string) {
	addr, _ := c.Flags().GetString("addr")
	if buf, err := cmd.RunGet(addr, "version"); err != nil {
		logger.Warning(err.Error())
	} else {
		logger.Info("\nServer Version:\n%s\n", buf)
	}

}
