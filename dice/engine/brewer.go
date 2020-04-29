package engine

import (
	"bufio"
	"container/list"
	"context"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

// ExecutionPlan represents complete plan.
type ExecutionPlan struct {
	Name string `json:"name"`
	CurrentStage *ExecutionStage `json:"currentStage"`
	Plan       *list.List                `json:"plan"`
	PlanMirror map[string]ExecutionStage `json:"planMirror"`
}

// ExecutionStage represents an unit of execution plan.
type ExecutionStage struct {
	Name          string            `json:"name"`
	Kind          string            `json:"kind"`     //cdk/k8s/helm/kustomization?
	WorkHome      string            `json:"workHome"` //root folder for execution
	Preparation   []string          `json:"preparation"`
	Command       *list.List        `json:"command"`
	CommandMirror map[string]string `json:"commandMirror"`
}

type Ts struct {
	TsLibs   []TsLib
	TsStacks []TsStack
}

type TsLib struct {
	TileName   string
	TileFolder string
}

type TsStack struct {
	TileName          string
	TileVariable      string
	TileStackName     string
	TileStackVariable string
	InputParameters   map[string]string //[]TsInputParameter
	TsManifests       *TsManifests
}

type TsInputParameter struct {
	InputName  string
	InputValue string
}

type TsManifests struct {
	ManifestType string
	Files        []string
	Folders      []string
}

// BrewerCore represent a group of core functions to execute & manage for
// execution plan.
type BrewerCore interface {
	ExecutePlan(ctx context.Context, out *websocket.Conn) error
	CommandExecutor(ctx context.Context, cmd []byte, out *websocket.Conn) ([]byte, error)
	CommandWrapperExecutor(ctx context.Context, stage *ExecutionStage, out *websocket.Conn) (string, error)
	WsTail(reader io.ReadCloser, out *websocket.Conn)
}

//ExecutePlan is a orchestrator to run execution plan.
func (ep *ExecutionPlan) ExecutePlan(ctx context.Context, out *websocket.Conn) error {
	for e := ep.Plan.Back(); e != nil; e = e.Prev() {
		stage := e.Value.(ExecutionStage)
		cmd, err := ep.CommandWrapperExecutor(ctx, &stage, out)
		if err != nil {
			return err
		}
		_, err = ep.CommandExecutor(ctx, []byte(cmd), out)
		if err != nil {
			return err
		}

	}
	return nil
}

// CommandExecutor exec command and return output.
func (ep *ExecutionPlan) CommandExecutor(ctx context.Context, cmdTxt []byte, out *websocket.Conn) ([]byte, error) {
	ct := string(cmdTxt)
	SendResponsef(out, "cmd => '%s'\n", ct)
	cts := strings.Split(ct, " ")
	cmd := exec.Command(cts[0], cts[1:len(cts)]...)

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()

	err := cmd.Start()
	if err != nil {
		SendResponsef(out, "cmd.Start() failed with '%s'\n", err)
		return nil, err
	}
	go ep.WsTail(stdoutIn, out)
	go ep.WsTail(stderrIn, out)

	go func() {
		select {
			case <-ctx.Done():
				err := cmd.Process.Kill()
				log.Printf("halted cmd with %s\n", err)
			}
	}()

	err = cmd.Wait()

	if err != nil {
		SendResponsef(out, "cmd.Run() failed with %s\n", err)
		return nil, err
	}

	return nil, nil
}

// CommandWrapperExecutor wrap commands as a unix script in order to execute.
func (ep *ExecutionPlan) CommandWrapperExecutor(ctx context.Context, stage *ExecutionStage, out *websocket.Conn) (string, error) {
	//stage.WorkHome
	script := stage.WorkHome + "script-" + stage.Name + "-" + randString(8) + ".sh"
	tContent := `#!/bin/bash
set -xe

{{range .Preparation}}
{{.}}
{{end}}

{{$map := .CommandMirror}}
{{range $key, $value := $map }}
{{$value}}
{{end}}
echo $?
`
	tp := template.New("script")
	tp, err := tp.Parse(tContent)
	if err != nil {
		return script, err
	}

	file, err := os.OpenFile(script, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) //Create(script)
	if err != nil {
		SendResponse(out, []byte(err.Error()))
		return script, err
	}
	defer file.Close()
	err = tp.Execute(file, stage)
	if err != nil {
		SendResponse(out, []byte(err.Error()))
		return script, err
	}
	// Show script
	buf, err := ioutil.ReadFile(script)
	SendResponse(out, buf)
	return script, err

}

// WsTail collect output from stdout/stderr, and also catch up defined output value & persist them.
// TODO:
//  Thinking about how to collect repose and persist required output value.
//  Output should from stdout? How to extract them?
func (ep *ExecutionPlan) WsTail(reader io.ReadCloser, out *websocket.Conn) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		buf := scanner.Bytes()
		log.Printf("%s\n", buf)
		if err := out.WriteMessage(websocket.TextMessage, buf); err != nil {
			log.Printf("write error: %s\n", err)
			return
		}
	}

}

func randString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
