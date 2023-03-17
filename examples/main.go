package main

import (
	"context"
	"github.com/oarkflow/frame"
	"github.com/oarkflow/frame/server"
	"github.com/oarkflow/swagger"
	_ "github.com/oarkflow/swagger/examples/docs"
)

// PingHandler Test handler
// @Summary Test Summary
// @Description Test Description
// @Accept application/json
// @Produce application/json
// @Router /ping [get]
func PingHandler(c context.Context, ctx *frame.Context) {
	ctx.JSON(200, map[string]string{
		"ping": "pong",
	})
}

// @title HertzTest
// @version 1.0
// @description This is a demo using Hertz.

// @contact.name hertz-contrib
// @contact.url https://github.com/hertz-contrib

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8888
// @BasePath /
// @schemes http
func main() {
	h := server.New()
	h.GET("/swagger/*any", swagger.New())
	h.GET("/ping", PingHandler)
	h.Spin()
}
