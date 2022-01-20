package client

import (
	"sync"

	grpc "github.com/apache/servicecomb-service-center/pkg/rpc"
	serverconfig "github.com/apache/servicecomb-service-center/server/config"
)

var (
	cfg  *grpc.TLSConfig
	once sync.Once
)

const (
	tlsEnable = "1"
)

func initConfig() {
	mode := serverconfig.GetString("ssl.mode", "0", serverconfig.WithStandby("ssl_mode"))
	if mode == tlsEnable {
		cfg = &grpc.TLSConfig{
			InsecureSkipVerify: true,
		}
	}
}

func RPClientConfig() *grpc.TLSConfig {
	once.Do(initConfig)
	return cfg
}
