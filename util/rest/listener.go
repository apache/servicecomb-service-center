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
	"net"
	"os"
	"syscall"
)

type restListener struct {
	net.Listener
	stopCh chan error
	closed bool
	server *Server
}

func newRestListener(l net.Listener, srv *Server) (el *restListener) {
	el = &restListener{
		Listener: l,
		stopCh:   make(chan error),
		server:   srv,
	}
	go func() {
		<-el.stopCh
		el.closed = true
		el.stopCh <- el.Listener.Close()
	}()
	return
}

func (rl *restListener) Accept() (c net.Conn, err error) {
	tc, err := rl.Listener.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return
	}

	if rl.server.KeepaliveTimeout > 0 {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(rl.server.KeepaliveTimeout)
	} else {
		tc.SetKeepAlive(false)
	}

	c = restConn{
		Conn:   tc,
		server: rl.server,
	}

	rl.server.AcceptOne()
	return
}

func (rl *restListener) Close() error {
	if rl.closed {
		return syscall.EINVAL
	}
	rl.stopCh <- nil
	return <-rl.stopCh
}

func (rl *restListener) File() *os.File {
	tl := rl.Listener.(*net.TCPListener)
	fl, _ := tl.File()
	return fl
}
