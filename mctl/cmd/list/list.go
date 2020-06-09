package list

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mctl/cmd"
)

var TilesInRepo = &cobra.Command{
	Use:   "tile",
	Short: "\tList Tiles in the Repo.",
	Long:  "\tList Tiles in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		log.Println("Tile")
	},
}

var HuInRepo = &cobra.Command{
	Use:   "hu",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		log.Println("Hu")
	},
}

var SchemaInRepo = &cobra.Command{
	Use:   "schema",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		log.Println("schema")
	},
}

var Deployment = &cobra.Command{
	Use:   "deployment",
	Short: "\tList all deployment has been triggered.",
	Long:  "\tList all deployment has been triggered.",
	Run: func(c *cobra.Command, args []string) {
		addr, _ := c.Flags().GetString("addr")
		buf, err := cmd.RunGetByVersion(addr, "ts")
		if err != nil {
			log.Printf("%s\n",err)
		}
		log.Printf("\n--------- Deployment Records --------- \n%s--------- ---------------- ---------\n", string(buf))
	},
}

var Repo = &cobra.Command{
	Use:   "list",
	Short: "\tList content (Tiles, Hu, etc.) in the Repo.",
	Long:  "\tList content (Tiles, Hu, etc.) in the Repo.",
}

func init() {
	Repo.AddCommand(TilesInRepo, HuInRepo, SchemaInRepo, Deployment)
}
