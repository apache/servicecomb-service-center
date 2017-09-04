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
	"github.com/ServiceComb/service-center/pkg/common"
	"github.com/ServiceComb/service-center/util"
	"github.com/ServiceComb/service-center/util/grace"
	"github.com/astaxie/beego"
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

var (
	defaultRESTfulServer *Server
)

type httpServerCfg struct {
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	KeepaliveTimeout  time.Duration
	GraceTimeout      time.Duration
	MaxHeaderBytes    int
}

func loadCfg() *httpServerCfg {
	readHeaderTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("read_header_timeout", "60s"))
	readTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("read_timeout", "60s"))
	writeTimeout, _ := time.ParseDuration(beego.AppConfig.DefaultString("write_timeout", "60s"))
	maxHeaderBytes := beego.AppConfig.DefaultInt("max_header_bytes", 16384)
	return &httpServerCfg{readTimeout, readHeaderTimeout, writeTimeout,
		3 * time.Minute, 3 * time.Second, maxHeaderBytes}
}

func NewServer(addr string, handler http.Handler) (server *Server, err error) {
	var tlsConfig *tls.Config
	if common.GetServerSSLConfig().SSLEnabled {
		verifyClient := common.GetServerSSLConfig().VerifyClient
		tlsConfig, err = GetServerTLSConfig(verifyClient)
		if err != nil {
			return nil, err
		}
	}
	srvCfg := loadCfg()
	server = &Server{
		Server: &http.Server{
			Addr:              addr,
			Handler:           handler,
			TLSConfig:         tlsConfig,
			ReadTimeout:       srvCfg.ReadTimeout,
			ReadHeaderTimeout: srvCfg.ReadHeaderTimeout,
			WriteTimeout:      srvCfg.WriteTimeout,
			MaxHeaderBytes:    srvCfg.MaxHeaderBytes,
		},
		KeepaliveTimeout: srvCfg.KeepaliveTimeout,
		GraceTimeout:     srvCfg.GraceTimeout,
		state:            serverStateInit,
		Network:          "tcp",
	}
	return server, nil
}

func initServer(addr string, handler http.Handler) error {
	if defaultRESTfulServer != nil {
		return nil
	}
	var err error
	defaultRESTfulServer, err = NewServer(addr, handler)
	if err != nil {
		return err
	}
	util.Logger().Warnf(nil, "listen on server %s.", addr)

	return nil
}

func ListenAndServeTLS(addr string, handler http.Handler) (err error) {
	err = initServer(addr, handler)
	if err != nil {
		return err
	}
	// 证书已经在config里加载，这里不需要再重新加载
	return defaultRESTfulServer.ListenAndServeTLS("", "")
}
func ListenAndServe(addr string, handler http.Handler) (err error) {
	err = initServer(addr, handler)
	if err != nil {
		return err
	}
	return defaultRESTfulServer.ListenAndServe()
}

func GracefulStop() {
	if defaultRESTfulServer == nil {
		return
	}
	defaultRESTfulServer.Shutdown()
}

func DefaultServer() *Server {
	return defaultRESTfulServer
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
		}
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
