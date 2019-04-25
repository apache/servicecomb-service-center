package config

import (
	"fmt"
	"os"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/peer"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins/repository/servicecenter"
	"github.com/apache/servicecomb-service-center/syncer/plugins/storage/memory"
)

var (
	DefaultMode           = "service-center"
	DefaultDCPort         = 30100
	DefaultTickerInterval = 30
)

type Config struct {
	*peer.Config
	LogFile           string `yaml:"log_file"`
	Mode              string `yaml:"mode"`
	DCAddr            string `yaml:"dc_addr"`
	JoinAddr          string `yaml:"join_addr"`
	TickerInterval    int    `yaml:"ticker_interval"`
	Profile           string `yaml:"profile"`
	EnableCompression bool   `yaml:"enable_compression"`
	AutoSync          bool   `yaml:"auto_sync"`
	StoragePlugin     string `json:"storage_plugin"`
	RepositoryPlugin  string `json:"repository_plugin"`
}

func DefaultConfig() *Config {
	peerConf := peer.DefaultConfig()
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf(err, "Error determining hostname: %s", err)
		return nil
	}
	peerConf.NodeName = hostname
	return &Config{
		LogFile:          "./syncer.log",
		Mode:             DefaultMode,
		DCAddr:           fmt.Sprintf("127.0.0.1:%d", DefaultDCPort),
		TickerInterval:   DefaultTickerInterval,
		Config:           peerConf,
		StoragePlugin:    memory.PluginName,
		RepositoryPlugin: servicecenter.PluginName,
	}
}

func (c *Config) Verification() error {
	ip, port, err := utils.SplitHostPort(c.BindAddr, peer.DefaultBindPort)
	if err != nil {
		return err
	}
	if ip == "127.0.0.1" {
		c.BindAddr = fmt.Sprintf("0.0.0.0:%d", port)
	}

	ip, port, err = utils.SplitHostPort(c.RPCAddr, peer.DefaultRPCPort)
	if err != nil {
		return err
	}
	c.RPCPort = port
	if ip == "127.0.0.1" {
		c.RPCAddr = fmt.Sprintf("0.0.0.0:%d", c.RPCPort)
	}
	return nil
}
