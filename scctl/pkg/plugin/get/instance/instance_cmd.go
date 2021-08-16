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

package instance

import (
	"context"
	"strings"

	"github.com/apache/servicecomb-service-center/client"
	model2 "github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/apache/servicecomb-service-center/scctl/pkg/model"
	"github.com/apache/servicecomb-service-center/scctl/pkg/plugin/get"
	"github.com/apache/servicecomb-service-center/scctl/pkg/writer"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/spf13/cobra"
)

func init() {
	NewInstanceCommand(get.RootCmd)
}

func NewInstanceCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instance [options]",
		Aliases: []string{"inst"},
		Short:   "Output the instance information of the service center ",
		Run:     InstanceCommandFunc,
	}
	parent.AddCommand(cmd)
	return cmd
}

func InstanceCommandFunc(_ *cobra.Command, args []string) {
	scClient, err := client.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}
	cache, scErr := scClient.GetScCache(context.Background())
	if scErr != nil {
		cmd.StopAndExit(cmd.ExitError, scErr)
	}

	svcMap := make(map[string]*model2.Microservice)
	for _, ms := range cache.Microservices {
		svcMap[ms.Value.ServiceId] = ms
	}

	records := make(map[string]*InstanceRecord)
	for _, inst := range cache.Instances {
		domainProject := model.GetDomainProject(inst)
		if !get.AllDomains && strings.Index(domainProject+core.SPLIT, get.Domain+core.SPLIT) != 0 {
			continue
		}

		svc, ok := svcMap[inst.Value.ServiceId]
		if !ok {
			continue
		}

		instance, ok := records[inst.Value.InstanceId]
		if !ok {
			instance = &InstanceRecord{
				Instance: model.Instance{
					DomainProject: domainProject,
					Host:          inst.Value.HostName,
					Endpoints:     inst.Value.Endpoints,
					Environment:   svc.Value.Environment,
					AppId:         svc.Value.AppId,
					ServiceName:   svc.Value.ServiceName,
					Version:       svc.Value.Version,
					Framework:     svc.Value.Framework,
				},
			}
			records[inst.Value.InstanceId] = instance
		}
		instance.SetLease(inst.Value.HealthCheck)
		instance.UpdateTimestamp(inst.Value.Timestamp)
	}

	sp := &InstancePrinter{Records: records}
	sp.SetOutputFormat(get.Output, get.AllDomains)
	writer.PrintTable(sp)
}
