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

package serf

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"github.com/hashicorp/serf/serf"
	"github.com/pkg/errors"
)

// HandleFunc handle user event
type HandleFunc func(data ...[]byte) bool

//// QueryHandler handle query event
//type QueryHandler func(data []byte) []byte

// CallbackFunc callback handler for query event
type CallbackFunc func(from string, data []byte)

// Server serf server
type Server struct {
	conf    *serf.Config
	serf    *serf.Serf
	running *utils.AtomicBool

	eventCh chan serf.Event
	readyCh chan struct{}
	stopCh  chan struct{}

	handlerMap *sync.Map
	peerAddr   []string
}

// NewServer new serf server with options
func NewServer(peerAddr string, opts ...Option) *Server {
	conf := serf.DefaultConfig()
	conf.Tags = map[string]string{}
	for _, opt := range opts {
		opt(conf)
	}

	eventCh := make(chan serf.Event, 64)
	conf.EventCh = eventCh

	address := strings.Split(peerAddr, ",")
	if len(peerAddr) == 0 {
		address = []string{}
	}
	return &Server{
		conf:       conf,
		running:    utils.NewAtomicBool(false),
		eventCh:    eventCh,
		readyCh:    make(chan struct{}),
		stopCh:     make(chan struct{}),
		handlerMap: &sync.Map{},
		peerAddr:   address,
	}
}

// Start serf server
func (s *Server) Start(ctx context.Context) {
	s.running.DoToReverse(false, func() {
		sf, err := serf.Create(s.conf)
		if err != nil {
			log.Error("serf: start server failed", err)
			close(s.stopCh)
			return
		}
		s.serf = sf
		s.Join(s.peerAddr)
		close(s.readyCh)
		go s.waitEvent(ctx)
	})
}

// Stop serf server
func (s *Server) Stop() {
	s.running.DoToReverse(true, func() {
		if s.serf != nil {
			log.Info("serf: begin shutdown")
			if err := s.serf.Shutdown(); err != nil {
				log.Error("serf: shutdown failed", err)
			}
			close(s.stopCh)
		}

		log.Info("serf: shutdown complete")
	})
}

// Ready Returns a channel that will be closed when serf is ready
func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

// Stopped Returns a channel that will be closed when serf is stopped
func (s *Server) Stopped() <-chan struct{} {
	return s.stopCh
}

// OnceEventHandler Add serf event handler, the handler will be automatically deleted after executing it once
func (s *Server) OnceEventHandler(handler EventHandler) {
	onceHandler := onceEventHandler(handler)
	s.AddEventHandler(onceHandler)
	go s.waitOnceEventHandler(onceHandler)
}

// AddEventHandler Add serf event handler
func (s *Server) AddEventHandler(handler EventHandler) {
	_, ok := s.handlerMap.Load(handler)
	if ok {
		log.Warn("serf: event handle is already exits, " + handler.String())
	}
	s.handlerMap.Store(handler, struct{}{})
}

// RemoveEventHandler remove serf event handler
func (s *Server) RemoveEventHandler(handler EventHandler) {
	_, ok := s.handlerMap.Load(handler)
	if !ok {
		log.Warn("serf: event handle is notfound, " + handler.String())
		return
	}
	s.handlerMap.Delete(handler)
}

func (s *Server) waitOnceEventHandler(once *onceHandler) {
	<-once.Ready()
	s.RemoveEventHandler(once)
}

// UserEvent send user event
func (s *Server) UserEvent(name string, payload []byte) error {
	err := s.serf.UserEvent(name, payload, true)
	if err != nil {
		err = errors.Wrapf(err, "serf: send user event '%s' failed", name)
	}
	return err
}

// Query send query
func (s *Server) Query(name string, payload []byte, callback CallbackFunc, opts ...QueryOption) error {
	param := s.serf.DefaultQueryParams()
	for _, opt := range opts {
		opt(param)
	}

	resp, err := s.serf.Query(name, payload, param)
	if err != nil {
		err = errors.Wrapf(err, "serf: send query '%s' failed", name)
		return err
	}
	go s.responseCallback(resp, callback)
	return nil
}

// Join asks the Serf instance to join. See the Serf.Join function.
func (s *Server) Join(addrs []string) (n int, err error) {
	log.Infof("serf: join to: %v replay : %v", addrs)
	n, err = s.serf.Join(addrs, true)
	if n > 0 {
		log.Infof("serf: joined: %d nodes", n)
	}
	if err != nil {
		log.Warnf("serf: error joining: %v", err)
	}
	return
}

// MembersByTags Returns members matching the tags
func (s *Server) MembersByTags(tags map[string]string) (members []serf.Member) {
	if s.serf == nil {
		return
	}

next:
	for _, member := range s.serf.Members() {
		for key, val := range tags {
			if member.Tags[key] != val {
				continue next
			}
		}
		members = append(members, member)
	}
	return
}

// LocalMember returns the Member information for the local node
func (s *Server) LocalMember() *serf.Member {
	if s.serf != nil {
		member := s.serf.LocalMember()
		return &member
	}
	return nil
}

// Member get member information with node
func (s *Server) Member(node string) *serf.Member {
	if s.serf != nil {
		ms := s.serf.Members()
		for _, m := range ms {
			if m.Name == node {
				return &m
			}
		}
	}
	return nil
}

func (s *Server) responseCallback(resp *serf.QueryResponse, callback CallbackFunc) {
	hourglass := time.After(resp.Deadline().Sub(time.Now()))
	for {
		select {
		case a := <-resp.AckCh():
			log.Infof("query response ack: %s", a)
		case r := <-resp.ResponseCh():
			log.Infof("query response: from %s, content %s", r.From, string(r.Payload))
			callback(r.From, r.Payload)
		case <-hourglass:
			log.Info("query response timeout")
			return
		}
	}
}

func (s *Server) waitEvent(ctx context.Context) {
	for {
		select {
		case e := <-s.eventCh:
			s.handlerMap.Range(func(key, value interface{}) bool {
				if handler, ok := key.(EventHandler); ok {
					handler.Handle(e)
				}
				return true
			})
		case <-s.serf.ShutdownCh():
			log.Warn("serf: server stopped, exited")
			s.Stop()
			return
		case <-ctx.Done():
			log.Warn("serf: cancel server by context")
			s.Stop()
			return
		}
	}
}
