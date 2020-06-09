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

// Router tells all routing definition
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
	// Run Linux commands for purpose
	r.GET("/v1alpha1/ws?linuxCommand=true", func(c *gin.Context) {
		WsHandler(ctx, c)
	})

	// Destroy API through WebSocket
	r.GET("/v1alpha1/destroy", func(c *gin.Context) {
		// TODO

	})

	// Return url of basic templates as per request
	r.GET("/v1alpha1/template/:what", func(c *gin.Context) {
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
	r.GET("/v1alpha1/ts/:sid", func(c *gin.Context) {
		Ts(ctx, c)
	})
	// List deployments in memory
	r.GET("/v1alpha1/ts", func(c *gin.Context) {
		AllTsD(ctx, c)
	})


	// Version of Dice
	r.GET("/version", func(c *gin.Context) {
		version := fmt.Sprintf("\tVersion:\t%s\n\tGo version:\t%s\n\tGit commit:\t%s\n\tBuilt:\t%s\n\tOS/Arch:\t%s/%s\n",
			utils.ServerVersion,
			runtime.Version(),
			utils.GitCommit,
			utils.Built,
			runtime.GOOS, runtime.GOARCH)
		c.String(http.StatusOK, version)
	})

	// '/toy' is a page for quick testing
	r.Use(static.Serve("/toy", static.LocalFile("./toy", true)))

	return r
}
