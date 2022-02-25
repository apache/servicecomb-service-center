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

package replicator

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	"github.com/apache/servicecomb-service-center/pkg/util"
	v1sync "github.com/apache/servicecomb-service-center/syncer/api/v1"
	syncerclient "github.com/apache/servicecomb-service-center/syncer/client"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/service/replicator/resource"
	"github.com/go-chassis/foundation/gopool"
	"google.golang.org/grpc"
)

const (
	schema      = "grpc"
	serviceName = "syncer"
)

const (
	reservedSize = 512 * 1024
	maxSize      = 10*1024*1024 - reservedSize
)

var (
	manager = NewManager(make(map[string]struct{}, 1000))
)

var (
	conn *grpc.ClientConn
)

func Work() error {
	err := InitSyncClient()
	if err != nil {
		return err
	}

	gopool.Go(func(ctx context.Context) {
		<-ctx.Done()

		Close()
	})

	resource.InitManager()
	return err
}

func InitSyncClient() error {
	cfg := config.GetConfig()
	peer := cfg.Sync.Peers[0]
	log.Info(fmt.Sprintf("peer is %v", peer))
	var err error
	conn, err = rpc.GetRoundRobinLbConn(&rpc.Config{
		Addrs:       peer.Endpoints,
		Scheme:      schema,
		ServiceName: serviceName,
		TLSConfig:   syncerclient.RPClientConfig(),
	})
	return err
}

func Close() {
	if conn == nil {
		return
	}

	err := conn.Close()
	if err != nil {
		log.Error("close conn failed", err)
	}
}

func Manager() Replicator {
	return manager
}

// Replicator define replicator manager, receive events from event manager
// and send events to remote syncer
type Replicator interface {
	Replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error)
	Persist(ctx context.Context, el *v1sync.EventList) []*resource.Result
}

func NewManager(cache map[string]struct{}) Replicator {
	return &replicatorManager{
		cache: cache,
	}
}

type replicatorManager struct {
	cache map[string]struct{}
}

func (r *replicatorManager) Replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error) {
	return r.replicate(ctx, el)
}

func pageEvents(source *v1sync.EventList, max int) []*v1sync.EventList {
	els := make([]*v1sync.EventList, 0, 5)

	size := 0
	el := &v1sync.EventList{
		Events: make([]*v1sync.Event, 0, 20),
	}

	for _, event := range source.Events {
		lv := len(event.Value)
		if size+lv < max {
			el.Events = append(el.Events, event)
			size += lv
			continue
		}

		log.Info(fmt.Sprintf("size is %d", size))
		els = append(els, el)

		el = &v1sync.EventList{
			Events: make([]*v1sync.Event, 0, 20),
		}
		el.Events = append(el.Events, event)
		size = lv
	}

	log.Info(fmt.Sprintf("size is %d", size))
	els = append(els, el)

	return els
}

func (r *replicatorManager) replicate(ctx context.Context, el *v1sync.EventList) (*v1sync.Results, error) {
	log.Info(fmt.Sprintf("start replicate events %d", len(el.Events)))

	set := client.NewSet(conn)

	els := pageEvents(el, maxSize)

	result := &v1sync.Results{
		Results: make(map[string]*v1sync.Result, len(el.Events)),
	}

	log.Info(fmt.Sprintf("page count %d to sync", len(els)))

	for _, in := range els {
		res, err := set.EventServiceClient.Sync(ctx, in)
		if err != nil {
			return nil, err
		}

		log.Info(fmt.Sprintf("replicate events success, count is %d", len(in.Events)))

		for k, v := range res.Results {
			log.Info(fmt.Sprintf("replicate event %s, %v", k, v))
			result.Results[k] = v
		}
	}

	log.Info(fmt.Sprintf("replicate events success %d", len(result.Results)))
	return result, nil
}

func (r *replicatorManager) Persist(ctx context.Context, el *v1sync.EventList) []*resource.Result {
	if el == nil || len(el.Events) == 0 {
		return []*resource.Result{}
	}

	results := make([]*resource.Result, 0, len(el.Events))
	for _, event := range el.Events {
		log.Info(fmt.Sprintf("start handle event %s", event.Flag()))

		r, result := resource.New(event)
		if result != nil {
			results = append(results, result.WithEventID(event.Id))
			continue
		}

		ctx = util.SetDomain(ctx, event.Opts[string(util.CtxDomain)])
		ctx = util.SetProject(ctx, event.Opts[string(util.CtxProject)])

		result = r.LoadCurrentResource(ctx)
		if result != nil {
			results = append(results, result.WithEventID(event.Id))
			continue
		}

		result = r.NeedOperate(ctx)
		if result != nil {
			results = append(results, result.WithEventID(event.Id))
			continue
		}

		result = r.Operate(ctx)
		results = append(results, result.WithEventID(event.Id))

		log.Info(fmt.Sprintf("operate resource %s, %s", event.Flag(), result.Flag()))
	}

	for _, result := range results {
		log.Info(fmt.Sprintf("handle event result %s", result.Flag()))
	}

	return results
}
