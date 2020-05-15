package apis

import (
	"context"
	"dice/utils"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime"
)

func Router(ctx context.Context) *gin.Engine {
	r := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("server-session", store))

	// Health API
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Deployment API through WebSocket
	r.GET("/v1alpha1/ws", func(c *gin.Context) {
		WsHandler(ctx, c)
	})

	// Deployment API through WebSocket, but dry run only
	r.GET("/v1alpha1/ws?dryRun=true", func(c *gin.Context) {
		WsHandler(ctx, c)
	})

	// Return url of basic templates as per request
	r.GET("/v1alpha1/template/:name", func(c *gin.Context) {
		RetrieveTemplate(ctx, c)
	})


	// Validate Tile specification
	r.POST("/v1alpha1/tile", func(c *gin.Context) {
		Tile(ctx, c)
	})

	// Validate Deployment specification
	r.POST("/v1alpha1/deployment", func(c *gin.Context) {
		Deployment(ctx, c)
	})

	// AllTs content in memory
	r.GET("/v1alpha1/ats/:sid", func(c *gin.Context) {
		AtsContent(ctx, c)
	})

	// Version
	r.GET("/version", func(c *gin.Context) {
		version := fmt.Sprintf("\tVersion:\t%s\n\tGo version:\t%s\n\tGit commit:\t%s\n\tBuilt:\t%s\n\tOS/Arch:\t%s/%s\n",
			utils.ServerVersion,
			runtime.Version(),
			utils.GitCommit,
			utils.Built,
			runtime.GOOS, runtime.GOARCH)
		c.String(http.StatusOK, version)
	})

	r.Use(static.Serve("/toy", static.LocalFile("./toy", true)))

	return r
}
