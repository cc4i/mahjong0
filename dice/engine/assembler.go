package engine

import (
	"container/list"
	"dice/utils"
	"fmt"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"text/template"
)

// Ts is key struct to fulfil super.ts template and key element to generate execution plan.
type Ts struct {
	TsLibs   []TsLib
	TsStacks []TsStack
	AllTiles map[string]Tile // "Category - TileName" -> Tile
}

type TsLib struct {
	TileName   string
	TileFolder string
	TileCategory string
}

type TsStack struct {
	TileName          string
	TileVariable      string
	TileStackName     string
	TileStackVariable string
	TileCategory string
	InputParameters   map[string]string //[]TsInputParameter
	TsManifests       *TsManifests
}

type TsInputParameter struct {
	InputName  string
	InputValue string
}

type TsManifests struct {
	ManifestType string
	Namespace string
	Files        []string
	Folders      []string
	VendorService string
	DependentTile string
	DependentTileVersion string
}


// AssemblerCore represents a group of functions to assemble CDK App.
type AssemblerCore interface {
	// Generate CDK App from base template with necessary tiles
	GenerateCdkApp(out *websocket.Conn) (*ExecutionPlan, error)

	// Pull Tile from repo
	PullTile(name string, version string, out *websocket.Conn, tsAll *Ts) error

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
	var override = make(map[string]*TileInputOverride) //TileName->TileInputOverride
	var ep *ExecutionPlan
	SendResponse(out, []byte("Loading Super ... from RePO."))
	_, err := s3Config.LoadSuper()
	if err != nil { return ep, err }
	SendResponse(out, []byte("Loading Super ... from RePO with success."))

	switch d.Spec.Template.Category {
	case Network.CString(), Compute.CString(), ContainerProvider.CString(), Storage.CString(), Database.CString(),
			Application.CString(), ContainerApplication.CString(), Analysis.CString(), ML.CString():
		for _, t := range d.Spec.Template.Tiles {
			if err := d.PullTile(t.TileReference, t.TileVersion, out, tsAll, override); err != nil {
				return ep, err
			}
		}
	}


	// 3. Generate super.ts
	if err := d.ApplyMainTs(out, tsAll); err != nil {
		return ep, err
	}

	//4. Generate execution plan
	return d.GenerateExecutePlan(out, tsAll)

}


func (d *Deployment) PullTile(tile string, version string, out *websocket.Conn, tsAll *Ts, override map[string]*TileInputOverride) error {

	// 1. Loading Tile from s3 & unzip
	tileSpecFile, err := s3Config.LoadTile(tile, version)
	if err != nil {
		SendResponsef(out, "Failed to pulling Tile < %s - %s > ... from RePO\n", tile, version)
	} else {
		SendResponsef(out, "Pulling Tile < %s - %s > ... from RePO with success\n", tile, version)
	}

	//parse tile-spec.yaml if need more tile
	SendResponsef(out, "Parsing specification of Tile: < %s - %s > ...\n", tile, version)
	buf, err := ioutil.ReadFile(tileSpecFile)
	if err != nil { return err }
	data := Data(buf)
	parsedTile, err := data.ParseTile()
	if err != nil  { return err }
	// TODO: to be refactor
	// Step 0. Caching the tile
	if tsAll.AllTiles == nil { tsAll.AllTiles = make(map[string]Tile) }
	tsAll.AllTiles[parsedTile.Metadata.Category + "-" + parsedTile.Metadata.Name] = *parsedTile
	////

	// Step 1. Caching deployment inputs
	deploymentInputs := make(map[string][]string) //tileName -> map[inputName]inputValues
	for _, tts := range d.Spec.Template.Tiles {
		for _, n := range tts.Inputs {

			if len(n.InputValues)>0 {
				deploymentInputs[tts.TileReference+"-"+n.Name]=n.InputValues
			} else {
				deploymentInputs[tts.TileReference+"-"+n.Name]=[]string{n.InputValue}
			}

		}
	}
	////

	// Step 2. Caching tile dependencies for further process
	dependenciesMap := make(map[string]string)
	for _, m := range parsedTile.Spec.Dependencies {
		dependenciesMap[m.Name] = m.TileReference
	}

	// Step 3. Caching tile override for further process; depends on Step 2
	for _, ov := range parsedTile.Spec.Inputs {
		if ov.Override.Name != "" {
			if _, ok := dependenciesMap[ov.Override.Name]; ok {
				override[dependenciesMap[ov.Override.Name]+"-"+ov.Override.Field] = &TileInputOverride {
					Name: ov.Override.Name,
					Field: ov.Override.Field,
					InputName: ov.Name,
				}
				//
			}
		}
	}
	////

	////
	// Step 4. Store import libs && avoid to add repeated one
	newTsLib := TsLib{
		TileName:   parsedTile.Metadata.Name,
		TileFolder: strings.ToLower(parsedTile.Metadata.Name),
		TileCategory: parsedTile.Metadata.Category,
	}
	if !containsTsLib(&newTsLib, tsAll.TsLibs) {
		tsAll.TsLibs = append(tsAll.TsLibs, newTsLib)
	}
	////

	// Step 5. Caching inputs <key, value> for further process
	// inputs: inputName -> inputValue
	inputs := make(map[string]string)
	for _, in := range parsedTile.Spec.Inputs {
		input := TsInputParameter{}
		// For value dependent on other Tile
		if in.Dependencies != nil {
			if len(in.Dependencies) == 1 {
				// single dependency
				input.InputName = in.Name
				stile := strings.ToLower(dependenciesMap[in.Dependencies[0].Name])
				input.InputValue = stile + "stack" + "var." + stile + "var." + in.Dependencies[0].Field
			} else {
				// multiple dependencies will be organized as an array
				input.InputName = in.Name
				v := "[ "
				for _, d := range in.Dependencies {
					stile := strings.ToLower(dependenciesMap[d.Name])
					val := stile + "stack" + "var." + stile + "var." + d.Field
					v = v + val + ","
				}
				input.InputValue = strings.TrimSuffix(v, ",") + " ]"
			}
			// For independent value
		} else {
			input.InputName = in.Name
			// Overwrite values as per Deployment
			if val, ok := deploymentInputs[parsedTile.Metadata.Name+"-"+in.Name]; ok {

				if len(val) >1 {
					v := "[ "
					for _, d := range val {
						if strings.Contains(in.InputType, String.IOTString()) {
							v = v + "'" + d + "',"
						} else {
							v = v + d + ","
						}

					}
					input.InputValue = strings.TrimSuffix(v, ",") + " ]"

				} else {
					if strings.Contains(in.InputType, String.IOTString()) {
						input.InputValue = "'" + val[0] + "'"
					} else {
						input.InputValue = val[0]
					}
				}

			} else {
				if in.DefaultValues != nil {
					vals := "[ "
					for _, d := range in.DefaultValues {
						if strings.Contains(in.InputType, String.IOTString()) {
							vals = vals + "'" + d + "',"
						} else {
							vals = vals + d + ","
						}

					}
					input.InputValue = strings.TrimSuffix(vals, ",") + " ]"

				} else if len(in.DefaultValue) > 0 {
					if strings.Contains(in.InputType, String.IOTString()) {
						input.InputValue = "'" + in.DefaultValue + "'"
					} else {
						input.InputValue = in.DefaultValue
					}
				}

			}
		}
		//lookup override
		if or, ok := override[parsedTile.Metadata.Name+"-"+input.InputName]; ok {
			if input.InputName == or.Field {
				input.InputValue = or.OverrideValue
			}
		}
		inputs[input.InputName] = input.InputValue
	}
	////

	// Step 6.Setup values for cached override, depend on Step 5
	for _, v := range override {
		if val, ok := inputs[v.InputName]; ok {
			v.OverrideValue = val
		}
	}
	////

	// Step 7. Caching manifest & overwrite
	// Overwrite namespace as deployment
	tm := &TsManifests{
		ManifestType: parsedTile.Spec.Manifests.ManifestType,
	}
	ns := ""
	for _, m := range d.Spec.Template.Tiles {
		if m.TileReference == parsedTile.Metadata.Name {
			ns = m.Manifests.Namespace
		}
	}
	if ns=="" { ns = parsedTile.Spec.Manifests.Namespace}
	tm.Namespace = ns

	// Overwrite files/folders as deployment
	var ffs []string
	var fds []string
	for _, m := range d.Spec.Template.Tiles {
		if m.TileReference == parsedTile.Metadata.Name {
			if m.Manifests.Files != nil { ffs = m.Manifests.Files }
			if m.Manifests.Folders != nil { fds = m.Manifests.Folders }
		}
	}
	if ffs==nil  {ffs = parsedTile.Spec.Manifests.Files}
	for _, m := range parsedTile.Spec.Manifests.Files {
		tm.Files = append(tm.Files, m)
	}
	if fds==nil  {fds = parsedTile.Spec.Manifests.Folders}
	for _, m := range parsedTile.Spec.Manifests.Folders {
		tm.Folders = append(tm.Folders, m)
	}
	////

	// Step 8. Store import Stacks && avoid repeated one
	ts := &TsStack {
		TileName:          parsedTile.Metadata.Name,
		TileVariable:      strings.ToLower(parsedTile.Metadata.Name + "var"),
		TileStackName:     parsedTile.Metadata.Name + "Stack",
		TileStackVariable: strings.ToLower(parsedTile.Metadata.Name + "stack" + "var"),
		InputParameters:   inputs,
		TileCategory: parsedTile.Metadata.Category,
		TsManifests: tm,
	}
	if !containsTsStack(ts, tsAll.TsStacks) {
		tsAll.TsStacks = append(tsAll.TsStacks, *ts)
	}
	////

	// recurred call
	for _, t := range parsedTile.Spec.Dependencies {
		if err = d.PullTile(t.TileReference, t.TileVersion, out, tsAll, override); err != nil {
			return err
		}
	}
	////
	// !!!Last job: checking vendor service before leaving, do it after recurring.
	// ???
	if parsedTile.Metadata.Category == ContainerApplication.CString() {
		for k , v := range tsAll.AllTiles {
			if strings.Contains(k, ContainerProvider.CString()) {
				ts.TsManifests.VendorService = v.Metadata.VendorService
				ts.TsManifests.DependentTile = v.Metadata.Name
				ts.TsManifests.DependentTileVersion = v.Metadata.Version
			}
		}
	}
	////

	SendResponsef(out, "Parsing specification of Tile: < %s - %s > was success.\n", tile, version)
	return nil
}

func containsTsLib(slice *TsLib, tls []TsLib) bool {
	for _, tl := range tls {
		if slice.TileName == tl.TileName && slice.TileFolder == tl.TileFolder {
			return true
		}
	}
	return false
}

func containsTsStack(slice *TsStack, tss []TsStack) bool {
	for _, ts := range tss {
		if slice.TileName == ts.TileName {
			return true
		}
	}
	return false
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
	SendResponse(out, []byte("--BO:-------------------------------------------------"))
	SendResponse(out, buf)
	SendResponse(out, []byte("--EO:-------------------------------------------------"))
	return nil
}

func (d *Deployment) GenerateExecutePlan(out *websocket.Conn, tsAll *Ts) (*ExecutionPlan, error) {
	SendResponse(out, []byte("Generating execution plan... "))
	var p = ExecutionPlan{
		Plan:       list.New(),
		PlanMirror: make(map[string]ExecutionStage),
	}
	for i, ts := range tsAll.TsStacks {
		workHome := s3Config.WorkHome + "/super"
		stage := ExecutionStage{
			Name:          ts.TileName,
			Command:       list.New(),
			CommandMirror: make(map[string]string),
			Kind:          ts.TileCategory,
			WorkHome:      workHome,
			Preparation:   []string{"cd " + workHome},
		}

		if ts.TileCategory == ContainerApplication.CString() {
			//stage.Preparation = ? how to connect to EKS
			switch ts.TsManifests.ManifestType {
			case K8s.MTString():
				if ts.TsManifests.VendorService == EKS.VSString() {
					stage.Preparation = append(stage.Preparation,
						fmt.Sprintf("aws eks update-kubeconfig --name %s --role-arn <master role arn> --kubeconfig <config>",ts.TsManifests.DependentTile))
				} else if ts.TsManifests.VendorService == Kubernetes.VSString() {
					// To be address
				}
				stage.Preparation = append(stage.Preparation,"")
				for j, f := range ts.TsManifests.Files {
					cmd := "kubectl -f ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + " -n " + ts.TsManifests.Namespace
					stage.Command.PushFront(cmd)
					stage.CommandMirror[strconv.Itoa(j)] = cmd
				}
			case Helm.MTString():
				// TODO: not quite yet to support Helm
				for j, f := range ts.TsManifests.Folders {
					cmd := "helm install " + ts.TileName + " ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + " -n " + ts.TsManifests.Namespace
					stage.Command.PushFront(cmd)
					stage.CommandMirror[strconv.Itoa(j)] = cmd
				}

			case Kustomize.MTString():
				// TODO: not quite yet to support Kustomize
				for j, f := range ts.TsManifests.Folders {
					cmd := "kustomize build -f ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + "|kubectl -f - " + " -n " + ts.TsManifests.Namespace
					stage.Command.PushFront(cmd)
					stage.CommandMirror[strconv.Itoa(j)] = cmd
				}
			}

		} else if ts.TileCategory == Application.CString() {
			//TODO: What to do with application?
		} else {

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

// extractValues