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

package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/apache/servicecomb-service-center/istio/pkg/bootstrap"

	"github.com/spf13/cobra"
	"istio.io/pkg/log"
)

var inputArgs *bootstrap.Args
var loggingOptions = log.DefaultOptions()

// NewRootCommand creates servicecomb-service-center-istio service cli args
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "servicecenter-to-istio",
		Short: "sc2mesh",
		Long:  "sc2mesh synchronizes data from servicecomb service center to Istio",
		Args:  cobra.ExactArgs(0),
		PreRunE: func(c *cobra.Command, args []string) error {
			if err := log.Configure(loggingOptions); err != nil {
				return err
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			// Create the stop channel for all of the servers.
			// ctx, cancelFunc := context.WithCancel(context.Background())
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()
			// defer cancelFunc()
			// Create the server for the servicecomb-service-center-istio service.
			server, err := bootstrap.NewServer(inputArgs)
			if err != nil {
				return fmt.Errorf("failed to create servicecomb-service-center-istio service: %v", err)
			}

			// Start the server
			if err := server.Start(ctx, inputArgs); err != nil {
				return fmt.Errorf("failed to start servicecomb-service-center-istio service: %v", err)
			}

			waitSignal(ctx)

			return nil
		},
	}
	addFlags(rootCmd)

	return rootCmd
}

// WaitSignal awaits for SIGINT or SIGTERM and closes the channel
func waitSignal(ctx context.Context) {
	<-ctx.Done()
	_ = log.Sync()
}

func addFlags(c *cobra.Command) {
	inputArgs = &bootstrap.Args{}

	// Process commandline args.

	// sc-addr is the service center registry centre address
	c.PersistentFlags().StringVar(&inputArgs.ServiceCenterAddr, "sc-addr", "localhost:30100",
		"servicecomb service center host ip address")
	// enable leader-election or not
	c.PersistentFlags().BoolVar(&inputArgs.HA, "ha", false,
		"enable k8s leader election or not for high avalibility")
	// kubectl config file path, if not set, will use in cluster kube config
	c.PersistentFlags().StringVar(&inputArgs.Kubeconfig, "kube-config", "", "service discovery kube config file")
	// Attach the Istio logging options to the command.
	loggingOptions.AttachCobraFlags(c)
}
