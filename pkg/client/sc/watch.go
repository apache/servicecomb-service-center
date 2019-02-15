package sc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/apache/servicecomb-service-center/server/core"
	pb "github.com/apache/servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/servicecomb-service-center/server/error"
	"github.com/gorilla/websocket"
)

const (
	apiWatcherURL        = "/v4/%s/registry/microservices/%s/watcher"
	apiListAndWatcherURL = "/v4/%s/registry/microservices/%s/listwatcher"
)

func (c *SCClient) Watch(ctx context.Context, domainProject, selfServiceId string, callback func(*pb.WatchInstanceResponse)) *scerr.Error {
	domain, project := core.FromDomainProject(domainProject)
	headers := c.CommonHeaders(ctx)
	headers.Set("X-Domain-Name", domain)

	conn, err := c.WebsocketDial(ctx, fmt.Sprintf(apiWatcherURL, project, selfServiceId), headers)
	if err != nil {
		return scerr.NewError(scerr.ErrInternal, err.Error())
	}

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		if messageType == websocket.TextMessage {
			data := &pb.WatchInstanceResponse{}
			err := json.Unmarshal(message, data)
			if err != nil {
				log.Println(err)
				break
			}
			callback(data)
		}
	}
	return scerr.NewError(scerr.ErrInternal, err.Error())
}
