package apis

import (
	"context"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
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

	// Validate Tile specification
	r.POST("/v1alpha1/tile", func(c *gin.Context) {
		Tile(ctx, c)
	})

	// Validate Deployment specification
	r.POST("/v1alpha1/deployment", func(c *gin.Context) {
		Deployment(ctx, c)
	})

	return r
}
