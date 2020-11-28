package client

import (
	"bou.ke/monkey"
	"context"
	"crypto/tls"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)
var c = NewSyncClient("", new(tls.Config))

func TestClient_IncrementPull(t *testing.T)  {
	t.Run("Test IncrementPull", func(t *testing.T) {
		_, err := c.IncrementPull(context.Background(), "http://127.0.0.1")
		assert.Error(t, err, "IncrementPull fail without grpc")
	})
	t.Run("Test IncrementPull", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.PatchInstanceMethod(reflect.TypeOf((*Client)(nil)),
			"IncrementPull", func(client *Client, ctx context.Context, string2 string) (*pb.SyncData, error) {
				return syncDataCreate(), nil
			})

		syncData, err := c.IncrementPull(context.Background(), "http://127.0.0.1")
		assert.NoError(t, err, "IncrementPull no err when client exist")
		assert.NotNil(t, syncData, "syncData not nil when client exist")
	})
}

func TestClient_DeclareDataLength(t *testing.T) {
	t.Run("DeclareDataLength test", func(t *testing.T) {
		_, err := c.DeclareDataLength(context.Background(), "http://127.0.0.1")
		assert.Error(t, err, "DeclareDataLength fail without grpc")
	})
	t.Run("DeclareDataLength test", func(t *testing.T) {
		defer monkey.UnpatchAll()

		monkey.PatchInstanceMethod(reflect.TypeOf((*Client)(nil)),
			"DeclareDataLength", func(client *Client, ctx context.Context, string2 string) (*pb.DeclareResponse, error) {
				return declareRespCreate(), nil
			})

		declareResp, err := c.DeclareDataLength(context.Background(), "http://127.0.0.1")
		assert.NoError(t, err, "DeclareDataLength no err when client exist")
		assert.NotNil(t, declareResp, "DeclareDataLength not nil when client exist")
	})
}

func syncDataCreate() *pb.SyncData {
	syncService := pb.SyncService{
		ServiceId: "a59f99611a6945677a21f28c0aeb05abb",
	}
	services := []*pb.SyncService{&syncService}
	syncInstance := pb.SyncInstance{
		InstanceId: "5e1140fc232111eb9bb600acc8c56b5b",
	}
	instances := []*pb.SyncInstance{&syncInstance}
	syncData := pb.SyncData{
		Services: services,
		Instances: instances,
	}
	return &syncData
}

func declareRespCreate() *pb.DeclareResponse {
	declareResp := pb.DeclareResponse{
		SyncDataLength: 3,
	}
	return &declareResp
}
