package main

import (
	"github.com/apache/servicecomb-service-center/control-panel/cp-backend/api"
	"github.com/apache/servicecomb-service-center/control-panel/cp-backend/change_decter"
	"github.com/apache/servicecomb-service-center/control-panel/cp-backend/pusher"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/url"
)

func main() {
	// New Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	setupServiceCenterAPIProxy(e)
	setupWebSocket(e)

	// Routes
	e.GET("/api/version", api.VersionGet)
	e.GET("/api/services", api.GetAllServices)

	// Start server
	e.Logger.Fatal(e.Start(":3000"))
}

func setupWebSocket(e *echo.Echo) {
	pusher.Events = change_decter.EventsChan
	e.GET("/websocket", pusher.Websocket)
}

func setupServiceCenterAPIProxy(e *echo.Echo) {
	url1, err := url.Parse("http://servicecenter:30100")
	if err != nil {
		e.Logger.Fatal(err)
	}
	targets := []*middleware.ProxyTarget{
		{
			URL: url1,
		},
	}
	g := e.Group("/v4")
	g.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(targets)))

}