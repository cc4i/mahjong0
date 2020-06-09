package engine

import (
	"container/list"
	"context"
	"dice/utils"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
	"text/template"
	"time"
)

// AssemblerCore represents a group of functions to assemble CDK App.
type AssemblerCore interface {
	// Generate CDK App from base template with necessary tiles
	GenerateCdkApp(ctx context.Context, out *websocket.Conn) (*ExecutionPlan, error)

	// Control Tile processing
	ProcessTiles(ctx context.Context, aTs *Ts, override map[string]*TileInputOverride, out *websocket.Conn)
	// Pull Tile from repo
	PullTile(ctx context.Context,
		tileInstance string,
		tile string,
		version string,
		executableOrder int,
		parentTileInstance string,
		rootTileInstance string,
		aTs *Ts,
		override map[string]*TileInputOverride,
		out *websocket.Conn) error

	//Generate Main Ts inside of CDK app
	ApplyMainTs(ctx context.Context, out *websocket.Conn, aTs *Ts) error

	//Generate execution plan to direct provision resources
	GenerateExecutePlan(ctx context.Context, out *websocket.Conn, aTs *Ts) (*ExecutionPlan, error)
}

var DiceConfig *utils.DiceConfig

func init() {
	// Must
	workHome, ok := os.LookupEnv("M_WORK_HOME")
	if !ok {
		log.Fatal("Failed to lookup M_WORK_HOME.")
	}
	// Must
	mode, ok := os.LookupEnv("M_MODE")
	if !ok {
		// Using prod mode and pulling tiles from S3 repo.
		log.Fatal("!!!Failed to lookup M_MODE and setup to 'prod' mode.!!!")
		mode = "prod"
	}
	// Must when 'prod' mode
	region, ok := os.LookupEnv("M_S3_BUCKET_REGION")
	if !ok && mode == "prod" {
		log.Fatal("Failed to lookup M_S3_BUCKET_REGION  on 'prod' mode.")
	}
	// Must when 'prod' mode
	bucketName, ok := os.LookupEnv("M_S3_BUCKET")
	if !ok && mode == "prod" {
		log.Fatal("Failed to lookup M_S3_BUCKET  on 'prod' mode.")
	}
	// Must when 'dev' mode
	localRepo, ok := os.LookupEnv("M_LOCAL_TILE_REPO")
	if !ok && mode == "dev" {
		log.Fatal("Failed to lookup M_LOCAL_TILE_REPO on 'dev' mode.")
	}

	DiceConfig = &utils.DiceConfig{
		WorkHome:   workHome,
		Region:     region,
		BucketName: bucketName,
		Mode:       mode,
		LocalRepo:  localRepo,
	}
	c, _ := yaml.Marshal(DiceConfig)
	log.Printf("Loaded configuration: \n%s\n", c)
}

// GenerateCdkApp return path where the base CDK App was generated.
func (d *Deployment) GenerateCdkApp(ctx context.Context, out *websocket.Conn) (*ExecutionPlan, error) {

	var aTs = &Ts{
		AllTilesN:    make(map[string]*Tile),
		CreatedTime:  time.Now(),
		TsLibsMap:    make(map[string]TsLib),
		TsStacksMapN: make(map[string]*TsStack),
		AllOutputsN:  make(map[string]*TsOutput),
	}
	// 1. Caching Ts
	// Cached here with point, so can be used in following procedures
	AllTs[ctx.Value("d-sid").(string)] = *aTs

	// 2. Loading Super from s3 & unzip
	var override = make(map[string]*TileInputOverride) //TileName->TileInputOverride
	var ep *ExecutionPlan
	SR(out, []byte("Loading Super ... from RePO."))
	_, err := DiceConfig.LoadSuper()
	if err != nil {
		return ep, err
	}
	SR(out, []byte("Loading Super ... from RePO with success."))

	// 3. Loading Tiles from s3 & unzip
	if err := d.ProcessTiles(ctx, aTs, override, out); err != nil {
		return ep, err
	}

	// 4. Generate super.ts
	if err := d.ApplyMainTs(ctx, out, aTs); err != nil {
		return ep, err
	}

	// 5. Generate execution plan
	return d.GenerateExecutePlan(ctx, out, aTs)

}

// ProcessTiles controls Tiles processing
func (d *Deployment) ProcessTiles(ctx context.Context, aTs *Ts, override map[string]*TileInputOverride, out *websocket.Conn) error {
	// order factor, every tiles family +1000
	executableOrder := 1000
	toBeProcessedTiles := make(map[string]DeploymentTemplateDetail) //tile-instance -> Tile
	var reversedTileInstance []string

	// Process by reversed order / revered order = right order we see in yaml
	for tileInstance, _ := range d.Spec.Template.Tiles {
		reversedTileInstance = append(reversedTileInstance, tileInstance)
	}
	for i := len(reversedTileInstance) - 1; i >= 0; i-- {
		tileInstance := reversedTileInstance[i]
		if tile, ok := d.Spec.Template.Tiles[tileInstance]; ok {
			parentTileInstance := "root"
			if tile.DependsOn != "" {
				parentTileInstance = tile.DependsOn
				if allTG, ok := AllTilesGrids[ctx.Value("d-sid").(string)]; ok && allTG != nil {
					if _, ok := (*allTG)[parentTileInstance]; ok {
						if err := d.PullTile(ctx,
							tileInstance,
							tile.TileReference,
							tile.TileVersion,
							executableOrder,
							parentTileInstance,
							parentTileInstance, //set same family with dependent tile if depends on tile
							aTs,
							override,
							out); err != nil {
							return err
						}

					} else {
						// caching and process later
						toBeProcessedTiles[tileInstance] = tile
					}
				}
			} else {
				if err := d.PullTile(ctx,
					tileInstance,
					tile.TileReference,
					tile.TileVersion,
					executableOrder,
					parentTileInstance,
					tileInstance,
					aTs,
					override,
					out); err != nil {
					return err
				}
			}
		}
		executableOrder = executableOrder + 1000
	}
	// Process Tiles with dependencies
	for tileInstance, tile := range toBeProcessedTiles {
		parentTileInstance := "root"
		if tile.DependsOn != "" {
			parentTileInstance = tile.DependsOn
		}
		if err := d.PullTile(ctx,
			tileInstance,
			tile.TileReference,
			tile.TileVersion,
			executableOrder,
			parentTileInstance,
			parentTileInstance,
			aTs,
			override,
			out); err != nil {
			return err
		}
		executableOrder = executableOrder + 1000
	}
	return nil
}

// PullTile pulls Tile from Tile Repo and extract & setup data along the way
func (d *Deployment) PullTile(ctx context.Context,
	tileInstance string,
	tile string,
	version string,
	executableOrder int,
	parentTileInstance string,
	rootTileInstance string,
	aTs *Ts,
	override map[string]*TileInputOverride,
	out *websocket.Conn) error {

	dSid := ctx.Value("d-sid").(string)
	id := generateAId()
	rStack := "Stack" + id

	// Pre-Process 1: Loading Tile from s3 & unzip
	tileSpecFile, err := DiceConfig.LoadTile(tile, version)
	if err != nil {
		SRf(out, "Failed to pulling Tile < %s - %s > ... from RePO\n", tile, version)
		return err
	} else {
		SRf(out, "Pulling Tile < %s - %s > ... from RePO with success\n", tile, version)
	}

	// Pre-Process 2: Parse tile-spec.yaml if need more tile
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

	// Pre-Process 3: Caching TilesGrid, which presents relation between Tiles
	ti := tileInstance
	if ti == "" {
		ti = fmt.Sprintf("%s-%s-%s", tile, id, "generated")
	}
	tg := TilesGrid{
		TileInstance:       ti,
		ExecutableOrder:    executableOrder - 1,
		TileName:           tile,
		TileVersion:        version,
		ParentTileInstance: parentTileInstance,
		RootTileInstance:   rootTileInstance,
		TileCategory:       parsedTile.Metadata.Category,
	}
	if allTG, ok := AllTilesGrids[dSid]; ok && allTG != nil {
		if !IsDuplicatedCategory(dSid, rootTileInstance, parsedTile.Metadata.Category) {
			(*allTG)[ti] = tg
		} else {
			log.Debugf("It's duplicated Tile under same group, Ignore : %s / %s / %s\n", tile, version, parsedTile.Metadata.Category)
			return nil
		}
	} else {
		val := make(map[string]TilesGrid)
		val[ti] = tg
		AllTilesGrids[dSid] = &val

	}
	parsedTile.TileInstance = tg.TileInstance

	// Kick start processing
	// Step 1. Caching the tile
	aTs.AllTilesN[ti] = parsedTile
	////

	// Step 2. Caching inputs & input values from deployment
	deploymentInputs := make(map[string][]string) //tileName-inputName -> map[inputName]inputValues
	if dt, ok := d.Spec.Template.Tiles[tileInstance]; ok {
		for _, input := range dt.Inputs {

			if len(input.InputValues) > 0 {
				deploymentInputs[dt.TileReference+"-"+input.Name] = input.InputValues
			} else {
				deploymentInputs[dt.TileReference+"-"+input.Name] = []string{input.InputValue}
			}
		}
	}
	////

	// Step 3. Caching tile dependencies for further process
	tileDependencies := make(map[string]string)
	for _, m := range parsedTile.Spec.Dependencies {
		tileDependencies[m.Name] = m.TileReference
	}

	// Step 4. Caching tile override for further process; depends on Step 2.
	// !!!override - includes all input value from deployment & could be using in dependent Tiles.
	for _, ov := range parsedTile.Spec.Inputs {
		if ov.Override.Name != "" {
			if tileName, ok := tileDependencies[ov.Override.Name]; ok {
				if val, ok := deploymentInputs[tile+"-"+ov.Override.Field]; ok {
					tlo := &TileInputOverride{
						Name:  ov.Override.Name,
						Field: ov.Override.Field,
						//OverrideValue: deploymentInputs
						//InputName: ov.Name,
					}
					if len(val) > 1 {
						tlo.OverrideValue = array2String(val, ov.InputType)
					} else {
						tlo.OverrideValue = str2string(val[0], ov.InputType)
					}
					override[tileName+"-"+ov.Override.Field] = tlo
				}
			}
		}
	}
	////

	// Step 5. Store import libs && avoid to add repeated one
	newTsLib := TsLib{
		TileInstance:      ti,
		TileName:          parsedTile.Metadata.Name,
		TileVersion:       parsedTile.Metadata.Version,
		TileConstructName: strcase.ToCamel(parsedTile.Metadata.Name),
		TileFolder:        strings.ToLower(parsedTile.Metadata.Name),
		TileCategory:      parsedTile.Metadata.Category,
	}
	if _, ok := aTs.TsLibsMap[parsedTile.Metadata.Name]; !ok {
		aTs.TsLibsMap[parsedTile.Metadata.Name] = newTsLib
	}
	////

	// Step 6. Caching manifest & overwrite
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

	// Step 7. recurred call for all dependent Tiles
	for _, t := range parsedTile.Spec.Dependencies {
		if err = d.PullTile(ctx,
			"",
			t.TileReference,
			t.TileVersion,
			tg.ExecutableOrder,
			tg.TileInstance,
			tg.RootTileInstance,
			aTs, override, out); err != nil {
			return err
		}
	}
	////

	// !!! Place here so that CDK can get dependent stack variables !!!
	// Step 8. Caching inputs <key, value> for further process
	// inputs: inputName -> inputValue
	inputs := make(map[string]TsInputParameter)
	for _, tileInput := range parsedTile.Spec.Inputs {

		input := TsInputParameter{}
		if tileInput.Dependencies != nil {
			// For value dependent on other Tile
			if parsedTile.Metadata.Category != ContainerApplication.CString() &&
				parsedTile.Metadata.Category != Application.CString() {
				if len(tileInput.Dependencies) == 1 {
					// single dependency
					input.InputName = tileInput.Name
					refTileName := tileDependencies[tileInput.Dependencies[0].Name]
					tsStack := ReferencedTsStack(dSid, rootTileInstance, refTileName)
					if tsStack != nil {
						input.InputValue = tsStack.TileStackVariable + "." + tsStack.TileVariable + "." + tileInput.Dependencies[0].Field
					}
				} else {
					// multiple dependencies will be organized as an array
					input.InputName = tileInput.Name
					v := "[ "
					for _, d := range tileInput.Dependencies {
						refTileName := tileDependencies[d.Name]
						tsStack := ReferencedTsStack(dSid, rootTileInstance, refTileName)
						val := tsStack.TileStackVariable + "." + tsStack.TileVariable + "." + d.Field
						v = v + val + ","
					}
					input.InputValue = strings.TrimSuffix(v, ",") + " ]"
				}
			} else {
				// output value can be retrieved after execution: $D-TBD_TileName.Output-Name
				// !!!Now support non-CDK tile can reference value from dependent Tile by injecting ENV!!!
				input.InputName = tileInput.Name
				input.InputValue = strings.ToUpper("$D_TBD_" +
					strcase.ToScreamingSnake(parsedTile.Metadata.Name) +
					"_" +
					tileInput.Dependencies[0].Field)
			}
		} else {
			// For independent value
			input.InputName = tileInput.Name
			// Overwrite values by values from Deployment
			if val, ok := deploymentInputs[parsedTile.Metadata.Name+"-"+tileInput.Name]; ok {
				if len(val) > 1 {
					input.InputValue = array2String(val, tileInput.InputType)
				} else {
					input.InputValue = str2string(val[0], tileInput.InputType)
				}

			} else {
				if tileInput.DefaultValues != nil {
					input.InputValue = array2String(tileInput.DefaultValues, tileInput.InputType)

				} else if len(tileInput.DefaultValue) > 0 {
					input.InputValue = str2string(tileInput.DefaultValue, tileInput.InputType)
				}
			}

			// Check out if include $cdk(tileInstance.tileName.field)
			if input.InputValue, err = CDKAllValueRef(dSid, input.InputValue); err != nil {
				return nil
			}

		}
		//lookup override
		if or, ok := override[parsedTile.Metadata.Name+"-"+input.InputName]; ok {
			if input.InputName == or.Field {
				input.InputValue = or.OverrideValue
			}
		}
		if tileInput.Override.Name != "" {
			input.IsOverrideField = "yes"
		}
		inputs[input.InputName] = input
	}
	////

	// Step 9. Store import Stacks && avoid repeated one
	ts := &TsStack{
		TileInstance:      ti,
		TileName:          parsedTile.Metadata.Name,
		TileVersion:       parsedTile.Metadata.Version,
		TileConstructName: strcase.ToCamel(parsedTile.Metadata.Name),
		TileVariable:      strings.ToLower(strcase.ToCamel(parsedTile.Metadata.Name)) + "Var",
		TileStackName:     strcase.ToCamel(parsedTile.Metadata.Name) + rStack,
		TileStackVariable: strings.ToLower(strcase.ToCamel(parsedTile.Metadata.Name)) + rStack + "Var",
		InputParameters:   inputs,
		TileCategory:      parsedTile.Metadata.Category,
		TsManifests:       tm,
	}
	if _, ok := aTs.TsStacksMapN[tg.TileInstance]; !ok {
		aTs.TsStacksMapN[tg.TileInstance] = ts
	}
	////

	// Step 10. !!!Caching Outputs
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
	aTs.AllOutputsN[tg.TileInstance] = to
	////

	SRf(out, "Parsing specification of Tile: < %s - %s > was success.\n", tile, version)
	return nil
}

func array2String(array []string, inputType string) string {
	val := "[ "
	switch inputType {
	case String.IOTString() + "[]":
		for _, d := range array {
			if strings.HasPrefix(d, "$(") || strings.HasPrefix(d, "$cdk(") {
				val = val + d + ","
			} else {
				val = val + "'" + d + "',"
			}
		}
	default:
		for _, d := range array {
			val = val + d + ","
		}

	}
	val = strings.TrimSuffix(val, ",") + " ]"
	return val
}

func str2string(str string, inputType string) string {
	val := ""
	switch inputType {
	case String.IOTString():
		if strings.HasPrefix(str, "$(") || strings.HasPrefix(str, "$cdk(") {
			val = str
		} else {
			val = "'" + str + "'"
		}

	default:
		val = str

	}
	return val
}

// ApplyMainTs apply values with super.ts template
func (d *Deployment) ApplyMainTs(ctx context.Context, out *websocket.Conn, aTs *Ts) error {
	dSid := ctx.Value("d-sid").(string)
	sts := DiceConfig.WorkHome + "/super/bin/super.ts"
	SR(out, []byte("Generating main.ts for Super ..."))

	tp, _ := template.ParseFiles(sts)

	file, err := os.Create(sts + "_new")
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}

	//!!! re-order aTs.TsStacks due to CDK require!!!
	stg := SortedTilesGrid(dSid)
	for _, s := range stg {
		aTs.TsStacks = append(aTs.TsStacks, aTs.TsStacksMapN[s.TileInstance])
	}
	for _, tl := range aTs.TsLibsMap {
		aTs.TsLibs = append(aTs.TsLibs, tl)
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
	os.Rename(sts, sts+"_old")
	os.Rename(sts+"_new", sts)
	buf, err := ioutil.ReadFile(sts)
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
		workHome := DiceConfig.WorkHome + "/super"
		stage := ExecutionStage{
			Name:        ts.TileInstance,
			Kind:        ts.TileCategory,
			WorkHome:    workHome,
			Preparation: []string{"cd " + workHome},
			TileName:    ts.TileName,
			TileVersion: ts.TileVersion,
		}
		// Define Kind of Stage
		if ts.TileCategory == ContainerApplication.CString() || ts.TileCategory == Application.CString() {
			stage.Kind = Command.SKString()
		} else {
			stage.Kind = CDK.SKString()
		}
		// Caching & injected work_home & tile_home
		//if ts.PredefinedEnv == nil {
		//	ts.PredefinedEnv = make(map[string]string)
		//}
		stage.InjectedEnv = append(stage.InjectedEnv, "export WORK_HOME="+DiceConfig.WorkHome+"/super")
		//ts.PredefinedEnv["WORK_HOME"] = DiceConfig.WorkHome + "/super"
		stage.InjectedEnv = append(stage.InjectedEnv, "export TILE_HOME="+DiceConfig.WorkHome+"/super/lib/"+strings.ToLower(ts.TileName))
		//ts.PredefinedEnv["TILE_HOME"] = DiceConfig.WorkHome + "/super/lib/" + strings.ToLower(ts.TileName)

		if ts.TileCategory == ContainerApplication.CString() || ts.TileCategory == Application.CString() {
			// ContainerApplication & Application
			// Inject namespace as environment or using default as namespace
			if ts.TsManifests.ManifestType != "" {
				if ts.TsManifests.Namespace != "" && ts.TsManifests.Namespace != "default" {
					stage.Preparation = append(stage.Preparation, "kubectl create ns "+ts.TsManifests.Namespace+" || true")
					stage.InjectedEnv = append(stage.InjectedEnv, "export NAMESPACE="+ts.TsManifests.Namespace)
					//ts.PredefinedEnv["NAMESPACE"] = ts.TsManifests.Namespace
				} else {
					ts.TsManifests.Namespace = "default"
					stage.InjectedEnv = append(stage.InjectedEnv, "export NAMESPACE=default")
					//ts.PredefinedEnv["NAMESPACE"] = "default"
				}

				// Process different manifests
				switch ts.TsManifests.ManifestType {
				case K8s.MTString():

					for _, f := range ts.TsManifests.Files {
						var cmd string
						cmd = "kubectl apply -f ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + " -n " + ts.TsManifests.Namespace
						stage.Commands = append(stage.Commands, cmd)
					}
				case Helm.MTString():
					// TODO: not quite yet to support Helm
					for _, f := range ts.TsManifests.Folders {
						cmd := "helm install " + ts.TileName + " ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + " -n " + ts.TsManifests.Namespace
						stage.Commands = append(stage.Commands, cmd)
					}

				case Kustomize.MTString():
					// TODO: not quite yet to support Kustomize
					for _, f := range ts.TsManifests.Folders {
						cmd := "kustomize build -f ./lib/" + strings.ToLower(ts.TileName) + "/lib/" + f + "|kubectl -f - " + " -n " + ts.TsManifests.Namespace
						stage.Commands = append(stage.Commands, cmd)
					}
				}
			}

			// Commands & output values to output.log
			fileName := DiceConfig.WorkHome + "/super/" + stage.Name + "-output.log"
			//Sleep 5 seconds to waiting pod's ready
			stage.Commands = append(stage.Commands, "sleep 10")
			if tile, ok := aTs.AllTilesN[ts.TileInstance]; ok {
				for _, o := range tile.Spec.Outputs {
					if o.DefaultValueCommand != "" {
						cmd := `echo "{\"` + o.Name + "=`" + o.DefaultValueCommand + "`" + `\"}" >>` + fileName
						stage.Commands = append(stage.Commands, cmd)
					} else if o.DefaultValue != "" {
						cmd := `echo "{\"` + o.Name + "=" + o.DefaultValue + `\"}" >>` + fileName
						stage.Commands = append(stage.Commands, cmd)
					}
				}
			}
		} else {
			// CDK based Tiles
			stage.Preparation = append(stage.Preparation, "npm install")
			stage.Preparation = append(stage.Preparation, "npm run build")
			stage.Preparation = append(stage.Preparation, "cdk list")
			cmd := "cdk deploy " + ts.TileStackName + " --require-approval never"
			stage.Commands = append(stage.Commands, cmd)
		}

		dSid := ctx.Value("d-sid").(string)
		if at, ok := AllTs[dSid]; ok {
			if tile, ok := at.AllTilesN[ts.TileInstance]; ok {
				// Inject Global environment variables
				for _, e := range tile.Spec.Global.Env {
					if e.Value != "" {
						if v, err := ValueRef(dSid, e.Value, ts.TileInstance); err != nil {
							log.Errorf("Inject Global environment : %s  was failed: %s\n", e.Value, err.Error())
						} else {
							stage.InjectedEnv = append(stage.InjectedEnv, fmt.Sprintf("export %s=%s", e.Name, v))
						}

					}
				}
				// Adding PreRun's commands into stage.Preparation
				for _, s := range tile.Spec.PreRun.Stages {
					stage.Preparation = append(stage.Preparation, s.Command)
				}

				// Adding PostRun's commands into stage.PostRunCommands
				for _, s := range tile.Spec.PostRun.Stages {
					stage.PostRunCommands = append(stage.PostRunCommands, s.Command)
				}
			}
		}

		p.Plan.PushFront(&stage)
		p.PlanMirror[ts.TileName] = &stage
	}

	buf, err := yaml.Marshal(p)
	SR(out, buf)
	SR(out, []byte("Generating execution plan... with success"))
	return &p, err
}

func generateAId() string {
	return uuid.New().String()[0:8]
}
