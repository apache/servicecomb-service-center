package change_decter

import (
	"encoding/json"
	"fmt"
	"github.com/go-chassis/go-chassis/pkg/scclient"
	"github.com/go-chassis/go-chassis/pkg/scclient/proto"
)

const ServerEventBufferSize = 100

var registryClient client.RegistryClient
var serviceId string
var EventsChan chan *client.MicroServiceInstanceChangedEvent

func initClient() {
	err := registryClient.Initialize(
		client.Options{
			Addrs: []string{"servicecenter:30100"},
		})
	if err != nil {
		fmt.Printf("err[%v]\n", err)
		return
	}
}

func registry() {
	cpService := &proto.MicroService{
		AppId:       "default",
		ServiceName: "control-panel",
		Version:     "0.0.1",
		Environment: "",
	}
	var err error
	serviceId, err = registryClient.RegisterService(cpService)

	if err != nil {
		fmt.Printf("err[%v]\n", err)
		return
	}

	fmt.Printf("sid[%v]\n", serviceId)
}

func notifyDependencyRelation(microServices []*proto.MicroService) {
	for _, m := range microServices {
		_, _ = registryClient.FindMicroServiceInstances(serviceId, m.AppId, m.ServiceName, m.Version)
	}
}

func notifyEvent(event *client.MicroServiceInstanceChangedEvent) {
	EventsChan <- event
	content, _ := json.Marshal(event)
	fmt.Printf("event[%v]\n", string(content))
}

func init() {
	EventsChan = make(chan *client.MicroServiceInstanceChangedEvent, ServerEventBufferSize)

	initClient()
	// registry self
	registry()
	//get all microservices
	services, err := registryClient.GetAllMicroServices()
	if err != nil {
		fmt.Println(err)
		return
	}
	//notifyDependencyRelation
	notifyDependencyRelation(services)
	//watch
	registryClient.WatchMicroService(serviceId, notifyEvent)
}
