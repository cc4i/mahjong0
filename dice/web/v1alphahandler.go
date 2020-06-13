package web

import (
	"context"
	"dice/apis/v1alpha1"
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

// WsBox is WebSocket struct with connection only so far
type WsBox struct {
	out *websocket.Conn
}

// WsWorker interface for all Websocket handlers
type WsWorker interface {
	// Processor acts as core interface between client and engine
	Processor(ctx context.Context, messageType int, p []byte, dryRun bool) error
	// SaveStatus persists the status of each deployment
	SaveStatus(ctx context.Context)
}

// WsHandler handle all coming request from WebSocket
func WsHandler(ctx context.Context, c *gin.Context) {
	log.Printf("%s connected to %s \n", c.Request.RemoteAddr, c.Request.RequestURI)
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

	linuxCommand := c.Query("linuxCommand") == "true"
	dryRun := c.Query("dryRun") == "true"
	for {
		mt, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			break
		}
		log.Printf("recv: %s\n", message)

		wb := WsBox{out: ws}
		if linuxCommand {
			ep := &engine.ExecutionPlan{}
			ep.LinuxCommandExecutor(stx, message, nil, wb.out)
		} else {
			err = wb.Processor(stx, mt, message, dryRun)
		}
		if err != nil {
			engine.SR(wb.out, []byte(err.Error()))
		}
		// Signal client all done & Could close connection if need to
		engine.SR(wb.out, []byte("d-done"))
	}
}

// WsCloseHandler handle close connection
func WsCloseHandler(cancel context.CancelFunc, code int, txt string) error {
	log.Printf("WebSocket connection was closed...error: %d - %s\n", code, txt)
	cancel()
	return nil
}

// Processor handle full process of deployment request
func (wb *WsBox) Processor(ctx context.Context, messageType int, p []byte, dryRun bool) error {
	var ep *engine.ExecutionPlan
	//
	// 1. Parsing YAML
	engine.SR(wb.out, []byte("Parsing Deployment..."))
	dt := v1alpha1.Data(p)
	deployment, err := dt.ParseDeployment(ctx)
	if err != nil {
		engine.SRf(wb.out, "Parsing Deployment error : %s \n", err)
		return err
	}
	engine.SR(wb.out, []byte("Parsing Deployment was success."))
	engine.SR(wb.out, []byte("--BO:-------------------------------------------------"))
	b, _ := yaml.Marshal(deployment)
	engine.SR(wb.out, b)
	engine.SR(wb.out, []byte("--EO:-------------------------------------------------"))

	// 2. Looking for the dSid of last deployment
	rdSid, isRepeated:= engine.IsRepeatedDeployment(deployment.Metadata.Name)
	if isRepeated {
		engine.SRf(wb.out, "Repeated deployment and last d-dSid = %s", rdSid)
	}
	dSid := uuid.New().String()
	ctx = context.WithValue(ctx, "d-sid", dSid)
	engine.SRf(wb.out, "Created a new d-dSid = %s", dSid)

	//
	// 3. assemble super app with base templates +
	engine.SR(wb.out, []byte("Generating main app..."))
	assemble := engine.AssembleData{
		Deployment: deployment,
	}
	ep, err = assemble.GenerateMainApp(ctx, wb.out)
	if err != nil {
		engine.SRf(wb.out, "GenerateMainApp error : %s \n", err)
		return err
	}
	engine.SR(wb.out, []byte("Generating main app... with success"))

	//
	// 4. execute cdk / manifest +
	err = ep.ExecutePlan(ctx, dryRun, wb.out)
	if err != nil {
		if aTs, ok := engine.AllTs[dSid]; ok {
			engine.UpdateDR(aTs.DR, engine.Interrupted.DSString())
		}
	} else {
		if aTs, ok := engine.AllTs[dSid]; ok {
			engine.UpdateDR(aTs.DR, engine.Done.DSString())
		}
	}
	return err

}

// RetrieveTemplate download template from S3 repo.
func RetrieveTemplate(ctx context.Context, c *gin.Context) {
	what := c.Param("what")

	switch what {
	case "sample-tile":
		tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s/%s.tgz",
			engine.DiceConfig.BucketName,
			engine.DiceConfig.Region,
			"sample-tile",
			"0.1.0",
			"sample-tile")
		c.String(http.StatusOK, tileUrl)
	case "tile":
		tileType := c.Query("type")
		tileUrl := fmt.Sprintf("https://%s.s3-%s.amazonaws.com/tiles-repo/%s/%s-tile.tgz",
			engine.DiceConfig.BucketName,
			engine.DiceConfig.Region,
			"tile",
			tileType)
		c.String(http.StatusOK, tileUrl)
	case "deployment":
		c.String(http.StatusOK, "not ready yet")
	case "hu":
		c.String(http.StatusOK, "not ready yet")
	case "super":
		c.String(http.StatusOK, "not ready yet")
	}
}

// Ts shows key content in memory as per sid
func Ts(ctx context.Context, c *gin.Context) {
	sid := c.Param("sid")
	if ts := engine.TsContent(sid); ts != nil {
		if buf, err := yaml.Marshal(ts); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		} else {
			c.String(http.StatusOK, string(buf))
		}

	} else {
		c.String(http.StatusNotFound, "Session ID : %s is not existed and checked out with CC.", sid)
	}
}

// AllTsD shows all recorded deployment in memory
func AllTsD(ctx context.Context, c *gin.Context) {
	c.JSON(http.StatusOK, engine.AllTsDeployment())

}

// Deployment validate deployment yaml
func Deployment(ctx context.Context, c *gin.Context) {
	buf, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := v1alpha1.Data(buf)
	deployment, err := d.ParseDeployment(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, _ := yaml.Marshal(deployment)
	c.String(http.StatusOK, string(b))
}

// Tile validate Tile yaml
func Tile(ctx context.Context, c *gin.Context) {

	buf, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := v1alpha1.Data(buf)
	tile, err := d.ParseTile(ctx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	b, _ := yaml.Marshal(tile)
	c.String(http.StatusOK, string(b))
}

// allowCORS allows cross site access
func allowCORS(ctx context.Context, c *gin.Context) {

	// allow CORS
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

}
