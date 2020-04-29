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
	Processor(messageType int, p []byte) ([]byte, error)

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
		_, err = wb.Processor(mt, message)
		if err != nil {
			engine.SendResponse(wb.out, []byte(err.Error()))
		}
	}

}

func WsCloseHandler(code int, txt string) error {
	log.Printf("WebSocket connection was closed...error: %d - %s\n", code, txt)
	return nil
}

func (wb *WsBox) Processor(messageType int, p []byte) ([]byte, error) {
	var ep *engine.ExecutionPlan
	//
	// 1. parse yaml +
	engine.SendResponse(wb.out,[]byte("Parsing Deployment..."))
	dt := engine.Data(p)
	deploy, err := dt.ParseDeployment()
	if err != nil {
		engine.SendResponsef(wb.out,"Parsing Deployment error : %s \n", err)
		engine.SendResponsef(wb.out,"!!! Treat input < %s > as command and to be executing...\n", p)
		resp, err := ep.CommandExecutor(p, wb.out)
		if err != nil {
			engine.SendResponsef(wb.out,"CommandExecutor error : %s \n", err)
			return resp, err
		}
		return resp, err
	}
	engine.SendResponse(wb.out,[]byte("Parsing Deployment was success."))
	engine.SendResponse(wb.out,[]byte("---"))
	b, _ := yaml.Marshal(deploy)
	engine.SendResponse(wb.out,b)
	engine.SendResponse(wb.out,[]byte("---"))

	//
	// 2. assemble super app with base templates +
	engine.SendResponse(wb.out,[]byte("Generating CDK App..."))
	ep, err = deploy.GenerateCdkApp(wb.out)
	if err != nil {
		engine.SendResponsef(wb.out,"GenerateCdkApp error : %s \n", err)
		return nil, err
	}
	engine.SendResponse(wb.out,[]byte("Generating CDK App was success."))

	//
	// 3. execute cdk / manifest +
	return nil, ep.ExecutePlan(wb.out)


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
