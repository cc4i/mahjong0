package list

import (
	"encoding/json"
	"github.com/kris-nova/logger"
	"github.com/spf13/cobra"
	"mctl/cmd"
	"time"
)

// deployment records
type deployment struct {
	SID string `json:"SID"` // session ID for each deployment
	Name string	`json:"Name"`//Unique name for each deployment
	CreatedTime  time.Time `json:"CreatedTime"`           // Created time
	SuperFolder	string	`json:"SuperFolder"` // Main folder for all stuff per deployment
	Status string	`json:"Status"` // Status of deployment
}

var TilesInRepo = &cobra.Command{
	Use:   "tile",
	Short: "\tList Tiles in the Repo.",
	Long:  "\tList Tiles in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		logger.Info("Tile")
	},
}

var HuInRepo = &cobra.Command{
	Use:   "hu",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		logger.Info("Hu")
	},
}

var SchemaInRepo = &cobra.Command{
	Use:   "schema",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		logger.Info("schema")
	},
}

var Deployment = &cobra.Command{
	Use:   "deployment",
	Short: "\tList all deployment has been triggered.",
	Long:  "\tList all deployment has been triggered.",
	Run: func(c *cobra.Command, args []string) {
		var dr []deployment
		addr, _ := c.Flags().GetString("addr")
		buf, err := cmd.RunGetByVersion(addr, "ts")
		if err != nil {
			logger.Warning("%s\n",err)
			return
		}
		err = json.Unmarshal(buf, &dr)
		if err != nil {
			logger.Warning("\n--------- Deployment Records --------- \n%s\n--------- ---------------- ---------\n", string(buf))
		} else {
			logger.Info("\n--------- Deployment Records --------- \n")
			logger.Info("SID\t\t Name\t\t Created Time\t\t Folder\t\t Status\t\t \n")
			for _, d := range dr {
				logger.Info("%s\t%s\t%s\t%s\t%s",d.SID, d.Name, d.CreatedTime, d.SuperFolder, d.Status)
			}

			logger.Info("\n--------- ---------------- -----------\n")

		}
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
