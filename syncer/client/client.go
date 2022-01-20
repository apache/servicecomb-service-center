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

func initConfig() {
	mode := serverconfig.GetBool("ssl.enable", false)
	if mode {
		cfg = &grpc.TLSConfig{
			InsecureSkipVerify: true,
		}
	}
}

func RPClientConfig() *grpc.TLSConfig {
	once.Do(initConfig)
	return cfg
}
