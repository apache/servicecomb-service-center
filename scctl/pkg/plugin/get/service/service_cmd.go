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

package service

import (
	"context"
	"strings"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/datasource/etcd/path"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/apache/servicecomb-service-center/scctl/pkg/model"
	"github.com/apache/servicecomb-service-center/scctl/pkg/plugin/get"
	"github.com/apache/servicecomb-service-center/scctl/pkg/writer"
	"github.com/spf13/cobra"
)

func init() {
	NewServiceCommand(get.RootCmd)
}

func NewServiceCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service [options]",
		Aliases: []string{"svc"},
		Short:   "Output the microservice information of the service center ",
		Run:     CommandFunc,
	}

	parent.AddCommand(cmd)
	return cmd
}

func CommandFunc(_ *cobra.Command, _ []string) {
	scClient, err := client.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}
	cache, scErr := scClient.GetScCache(context.Background())
	if scErr != nil {
		cmd.StopAndExit(cmd.ExitError, scErr)
	}

	endpointMap := make(map[string][]string)
	for i := 0; i < len(cache.Instances); i++ {
		serviceID := cache.Instances[i].Value.ServiceId
		endpointMap[serviceID] = append(endpointMap[serviceID], cache.Instances[i].Value.Endpoints...)
	}

	records := make(map[string]*Record)
	for _, ms := range cache.Microservices {
		domainProject := model.GetDomainProject(ms)
		if !get.AllDomains && strings.Index(domainProject+path.SPLIT, get.Domain+path.SPLIT) != 0 {
			continue
		}

		appID := ms.Value.AppId
		env := ms.Value.Environment
		name := ms.Value.ServiceName

		key := util.StringJoin([]string{domainProject, env, appID, name}, "/")
		svc, ok := records[key]
		if !ok {
			svc = &Record{
				Service: model.Service{
					DomainProject: domainProject,
					Environment:   env,
					AppID:         appID,
					ServiceName:   name,
				},
			}
			records[key] = svc
		}
		svc.AppendVersion(ms.Value.Version)
		svc.AppendFramework(ms.Value.Framework)
		svc.AppendEndpoints(endpointMap[ms.Value.ServiceId])
		svc.UpdateTimestamp(ms.Value.Timestamp)
	}

	sp := &Printer{Records: records}
	sp.SetOutputFormat(get.Output, get.AllDomains)
	writer.PrintTable(sp)
}
