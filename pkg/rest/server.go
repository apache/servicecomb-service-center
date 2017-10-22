//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package rest

import (
	"crypto/tls"
	"github.com/ServiceComb/service-center/pkg/grace"
	"github.com/ServiceComb/service-center/pkg/util"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	serverStateInit = iota
	serverStateRunning
	serverStateTerminating
	serverStateClosed
)

type ServerConfig struct {
	Addr              string
	Handler           http.Handler
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	IdleTimeout       time.Duration
	WriteTimeout      time.Duration
	KeepAliveTimeout  time.Duration
	GraceTimeout      time.Duration
	MaxHeaderBytes    int
	TLSConfig         *tls.Config
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
		IdleTimeout:       60 * time.Second,
		WriteTimeout:      60 * time.Second,
		KeepAliveTimeout:  1 * time.Minute,
		GraceTimeout:      3 * time.Second,
		MaxHeaderBytes:    16384,
	}
}

func NewServer(srvCfg *ServerConfig) *Server {
	if srvCfg == nil {
		srvCfg = DefaultServerConfig()
	}
	return &Server{
		Server: &http.Server{
			Addr:              srvCfg.Addr,
			Handler:           srvCfg.Handler,
			TLSConfig:         srvCfg.TLSConfig,
			ReadTimeout:       srvCfg.ReadTimeout,
			ReadHeaderTimeout: srvCfg.ReadHeaderTimeout,
			IdleTimeout:       srvCfg.IdleTimeout,
			WriteTimeout:      srvCfg.WriteTimeout,
			MaxHeaderBytes:    srvCfg.MaxHeaderBytes,
		},
		KeepaliveTimeout: srvCfg.KeepAliveTimeout,
		GraceTimeout:     srvCfg.GraceTimeout,
		state:            serverStateInit,
		Network:          "tcp",
	}
}

type Server struct {
	*http.Server

	Network          string
	KeepaliveTimeout time.Duration
	GraceTimeout     time.Duration

	registerListener net.Listener
	restListener     net.Listener
	innerListener    *restListener

	conns int64
	wg    sync.WaitGroup
	state uint8
}

func (srv *Server) Serve() (err error) {
	defer func() {
		srv.state = serverStateClosed
	}()
	defer util.RecoverAndReport()
	srv.state = serverStateRunning
	err = srv.Server.Serve(srv.restListener)
	util.Logger().Debugf("server serve failed(%s)", err)
	srv.wg.Wait()
	return
}

func (srv *Server) AcceptOne() {
	defer util.RecoverAndReport()
	srv.wg.Add(1)
	atomic.AddInt64(&srv.conns, 1)
}

func (srv *Server) CloseOne() bool {
	defer util.RecoverAndReport()
	for {
		left := atomic.LoadInt64(&srv.conns)
		if left <= 0 {
			return false
		}
		if atomic.CompareAndSwapInt64(&srv.conns, left, left-1) {
			srv.wg.Done()
			return true
		}
	}
}

func (srv *Server) Listen() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}

	l, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	srv.restListener = newRestListener(l, srv)
	grace.RegisterFiles(addr, srv.File())
	return nil
}

func (srv *Server) ListenTLS() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":https"
	}

	l, err := srv.getListener(addr)
	if err != nil {
		return err
	}

	srv.innerListener = newRestListener(l, srv)
	srv.restListener = tls.NewListener(srv.innerListener, srv.TLSConfig)
	grace.RegisterFiles(addr, srv.File())
	return nil
}

func (srv *Server) ListenAndServe() (err error) {
	err = srv.Listen()
	if err != nil {
		return
	}
	return srv.Serve()
}

func (srv *Server) ListenAndServeTLS(certFile, keyFile string) (err error) {
	if srv.TLSConfig == nil {
		srv.TLSConfig = &tls.Config{}
		srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
		srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return
		}
	}
	if srv.TLSConfig.NextProtos == nil {
		srv.TLSConfig.NextProtos = []string{"http/1.1"}
	}

	err = srv.ListenTLS()
	if err != nil {
		return
	}
	return srv.Serve()
}

func (srv *Server) RegisterListener(l net.Listener) {
	srv.registerListener = l
}

func (srv *Server) getListener(addr string) (l net.Listener, err error) {
	l = srv.registerListener
	if l != nil {
		return
	}
	if !grace.IsFork() {
		return net.Listen(srv.Network, addr)
	}

	offset := grace.ExtraFileOrder(addr)
	if offset < 0 {
		return net.Listen(srv.Network, addr)
	}

	f := os.NewFile(uintptr(3+offset), "")
	l, err = net.FileListener(f)
	if err != nil {
		f.Close()
		return nil, err
	}
	return
}

func (srv *Server) Shutdown() {
	if srv.state != serverStateRunning {
		return
	}

	srv.state = serverStateTerminating
	err := srv.restListener.Close()
	if err != nil {
		util.Logger().Errorf(err, "server listener close failed")
	}

	if srv.GraceTimeout >= 0 {
		srv.gracefulStop(srv.GraceTimeout)
	}
}

func (srv *Server) gracefulStop(d time.Duration) {
	if srv.state != serverStateTerminating {
		return
	}

	<-time.After(d)

	n := 0
	for {
		if srv.state == serverStateClosed {
			break
		}

		if srv.CloseOne() {
			n++
			continue
		}
		break
	}

	if n != 0 {
		util.Logger().Warnf(nil, "%s timed out, force close %d connection(s)", d, n)
		err := srv.Server.Close()
		if err != nil {
			util.Logger().Warnf(err, "server close failed")
		}
	}
}

func (srv *Server) File() *os.File {
	switch srv.restListener.(type) {
	case *restListener:
		return srv.restListener.(*restListener).File()
	default:
		return srv.innerListener.File()
	}
}
