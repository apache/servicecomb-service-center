package peer

import (
	"fmt"

	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/serf/cmd/serf/command/agent"
	"github.com/hashicorp/serf/serf"
)

const (
	DefaultBindPort = 30190
	DefaultRPCPort  = 30191
)

func DefaultConfig() *Config {
	agentConf := agent.DefaultConfig()
	agentConf.BindAddr = fmt.Sprintf("0.0.0.0:%d", DefaultBindPort)
	agentConf.RPCAddr = fmt.Sprintf("0.0.0.0:%d", DefaultRPCPort)
	return &Config{Config: agentConf}
}

type Config struct {
	*agent.Config
	RPCPort int `yaml:"-"`
}

func (c *Config) readConfigFile(filepath string) error {
	if filepath != "" {

	}
	return nil
}

func (c *Config) convertToSerf() (*serf.Config, error) {
	serfConf := serf.DefaultConfig()

	bindIP, bindPort, err := utils.SplitHostPort(c.BindAddr, DefaultBindPort)
	if err != nil {
		return nil, fmt.Errorf("invalid bind address: %s", err)
	}

	switch c.Profile {
	case "lan":
		serfConf.MemberlistConfig = memberlist.DefaultLANConfig()
	case "wan":
		serfConf.MemberlistConfig = memberlist.DefaultWANConfig()
	case "local":
		serfConf.MemberlistConfig = memberlist.DefaultLocalConfig()
	default:
		serfConf.MemberlistConfig = memberlist.DefaultLANConfig()
	}

	serfConf.MemberlistConfig.BindAddr = bindIP
	serfConf.MemberlistConfig.BindPort = bindPort
	serfConf.NodeName = c.NodeName
	return serfConf, nil
}
