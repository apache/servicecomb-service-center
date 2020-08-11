package main

import (
	"github.com/alec-z/cp-backend/api"
	"github.com/alec-z/cp-backend/change_decter"
	"github.com/alec-z/cp-backend/pusher"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// New Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	pusher.Events = change_decter.Events
	go change_decter.SampleWorker()

	// WebSocket
	e.GET("/websocket", pusher.Websocket)
	// Routes
	e.GET("/api/version", api.VersionGet)

	// Start server
	e.Logger.Info(e.Start(":3000"))
}