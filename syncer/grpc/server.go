package grpc

import (
	"context"
	"net"

	"github.com/apache/servicecomb-service-center/pkg/gopool"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"google.golang.org/grpc"
)

type GRPCHandler interface {
	GetData() *pb.SyncData
}
type PullHandle func() *pb.SyncData

// Server struct
type Server struct {
	lsn     net.Listener
	addr    string
	handler GRPCHandler
}

// NewServer new grpc server
func NewServer(addr string, handler GRPCHandler) *Server {
	return &Server{addr: addr, handler: handler}
}

// Provide consumers with an interface to pull data
func (s *Server) Pull(ctx context.Context, in *pb.PullRequest) (*pb.SyncData, error) {
	return s.handler.GetData(), nil
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
