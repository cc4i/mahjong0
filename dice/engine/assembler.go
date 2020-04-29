package engine

import (
	"container/list"
	"dice/utils"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"
)

// AssemblerCore represents a group of functions to assemble CDK App.
type AssemblerCore interface {
	// Generate CDK App from base template with necessary tiles
	GenerateCdkApp(out *websocket.Conn) (*ExecutionPlan, error)

	// Pull Tile from repo
	PullTile(name string, version string, out *websocket.Conn, tsAll *Ts) error

	// Validate Tile as per tile-spec.yaml: name='folder' version='version_folder'
	ValidateTile(name string, version string, out *websocket.Conn) error

	//Generate Main Ts inside of CDK app
	ApplyMainTs(out *websocket.Conn, tsAll *Ts) error

	//Generate execution plan to direct provision resources
	GenerateExecutePlan(out *websocket.Conn, tsAll *Ts) (*ExecutionPlan, error)
}

// GenerateCdkApp return path where the base CDK App was generated.
func (d *Deployment) GenerateCdkApp(out *websocket.Conn) (*ExecutionPlan, error) {

	// 1. Pull super.zip from s3 & unzip
	// 2. Pull tiles.zip from s3
	var tsAll = &Ts{}
	var ep *ExecutionPlan

	repoDir := "../tiles-repo/"
	destDir := "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace/"
	if err := utils.Copy(repoDir+"super", destDir+"super"); err != nil {
		return ep, err
	}
	// Network
	for _, t := range d.Spec.Template.Network {
		SendResponse(out, []byte(t.TileReference))
	}
	// Compute
	for _, t := range d.Spec.Template.Compute {
		SendResponse(out, []byte(t.TileReference))
	}
	// Container
	for _, t := range d.Spec.Template.Container {
		if err := d.PullTile(destDir+"super", t.TileReference, t.TileVersion, out, tsAll); err != nil {
			return ep, err
		}
	}
	// Database
	for _, t := range d.Spec.Template.Database {
		SendResponse(out, []byte(t.TileReference))
	}
	// Application
	for _, t := range d.Spec.Template.Application {
		SendResponse(out, []byte(t.TileReference))
	}
	// Analysis
	for _, t := range d.Spec.Template.Analysis {
		SendResponse(out, []byte(t.TileReference))
	}
	// ML
	for _, t := range d.Spec.Template.ML {
		SendResponse(out, []byte(t.TileReference))
	}

	// 3. Generate super.ts
	if err := d.ApplyMainTs(out, tsAll); err != nil {
		return ep, err
	}

	//4. Generate execution plan
	return d.GenerateExecutePlan(out, tsAll)

}

func (d *Deployment) PullTile(to string, tile string, version string, out *websocket.Conn, tsAll *Ts) error {

	// 1. Download tile from s3 & unzip
	repoDir := "../tiles-repo/"
	srcDir := repoDir + strings.ToLower(tile) + "/" + strings.ToLower(version)
	destDir := to + "/lib/" + strings.ToLower(tile)
	tileSpecFile := destDir + "/tile-spec.yaml"

	SendResponsef(out, "Pulling Tile < %s - %s > ...\n", tile, version)

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
			// Ref -> Tile
			dependenciesMap := make(map[string]string)
			for _, m := range tile.Spec.Dependencies {
				dependenciesMap[m.Name] = m.TileReference
			}

			tsAll.TsLibs = append(tsAll.TsLibs, TsLib{
				TileName:   tile.Metadata.Name,
				TileFolder: strings.ToLower(tile.Metadata.Name),
			})
			inputs := make(map[string]string) //[]TsInputParameter  {}
			for _, in := range tile.Spec.Inputs {
				input := TsInputParameter{}
				if in.Require == "true" {
					if in.Dependencies != nil {
						if len(in.Dependencies) == 1 {
							input.InputName = in.Name
							stile := strings.ToLower(dependenciesMap[in.Dependencies[0].Name])
							input.InputValue = stile + "stack" + "var." + stile + "var." + in.Dependencies[0].Field
						} else {
							input.InputName = in.Name
							vals := "{ "
							for _, d := range in.Dependencies {
								stile := strings.ToLower(dependenciesMap[d.Name])
								val := stile + "stack" + "var." + stile + "var." + d.Field
								vals = vals + val + ","
							}
							input.InputValue = strings.TrimSuffix(vals, ",") + " }"
						}

					} else {
						input.InputName = in.Name
						if in.DefaultValues != nil {
							vals := "{ "
							for _, d := range in.DefaultValues {
								if in.InputType == "string" || in.InputType == "string[]" {
									vals = vals + "'" + d + "',"
								} else {
									vals = vals + d + ","
								}

							}
							input.InputValue = strings.TrimSuffix(vals, ",") + " }"

						} else if len(in.DefaultValue) > 0 {
							if in.InputType == "string" || in.InputType == "string[]" {
								input.InputValue = "'" + in.DefaultValue + "'"
							} else {
								input.InputValue = in.DefaultValue
							}

						}

					}
					inputs[input.InputName] = input.InputValue //append(inputs, input)

				}
			}

			tsAll.TsStacks = append(tsAll.TsStacks, TsStack{
				TileName:          tile.Metadata.Name,
				TileVariable:      strings.ToLower(tile.Metadata.Name + "var"),
				TileStackName:     tile.Metadata.Name + "Stack",
				TileStackVariable: strings.ToLower(tile.Metadata.Name + "stack" + "var"),
				InputParameters:   inputs,
			})

			for _, t := range tile.Spec.Dependencies {
				if err = d.PullTile(to, t.TileReference, t.TileVersion, out, tsAll); err != nil {
					return err
				}
			}
		}

	}
	SendResponsef(out, "Pulling Tile < %s - %s > was success.\n", tile, version)
	return nil
}

// TODO: Simplify & refactor !!!
func (d *Deployment) ApplyMainTs(out *websocket.Conn, tsAll *Ts) error {
	destDir := "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace/"
	superts := destDir + "super/bin/super.ts"

	tp, _ := template.ParseFiles(superts)

	file, err := os.Create(superts + "_new")
	if err != nil {
		SendResponse(out, []byte(err.Error()))
		return err
	}

	//!!!reverse tsAll.TsStacks due to CDK require!!!
	for i := len(tsAll.TsStacks)/2 - 1; i >= 0; i-- {
		opp := len(tsAll.TsStacks) - 1 - i
		tsAll.TsStacks[i], tsAll.TsStacks[opp] = tsAll.TsStacks[opp], tsAll.TsStacks[i]
	}

	err = tp.Execute(file, tsAll)
	if err != nil {
		SendResponse(out, []byte(err.Error()))
		return err
	}
	err = file.Close()
	if err != nil {
		SendResponse(out, []byte(err.Error()))
		return err
	}
	os.Rename(superts, superts+"_old")
	os.Rename(superts+"_new", superts)
	buf, err := ioutil.ReadFile(superts)
	if err != nil {
		SendResponse(out, []byte(err.Error()))
		return err
	}
	SendResponse(out, buf)
	return nil
}

func (d *Deployment) GenerateExecutePlan(out *websocket.Conn, tsAll *Ts) (*ExecutionPlan, error) {
	destDir := "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace/super/"

	var p = ExecutionPlan{
		Plan:       list.New(),
		PlanMirror: make(map[string]ExecutionStage),
	}
	for i, ts := range tsAll.TsStacks {

		stage := ExecutionStage{
			Name:          ts.TileName,
			Command:       list.New(),
			CommandMirror: make(map[string]string),
		}
		if ts.TsManifests != nil {
			stage.Kind = ts.TsManifests.ManifestType
			for j, f := range ts.TsManifests.Files {
				//TODO: support other type of manifest
				switch stage.Kind {
				case "k8s_manifest":
					cmd := "kubectl -f " + destDir + "./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f
					stage.Command.PushFront(cmd)
					stage.CommandMirror[strconv.Itoa(j)] = cmd
					break
				}

			}
			//TODO: support folder with other type of manifest
			for j, f := range ts.TsManifests.Folders {

				stage.Command.PushFront(f)
				stage.CommandMirror[strconv.Itoa(j)] = f
			}
		} else {
			stage.Kind = "cdk"
			stage.WorkHome = destDir
			stage.Preparation = append(stage.Preparation, "cd "+destDir)
			stage.Preparation = append(stage.Preparation, "npm install")
			stage.Preparation = append(stage.Preparation, "npm run build")
			stage.Preparation = append(stage.Preparation, "cdk list")
			cmd := "cdk deploy " + ts.TileStackName + " --require-approval never"
			stage.Command.PushFront(cmd)
			stage.CommandMirror[strconv.Itoa(1)] = cmd
		}
		p.Plan.PushFront(stage)
		p.PlanMirror[strconv.Itoa(i)] = stage
		//ts.TileStackName
	}

	buf, _ := yaml.Marshal(p)
	err := SendResponse(out, buf)
	return &p, err
}
