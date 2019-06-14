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
	conf = config.DefaultConfig()
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

	syncerCmd.Flags().StringVar(&conf.NodeName, "node", conf.NodeName,
		"node name")

	syncerCmd.Flags().StringVar(&conf.BindAddr, "bind-addr", conf.BindAddr,
		"address to bind listeners to")

	syncerCmd.Flags().StringVar(&conf.RPCAddr, "rpc-addr", conf.RPCAddr,
		"port to bind RPC listener to")

	syncerCmd.Flags().StringVar(&conf.JoinAddr, "join-addr", conf.JoinAddr,
		"address to join the cluster by specifying at least one existing member")

	syncerCmd.Flags().StringVar(&conf.SCAddr, "sc-addr", conf.SCAddr,
		"address to monitor the service-center")

	syncerCmd.Flags().StringVar(&conf.ClusterName, "cluster-name", conf.ClusterName,
		"name to group members into cluster")

	syncerCmd.Flags().IntVar(&conf.ClusterPort, "cluster-port", conf.ClusterPort,
		"port to communicate between cluster members")
}

// runSyncer Runs the Syncer service.
func runSyncer(cmd *cobra.Command, args []string) {
	if err := conf.Verification(); err != nil {
		log.Errorf(err, "verification syncer config failed")
		return
	}

	server.NewServer(conf).Run(context.Background())
}
