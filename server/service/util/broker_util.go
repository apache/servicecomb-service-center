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
package util

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ServiceComb/service-center/pkg/cache"
	"github.com/ServiceComb/service-center/pkg/lager"
	"github.com/ServiceComb/service-center/pkg/lager/core"
	"github.com/ServiceComb/service-center/pkg/util"
	apt "github.com/ServiceComb/service-center/server/core"
	pb "github.com/ServiceComb/service-center/server/core/proto"
	"github.com/ServiceComb/service-center/server/core/registry"
	"github.com/ServiceComb/service-center/server/core/registry/store"
)

const (
	BROKER_HOME_URL                      = "/broker"
	BROKER_PARTICIPANTS_URL              = "/broker/participants"
	BROKER_PARTICIPANT_URL               = "/broker/participants/:participantId"
	BROKER_PARTY_VERSIONS_URL            = "/broker/participants/:participantId/versions"
	BROKER_PARTY_LATEST_VERSION_URL      = "/broker/participants/:participantId/versions/latest"
	BROKER_PARTY_VERSION_URL             = "/broker/participants/:participantId/versions/:number"
	BROKER_PROVIDER_URL                  = "/broker/pacts/provider"
	BROKER_PROVIDER_LATEST_PACTS_URL     = "/broker/pacts/provider/:providerId/latest"
	BROKER_PROVIDER_LATEST_PACTS_TAG_URL = "/broker/pacts/provider/:providerId/latest/:tag"
	BROKER_PACTS_LATEST_URL              = "/broker/pacts/latest"

	BROKER_PUBLISH_URL              = "/broker/pacts/provider/:providerId/consumer/:consumerId/version/:number"
	BROKER_PUBLISH_VERIFICATION_URL = "/broker/pacts/provider/:providerId/consumer/:consumerId/pact-version/:pact/verification-results"
	BROKER_WEBHOOHS_URL             = "/broker/webhooks"

	BROKER_CURIES_URL = "/broker/doc/:rel"
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

var PactLogger core.Logger
var brokerCache *cache.Cache

func init() {
	//define broker objects cache
	d, _ := time.ParseDuration("2m")
	brokerCache = cache.New(d, d)
	//define Broker logger
	lager.Init(lager.Config{
		LoggerLevel:   "INFO",
		LoggerFile:    "broker_srvc.log",
		EnableRsyslog: false,
	})
	PactLogger = lager.NewLogger("broker_srvc")
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
func CreateBrokerHomeResponse(host string, scheme string) *pb.BrokerHomeResponse {

	var apiEntries map[string]*pb.BrokerAPIInfoEntry
	apiEntries = make(map[string]*pb.BrokerAPIInfoEntry)

	for k := range brokerAPILinksValues {
		apiEntries[k] = &pb.BrokerAPIInfoEntry{
			Href:      GetBrokerHomeLinksAPIS(scheme, host, k),
			Title:     brokerAPILinksTitles[k],
			Templated: brokerAPILinksTempl[k],
		}
	}

	curies := []*pb.BrokerAPIInfoEntry{}
	curies = append(curies, &pb.BrokerAPIInfoEntry{
		Name: "pb",
		Href: GenerateBrokerAPIPath(scheme, host, BROKER_CURIES_URL,
			strings.NewReplacer(":rel", "{rel}")),
	})

	return &pb.BrokerHomeResponse{
		Response: pb.CreateResponse(pb.Response_SUCCESS, "Broker Home."),
		XLinks:   apiEntries,
		Curies:   curies,
	}
}

//GetBrokerHomeResponse gets the homeResponse from cache if it exists
func GetBrokerHomeResponse(host string, scheme string) *pb.BrokerHomeResponse {
	brokerResponse, ok := brokerCache.Get(scheme + "://" + host)
	if !ok {
		brokerResp := CreateBrokerHomeResponse(host, scheme)
		if brokerResp == nil {
			return nil
		}
		d, _ := time.ParseDuration("10m")
		brokerCache.Set(scheme+"://"+host, brokerResp, d)
		return brokerResp
	}
	return brokerResponse.(*pb.BrokerHomeResponse)
}

//GetBrokerParticipantUtils returns the participant from ETCD
func GetBrokerParticipantUtils(ctx context.Context, tenant string, appId string,
	serviceName string, opts ...registry.PluginOpOption) (*pb.Participant, error) {

	key := apt.GenerateBrokerParticipantKey(tenant, appId, serviceName)
	opts = append(opts, registry.WithStrKey(key))
	participants, err := store.Store().Participant().Search(ctx, opts...)

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant with, could not be searched.")
		return nil, err
	}

	if len(participants.Kvs) == 0 {
		PactLogger.Info("GetParticipant found no participant")
		return nil, nil
	}

	participant := &pb.Participant{}
	err = json.Unmarshal(participants.Kvs[0].Value, participant)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetParticipant: (%d, %s, %s)", participant.Id, participant.AppId,
		participant.ServiceName)
	return participant, nil
}

//GetBrokerParticipantFromServiceId returns the participant and the service from ETCD
func GetBrokerParticipantFromServiceId(ctx context.Context, serviceId string) (*pb.Participant,
	*pb.MicroService, error, error) {

	tenant := GetDefaultTenantProject()
	serviceParticipant, err := GetService(ctx, tenant, serviceId)
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
	microservice *pb.MicroService) (*pb.Participant, error) {
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
	serviceName string) (*pb.Participant, error) {
	key := apt.GenerateBrokerParticipantKey(domain, appId, serviceName)
	participants, err := store.Store().Participant().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		PactLogger.Info("GetParticipant found no participant")
		return nil, nil
	}
	participant := &pb.Participant{}
	err = json.Unmarshal(participants.Kvs[0].Value, participant)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetParticipant: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
	return participant, nil
}

func GetVersion(ctx context.Context, domain string, number string,
	participantId int32) (*pb.Version, error) {
	key := apt.GenerateBrokerVersionKey(domain, number, participantId)
	versions, err := store.Store().Version().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	version := &pb.Version{}
	err = json.Unmarshal(versions.Kvs[0].Value, version)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetVersion: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	return version, nil
}

func GetPact(ctx context.Context, domain string, consumerParticipantId int32, producerParticipantId int32, sha []byte) (*pb.Pact, error) {
	key := apt.GenerateBrokerPactKey(domain, consumerParticipantId, producerParticipantId, sha)
	versions, err := store.Store().Pact().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	pact := &pb.Pact{}
	err = json.Unmarshal(versions.Kvs[0].Value, pact)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetPact: (%d, %d, %d, %s, %s)", pact.Id, pact.ConsumerParticipantId, pact.ProviderParticipantId, string(pact.Sha), string(pact.Content))
	return pact, nil
}

func GetPactVersion(ctx context.Context, domain string, versionId int32,
	pactId int32) (*pb.PactVersion, error) {
	key := apt.GenerateBrokerPactVersionKey(domain, versionId, pactId)
	versions, err := store.Store().PactVersion().Search(ctx, registry.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	pactVersion := &pb.PactVersion{}
	err = json.Unmarshal(versions.Kvs[0].Value, pactVersion)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetPactVersion: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
	return pactVersion, nil
}

func GetData(ctx context.Context, key string) (int, error) {
	values, err := store.Store().PactLatest().Search(ctx, registry.WithStrKey(key))
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
	_, err := registry.GetRegisterCenter().Do(ctx, registry.PUT,
		registry.WithStrKey(key),
		registry.WithValue([]byte(value)))
	return err
}

func CreateParticipant(pactLogger core.Logger, ctx context.Context, participantKey string, participant pb.Participant) (*pb.PublishPactResponse, error) {
	data, err := json.Marshal(participant)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "participant cannot be created."),
		}, err
	}

	_, err = registry.GetRegisterCenter().Do(ctx, registry.PUT,
		registry.WithStrKey(participantKey),
		registry.WithValue(data))

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "participant cannot be created."),
		}, err
	}

	k := apt.GetBrokerLatestParticipantIDKey()
	v := strconv.Itoa(int(participant.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "participant cannot be created."),
		}, err
	}
	PactLogger.Infof("Participant created for key: %s", participantKey)
	return nil, nil
}

func CreateVersion(pactLogger core.Logger, ctx context.Context, versionKey string,
	version pb.Version) (*pb.PublishPactResponse, error) {
	data, err := json.Marshal(version)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be created."),
		}, err
	}

	_, err = registry.GetRegisterCenter().Do(ctx, registry.PUT,
		registry.WithStrKey(versionKey),
		registry.WithValue(data))
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be created."),
		}, err
	}
	k := apt.GetBrokerLatestVersionIDKey()
	v := strconv.Itoa(int(version.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be created."),
		}, err
	}
	PactLogger.Infof("Version created for key: %s", versionKey)
	return nil, nil
}

func CreatePact(pactLogger core.Logger, ctx context.Context,
	pactKey string, pact pb.Pact) (*pb.PublishPactResponse, error) {
	data, err := json.Marshal(pact)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be created."),
		}, err
	}

	_, err = registry.GetRegisterCenter().Do(ctx,
		registry.PUT,
		registry.WithStrKey(pactKey),
		registry.WithValue(data))

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be created."),
		}, err
	}
	k := apt.GetBrokerLatestPactIDKey()
	v := strconv.Itoa(int(pact.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be created."),
		}, err
	}
	PactLogger.Infof("Pact created for key: %s", pactKey)
	return nil, nil
}

func CreatePactVersion(pactLogger core.Logger, ctx context.Context, pactVersionKey string, pactVersion pb.PactVersion) (*pb.PublishPactResponse, error) {
	data, err := json.Marshal(pactVersion)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be created."),
		}, err
	}

	_, err = registry.GetRegisterCenter().Do(ctx,
		registry.PUT, registry.WithValue(data), registry.WithStrKey(pactVersionKey))
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be created."),
		}, err
	}
	k := apt.GetBrokerLatestPactVersionIDKey()
	v := strconv.Itoa(int(pactVersion.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &pb.PublishPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact version cannot be created."),
		}, err
	}
	PactLogger.Infof("Pact version created for key: %s", pactVersionKey)
	return nil, nil
}

func CreateVerification(pactLogger core.Logger, ctx context.Context,
	verificationKey string, verification pb.Verification) (*pb.PublishVerificationResponse, error) {
	data, err := json.Marshal(verification)
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result marshal error.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "verification result marshal error."),
		}, err
	}

	_, err = registry.GetRegisterCenter().Do(ctx, registry.PUT,
		registry.WithStrKey(verificationKey),
		registry.WithValue(data))
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result cannot be created.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "verification result cannot be created."),
		}, err
	}
	k := apt.GetBrokerLatestVerificationIDKey()
	v := strconv.Itoa(int(verification.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = StoreData(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result cannot be created.")
		return &pb.PublishVerificationResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "verification result cannot be created."),
		}, err
	}
	PactLogger.Infof("Verification result created for key: %s", verificationKey)
	return nil, nil
}

func GetLastestVersionNumberForParticipant(ctx context.Context,
	tenant string, participantId int32) int32 {
	key := util.StringJoin([]string{
		apt.GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := store.Store().Version().Search(ctx,
		registry.WithStrKey(key),
		registry.WithPrefix())

	if err != nil || len(versions.Kvs) == 0 {
		return -1
	}
	order := int32(math.MinInt32)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &pb.Version{}
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
	in *pb.GetProviderConsumerVersionPactRequest) (*pb.GetProviderConsumerVersionPactResponse, int32, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 || len(in.Version) == 0 {
		PactLogger.Errorf(nil, "pact retrieve request failed: invalid params.")
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Request format invalid."),
		}, -1, nil
	}
	tenant := GetDefaultTenantProject()
	// Get provider microservice
	provider, err := GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		PactLogger.Errorf(err, "pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query provider failed."),
		}, -1, err
	}
	if provider == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider does not exist."),
		}, -1, nil
	}
	// Get consumer microservice
	consumer, err := GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		PactLogger.Errorf(err, "pact retrieve failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Query consumer failed."),
		}, -1, err
	}
	if consumer == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Consumer does not exist."),
		}, -1, nil
	}
	// Get provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, provider participant %s cannot be searched.", in.ProviderId)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "Provider participant cannot be searched."),
		}, -1, err
	}
	// Get consumer participant
	//consumerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, consumer.AppId, consumer.ServiceName)
	consumerParticipant, err := GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, consumer participant %s cannot be searched.", in.ConsumerId)
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "consumer participant cannot be searched."),
		}, -1, err
	}
	// Get or create version
	//versionKey := apt.GenerateBrokerVersionKey(tenant, in.Version, consumerParticipant.Id)
	version, err := GetVersion(ctx, tenant, in.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, version cannot be searched.")
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "version cannot be searched."),
		}, -1, err
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{
		apt.GetBrokerPactVersionKey(tenant),
		strconv.Itoa(int(version.Id))},
		"/")
	pactVersions, err := store.Store().PactVersion().Search(ctx,
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
		pactVersion := &pb.PactVersion{}
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
		return &pb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be found."),
		}, -1, err
	}
	pactKey := util.StringJoin([]string{
		apt.GetBrokerPactKey(tenant),
		strconv.Itoa(int(consumerParticipant.Id)),
		strconv.Itoa(int(providerParticipant.Id))},
		"/")
	pacts, err := store.Store().PactVersion().Search(ctx,
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
		pactObj := &pb.Pact{}
		err = json.Unmarshal(pacts.Kvs[i].Value, pactObj)
		if err != nil {
			return nil, -1, err
		}
		if _, ok := pactIds[pactObj.Id]; ok {
			//PactLogger.Infof("pact retrieve succeeded, found pact: %s", string(pactObj.Content))
			return &pb.GetProviderConsumerVersionPactResponse{
				Response: pb.CreateResponse(pb.Response_SUCCESS, "pact found."),
				Pact:     pactObj.Content,
			}, pactObj.Id, nil
		}
	}
	PactLogger.Errorf(nil, "pact retrieve failed, pact cannot be found.")
	return &pb.GetProviderConsumerVersionPactResponse{
		Response: pb.CreateResponse(pb.Response_FAIL, "pact cannot be found."),
	}, -1, nil
}

func DeletePactData(ctx context.Context,
	in *pb.BaseBrokerRequest) (*pb.Response, error) {
	//tenant := util.ParseTenantProject(ctx)
	allPactKey := apt.GetBrokerRootKey()//GetBrokerVerificationKey("default") //util.StringJoin([]string{ apt.GetRootKey(), apt.REGISTRY_PACT_ROOT_KEY }, "/")

	_, err := registry.GetRegisterCenter().Do(ctx,
		registry.DEL, registry.WithStrKey(allPactKey), registry.WithPrefix())
	if err != nil {
		return pb.CreateResponse(pb.Response_FAIL, "error deleting pacts."), err
	}
	return pb.CreateResponse(pb.Response_SUCCESS, "deleting pacts Succeed."), nil
}
