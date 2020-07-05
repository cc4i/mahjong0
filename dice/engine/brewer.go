package engine

import (
	"bufio"
	"bytes"
	"container/list"
	"context"
	"dice/apis/v1alpha1"
	"dice/utils"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

// ExecutionPlan represents complete plan.
type ExecutionPlan struct {
	Name             string                     `json:"name"`         // The name of execution plan
	CurrentStage     *ExecutionStage            `json:"currentStage"` // Current executing stage
	Plan             *list.List                 `json:"plan"`         // Ordered plans
	PlanMirror       map[string]*ExecutionStage `json:"planMirror"`
	ParallelPlan     []*list.List               `json:"parallelPlan"`     // Parallel plans
	OriginDeployment *v1alpha1.Deployment       `json:"originDeployment"` // Original deployment data
}

// ExecutionStage represents an unit of execution plan.
type ExecutionStage struct {
	// Name
	Name string `json:"name"` // = TileInstance
	// Stage type
	Kind            string                             `json:"kind"`            // CDK/Command
	WorkHome        string                             `json:"workHome"`        // root folder for execution
	InjectedEnv     []string                           `json:"injectedEnv"`     // example: "export variable=value"
	Preparation     []string                           `json:"preparation"`     // Includes pre-run commands
	Commands        []string                           `json:"commands"`        // Command
	TileName        string                             `json:"tileName"`        // The name of Tile
	TileVersion     string                             `json:"tileVersion"`     // Tile version
	PostRunCommands []string                           `json:"postRunCommands"` // Includes post-run commands
	ProbeCommands   map[string]v1alpha1.ReadinessProbe `json:"probeCommands"`   // Commands for readiness probe
}

// StageKind defines type of stage
type StageKind int

const (
	// CDK based Tile
	CDK StageKind = iota
	// Non-CDK based Tile
	Command
)

// Convert enumeration into string
func (sk StageKind) SKString() string {
	return [...]string{"CDK", "Command", "FromCommand"}[sk]
}

// BrewerCore represent a group of core functions to execute & manage for
// execution plan.
type BrewerCore interface {
	// ExecutePlan executes the generated plan
	ExecutePlan(ctx context.Context, dryRun bool, out *websocket.Conn) error
	ExecuteParallelPlan(ctx context.Context, dryRun bool, out *websocket.Conn) error
	RunPlan(ctx context.Context, dryRun bool, wg *sync.WaitGroup, out *websocket.Conn) error

	// CommandExecutor executes the generated script or wire simulated data
	CommandExecutor(ctx context.Context, dryRun bool, cmdTxt []byte, out *websocket.Conn) error
	// LinuxCommandExecutor run a command/script
	LinuxCommandExecutor(ctx context.Context, cmdTxt []byte, stageLog *log.Logger, out *websocket.Conn) error
	// CommandWrapperExecutor wrap all parameters and commands into a script
	CommandWrapperExecutor(ctx context.Context, dryRun bool, out *websocket.Conn) (string, error)
	// WsTail collects all output data
	WsTail(ctx context.Context, reader io.ReadCloser, stageLog *log.Logger, wg *sync.WaitGroup, out *websocket.Conn)
	// ExtractValue extract output data from logs
	ExtractValue(ctx context.Context, buf []byte, out *websocket.Conn) error
	// ScanOutput scans output values from logs
	ScanOutput(regx *regexp.Regexp, buf []byte, outputDetail *TsOutputDetail, out *websocket.Conn) error
	// PostRun execute post jobs after major work
	PostRun(ctx context.Context, dryRun bool, out *websocket.Conn) (string, error)
	// GenerateSummary generate report for Deployment
	GenerateSummary(ctx context.Context, out *websocket.Conn) error
	// ExtractAllEnv extracts all possible key,value from environment variables
	ExtractAllEnv() map[string]string
	// ReplaceAll replaces value reference and environment variable to actual value
	ReplaceAll(str string, dSid string, kv map[string]string) string
	// ReplaceAllEnv replaces environment variable to actual value
	ReplaceAllEnv(str string, allEnv map[string]string) string
	// ReplaceAllValueRef replaces value reference to actual value
	ReplaceAllValueRef(str string, dSid string, ti string) string
}

// ExecutePlan is a orchestrator to run execution plan.
// Execution plan would only parse and use test data provided by Tile, but no commands would be sent
// if dryRun is true
func (ep *ExecutionPlan) ExecutePlan(ctx context.Context, dryRun bool, out *websocket.Conn) error {
	// 1. Run Plan
	if err := ep.RunPlan(ctx, dryRun, nil, out); err != nil {
		return err
	}
	//

	// 2. GENERATE REPORT
	if ep.Plan.Len() > 0 { //avoid empty plan
		if err := ep.GenerateSummary(ctx, out); err != nil {
			return err
		}
	}
	//
	return nil
}

func (ep *ExecutionPlan) RunPlan(ctx context.Context, dryRun bool, wg *sync.WaitGroup, out *websocket.Conn) error {
	dSid := ctx.Value("d-sid").(string)
	if aTs, ok := AllTs[dSid]; ok {
		for e := ep.Plan.Back(); e != nil; e = e.Prev() {
			stage := e.Value.(*ExecutionStage)
			for {
				if dependency, ok, err := IsDependenciesDone(dSid, stage.Name); ok {
					break
				} else {
					if err != nil {
						return err
					} else {
						log.Infof("Dependent tile : %s was not ready yet and waiting 5s ...", dependency)
						time.Sleep(5 * time.Second)
					}
				}
			}

			ep.CurrentStage = stage
			setStatus(dSid, stage.Name, Progress.DSString())

			// 1. Wrap commands into a shell script
			cmd, err := ep.CommandWrapperExecutor(ctx, dryRun, out)
			if err != nil {
				setStatus(dSid, stage.Name, Interrupted.DSString())
				return err
			}
			//

			// 2. Execute wrapped script
			if err := ep.CommandExecutor(ctx, dryRun, []byte(cmd), out); err != nil {
				setStatus(dSid, stage.Name, Interrupted.DSString())
				return err
			}
			//

			// 3.Extract output values & caching results
			buf, err := ioutil.ReadFile(DiceConfig.WorkHome + aTs.DR.SuperFolder + "/" + stage.Name + "-output.log")
			if err != nil {
				setStatus(dSid, stage.Name, Interrupted.DSString())
				return err
			}
			err = ep.ExtractValue(ctx, buf, out)
			if err != nil {
				setStatus(dSid, stage.Name, Interrupted.DSString())
				return err
			}
			//

			// 4. Post run with commands
			if ep.CurrentStage.PostRunCommands != nil {
				cmd, err := ep.PostRun(ctx, dryRun, out)
				if err != nil {
					setStatus(dSid, stage.Name, Interrupted.DSString())
					return err
				}
				err = ep.CommandExecutor(ctx, dryRun, []byte(cmd), out)
				if err != nil {
					setStatus(dSid, stage.Name, Interrupted.DSString())
					return err
				}
			}
			//
			setStatus(dSid, stage.Name, Done.DSString())

		}
	}
	if wg!=nil {wg.Done()}
	return nil
}

func (ep *ExecutionPlan) ExecuteParallelPlan(ctx context.Context, dryRun bool, out *websocket.Conn) error {
	// 1. Run plan by parallel
	wg := new(sync.WaitGroup)
	for _, pp := range ep.ParallelPlan {
		sep := &ExecutionPlan{
			Name: "parallel-plan",
			Plan: pp,
		}

		go sep.RunPlan(ctx, dryRun, wg, out)
		wg.Add(1)
	}
	wg.Wait()

	// 2. GENERATE REPORT
	if ep.Plan.Len() > 0 { //avoid empty plan
		if err := ep.GenerateSummary(ctx, out); err != nil {
			return err
		}
	}
	//

	return nil
}

// GenerateSummary generate summary after running execution plan.
func (ep *ExecutionPlan) GenerateSummary(ctx context.Context, out *websocket.Conn) error {
	file, err := os.OpenFile(DiceConfig.WorkHome+"/"+ep.Name+"output-summary.txt",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		SRf(out, "Failed to write summary, %s\n", err)
		return err
	}
	defer file.Close()

	SRF(out, file, []byte("\n"))
	SRF(out, file, []byte("\n\n============================Summary====================================\n\n"))
	dSid := ctx.Value("d-sid").(string)
	kv := ep.ExtractAllEnv()
	if ep.OriginDeployment.Spec.Summary.Description != "" {

		SRF(out, file, []byte(ep.ReplaceAll(ep.OriginDeployment.Spec.Summary.Description, dSid, kv)+"\n"))
	}
	SRF(out, file, []byte("\n"))
	for _, ot := range ep.OriginDeployment.Spec.Summary.Outputs {
		SRF(out, file, []byte(fmt.Sprintf("%s = %s\n", ot.Name, ep.ReplaceAll(ot.Value, dSid, kv))))
	}
	SRF(out, file, []byte("\n"))
	for _, n := range ep.OriginDeployment.Spec.Summary.Notes {
		SRF(out, file, []byte(ep.ReplaceAll(n, dSid, kv)+"\n"))
	}
	SRF(out, file, []byte("\n\n=======================================================================\n"))
	return nil
}

func (ep *ExecutionPlan) ExtractAllEnv() map[string]string {
	env := make(map[string]string)
	for e := ep.Plan.Back(); e != nil; e = e.Prev() {
		stage := e.Value.(*ExecutionStage)
		//replace by env value
		for _, val := range stage.InjectedEnv {
			re := regexp.MustCompile(`^export (.*)=(.*)$`)
			kv := re.FindStringSubmatch(val)
			if len(kv) == 3 {
				env[kv[1]] = kv[2]
			}
		}
	}
	return env
}

// Replace all possible env & value reference
func (ep *ExecutionPlan) ReplaceAll(str string, dSid string, kv map[string]string) string {
	str = ep.ReplaceAllEnv(str, kv)
	str = ep.ReplaceAllValueRef(str, dSid, "")                   //replace 'anything else'
	return str
}

// Replace reference by value
func (ep *ExecutionPlan) ReplaceAllValueRef(str string, dSid string, ti string) string {
	max := strings.Count(str, "$")
	for {
		re := regexp.MustCompile(`.*(\$\([[:alnum:]]*\.[[:alnum:]]*\.[[:alnum:]]*\)).*`)
		s := re.FindStringSubmatch(str)
		//
		if len(s) == 2 {
			if v, err := ValueRef(dSid, s[1], ti); err != nil {
				log.Errorf("Replace value reference was failed : %s \n", err)
				break
			} else {
				if v != "" {
					str = strings.ReplaceAll(str, s[1], v)
				}
			}
		} else {
			break
		}
		// avoid infinite loop due to replacement failure
		max--
		if max < 0 {
			break
		}
	}
	return str
}

// Replace env by value
func (ep *ExecutionPlan) ReplaceAllEnv(str string, allEnv map[string]string) string {
	for k, v := range allEnv {
		str = strings.ReplaceAll(str, "$"+k, v)
	}
	return str
}

// CommandExecutor exec command and return output.
func (ep *ExecutionPlan) CommandExecutor(ctx context.Context, dryRun bool, cmdTxt []byte, out *websocket.Conn) error {
	dSid := ctx.Value("d-sid").(string)
	aTs := AllTs[dSid]
	var stageLog *log.Logger
	SR(out, []byte("Initializing stage log file ..."))
	stageLog = log.New()
	fileName := DiceConfig.WorkHome + aTs.DR.SuperFolder + "/" + ep.CurrentStage.Name + "-output.log"
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		SRf(out, "Failed to save stage log : %s\n", err)
		return err
	}
	defer logFile.Sync()
	defer logFile.Close()

	stageLog.SetOutput(logFile)
	stageLog.SetFormatter(&log.JSONFormatter{DisableTimestamp: true})

	SR(out, []byte("Initializing stage log file with success"))

	SRf(out, "cmd => '%s'\n", cmdTxt)
	if !dryRun {
		return ep.LinuxCommandExecutor(ctx, cmdTxt, stageLog, out)
	} else {
		testData, err := DiceConfig.LoadTestOutput(ep.CurrentStage.TileName, aTs.DR.SuperFolder)
		if err != nil {
			log.Printf("No testing output for %s\n", ep.CurrentStage.TileName)
		} else {
			logFile.Write(testData)
		}
	}
	return nil
}

func (ep *ExecutionPlan) LinuxCommandExecutor(ctx context.Context, cmdTxt []byte, stageLog *log.Logger, out *websocket.Conn) error {
	ct := strings.TrimSpace(string(cmdTxt))
	cts := strings.Split(ct, " ")
	cmd := exec.Command(cts[0], cts[1:]...)

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	// Wait for logs flush out
	wg := new(sync.WaitGroup)
	err := cmd.Start()
	if err != nil {
		SRf(out, "cmd.Start() failed with '%s'\n", err)
		return err
	}
	go ep.WsTail(ctx, stdoutIn, stageLog, wg, out)
	wg.Add(1)
	go ep.WsTail(ctx, stderrIn, stageLog, wg, out)
	wg.Add(1)
	go func() {
		select {
		case <-ctx.Done():
			err := cmd.Process.Kill()
			log.Printf("halted cmd with %s\n", err)
		}
	}()

	err = cmd.Wait()
	wg.Wait()

	if err != nil {
		SRf(out, "cmd.Run() failed with %s\n", err)
	}
	return err
}

// CommandWrapperExecutor wrap commands as a unix script in order to execute.
func (ep *ExecutionPlan) CommandWrapperExecutor(ctx context.Context, dryRun bool, out *websocket.Conn) (string, error) {
	//stage.WorkHome
	script := ep.CurrentStage.WorkHome + "/script-" + ep.CurrentStage.Name + "-" + utils.RandString(8) + ".sh"
	// context id
	dSid := ctx.Value(`d-sid`).(string)
	tContent := `#!/bin/bash
set -xe
{{range .InjectedEnv}}
{{.}}
{{end}}
{{range .Preparation}}
{{.}}
{{end}}

{{range .Commands}}
{{.}}
{{end}}
echo $?
`

	tContent4K8s := `#!/bin/bash
set -xe
[kube.config]
{{range .InjectedEnv}}
{{.}}
{{end}}
{{range .Preparation}}
{{.}}
{{end}}

{{range .Commands}}
{{.}}
{{end}}
echo $?
`
	//Inject kube.config if need to
	if at, ok := AllTs[dSid]; ok {
		// Looking for initial kube.config. For EKS, require clusterName, masterRoleARN ; For others, not implementing.
		if tile, ok := at.AllTilesN[ep.CurrentStage.Name]; ok {
			if tile.Spec.Manifests.ManifestType != "" || tile.Metadata.DependentOnVendorService == v1alpha1.EKS.VSString() {
				var clusterName, masterRoleARN string
				// Tile with dependency
				if dependentTile := DependentEKSTile(dSid, tile.TileInstance); dependentTile != nil {
					if outputs, ok := (*at.AllOutputsN)[dependentTile.TileInstance]; ok {
						if cn, ok := (*outputs.TsOutputs)["clusterName"]; ok {
							clusterName = cn.OutputValue
						}
						if arn, ok := (*outputs.TsOutputs)["masterRoleARN"]; ok {
							masterRoleARN = arn.OutputValue
						}
					}
				}
				// Tile without dependency but input parameters
				if s, ok := at.TsStacksMapN[ep.CurrentStage.Name]; ok && (clusterName == "" || masterRoleARN == "") {
					if inputParameters, ok := s.InputParameters["clusterName"]; ok {
						clusterName = ep.ReplaceAllValueRef(inputParameters.InputValue, dSid, ep.CurrentStage.Name)
						clusterName = ep.ReplaceAllValueRef(inputParameters.InputValue, dSid, "")
					}
					if inputParameters, ok := s.InputParameters["masterRoleARN"]; ok {
						masterRoleARN = ep.ReplaceAllValueRef(inputParameters.InputValue, dSid, ep.CurrentStage.Name)
						masterRoleARN = ep.ReplaceAllValueRef(inputParameters.InputValue, dSid, "")
					}
				}

				if (clusterName == "" || masterRoleARN == "") && !dryRun {
					return script, errors.New("ContainerProvider with EKS didn't include output: clusterName & masterRoleARN")
				}

				tContent4K8s = strings.ReplaceAll(tContent4K8s, "[kube.config]",
					fmt.Sprintf("aws eks update-kubeconfig --name %s --role-arn %s --kubeconfig %s\nexport KUBECONFIG=%s",
						clusterName,
						masterRoleARN,
						DiceConfig.WorkHome+at.DR.SuperFolder+at.TsStacksMapN[tile.TileInstance].TileFolder+"/kube.config",
						DiceConfig.WorkHome+at.DR.SuperFolder+at.TsStacksMapN[tile.TileInstance].TileFolder+"/kube.config",
					))
				tContent = tContent4K8s
			}
		}
		////

		// Inject output values as env which can be retrieved only after execution
		dependentTiles := AllDependentTiles(dSid, ep.CurrentStage.Name)
		for _, tile := range dependentTiles {
			if to, ok := (*at.AllOutputsN)[tile.TileInstance]; ok {
				for k, v := range *to.TsOutputs {
					//$D-TBD_TileName.Output-Name
					if v.OutputValue != "" {
						ep.CurrentStage.InjectedEnv = append(ep.CurrentStage.InjectedEnv, fmt.Sprintf("export D_TBD_%s_%s=%s",
							strcase.ToScreamingSnake(strings.ToUpper(ep.CurrentStage.TileName)),
							strings.ToUpper(k),
							v.OutputValue))
					}
				}
			}
		}
		////
	}

	tp := template.New("script")
	tp, err := tp.Parse(tContent)
	if err != nil {
		return script, err
	}

	file, err := os.OpenFile(script, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //Create(script)
	if err != nil {
		SR(out, []byte(err.Error()))
		return script, err
	}
	defer file.Close()

	// !!! Replace $(value) to actual value !!!
	for _, kvs := range [][]string{ep.CurrentStage.InjectedEnv,
		ep.CurrentStage.Preparation,
		ep.CurrentStage.Commands} {
		for i, _ := range kvs {

			kvs[i] = ep.ReplaceAllValueRef(kvs[i], dSid, ep.CurrentStage.Name) //replace 'self'
			kvs[i] = ep.ReplaceAllValueRef(kvs[i], dSid, "")                   //replace 'anything else'
		}
	}
	// !!! setup probe !!!
	for i, _ := range ep.CurrentStage.Preparation {
		if strings.Contains(ep.CurrentStage.Preparation[i], "dice-probe-") {
			newCmd, err := ep.ProbeWrapper(ctx, ep.CurrentStage.Preparation[i], ep.CurrentStage.ProbeCommands[ep.CurrentStage.Preparation[i]])
			if err != nil {
				return script, err
			}
			ep.CurrentStage.Preparation[i] = newCmd
		}
	}
	////

	err = tp.Execute(file, ep.CurrentStage)
	if err != nil {
		SR(out, []byte(err.Error()))
		return script, err
	}
	// Show script
	buf, err := ioutil.ReadFile(script)
	SRf(out, "Generated script -  %s with content: \n", script)
	SR(out, []byte("--BO:-------------------------------------------------"))
	SR(out, buf)
	SR(out, []byte("--EO:-------------------------------------------------"))

	return script, err

}

func (ep *ExecutionPlan) ProbeWrapper(ctx context.Context, id string, probe v1alpha1.ReadinessProbe) (string, error) {
	dSid := ctx.Value(`d-sid`).(string)
	script := ep.CurrentStage.WorkHome + "/" + id + "-" + utils.RandString(8) + ".sh"
	tContent := `#!/bin/bash
set -xe

{{.}}
`
	file, err := os.OpenFile(script, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //Create(script)
	if err != nil {
		return script, err
	}
	defer file.Close()

	tp := template.New("script")
	tp, err = tp.Parse(tContent)
	if err != nil {
		return script, err
	}

	// !!! Replace $(value) to actual value !!!
	probe.Command = ep.ReplaceAllValueRef(probe.Command, dSid, ep.CurrentStage.Name) //replace 'self'
	probe.Command = ep.ReplaceAllValueRef(probe.Command, dSid, "")                   //replace 'anything else'
	////

	err = tp.Execute(file, probe.Command)
	if err != nil {
		return script, err
	}
	return fmt.Sprintf("probe -command %s "+
		"-initialDelaySeconds %d "+
		"-periodSeconds %d "+
		"-timeoutSeconds %d "+
		"-successThreshold %d "+
		"-failureThreshold %d",
		script,
		probe.InitialDelaySeconds,
		probe.PeriodSeconds,
		probe.TimeoutSeconds,
		probe.SuccessThreshold,
		probe.FailureThreshold), nil
}

// WsTail collect output from stdout/stderr, and also catch up defined output value & persist them.
func (ep *ExecutionPlan) WsTail(ctx context.Context, reader io.ReadCloser, stageLog *log.Logger, wg *sync.WaitGroup, out *websocket.Conn) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		buf := scanner.Bytes()
		if stageLog != nil {
			stageLog.Printf("%s", buf)
		}
		SR(out, buf)
	}
	if wg != nil {
		wg.Done()
	}
}

// ExtractValue retrieve values from output logs.
func (ep *ExecutionPlan) ExtractValue(ctx context.Context, buf []byte, out *websocket.Conn) error {
	dSid := ctx.Value(`d-sid`).(string)
	allEnv := ep.ExtractAllEnv()
	if ts, ok := AllTs[dSid]; ok {
		tileInstance := ep.CurrentStage.Name
		var tileCategory string
		//var vendorService string
		if stack, ok := ts.TsStacksMapN[tileInstance]; ok {
			tileCategory = stack.TileCategory
		}

		if outputs, ok := (*ts.AllOutputsN)[tileInstance]; ok {
			outputs.StageName = ep.CurrentStage.Name
			for _, outputName := range outputs.OutputsOrder {
				if outputDetail, ok := (*outputs.TsOutputs)[outputName]; ok {
					if tileCategory == v1alpha1.ContainerApplication.CString() || tileCategory == v1alpha1.Application.CString() {
						// Extract dSid, value from Command outputs
						regx := regexp.MustCompile("^\\{\"(" +
							outputName +
							"=" +
							".*?)\"}$")
						regxO := regexp.MustCompile("^\\{\"level\":\"info\"\\,\"msg\":\"(" +
							outputName +
							"=" +
							".*?)\"}$")
						err := ep.ScanOutput(regx, buf, outputDetail, out)
						if err != nil {
							return err
						}
						err = ep.ScanOutput(regxO, buf, outputDetail, out)
						if err != nil {
							return err
						}

					} else if stack, ok := ts.TsStacksMapN[tileInstance]; ok {
						// Extract dSid, value from CDK outputs
						regx := regexp.MustCompile("^\\{\"level\":\"info\"\\,\"msg\":\"(" +
							strcase.ToCamel(stack.TileName) + "." +
							".*" +
							outputName +
							".*?)\"}$")
						err := ep.ScanOutput(regx, buf, outputDetail, out)
						if err != nil {
							return err
						}

					} else {
						return errors.New("the handler of regular expression wasn't existed")
					}

					// Replace possible ENV in output
					if strings.Contains(outputDetail.OutputValue, "$") {
						outputDetail.OutputValue = ep.ReplaceAllEnv(outputDetail.OutputValue, allEnv)
					}

				} else {
					return errors.New("outputs & outputs' order weren't consistent")
				}
			}
		}

		// Replace possible value reference
		if outputs, ok := (*ts.AllOutputsN)[tileInstance]; ok {
			for outputName, outputDetail := range *outputs.TsOutputs {
				if strings.Contains(outputDetail.OutputValue, "$") {
					outputDetail.OutputValue = ep.ReplaceAllValueRef(outputDetail.OutputValue, dSid, ep.CurrentStage.Name)
					SRf(out, "Replace reference for extracted output: [%s] = [%s] ", outputName, outputDetail.OutputValue)
				}
			}
		}

		// Pass output values to parent stack
		if parentTileInstances := ParentTileInstance(dSid, tileInstance); parentTileInstances != nil {
			for _, parentTileInstance := range parentTileInstances {
				if outputs, ok := (*ts.AllOutputsN)[tileInstance]; ok {
					if parentOutputs, ok := (*ts.AllOutputsN)[parentTileInstance]; ok {
						for k, v := range *outputs.TsOutputs {
							(*parentOutputs.TsOutputs)[k] = v
						}
					}
				}
			}
		}
	}

	return nil
}

func (ep *ExecutionPlan) ScanOutput(regx *regexp.Regexp, buf []byte, outputDetail *TsOutputDetail, out *websocket.Conn) error {
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		txt := scanner.Text()
		match := regx.FindStringSubmatch(txt)
		if len(match) > 1 {
			key, value, err := FindPair(match[1])
			if err != nil {
				return err
			} else if !strings.Contains(key, outputDetail.Name) {
				return errors.New("matched name wasn't expected: " + key)
			} else {
				outputDetail.OutputValue = value
				SRf(out, "Extract outputs: [%s] = [%s] ", outputDetail.Name, outputDetail.OutputValue)
				break
			}
		}
	}
	return nil
}

func FindPair(str string) (string, string, error) {
	key := ""
	value := ""
	kv := strings.Split(str, " = ")
	if len(kv) < 2 {
		kv = strings.Split(str, "=")
	}
	if len(kv) < 2 {
		return "", "", errors.New("found matched string but failed to extract values")
	} else {
		key = strings.TrimSpace(kv[0])
		if len(kv) == 2 {
			value = strings.TrimSpace(kv[1])
		} else {
			value = strings.TrimSpace(str[len(kv[0])+1:])
		}
	}
	return key, value, nil
}

// PostRun manages and executes commands after provision
func (ep *ExecutionPlan) PostRun(ctx context.Context, dryRun bool, out *websocket.Conn) (string, error) {
	stage := ep.CurrentStage
	script := stage.WorkHome + "/script-" + stage.Name + "-Post-" + utils.RandString(8) + ".sh"
	file, err := os.OpenFile(script, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //Create(script)
	if err != nil {
		SR(out, []byte(err.Error()))
		return script, err
	}
	defer file.Close()

	tContent := `#!/bin/bash
set -xe
{{range .InjectedEnv}}
{{.}}
{{end}}


{{range .PostRunCommands}}
{{.}}
{{end}}
echo $?
`
	// Injected output values of current Tile as env
	dSid := ctx.Value(`d-sid`).(string)
	if at, ok := AllTs[dSid]; ok {
		if tile, ok := at.AllTilesN[stage.Name]; ok {
			if to, ok := (*at.AllOutputsN)[tile.TileInstance]; ok {
				for k, v := range *to.TsOutputs {
					//$D-TBD_TileName.Output-Name
					if v.OutputValue != "" {
						stage.InjectedEnv = append(stage.InjectedEnv, fmt.Sprintf("export D_TBD_%s_%s=%s",
							strcase.ToScreamingSnake(strings.ToUpper(stage.TileName)),
							strings.ToUpper(k),
							v.OutputValue))
					}
				}
			}
		}
	}

	tp := template.New("script")
	tp, err = tp.Parse(tContent)
	if err != nil {
		return script, err
	}

	// Replace reference value
	for i, _ := range ep.CurrentStage.PostRunCommands {
		ep.CurrentStage.PostRunCommands[i] = ep.ReplaceAllValueRef(ep.CurrentStage.PostRunCommands[i], dSid, ep.CurrentStage.Name) //replace 'self'
		ep.CurrentStage.PostRunCommands[i] = ep.ReplaceAllValueRef(ep.CurrentStage.PostRunCommands[i], dSid, "")                   //replace 'anything else'
	}
	// !!! setup probe !!!
	for i, _ := range ep.CurrentStage.PostRunCommands {
		if strings.Contains(ep.CurrentStage.PostRunCommands[i], "dice-probe-") {
			newCmd, err := ep.ProbeWrapper(ctx, ep.CurrentStage.PostRunCommands[i], ep.CurrentStage.ProbeCommands[ep.CurrentStage.PostRunCommands[i]])
			if err != nil {
				return script, err
			}
			ep.CurrentStage.PostRunCommands[i] = newCmd
		}
	}
	////

	err = tp.Execute(file, stage)
	if err != nil {
		SR(out, []byte(err.Error()))
		return script, err
	}
	// Show script
	cnt, err := ioutil.ReadFile(script)
	SRf(out, "Generated script -  %s with content: \n", script)
	SR(out, []byte("--BO:-------------------------------------------------"))
	SR(out, cnt)
	SR(out, []byte("--EO:-------------------------------------------------"))

	return script, nil

}

//
func setStatus(dSid string, tileInstance string, status string) {
	if tilesGrid, ok := AllTilesGrids[dSid]; ok {
		if tg, ok := (*tilesGrid)[tileInstance]; ok {
			tg.Status = status
		}
	}
}
