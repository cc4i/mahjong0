package engine

import (
	"dice/utils"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"strings"
)

// AssemblerCore represents a group of functions to assemble CDK App.
type AssemblerCore interface {
	// Generate CDK App from base template with necessary tiles
	GenerateCdkApp(out *websocket.Conn)(string, error)

	// Pull Tile from repo
	PullTile(name string, version string, out *websocket.Conn) error

	// Validate Tile as per tile-spec.yaml: name='folder' version='version_folder'
	ValidateTile(name string, version string, out *websocket.Conn) error

	//Generate Main Ts inside of CDK app
	ApplyMainTs(out *websocket.Conn) error

	//Generate execution plan to direct provision resources
	GenerateExecutePlan(out *websocket.Conn) error

}

// GenerateCdkApp return path where the base CDK App was generated.
func (d *Deployment) GenerateCdkApp(out *websocket.Conn)(string, error) {

	// 1. Pull super.zip from s3 & unzip
	// 2. Pull tiles.zip from s3
	repoDir := "../tiles-repo/"
	destDir := "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace/"
	if err := utils.Copy(repoDir+"super", destDir+"super"); err != nil {
		return destDir, err
	}
	// Network
	for _, t := range d.Spec.Template.Network {
		sendResponse(out, []byte(t.TileReference))
	}
	// Compute
	for _, t := range d.Spec.Template.Compute {
		sendResponse(out, []byte(t.TileReference))
	}
	// Container
	for _, t := range d.Spec.Template.Container {
		if err := d.PullTile(destDir+"super", t.TileReference, t.TileVersion, out); err != nil {
			return destDir, err
		}
	}
	// Database
	for _, t := range d.Spec.Template.Database {
		sendResponse(out, []byte(t.TileReference))
	}
	// Application
	for _, t := range d.Spec.Template.Application {
		sendResponse(out, []byte(t.TileReference))
	}
	// Analysis
	for _, t := range d.Spec.Template.Analysis {
		sendResponse(out, []byte(t.TileReference))
	}
	// ML
	for _, t := range d.Spec.Template.ML {
		sendResponse(out, []byte(t.TileReference))
	}

	// 3. Generate super.ts


	return destDir, nil
}

func (d *Deployment) PullTile(to string, tile string, version string, out *websocket.Conn) error {

	// 1. Download tile from s3 & unzip
	repoDir := "../tiles-repo/"
	srcDir := repoDir+ strings.ToLower(tile)+"/"+strings.ToLower(version)
	destDir := to+"/lib/"+strings.ToLower(tile)
	tileSpecFile := destDir+"/tile-spec.yaml"

	sendResponsef(out, "Pulling Tile < %s - %s > ...\n", tile, version)

	if err := utils.Copy(srcDir, destDir,
		utils.Options{
			OnSymlink: func(src string) utils.SymlinkAction {
				return utils.Skip
			},
			Skip: func(src string) bool {
				return strings.Contains(src, "node_modules")
			},
		}); err != nil {
		return err
	}

	//parse tile-spec.yaml if need more tile
	if buf, err := ioutil.ReadFile(tileSpecFile); err != nil {
		return err
	} else {
		data := Data(buf)
		if tile, err := data.ParseTile(); err != nil {
			return err
		} else {
			for _, t := range tile.Spec.Dependencies {
				if err =d.PullTile(to, t.TileReference, t.TileVersion, out); err != nil {
					return err
				}
			}
		}

	}
	sendResponsef(out, "Pulling Tile < %s - %s > was success.\n", tile, version)
	return nil
}


// TODO: ApplyMainTs
func (d *Deployment)ApplyMainTs(out *websocket.Conn) error {
	//destDir := "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace/"
	//superts := destDir + "bin/super.ts"



	return nil
}

func sendResponse(out *websocket.Conn, response []byte) error {
	log.Printf("%s\n", response)
	err := out.WriteMessage(websocket.TextMessage, response)
	if err !=nil {
		log.Printf("write error: %s\n", err)
	}
	return err

}

func sendResponsef(out *websocket.Conn, format string, v ...interface{}) error {
	log.Printf(format, v...)
	err := out.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(format, v...)))
	if err !=nil {
		log.Printf("write error: %s\n", err)
	}
	return err

}