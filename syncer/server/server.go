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

package server

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"syscall"

	"github.com/apache/servicecomb-service-center/pkg/dump"
	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rpc"
	"github.com/apache/servicecomb-service-center/syncer/client"
	"github.com/apache/servicecomb-service-center/syncer/config"
	"github.com/apache/servicecomb-service-center/syncer/etcd"
	"github.com/apache/servicecomb-service-center/syncer/grpc"
	"github.com/apache/servicecomb-service-center/syncer/pkg/syssig"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"github.com/apache/servicecomb-service-center/syncer/serf"
	"github.com/apache/servicecomb-service-center/syncer/servicecenter"
	"github.com/apache/servicecomb-service-center/syncer/task"
	ggrpc "google.golang.org/grpc"

	// import plugins
	_ "github.com/apache/servicecomb-service-center/syncer/plugins/eureka"
	_ "github.com/apache/servicecomb-service-center/syncer/plugins/servicecenter"

	// import task
	_ "github.com/apache/servicecomb-service-center/syncer/task/idle"
	_ "github.com/apache/servicecomb-service-center/syncer/task/ticker"
)

var ErrStopChan = errors.New("stopped syncer by stopCh")

type moduleServer interface {
	// Starts launches the module server, the returned is not guaranteed that the server is ready
	// The moduleServer.Ready() channel will be transmit a message when server completed
	Start(ctx context.Context)

	// Returns a channel that will be closed when the module server is ready
	Ready() <-chan struct{}

	// Returns a channel that will be closed when the module server is stopped
	Stopped() <-chan struct{}
}

// Server struct for syncer
type Server struct {
	ctx    context.Context
	cancel context.CancelFunc
	// Syncer configuration
	conf *config.Config

	// task for discovery
	task task.Tasker

	// Wrap the servicecenter
	servicecenter servicecenter.Servicecenter

	etcd *etcd.Server

	// Wraps the serf agent
	serf *serf.Server

	// Wraps the grpc server
	grpc *grpc.Server

	revisionMap map[string]record

	eventQueue []*dump.WatchInstanceChangedEvent

	mapLock sync.RWMutex

	queueLock sync.RWMutex

	mux sync.RWMutex

	triggered bool

	// The channel will be closed when receiving a system interrupt signal
	stopCh chan struct{}
}

type record struct {
	revision int64
	action   string
}

// NewServer new server with Config
func NewServer(conf *config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		ctx:         ctx,
		cancel:      cancel,
		conf:        conf,
		stopCh:      make(chan struct{}),
		revisionMap: make(map[string]record),
		eventQueue:  make([]*dump.WatchInstanceChangedEvent, 0),
		triggered:   true,
	}
}

// Run syncer Server
func (s *Server) Run(ctx context.Context) {
	var err error
	s.initPlugin()
	if err = s.initialization(); err != nil {
		return
	}

	// Start system signal listening, wait for user interrupt program
	gopool.Go(syssig.Run)

	err = s.startModuleServer(s.serf)
	if err != nil {
		return
	}

	err = s.startModuleServer(s.etcd)
	if err != nil {
		return
	}

	err = s.startModuleServer(s.grpc)
	if err != nil {
		return
	}

	s.servicecenter.SetStorageEngine(s.etcd.Storage())

	s.task.Handle(s.tickHandler)

	s.DataRemoveTickHandler()

	s.task.Run(ctx)

	go s.NewHTTPServer()

	err = s.watchInstance()
	if err != nil {
		log.Error("watch instance error:%s", err)
		return
	}

	log.Info("start service done")

	<-s.stopCh

	s.Stop()
}

// Stop Syncer Server
func (s *Server) Stop() {
	if s.serf != nil {
		//stop serf agent
		s.serf.Stop()
	}

	if s.grpc != nil {
		s.grpc.Stop()
	}

	if s.etcd != nil {
		s.etcd.Stop()
	}

	s.cancel()

	// Closes all goroutines in the pool
	gopool.CloseAndWait()
}

func (s *Server) startModuleServer(module moduleServer) (err error) {
	gopool.Go(module.Start)
	select {
	case <-module.Ready():
		return nil
	case <-module.Stopped():
	case <-s.stopCh:
	}
	s.Stop()
	return ErrStopChan
}

// initialization Initialize the starter of the syncer
func (s *Server) initialization() (err error) {
	err = syssig.AddSignalsHandler(func() {
		log.Info("close svr stop chan")
		close(s.stopCh)
	}, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	if err != nil {
		log.Error("listen system signal failed", err)
		return
	}

	rpc.RegisterService(func(svr *ggrpc.Server) {
		pb.RegisterSyncServer(svr, s)
	})

	s.serf = serf.NewServer(s.conf.Join.Address, convertSerfOptions(s.conf)...)
	s.serf.OnceEventHandler(serf.NewEventHandler(serf.MemberJoinFilter(), s.waitClusterMembers))

	s.etcd, err = etcd.NewServer(convertEtcdOptions(s.conf)...)
	if err != nil {
		log.Errorf(err, "Create etcd failed, %s", err)
		return
	}

	s.task, err = task.GenerateTasker(s.conf.Task.Kind, convertTaskOptions(s.conf)...)
	if err != nil {
		log.Errorf(err, "Create tasker failed, %s", err)
		return
	}

	s.servicecenter, err = servicecenter.NewServicecenter(convertSCConfigOption(s.conf)...)
	if err != nil {
		log.Error("create servicecenter failed", err)
		return
	}

	s.grpc, err = grpc.NewServer(convertGRPCOptions(s.conf)...)
	if err != nil {
		log.Error("create grpc failed", err)
		return
	}

	return nil
}

// initPlugin Initialize the plugin and load the external plugin according to the configuration
func (s *Server) initPlugin() {
	plugins.SetPluginConfig(plugins.PluginServicecenter.String(), s.conf.Registry.Plugin)
	plugins.LoadPlugins()
}

func (s *Server) waitClusterMembers(data ...[]byte) bool {
	if s.conf.Mode == config.ModeCluster {
		tags := map[string]string{tagKeyClusterName: s.conf.Cluster}
		if len(s.serf.MembersByTags(tags)) < groupExpect {
			return false
		}
		err := s.configureCluster()
		if err != nil {
			log.Error("configure cluster failed", err)
			s.Stop()
			return false
		}
	}
	s.serf.AddEventHandler(serf.NewEventHandler(serf.UserEventFilter(EventDiscovered), s.userEvent))
	s.serf.AddEventHandler(serf.NewEventHandler(serf.UserEventFilter(EventIncrementPulled), s.incrementUserEvent))
	s.serf.AddEventHandler(serf.NewEventHandler(serf.UserEventFilter(EventNotifyFullPulled), s.notifyUserEvent))
	return true
}

// configureCluster Configuring the cluster by serf group member information
func (s *Server) configureCluster() error {
	// get local member of serf
	self := s.serf.LocalMember()
	_, peerPort, _ := utils.SplitAddress(s.conf.Listener.PeerAddr)
	ops := []etcd.Option{etcd.WithPeerAddr(self.Addr.String() + ":" + strconv.Itoa(peerPort))}

	// group members from serf as initial cluster members
	tags := map[string]string{tagKeyClusterName: s.conf.Cluster}
	for _, member := range s.serf.MembersByTags(tags) {
		ops = append(ops, etcd.WithAddPeers(member.Name, member.Addr.String()+":"+member.Tags[tagKeyClusterPort]))
	}

	return s.etcd.AddOptions(ops...)
}
func (s *Server) watchInstance() error {
	cli := client.NewWatchClient(s.conf.Registry.Address)

	err := cli.WatchInstances(s.addToQueue)

	if err != nil {
		return err
	}

	cli.WatchInstanceHeartbeat(s.addToQueue)

	return nil
}

func (s *Server) addToQueue(event *dump.WatchInstanceChangedEvent) {
	mapping := s.servicecenter.GetSyncMapping()

	for _, m := range mapping {
		if instFromOtherSC(event.Instance, m) {
			log.Debugf("instance[curId:%s, originId:%s] is from another sc, no need to put to queue",
				m.CurInstanceID, m.OrgInstanceID)
			return
		}
	}

	s.queueLock.Lock()
	s.eventQueue = append(s.eventQueue, event)
	log.Debugf("success add instance event to queue:%s   len:%s", event, len(s.eventQueue))
	s.queueLock.Unlock()
}

func instFromOtherSC(instance *dump.Instance, m *pb.MappingEntry) bool {
	if instance.Value.InstanceId == m.CurInstanceID && m.OrgInstanceID != "" {
		return true
	}
	return false
}

func (s *Server) getRevision(addr string) int64 {
	s.mapLock.RLock()
	value, ok := s.revisionMap[addr]
	s.mapLock.RUnlock()
	if ok {
		return value.revision
	}
	return -1
}

func (s *Server) getAction(addr string) string {
	s.mapLock.RLock()
	value, ok := s.revisionMap[addr]
	s.mapLock.RUnlock()
	if ok {
		return value.action
	}
	return ""
}

func (s *Server) getSyncDataLength(addr string) (response *pb.DeclareResponse) {
	response = &pb.DeclareResponse{
		SyncDataLength: int64(len(s.GetIncrementQueue(addr))),
	}
	return response
}

func (s *Server) updateRevisionMap(addr string, incrementQueue []*dump.WatchInstanceChangedEvent) {
	if len(incrementQueue) == 0 {
		log.Info("incrementQueue is empty, no need to update RevisionMap")
		return
	}

	log.Debug(fmt.Sprintf("update RevisionMap, addr = %s", addr))
	s.mapLock.Lock()
	s.revisionMap[addr] = record{
		incrementQueue[len(incrementQueue)-1].Revision,
		incrementQueue[len(incrementQueue)-1].Action,
	}
	s.mapLock.Unlock()
}

func (s *Server) GetIncrementQueue(addr string) []*dump.WatchInstanceChangedEvent {
	revision := s.getRevision(addr)
	action := s.getAction(addr)

	s.queueLock.RLock()
	defer s.queueLock.RUnlock()

	length := len(s.eventQueue)
	if length == 0 {
		log.Info("eventQueue is empty")
		return nil
	}

	if revision == -1 {
		return s.eventQueue
	}

	index := 0
	for _, event := range s.eventQueue {
		if event.Revision == revision && event.Action == action {
			break
		}
		index++
	}

	if index == length-1 {
		log.Info("no incremental event in the queue")
		return nil
	}

	if index == length {
		log.Info(fmt.Sprintf("fail to find the event in the queue by RevisionMap, Revision = %d, Action = %s", revision, action))
		return s.eventQueue
	}
	return s.eventQueue[index+1:]
}

func (s *Server) GetIncrementData(ctx context.Context, incrementQueue []*dump.WatchInstanceChangedEvent) (data *pb.SyncData) {
	return s.EventQueueToSyncData(ctx, incrementQueue)
}
