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
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	"os"
	"path/filepath"
)

type Config struct {
	SaveDir string
}

type SchemaWriter interface {
	Write([]*pb.Schema) error
}

func NewSchemaWriter(cfg Config) SchemaWriter {
	switch {
	case len(cfg.SaveDir) == 0:
		return &SchemaStdoutWriter{}
	default:
		return &SchemaFileWriter{cfg.SaveDir}
	}
}

type SchemaStdoutWriter struct {
}

func (w *SchemaStdoutWriter) Write(schemas []*pb.Schema) error {
	for _, schema := range schemas {
		_, err := os.Stdout.WriteString(schema.Schema)
		if err != nil {
			return err
		}
	}
	return nil
}

type SchemaFileWriter struct {
	Dir string
}

func (w *SchemaFileWriter) Write(schemas []*pb.Schema) error {
	err := os.MkdirAll(w.Dir, 0750)
	if err != nil {
		return err
	}
	for _, schemas := range schemas {
		file, err := os.OpenFile(filepath.Join(w.Dir, schemas.SchemaId+".yaml"),
			os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return err
		}
		_, err = file.WriteString(schemas.Schema)
		if err != nil {
			file.Close()
			return err
		}
		file.Close()
	}
	return nil
}
