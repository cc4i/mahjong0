package engine

import (
	"container/list"
	"context"
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

// AssemblerCore represents a group of functions to assemble CDK App.
type AssemblerCore interface {
	// Generate CDK App from base template with necessary tiles
	GenerateCdkApp(ctx context.Context, out *websocket.Conn) (*ExecutionPlan, error)

	// Pull Tile from repo
	PullTile(ctx context.Context, name string, version string, out *websocket.Conn, aTs *Ts) error

	//Generate Main Ts inside of CDK app
	ApplyMainTs(ctx context.Context, out *websocket.Conn, aTs *Ts) error

	//Generate execution plan to direct provision resources
	GenerateExecutePlan(ctx context.Context, out *websocket.Conn, aTs *Ts) (*ExecutionPlan, error)
}

var s3Config *utils.S3Config

func init() {
	// TODO: load from config
	s3Config = &utils.S3Config{
		WorkHome:   "/Users/chuancc/mywork/mylabs/csdc/mahjong-workspace",
		Region:     "ap-southeast-1",
		BucketName: "cc-mahjong-0",
		Mode:       "dev",
		LocalRepo:  "/Users/chuancc/mywork/mylabs/csdc/mahjong-0/tiles-repo",
	}
}

// GenerateCdkApp return path where the base CDK App was generated.
func (d *Deployment) GenerateCdkApp(ctx context.Context, out *websocket.Conn) (*ExecutionPlan, error) {

	// 1. Loading Super from s3 & unzip
	// 2. Loading Tiles from s3 & unzip
	var aTs = &Ts{}
	var override = make(map[string]*TileInputOverride) //TileName->TileInputOverride
	var ep *ExecutionPlan
	SR(out, []byte("Loading Super ... from RePO."))
	_, err := s3Config.LoadSuper()
	if err != nil {
		return ep, err
	}
	SR(out, []byte("Loading Super ... from RePO with success."))

	for _, t := range d.Spec.Template.Tiles {
		if err := d.PullTile(ctx, t.TileReference, t.TileVersion, out, aTs, override); err != nil {
			return ep, err
		}
	}

	// 3. Generate super.ts
	if err := d.ApplyMainTs(ctx, out, aTs); err != nil {
		return ep, err
	}

	// 4. Caching Ts
	AllTs[ctx.Value("d-sid").(string)] = *aTs

	// 4. Generate execution plan
	return d.GenerateExecutePlan(ctx, out, aTs)

}

func (d *Deployment) PullTile(ctx context.Context, tile string, version string, out *websocket.Conn, aTs *Ts, override map[string]*TileInputOverride) error {

	var rStack = "Stack"//+ctx.Value("d-sid").(string)[0:8]

	// 1. Loading Tile from s3 & unzip
	tileSpecFile, err := s3Config.LoadTile(tile, version)
	if err != nil {
		SRf(out, "Failed to pulling Tile < %s - %s > ... from RePO\n", tile, version)
	} else {
		SRf(out, "Pulling Tile < %s - %s > ... from RePO with success\n", tile, version)
	}

	//parse tile-spec.yaml if need more tile
	SRf(out, "Parsing specification of Tile: < %s - %s > ...\n", tile, version)
	buf, err := ioutil.ReadFile(tileSpecFile)
	if err != nil {
		return err
	}
	data := Data(buf)
	parsedTile, err := data.ParseTile(ctx)
	if err != nil {
		return err
	}
	// TODO: to be refactor
	// Step 0. Caching the tile
	if aTs.AllTiles == nil {
		aTs.AllTiles = make(map[string]Tile)
	}
	aTs.AllTiles[parsedTile.Metadata.Category+"-"+parsedTile.Metadata.Name] = *parsedTile
	////

	// Step 1. Caching deployment inputs
	deploymentInputs := make(map[string][]string) //tileName -> map[inputName]inputValues
	for _, tts := range d.Spec.Template.Tiles {
		for _, n := range tts.Inputs {

			if len(n.InputValues) > 0 {
				deploymentInputs[tts.TileReference+"-"+n.Name] = n.InputValues
			} else {
				deploymentInputs[tts.TileReference+"-"+n.Name] = []string{n.InputValue}
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
				override[dependenciesMap[ov.Override.Name]+"-"+ov.Override.Field] = &TileInputOverride{
					Name:      ov.Override.Name,
					Field:     ov.Override.Field,
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
		TileName:     parsedTile.Metadata.Name,
		TileVersion:  parsedTile.Metadata.Version,
		TileConstructName: strings.ReplaceAll(parsedTile.Metadata.Name, "-",""),
		TileFolder:   strings.ToLower(parsedTile.Metadata.Name),
		TileCategory: parsedTile.Metadata.Category,
	}
	if aTs.TsLibsMap == nil {
		aTs.TsLibsMap = make(map[string]TsLib)
	}
	if _, ok := aTs.TsLibsMap[parsedTile.Metadata.Name]; !ok {
		aTs.TsLibsMap[parsedTile.Metadata.Name] = newTsLib
	}
	////

	// Step 5. Caching inputs <key, value> for further process
	// inputs: inputName -> inputValue
	inputs := make(map[string]TsInputParameter)
	for _, in := range parsedTile.Spec.Inputs {

		input := TsInputParameter{}
		if in.Dependencies != nil {
			// For value dependent on other Tile
			if parsedTile.Metadata.Category != ContainerApplication.CString() &&
				parsedTile.Metadata.Category != Application.CString() {
				if len(in.Dependencies) == 1 {
					// single dependency
					input.InputName = in.Name
					stile := strings.ReplaceAll(strings.ToLower(dependenciesMap[in.Dependencies[0].Name]),"-","")
					input.InputValue = stile + rStack + "var." + stile + "var." + in.Dependencies[0].Field
				} else {
					// multiple dependencies will be organized as an array
					input.InputName = in.Name
					v := "[ "
					for _, d := range in.Dependencies {
						stile := strings.ReplaceAll(strings.ToLower(dependenciesMap[d.Name]),"-","")
						val := stile + rStack + "var." + stile + "var." + d.Field
						v = v + val + ","
					}
					input.InputValue = strings.TrimSuffix(v, ",") + " ]"
				}
			} else {
				// output value can be retrieved after execution: $D-TBD_TileName.Output-Name
				// !!!Now support non-CDK tile can reference value from dependent Tile by injecting ENV!!!
				input.InputName = in.Name
				input.InputValue = strings.ToUpper("$D_TBD_"+
					strings.ReplaceAll(parsedTile.Metadata.Name,"-","_")+
					"_"+in.Dependencies[0].Field)
			}
		} else {
			// For independent value
			input.InputName = in.Name
			// Overwrite values as per Deployment
			if val, ok := deploymentInputs[parsedTile.Metadata.Name+"-"+in.Name]; ok {

				if len(val) > 1 {
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
		if in.Override.Name != "" { input.IsOverrideField = "yes" }
		inputs[input.InputName] = input
	}
	////

	// Step 6.Setup values for cached override, depend on Step 5
	for _, v := range override {
		if input, ok := inputs[v.InputName]; ok {
			v.OverrideValue = input.InputValue
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
	if ns == "" {
		ns = parsedTile.Spec.Manifests.Namespace
	}
	tm.Namespace = ns

	// Overwrite files/folders as deployment
	var ffs []string
	var fds []string
	for _, m := range d.Spec.Template.Tiles {
		if m.TileReference == parsedTile.Metadata.Name {
			if m.Manifests.Files != nil {
				ffs = m.Manifests.Files
			}
			if m.Manifests.Folders != nil {
				fds = m.Manifests.Folders
			}
		}
	}
	if ffs == nil {
		ffs = parsedTile.Spec.Manifests.Files
	}
	for _, m := range parsedTile.Spec.Manifests.Files {
		tm.Files = append(tm.Files, m)
	}
	if fds == nil {
		fds = parsedTile.Spec.Manifests.Folders
	}
	for _, m := range parsedTile.Spec.Manifests.Folders {
		tm.Folders = append(tm.Folders, m)
	}
	////

	// Step 8. Store import Stacks && avoid repeated one
	ts := &TsStack{
		TileName:          parsedTile.Metadata.Name,
		TileVersion:       parsedTile.Metadata.Version,
		TileConstructName: strings.ReplaceAll(parsedTile.Metadata.Name, "-",""),
		TileVariable:      strings.ReplaceAll(strings.ToLower(parsedTile.Metadata.Name),"-","") + "var",
		TileStackName:     strings.ReplaceAll(parsedTile.Metadata.Name,"-","") + rStack,
		TileStackVariable: strings.ReplaceAll(strings.ToLower(parsedTile.Metadata.Name),"-","") + rStack + "var",
		InputParameters:   inputs,
		TileCategory:      parsedTile.Metadata.Category,
		TsManifests:       tm,
	}
	if aTs.TsStacksMap == nil {
		aTs.TsStacksMap = make(map[string]*TsStack)
	}
	if _, ok := aTs.TsStacksMap[parsedTile.Metadata.Name]; !ok {
		aTs.TsStacksMap[parsedTile.Metadata.Name] = ts
		if aTs.TsStacksOrder == nil {
			aTs.TsStacksOrder = list.New()
		}
		aTs.TsStacksOrder.PushFront(parsedTile.Metadata.Name)
	}
	////

	// recurred call
	for _, t := range parsedTile.Spec.Dependencies {
		if err = d.PullTile(ctx, t.TileReference, t.TileVersion, out, aTs, override); err != nil {
			return err
		}
	}
	////
	// !!!Last job: checking vendor service before leaving, do it after recurring.
	// ???
	if parsedTile.Metadata.Category == ContainerApplication.CString() {
		for k, v := range aTs.AllTiles {
			if strings.Contains(k, ContainerProvider.CString()) {
				ts.TsManifests.VendorService = v.Metadata.VendorService
				ts.TsManifests.DependentTile = v.Metadata.Name
				ts.TsManifests.DependentTileVersion = v.Metadata.Version
			}
		}
	}
	////

	// !!!Caching Outputs
	if aTs.AllOutputs == nil {
		aTs.AllOutputs = make(map[string]*TsOutput)
	}
	to := &TsOutput{
		TileName:    tile,
		TileVersion: parsedTile.Metadata.Version,
		TsOutputs:   make(map[string]*TsOutputDetail),
	}
	for _, o := range parsedTile.Spec.Outputs {
		to.TsOutputs[o.Name] = &TsOutputDetail{
			Name:                o.Name,
			OutputType:          o.OutputType,
			DefaultValue:        o.DefaultValue,
			DefaultValueCommand: o.DefaultValueCommand,
			OutputValue:         o.DefaultValue,
			Description:         o.Description,
		}
	}
	aTs.AllOutputs[tile] = to
	////

	SRf(out, "Parsing specification of Tile: < %s - %s > was success.\n", tile, version)
	return nil
}

// ApplyMainTs apply values with super.ts template
func (d *Deployment) ApplyMainTs(ctx context.Context, out *websocket.Conn, aTs *Ts) error {
	superts := s3Config.WorkHome + "/super/bin/super.ts"
	SR(out, []byte("Generating main.ts for Super ..."))

	tp, _ := template.ParseFiles(superts)

	file, err := os.Create(superts + "_new")
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}

	//!!!reverse aTs.TsStacks due to CDK require!!!
	//for i := len(aTs.TsStacks)/2 - 1; i >= 0; i-- {
	//	opp := len(aTs.TsStacks) - 1 - i
	//	aTs.TsStacks[i], aTs.TsStacks[opp] = aTs.TsStacks[opp], aTs.TsStacks[i]
	//}
	if d.Spec.Template.ForceOrder != nil {
		// forced order
		for _,t := range d.Spec.Template.ForceOrder {
			aTs.TsLibs = append(aTs.TsLibs, aTs.TsLibsMap[t])
			aTs.TsStacks = append(aTs.TsStacks, aTs.TsStacksMap[t])
		}
	} else {
		// natural reversed order
		for e := aTs.TsStacksOrder.Front(); e != nil; e = e.Next() {
			n := e.Value.(string)
			aTs.TsLibs = append(aTs.TsLibs, aTs.TsLibsMap[n])
			aTs.TsStacks = append(aTs.TsStacks, aTs.TsStacksMap[n])
		}
	}

	err = tp.Execute(file, aTs)
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	err = file.Close()
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	os.Rename(superts, superts+"_old")
	os.Rename(superts+"_new", superts)
	buf, err := ioutil.ReadFile(superts)
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	SR(out, []byte("Generating main.ts for Super ... with success"))
	SR(out, []byte("--BO:-------------------------------------------------"))
	SR(out, buf)
	SR(out, []byte("--EO:-------------------------------------------------"))
	return nil
}

func (d *Deployment) GenerateExecutePlan(ctx context.Context, out *websocket.Conn, aTs *Ts) (*ExecutionPlan, error) {
	SR(out, []byte("Generating execution plan... "))
	var p = ExecutionPlan{
		Plan:             list.New(),
		PlanMirror:       make(map[string]*ExecutionStage),
		OriginDeployment: d,
	}
	for _, ts := range aTs.TsStacks {
		workHome := s3Config.WorkHome + "/super"
		stage := ExecutionStage{
			Name:          ts.TileName,
			Command:       list.New(),
			CommandMirror: make(map[string]string),
			Kind:          ts.TileCategory,
			WorkHome:      workHome,
			Preparation:   []string{"cd " + workHome},
			TileName:      ts.TileName,
			TileVersion:   ts.TileVersion,
		}
		// Define Kind of Stage
		if ts.TileCategory == ContainerApplication.CString() || ts.TileCategory == Application.CString() {
			stage.Kind = Command.SKString()
		} else {
			stage.Kind = CDK.SKString()
		}
		// Caching env
		if ts.EnvList == nil { ts.EnvList = make(map[string]string) }

		if ts.TileCategory == ContainerApplication.CString() || ts.TileCategory == Application.CString() {
			// Reserved stage.Preparation & inject reserved environment variables
			if ts.TsManifests.Namespace != "" && ts.TsManifests.Namespace != "default" {
				stage.Preparation = append(stage.Preparation, "kubectl create ns "+ts.TsManifests.Namespace+" || true")
				stage.Preparation = append(stage.Preparation, "export NAMESPACE="+ts.TsManifests.Namespace)
				ts.EnvList["NAMESPACE"]=ts.TsManifests.Namespace
			} else {
				ts.TsManifests.Namespace="default"
				ts.EnvList["NAMESPACE"]="default"
			}
			stage.Preparation = append(stage.Preparation, "export WORK_HOME="+s3Config.WorkHome+"/super")
			ts.EnvList["WORK_HOME"]=s3Config.WorkHome+"/super"
			stage.Preparation = append(stage.Preparation, "export TILE_HOME="+s3Config.WorkHome+"/super/lib/"+strings.ToLower(ts.TileName))
			ts.EnvList["TILE_HOME"]=s3Config.WorkHome+"/super/lib/"+strings.ToLower(ts.TileName)

			// Inject Global environment variables & Adding PreRun...Commands into stage.Preparation
			dSid := ctx.Value("d-sid").(string)
			if at, ok := AllTs[dSid]; ok {
				if tile, ok := at.AllTiles[ts.TileCategory+"-"+ts.TileName]; ok {
					for _, e := range tile.Spec.Global.Env {
						if e.Value != "" {
							stage.Preparation = append(stage.Preparation, fmt.Sprintf("export %s=%s", e.Name, e.Value))
							ts.EnvList[e.Name]=e.Value
						} else if e.Value == "" && e.ValueRef != "" {
							if v, ok := ts.InputParameters[e.ValueRef]; ok {
								stage.Preparation = append(stage.Preparation, fmt.Sprintf("export %s=%s", e.Name, v))
								ts.EnvList[e.Name]=v.InputValue
							}
						}
					}
					for _, s := range tile.Spec.PreRun.Stages {
						stage.Preparation = append(stage.Preparation, s.Command)
					}
				}
			}

			switch ts.TsManifests.ManifestType {
			case K8s.MTString():

				for j, f := range ts.TsManifests.Files {
					var cmd string
					cmd = "kubectl apply -f ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + " -n " + ts.TsManifests.Namespace
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
			//Post commands, output values to output.log
			fileName := s3Config.WorkHome + "/super/" + stage.Name + "-output.log"
			//Sleep 5 seconds to waiting pod's ready
			stage.Command.PushFront("sleep 10")
			stage.CommandMirror[strconv.Itoa(stage.Command.Len()+1)] = "sleep 10"
			if tile, ok := aTs.AllTiles[ts.TileCategory+"-"+ts.TileName]; ok {
				for _, o := range tile.Spec.Outputs {
					if o.DefaultValueCommand != "" {
						cmd := `echo "{\"` + o.Name + "=`" + o.DefaultValueCommand + "`" + `\"}" >>` + fileName
						stage.Command.PushFront(cmd)
						stage.CommandMirror[strconv.Itoa(stage.Command.Len()+1)] = cmd
					} else if o.DefaultValue != "" {
						cmd := `echo "{\"` + o.Name + "=" + o.DefaultValue + `\"}" >>` + fileName
						stage.Command.PushFront(cmd)
						stage.CommandMirror[strconv.Itoa(stage.Command.Len()+1)] = cmd
					}
				}
			}
		} else {
			stage.Preparation = append(stage.Preparation, "npm install")
			stage.Preparation = append(stage.Preparation, "npm run build")
			stage.Preparation = append(stage.Preparation, "cdk list")
			cmd := "cdk deploy " + ts.TileStackName + " --require-approval never"
			stage.Command.PushFront(cmd)
			stage.CommandMirror[strconv.Itoa(1)] = cmd
		}
		p.Plan.PushFront(&stage)
		p.PlanMirror[ts.TileName] = &stage
	}

	buf, _ := yaml.Marshal(p)
	err := SR(out, buf)
	SR(out, []byte("Generating execution plan... with success"))
	return &p, err
}
