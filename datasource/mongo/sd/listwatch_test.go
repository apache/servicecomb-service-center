package sd

import (
	"context"
	"github.com/apache/servicecomb-service-center/datasource/mongo/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/go-chassis/v2/storage"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"
)

func TestListWatchConfig_String(t *testing.T) {
	t.Run("TestListWatchConfig_String", func(t *testing.T) {
		config := ListWatchConfig{
			Timeout: 666,
		}
		ret := config.String()
		assert.Equal(t, "{timeout: 666ns}", ret)
	})
	t.Run("when time is nil", func(t *testing.T) {
		config := ListWatchConfig{}
		ret := config.String()
		assert.Equal(t, "{timeout: 0s}", ret)
	})
}

func TestInnerListWatch_List(t *testing.T) {
	t.Run("with out key",  func(t *testing.T) {
		config := storage.Options{
			URI: "mongodb://localhost:27017",
		}
		client.NewMongoClient(config)
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
		}
		cfg := ListWatchConfig{
			Timeout: time.Minute,
			Context: context.Background(),
		}
		mlwr, err := ilw.List(cfg)
		assert.Nil(t, mlwr)
		assert.NotNil(t, err)
	})

	t.Run("with key instance",  func(t *testing.T) {
		config := storage.Options{
			URI: "mongodb://localhost:27017",
		}
		client.NewMongoClient(config)
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: INSTANCE,
		}
		cfg := ListWatchConfig{
			Timeout: time.Minute,
			Context: context.Background(),
		}
		mlwr, err := ilw.List(cfg)
		assert.NotNil(t, mlwr)
		assert.Nil(t, err)
	})

	t.Run("with key service",  func(t *testing.T) {
		config := storage.Options{
			URI: "mongodb://localhost:27017",
		}
		client.NewMongoClient(config)
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: SERVICE,
		}
		cfg := ListWatchConfig{
			Timeout: time.Minute,
			Context: context.Background(),
		}
		mlwr, err := ilw.List(cfg)
		assert.NotNil(t, mlwr)
		assert.Nil(t, err)
	})
}

func TestInnerListWatch_ResumeToken(t *testing.T) {
	t.Run("get resume token test", func(t *testing.T) {
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: INSTANCE,
			resumeToken: bson.Raw("resumToken"),
		}
		res := ilw.ResumeToken()
		assert.NotNil(t, res)
		assert.Equal(t, bson.Raw("resumToken"), res)
	})
}

func TestInnerListWatch_Watch(t *testing.T) {
	t.Run("config watch", func(t *testing.T) {
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: INSTANCE,
			resumeToken: bson.Raw("resumToken"),
		}
		op := ListWatchConfig{
			Timeout: time.Minute,
			Context: context.Background(),
		}
		iw := ilw.Watch(op)
		assert.NotNil(t, iw)
		assert.NotNil(t, iw.EventBus())
	})
}

func TestInnerWatcher_Stop(t *testing.T) {

}

func TestInnerWatcher_DoWatch(t *testing.T) {
	t.Run("Dowatch test", func(t *testing.T) {
		config := storage.Options{
			URI: "mongodb://localhost:27017",
		}
		client.NewMongoClient(config)
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: INSTANCE,
			resumeToken: bson.Raw("resumToken"),
		}
		err := ilw.DoWatch(context.Background(), func(resp *MongoListWatchResponse) {
			log.Info("inner func run")
		})
		assert.NotNil(t, err, "length of token exceeds watch table instance failed")
	})

	t.Run("", func(t *testing.T) {
		config := storage.Options{
			URI: "mongodb://localhost:27017",
		}
		client.NewMongoClient(config)
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: SERVICE,
			resumeToken: bson.Raw("1701539700"),
		}
		err := ilw.DoWatch(context.Background(), func(resp *MongoListWatchResponse) {
			log.Info("inner func run")
		})
		assert.NotNil(t, err)
	})
	t.Run("Dowatch test without token", func(t *testing.T) {
		config := storage.Options{
			URI: "mongodb://localhost:27017",
		}
		client.NewMongoClient(config)
		ilw := innerListWatch{
			Client: client.GetMongoClient(),
			Key: INSTANCE,
		}
		err := ilw.DoWatch(context.Background(), func(resp *MongoListWatchResponse) {
			log.Info("inner func run")
		})
		assert.Nil(t, err, "token is nil, but watch table instance start")
	})
}