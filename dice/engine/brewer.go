package engine

import (
	"bufio"
	"container/list"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"os/exec"
	"strings"
)

type ExecutionPlan struct {
	Plan *list.List
	PlanMirror map[string]ExecutionStage `json:"plan"`
}

type ExecutionStage struct {
	Name string `json:"name"`
	Kind string `json:"kind"` //cdk/k8s/helm/kustomization?
	WorkHome string `json:"workHome"` //root folder for execution
	Preparation []string `json:"preparation"`
	Command *list.List
	CommandMirror map[string]string `json:"command"`

}

type Ts struct {
	TsLibs []TsLib
	TsStacks []TsStack
}

type TsLib struct {
	TileName string
	TileFolder string
}

type TsStack struct {
	TileName string
	TileVariable string
	TileStackName string
	TileStackVariable string
	InputParameters map[string]string//[]TsInputParameter
	TsManifests *TsManifests
}

type TsInputParameter struct {
	InputName string
	InputValue string
}

type TsManifests struct {
	ManifestType string
	Files []string
	Folders []string
}

type BrewerCore interface {
	ExecutePlan(out *websocket.Conn) error
	CommandExecutor(cmd []byte,out *websocket.Conn) ([]byte, error)
	CommandWrapper(cmd []byte,out *websocket.Conn)([]byte, error)
	WsTail(reader io.ReadCloser, out *websocket.Conn)

}


func (ep *ExecutionPlan)ExecutePlan(out *websocket.Conn) error {
	for e := ep.Plan.Back(); e != nil; e = e.Next() {
		stage := e.Value.(ExecutionStage)
		for c := stage.Command.Back(); c != nil; c = c.Next() {
			_,err := ep.CommandExecutor([]byte(fmt.Sprintf("%s",c.Value)), out)
			if err != nil {
				log.Printf("commandExecutor error : %s \n", err)
				return err
			}
		}
	}
	return nil
}


func (ep *ExecutionPlan) CommandExecutor(cmdTxt []byte, out *websocket.Conn) ([]byte, error) {
	ct := string(cmdTxt)
	SendResponsef(out,"cmd => '%s'\n", ct)
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

	err = cmd.Wait()
	if err != nil {
		SendResponsef(out,"cmd.Run() failed with %s\n", err)
		return nil, err
	}

	return nil, nil
}


func (ep *ExecutionPlan) CommandWrapper(cmd []byte,out *websocket.Conn)([]byte, error) {
	return nil, nil
}

func (ep *ExecutionPlan) WsTail(reader io.ReadCloser, out *websocket.Conn) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		buf := scanner.Bytes()
		log.Printf("%s\n", buf)
		if err := out.WriteMessage(websocket.TextMessage, buf); err != nil {
			log.Printf("write error: %s\n", err)
		}
	}

}

