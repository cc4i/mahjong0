package initial

import (
	"bufio"
	"crypto/tls"
	"github.com/spf13/cobra"
	"log"
	"mctl/cmd"
	"net/http"
)

var Init = &cobra.Command{
	Use:              "init",
	Short:            "\tInitial Tile or Deployment.",
	Long:             "\tInitial Tile or Deployment with all basic needs.",
	TraverseChildren: true,
	//Run: func(c *cobra.Command, args []string) {
	//	fmt.Println("Tile")
	//},
}

var Tile = &cobra.Command{
	Use:   "tile",
	Short: "\tInitial Tile with basic template.",
	Long:  "\tInitial Tile to kick off building a nwe Tile, with all basic needs.",
	Run: func(c *cobra.Command, args []string) {
		tileFunc(c, args)
	},
}

var SampleTile = &cobra.Command{
	Use:   "sample-tile",
	Short: "\tInitial Sample Tile with basic template.",
	Long:  "\tInitial Sample Tile to kick off building a nwe Tile, with all basic needs.",
	Run: func(c *cobra.Command, args []string) {
		sampleTileFunc(c, args)
	},
}

var Deployment = &cobra.Command{
	Use:   "deployment",
	Short: "\tInitial Deployment with basic template.",
	Long:  "\tInitial Deployment to start a nwe deployment, with all basic needs.",
	Run: func(c *cobra.Command, args []string) {
		deploymentFunc(c, args)
	},
}

func init() {
	Init.PersistentFlags().StringP("name", "n","", "The name of building Tile/Deployment")
	Init.PersistentFlags().String("version", "0.0.1", "The version of building Tile/Deployment")
	Init.PersistentFlags().String("directory", ".", "Where to place the Tile/Deployment with templates")

	Init.AddCommand(Tile, SampleTile, Deployment)

}

func sampleTileFunc(c *cobra.Command, args []string) {
	log.Printf("Loading Sample Tile from Tile-Repo started ...")

	destDir := "./sample-tile/0.1.0"
	uri := "v1alpha1/template/tile"
	addr, _ := c.Flags().GetString("addr")
	url, err := cmd.RunGet(addr, uri)
	if err != nil {
		log.Printf("Loading Sample Tile was failed with Err: %s. \n", err)
		return
	}

	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: config}
	httpClient := &http.Client{Transport: tr}

	req, err := http.NewRequest(http.MethodGet, string(url), nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Downloading was failed from %s with Err: %s. \n", string(url), err)
		return
	}

	err = cmd.UnTarGz(destDir, bufio.NewReader(resp.Body))
	if err != nil {
		log.Printf("Unzip tar was failed with Err: %s \n", err.Error())
	}
	log.Printf("Loading Sample Tile finished")
}

func tileFunc(c *cobra.Command, args []string) {

}

func deploymentFunc(c *cobra.Command, args []string) {

}
