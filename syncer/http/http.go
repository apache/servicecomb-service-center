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

package http

import (
	"compress/gzip"
	"context"
	"github.com/NYTimes/gziphandler"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/utils"
	"io"
	"net/http"
	"sync"
)

const (
	Path    = "/v1/syncer/notify"
	Message = "Deliver full synchronization task success!"
)

// Server struct
type Server struct {
	server    *http.Server
	running   *utils.AtomicBool
	Triggered bool
	Mux       sync.RWMutex

	readyCh chan struct{}
	stopCh  chan struct{}
}

// NewServer new http server
func NewServer(ops ...Option) (*Server, error) {
	conf := toHttpServerConfig(ops...)

	srv := &http.Server{
		Addr:      conf.addr,
		TLSConfig: conf.tlsConfig,
	}
	if conf.compressed && conf.compressMinBytes > 0 {
		wrapper, _ := gziphandler.NewGzipLevelAndMinSize(gzip.DefaultCompression, conf.compressMinBytes)
		srv.Handler = wrapper(http.NewServeMux())
	}

	return &Server{
		server:    srv,
		running:   utils.NewAtomicBool(false),
		Triggered: false,
		readyCh:   make(chan struct{}),
		stopCh:    make(chan struct{}),
	}, nil
}

func (s *Server) Start(ctx context.Context) {
	s.running.DoToReverse(false, func() {
		go func() {
			http.HandleFunc(Path, func(w http.ResponseWriter, r *http.Request) {
				s.Mux.Lock()
				s.Triggered = true
				s.Mux.Unlock()
				_, err := io.WriteString(w, Message)
				if err != nil {
					log.Error("", err)
				}
			})
			err := s.server.ListenAndServe()
			if err != nil {
				log.Error("httpserver: start server failed", err)
				s.Stop()
			}
		}()
		close(s.readyCh)
		go s.wait(ctx)
	})
}

// Stop http server
func (s *Server) Stop() {
	s.running.DoToReverse(true, func() {
		if s.server != nil {
			log.Info("httpserver: begin shutdown")
			err := s.server.Close()
			if err != nil {
				log.Errorf(err, "httpserver close failed")
			}
			close(s.stopCh)
		}
		log.Info("httpserver: shutdown complete")
	})
	return
}

func (s *Server) wait(ctx context.Context) {
	select {
	case <-s.stopCh:
		log.Warn("httpserver: server stopped, exited")
	case <-ctx.Done():
		log.Warn("httpserver: cancel server by context")
		s.Stop()
	}
}

// Ready Returns a channel that will be closed when httpserver is ready
func (s *Server) Ready() <-chan struct{} {
	return s.readyCh
}

// Stopped Returns a channel that will be closed when httpserver is stopped
func (s *Server) Stopped() <-chan struct{} {
	return s.stopCh
}
