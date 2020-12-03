package server

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/syncer/config"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/serf"
	"github.com/go-chassis/cari/discovery"
	"github.com/stretchr/testify/assert"
)

var s Server

func TestServer_IncrementPull(t *testing.T) {
	confCreate()
	s.channelMap = setChannel()

	t.Run("increment when address exist", func(t *testing.T) {
		iPReq := pb.IncrementPullRequest{
			Addr: "1",
		}
		syncData, err := s.IncrementPull(context.Background(), &iPReq)
		assert.NoError(t, err, "no error when DeclareDataLength")
		assert.NotEmpty(t, syncData, "no increase")
	})
}

func TestServer_DeclareDataLength(t *testing.T) {
	s.channelMap = setChannel()
	t.Run("when address exist", func(t *testing.T) {
		dReq := pb.DeclareRequest{
			Addr: "3",
		}
		declareResp, err := s.DeclareDataLength(context.Background(), &dReq)

		assert.NoError(t, err, "error when DeclareDataLength")
		assert.Empty(t, err, "declareResp.SyncDataLength is empty")
		assert.NotZero(t, declareResp.SyncDataLength, "declareResp.SyncDataLength is empty")

	})
}

func TestService_incrementUserEvent(t *testing.T) {

	t.Run("increment event fail", func(t *testing.T) {
		//membersCreate()

		svr := defaultServer()
		s.serf = svr

		confCreate()
		result := s.incrementUserEvent([]byte("servicecenter"))
		assert.Error(t, errors.New("members is nil"), "increment event fail when members is nil")
		assert.False(t, result, "increment event fail with cluster name servicecenter")
	})
}

func setChannel() map[string]chan *dump.WatchInstanceChangedEvent {
	var channelMap = make(map[string]chan *dump.WatchInstanceChangedEvent)
	var ch1 = make(chan *dump.WatchInstanceChangedEvent, 1000)
	var ch2 = make(chan *dump.WatchInstanceChangedEvent, 1000)
	var ch3 = make(chan *dump.WatchInstanceChangedEvent, 1000)
	channelMap["1"] = ch1
	channelMap["2"] = ch2
	channelMap["3"] = ch3

	var event1 = instanceAndServiceCreate(1)

	for _, ch := range channelMap {
		select {
		case ch <- event1:
		default:
		}
	}
	return channelMap
}

func confCreate() {
	tlsMount := config.Mount{
		Enabled: false,
		Name:    "servicecenter",
	}
	tlsMount1 := config.Mount{
		Enabled: false,
		Name:    "syncer",
	}
	listener := config.Listener{
		BindAddr:      "0.0.0.0:30190",
		AdvertiseAddr: "",
		RPCAddr:       "0.0.0.0:30191",
		PeerAddr:      "127.0.0.1:30192",
		TLSMount:      tlsMount1,
	}
	registry := config.Registry{
		Address:  "http://127.0.0.1:30100",
		Plugin:   "servicecenter",
		TLSMount: tlsMount,
	}
	join := config.Join{
		Enabled:       false,
		Address:       "127.0.0.1:30190",
		RetryMax:      3,
		RetryInterval: "30s",
	}
	lable := config.Label{
		Key:   "interval",
		Value: "30s",
	}
	lables := append([]config.Label{}, lable)
	task := config.Task{
		Kind:   "ticker",
		Params: lables,
	}
	ciphers := []string{"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_RSA_WITH_AES_256_GCM_SHA384", "TLS_RSA_WITH_AES_128_GCM_SHA256"}
	tlsConfig1 := config.TLSConfig{
		Name:       "syncer",
		VerifyPeer: true,
		MinVersion: "TLSv1.2",
		Passphrase: "",
		CAFile:     "./certs/trust.cer",
		CertFile:   "./certs/server.cer",
		KeyFile:    "./certs/server_key.pem",
		Ciphers:    ciphers,
	}
	tlsConfig2 := config.TLSConfig{
		Name:       "servicecenter",
		VerifyPeer: false,
		CAFile:     "./certs/trust.cer",
		CertFile:   "./certs/server.cer",
		KeyFile:    "./certs/server_key.pem",
	}
	tlsConfigs := []*config.TLSConfig{&tlsConfig1, &tlsConfig2}
	var conf = config.Config{
		Mode:       "signle",
		Node:       "syncer-node",
		Cluster:    "syncer-cluster",
		DataDir:    "./syncer-data/",
		Listener:   listener,
		Join:       join,
		Task:       task,
		Registry:   registry,
		TLSConfigs: tlsConfigs,
	}
	s.conf = &conf
}

func instanceAndServiceCreate(i int) *dump.WatchInstanceChangedEvent {
	var event = new(dump.WatchInstanceChangedEvent)
	status := []string{"UNKNOWN", "UP", "DOWN"}
	var ss = new(dump.Microservice)
	var sv = new(discovery.MicroService)
	sv.AppId = "serviceApp" + strconv.FormatInt(int64(i), 10)
	sv.Environment = "env"
	sv.ServiceId = "a59f99611a6945677a21f28c0aeb05abb" + strconv.FormatInt(int64(i/2), 10)
	sv.Status = status[i%3]
	sv.Version = "1.0.0"
	var sk = new(dump.KV)
	sk.Key = "/cse-sr/ms/files/default/default/" + sv.ServiceId
	sk.Rev = int64(i)

	ss.Value = sv
	ss.KV = sk

	is := new(dump.Instance)
	insStatus := []string{"UNKNOWN", "UP", "STARTING", "DOWN", "OUTOFSERVICE"}
	healthCheckModes := []string{"UNKNOWN", "PUSH", "PULL"}
	healthCheck := discovery.HealthCheck{
		Mode:     healthCheckModes[i%3],
		Interval: 30,
		Times:    30,
	}
	var iv = new(discovery.MicroServiceInstance)
	iv.HostName = "provider_demo" + strconv.FormatInt(int64(i), 10)
	iv.Endpoints = []string{"rest://127.0.0.1:8080"}
	iv.InstanceId = "5e1140fc232111eb9bb600acc8c56b5b" + strconv.FormatInt(int64(i/2), 10)
	iv.HealthCheck = &healthCheck
	if i == 10 {
		iv.ServiceId = "a59f99611a6945677a21f28c0aeb05abb" + strconv.FormatInt(int64(i), 10)
	} else {
		iv.ServiceId = "a59f99611a6945677a21f28c0aeb05abb" + strconv.FormatInt(int64(i/2), 10)
	}
	iv.Status = insStatus[i%5]
	iv.Version = "1.0.0"

	var ik = new(dump.KV)
	ik.Key = "/cse-sr/inst/files/default/default/" + sv.ServiceId + iv.InstanceId
	ik.Rev = int64(i)
	is.KV = ik
	is.Value = iv
	is.Rev = int64(i)

	event.Instance = is
	event.Service = ss

	return event
}

func defaultServer() *serf.Server {
	return serf.NewServer(
		"",
		serf.WithNode("syncer-test"),
		serf.WithBindAddr("127.0.0.1"),
		serf.WithBindPort(35151),
	)
}
