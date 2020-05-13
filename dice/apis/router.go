package apis

import (
	"context"
	"dice/engine"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"sigs.k8s.io/yaml"
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

	r.GET("/v1alpha1/template/tile", func(c *gin.Context) {
		//TODO
		// 1. CDK style/ Application style
		// 2. Added tile-spec.yaml
		c.String(http.StatusOK, "building...")
	})
	r.GET("/v1alpha1/template/deployment", func(c *gin.Context) {
		//TODO
		// 1. generate a deployment-spec
		c.String(http.StatusOK, "building...")
	})

	// Validate Tile specification
	r.POST("/v1alpha1/tile", func(c *gin.Context) {
		Tile(ctx, c)
	})

	// Validate Deployment specification
	r.POST("/v1alpha1/deployment", func(c *gin.Context) {
		Deployment(ctx, c)
	})

	r.GET("/v1alpha1/session/:sid", func(c *gin.Context) {
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
	})

	r.Use(static.Serve("/toy", static.LocalFile("./toy", true)))

	return r
}
