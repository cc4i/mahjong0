package version

import (
	"fmt"
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
	fmt.Printf("Client Version:\n")
	fmt.Printf("\tVersion:\t%s\n", ClientVersion)
	fmt.Printf("\tGo version:\t%s\n", runtime.Version())
	fmt.Printf("\tGit commit:\t%s\n", GitCommit)
	fmt.Printf("\tBuilt:\t%s\n", Built)
	fmt.Printf("\tOS/Arch:\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
}

func serverVersion(c *cobra.Command, args []string) {
	//TODO Query from server
	fmt.Printf("Server Version:\n")
	addr, _ := c.Flags().GetString("addr")
	if buf, err := cmd.RunGet(addr, "version"); err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Printf("%s\n", buf)
	}

}
