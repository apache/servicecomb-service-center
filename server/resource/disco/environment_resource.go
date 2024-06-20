package disco

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	pb "github.com/go-chassis/cari/discovery"
	ev "github.com/go-chassis/cari/env"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/pkg/rest"
	"github.com/apache/servicecomb-service-center/pkg/util"
	discosvc "github.com/apache/servicecomb-service-center/server/service/disco"
)

type EnvironmentResource struct {
	//
}

func (s *EnvironmentResource) URLPatterns() []rest.Route {
	return []rest.Route{
		{Method: http.MethodGet, Path: "/v4/:project/registry/environments", Func: s.ListEnvironments},
		{Method: http.MethodPost, Path: "/v4/:project/registry/environments", Func: s.RegistryEnvironment},
		{Method: http.MethodGet, Path: "/v4/:project/registry/environments/:environmentId", Func: s.GetEnvironment},
		{Method: http.MethodPut, Path: "/v4/:project/registry/environments/:environmentId", Func: s.UpdateEnvironment},
		{Method: http.MethodDelete, Path: "/v4/:project/registry/environments/:environmentId", Func: s.UnRegistryEnvironment},
	}
}

func (s *EnvironmentResource) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	resp, err := discosvc.ListEnvironments(r.Context())
	if err != nil {
		log.Error("list envs failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *EnvironmentResource) RegistryEnvironment(w http.ResponseWriter, r *http.Request) {
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	var request ev.CreateEnvironmentRequest
	err = json.Unmarshal(message, &request.Environment)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	resp, err := discosvc.RegistryEnvironment(r.Context(), &request)
	if err != nil {
		log.Error("create service failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, resp)
}

func (s *EnvironmentResource) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	request := &ev.GetEnvironmentRequest{
		EnvironmentId: r.URL.Query().Get(":environmentId"),
	}
	environment, err := discosvc.GetEnvironment(r.Context(), request)
	if err != nil {
		log.Error(fmt.Sprintf("get environment[%s] failed", request.EnvironmentId), err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, &ev.GetEnvironmentResponse{Environment: environment})
}

func (s *EnvironmentResource) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	environmentID := query.Get(":environmentId")
	message, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	var request ev.UpdateEnvironmentRequest
	err = json.Unmarshal(message, &request.Environment)
	if err != nil {
		log.Error(fmt.Sprintf("invalid json: %s", util.BytesToStringWithNoCopy(message)), err)
		rest.WriteError(w, pb.ErrInvalidParams, err.Error())
		return
	}
	request.Environment.ID = environmentID
	err = discosvc.UpdateEnvironment(r.Context(), &request)
	if err != nil {
		log.Error("update environment failed", err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}

func (s *EnvironmentResource) UnRegistryEnvironment(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	environmentID := query.Get(":environmentId")

	request := &ev.DeleteEnvironmentRequest{
		EnvironmentId: environmentID,
	}
	err := discosvc.UnRegistryEnvironment(r.Context(), request)
	if err != nil {
		log.Error(fmt.Sprintf("delete environment[%s] failed", environmentID), err)
		rest.WriteServiceError(w, err)
		return
	}
	rest.WriteResponse(w, r, nil, nil)
}
