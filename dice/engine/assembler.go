package engine

import (
	"container/list"
	"context"
	"dice/apis/v1alpha1"
	"dice/utils"
	"errors"
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

type AssembleData struct {
	Deployment *v1alpha1.Deployment
}

// AssemblerCore represents a group of functions to assemble CDK App.
type AssembleCore interface {
	// GenerateMainApp generate main app from base template with necessary tiles
	GenerateMainApp(ctx context.Context, out *websocket.Conn) (*ExecutionPlan, error)

	// validateDependsOn validates dependent tiles
	validateDependsOn(tileInstance []string) error

	// ProcessTiles controls Tile processing
	ProcessTiles(ctx context.Context, aTs *Ts, override map[string]*v1alpha1.TileInputOverride, out *websocket.Conn) error
	// PullTile pulls Tile from repo
	PullTile(ctx context.Context,
		tileInstance string,
		tileName string,
		version string,
		executableOrder int,
		parentTileInstance []string,
		rootTileInstances string,
		aTs *Ts,
		override map[string]*v1alpha1.TileInputOverride,
		region string,
		profile string,
		out *websocket.Conn) error

	// ApplyMainTs applies Ts to main CDK app
	ApplyMainTs(ctx context.Context, aTs *Ts, out *websocket.Conn) error

	//GenerateExecutePlan generates execution plan to direct provision resources
	GenerateExecutePlan(ctx context.Context, aTs *Ts, out *websocket.Conn) (*ExecutionPlan, error)

	//GenerateParallelPlan generates parallel plan
	GenerateParallelPlan(ctx context.Context, aTs *Ts, out *websocket.Conn) error
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

func UpdateDR(dr *DeploymentRecord, status string) {
	dr.Status = status
	dr.Updated = time.Now()
}

// GenerateMainApp return path where the base CDK App was generated.
func (d *AssembleData) GenerateMainApp(ctx context.Context, out *websocket.Conn) (*ExecutionPlan, error) {

	dSid := ctx.Value("d-sid").(string)
	aon := make(map[string]*TsOutput)
	var aTs = &Ts{
		DR: &DeploymentRecord{
			SID:         dSid,
			Name:        d.Deployment.Metadata.Name,
			Created:     time.Now(),
			SuperFolder: "/" + d.Deployment.Metadata.Name,
			Status:      Created.DSString(),
		},
		AllTilesN:    make(map[string]*v1alpha1.Tile),
		TsLibsMap:    make(map[string]TsLib),
		TsStacksMapN: make(map[string]*TsStack),
		AllOutputsN:  &aon,
	}

	// 1. Caching Ts
	// Cached here with point, so can be used in following procedures
	AllTs[dSid] = *aTs

	// 2. Loading Super from s3 & unzip
	UpdateDR(aTs.DR, Progress.DSString())
	var override = make(map[string]*v1alpha1.TileInputOverride) //TileName->TileInputOverride
	var ep *ExecutionPlan
	SR(out, []byte("Loading Super ... from RePO."))
	_, err := DiceConfig.LoadSuper(aTs.DR.SuperFolder)
	if err != nil {
		UpdateDR(aTs.DR, Interrupted.DSString())
		return ep, err
	}
	SR(out, []byte("Loading Super ... from RePO with success."))

	// 3. Loading Tiles from s3 & unzip
	if err := d.ProcessTiles(ctx, aTs, override, out); err != nil {
		UpdateDR(aTs.DR, Interrupted.DSString())
		return ep, err
	}
	if len(aTs.TsStacksMapN) < 1 {
		UpdateDR(aTs.DR, Interrupted.DSString())
		return ep, errors.New("invalid deployment without parsed Tiles")
	}

	// 4. Generate super.ts
	if err := d.ApplyMainTs(ctx, aTs, out); err != nil {
		UpdateDR(aTs.DR, Interrupted.DSString())
		return ep, err
	}

	// 5. Generate execution plan
	plan, err := d.GenerateExecutePlan(ctx, aTs, out)
	if err != nil {
		UpdateDR(aTs.DR, Interrupted.DSString())
		return nil, err
	}
	// 6. Store execution plan
	AllPlans[dSid] = plan

	//7. Generate parallel
	d.GenerateParallelPlan(ctx, aTs, out)

	return plan, nil
}

func (d *AssembleData) validateDependsOn(tileInstance []string) error {
	for _, ti := range tileInstance {
		if !utils.Contains(d.Deployment.OriginalOrder, ti) {
			return errors.New(ti + " wasn't existed in the deployment")
		}
	}
	return nil
}

// ProcessTiles controls Tiles processing
func (d *AssembleData) ProcessTiles(ctx context.Context, aTs *Ts, override map[string]*v1alpha1.TileInputOverride, out *websocket.Conn) error {
	// order factor, every tiles family +1000
	dSid := ctx.Value("d-sid").(string)
	executableOrder := 1000
	toBeProcessedTiles := make(map[string]v1alpha1.DeploymentTemplateDetail) //deploy-instance -> Tile

	for _, tileInstance := range d.Deployment.OriginalOrder {
		if deploy, ok := d.Deployment.Spec.Template.Tiles[tileInstance]; ok {
			parentTileInstances := []string{"root"}
			if deploy.DependsOn != nil {
				if err := d.validateDependsOn(deploy.DependsOn); err != nil {
					return err
				}
				parentTileInstances = deploy.DependsOn
				rootTileInstance := RootTileInstance(dSid, parentTileInstances[0])
				if IsProcessed(dSid, parentTileInstances) {
					if _, err := d.PullTile(ctx,
						tileInstance,
						deploy.TileReference,
						deploy.TileVersion,
						executableOrder,
						parentTileInstances,
						rootTileInstance, //set same family with dependent deploy if depends on deploy
						aTs,
						override,
						deploy.Region,
						deploy.Profile,
						out); err != nil {
						return err
					}

				} else {
					// caching and process later
					toBeProcessedTiles[tileInstance] = deploy
				}

			} else {
				if _, err := d.PullTile(ctx,
					tileInstance,
					deploy.TileReference,
					deploy.TileVersion,
					executableOrder,
					parentTileInstances,
					tileInstance,
					aTs,
					override,
					deploy.Region,
					deploy.Profile,
					out); err != nil {
					return err
				}
			}
		}
		executableOrder = executableOrder + 1000
	}
	// Process Tiles with dependencies
	for tileInstance, deploy := range toBeProcessedTiles {
		parentTileInstances := []string{"root"}
		if deploy.DependsOn != nil {
			parentTileInstances = deploy.DependsOn
		}
		rootTileInstance := RootTileInstance(dSid, parentTileInstances[0])
		if _, err := d.PullTile(ctx,
			tileInstance,
			deploy.TileReference,
			deploy.TileVersion,
			executableOrder,
			parentTileInstances,
			rootTileInstance,
			aTs,
			override,
			deploy.Region,
			deploy.Profile,
			out); err != nil {
			return err
		}
		executableOrder = executableOrder + 1000
	}
	return nil
}

// PullTile pulls Tile from Tile Repo and extract & setup data along the way
func (d *AssembleData) PullTile(ctx context.Context,
	tileInstance string,
	tileName string,
	version string,
	executableOrder int,
	parentTileInstances []string,
	rootTileInstance string,
	aTs *Ts,
	override map[string]*v1alpha1.TileInputOverride,
	region string,
	profile string,
	out *websocket.Conn) (string, error) {

	dSid := ctx.Value("d-sid").(string)
	ti := generateTileInstance(tileInstance, tileName, rootTileInstance)
	rStack := "Stack" + ti

	// Pre-Process 1: Loading Tile from s3 & unzip
	tileSpecFile, err := DiceConfig.LoadTile(tileName, version, aTs.DR.SuperFolder)
	if err != nil {
		SRf(out, "Failed to pulling Tile < %s - %s > ... from RePO\n", tileName, version)
		return ti, err
	} else {
		SRf(out, "Pulling Tile < %s - %s > ... from RePO with success\n", tileName, version)
	}

	// Pre-Process 2: Parse tileName-spec.yaml if need more tileName
	SRf(out, "Parsing specification of Tile: < %s - %s > ...\n", tileName, version)
	buf, err := ioutil.ReadFile(tileSpecFile)
	if err != nil {
		return ti, err
	}
	data := v1alpha1.Data(buf)
	parsedTile, err := data.ParseTile(ctx)
	if err != nil {
		return ti, err
	}

	// Pre-Process 3: Caching TilesGrid, which presents relation between Tiles
	tg := TilesGrid{
		TileInstance:        ti,
		ExecutableOrder:     executableOrder - 1,
		TileName:            tileName,
		TileVersion:         version,
		ParentTileInstances: parentTileInstances,
		RootTileInstance:    rootTileInstance,
		TileCategory:        parsedTile.Metadata.Category,
		Status:              Created.DSString(),
	}
	if parsedTile.Spec.Dependencies == nil && tileInstance == "" {
		tg.ParentTileInstances = []string{"root"}
	}
	if allTG, ok := AllTilesGrids[dSid]; ok && allTG != nil {
		if !IsDuplicatedTile(dSid, rootTileInstance, parsedTile.Metadata.Name) {
			(*allTG)[ti] = &tg
		} else {
			log.Debugf("It's duplicated Tile under same group, Ignore : %s / %s / %s\n", tileName, version, parsedTile.Metadata.Category)
			return ti, nil
		}
	} else {
		val := make(map[string]*TilesGrid)
		val[ti] = &tg
		AllTilesGrids[dSid] = &val

	}
	parsedTile.TileInstance = tg.TileInstance

	// Kick start processing
	// Step 1. Caching the tileName
	aTs.AllTilesN[ti] = parsedTile
	////

	// Step 2. Caching inputs & input values from deployment
	deploymentInputs := make(map[string][]string) //tileName-inputName -> map[inputName]inputValues
	if dt, ok := d.Deployment.Spec.Template.Tiles[tileInstance]; ok {
		for _, input := range dt.Inputs {

			if len(input.InputValues) > 0 {
				deploymentInputs[dt.TileReference+"-"+input.Name] = input.InputValues
			} else {
				deploymentInputs[dt.TileReference+"-"+input.Name] = []string{input.InputValue}
			}
		}
	}
	////

	// Step 3. Caching tileName dependencies for further process
	tileDependencies := make(map[string]string)
	for _, m := range parsedTile.Spec.Dependencies {
		tileDependencies[m.Name] = m.TileReference
	}

	// Step 4. Caching tileName override for further process; depends on Step 2.
	// !!!override - includes all input value from deployment & could be using in dependent Tiles.
	for _, ov := range parsedTile.Spec.Inputs {
		if ov.Override.Name != "" {
			if tileName, ok := tileDependencies[ov.Override.Name]; ok {
				if val, ok := deploymentInputs[tileName+"-"+ov.Override.Field]; ok {
					tlo := &v1alpha1.TileInputOverride{
						Name:  ov.Override.Name,
						Field: ov.Override.Field,
					}
					tlo.OverrideValue = array2String(val, ov.InputType)
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
	if ns := namespace(parsedTile.Metadata.Name, d.Deployment); ns == "" {
		tm.Namespace = parsedTile.Spec.Manifests.Namespace
	} else {
		tm.Namespace = ns
	}

	if parsedTile.Spec.Manifests.Files != nil {
		tm.Files = append(tm.Files, parsedTile.Spec.Manifests.Files...)
	}
	if parsedTile.Spec.Manifests.Folders != nil {
		tm.Folders = append(tm.Folders, parsedTile.Spec.Manifests.Folders...)
	}
	if parsedTile.Spec.Manifests.Flags != nil {
		tm.Flags = append(tm.Flags, parsedTile.Spec.Manifests.Flags...)
	}
	////

	// Step 7. recurred call for all dependent Tiles
	for _, t := range parsedTile.Spec.Dependencies {
		var tis []string
		if nti, err := d.PullTile(ctx,
			"",
			t.TileReference,
			t.TileVersion,
			tg.ExecutableOrder,
			[]string{tg.TileInstance},
			tg.RootTileInstance,
			aTs,
			override,
			region,
			profile,
			out); err != nil {
			return ti, err
		} else {
			tis = append(tis, nti)
		}
		if tis != nil {
			tg.ParentTileInstances = tis
		}
	}
	////

	// !!! Place here so that CDK can get dependent stack variables !!!
	// Step 8. Caching inputs <key, value> for further process
	// inputs: inputName -> inputValue
	inputs := make(map[string]*TsInputParameter)
	for _, tileInput := range parsedTile.Spec.Inputs {

		input := TsInputParameter{
			InputName: tileInput.Name,
			InputType: tileInput.InputType,
		}
		if tileInput.Dependencies != nil {
			// For value dependent on other Tile
			if parsedTile.Metadata.Category != v1alpha1.ContainerApplication.CString() &&
				parsedTile.Metadata.Category != v1alpha1.Application.CString() {
				if len(tileInput.Dependencies) == 1 {
					// single dependency
					//input.InputName = tileInput.Name
					refTileName := tileDependencies[tileInput.Dependencies[0].Name]
					tsStack := ReferencedTsStack(dSid, rootTileInstance, refTileName)
					if tsStack != nil {
						input.InputValue = tsStack.TileStackVariable + "." + tsStack.TileVariable + "." + tileInput.Dependencies[0].Field
					}
				} else {
					// multiple dependencies will be organized as an array
					//input.InputName = tileInput.Name
					v := ""
					for _, dependency := range tileInput.Dependencies {
						refTileName := tileDependencies[dependency.Name]
						tsStack := ReferencedTsStack(dSid, rootTileInstance, refTileName)
						val := tsStack.TileStackVariable + "." + tsStack.TileVariable + "." + dependency.Field
						v = v + val + ","
					}
					input.InputValue = strings.TrimSuffix(v, ",")
				}
			} else {
				// output value can be retrieved after execution: $D-TBD_TileName.Output-Name
				// !!!Now support non-CDK tileName can reference value from dependent Tile by injecting ENV!!!
				//input.InputName = tileInput.Name
				input.InputValue = strings.ToUpper("$D_TBD_" +
					strcase.ToScreamingSnake(parsedTile.Metadata.Name) +
					"_" +
					tileInput.Dependencies[0].Field)
			}
		} else {
			// For independent value
			//input.InputName = tileInput.Name
			// Overwrite values by values from Deployment
			if val, ok := deploymentInputs[parsedTile.Metadata.Name+"-"+tileInput.Name]; ok {
				input.InputValue = array2String(val, tileInput.InputType)

			} else {
				if tileInput.DefaultValues != nil {
					input.InputValue = array2String(tileInput.DefaultValues, tileInput.InputType)

				} else if len(tileInput.DefaultValue) > 0 {
					input.InputValue = array2String([]string{tileInput.DefaultValue}, tileInput.InputType)
				}
			}

			// Check out if include $cdk(tileInstance.tileName.field)
			if input.InputValue, err = CDKAllValueRef(dSid, input.InputValue); err != nil {
				return ti, nil
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
		inputs[input.InputName] = &input
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
		TileFolder:        "/lib/" + strings.ToLower(parsedTile.Metadata.Name),
		Region:            region,
		Profile:           profile,
	}
	if _, ok := aTs.TsStacksMapN[tg.TileInstance]; !ok {
		aTs.TsStacksMapN[tg.TileInstance] = ts
	}
	////

	// Step 10. !!!Caching Outputs
	tos := make(map[string]*TsOutputDetail)
	to := &TsOutput{
		TileName:     tileName,
		TileVersion:  parsedTile.Metadata.Version,
		TsOutputs:    &tos,
		OutputsOrder: parsedTile.Spec.OutputsOrder,
	}
	for _, o := range parsedTile.Spec.Outputs {
		(*to.TsOutputs)[o.Name] = &TsOutputDetail{
			Name:                o.Name,
			OutputType:          o.OutputType,
			DefaultValue:        o.DefaultValue,
			DefaultValueCommand: o.DefaultValueCommand,
			OutputValue:         o.DefaultValue,
			Description:         o.Description,
		}
	}
	(*aTs.AllOutputsN)[tg.TileInstance] = to
	////

	SRf(out, "Parsing specification of Tile: < %s - %s > was success.\n", tileName, version)
	return ti, nil
}

func namespace(tileName string, deployment *v1alpha1.Deployment) string {
	for _, m := range deployment.Spec.Template.Tiles {
		if m.TileReference == tileName {
			return m.Manifests.Namespace
		}
	}
	return ""
}

func array2String(array []string, inputType string) string {

	val := ""
	if strings.Contains(inputType, "[]") && len(array) > 1 {
		for _, d := range array {
			val = val + d + ","
		}
		val = strings.TrimSuffix(val, ",")
	} else {
		val = array[0]
	}
	return val

}

// ApplyMainTs apply values with super.ts template
func (d *AssembleData) ApplyMainTs(ctx context.Context, aTs *Ts, out *websocket.Conn) error {
	dSid := ctx.Value("d-sid").(string)
	superFile := DiceConfig.WorkHome + aTs.DR.SuperFolder + "/bin/super.ts"
	SR(out, []byte("Generating main.ts for Super ..."))

	tp, _ := template.ParseFiles(superFile)

	tmpFile, err := os.Create(superFile + "_new")
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
	for _, tsStack := range aTs.TsStacks {
		for _, ip := range tsStack.InputParameters {
			switch ip.InputType {
			case v1alpha1.String.IOTString(), v1alpha1.Secret.IOTString():
				ip.InputValueForTemplate = "'" + ip.InputValue + "'"
			case v1alpha1.String.IOTString() + "[]", v1alpha1.Secret.IOTString() + "[]":
				values := strings.Split(ip.InputValue, ",")
				str := "['"
				for _, v := range values {
					str = str + v + "','"
				}
				ip.InputValueForTemplate = strings.TrimSuffix(str, ",'") + " ]"
			default:
				if strings.Contains(ip.InputType, "[]") {
					ip.InputValueForTemplate = "[" + ip.InputValue + "]"
				} else {
					ip.InputValueForTemplate = ip.InputValue
				}

			}

		}
	}
	err = tp.Execute(tmpFile, aTs)
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	err = tmpFile.Close()
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	err = os.Rename(superFile, superFile+"_old")
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	err = os.Rename(superFile+"_new", superFile)
	if err != nil {
		SR(out, []byte(err.Error()))
		return err
	}
	SR(out, []byte("Generating main.ts for Super ... with success"))
	return nil
}

func (d *AssembleData) GenerateExecutePlan(ctx context.Context, aTs *Ts, out *websocket.Conn) (*ExecutionPlan, error) {
	//dSid := ctx.Value("d-sid").(string)
	SR(out, []byte("Generating execution plan... "))
	var p = ExecutionPlan{
		Name:             aTs.DR.Name,
		Plan:             list.New(),
		PlanMirror:       make(map[string]*ExecutionStage),
		OriginDeployment: d.Deployment,
	}

	for _, ts := range aTs.TsStacks {
		workHome := DiceConfig.WorkHome + aTs.DR.SuperFolder
		stage := ExecutionStage{
			Name:          ts.TileInstance,
			Kind:          ts.TileCategory,
			WorkHome:      workHome,
			Preparation:   []string{"cd $WORK_HOME"},
			TileName:      ts.TileName,
			TileVersion:   ts.TileVersion,
			ProbeCommands: make(map[string]v1alpha1.ReadinessProbe),
		}
		// Define Kind of Stage
		if ts.TileCategory == v1alpha1.ContainerApplication.CString() ||
			ts.TileCategory == v1alpha1.Application.CString() {
			stage.Kind = Command.SKString()
		} else {
			stage.Kind = CDK.SKString()
		}
		stage.InjectedEnv = append(stage.InjectedEnv, "export WORK_HOME="+DiceConfig.WorkHome+aTs.DR.SuperFolder)
		stage.InjectedEnv = append(stage.InjectedEnv, "export TILE_HOME="+DiceConfig.WorkHome+aTs.DR.SuperFolder+ts.TileFolder)
		if ts.Region != "" {
			stage.InjectedEnv = append(stage.InjectedEnv, "export AWS_DEFAULT_REGION="+ts.Region)
		}
		if ts.Profile != "" {
			stage.InjectedEnv = append(stage.InjectedEnv, "export AWS_DEFAULT_PROFILE="+ts.Profile)
		}

		if ts.TileCategory == v1alpha1.ContainerApplication.CString() ||
			ts.TileCategory == v1alpha1.Application.CString() {
			// ContainerApplication & Application
			// Inject namespace as environment or using default as namespace
			if ts.TsManifests.ManifestType != "" {
				if ts.TsManifests.Namespace != "" && ts.TsManifests.Namespace != "default" {
					stage.Preparation = append(stage.Preparation, "kubectl create ns "+ts.TsManifests.Namespace+" || true")
					stage.InjectedEnv = append(stage.InjectedEnv, "export NAMESPACE="+ts.TsManifests.Namespace)
				} else {
					ts.TsManifests.Namespace = "default"
					stage.InjectedEnv = append(stage.InjectedEnv, "export NAMESPACE=default")
				}

				// Process different manifests
				prefix := DiceConfig.WorkHome + aTs.DR.SuperFolder + ts.TileFolder + "/lib/"
				switch ts.TsManifests.ManifestType {
				case v1alpha1.K8s.MTString():

					for _, f := range ts.TsManifests.Files {
						var cmd string
						cmd = "kubectl apply -f" +
							" " + prefix + f + // applied file
							" -n " + ts.TsManifests.Namespace // namespace
						stage.Commands = append(stage.Commands, cmd)
					}
				case v1alpha1.Helm.MTString():
					flags := ""
					for _, flag := range ts.TsManifests.Flags {
						flags = flags + " " + flag
					}
					for _, f := range ts.TsManifests.Folders {
						cmd := "helm install  --namespace=" + ts.TsManifests.Namespace +
							flags + // flags
							" -g" + //generate the name // !!!or ts.TileInstance + // helm name!!!
							" " + prefix + f // applied folder
						stage.Commands = append(stage.Commands, cmd)
					}

				case v1alpha1.Kustomize.MTString():
					// TODO: not quite yet to support Kustomize
					for _, f := range ts.TsManifests.Folders {
						cmd := "kustomize build -f" +
							" " + prefix + f + " | kubectl -f - " + // applied folder
							" -n " + ts.TsManifests.Namespace // namespace
						stage.Commands = append(stage.Commands, cmd)
					}
				}
			}

			// Commands & output values to output.log
			fileName := DiceConfig.WorkHome + aTs.DR.SuperFolder + "/" + stage.Name + "-output.log"
			//Sleep 5 seconds to waiting pod's ready
			stage.Commands = append(stage.Commands, "sleep 10")
			if tile, ok := aTs.AllTilesN[ts.TileInstance]; ok {
				for _, o := range tile.Spec.Outputs {
					cmd := ""
					if o.DefaultValueCommand != "" {
						cmd = `echo "{\"` + o.Name + "=`" + o.DefaultValueCommand + "`" + `\"}" >>` + fileName
					} else if o.DefaultValue != "" {
						cmd = `echo "{\"` + o.Name + "=" + o.DefaultValue + `\"}" >>` + fileName
					}
					// Replace $( -> \$( to avoid "command not found" error in script
					if cmd != "" {
						if strings.Contains(cmd, `$(`) {
							cmd = strings.ReplaceAll(cmd, `$(`, `\$(`)
						}
						stage.Commands = append(stage.Commands, cmd)
					}
				}
			}
		} else {
			// CDK based Tiles
			stage.Preparation = append(stage.Preparation, "npm install")
			stage.Preparation = append(stage.Preparation, "npm run build")
			stage.Preparation = append(stage.Preparation, "cdk list")
			cmd := "cdk deploy " + ts.TileStackName + " --require-approval never --exclusively true"
			stage.Commands = append(stage.Commands, cmd)
			stage.Commands = append(stage.Commands, "sleep 10")
			// Force to get output in case CDK doesn't output all
			cmd = `aws cloudformation describe-stacks --stack-name ` +
				ts.TileStackName + ` --query "Stacks[0].Outputs[]" | ` +
				`jq -r  '.[] | [.OutputKey, .OutputValue] | reduce .[1:][] as $i ("` +
				ts.TileStackName + `.\(.[0])"; . + " = \($i)")'`
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
					stage.Preparation = append(stage.Preparation, "# "+s.Name)
					stage.Preparation = append(stage.Preparation, s.Command)
					if s.ReadinessProbe != nil {
						id := "dice-probe-" + uuid.New().String()
						stage.Preparation = append(stage.Preparation, id)
						stage.ProbeCommands[id] = *s.ReadinessProbe
					}

				}

				// Adding PostRun's commands into stage.PostRunCommands
				for _, s := range tile.Spec.PostRun.Stages {
					stage.PostRunCommands = append(stage.PostRunCommands, "# "+s.Name)
					stage.PostRunCommands = append(stage.PostRunCommands, s.Command)
					if s.ReadinessProbe != nil {
						id := "dice-probe-" + uuid.New().String()
						stage.PostRunCommands = append(stage.PostRunCommands, id)
						stage.ProbeCommands[id] = *s.ReadinessProbe
					}
				}
			}
		}

		p.Plan.PushFront(&stage)
		p.PlanMirror[ts.TileInstance] = &stage
	}

	//buf, err := yaml.Marshal(p)
	//SR(out, buf)
	SR(out, []byte("Generating execution plan... with success"))
	SR(out, []byte("\n+--------------Execution Flow--------------+\n"))
	SR(out, []byte(ToFlow(&p)))
	SR(out, []byte("\n+--------------Execution Flow--------------+\n"))
	return &p, nil
}

// GenerateParallelPlan generates parallel execution
func (d *AssembleData) GenerateParallelPlan(ctx context.Context, aTs *Ts, out *websocket.Conn) error {
	dSid := ctx.Value("d-sid").(string)
	if plan, ok := AllPlans[dSid]; ok {
		roots := AllRootTileInstance(dSid)
		for _, root := range roots {
			l := list.New()
			family := FamilyTileInstance(dSid, root)
			for e := plan.Plan.Back(); e != nil; e = e.Prev() {
				stage := e.Value.(*ExecutionStage)
				if utils.Contains(family, stage.Name) {
					l.PushFront(stage)
				}
			}
			plan.ParallelPlan = append(plan.ParallelPlan, l)
		}
	}
	return nil
}

func ToFlow(p *ExecutionPlan) string {

	flow := "{Start}"
	for e := p.Plan.Back(); e != nil; e = e.Prev() {
		stage := e.Value.(*ExecutionStage)
		flow = flow + " -> " + stage.Name
	}
	flow = flow + " -> {Stop}"
	return flow
}

func ToParallelFlow(p *ExecutionPlan) []string {

	var flows []string
	for _, parallel := range p.ParallelPlan {
		flow := "{Start}"
		for e := parallel.Back(); e != nil; e = e.Prev() {
			stage := e.Value.(*ExecutionStage)
			flow = flow + " -> " + stage.Name
		}
		flow = flow + " -> {Stop}"
		flows = append(flows, flow)
	}

	return flows
}

func generateTileInstance(tileInstance string, tileName string, rootTileInstance string) string {
	if tileInstance == "" {
		return fmt.Sprintf("%s%s%s", tileName, rootTileInstance, "Generated")
	} else {
		return tileInstance
	}

}
