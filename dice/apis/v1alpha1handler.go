package apis

import (
	"context"
	"dice/engine"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sigs.k8s.io/yaml"
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
	Processor(ctx context.Context, messageType int, p []byte, dryRun bool) error
}

func WsHandler(ctx context.Context, c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade error:", err)
		return
	}
	stx, cancel := context.WithCancel(ctx)
	ws.SetCloseHandler(func(code int, txt string) error {
		return WsCloseHandler(cancel, code, txt)
	})
	//ws.SetWriteDeadline(time.Now().Add(5 * time.Second))
	defer ws.Close()

	dryRun := c.Query("dryRun")=="true"
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("recv: %s\n", message)

		wb := WsBox{out: ws}
		err = wb.Processor(stx, mt, message, dryRun)
		if err != nil {
			engine.SendResponse(wb.out, []byte(err.Error()))
		}
	}

}

func WsCloseHandler(cancel context.CancelFunc, code int, txt string) error {
	log.Printf("WebSocket connection was closed...error: %d - %s\n", code, txt)
	cancel()
	return nil
}

func (wb *WsBox) Processor(ctx context.Context, messageType int, p []byte, dryRun bool) error {
	var ep *engine.ExecutionPlan
	//
	// 1. parse yaml +
	engine.SendResponse(wb.out, []byte("Parsing Deployment..."))
	dt := engine.Data(p)
	deploy, err := dt.ParseDeployment()
	if err != nil {
		engine.SendResponsef(wb.out, "Parsing Deployment error : %s \n", err)
		engine.SendResponsef(wb.out, "!!! Treat input < %s > as command and to be executing...\n", p)
		_, err := ep.CommandExecutor(ctx, p, wb.out)
		if err != nil {
			engine.SendResponsef(wb.out, "CommandExecutor error : %s \n", err)
			return err
		}
		return err
	}
	engine.SendResponse(wb.out, []byte("Parsing Deployment was success."))
	engine.SendResponse(wb.out, []byte("--BO:-------------------------------------------------"))
	b, _ := yaml.Marshal(deploy)
	engine.SendResponse(wb.out, b)
	engine.SendResponse(wb.out, []byte("--EO:-------------------------------------------------"))

	//
	// 2. assemble super app with base templates +
	engine.SendResponse(wb.out, []byte("Generating CDK App..."))
	ep, err = deploy.GenerateCdkApp(wb.out)
	if err != nil {
		engine.SendResponsef(wb.out, "GenerateCdkApp error : %s \n", err)
		return err
	}
	engine.SendResponse(wb.out, []byte("Generating CDK App was success."))

	//
	// 3. execute cdk / manifest +
	return ep.ExecutePlan(ctx, dryRun, wb.out)

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
