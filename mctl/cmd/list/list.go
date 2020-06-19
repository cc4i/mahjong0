package list

import (
	"encoding/json"
	"fmt"
	"github.com/kris-nova/logger"
	"github.com/spf13/cobra"
	"mctl/cmd"
	"time"
)

// deployment records
type deployment struct {
	SID         string    `json:"SID"`         // session ID for each deployment
	Name        string    `json:"Name"`        //Unique name for each deployment
	Created     time.Time `json:"Created"`     // Created time
	Updated     time.Time `json:"Updated"`     // Last update
	SuperFolder string    `json:"SuperFolder"` // Main folder for all stuff per deployment
	Status      string    `json:"Status"`      // Status of deployment
}

var TilesInRepo = &cobra.Command{
	Use:   "tile",
	Short: "\tList Tiles in the Repo.",
	Long:  "\tList Tiles in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		addr, _ := c.Flags().GetString("addr")
		buf, err := cmd.RunGetByVersion(addr, "repo/tile")
		if err != nil {
			logger.Warning("%s\n", err)
			return
		}
		output, _ := c.Flags().GetString("output")
		if output != "" {
			if output == "json" {
				fmt.Printf("%s\n", string(buf))
			} else {
				logger.Warning("unsupported format")
			}
		} else {
			var tms []TileMetadata
			err = json.Unmarshal(buf, &tms)
			if err != nil {
				logger.Warning("%s\n", err)
				return
			} else {
				logger.Info("%s\t\t %s\t\t %s\t\t %s\n", "Name", "Version", "License", "Released")
				for _, tm := range tms {
					logger.Info("%s\t\t %s\t\t %s\t\t %s\n", tm.Name, tm.Version, tm.License, tm.Released.Local().Format("2006-01-02 15:04:05"))
				}

			}
		}

	},
}

var HuInRepo = &cobra.Command{
	Use:   "hu",
	Short: "\tList Hu in the Repo.",
	Long:  "\tList Hu in the Repo.",
	Run: func(c *cobra.Command, args []string) {
		addr, _ := c.Flags().GetString("addr")
		buf, err := cmd.RunGetByVersion(addr, "repo/hu")
		if err != nil {
			logger.Warning("%s\n", err)
			return
		}
		output, _ := c.Flags().GetString("output")
		if output != "" {
			if output == "json" {
				fmt.Printf("%s\n", string(buf))
			} else {
				logger.Warning("unsupported format")
			}
		} else {
			var tms []TileMetadata
			err = json.Unmarshal(buf, &tms)
			if err != nil {
				logger.Warning("%s\n", err)
				return
			} else {
				logger.Info("%s\t\t %s\t\t %s\t\t %s\n", "Name", "Version", "License", "Released")
				for _, tm := range tms {
					logger.Info("%s\t\t %s\t\t %s\t\t %s\n", tm.Name, tm.Version, tm.License, tm.Released.Local().Format("2006-01-02 15:04:05"))
				}

			}
		}

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
			logger.Warning("%s\n", err)
			return
		}
		output, _ := c.Flags().GetString("output")
		if output != "" {
			if output == "json" {
				fmt.Printf("%s\n", string(buf))
			} else {
				logger.Warning("unsupported format")
			}

		} else {
			err = json.Unmarshal(buf, &dr)
			if err != nil {
				logger.Warning("\n--------- Deployment Records --------- \n%s\n--------- ---------------- ---------\n", string(buf))
			} else {
				logger.Info("--------- Deployment Records --------- \n")
				logger.Info("Name\t\t Created time\t\t Last update\t\t Folder\t\t Status\n")
				for _, d := range dr {
					logger.Info("%s\t\t %s\t\t %s\t\t %s\t\t %s\n", d.Name,
						d.Created.Local().Format("2006-01-02 15:04:05"),
						d.Updated.Local().Format("2006-01-02 15:04:05"),
						d.SuperFolder, d.Status)
				}

				logger.Info("--------- ---------------- -----------\n")

			}
		}

	},
}

var Repo = &cobra.Command{
	Use:   "list",
	Short: "\tList content (Tiles, Hu, etc.) in the Repo.",
	Long:  "\tList content (Tiles, Hu, etc.) in the Repo.",
}

func init() {
	Repo.PersistentFlags().StringP("output", "o", "", "output format: json ")
	Repo.AddCommand(TilesInRepo, HuInRepo, Deployment)
}
