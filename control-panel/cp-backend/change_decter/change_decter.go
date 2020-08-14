package change_decter

import (
	"github.com/alec-z/cp-backend/model"
	"time"
)

var Events chan model.ServerEvent

const ServerEventBufferSize = 100

func init() {
	Events = make(chan model.ServerEvent, ServerEventBufferSize)
}

func SampleWorker() {
	for {
		event := model.ServerEvent{
			EventType: "ServiceChanged",
			Content:   "Service A is down, Service B is up",
			CreatedAt: time.Now(),
		}
		time.Sleep(3 * time.Second)
		Events <- event
	}
}
