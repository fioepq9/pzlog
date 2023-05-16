package pzlog

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func TestGinDefaultWriter(t *testing.T) {
	log := zerolog.New(NewPtermWriter()).With().Timestamp().Logger()
	gin.DefaultWriter = NewGinDefaultWriter(&log, zerolog.DebugLevel)
	r := gin.Default()
	r.GET("/adasdsa", func(ctx *gin.Context) { ctx.JSON(200, "ok") })
	r.GET("/foooooo", func(ctx *gin.Context) { ctx.JSON(200, "ok") })
	r.GET("/barfoqwe", func(ctx *gin.Context) { ctx.JSON(200, "ok") })
	go r.Run(":9090")
	time.Sleep(time.Second)
}
