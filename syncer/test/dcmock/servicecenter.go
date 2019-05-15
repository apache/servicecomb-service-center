package dcmock

import (
	"context"

	scpb "github.com/apache/servicecomb-service-center/server/core/proto"
	"github.com/apache/servicecomb-service-center/syncer/plugins"
	pb "github.com/apache/servicecomb-service-center/syncer/proto"
)

const PluginName = "testmock"

var (
	scCacheHandler func(ctx context.Context) (*pb.SyncData, error)
)

func SetGetAll(handler func(ctx context.Context) (*pb.SyncData, error)) {
	scCacheHandler = handler
}

func init() {
	plugins.RegisterPlugin(&plugins.Plugin{
		Kind: plugins.PluginDatacenter,
		Name: PluginName,
		New:  New,
	})
}

type adaptor struct{}

func New() plugins.PluginInstance {
	return &adaptor{}
}

func (*adaptor) New(endpoints []string) (plugins.Datacenter, error) {
	return &mockPlugin{}, nil
}

type mockPlugin struct {
}

func (c *mockPlugin) GetAll(ctx context.Context) (*pb.SyncData, error) {
	if scCacheHandler != nil {
		return scCacheHandler(ctx)
	}
	return &pb.SyncData{
		Services: []*pb.SyncService{
			{
				DomainProject: "default/default",
				Service: &scpb.MicroService{
					ServiceId:   "5db1b794aa6f8a875d6e68110260b5491ee7e223",
					AppId:       "default",
					ServiceName: "SERVICECENTER",
					Version:     "1.1.0",
					Level:       "BACK",
					Schemas: []string{
						"servicecenter.grpc.api.ServiceCtrl",
						"servicecenter.grpc.api.ServiceInstanceCtrl",
					},
					Status: "UP",
					Properties: map[string]string{
						"allowCrossApp": "true",
					},
					Timestamp:    "1552626180",
					ModTimestamp: "1552626180",
					Environment:  "production",
				},
				Instances: []*scpb.MicroServiceInstance{
					{
						InstanceId: "4d41a637471f11e9888cfa163eca30e0",
						ServiceId:  "5db1b794aa6f8a875d6e68110260b5491ee7e223",
						Endpoints: []string{
							"rest://127.0.0.1:30100/",
						},
						HostName: "testmock",
						Status:   "UP",
						HealthCheck: &scpb.HealthCheck{
							Mode:     "push",
							Interval: 30,
							Times:    3,
						},
						Timestamp:    "1552653537",
						ModTimestamp: "1552653537",
						Version:      "1.1.0",
					},
				},
			},
		},
	}, nil
}
