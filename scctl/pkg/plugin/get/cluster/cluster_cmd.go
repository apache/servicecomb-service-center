// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/apache/servicecomb-service-center/scctl/pkg/plugin/get"
	"github.com/apache/servicecomb-service-center/scctl/pkg/writer"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func init() {
	NewClusterCommand(get.RootCmd)
}

func NewClusterCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster [options]",
		Short: "Output the registry clusters managed by service center",
		Run:   ClusterCommandFunc,
	}

	parent.AddCommand(cmd)
	return cmd
}

func ClusterCommandFunc(_ *cobra.Command, args []string) {
	scClient, err := sc.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}
	clusters, scErr := scClient.GetClusters(context.Background())
	if scErr != nil {
		cmd.StopAndExit(cmd.ExitError, scErr)
	}
	records := make(map[string]*ClusterRecord)
	for name, endpoints := range clusters {
		records[name] = &ClusterRecord{
			Name: name, Endpoints: endpoints,
		}
	}
	sp := &ClustersPrinter{Records: records}
	sp.SetOutputFormat(get.Output, get.AllDomains)
	writer.PrintTable(sp)
}
