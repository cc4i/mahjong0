package apis

import (
	"bufio"
	"context"
	"dice/engine"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net/http"
	"os/exec"
	"sigs.k8s.io/yaml"
	"strings"
)


var upGrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WsBox struct {
	out *websocket.Conn
}

type WsWorker interface {
	Processor(messageType int, p []byte)([]byte, error)
	CommandExecutor(cmd []byte)([]byte, error)
	SendResponse(response []byte) error
	WsTail(reader io.ReadCloser)
}


func WsHandler(ctx context.Context, c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	ws.SetCloseHandler(func(code int, txt string) error {
		return WsCloseHandler(code, txt)
	})
	defer ws.Close()

	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("recv: %s\n", message)


		wb := WsBox{out: ws}
		_,err = wb.Processor(mt, message)
		if err != nil {
			wb.SendResponse([]byte(err.Error()))

		}

	}

}

func WsCloseHandler(code int, txt string) error {
	log.Printf("WebSocket connection was closed...error: %d - %s\n", code, txt)
	return nil
}

func (wb *WsBox) Processor(messageType int, p []byte)([]byte, error) {

	//
	// 1. parse yaml +
	wb.SendResponse([]byte("Parsing Deployment..."))
	dt := engine.Data(p)
	deploy, err := dt.ParseDeployment()
	if err != nil {
		wb.SendResponsef("Parsing Deployment error : %s \n", err)
		wb.SendResponsef("!!! Treat input < %s > as command and to be executing...\n", p)
		resp, err := wb.CommandExecutor(p)
		if err != nil {
			wb.SendResponsef("CommandExecutor error : %s \n", err)
			return resp, err
		}
		return resp, err
	}
	wb.SendResponse([]byte("Parsing Deployment was success."))
	wb.SendResponse([]byte("---"))
	b, _ := yaml.Marshal(deploy)
	wb.SendResponse(b)
	wb.SendResponse([]byte("---"))


	//
	// 2. assemble super app with base templates +
	wb.SendResponse([]byte("Generating CDK App..."))
	_, err = deploy.GenerateCdkApp(wb.out)
	if err != nil {
		wb.SendResponsef("GenerateCdkApp error : %s \n", err)
		return nil, err
	}
	wb.SendResponse([]byte("Generating CDK App was success."))


	//
	// 3. execute cdk / manifest +
	//resp,err := wb.CommandExecutor(p)
	//if err != nil {
	//	log.Printf("commandExecutor error : %s \n", err)
	//	return nil, err
	//}
	//
	return nil, err
}

func (wb *WsBox)SendResponse(response []byte) error {
	log.Printf("%s\n", response)
	err := wb.out.WriteMessage(websocket.TextMessage, response)
	if err !=nil {
		log.Printf("write error: %s\n", err)
	}
	return err

}

func (wb *WsBox)SendResponsef(format string, v ...interface{}) error {
	str := fmt.Sprintf(format, v...)
	log.Println(str)
	err := wb.out.WriteMessage(websocket.TextMessage, []byte(str))
	if err !=nil {
		log.Printf("write error: %s\n", err)
	}
	return err

}

func (wb *WsBox) CommandExecutor(cmdTxt []byte)([]byte, error) {
	ct := string(cmdTxt)
	cts := strings.Split(ct, " ")
	cmd := exec.Command(cts[0], cts[1:len(cts)]...)

	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()


	err := cmd.Start()
	if err != nil {
		wb.SendResponsef("cmd.Start() failed with '%s'\n", err)
		return nil, err
	}
	go wb.WsTail(stdoutIn)
	go wb.WsTail(stderrIn)

	err = cmd.Wait()
	if err != nil {
		wb.SendResponsef("cmd.Run() failed with %s\n", err)
		return nil, err
	}

	return nil, nil
}


func (wb *WsBox)WsTail(reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		buf := scanner.Bytes()
		if err := wb.out.WriteMessage(websocket.TextMessage, buf); err !=nil {
			log.Printf("write error: %s\n", err)
		}
	}

}


func Deployment(ctx context.Context, c *gin.Context) {
	buf, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := engine.Data(buf)
	deployment, err := d.ParseDeployment()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, _ := yaml.Marshal(deployment)
	c.String(http.StatusOK, string(b))
}

func Tile(ctx context.Context, c *gin.Context) {

	buf, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := engine.Data(buf)
	tile, err := d.ParseTile()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, _ := yaml.Marshal(tile)
	c.String(http.StatusOK, string(b))
}


func allowCORS(ctx context.Context, c *gin.Context) {

	// allow CORS
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

}