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

package diagnose

import (
	"bytes"
	"context"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/client/sc"
	"github.com/apache/servicecomb-service-center/pkg/model"
	"github.com/apache/servicecomb-service-center/scctl/etcd"
	"github.com/apache/servicecomb-service-center/scctl/pkg/cmd"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/spf13/cobra"
)

const (
	service  = "service"
	instance = "instance"
)

const (
	greater = iota
	mismatch
	less
)

var typeMap = map[string]string{
	service:  "/cse-sr/ms/files/",
	instance: "/cse-sr/inst/files/",
}

type etcdResponse map[string][]*mvccpb.KeyValue

func DiagnoseCommandFunc(_ *cobra.Command, args []string) {
	// initialize sc/etcd clients
	scClient, err := sc.NewSCClient(cmd.ScClientConfig)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}
	etcdClient, err := etcd.NewEtcdClient(EtcdClientConfig)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}
	defer etcdClient.Close()

	// query etcd
	etcdResp, err := getEtcdResponse(context.Background(), etcdClient)
	if err != nil {
		cmd.StopAndExit(cmd.ExitError, err)
	}

	// query sc
	cache, scErr := scClient.GetScCache(context.Background())
	if scErr != nil {
		cmd.StopAndExit(cmd.ExitError, scErr)
	}

	// diagnose go...
	err, details := diagnose(cache, etcdResp)
	if err != nil {
		fmt.Println(details)                // stdout
		cmd.StopAndExit(cmd.ExitError, err) // stderr
	}
}

func getEtcdResponse(ctx context.Context, etcdClient *clientv3.Client) (etcdResponse, error) {
	etcdResp := make(etcdResponse)
	for t, prefix := range typeMap {
		if err := setResponse(ctx, etcdClient, t, prefix, etcdResp); err != nil {
			return nil, err
		}
	}
	return etcdResp, nil
}

func setResponse(ctx context.Context, etcdClient *clientv3.Client, key, prefix string, etcdResp etcdResponse) error {
	resp, err := etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	etcdResp[key] = resp.Kvs
	return nil
}

func diagnose(cache *model.Cache, etcdResp etcdResponse) (err error, details string) {
	var (
		service  = ServiceCompareHolder{Cache: cache.Microservices, Kvs: etcdResp[service]}
		instance = InstanceCompareHolder{Cache: cache.Instances, Kvs: etcdResp[instance]}
	)

	sr := service.Compare()
	ir := instance.Compare()

	var (
		b    bytes.Buffer
		full bytes.Buffer
	)
	writeResult(&b, &full, sr, ir)
	if b.Len() > 0 {
		return fmt.Errorf("error: %s", b.String()), full.String()
	}
	return nil, ""
}

func writeResult(b *bytes.Buffer, full *bytes.Buffer, rss ...*CompareResult) {
	g, m, l := make(map[string][]string), make(map[string][]string), make(map[string][]string)
	for _, rs := range rss {
		for t, arr := range rs.Results {
			switch t {
			case greater:
				g[rs.Name] = arr
			case mismatch:
				m[rs.Name] = arr
			case less:
				l[rs.Name] = arr
			}
		}
	}

	i := 0
	if s := len(g); s > 0 {
		i++
		header := fmt.Sprintf("%d. found in cache but not in etcd ", i)
		b.WriteString(header)
		full.WriteString(header)
		writeBody(full, g)
	}
	if s := len(m); s > 0 {
		i++
		header := fmt.Sprintf("%d. found different between cache and etcd ", i)
		b.WriteString(header)
		full.WriteString(header)
		writeBody(full, m)
	}
	if s := len(l); s > 0 {
		i++
		header := fmt.Sprintf("%d. found in etcd but not in cache ", i)
		b.WriteString(header)
		full.WriteString(header)
		writeBody(full, l)
	}
	if l := b.Len(); l > 0 {
		b.Truncate(l - 1)
		full.Truncate(full.Len() - 1)
	}
}

func writeBody(b *bytes.Buffer, r map[string][]string) {
	b.WriteString("\b, details:\n")
	for t, v := range r {
		writeSection(b, t)
		b.WriteString(fmt.Sprint(v))
		b.WriteRune('\n')
	}
}

func writeSection(b *bytes.Buffer, t string) {
	b.WriteString("  ")
	b.WriteString(t)
	b.WriteString(": ")
}
