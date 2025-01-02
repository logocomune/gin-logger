package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/logocomune/gin-logger"
	"log/slog"
)

func main() {
	r := gin.New()
	logger := slogger.New(context.Background(),
		slogger.WithStaticLogEntries(map[string]string{
			"_appName":    "test",
			"_appVersion": "1.0.0",
		}))
	r.Use(logger.Middleware())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	slog.Info("starting server. Test at http://localhost:8080/ping")
	r.Run(":8080")
}
