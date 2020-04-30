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

var s3Config *utils.S3Config

func init() {
	// TODO: load from config
	s3Config = &utils.S3Config {
		WorkHome: "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace",
		Region: "",
		BucketName: "",
		Mode: "dev",
		LocalRepo: "/Users/chuancc/mywork/mylabs/csdc/mahjong-0/tiles-repo",
	}
}

// GenerateCdkApp return path where the base CDK App was generated.
func (d *Deployment) GenerateCdkApp(out *websocket.Conn) (*ExecutionPlan, error) {

	// 1. Loading Super from s3 & unzip
	// 2. Loading Tiles from s3 & unzip
	var tsAll = &Ts{}
	var ep *ExecutionPlan
	SendResponse(out, []byte("Loading Super ... from RePO."))
	_, err := s3Config.LoadSuper()
	if err != nil {
		return ep, err
	}
	SendResponse(out, []byte("Loading Super ... from RePO with success."))

	switch d.Spec.Template.Category {
	case Network.CString():
		break
	case Compute.CString():
		break
	case ContainerProvider.CString():
		for _, t := range d.Spec.Template.tiles {
			if err := d.PullTile(t.TileReference, t.TileVersion, out, tsAll); err != nil {
				return ep, err
			}
		}
		break
	case Storage.CString():
		break
	case Database.CString():
		break
	case Application.CString():
		break
	case ContainerApplication.CString():
		break
	case Analysis.CString():
		break
	case ML.CString():
		break
	}

	// 3. Generate super.ts
	if err := d.ApplyMainTs(out, tsAll); err != nil {
		return ep, err
	}

	//4. Generate execution plan
	return d.GenerateExecutePlan(out, tsAll)

}

func (d *Deployment) PullTile(tile string, version string, out *websocket.Conn, tsAll *Ts) error {

	// 1. Loading Tile from s3 & unzip
	tileSpecFile, err := s3Config.LoadTile(tile, version)
	if err != nil {
		SendResponsef(out, "Failed to pulling Tile < %s - %s > ... from RePO\n", tile, version)
	} else {
		SendResponsef(out, "Pulling Tile < %s - %s > ... from RePO with success\n", tile, version)
	}

	//parse tile-spec.yaml if need more tile
	SendResponsef(out, "Parsing spec of Tile: < %s - %s > ...\n", tile, version)
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
				if err = d.PullTile(t.TileReference, t.TileVersion, out, tsAll); err != nil {
					return err
				}
			}
		}

	}
	SendResponsef(out, "Parsing spec of Tile: < %s - %s > was success.\n", tile, version)
	return nil
}

// TODO: Simplify & refactor !!!
func (d *Deployment) ApplyMainTs(out *websocket.Conn, tsAll *Ts) error {
	superts := s3Config.WorkHome + "/super/bin/super.ts"
	SendResponse(out, []byte("Generating main.ts for Super ..."))

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
	SendResponse(out, []byte("Generating main.ts for Super ... with success"))
	SendResponse(out, buf)
	return nil
}

func (d *Deployment) GenerateExecutePlan(out *websocket.Conn, tsAll *Ts) (*ExecutionPlan, error) {
	SendResponse(out, []byte("Generating execution plan... "))
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
		if ts.TileCategory == ContainerApplication.CString() {
			switch ts.TsManifests.ManifestType {
			case K8s.MTString():
				for j, f := range ts.TsManifests.Files {
					cmd := "kubectl -f ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f
					stage.Command.PushFront(cmd)
					stage.CommandMirror[strconv.Itoa(j)] = cmd
				}
				break
			case Helm.MTString():
				// TODO: not quite yet to support Helm
				break
			case Kustomize.MTString():
				// TODO: not quite yet to support Kustomize
				break
			}

		} else {
			stage.Kind = ts.TileCategory
			stage.WorkHome = s3Config.WorkHome
			stage.Preparation = append(stage.Preparation, "cd " + s3Config.WorkHome)
			stage.Preparation = append(stage.Preparation, "npm install")
			stage.Preparation = append(stage.Preparation, "npm run build")
			stage.Preparation = append(stage.Preparation, "cdk list")
			cmd := "cdk deploy " + ts.TileStackName + " --require-approval never"
			stage.Command.PushFront(cmd)
			stage.CommandMirror[strconv.Itoa(1)] = cmd
		}
		p.Plan.PushFront(stage)
		p.PlanMirror[strconv.Itoa(i)] = stage
	}

	buf, _ := yaml.Marshal(p)
	err := SendResponse(out, buf)
	SendResponse(out, []byte("Generating execution plan... with success"))
	return &p, err
}
