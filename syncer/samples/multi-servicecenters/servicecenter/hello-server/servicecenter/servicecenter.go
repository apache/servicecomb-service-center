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

package servicecenter

import (
	"context"
	"fmt"
	client2 "github.com/apache/servicecomb-service-center/client"
	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/ticker"
	"github.com/go-chassis/cari/discovery"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	domain, project       string
	cli                   *client2.Client
	once                  sync.Once
	heartbeatInterval     = 30
	providerCaches        = &sync.Map{}
	service               *discovery.MicroService
	instance              *discovery.MicroServiceInstance
	retryDiscover         = 3
	retryDiscoverInterval = 30
)

func Start(ctx context.Context, conf *Config) (err error) {
	once.Do(func() {
		cli, err = client2.NewSCClient(client2.Config{Endpoints: conf.Registry.Endpoints})
		if err != nil {
			log.Error("new service center client failed", err)
			return
		}

		if conf.Tenant != nil {
			domain, project = conf.Tenant.Domain, conf.Tenant.Project
		}

		consumerID := ""
		if conf.Service != nil {
			service, err = createService(ctx, conf.Service)
			if err != nil {
				return
			}

			if conf.Service.Instance != nil {
				instance, err = registerInstance(ctx, service.ServiceId, conf.Service.Instance)
				if err != nil {
					return
				}
				go heartbeat(ctx, instance)
			}
			consumerID = service.ServiceId
		}

		for i := 0; i <= retryDiscover; i++ {
			// 定时发送心跳
			err1 := discoveryToCaches(ctx, consumerID, conf.Provider)
			if err1 == nil {
				err = nil
				log.Infof("discovery provider success, appID = %s, name = %s, version = %s",
					conf.Provider.AppID, conf.Provider.Name, conf.Provider.Version)
				break
			}
			err = err1
			log.Warnf("discovery provider failed, appID = %s, name = %s, version = %s",
				conf.Provider.AppID, conf.Provider.Name, conf.Provider.Version)
			log.Info("waiting for retry")
			time.Sleep(time.Duration(retryDiscoverInterval) * time.Second)

		}

		go watchAndRenewCaches(ctx, conf.Provider)
	})
	return
}

func Stop(ctx context.Context) error {
	if instance != nil {
		// 注销微服务实例
		err := cli.UnregisterInstance(ctx, domain, project, instance.ServiceId, instance.InstanceId)
		if err != nil {
			return err
		}
	}

	// 实例注销后，服务中心清理数据需要一些时间，稍作延后
	time.Sleep(time.Second * 3)
	// 注销微服务
	return cli.DeleteService(ctx, domain, project, service.ServiceId)
}

func Do(ctx context.Context, method, addr string, headers http.Header, body []byte) (resp *http.Response, err error) {
	raw, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	endpoints, err := serverNameToEndpoints(raw.Hostname())

	client, err := client2.NewLBClient(endpoints, (&client2.Config{Endpoints: endpoints}).Merge())
	if err != nil {
		return nil, err
	}

	return client.RestDoWithContext(ctx, method, raw.Path, headers, body)
}

func createService(ctx context.Context, svc *MicroService) (*discovery.MicroService, error) {
	service := transformMicroService(svc)

	// 检测微服务是否存在
	serviceID, err := cli.ServiceExistence(ctx, domain, project, service.AppId, service.ServiceName, service.Version, "")
	if serviceID == "" {
		// 注册微服务
		serviceID, err = cli.CreateService(ctx, domain, project, service)
		if err != nil {
			return nil, err
		}
	}
	service.ServiceId = serviceID
	return service, nil
}

func registerInstance(ctx context.Context, serviceID string, ins *Instance) (*discovery.MicroServiceInstance, error) {
	instance := transformInstance(ins)
	instanceID, err := cli.RegisterInstance(ctx, domain, project, serviceID, instance)
	if err != nil {
		return nil, err
	}
	instance.ServiceId = serviceID
	instance.InstanceId = instanceID
	return instance, nil
}

func heartbeat(ctx context.Context, ins *discovery.MicroServiceInstance) {
	// 启动定时器：间隔30s
	t := ticker.NewTaskTicker(heartbeatInterval, func(ctx context.Context) {
		// 定时发送心跳
		err := cli.Heartbeat(ctx, domain, project, ins.ServiceId, ins.InstanceId)
		if err != nil {
			log.Errorf(err, "send heartbeat failed, domain = %s, serviceID = %s, instanceID = %s", domain, project, ins.ServiceId, ins.InstanceId)
			return
		}
		log.Debug("send heartbeat success")
	})
	defer t.Stop()
	t.Start(ctx)
}

func discoveryToCaches(ctx context.Context, consumerID string, provider *MicroService) error {
	service := transformMicroService(provider)
	list, err := cli.DiscoveryInstances(ctx, domain, project, consumerID, service.AppId, service.ServiceName, service.Version)
	if err != nil || len(list) == 0 {
		return fmt.Errorf("provider not found, serviceName: %s appID: %s, version: %s",
			provider.Name, provider.AppID, provider.Version)
	}
	// 缓存 provider 实例信息
	providerCaches.Store(provider.Name, list)
	provider.ID = list[0].ServiceId
	return nil
}

func watchAndRenewCaches(ctx context.Context, provider *MicroService) {
	err := cli.Watch(ctx, domain, project, provider.ID, func(result *discovery.WatchInstanceResponse) {
		log.Debug("reply from watch service")
		list, ok := providerCaches.Load(result.Instance.ServiceId)
		if !ok {
			log.Infof("provider \"%s\" not found", result.Instance.ServiceId)
			return
		}
		providerList := list.([]*discovery.MicroServiceInstance)

		renew := false
		for i, item := range providerList {
			if item.InstanceId != result.Instance.InstanceId {
				continue
			}
			if result.Action == "DELETE" {
				providerList = append(providerList[:i], providerList[i+1:]...)
			} else {
				providerList[i] = result.Instance
			}
			renew = true
			break
		}
		if !renew && result.Action != "DELETE" {
			providerList = append(providerList, result.Instance)
		}
		log.Debugf("update provider list: %s", providerList)
		providerCaches.Store(provider.Name, providerList)
	})
	if err != nil {
		log.Errorf(err, "watch service %s failed", provider.ID)
	}
}

func serverNameToEndpoints(name string) ([]string, error) {
	list, ok := providerCaches.Load(name)
	if !ok {
		return nil, fmt.Errorf("provider \"%s\" not found", name)
	}
	providerList := list.([]*discovery.MicroServiceInstance)
	endpointList := make([]string, 0, len(providerList))
	for i := 0; i < len(providerList); i++ {
		endpoints := providerList[i].Endpoints
		for j := 0; j < len(endpoints); j++ {
			addr, err := url.Parse(endpoints[j])
			if err != nil {
				log.Error("parse provider endpoint failed", err)
				continue
			}
			if addr.Scheme == "rest" {
				addr.Scheme = "http"
			}
			endpointList = append(endpointList, addr.String())
		}
	}
	return endpointList, nil
}

func transformMicroService(service *MicroService) *discovery.MicroService {
	return &discovery.MicroService{
		AppId:       service.AppID,
		ServiceId:   service.ID,
		ServiceName: service.Name,
		Version:     service.Version,
	}
}

func transformInstance(instance *Instance) *discovery.MicroServiceInstance {
	return &discovery.MicroServiceInstance{
		InstanceId: instance.ID,
		HostName:   instance.Hostname,
		Endpoints:  []string{instance.Protocol + "://" + instance.ListenAddress},
	}
}
