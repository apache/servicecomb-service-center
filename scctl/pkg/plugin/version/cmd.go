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

package version

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/apache/servicecomb-service-center/scctl/pkg/version"
	"github.com/spf13/cobra"
)

var (
	RootCmd *cobra.Command
)

func init() {
	RootCmd = NewGetCommand(cmd.RootCmd())
}

func NewGetCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "Output the version of tool and service center",
		Run:     VersionCommandFunc,
	}
	parent.AddCommand(cmd)
	return cmd
}

func VersionCommandFunc(_ *cobra.Command, _ []string) {
	defer cmd.StopAndExit(cmd.ExitSuccess)
	fmt.Print(version.TOOL_NAME, " ")
	version.Ver().Print()

	scClient, err := sc.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		return
	}
	v, scErr := scClient.GetScVersion(context.Background())
	if scErr != nil {
		return
	}

	fmt.Println()
	fmt.Print("service center ")
	v.Print()
}
