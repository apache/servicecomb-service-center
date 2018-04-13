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
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/ServiceComb/paas-lager/third_party/forked/cloudfoundry/lager"
	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
	"github.com/apache/incubator-servicecomb-service-center/server/core"
	"github.com/apache/incubator-servicecomb-service-center/server/core/backend"
	pb "github.com/apache/incubator-servicecomb-service-center/server/core/proto"
	scerr "github.com/apache/incubator-servicecomb-service-center/server/error"
	"github.com/apache/incubator-servicecomb-service-center/server/infra/registry"
	serviceUtil "github.com/apache/incubator-servicecomb-service-center/server/service/util"
	"path/filepath"
	"time"
)

var PactLogger lager.Logger

const (
	BROKER_HOME_URL                      = "/"
	BROKER_PARTICIPANTS_URL              = "/participants"
	BROKER_PARTICIPANT_URL               = "/participants/:participantId"
	BROKER_PARTY_VERSIONS_URL            = "/participants/:participantId/versions"
	BROKER_PARTY_LATEST_VERSION_URL      = "/participants/:participantId/versions/latest"
	BROKER_PARTY_VERSION_URL             = "/participants/:participantId/versions/:number"
	BROKER_PROVIDER_URL                  = "/pacts/provider"
	BROKER_PROVIDER_LATEST_PACTS_URL     = "/pacts/provider/:providerId/latest"
	BROKER_PROVIDER_LATEST_PACTS_TAG_URL = "/pacts/provider/:providerId/latest/:tag"
	BROKER_PACTS_LATEST_URL              = "/pacts/latest"

	BROKER_PUBLISH_URL              = "/pacts/provider/:providerId/consumer/:consumerId/version/:number"
	BROKER_PUBLISH_VERIFICATION_URL = "/pacts/provider/:providerId/consumer/:consumerId/pact-version/:pact/verification-results"
	BROKER_WEBHOOHS_URL             = "/webhooks"

	BROKER_CURIES_URL = "/doc/:rel"
)

var brokerAPILinksValues = map[string]string{
	"self":                              BROKER_HOME_URL,
	"pb:publish-pact":                   BROKER_PUBLISH_URL,
	"pb:latest-pact-versions":           BROKER_PACTS_LATEST_URL,
	"pb:pacticipants":                   BROKER_PARTICIPANTS_URL,
	"pb:latest-provider-pacts":          BROKER_PROVIDER_LATEST_PACTS_URL,
	"pb:latest-provider-pacts-with-tag": BROKER_PROVIDER_LATEST_PACTS_TAG_URL,
	"pb:webhooks":                       BROKER_WEBHOOHS_URL,
}

var brokerAPILinksTempl = map[string]bool{
	"self":                              false,
	"pb:publish-pact":                   true,
	"pb:latest-pact-versions":           false,
	"pb:pacticipants":                   false,
	"pb:latest-provider-pacts":          true,
	"pb:latest-provider-pacts-with-tag": true,
	"pb:webhooks":                       false,
}

var brokerAPILinksTitles = map[string]string{
	"self":                              "Index",
	"pb:publish-pact":                   "Publish a pact",
	"pb:latest-pact-versions":           "Latest pact versions",
	"pb:pacticipants":                   "Pacticipants",
	"pb:latest-provider-pacts":          "Latest pacts by provider",
	"pb:latest-provider-pacts-with-tag": "Latest pacts by provider with a specified tag",
	"pb:webhooks":                       "Webhooks",
}

func init() {
	//define Broker logger
	name := ""
	if len(core.ServerInfo.Config.LogFilePath) != 0 {
		name = filepath.Join(filepath.Dir(core.ServerInfo.Config.LogFilePath), "broker_srvc.log")
	}
	PactLogger = util.NewLogger(util.LoggerConfig{
		LoggerLevel:     core.ServerInfo.Config.LogLevel,
		LoggerFile:      name,
		LogFormatText:   core.ServerInfo.Config.LogFormat == "text",
		LogRotatePeriod: 30 * time.Second,
		LogRotateSize:   int(core.ServerInfo.Config.LogRotateSize),
		LogBackupCount:  int(core.ServerInfo.Config.LogBackupCount),
	})
}

func GetDefaultTenantProject() string {
	return util.StringJoin([]string{"default", "default"}, "/")
}

//GenerateBrokerAPIPath creates the API link from the constant template
func GenerateBrokerAPIPath(scheme string, host string, apiPath string,
	replacer *strings.Replacer) string {
	genPath := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   apiPath,
	}
	if replacer != nil {
		return replacer.Replace(genPath.String())
	}
	return genPath.String()
}

//GetBrokerHomeLinksAPIS return the generated Home links
func GetBrokerHomeLinksAPIS(scheme string, host string, apiKey string) string {
	return GenerateBrokerAPIPath(scheme, host, brokerAPILinksValues[apiKey],
		strings.NewReplacer(":providerId", "{provider}",
			":consumerId", "{consumer}",
			":number", "{consumerApplicationVersion}",
			":tag", "{tag}"))
}

//CreateBrokerHomeResponse create the templated broker home response
func CreateBrokerHomeResponse(host string, scheme string) *BrokerHomeResponse {

	var apiEntries map[string]*BrokerAPIInfoEntry
	apiEntries = make(map[string]*BrokerAPIInfoEntry)

	for k := range brokerAPILinksValues {
		apiEntries[k] = &BrokerAPIInfoEntry{
			Href:      GetBrokerHomeLinksAPIS(scheme, host, k),
			Title:     brokerAPILinksTitles[k],
			Templated: brokerAPILinksTempl[k],
		}
	}

	curies := []*BrokerAPIInfoEntry{}
	curies = append(curies, &BrokerAPIInfoEntry{
		Name: "pb",
		Href: GenerateBrokerAPIPath(scheme, host, BROKER_CURIES_URL,
			strings.NewReplacer(":rel", "{rel}")),
	})

	return &BrokerHomeResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Broker Home."),
		XLinks:   apiEntries,
		Curies:   curies,
	}
}

//GetBrokerHomeResponse gets the homeResponse from cache if it exists
func GetBrokerHomeResponse(host string, scheme string) *BrokerHomeResponse {
	brokerResp := CreateBrokerHomeResponse(host, scheme)
	if brokerResp == nil {
		return nil
	}
	return brokerResp
}

//GetBrokerParticipantUtils returns the participant from ETCD
func GetBrokerParticipantUtils(ctx context.Context, tenant string, appId string,
	serviceName string, opts ...registry.PluginOpOption) (*Participant, error) {

	key := GenerateBrokerParticipantKey(tenant, appId, serviceName)
	opts = append(opts, registry.WithStrKey(key))
	participants, err := Store().Participant().Search(ctx, opts...)

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant with, could not be searched.")
		return nil, err
	}

	if len(participants.Kvs) == 0 {
		PactLogger.Info("GetParticipant found no participant")
		return nil, nil
	}

	participant := &Participant{}
	err = json.Unmarshal(participants.Kvs[0].Value, participant)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetParticipant: (%d, %s, %s)", participant.Id, participant.AppId,
		participant.ServiceName)
	return participant, nil
}

//GetBrokerParticipantFromServiceId returns the participant and the service from ETCD
func GetBrokerParticipantFromServiceId(ctx context.Context, serviceId string) (*Participant,
	*pb.MicroService, error, error) {

	tenant := GetDefaultTenantProject()
	serviceParticipant, err := serviceUtil.GetService(ctx, tenant, serviceId)
	if err != nil {
		PactLogger.Errorf(err,
			"get participant failed, serviceId is %s: query provider failed.", serviceId)
		return nil, nil, nil, err
	}
	if serviceParticipant == nil {
		PactLogger.Errorf(nil,
			"get participant failed, serviceId is %s: service not exist.", serviceId)
		return nil, nil, nil, errors.New("get participant, serviceId not exist.")
	}
	// Get or create provider participant
	participant, errBroker := GetBrokerParticipantUtils(ctx, tenant, serviceParticipant.AppId,
		serviceParticipant.ServiceName)
	if errBroker != nil {
		PactLogger.Errorf(errBroker,
			"get participant failed, serviceId %s: query participant failed.", serviceId)
		return nil, serviceParticipant, errBroker, err
	}
	if participant == nil {
		PactLogger.Errorf(nil,
			"get participant failed, particpant does not exist for serviceId %s", serviceId)
		return nil, serviceParticipant, errors.New("particpant does not exist for serviceId."), err
	}

	return participant, serviceParticipant, errBroker, nil
}

//GetBrokerParticipantFromService returns the participant given the microservice
func GetBrokerParticipantFromService(ctx context.Context,
	microservice *pb.MicroService) (*Participant, error) {
	if microservice == nil {
		return nil, nil
	}
	tenant := GetDefaultTenantProject()
	participant, errBroker := GetBrokerParticipantUtils(ctx, tenant, microservice.AppId,
		microservice.ServiceName)
	if errBroker != nil {
		PactLogger.Errorf(errBroker,
			"get participant failed, serviceId %s: query participant failed.",
			microservice.ServiceId)
		return nil, errBroker
	}
	return participant, errBroker
}

func GetParticipant(ctx context.Context, domain string, appId string,
	serviceName string) (*Participant, error) {
	key := GenerateBrokerParticipantKey(domain, appId, serviceName)
	participants, err := Store().Participant().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		PactLogger.Info("GetParticipant found no participant")
		return nil, nil
	}
	participant := &Participant{}
	err = json.Unmarshal(participants.Kvs[0].Value, participant)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetParticipant: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
	return participant, nil
}

func GetVersion(ctx context.Context, domain string, number string,
	participantId int32) (*Version, error) {
	key := GenerateBrokerVersionKey(domain, number, participantId)
	versions, err := Store().Version().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	version := &Version{}
	err = json.Unmarshal(versions.Kvs[0].Value, version)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetVersion: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	return version, nil
}

func GetPact(ctx context.Context, domain string, consumerParticipantId int32, producerParticipantId int32, sha []byte) (*Pact, error) {
	key := GenerateBrokerPactKey(domain, consumerParticipantId, producerParticipantId, sha)
	versions, err := Store().Pact().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	pact := &Pact{}
	err = json.Unmarshal(versions.Kvs[0].Value, pact)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetPact: (%d, %d, %d, %s, %s)", pact.Id, pact.ConsumerParticipantId, pact.ProviderParticipantId, string(pact.Sha), string(pact.Content))
	return pact, nil
}

func GetPactVersion(ctx context.Context, domain string, versionId int32,
	pactId int32) (*PactVersion, error) {
	key := GenerateBrokerPactVersionKey(domain, versionId, pactId)
	versions, err := Store().PactVersion().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	pactVersion := &PactVersion{}
	err = json.Unmarshal(versions.Kvs[0].Value, pactVersion)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetPactVersion: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
	return pactVersion, nil
}

func GetData(ctx context.Context, key string) (int, error) {
	values, err := Store().PactLatest().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return -1, err
	}
	if len(values.Kvs) == 0 {
		return -1, nil
	}
	id, err := strconv.Atoi(string(values.Kvs[0].Value))
	if err != nil {
		return -1, err
	}
	return id, nil
}

func StoreData(ctx context.Context, key string, value string) error {
	_, err := backend.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue([]byte(value)))
	return err
}

func CreateParticipant(pactLogger lager.Logger, ctx context.Context, participantKey string, participant Participant) (*PublishPactResponse, error) {
	data, err := json.Marshal(participant)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "participant cannot be created."),
		}, err
	}

	_, err = backend.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(participantKey),
		registry.WithValue(data))

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "participant cannot be created."),
		}, err
	}

	k := GetBrokerLatestParticipantIDKey()
	v := strconv.Itoa(int(participant.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "participant cannot be created."),
		}, err
	}
	PactLogger.Infof("Participant created for key: %s", participantKey)
	return nil, nil
}

func CreateVersion(pactLogger lager.Logger, ctx context.Context, versionKey string,
	version Version) (*PublishPactResponse, error) {
	data, err := json.Marshal(version)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be created."),
		}, err
	}

	_, err = backend.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(versionKey),
		registry.WithValue(data))
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be created."),
		}, err
	}
	k := GetBrokerLatestVersionIDKey()
	v := strconv.Itoa(int(version.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be created."),
		}, err
	}
	PactLogger.Infof("Version created for key: %s", versionKey)
	return nil, nil
}

func CreatePact(pactLogger lager.Logger, ctx context.Context,
	pactKey string, pact Pact) (*PublishPactResponse, error) {
	data, err := json.Marshal(pact)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be created."),
		}, err
	}

	_, err = backend.Registry().Do(ctx,
		registry.PUT,
		registry.WithStrKey(pactKey),
		registry.WithValue(data))

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be created."),
		}, err
	}
	k := GetBrokerLatestPactIDKey()
	v := strconv.Itoa(int(pact.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be created."),
		}, err
	}
	PactLogger.Infof("Pact created for key: %s", pactKey)
	return nil, nil
}

func CreatePactVersion(pactLogger lager.Logger, ctx context.Context, pactVersionKey string, pactVersion PactVersion) (*PublishPactResponse, error) {
	data, err := json.Marshal(pactVersion)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact version cannot be created."),
		}, err
	}

	_, err = backend.Registry().Do(ctx,
		registry.PUT, registry.WithValue(data), registry.WithStrKey(pactVersionKey))
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact version cannot be created."),
		}, err
	}
	k := GetBrokerLatestPactVersionIDKey()
	v := strconv.Itoa(int(pactVersion.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact version cannot be created."),
		}, err
	}
	PactLogger.Infof("Pact version created for key: %s", pactVersionKey)
	return nil, nil
}

func CreateVerification(pactLogger lager.Logger, ctx context.Context,
	verificationKey string, verification Verification) (*PublishVerificationResponse, error) {
	data, err := json.Marshal(verification)
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result marshal error.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "verification result marshal error."),
		}, err
	}

	_, err = backend.Registry().Do(ctx, registry.PUT,
		registry.WithStrKey(verificationKey),
		registry.WithValue(data))
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result cannot be created.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "verification result cannot be created."),
		}, err
	}
	k := GetBrokerLatestVerificationIDKey()
	v := strconv.Itoa(int(verification.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result cannot be created.")
		return &PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "verification result cannot be created."),
		}, err
	}
	PactLogger.Infof("Verification result created for key: %s", verificationKey)
	return nil, nil
}

func GetLastestVersionNumberForParticipant(ctx context.Context,
	tenant string, participantId int32) int32 {
	key := util.StringJoin([]string{
		GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := Store().Version().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())

	if err != nil || len(versions.Kvs) == 0 {
		return -1
	}
	order := int32(math.MinInt32)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &Version{}
		err = json.Unmarshal(versions.Kvs[i].Value, &version)
		if err != nil {
			return -1
		}
		if version.ParticipantId != participantId {
			continue
		}
		if version.Order > order {
			order = version.Order
		}
	}
	return order
}

func RetrieveProviderConsumerPact(ctx context.Context,
	in *GetProviderConsumerVersionPactRequest) (*GetProviderConsumerVersionPactResponse, int32, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 || len(in.Version) == 0 {
		PactLogger.Errorf(nil, "pact retrieve request failed: invalid params.")
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Request format invalid."),
		}, -1, nil
	}
	tenant := GetDefaultTenantProject()
	// Get provider microservice
	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		PactLogger.Errorf(err, "pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Query provider failed."),
		}, -1, err
	}
	if provider == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Provider does not exist."),
		}, -1, nil
	}
	// Get consumer microservice
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		PactLogger.Errorf(err, "pact retrieve failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Query consumer failed."),
		}, -1, err
	}
	if consumer == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Consumer does not exist."),
		}, -1, nil
	}
	// Get provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, provider participant %s cannot be searched.", in.ProviderId)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Provider participant cannot be searched."),
		}, -1, err
	}
	// Get consumer participant
	//consumerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, consumer.AppId, consumer.ServiceName)
	consumerParticipant, err := GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, consumer participant %s cannot be searched.", in.ConsumerId)
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "consumer participant cannot be searched."),
		}, -1, err
	}
	// Get or create version
	//versionKey := apt.GenerateBrokerVersionKey(tenant, in.Version, consumerParticipant.Id)
	version, err := GetVersion(ctx, tenant, in.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, version cannot be searched.")
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be searched."),
		}, -1, err
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{
		GetBrokerPactVersionKey(tenant),
		strconv.Itoa(int(version.Id))},
		"/")
	pactVersions, err := Store().PactVersion().Search(ctx,
		registry.WithPrefix(),
		registry.WithStrKey(pactVersionKey))

	if err != nil {
		return nil, -1, err
	}
	if len(pactVersions.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPact] No pact version found, sorry")
		return nil, -1, nil
	}
	pactIds := make(map[int32]int32)
	for i := 0; i < len(pactVersions.Kvs); i++ {
		pactVersion := &PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[i].Value, pactVersion)
		if err != nil {
			return nil, -1, err
		}
		// Obviously true, but checking it anyways
		if pactVersion.VersionId == version.Id {
			pactid := pactVersion.PactId
			pactIds[pactid] = pactid
		}
	}
	if len(pactIds) == 0 {
		PactLogger.Errorf(nil, "pact retrieve failed, pact cannot be found.")
		return &GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be found."),
		}, -1, err
	}
	pactKey := util.StringJoin([]string{
		GetBrokerPactKey(tenant),
		strconv.Itoa(int(consumerParticipant.Id)),
		strconv.Itoa(int(providerParticipant.Id))},
		"/")
	pacts, err := Store().PactVersion().Search(ctx,
		registry.WithStrKey(pactKey),
		registry.WithPrefix())

	if err != nil {
		return nil, -1, err
	}
	if len(pacts.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPact] No pact version found, sorry")
		return nil, -1, nil
	}
	for i := 0; i < len(pacts.Kvs); i++ {
		pactObj := &Pact{}
		err = json.Unmarshal(pacts.Kvs[i].Value, pactObj)
		if err != nil {
			return nil, -1, err
		}
		if _, ok := pactIds[pactObj.Id]; ok {
			//PactLogger.Infof("pact retrieve succeeded, found pact: %s", string(pactObj.Content))
			return &GetProviderConsumerVersionPactResponse{
				Response: pb.CreateResponse(pb.Response_SUCCESS, "pact found."),
				Pact:     pactObj.Content,
			}, pactObj.Id, nil
		}
	}
	PactLogger.Errorf(nil, "pact retrieve failed, pact cannot be found.")
	return &GetProviderConsumerVersionPactResponse{
		Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be found."),
	}, -1, nil
}

func DeletePactData(ctx context.Context,
	in *BaseBrokerRequest) (*pb.Response, error) {
	//tenant := util.ParseTenantProject(ctx)
	allPactKey := GetBrokerRootKey() //GetBrokerVerificationKey("default") //util.StringJoin([]string{ apt.GetRootKey(), apt.REGISTRY_PACT_ROOT_KEY }, "/")

	_, err := backend.Registry().Do(ctx,
		registry.DEL, registry.WithStrKey(allPactKey), registry.WithPrefix())
	if err != nil {
		return pb.CreateResponse(scerr.ErrInternal, "error deleting pacts."), err
	}
	return pb.CreateResponse(pb.Response_SUCCESS, "deleting pacts Succeed."), nil
}
