package event

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/datasource/mongo/sd"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	config := storage.Options{
		URI: "mongodb://localhost:27017",
	}
	client.NewMongoClient(config)
}

func TestInstanceEventHandler_OnEvent(t *testing.T) {

	t.Run("OnEvent test by mongo watch", func(t *testing.T) {
		var store = &sd.TypeStore{}
		store.Initialize()
		h := InstanceEventHandler{}
		h.OnEvent(mongoWatch())
		assert.NotNil(t, discovery.MicroService{})
	})
	t.Run("OnEvent test by request config", func(t *testing.T) {
		h := InstanceEventHandler{}
		h.OnEvent(mongoAssign())
		assert.NotNil(t, discovery.MicroService{})
	})
	t.Run("OnEvent test by request config", func(t *testing.T) {
		h := InstanceEventHandler{}
		h.OnEvent(mongoAssign())
		assert.Error(t, assert.AnError)
	})

}

func TestNotifySyncerInstanceEvent(t *testing.T) {
	evt := mongoAssign()
	ms := getMicroService()
	NotifySyncerInstanceEvent(evt, ms)
}

func mongoWatch() sd.MongoEvent {

	var pipeline []interface{}
	ctx2 := context.Background()
	for  {
		//cur, err := coll.Watch(ctx2,pipeline)
		cur, err := client.GetMongoClient().Watch(ctx2, "instance", pipeline)
		if err != nil {
			log.Fatal(err)
		}
		//defer cur.Close(ctx2)
		for cur.Next(ctx2) { // 下面是数据解析
			resp := MongoWatchResponse{}
			me := sd.MongoEvent{}
			err2 := bson.Unmarshal(cur.Current, &resp)
			switch resp.OperationType {
			case "insert":
				me.Type = discovery.EVT_CREATE
			case "delete":
				me.Type = discovery.EVT_DELETE
			}
			if  err2 != nil {
				log.Fatal("", err)
			}

			fullDocument := Fulldocument{}
			err2 = bson.Unmarshal(resp.MongoEvent, &fullDocument)
			if err2 != nil {
				log.Fatal(fmt.Sprintf("Unmarshal data from db failed : %v", err2))
			}
			instance := discovery.MicroServiceInstance{}
			err3 := bson.Unmarshal(fullDocument.Instanceinfo, &instance)
			if err3 != nil {
				log.Fatal(fmt.Sprintf("Unmarshal Instanceinfo failed : %v", err2))
			}
			me.DocumentId = fullDocument.Id.Hex()
			mInstance := sd.Instance{}
			mInstance.Domain = fullDocument.Domain
			mInstance.Project = fullDocument.Project
			mInstance.InstanceInfo = &instance
			me.Value = mInstance

			return me
		}
	}
}
func mongoAssign() sd.MongoEvent {
	sd.Store().Service().Cache()
	endPoints := []string{"127.0.0.1:27017"}
	instance := discovery.MicroServiceInstance{
		InstanceId: "f73dceb440f711eba63ffa163e7cdcb8",
		ServiceId:  "2a20507274fc71c925d138341517dce14b600744",
		Endpoints: endPoints,
	}
	mInstance := sd.Instance{}
	mInstance.InstanceInfo = &instance
	mInstance.Domain = "default"
	mInstance.Project = "default"
	me := sd.MongoEvent{}
	me.DocumentId = "5fdc483b4a885f69317e3505"
	me.Value = mInstance
	me.Type = discovery.EVT_CREATE
	me.BusinessId = "f73dceb440f711eba63ffa163e7cdcb8"
	return me
}

func getMicroService() *discovery.MicroService {
	ms := discovery.MicroService{
		ServiceId: "1efe8be8eacce1efbf67967978133572fb8b5667",
		AppId: "default",
		ServiceName: "ProviderDemoService1-2",
		Version: "1.0.0",
		Level: "BACK",
		Status: "UP",
		Timestamp: "1608260891",
		ModTimestamp: "1608260891",
	}
	return &ms
}

type Fulldocument struct{
	Id             primitive.ObjectID `bson:"_id"`
	Domain         string
	Project        string
	Refreshtime    bson.Raw
	Tags           bson.Raw
	Serviceinfo    bson.Raw
	Instanceinfo   bson.Raw
}

type MongoWatchResponse struct {
	Id                  bson.Raw `bson:"_id"`
	OperationType       string
	Action              discovery.EventType
	ClusterTime         bson.Raw // nil
	MongoEvent          bson.Raw `bson:"fullDocument"`
	Ns                  bson.Raw
	DocumentKey         bson.Raw
}