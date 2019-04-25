package client

import (
	"context"

	pb "github.com/apache/servicecomb-service-center/syncer/proto"
	"google.golang.org/grpc"
)

type Client struct {
	addr string
	cli  pb.SyncClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &Client{cli: pb.NewSyncClient(conn), addr: addr}, nil
}

func (c *Client) Pull(ctx context.Context) (*pb.SyncData, error) {
	return c.cli.Pull(ctx, &pb.PullRequest{})
}
