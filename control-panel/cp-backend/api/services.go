package api

import (
	"encoding/json"
	"fmt"
	client "github.com/go-chassis/go-chassis/pkg/scclient"
	"github.com/labstack/echo/v4"
	"net/http"
)

var scClient client.RegistryClient

func GetAllServices(c echo.Context) (err error) {
	services, err := scClient.GetAllMicroServices()
	if err != nil {
		fmt.Println(err)
		return c.JSON(http.StatusInternalServerError, "cannot get all microservice list from service center")
	}
	sb, err := json.Marshal(services)
	return c.JSON(http.StatusOK, string(sb))
}

func init() {
	err := scClient.Initialize(
		client.Options{
			Addrs: []string{"servicecenter:30100"},
		})
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		return
	}
}
