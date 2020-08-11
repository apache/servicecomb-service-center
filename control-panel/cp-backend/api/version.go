package api

import (
	"github.com/alec-z/cp-backend/model"
	"github.com/labstack/echo/v4"
	"net/http"
)

func VersionGet(c echo.Context)  (err error) {
	version := model.Version{Name: "control-panel", Tag: "0.0.1"}
	return c.JSON(http.StatusOK, version)
}