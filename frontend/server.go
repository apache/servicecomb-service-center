package main

import (
	"log"
	"net/url"
	"os"
	"path/filepath"

	"github.com/apache/incubator-servicecomb-service-center/frontend/schema"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func Serve(c Config) {
	e := echo.New()
	e.HideBanner = true
	// handle all requests by serving a file of the same name
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cant get cwd, error:%s", err)
	}
	staticPath := filepath.Join(dir, "app")
	e.Static("/", staticPath)

	e.Any("/testSchema/", schema.SchemaHandleFunc)

	scProxy(c, e)

	log.Printf("Error: %s", e.Start(c.frontendAddr))
}

// setup proxy for requests to service center
func scProxy(c Config, e *echo.Echo) {
	scUrl, err := url.Parse(c.scAddr)
	if err != nil {
		log.Fatalf("Error parsing service center address:%s, err:%s", c.scAddr, err)
	}

	targets := []*middleware.ProxyTarget{
		{
			URL: scUrl,
		},
	}
	g := e.Group("/sc")
	balancer := middleware.NewRoundRobinBalancer(targets)
	pcfg := middleware.ProxyConfig{
		Balancer: balancer,
		Skipper:  middleware.DefaultSkipper,
		Rewrite: map[string]string{
			"/sc/*": "/$1",
		},
	}
	g.Use(middleware.ProxyWithConfig(pcfg))
}
