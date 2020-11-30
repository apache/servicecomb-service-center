package server

import (
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/go-chassis/go-chassis/v2"
	rf "github.com/go-chassis/go-chassis/v2/server/restful"
	"net/http"
)

const (
	Message = "Deliver full synchronization task success!"
)

func (s *Server) FullSync(b *rf.Context) {
	s.mux.Lock()
	s.triggered = true
	s.mux.Unlock()
	err := b.Write([]byte(Message))
	if err != nil {
		log.Error("", err)
	}
	return
}

func (s *Server) URLPatterns() []rf.Route {
	return []rf.Route{
		{Method: http.MethodGet, Path: "/v1/syncer/full-synchronization", ResourceFunc: s.FullSync},
	}
}

//if you use go run main.go instead of binary run, plz export CHASSIS_HOME=/{path}/{to}/server/

func (s *Server) NewHttpServer() {
	chassis.RegisterSchema("rest", s)
	if err := chassis.Init(); err != nil {
		log.Error("Init failed.", err)
		return
	}
	err := chassis.Run()
	if err != nil {
		log.Error("fail to run http server", err)
	}
}
