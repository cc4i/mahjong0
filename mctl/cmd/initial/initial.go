package initial

import (
	"bufio"
	"crypto/tls"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"log"
	"mctl/cmd"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
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

var Deployment = &cobra.Command{
	Use:   "deployment",
	Short: "\tInitial Deployment with basic template.",
	Long:  "\tInitial Deployment to start a nwe deployment, with all basic needs.",
	Run: func(c *cobra.Command, args []string) {
		deploymentFunc(c, args)
	},
}

// TileTemplateData used to replace template
type TileTemplateData struct {
	TileName          string // Tile Name
	TileNameLowerCase string // Tile Name >>> Tile Name with Lower Case
	TileNameCDK       string // TileName >>> Tile Name with Camel Case

}

func init() {
	Init.PersistentFlags().StringP("name", "n", "", "The name of building Tile/Deployment")
	Init.PersistentFlags().StringP("type", "t", "cdk", "The type of building Tile, cdk - cdk tiles/app - app tiles.")
	Init.PersistentFlags().StringP("version", "v", "0.1.0", "The version of building Tile/Deployment")
	Init.PersistentFlags().StringP("directory", "d", ".", "Where to place the Tile/Deployment with templates")

	Init.AddCommand(Tile, Deployment)

}

func tileFunc(c *cobra.Command, args []string) {
	name, _ := c.Flags().GetString("name")
	if destDir, err := download(c, args, name); err != nil {
		return
	} else {
		if name != "sample-tile" {
			tt := &TileTemplateData{
				TileName:          name,
				TileNameLowerCase: strings.ToLower(name),
				TileNameCDK:       strcase.ToCamel(name),
			}
			filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() && strings.HasSuffix(path, ".tp") {
					//tp = tp.New("x")
					tp, err := template.ParseFiles(path)
					if err != nil {
						log.Printf("%s\n", err)
					}
					f, err := os.Create(strings.TrimSuffix(path, ".tp"))
					if err != nil {
						log.Printf("%s\n", err)
					}

					err = tp.Execute(f, tt)
					if err != nil {
						log.Printf("file: %s : %s\n", path, err)
					}
					defer f.Close()
					os.Remove(path)
				}
				log.Printf("Generated file - %s \n", path)
				return nil
			})

		}

	}

}

func deploymentFunc(c *cobra.Command, args []string) {

	deploymentExample:=`
apiVersion: mahjong.io/v1alpha1
kind: Deployment 
metadata:
  name: simple-eks
spec:
  template:
    tiles:
      tileEKS005:
        tileReference: Eks0
        tileVersion: 0.0.5
        inputs:
          - name: clusterName
            inputValue: simple-eks-cluster
          - name: capacity
            inputValue: 2
          - name: capacityInstance
            inputValue: m5.large
          - name: version
            inputValue: 1.16
  summary:
    description: 
    outputs:
      - name: EKS Cluster Name
        valueRef: $(tileEKS005.outputs.clusterName)
      - name: Master role arn for EKS Cluster
        valueRef: $(tileEKS005.outputs.masterRoleARN)
      - name: The API endpoint EKS Cluster
        valueRef: $(tileEKS005.outputs.clusterEndpoint)
      - name: Instance type of worker node
        valueRef: $(tileEKS005.outputs.capacityInstance)
      - name: Default capacity of worker node
        valueRef: $(tileEKS005.outputs.capacity)

    notes: []

`
	log.Println("Generating simple example for deployment ... ...")
	file, err := os.Create("simple-eks.yaml")
	if err!=nil{
		log.Printf("Generating simple example for deployment was failed, with error: %s\n", err)
	}
	defer file.Close()
	_, err = file.Write([]byte(deploymentExample))
	if err != nil {
		log.Printf("Generating simple example for deployment was failed, with error: %s\n", err)
	}
	log.Printf("Generated a simple example. ")
	log.Printf("Download https://github.com/cc4i/mahjong0/blob/master/templates/deployment-schema.json for schema, and jump to https://github.com/cc4i/mahjong0#examples for more examples.\n")
}

func download(c *cobra.Command, args []string, name string) (string, error) {
	log.Printf("Loading %s templates from Tile-Repo started ...", name)

	addr, _ := c.Flags().GetString("addr")

	t, _ := c.Flags().GetString("type")
	uri := "v1alpha1/template/"
	if name == "sample-tile" {
		uri = uri + name
	} else {
		if t != "cdk" {
			uri = uri + "tile?type=app"
		} else {
			uri = uri + "tile?type=cdk"
		}
	}

	destDir := "/" + name + "/"
	oDir, _ := c.Flags().GetString("directory")
	if oDir == "" {
		destDir = "." + destDir
	} else {
		destDir = oDir + destDir
	}

	version, _ := c.Flags().GetString("version")
	destDir = destDir + version

	url, err := cmd.RunGet(addr, uri)
	if err != nil {
		log.Printf("Loading %s was failed with Err: %s. %s\n", name, err, url)
		return destDir, err
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
		return destDir, err
	}

	err = cmd.UnTarGz(destDir, bufio.NewReader(resp.Body))
	if err != nil {
		log.Printf("Unzip tar was failed with Err: %s \n", err.Error())
		return destDir, err
	}

	log.Printf("Loading %s templates finished.", name)
	return destDir, nil
}
