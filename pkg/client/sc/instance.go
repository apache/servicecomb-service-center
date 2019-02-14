package sc

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/error"
)

const (
	apiDiscoveryInstancesURL = "/v4/%s/registry/instances"
	apiHeartbeatSetURL       = "/v4/%s/registry/heartbeats"
	apiInstanceHeartbeatURL  = "/v4/%s/registry/microservices/%s/instances/%s/heartbeat"
)

func (c *SCClient) RegisterInstance(ctx context.Context, domainProject, serviceId string, instance *pb.MicroServiceInstance) (string, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.RegisterInstanceRequest{Instance: instance})
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPost,
		fmt.Sprintf(apiInstancesURL, project, serviceId),
		headers, reqBody)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return "", c.toError(body)
	}

	instancesResp := &pb.RegisterInstanceResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return "", scerr.NewError(scerr.ErrInternal, err.Error())
	}
	return instancesResp.InstanceId, nil
}

func (c *SCClient) UnregisterInstance(ctx context.Context, domainProject, serviceId, instanceId string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodDelete,
		fmt.Sprintf(apiInstanceURL, project, serviceId, instanceId),
		headers, nil)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *SCClient) DiscoveryInstances(ctx context.Context, domainProject, consumerId string, provider *pb.MicroService) ([]*pb.MicroServiceInstance, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)
	headers.Set("X-ConsumerId", consumerId)

	query := url.Values{}
	query.Set("appId", provider.AppId)
	query.Set("serviceName", provider.ServiceName)
	query.Set("version", provider.Version)

	resp, err := c.RestDoWithContext(ctx, http.MethodGet,
		fmt.Sprintf(apiDiscoveryInstancesURL, project)+"?"+c.parseQuery(ctx)+"&"+query.Encode(),
		headers, nil)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &pb.GetInstancesResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}

func (c *SCClient) Heartbeat(ctx context.Context, domainProject, serviceId, instanceId string) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiInstanceHeartbeatURL, project, serviceId, instanceId),
		headers, nil)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return c.toError(body)
	}

	return nil
}

func (c *SCClient) HeartbeatSet(ctx context.Context, domainProject string, instances ...*pb.HeartbeatSetElement) ([]*pb.InstanceHbRst, *scerr.Error) {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	reqBody, err := json.Marshal(&pb.HeartbeatSetRequest{Instances: instances})
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	resp, err := c.RestDoWithContext(ctx, http.MethodPut,
		fmt.Sprintf(apiHeartbeatSetURL, project),
		headers, reqBody)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.toError(body)
	}

	instancesResp := &pb.HeartbeatSetResponse{}
	err = json.Unmarshal(body, instancesResp)
	if err != nil {
		return nil, scerr.NewError(scerr.ErrInternal, err.Error())
	}

	return instancesResp.Instances, nil
}
