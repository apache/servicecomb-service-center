package server

import (
	"testing"

	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/stretchr/testify/assert"
)

func TestConvert_convertHttpOptions(t *testing.T) {
	t.Run("Run convertHttpOptions when ListenerTlsEnable is true", func(t *testing.T) {
		conf := confCreate1(true)
		httpOptions := convertHttpOptions(&conf)
		assert.True(t, cap(httpOptions) > len(httpOptions), "size of httpOptions less then cap")
	})
	t.Run("Run convertHttpOptions when ListenerTlsEnable is false", func(t *testing.T) {
		conf := confCreate1(false)
		httpOptions := convertHttpOptions(&conf)
		assert.True(t, cap(httpOptions) == len(httpOptions), "size of httpOptions equals cap")
	})
}

func confCreate1(ListenerTlsEnable bool) config.Config {
	tlsMount := config.Mount{
		Enabled: false,
		Name:    "servicecenter",
	}
	tlsMount1 := config.Mount{
		Enabled: ListenerTlsEnable,
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
	return conf
}
