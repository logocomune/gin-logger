package main

import (
	"context"
	slogger "gilhub.com/logocomune/gin-logger"
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func main() {
	r := gin.New()
	logger := slogger.New(context.Background(),
		slogger.WithStaticLogEntries(map[string]string{
			"_appName":    "test",
			"_appVersion": "1.0.0",
		}),
		slogger.WithAggregation(true), slogger.WithTimeAggregation(time.Second*10))
	r.Use(logger.Middleware())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	slog.Info("starting server. Test at http://localhost:8080/ping")
	r.Run(":8080")
}
