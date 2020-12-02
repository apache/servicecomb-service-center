/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/server"
	"github.com/spf13/cobra"
)

var (
	conf       = &config.Config{}
	configFile = ""
)

var syncerCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start a syncer daemon",
	Run:   runSyncer,
}

func init() {
	rootCmd.AddCommand(syncerCmd)

	syncerCmd.Flags().StringVar(&conf.Mode, "mode", conf.Mode,
		"run mode")

	syncerCmd.Flags().StringVar(&conf.Node, "node", conf.Node,
		"node name")

	syncerCmd.Flags().StringVar(&conf.Cluster, "cluster", conf.Cluster,
		"cluster name")

	syncerCmd.Flags().StringVar(&conf.Listener.BindAddr, "bind-addr", conf.Listener.BindAddr,
		"address used to network with other Syncers")

	syncerCmd.Flags().StringVar(&conf.Listener.RPCAddr, "rpc-addr", conf.Listener.RPCAddr,
		"port used to synchronize data with other Syncers")

	syncerCmd.Flags().StringVar(&conf.Listener.PeerAddr, "peer-addr", conf.Listener.PeerAddr,
		"port used to communicate with other cluster members")

	syncerCmd.Flags().StringVar(&conf.Join.Address, "join", "",
		"address to join the network by specifying at least one existing member")

	syncerCmd.Flags().StringVar(&conf.Registry.Address, "registry", conf.Registry.Address,
		"address to monitor the registry")

	syncerCmd.Flags().StringVar(&conf.Registry.Plugin, "plugin", conf.Registry.Plugin,
		"plugin name of servicecenter")

	syncerCmd.Flags().StringVar(&configFile, "config", "",
		"configuration from file")
}

// runSyncer Runs the Syncer service.
func runSyncer(cmd *cobra.Command, args []string) {
	if conf.Join.Address != "" {
		conf.Join.Enabled = true
	}

	defaultConfig := config.DefaultConfig()

	if configFile != "" {
		fromFile, err := config.LoadConfig(configFile)
		if err != nil {
			log.Errorf(err, "load config file failed")
			return
		}
		if fromFile != nil {
			*defaultConfig = config.Merge(*defaultConfig, *fromFile)
		}
	}

	*conf = config.Merge(*defaultConfig, *conf)
	err := config.Verify(conf)
	if err != nil {
		log.Errorf(err, "verify syncer config failed")
		return
	}

	server.NewServer(conf).Run(context.Background())
}
