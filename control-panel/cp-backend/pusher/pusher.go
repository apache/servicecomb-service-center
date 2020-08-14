package pusher

import (
	"encoding/json"
	"github.com/alec-z/cp-backend/model"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

var Events chan model.ServerEvent

func Websocket(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			event, ok := <-Events
			if !ok {
				c.Logger().Error("The main event chan is closed!")
			}
			b, err := json.Marshal(event)
			if err != nil {
				c.Logger().Error(err)
				continue
			}
			err = websocket.Message.Send(ws, string(b))
			if err != nil {
				c.Logger().Error(err)
			}
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}
