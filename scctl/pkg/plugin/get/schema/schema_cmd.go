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

package schema

import (
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	model2 "github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/apache/servicecomb-service-center/scctl/pkg/model"
	"github.com/apache/servicecomb-service-center/scctl/pkg/plugin/get"
	"github.com/apache/servicecomb-service-center/scctl/pkg/progress-bar"
	"github.com/apache/servicecomb-service-center/server/core"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	AppId       string
	ServiceName string
	Version     string
	SaveDir     string
)

func init() {
	NewSchemaCommand(get.RootCmd)
}

func NewSchemaCommand(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema [options]",
		Short: "Output the microservice schema information of the service center ",
		Run:   SchemaCommandFunc,
	}

	cmd.Flags().StringVarP(&SaveDir, "save-dir", "s", "",
		"the directory to save the schemas data")
	cmd.Flags().StringVar(&AppId, "app", "", "the application name of microservice")
	cmd.Flags().StringVar(&ServiceName, "name", "", "the name of microservice")
	cmd.Flags().StringVar(&Version, "version", "", "the semantic version of microservice")

	parent.AddCommand(cmd)
	return cmd
}

// schemas/[${domain}/][${project}/][${env}/]${app}/${microservice}.${version}/${schemaId}.yaml
func saveDirectory(root string, ms *model2.Microservice) string {
	if len(root) == 0 {
		return ""
	}
	domain, project := core.FromDomainProject(model.GetDomainProject(ms))
	if domain == core.RegistryDomain {
		domain = ""
	}
	if project == core.RegistryDomain {
		project = ""
	}
	return filepath.Join(root, "schemas", domain, project, ms.Value.Environment, ms.Value.AppId, ms.Value.ServiceName+".v"+ms.Value.Version)
}

func SchemaCommandFunc(_ *cobra.Command, args []string) {
	scClient, err := sc.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}
	cache, scErr := scClient.GetScCache(context.Background())
	if scErr != nil {
		cmd.StopAndExit(cmd.ExitError, scErr)
	}

	var progressBarWriter io.Writer = os.Stdout
	if len(SaveDir) == 0 {
		progressBarWriter = ioutil.Discard
	}
	progressBar := pb.NewProgressBar(len(cache.Microservices), progressBarWriter)
	defer progressBar.FinishPrint("Finished.")

	for _, ms := range cache.Microservices {
		progressBar.Increment()

		domainProject := model.GetDomainProject(ms)
		if !get.AllDomains && strings.Index(domainProject+core.SPLIT, get.Domain+core.SPLIT) != 0 {
			continue
		}
		if len(AppId) > 0 && ms.Value.AppId != AppId {
			continue
		}
		if len(ServiceName) > 0 && ms.Value.ServiceName != ServiceName {
			continue
		}
		if len(Version) > 0 && ms.Value.Version != Version {
			continue
		}

		schemas, err := scClient.GetSchemasByServiceID(context.Background(), domainProject, ms.Value.ServiceId)
		if err != nil {
			cmd.StopAndExit(cmd.ExitError, err)
		}
		if len(schemas) == 0 {
			continue
		}

		writer := NewSchemaWriter(Config{SaveDir: saveDirectory(SaveDir, ms)})
		if err := writer.Write(schemas); err != nil {
			fmt.Fprintln(os.Stderr, "output schema data failed", err.Error())
		}
	}
}
