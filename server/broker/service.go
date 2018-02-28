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
package broker

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	apt "github.com/apache/incubator-servicecomb-service-center/server/core"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"golang.org/x/net/context"
)

var BrokerServiceAPI *BrokerService = &BrokerService{}

type BrokerService struct {
}

func (*BrokerService) GetBrokerHome(ctx context.Context,
	in *BaseBrokerRequest) (*BrokerHomeResponse, error) {

	if in == nil || len(in.HostAddress) == 0 {
		PactLogger.Errorf(nil, "Get Participant versions request failed: invalid params.")
		return &BrokerHomeResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}

	return GetBrokerHomeResponse(in.HostAddress, in.Scheme), nil
}

func (*BrokerService) GetPactsOfProvider(ctx context.Context,
	in *GetProviderConsumerVersionPactRequest) (*GetProviderConsumerVersionPactResponse, error) {
	PactLogger.Infof("GetPactsOfProvider: (%s, %s, %s)\n",
		in.ProviderId, in.ConsumerId, in.Version)

	resp, pactId, err := RetrieveProviderConsumerPact(ctx, in)
	if err != nil || resp.GetPact() == nil || pactId == -1 {
		PactLogger.Errorf(nil, "Get pacts of provider failed: %s\n",
			resp.Response.Message)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, err.Error()),
		}, err
	}

	urlValue := GenerateBrokerAPIPath(in.BaseUrl.Scheme, in.BaseUrl.HostAddress,
		BROKER_PUBLISH_VERIFICATION_URL,
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

	return &GetProviderConsumerVersionPactResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Success."),
		Pact:     finalBytes,
	}, nil
	//controller.WriteText(http.StatusBadRequest, resp.Response.Message, w)

}

func (*BrokerService) DeletePacts(ctx context.Context,
	in *BaseBrokerRequest) (*pb.Response, error) {

	resp, err := DeletePactData(ctx, in)

	return resp, err
}

func (*BrokerService) RetrieveProviderPacts(ctx context.Context,
	in *GetAllProviderPactsRequest) (*GetAllProviderPactsResponse, error) {
	if in == nil || len(in.ProviderId) == 0 {
		PactLogger.Errorf(nil, "all provider pact retrieve request failed: invalid params.")
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	tenant := GetDefaultTenantProject()

	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		PactLogger.Errorf(err, "all provider pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Query provider failed."),
		}, err
	}
	if provider == nil {
		PactLogger.Errorf(nil, "all provider pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Provider does not exist."),
		}, nil
	}
	// Get the provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		PactLogger.Errorf(nil, "all provider pact retrieve failed, provider participant cannot be searched.", in.ProviderId)
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Provider participant cannot be searched."),
		}, err
	}
	PactLogger.Infof("[RetrieveProviderPacts] Provider participant id : %d", providerParticipant.Id)
	// Get all versions
	versionKey := util.StringJoin([]string{GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := Store().Version().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(versionKey))

	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPacts] No versions found, sorry")
		return nil, nil
	}
	// Store versions in a map
	versionObjects := make(map[int32]Version)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &Version{}
		err = json.Unmarshal(versions.Kvs[i].Value, version)
		if err != nil {
			return nil, err
		}
		PactLogger.Infof("[RetrieveProviderPacts] Version found : (%d, %s)", version.Id, version.Number)
		versionObjects[version.Id] = *version
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{GetBrokerPactVersionKey(tenant), ""}, "/")
	pactVersions, err := Store().PactVersion().Search(ctx,
		registry.WithStrKey(pactVersionKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(pactVersions.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPacts] No pact version found, sorry")
		return nil, nil
	}
	participantToVersionObj := make(map[int32]Version)
	for i := 0; i < len(pactVersions.Kvs); i++ {
		pactVersion := &PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[i].Value, pactVersion)
		if err != nil {
			return nil, err
		}
		if pactVersion.ProviderParticipantId != providerParticipant.Id {
			continue
		}
		PactLogger.Infof("[RetrieveProviderPacts] Pact version found: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
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
	participantKey := util.StringJoin([]string{GetBrokerParticipantKey(tenant), ""}, "/")
	participants, err := Store().Participant().Search(ctx,
		registry.WithStrKey(participantKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		return nil, nil
	}
	consumerInfoArr := make([]*ConsumerInfo, 0)
	for i := 0; i < len(participants.Kvs); i++ {
		participant := &Participant{}
		err = json.Unmarshal(participants.Kvs[i].Value, participant)
		if err != nil {
			return nil, err
		}
		if _, ok := participantToVersionObj[participant.Id]; !ok {
			continue
		}
		PactLogger.Infof("[RetrieveProviderPacts] Consumer found: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
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
		PactLogger.Infof("[RetrieveProviderPacts] Consumer microservice found: %s", consumerId)

		urlValue := GenerateBrokerAPIPath(in.BaseUrl.Scheme, in.BaseUrl.HostAddress,
			BROKER_PUBLISH_URL,
			strings.NewReplacer(":providerId", in.ProviderId,
				":consumerId", consumerId,
				":number", consumerVersion))

		consumerInfo := &ConsumerInfo{
			Href: urlValue,
			Name: consumerId,
		}
		consumerInfoArr = append(consumerInfoArr, consumerInfo)
	}
	links := &Links{
		Pacts: consumerInfoArr,
	}
	resJson, err := json.Marshal(links)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("Json : %s", string(resJson))
	response := &GetAllProviderPactsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "retrieve provider pact info succeeded."),
		XLinks:   links,
	}
	return response, nil
}

func (*BrokerService) GetAllProviderPacts(ctx context.Context,
	in *GetAllProviderPactsRequest) (*GetAllProviderPactsResponse, error) {

	if in == nil || len(in.ProviderId) == 0 {
		PactLogger.Errorf(nil, "all provider pact retrieve request failed: invalid params.")
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	tenant := GetDefaultTenantProject()

	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		PactLogger.Errorf(err, "all provider pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Query provider failed."),
		}, err
	}
	if provider == nil {
		PactLogger.Errorf(nil, "all provider pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Provider does not exist."),
		}, nil
	}
	// Get the provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		PactLogger.Errorf(nil, "all provider pact retrieve failed, provider participant cannot be searched.", in.ProviderId)
		return &GetAllProviderPactsResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Provider participant cannot be searched."),
		}, err
	}
	PactLogger.Infof("[RetrieveProviderPacts] Provider participant id : %d", providerParticipant.Id)
	// Get all versions
	versionKey := util.StringJoin([]string{GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := Store().Version().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(versionKey))

	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPacts] No versions found, sorry")
		return nil, nil
	}
	// Store versions in a map
	versionObjects := make(map[int32]Version)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &Version{}
		err = json.Unmarshal(versions.Kvs[i].Value, version)
		if err != nil {
			return nil, err
		}
		PactLogger.Infof("[RetrieveProviderPacts] Version found : (%d, %s)", version.Id, version.Number)
		versionObjects[version.Id] = *version
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{GetBrokerPactVersionKey(tenant), ""}, "/")
	pactVersions, err := Store().PactVersion().Search(ctx,
		registry.WithStrKey(pactVersionKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(pactVersions.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPacts] No pact version found, sorry")
		return nil, nil
	}
	participantToVersionObj := make(map[int32]Version)
	for i := 0; i < len(pactVersions.Kvs); i++ {
		pactVersion := &PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[i].Value, pactVersion)
		if err != nil {
			return nil, err
		}
		if pactVersion.ProviderParticipantId != providerParticipant.Id {
			continue
		}
		PactLogger.Infof("[RetrieveProviderPacts] Pact version found: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
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
	participantKey := util.StringJoin([]string{GetBrokerParticipantKey(tenant), ""}, "/")
	participants, err := Store().Participant().Search(ctx,
		registry.WithStrKey(participantKey),
		registry.WithPrefix())

	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		return nil, nil
	}
	consumerInfoArr := make([]*ConsumerInfo, 0)
	for i := 0; i < len(participants.Kvs); i++ {
		participant := &Participant{}
		err = json.Unmarshal(participants.Kvs[i].Value, participant)
		if err != nil {
			return nil, err
		}
		if _, ok := participantToVersionObj[participant.Id]; !ok {
			continue
		}
		PactLogger.Infof("[RetrieveProviderPacts] Consumer found: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
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
		PactLogger.Infof("[RetrieveProviderPacts] Consumer microservice found: %s", consumerId)

		urlValue := GenerateBrokerAPIPath(in.BaseUrl.Scheme, in.BaseUrl.HostAddress,
			BROKER_PUBLISH_URL,
			strings.NewReplacer(":providerId", in.ProviderId,
				":consumerId", consumerId,
				":number", consumerVersion))

		consumerInfo := &ConsumerInfo{
			Href: urlValue,
			Name: consumerId,
		}
		consumerInfoArr = append(consumerInfoArr, consumerInfo)
	}
	links := &Links{
		Pacts: consumerInfoArr,
	}
	resJson, err := json.Marshal(links)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("Json : %s", string(resJson))
	response := &GetAllProviderPactsResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "retrieve provider pact info succeeded."),
		XLinks:   links,
	}
	return response, nil
}

func (*BrokerService) RetrieveVerificationResults(ctx context.Context, in *RetrieveVerificationRequest) (*RetrieveVerificationResponse, error) {
	if in == nil || len(in.ConsumerId) == 0 || len(in.ConsumerVersion) == 0 {
		PactLogger.Errorf(nil, "verification result retrieve request failed: invalid params.")
		return &RetrieveVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	tenant := GetDefaultTenantProject()
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		PactLogger.Errorf(err, "verification result retrieve request failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &RetrieveVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Query consumer failed."),
		}, err
	}
	if consumer == nil {
		PactLogger.Errorf(nil, "verification result retrieve request failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &RetrieveVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Consumer does not exist."),
		}, nil
	}
	PactLogger.Infof("Consumer service found: (%s, %s, %s, %s)", consumer.ServiceId, consumer.AppId, consumer.ServiceName, consumer.Version)
	// Get consumer participant
	consumerParticipant, err := GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		PactLogger.Errorf(nil, "verification result retrieve request failed, consumer participant cannot be searched.", in.ConsumerId)
		return &RetrieveVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "consumer participant cannot be searched."),
		}, err
	}
	PactLogger.Infof("Consumer participant found: (%d, %s, %s)", consumerParticipant.Id, consumerParticipant.AppId, consumerParticipant.ServiceName)
	// Get version
	version, err := GetVersion(ctx, tenant, consumer.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		PactLogger.Errorf(nil, "verification result retrieve request failed, version cannot be searched.")
		return &RetrieveVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "version cannot be searched."),
		}, err
	}
	PactLogger.Infof("Version found/created: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	key := util.StringJoin([]string{GetBrokerPactVersionKey(tenant), strconv.Itoa(int(version.Id))}, "/")
	pactVersions, err := Store().PactVersion().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(key))

	if err != nil || len(pactVersions.Kvs) == 0 {
		PactLogger.Errorf(nil, "verification result publish request failed, pact version cannot be searched.")
		return &RetrieveVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact version cannot be searched."),
		}, err
	}
	overAllSuccess := false

	successfuls := make([]string, 0)
	fails := make([]string, 0)
	unknowns := make([]string, 0)

	verificationDetailsArr := make([]*VerificationDetail, 0)
	for j := 0; j < len(pactVersions.Kvs); j++ {
		pactVersion := &PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[j].Value, &pactVersion)
		if err != nil {
			PactLogger.Errorf(nil, "verification result retrieve request failed, pact version cannot be searched.")
			return &RetrieveVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact version cannot be searched."),
			}, err
		}
		key = util.StringJoin([]string{GetBrokerVerificationKey(tenant), strconv.Itoa(int(pactVersion.Id))}, "/")
		verifications, err := Store().Verification().Search(ctx,
			registry.WithPrefix(),
			registry.WithStrKey(key))

		if err != nil || len(verifications.Kvs) == 0 {
			PactLogger.Errorf(nil, "verification result retrieve request failed, verification results cannot be searched.")
			return &RetrieveVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification results cannot be searched."),
			}, err
		}
		lastNumber := int32(math.MinInt32)
		var lastVerificationResult *Verification
		for i := 0; i < len(verifications.Kvs); i++ {
			verification := &Verification{}
			err = json.Unmarshal(verifications.Kvs[i].Value, &verification)
			if err != nil {
				PactLogger.Errorf(nil, "verification result retrieve request failed, verification result unmarshall error.")
				return &RetrieveVerificationResponse{
					Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification result unmarshall error."),
				}, err
			}
			if verification.Number > lastNumber {
				lastNumber = verification.Number
				lastVerificationResult = verification
			}
		}
		if lastVerificationResult == nil {
			PactLogger.Errorf(nil, "verification result retrieve request failed, verification result cannot be found.")
			return &RetrieveVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification result cannot be found."),
			}, err
		}
		PactLogger.Infof("Verification result found: (%d, %d, %d, %t, %s, %s, %s)",
			lastVerificationResult.Id, lastVerificationResult.Number, lastVerificationResult.PactVersionId,
			lastVerificationResult.Success, lastVerificationResult.ProviderVersion,
			lastVerificationResult.BuildUrl, lastVerificationResult.VerificationDate)

		key = util.StringJoin([]string{GetBrokerParticipantKey(tenant), ""}, "/")
		participants, err := Store().Participant().Search(ctx,
			registry.WithStrKey(key),
			registry.WithPrefix())

		if err != nil || len(participants.Kvs) == 0 {
			PactLogger.Errorf(nil, "verification result retrieve request failed, provider participant cannot be searched.")
			return &RetrieveVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "provider participant cannot be searched."),
			}, err
		}
		var providerParticipant *Participant
		for i := 0; i < len(participants.Kvs); i++ {
			participant := &Participant{}
			err = json.Unmarshal(participants.Kvs[i].Value, &participant)
			if err != nil {
				PactLogger.Errorf(nil, "verification result retrieve request failed, verification result unmarshall error.")
				return &RetrieveVerificationResponse{
					Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification result unmarshall error."),
				}, err
			}
			if participant.Id == pactVersion.ProviderParticipantId {
				providerParticipant = participant
				break
			}
		}
		if providerParticipant == nil {
			PactLogger.Errorf(nil, "verification result retrieve request failed, verification result unmarshall error.")
			return &RetrieveVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification result unmarshall error."),
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
			PactLogger.Errorf(nil, "verification result retrieve request failed, provider service cannot be found.")
			return &RetrieveVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "provider service cannot be found."),
			}, err
		}
		providerName := resp.ServiceId
		verificationDetail := &VerificationDetail{
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
	verificationDetails := &VerificationDetails{VerificationResults: verificationDetailsArr}
	verificationSummary := &VerificationSummary{Successful: successfuls, Failed: fails, Unknown: unknowns}
	verificationResult := &VerificationResult{Success: overAllSuccess, ProviderSummary: verificationSummary, XEmbedded: verificationDetails}
	PactLogger.Infof("Verification result retrieved successfully ...")
	return &RetrieveVerificationResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Verification result retrieved successfully."),
		Result:   verificationResult,
	}, nil
}

func (*BrokerService) PublishVerificationResults(ctx context.Context, in *PublishVerificationRequest) (*PublishVerificationResponse, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 {
		PactLogger.Errorf(nil, "verification result publish request failed: invalid params.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	tenant := GetDefaultTenantProject()
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		PactLogger.Errorf(err, "verification result publish request failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Query consumer failed."),
		}, err
	}
	if consumer == nil {
		PactLogger.Errorf(nil, "verification result publish request failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Consumer does not exist."),
		}, nil
	}
	PactLogger.Infof("Consumer service found: (%s, %s, %s, %s)", consumer.ServiceId, consumer.AppId, consumer.ServiceName, consumer.Version)
	// Get consumer participant
	consumerParticipant, err := GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		PactLogger.Errorf(nil, "verification result publish request failed, consumer participant cannot be searched.", in.ConsumerId)
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "consumer participant cannot be searched."),
		}, err
	}
	PactLogger.Infof("Consumer participant found: (%d, %s, %s)", consumerParticipant.Id, consumerParticipant.AppId, consumerParticipant.ServiceName)
	// Get version
	version, err := GetVersion(ctx, tenant, consumer.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		PactLogger.Errorf(nil, "verification result publish request failed, version cannot be searched.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "version cannot be searched."),
		}, err
	}
	PactLogger.Infof("Version found/created: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	key := util.StringJoin([]string{GetBrokerPactKey(tenant), ""}, "/")
	pacts, err := Store().Pact().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())

	if err != nil || len(pacts.Kvs) == 0 {
		PactLogger.Errorf(nil, "verification result publish request failed, pact cannot be searched.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact cannot be searched."),
		}, err
	}
	pactExists := false
	for i := 0; i < len(pacts.Kvs); i++ {
		pact := &Pact{}
		err = json.Unmarshal(pacts.Kvs[i].Value, &pact)
		if err != nil {
			PactLogger.Errorf(nil, "verification result publish request failed, pact cannot be searched.")
			return &PublishVerificationResponse{
				Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact cannot be searched."),
			}, err
		}
		if pact.Id == in.PactId {
			pactExists = true
		}
	}
	if pactExists == false {
		PactLogger.Errorf(nil, "verification result publish request failed, pact does not exists.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact does not exists."),
		}, err
	}
	pactVersion, err := GetPactVersion(ctx, tenant, version.Id, in.PactId)
	if err != nil || pactVersion == nil {
		PactLogger.Errorf(nil, "verification result publish request failed, pact version cannot be searched.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact version cannot be searched."),
		}, err
	}
	// Check if some verification results already exists
	key = util.StringJoin([]string{GetBrokerVerificationKey(tenant), strconv.Itoa(int(pactVersion.Id))}, "/")
	verifications, err := Store().Verification().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())

	if err != nil {
		PactLogger.Errorf(nil, "verification result publish request failed, verification result cannot be searched.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification result cannot be searched."),
		}, err
	}
	lastNumber := int32(math.MinInt32)
	if len(verifications.Kvs) != 0 {
		for i := 0; i < len(verifications.Kvs); i++ {
			verification := &Verification{}
			err = json.Unmarshal(verifications.Kvs[i].Value, &verification)
			if err != nil {
				PactLogger.Errorf(nil, "verification result publish request failed, verification result unmarshall error.")
				return &PublishVerificationResponse{
					Response: pb.CreateResponse(scerr.ErrInvalidParams, "verification result unmarshall error."),
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
	verificationKey := GenerateBrokerVerificationKey(tenant, pactVersion.Id, lastNumber)
	id, err := GetData(ctx, GetBrokerLatestVerificationIDKey())
	verification := &Verification{
		Id:               int32(id) + 1,
		Number:           lastNumber,
		PactVersionId:    pactVersion.Id,
		Success:          in.Success,
		ProviderVersion:  in.ProviderApplicationVersion,
		BuildUrl:         "",
		VerificationDate: verificationDate,
	}
	response, err := CreateVerification(PactLogger, ctx, verificationKey, *verification)
	if err != nil {
		return response, err
	}
	PactLogger.Infof("Verification result inserted: (%d, %d, %d, %t, %s, %s, %s)",
		verification.Id, verification.Number, verification.PactVersionId,
		verification.Success, verification.ProviderVersion, verification.BuildUrl, verification.VerificationDate)
	verificationResponse := &VerificationDetail{
		ProviderName:               in.ProviderId,
		ProviderApplicationVersion: verification.ProviderVersion,
		Success:                    verification.Success,
		VerificationDate:           verification.VerificationDate,
	}
	PactLogger.Infof("Verification result published successfully ...")
	return &PublishVerificationResponse{
		Response:     pb.CreateResponse(pb.Response_SUCCESS, "Verification result published successfully."),
		Confirmation: verificationResponse,
	}, nil
}

func (*BrokerService) PublishPact(ctx context.Context, in *PublishPactRequest) (*PublishPactResponse, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 || len(in.Version) == 0 || len(in.Pact) == 0 {
		PactLogger.Errorf(nil, "pact publish request failed: invalid params.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Request format invalid."),
		}, nil
	}
	tenant := GetDefaultTenantProject()

	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		PactLogger.Errorf(err, "pact publish failed, providerId is %s: query provider failed.", in.ProviderId)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Query provider failed."),
		}, err
	}
	if provider == nil {
		PactLogger.Errorf(nil, "pact publish failed, providerId is %s: provider not exist.", in.ProviderId)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Provider does not exist."),
		}, nil
	}
	PactLogger.Infof("Provider service found: (%s, %s, %s, %s)", provider.ServiceId, provider.AppId, provider.ServiceName, provider.Version)
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		PactLogger.Errorf(err, "pact publish failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Query consumer failed."),
		}, err
	}
	if consumer == nil {
		PactLogger.Errorf(nil, "pact publish failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Consumer does not exist."),
		}, nil
	}

	// check that the consumer has that vesion in the url
	if strings.Compare(consumer.GetVersion(), in.Version) != 0 {
		util.Logger().Errorf(nil,
			"pact publish failed, version (%s) does not exist for consmer", in.Version)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Consumer Version does not exist."),
		}, nil
	}

	PactLogger.Infof("Consumer service found: (%s, %s, %s, %s)", consumer.ServiceId, consumer.AppId, consumer.ServiceName, consumer.Version)
	// Get or create provider participant
	providerParticipantKey := GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, provider participant cannot be searched.", in.ProviderId)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "Provider participant cannot be searched."),
		}, err
	}
	if providerParticipant == nil {
		id, err := GetData(ctx, GetBrokerLatestParticipantIDKey())
		providerParticipant = &Participant{Id: int32(id) + 1, AppId: provider.AppId, ServiceName: provider.ServiceName}
		response, err := CreateParticipant(PactLogger, ctx, providerParticipantKey, *providerParticipant)
		if err != nil {
			return response, err
		}
	}
	PactLogger.Infof("Provider participant found: (%d, %s, %s)", providerParticipant.Id, providerParticipant.AppId, providerParticipant.ServiceName)
	// Get or create consumer participant
	consumerParticipantKey := GenerateBrokerParticipantKey(tenant, consumer.AppId, consumer.ServiceName)
	consumerParticipant, err := GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, consumer participant cannot be searched.", in.ConsumerId)
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "consumer participant cannot be searched."),
		}, err
	}
	if consumerParticipant == nil {
		id, err := GetData(ctx, GetBrokerLatestParticipantIDKey())
		consumerParticipant = &Participant{Id: int32(id) + 1, AppId: consumer.AppId, ServiceName: consumer.ServiceName}
		response, err := CreateParticipant(PactLogger, ctx, consumerParticipantKey, *consumerParticipant)
		if err != nil {
			return response, err
		}
	}
	PactLogger.Infof("Consumer participant found: (%d, %s, %s)", consumerParticipant.Id, consumerParticipant.AppId, consumerParticipant.ServiceName)
	// Get or create version
	versionKey := GenerateBrokerVersionKey(tenant, in.Version, consumerParticipant.Id)
	version, err := GetVersion(ctx, tenant, in.Version, consumerParticipant.Id)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be searched.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "version cannot be searched."),
		}, err
	}
	if version == nil {
		order := GetLastestVersionNumberForParticipant(ctx, tenant, consumerParticipant.Id)
		PactLogger.Infof("Old version order: %d", order)
		order++
		id, err := GetData(ctx, GetBrokerLatestVersionIDKey())
		version = &Version{Id: int32(id) + 1, Number: in.Version, ParticipantId: consumerParticipant.Id, Order: order}
		response, err := CreateVersion(PactLogger, ctx, versionKey, *version)
		if err != nil {
			return response, err
		}
	}
	PactLogger.Infof("Version found/created: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	// Get or create pact
	sha1 := sha1.Sum(in.Pact)
	var sha []byte = sha1[:]
	pactKey := GenerateBrokerPactKey(tenant, consumerParticipant.Id, providerParticipant.Id, sha)
	pact, err := GetPact(ctx, tenant, consumerParticipant.Id, providerParticipant.Id, sha)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be searched.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact cannot be searched."),
		}, err
	}
	if pact == nil {
		id, err := GetData(ctx, GetBrokerLatestPactIDKey())
		pact = &Pact{Id: int32(id) + 1, ConsumerParticipantId: consumerParticipant.Id,
			ProviderParticipantId: providerParticipant.Id, Sha: sha, Content: in.Pact}
		response, err := CreatePact(PactLogger, ctx, pactKey, *pact)
		if err != nil {
			return response, err
		}
	}
	PactLogger.Infof("Pact found/created: (%d, %d, %d, %s)", pact.Id, pact.ConsumerParticipantId, pact.ProviderParticipantId, pact.Sha)
	// Get or create pact version
	pactVersionKey := GenerateBrokerPactVersionKey(tenant, version.Id, pact.Id)
	pactVersion, err := GetPactVersion(ctx, tenant, version.Id, pact.Id)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be searched.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInvalidParams, "pact version cannot be searched."),
		}, err
	}
	if pactVersion == nil {
		id, err := GetData(ctx, GetBrokerLatestPactVersionIDKey())
		pactVersion = &PactVersion{Id: int32(id) + 1, VersionId: version.Id, PactId: pact.Id, ProviderParticipantId: providerParticipant.Id}
		response, err := CreatePactVersion(PactLogger, ctx, pactVersionKey, *pactVersion)
		if err != nil {
			return response, err
		}
	}
	PactLogger.Infof("PactVersion found/create: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
	PactLogger.Infof("Pact published successfully ...")
	return &PublishPactResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Pact published successfully."),
	}, nil
}
