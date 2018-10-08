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

package get

import (
	"github.com/apache/incubator-servicecomb-service-center/scctl/pkg/cmd"
	"github.com/spf13/cobra"
)

var (
	Domain     string
	Output     string
	AllDomains bool
	RootCmd    *cobra.Command
)

func init() {
	RootCmd = NewGetCommand(cmd.RootCmd())
}

func NewGetCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <command> [options]",
		Short: "Output the resources information of service center",
	}
	parent.AddCommand(cmd)
	cmd.PersistentFlags().StringVarP(&Domain, "domain", "d", "default", "print the information under the specified domain in service center")
	cmd.PersistentFlags().StringVarP(&Output, "output", "o", "", "output the complete microservice information(e.g., framework, endpoints)")
	cmd.PersistentFlags().BoolVar(&AllDomains, "all-domains", false, "print the information under all domains in service center")

	return cmd
}
