package pusher

import (
	"encoding/json"
	client "github.com/go-chassis/go-chassis/pkg/scclient"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
)

var Events chan *client.MicroServiceInstanceChangedEvent

var clients = make(map[*websocket.Conn]bool) // connected clients

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Websocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		c.Logger().Error("upgrade errors:", err)
	}
	defer ws.Close()

	clients[ws] = true
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
		err = ws.WriteMessage(1, b)
		if err != nil {
			c.Logger().Error(err)
			delete(clients, ws)
		}
	}
}
