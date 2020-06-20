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

package health

import (
	"context"
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"github.com/spf13/cobra"
)

const (
	ExistInternal    = iota + cmd.ExitError
	ExistUnavailable // connection timeout or refuse
	ExistAbnormal    // abnormal
)

func init() {
	NewHealthCommand(cmd.RootCmd())
}

func NewHealthCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health [options]",
		Short: "Output the health check result of service center",
		Run:   HealthCommandFunc,
	}

	parent.AddCommand(cmd)
	return cmd
}

func HealthCommandFunc(_ *cobra.Command, args []string) {
	scClient, err := sc.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		cmd.StopAndExit(ExistInternal, err)
	}
	scErr := scClient.HealthCheck(context.Background())
	if scErr != nil {
		switch scErr.Code {
		case scerr.ErrInternal:
			cmd.StopAndExit(ExistUnavailable, scErr)
		default:
			cmd.StopAndExit(ExistAbnormal, scErr)
		}
	}
}
