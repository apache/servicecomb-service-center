package server

import (
	"context"
	"net"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	"github.com/apache/servicecomb-service-center/syncer/datacenter"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"google.golang.org/grpc"
)

// Server struct
type Server struct {
	lsn   net.Listener
	addr  string
	store datacenter.Store
}

// NewServer new grpc server
func NewServer(addr string, store datacenter.Store) *Server {
	return &Server{addr: addr, store: store}
}

// Provide consumers with an interface to pull data
func (s *Server) Pull(ctx context.Context, in *pb.PullRequest) (*pb.SyncData, error) {
	return s.store.LocalInfo(), nil
}

// Stop grpc server
func (s *Server) Stop() {
	if s.lsn == nil {
		return
	}
	s.lsn.Close()
}

// Run grpc server
func (s *Server) Run() (err error) {
	s.lsn, err = net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	svc := grpc.NewServer()
	pb.RegisterSyncServer(svc, s)
	gopool.Go(func(ctx context.Context) {
		svc.Serve(s.lsn)
	})
	return nil
}
