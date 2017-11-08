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
package service

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
	serviceUtil "github.com/ServiceComb/service-center/server/service/util"
	"golang.org/x/net/context"
)

type BrokerController struct {
}

func (s *BrokerController) GetBrokerHome(ctx context.Context,
	in *pb.BaseBrokerRequest) (*pb.BrokerHomeResponse, error) {

	if in == nil || len(in.HostAddress) == 0 {
		util.Logger().Errorf(nil, "Get Participant versions request failed: invalid params.")
		return &pb.BrokerHomeResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}

	return serviceUtil.GetBrokerHomeResponse(in.HostAddress, in.Scheme), nil
}

func (s *BrokerController) GetPactsOfProvider(ctx context.Context,
	in *pb.GetProviderConsumerVersionPactRequest) (*pb.GetProviderConsumerVersionPactResponse, error) {
	serviceUtil.PactLogger.Infof("GetPactsOfProvider: (%s, %s, %s)\n",
		in.ProviderId, in.ConsumerId, in.Version)

	resp, pactId, err := serviceUtil.RetrieveProviderConsumerPact(ctx, in)
	if err != nil || resp.GetPact() == nil || pactId == -1 {
		serviceUtil.PactLogger.Errorf(nil, "Get pacts of provider failed: %s\n",
			resp.Response.Message)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, err.Error()),
		}, err
	}

	urlValue := serviceUtil.GenerateBrokerAPIPath(in.BaseUrl.Scheme, in.BaseUrl.HostAddress,
		serviceUtil.BROKER_PUBLISH_VERIFICATION_URL,
		strings.NewReplacer(":providerId", in.ProviderId,
			":consumerId", in.ConsumerId,
			":pact", fmt.Sprint(pactId)))

	links := ",\"_links\": {" +
		"\"pb:publish-verification-results\": {" +
		"\"title\": \"Publish verification results\"," +
		"\"href\": \"" + urlValue +
		"\"" +
		"}" +
		"}}"

	linksBytes := []byte(links)
	pactBytes := resp.GetPact()
	sliceOfResp := pactBytes[0 : len(pactBytes)-2]
	finalBytes := append(sliceOfResp, linksBytes...)

	return &pb.GetProviderConsumerVersionPactResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Success."),
		Pact:     finalBytes,
	}, nil
	//controller.WriteText(http.StatusBadRequest, resp.Response.Message, w)

}

func (s *BrokerController) DeletePacts(ctx context.Context,
	in *pb.BaseBrokerRequest) (*pb.Response, error) {

	resp, err := serviceUtil.DeletePactData(ctx, in)

	return resp, err
}

//TODO: Redundant to be removed
func (s *BrokerController) RetrieveProviderPacts(ctx context.Context,
	in *pb.GetAllProviderPactsRequest) (*pb.GetAllProviderPactsReponse, error) {
	if in == nil || len(in.ProviderId) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "all provider pact retrieve request failed: invalid params.")
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tenant := serviceUtil.GetDefaultTenantProject()

	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		serviceUtil.PactLogger.Errorf(err, "all provider pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query provider failed."),
		}, err
	}
	if provider == nil {
		serviceUtil.PactLogger.Errorf(nil, "all provider pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider does not exist."),
		}, nil
	}
	// Get the provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := serviceUtil.GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		serviceUtil.PactLogger.Errorf(nil, "all provider pact retrieve failed, provider participant cannot be searched.", in.ProviderId)
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider participant cannot be searched."),
		}, err
	}
	serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Provider participant id : %d", providerParticipant.Id)
	// Get all versions
	versionKey := util.StringJoin([]string{apt.GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := store.Store().Version().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(versionKey))

	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		serviceUtil.PactLogger.Info("[RetrieveProviderPacts] No versions found, sorry")
		return nil, nil
	}
	// Store versions in a map
	versionObjects := make(map[int32]pb.Version)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &pb.Version{}
		err = json.Unmarshal(versions.Kvs[i].Value, version)
		if err != nil {
			return nil, err
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Version found : (%d, %s)", version.Id, version.Number)
		versionObjects[version.Id] = *version
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{apt.GetBrokerPactVersionKey(tenant), ""}, "/")
	pactVersions, err := store.Store().PactVersion().Search(ctx,
		registry.WithStrKey(pactVersionKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(pactVersions.Kvs) == 0 {
		serviceUtil.PactLogger.Info("[RetrieveProviderPacts] No pact version found, sorry")
		return nil, nil
	}
	participantToVersionObj := make(map[int32]pb.Version)
	for i := 0; i < len(pactVersions.Kvs); i++ {
		pactVersion := &pb.PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[i].Value, pactVersion)
		if err != nil {
			return nil, err
		}
		if pactVersion.ProviderParticipantId != providerParticipant.Id {
			continue
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Pact version found: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
		vObj := versionObjects[pactVersion.VersionId]
		if v1Obj, ok := participantToVersionObj[vObj.ParticipantId]; ok {
			if vObj.Order > v1Obj.Order {
				participantToVersionObj[vObj.ParticipantId] = vObj
			}
		} else {
			participantToVersionObj[vObj.ParticipantId] = vObj
		}
	}
	// Get all participants
	participantKey := util.StringJoin([]string{apt.GetBrokerParticipantKey(tenant), ""}, "/")
	participants, err := store.Store().Participant().Search(ctx,
		registry.WithStrKey(participantKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		return nil, nil
	}
	consumerInfoArr := make([]*pb.ConsumerInfo, 0)
	for i := 0; i < len(participants.Kvs); i++ {
		participant := &pb.Participant{}
		err = json.Unmarshal(participants.Kvs[i].Value, participant)
		if err != nil {
			return nil, err
		}
		if _, ok := participantToVersionObj[participant.Id]; !ok {
			continue
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Consumer found: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
		consumerVersion := participantToVersionObj[participant.Id].Number
		consumerId, err := serviceUtil.GetServiceId(ctx, &pb.MicroServiceKey{
			Tenant:      tenant,
			AppId:       participant.AppId,
			ServiceName: participant.ServiceName,
			Version:     consumerVersion,
		})
		if err != nil {
			return nil, err
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Consumer microservice found: %s", consumerId)

		urlValue := serviceUtil.GenerateBrokerAPIPath(in.BaseUrl.Scheme, in.BaseUrl.HostAddress,
			serviceUtil.BROKER_PUBLISH_URL,
			strings.NewReplacer(":providerId", in.ProviderId,
				":consumerId", consumerId,
				":number", consumerVersion))

		consumerInfo := &pb.ConsumerInfo{
			Href: urlValue,
			Name: consumerId,
		}
		consumerInfoArr = append(consumerInfoArr, consumerInfo)
	}
	links := &pb.Links{
		Pacts: consumerInfoArr,
	}
	resJson, err := json.Marshal(links)
	if err != nil {
		return nil, err
	}
	serviceUtil.PactLogger.Infof("Json : %s", string(resJson))
	response := &pb.GetAllProviderPactsReponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "retrieve provider pact info succeeded."),
		XLinks:   links,
	}
	return response, nil
}

func (s *BrokerController) GetAllProviderPacts(ctx context.Context,
	in *pb.GetAllProviderPactsRequest) (*pb.GetAllProviderPactsReponse, error) {

	if in == nil || len(in.ProviderId) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "all provider pact retrieve request failed: invalid params.")
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tenant := serviceUtil.GetDefaultTenantProject()

	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		serviceUtil.PactLogger.Errorf(err, "all provider pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query provider failed."),
		}, err
	}
	if provider == nil {
		serviceUtil.PactLogger.Errorf(nil, "all provider pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider does not exist."),
		}, nil
	}
	// Get the provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := serviceUtil.GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		serviceUtil.PactLogger.Errorf(nil, "all provider pact retrieve failed, provider participant cannot be searched.", in.ProviderId)
		return &pb.GetAllProviderPactsReponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider participant cannot be searched."),
		}, err
	}
	serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Provider participant id : %d", providerParticipant.Id)
	// Get all versions
	versionKey := util.StringJoin([]string{apt.GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := store.Store().Version().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(versionKey))

	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		serviceUtil.PactLogger.Info("[RetrieveProviderPacts] No versions found, sorry")
		return nil, nil
	}
	// Store versions in a map
	versionObjects := make(map[int32]pb.Version)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &pb.Version{}
		err = json.Unmarshal(versions.Kvs[i].Value, version)
		if err != nil {
			return nil, err
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Version found : (%d, %s)", version.Id, version.Number)
		versionObjects[version.Id] = *version
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{apt.GetBrokerPactVersionKey(tenant), ""}, "/")
	pactVersions, err := store.Store().PactVersion().Search(ctx,
		registry.WithStrKey(pactVersionKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(pactVersions.Kvs) == 0 {
		serviceUtil.PactLogger.Info("[RetrieveProviderPacts] No pact version found, sorry")
		return nil, nil
	}
	participantToVersionObj := make(map[int32]pb.Version)
	for i := 0; i < len(pactVersions.Kvs); i++ {
		pactVersion := &pb.PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[i].Value, pactVersion)
		if err != nil {
			return nil, err
		}
		if pactVersion.ProviderParticipantId != providerParticipant.Id {
			continue
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Pact version found: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
		vObj := versionObjects[pactVersion.VersionId]
		if v1Obj, ok := participantToVersionObj[vObj.ParticipantId]; ok {
			if vObj.Order > v1Obj.Order {
				participantToVersionObj[vObj.ParticipantId] = vObj
			}
		} else {
			participantToVersionObj[vObj.ParticipantId] = vObj
		}
	}
	// Get all participants
	participantKey := util.StringJoin([]string{apt.GetBrokerParticipantKey(tenant), ""}, "/")
	participants, err := store.Store().Participant().Search(ctx,
		registry.WithStrKey(participantKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		return nil, nil
	}
	consumerInfoArr := make([]*pb.ConsumerInfo, 0)
	for i := 0; i < len(participants.Kvs); i++ {
		participant := &pb.Participant{}
		err = json.Unmarshal(participants.Kvs[i].Value, participant)
		if err != nil {
			return nil, err
		}
		if _, ok := participantToVersionObj[participant.Id]; !ok {
			continue
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Consumer found: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
		consumerVersion := participantToVersionObj[participant.Id].Number
		consumerId, err := serviceUtil.GetServiceId(ctx, &pb.MicroServiceKey{
			Tenant:      tenant,
			AppId:       participant.AppId,
			ServiceName: participant.ServiceName,
			Version:     consumerVersion,
		})
		if err != nil {
			return nil, err
		}
		serviceUtil.PactLogger.Infof("[RetrieveProviderPacts] Consumer microservice found: %s", consumerId)

		urlValue := serviceUtil.GenerateBrokerAPIPath(in.BaseUrl.Scheme, in.BaseUrl.HostAddress,
			serviceUtil.BROKER_PUBLISH_URL,
			strings.NewReplacer(":providerId", in.ProviderId,
				":consumerId", consumerId,
				":number", consumerVersion))

		consumerInfo := &pb.ConsumerInfo{
			Href: urlValue,
			Name: consumerId,
		}
		consumerInfoArr = append(consumerInfoArr, consumerInfo)
	}
	links := &pb.Links{
		Pacts: consumerInfoArr,
	}
	resJson, err := json.Marshal(links)
	if err != nil {
		return nil, err
	}
	serviceUtil.PactLogger.Infof("Json : %s", string(resJson))
	response := &pb.GetAllProviderPactsReponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "retrieve provider pact info succeeded."),
		XLinks:   links,
	}
	return response, nil
}

func (s *BrokerController) RetrieveVerificationResults(ctx context.Context, in *pb.RetrieveVerificationRequest) (*pb.RetrieveVerificationResponse, error) {
	if in == nil || len(in.ConsumerId) == 0 || len(in.ConsumerVersion) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed: invalid params.")
		return &pb.RetrieveVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tenant := serviceUtil.GetDefaultTenantProject()
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		serviceUtil.PactLogger.Errorf(err, "verification result retrieve request failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &pb.RetrieveVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query consumer failed."),
		}, err
	}
	if consumer == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &pb.RetrieveVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Consumer does not exist."),
		}, nil
	}
	serviceUtil.PactLogger.Infof("Consumer service found: (%s, %s, %s, %s)", consumer.ServiceId, consumer.AppId, consumer.ServiceName, consumer.Version)
	// Get consumer participant
	consumerParticipant, err := serviceUtil.GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, consumer participant cannot be searched.", in.ConsumerId)
		return &pb.RetrieveVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "consumer participant cannot be searched."),
		}, err
	}
	serviceUtil.PactLogger.Infof("Consumer participant found: (%d, %s, %s)", consumerParticipant.Id, consumerParticipant.AppId, consumerParticipant.ServiceName)
	// Get version
	version, err := serviceUtil.GetVersion(ctx, tenant, consumer.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, version cannot be searched.")
		return &pb.RetrieveVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be searched."),
		}, err
	}
	serviceUtil.PactLogger.Infof("Version found/created: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	key := util.StringJoin([]string{apt.GetBrokerPactVersionKey(tenant), strconv.Itoa(int(version.Id))}, "/")
	pactVersions, err := store.Store().PactVersion().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(key))

	if err != nil || len(pactVersions.Kvs) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, pact version cannot be searched.")
		return &pb.RetrieveVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be searched."),
		}, err
	}
	overAllSuccess := false

	successfuls := make([]string, 0)
	fails := make([]string, 0)
	unknowns := make([]string, 0)

	verificationDetailsArr := make([]*pb.VerificationDetail, 0)
	for j := 0; j < len(pactVersions.Kvs); j++ {
		pactVersion := &pb.PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[j].Value, &pactVersion)
		if err != nil {
			serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, pact version cannot be searched.")
			return &pb.RetrieveVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be searched."),
			}, err
		}
		key = util.StringJoin([]string{apt.GetBrokerVerificationKey(tenant), strconv.Itoa(int(pactVersion.Id))}, "/")
		verifications, err := store.Store().Verification().Search(ctx,
			registry.WithPrefix(),
			registry.WithStrKey(key))

		if err != nil || len(verifications.Kvs) == 0 {
			serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, verification results cannot be searched.")
			return &pb.RetrieveVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "verification results cannot be searched."),
			}, err
		}
		lastNumber := int32(math.MinInt32)
		var lastVerificationResult *pb.Verification
		for i := 0; i < len(verifications.Kvs); i++ {
			verification := &pb.Verification{}
			err = json.Unmarshal(verifications.Kvs[i].Value, &verification)
			if err != nil {
				serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, verification result unmarshall error.")
				return &pb.RetrieveVerificationResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "verification result unmarshall error."),
				}, err
			}
			if verification.Number > lastNumber {
				lastNumber = verification.Number
				lastVerificationResult = verification
			}
		}
		if lastVerificationResult == nil {
			serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, verification result cannot be found.")
			return &pb.RetrieveVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "verification result cannot be found."),
			}, err
		}
		serviceUtil.PactLogger.Infof("Verification result found: (%d, %d, %d, %t, %s, %s, %s)",
			lastVerificationResult.Id, lastVerificationResult.Number, lastVerificationResult.PactVersionId,
			lastVerificationResult.Success, lastVerificationResult.ProviderVersion,
			lastVerificationResult.BuildUrl, lastVerificationResult.VerificationDate)

		key = util.StringJoin([]string{apt.GetBrokerParticipantKey(tenant), ""}, "/")
		participants, err := store.Store().Participant().Search(ctx,
			registry.WithStrKey(key),
			registry.WithPrefix())

		if err != nil || len(participants.Kvs) == 0 {
			serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, provider participant cannot be searched.")
			return &pb.RetrieveVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "provider participant cannot be searched."),
			}, err
		}
		var providerParticipant *pb.Participant
		for i := 0; i < len(participants.Kvs); i++ {
			participant := &pb.Participant{}
			err = json.Unmarshal(participants.Kvs[i].Value, &participant)
			if err != nil {
				serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, verification result unmarshall error.")
				return &pb.RetrieveVerificationResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "verification result unmarshall error."),
				}, err
			}
			if participant.Id == pactVersion.ProviderParticipantId {
				providerParticipant = participant
				break
			}
		}
		if providerParticipant == nil {
			serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, verification result unmarshall error.")
			return &pb.RetrieveVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "verification result unmarshall error."),
			}, err
		}
		serviceFindReq := &pb.GetExistenceRequest{
			Type:        "microservice",
			AppId:       providerParticipant.AppId,
			ServiceName: providerParticipant.ServiceName,
			Version:     lastVerificationResult.ProviderVersion,
		}
		resp, err := apt.ServiceAPI.Exist(ctx, serviceFindReq)
		if err != nil {
			serviceUtil.PactLogger.Errorf(nil, "verification result retrieve request failed, provider service cannot be found.")
			return &pb.RetrieveVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "provider service cannot be found."),
			}, err
		}
		providerName := resp.ServiceId
		verificationDetail := &pb.VerificationDetail{
			ProviderName:               providerName,
			ProviderApplicationVersion: lastVerificationResult.ProviderVersion,
			Success:                    lastVerificationResult.Success,
			VerificationDate:           lastVerificationResult.VerificationDate,
		}
		verificationDetailsArr = append(verificationDetailsArr, verificationDetail)
		if verificationDetail.Success == true {
			successfuls = append(successfuls, providerName)
		} else {
			fails = append(fails, providerName)
		}
		overAllSuccess = overAllSuccess && verificationDetail.Success
	}
	verificationDetails := &pb.VerificationDetails{VerificationResults: verificationDetailsArr}
	verificationSummary := &pb.VerificationSummary{Successful: successfuls, Failed: fails, Unknown: unknowns}
	verificationResult := &pb.VerificationResult{Success: overAllSuccess, ProviderSummary: verificationSummary, XEmbedded: verificationDetails}
	serviceUtil.PactLogger.Infof("Verification result retrieved successfully ...")
	return &pb.RetrieveVerificationResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Verification result retrieved successfully."),
		Result:   verificationResult,
	}, nil
}

func (s *BrokerController) PublishVerificationResults(ctx context.Context, in *pb.PublishVerificationRequest) (*pb.PublishVerificationResponse, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed: invalid params.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tenant := serviceUtil.GetDefaultTenantProject()
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		serviceUtil.PactLogger.Errorf(err, "verification result publish request failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query consumer failed."),
		}, err
	}
	if consumer == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Consumer does not exist."),
		}, nil
	}
	serviceUtil.PactLogger.Infof("Consumer service found: (%s, %s, %s, %s)", consumer.ServiceId, consumer.AppId, consumer.ServiceName, consumer.Version)
	// Get consumer participant
	consumerParticipant, err := serviceUtil.GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, consumer participant cannot be searched.", in.ConsumerId)
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "consumer participant cannot be searched."),
		}, err
	}
	serviceUtil.PactLogger.Infof("Consumer participant found: (%d, %s, %s)", consumerParticipant.Id, consumerParticipant.AppId, consumerParticipant.ServiceName)
	// Get version
	version, err := serviceUtil.GetVersion(ctx, tenant, consumer.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, version cannot be searched.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be searched."),
		}, err
	}
	serviceUtil.PactLogger.Infof("Version found/created: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	key := util.StringJoin([]string{apt.GetBrokerPactKey(tenant), ""}, "/")
	pacts, err := store.Store().Pact().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())

	if err != nil || len(pacts.Kvs) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, pact cannot be searched.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be searched."),
		}, err
	}
	pactExists := false
	for i := 0; i < len(pacts.Kvs); i++ {
		pact := &pb.Pact{}
		err = json.Unmarshal(pacts.Kvs[i].Value, &pact)
		if err != nil {
			serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, pact cannot be searched.")
			return &pb.PublishVerificationResponse{
				Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be searched."),
			}, err
		}
		if pact.Id == in.PactId {
			pactExists = true
		}
	}
	if pactExists == false {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, pact does not exists.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact does not exists."),
		}, err
	}
	pactVersion, err := serviceUtil.GetPactVersion(ctx, tenant, version.Id, in.PactId)
	if err != nil || pactVersion == nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, pact version cannot be searched.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be searched."),
		}, err
	}
	// Check if some verification results already exists
	key = util.StringJoin([]string{apt.GetBrokerVerificationKey(tenant), strconv.Itoa(int(pactVersion.Id))}, "/")
	verifications, err := store.Store().Verification().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())

	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, verification result cannot be searched.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "verification result cannot be searched."),
		}, err
	}
	lastNumber := int32(math.MinInt32)
	if len(verifications.Kvs) != 0 {
		for i := 0; i < len(verifications.Kvs); i++ {
			verification := &pb.Verification{}
			err = json.Unmarshal(verifications.Kvs[i].Value, &verification)
			if err != nil {
				serviceUtil.PactLogger.Errorf(nil, "verification result publish request failed, verification result unmarshall error.")
				return &pb.PublishVerificationResponse{
					Response: pb.CreateResponse(pb.Response_FAIL, "verification result unmarshall error."),
				}, err
			}
			if verification.Number > lastNumber {
				lastNumber = verification.Number
			}
		}
	}
	if lastNumber < 0 {
		lastNumber = 0
	} else {
		lastNumber++
	}
	verificationDate := time.Now().Format(time.RFC3339)
	verificationKey := apt.GenerateBrokerVerificationKey(tenant, pactVersion.Id, lastNumber)
	id, err := serviceUtil.GetData(ctx, apt.GetBrokerLatestVerificationIDKey())
	verification := &pb.Verification{
		Id:               int32(id) + 1,
		Number:           lastNumber,
		PactVersionId:    pactVersion.Id,
		Success:          in.Success,
		ProviderVersion:  in.ProviderApplicationVersion,
		BuildUrl:         "",
		VerificationDate: verificationDate,
	}
	response, err := serviceUtil.CreateVerification(serviceUtil.PactLogger, ctx, verificationKey, *verification)
	if err != nil {
		return response, err
	}
	serviceUtil.PactLogger.Infof("Verification result inserted: (%d, %d, %d, %t, %s, %s, %s)",
		verification.Id, verification.Number, verification.PactVersionId,
		verification.Success, verification.ProviderVersion, verification.BuildUrl, verification.VerificationDate)
	verificationResponse := &pb.VerificationDetail{
		ProviderName:               in.ProviderId,
		ProviderApplicationVersion: verification.ProviderVersion,
		Success:                    verification.Success,
		VerificationDate:           verification.VerificationDate,
	}
	serviceUtil.PactLogger.Infof("Verification result published successfully ...")
	return &pb.PublishVerificationResponse{
		Response:     pb.CreateResponse(pb.Response_SUCCESS, "Verification result published successfully."),
		Confirmation: verificationResponse,
	}, nil
}

func (s *BrokerController) PublishPact(ctx context.Context, in *pb.PublishPactRequest) (*pb.PublishPactResponse, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 || len(in.Version) == 0 || len(in.Pact) == 0 {
		serviceUtil.PactLogger.Errorf(nil, "pact publish request failed: invalid params.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, nil
	}
	tenant := serviceUtil.GetDefaultTenantProject()

	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		serviceUtil.PactLogger.Errorf(err, "pact publish failed, providerId is %s: query provider failed.", in.ProviderId)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query provider failed."),
		}, err
	}
	if provider == nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, providerId is %s: provider not exist.", in.ProviderId)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider does not exist."),
		}, nil
	}
	serviceUtil.PactLogger.Infof("Provider service found: (%s, %s, %s, %s)", provider.ServiceId, provider.AppId, provider.ServiceName, provider.Version)
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		serviceUtil.PactLogger.Errorf(err, "pact publish failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query consumer failed."),
		}, err
	}
	if consumer == nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Consumer does not exist."),
		}, nil
	}

	// check that the consumer has that vesion in the url
	if strings.Compare(consumer.GetVersion(), in.Version) != 0 {
		util.Logger().Errorf(nil,
			"pact publish failed, version (%s) does not exist for consmer", in.Version)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Consumer Version does not exist."),
		}, nil
	}

	serviceUtil.PactLogger.Infof("Consumer service found: (%s, %s, %s, %s)", consumer.ServiceId, consumer.AppId, consumer.ServiceName, consumer.Version)
	// Get or create provider participant
	providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := serviceUtil.GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, provider participant cannot be searched.", in.ProviderId)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider participant cannot be searched."),
		}, err
	}
	if providerParticipant == nil {
		id, err := serviceUtil.GetData(ctx, apt.GetBrokerLatestParticipantIDKey())
		providerParticipant = &pb.Participant{Id: int32(id) + 1, AppId: provider.AppId, ServiceName: provider.ServiceName}
		response, err := serviceUtil.CreateParticipant(serviceUtil.PactLogger, ctx, providerParticipantKey, *providerParticipant)
		if err != nil {
			return response, err
		}
	}
	serviceUtil.PactLogger.Infof("Provider participant found: (%d, %s, %s)", providerParticipant.Id, providerParticipant.AppId, providerParticipant.ServiceName)
	// Get or create consumer participant
	consumerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, consumer.AppId, consumer.ServiceName)
	consumerParticipant, err := serviceUtil.GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, consumer participant cannot be searched.", in.ConsumerId)
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "consumer participant cannot be searched."),
		}, err
	}
	if consumerParticipant == nil {
		id, err := serviceUtil.GetData(ctx, apt.GetBrokerLatestParticipantIDKey())
		consumerParticipant = &pb.Participant{Id: int32(id) + 1, AppId: consumer.AppId, ServiceName: consumer.ServiceName}
		response, err := serviceUtil.CreateParticipant(serviceUtil.PactLogger, ctx, consumerParticipantKey, *consumerParticipant)
		if err != nil {
			return response, err
		}
	}
	serviceUtil.PactLogger.Infof("Consumer participant found: (%d, %s, %s)", consumerParticipant.Id, consumerParticipant.AppId, consumerParticipant.ServiceName)
	// Get or create version
	versionKey := apt.GenerateBrokerVersionKey(tenant, in.Version, consumerParticipant.Id)
	version, err := serviceUtil.GetVersion(ctx, tenant, in.Version, consumerParticipant.Id)
	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, version cannot be searched.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be searched."),
		}, err
	}
	if version == nil {
		order := serviceUtil.GetLastestVersionNumberForParticipant(ctx, tenant, consumerParticipant.Id)
		serviceUtil.PactLogger.Infof("Old version order: %d", order)
		order++
		id, err := serviceUtil.GetData(ctx, apt.GetBrokerLatestVersionIDKey())
		version = &pb.Version{Id: int32(id) + 1, Number: in.Version, ParticipantId: consumerParticipant.Id, Order: order}
		response, err := serviceUtil.CreateVersion(serviceUtil.PactLogger, ctx, versionKey, *version)
		if err != nil {
			return response, err
		}
	}
	serviceUtil.PactLogger.Infof("Version found/created: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	// Get or create pact
	sha1 := sha1.Sum(in.Pact)
	var sha []byte = sha1[:]
	pactKey := apt.GenerateBrokerPactKey(tenant, consumerParticipant.Id, providerParticipant.Id, sha)
	pact, err := serviceUtil.GetPact(ctx, tenant, consumerParticipant.Id, providerParticipant.Id, sha)
	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, pact cannot be searched.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be searched."),
		}, err
	}
	if pact == nil {
		id, err := serviceUtil.GetData(ctx, apt.GetBrokerLatestPactIDKey())
		pact = &pb.Pact{Id: int32(id) + 1, ConsumerParticipantId: consumerParticipant.Id,
			ProviderParticipantId: providerParticipant.Id, Sha: sha, Content: in.Pact}
		response, err := serviceUtil.CreatePact(serviceUtil.PactLogger, ctx, pactKey, *pact)
		if err != nil {
			return response, err
		}
	}
	serviceUtil.PactLogger.Infof("Pact found/created: (%d, %d, %d, %s)", pact.Id, pact.ConsumerParticipantId, pact.ProviderParticipantId, pact.Sha)
	// Get or create pact version
	pactVersionKey := apt.GenerateBrokerPactVersionKey(tenant, version.Id, pact.Id)
	pactVersion, err := serviceUtil.GetPactVersion(ctx, tenant, version.Id, pact.Id)
	if err != nil {
		serviceUtil.PactLogger.Errorf(nil, "pact publish failed, pact version cannot be searched.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be searched."),
		}, err
	}
	if pactVersion == nil {
		id, err := serviceUtil.GetData(ctx, apt.GetBrokerLatestPactVersionIDKey())
		pactVersion = &pb.PactVersion{Id: int32(id) + 1, VersionId: version.Id, PactId: pact.Id, ProviderParticipantId: providerParticipant.Id}
		response, err := serviceUtil.CreatePactVersion(serviceUtil.PactLogger, ctx, pactVersionKey, *pactVersion)
		if err != nil {
			return response, err
		}
	}
	serviceUtil.PactLogger.Infof("PactVersion found/create: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
	serviceUtil.PactLogger.Infof("Pact published successfully ...")
	return &pb.PublishPactResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Pact published successfully."),
	}, nil
}
