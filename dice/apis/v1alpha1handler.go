package apis

import (
	"context"
	"dice/engine"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
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

	dryRun := c.Query("dryRun") == "true"
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
			engine.SR(wb.out, []byte(err.Error()))
		}
		// Signal client all done & Could close connection if need to
		engine.SR(wb.out, []byte("d-done"))

	}

}

func WsCloseHandler(cancel context.CancelFunc, code int, txt string) error {
	log.Printf("WebSocket connection was closed...error: %d - %s\n", code, txt)
	cancel()
	return nil
}

func (wb *WsBox) Processor(ctx context.Context, messageType int, p []byte, dryRun bool) error {
	sid := uuid.New().String()
	ctx = context.WithValue(ctx, "d-sid", sid)
	engine.SRf(wb.out, "Created a new session & d-sid = %s", sid)
	var ep *engine.ExecutionPlan
	//
	// 1. parse yaml +
	engine.SR(wb.out, []byte("Parsing Deployment..."))
	dt := engine.Data(p)
	deploy, err := dt.ParseDeployment(ctx)
	if err != nil {
		engine.SRf(wb.out, "Parsing Deployment error : %s \n", err)
		engine.SRf(wb.out, "!!! Treat input < %s > as command and to be executing...\n", p)
		if err := ep.CommandExecutor(ctx, nil, p, wb.out); err != nil {
			engine.SRf(wb.out, "CommandExecutor error : %s \n", err)
			return err
		}
		return nil
	}
	engine.SR(wb.out, []byte("Parsing Deployment was success."))
	engine.SR(wb.out, []byte("--BO:-------------------------------------------------"))
	b, _ := yaml.Marshal(deploy)
	engine.SR(wb.out, b)
	engine.SR(wb.out, []byte("--EO:-------------------------------------------------"))

	//
	// 2. assemble super app with base templates +
	engine.SR(wb.out, []byte("Generating CDK App..."))
	ep, err = deploy.GenerateCdkApp(ctx, wb.out)
	if err != nil {
		engine.SRf(wb.out, "GenerateCdkApp error : %s \n", err)
		return err
	}
	engine.SR(wb.out, []byte("Generating CDK App was success."))

	//
	// 3. execute cdk / manifest +
	return ep.ExecutePlan(ctx, dryRun, wb.out)

}

func RetrieveTemplate(ctx context.Context, c *gin.Context) {
	name := c.Param("name")
	switch name {
	case "tile":
		tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s/%s.tgz",
			engine.DiceConfig.BucketName,
			engine.DiceConfig.Region,
			"sample-tile",
			"0.1.0",
			"sample-tile")
		c.String(http.StatusOK, tileUrl)
	case "deployment":
		c.String(http.StatusOK, "not ready yet")
	case "hu":
		c.String(http.StatusOK, "not ready yet")
	case "super":
		c.String(http.StatusOK, "not ready yet")
	}
}

func AtsContent(ctx context.Context, c *gin.Context) {
	sid := c.Param("sid")
	if at, ok := engine.AllTs[sid]; ok {
		if buf, err := yaml.Marshal(at); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.String(http.StatusOK, string(buf))
		}

	} else {
		c.String(http.StatusNotFound, "Session ID : %s is not existed and checked out with CC.", sid)
	}
}

func Deployment(ctx context.Context, c *gin.Context) {
	buf, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := engine.Data(buf)
	deployment, err := d.ParseDeployment(ctx)
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
	tile, err := d.ParseTile(ctx)
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
