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
	"github.com/apache/servicecomb-service-center/server/config"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/apache/servicecomb-service-center/datasource/etcd/client"
	serviceUtil "github.com/apache/servicecomb-service-center/datasource/etcd/util"
	"github.com/apache/servicecomb-service-center/pkg/log"
	pb "github.com/apache/servicecomb-service-center/pkg/registry"
	"github.com/apache/servicecomb-service-center/pkg/util"
	"github.com/apache/servicecomb-service-center/server/broker/brokerpb"
	scerr "github.com/apache/servicecomb-service-center/server/scerror"
	"path/filepath"
)

var PactLogger *log.Logger

const (
	HomeURL                   = "/"
	ParticipantsURL           = "/participants"
	ProviderLatestPactsURL    = "/pacts/provider/:providerId/latest"
	ProviderLatestPactsTagURL = "/pacts/provider/:providerId/latest/:tag"
	PactsLatestURL            = "/pacts/latest"

	PublishURL             = "/pacts/provider/:providerId/consumer/:consumerId/version/:number"
	PublishVerificationURL = "/pacts/provider/:providerId/consumer/:consumerId/pact-version/:pact/verification-results"
	WebhooksURL            = "/webhooks"

	CuriesURL = "/doc/:rel"
)

var brokerAPILinksValues = map[string]string{
	"self":                              HomeURL,
	"pb:publish-pact":                   PublishURL,
	"pb:latest-pact-versions":           PactsLatestURL,
	"pb:pacticipants":                   ParticipantsURL,
	"pb:latest-provider-pacts":          ProviderLatestPactsURL,
	"pb:latest-provider-pacts-with-tag": ProviderLatestPactsTagURL,
	"pb:webhooks":                       WebhooksURL,
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
	if len(config.GetLog().LogFilePath) != 0 {
		name = filepath.Join(filepath.Dir(config.GetLog().LogFilePath), "broker_srvc.log")
	}
	PactLogger = log.NewLogger(log.Config{
		LoggerLevel:    config.GetLog().LogLevel,
		LoggerFile:     name,
		LogFormatText:  config.GetLog().LogFormat == "text",
		LogRotateSize:  int(config.GetLog().LogRotateSize),
		LogBackupCount: int(config.GetLog().LogBackupCount),
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
func CreateBrokerHomeResponse(host string, scheme string) *brokerpb.BrokerHomeResponse {
	apiEntries := make(map[string]*brokerpb.BrokerAPIInfoEntry)

	for k := range brokerAPILinksValues {
		apiEntries[k] = &brokerpb.BrokerAPIInfoEntry{
			Href:      GetBrokerHomeLinksAPIS(scheme, host, k),
			Title:     brokerAPILinksTitles[k],
			Templated: brokerAPILinksTempl[k],
		}
	}

	curies := []*brokerpb.BrokerAPIInfoEntry{}
	curies = append(curies, &brokerpb.BrokerAPIInfoEntry{
		Name: "pb",
		Href: GenerateBrokerAPIPath(scheme, host, CuriesURL,
			strings.NewReplacer(":rel", "{rel}")),
	})

	return &brokerpb.BrokerHomeResponse{
		Response: pb.CreateResponse(pb.ResponseSuccess, "Broker Home."),
		XLinks:   apiEntries,
		Curies:   curies,
	}
}

//GetBrokerHomeResponse gets the homeResponse from cache if it exists
func GetBrokerHomeResponse(host string, scheme string) *brokerpb.BrokerHomeResponse {
	brokerResp := CreateBrokerHomeResponse(host, scheme)
	if brokerResp == nil {
		return nil
	}
	return brokerResp
}

func GetParticipant(ctx context.Context, domain string, appID string,
	serviceName string) (*brokerpb.Participant, error) {
	key := GenerateBrokerParticipantKey(domain, appID, serviceName)
	participants, err := Store().Participant().Search(ctx, client.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(participants.Kvs) == 0 {
		PactLogger.Info("GetParticipant found no participant")
		return nil, nil
	}
	participant := &brokerpb.Participant{}
	err = json.Unmarshal(participants.Kvs[0].Value.([]byte), participant)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetParticipant: (%d, %s, %s)", participant.Id, participant.AppId, participant.ServiceName)
	return participant, nil
}

func GetVersion(ctx context.Context, domain string, number string,
	participantID int32) (*brokerpb.Version, error) {
	key := GenerateBrokerVersionKey(domain, number, participantID)
	versions, err := Store().Version().Search(ctx, client.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	version := &brokerpb.Version{}
	err = json.Unmarshal(versions.Kvs[0].Value.([]byte), version)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetVersion: (%d, %s, %d, %d)", version.Id, version.Number, version.ParticipantId, version.Order)
	return version, nil
}

func GetPact(ctx context.Context, domain string, consumerParticipantID int32, producerParticipantID int32, sha []byte) (*brokerpb.Pact, error) {
	key := GenerateBrokerPactKey(domain, consumerParticipantID, producerParticipantID, sha)
	versions, err := Store().Pact().Search(ctx, client.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	pact := &brokerpb.Pact{}
	err = json.Unmarshal(versions.Kvs[0].Value.([]byte), pact)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetPact: (%d, %d, %d, %s, %s)", pact.Id, pact.ConsumerParticipantId, pact.ProviderParticipantId, string(pact.Sha), string(pact.Content))
	return pact, nil
}

func GetPactVersion(ctx context.Context, domain string, versionID int32,
	pactID int32) (*brokerpb.PactVersion, error) {
	key := GenerateBrokerPactVersionKey(domain, versionID, pactID)
	versions, err := Store().PactVersion().Search(ctx, client.WithStrKey(key))
	if err != nil {
		return nil, err
	}
	if len(versions.Kvs) == 0 {
		return nil, nil
	}
	pactVersion := &brokerpb.PactVersion{}
	err = json.Unmarshal(versions.Kvs[0].Value.([]byte), pactVersion)
	if err != nil {
		return nil, err
	}
	PactLogger.Infof("GetPactVersion: (%d, %d, %d, %d)", pactVersion.Id, pactVersion.VersionId, pactVersion.PactId, pactVersion.ProviderParticipantId)
	return pactVersion, nil
}

func GetData(ctx context.Context, key string) (int, error) {
	values, err := Store().PactLatest().Search(ctx, client.WithStrKey(key))
	if err != nil {
		return -1, err
	}
	if len(values.Kvs) == 0 {
		return -1, nil
	}
	id, err := strconv.Atoi(string(values.Kvs[0].Value.([]byte)))
	if err != nil {
		return -1, err
	}
	return id, nil
}

func CreateParticipant(ctx context.Context, participantKey string, participant brokerpb.Participant) (*brokerpb.PublishPactResponse, error) {
	data, err := json.Marshal(participant)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "participant cannot be created."),
		}, err
	}

	_, err = client.Instance().Do(ctx, client.PUT,
		client.WithStrKey(participantKey),
		client.WithValue(data))

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "participant cannot be created."),
		}, err
	}

	k := GetBrokerLatestParticipantIDKey()
	v := strconv.Itoa(int(participant.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = client.Put(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, participant cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "participant cannot be created."),
		}, err
	}
	PactLogger.Infof("Participant created for key: %s", participantKey)
	return nil, nil
}

func CreateVersion(ctx context.Context, versionKey string,
	version brokerpb.Version) (*brokerpb.PublishPactResponse, error) {
	data, err := json.Marshal(version)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be created."),
		}, err
	}

	_, err = client.Instance().Do(ctx, client.PUT,
		client.WithStrKey(versionKey),
		client.WithValue(data))
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be created."),
		}, err
	}
	k := GetBrokerLatestVersionIDKey()
	v := strconv.Itoa(int(version.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = client.Put(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, version cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be created."),
		}, err
	}
	PactLogger.Infof("Version created for key: %s", versionKey)
	return nil, nil
}

func CreatePact(ctx context.Context,
	pactKey string, pact brokerpb.Pact) (*brokerpb.PublishPactResponse, error) {
	data, err := json.Marshal(pact)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be created."),
		}, err
	}

	_, err = client.Instance().Do(ctx,
		client.PUT,
		client.WithStrKey(pactKey),
		client.WithValue(data))

	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be created."),
		}, err
	}
	k := GetBrokerLatestPactIDKey()
	v := strconv.Itoa(int(pact.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = client.Put(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be created."),
		}, err
	}
	PactLogger.Infof("Pact created for key: %s", pactKey)
	return nil, nil
}

func CreatePactVersion(ctx context.Context, pactVersionKey string, pactVersion brokerpb.PactVersion) (*brokerpb.PublishPactResponse, error) {
	data, err := json.Marshal(pactVersion)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact version cannot be created."),
		}, err
	}

	_, err = client.Instance().Do(ctx,
		client.PUT, client.WithValue(data), client.WithStrKey(pactVersionKey))
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact version cannot be created."),
		}, err
	}
	k := GetBrokerLatestPactVersionIDKey()
	v := strconv.Itoa(int(pactVersion.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = client.Put(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "pact publish failed, pact version cannot be created.")
		return &brokerpb.PublishPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact version cannot be created."),
		}, err
	}
	PactLogger.Infof("Pact version created for key: %s", pactVersionKey)
	return nil, nil
}

func CreateVerification(ctx context.Context,
	verificationKey string, verification brokerpb.Verification) (*brokerpb.PublishVerificationResponse, error) {
	data, err := json.Marshal(verification)
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result marshal error.")
		return &brokerpb.PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "verification result marshal error."),
		}, err
	}

	_, err = client.Instance().Do(ctx, client.PUT,
		client.WithStrKey(verificationKey),
		client.WithValue(data))
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result cannot be created.")
		return &brokerpb.PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "verification result cannot be created."),
		}, err
	}
	k := GetBrokerLatestVerificationIDKey()
	v := strconv.Itoa(int(verification.Id))
	PactLogger.Infof("Inserting (%s, %s)", k, v)
	err = client.Put(ctx, k, v)
	if err != nil {
		PactLogger.Errorf(nil, "verification result publish failed, verification result cannot be created.")
		return &brokerpb.PublishVerificationResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "verification result cannot be created."),
		}, err
	}
	PactLogger.Infof("Verification result created for key: %s", verificationKey)
	return nil, nil
}

func GetLastestVersionNumberForParticipant(ctx context.Context,
	tenant string, participantID int32) int32 {
	key := util.StringJoin([]string{
		GetBrokerVersionKey(tenant), ""}, "/")
	versions, err := Store().Version().Search(ctx,
		client.WithStrKey(key),
		client.WithPrefix())

	if err != nil || len(versions.Kvs) == 0 {
		return -1
	}
	order := int32(math.MinInt32)
	for i := 0; i < len(versions.Kvs); i++ {
		version := &brokerpb.Version{}
		err = json.Unmarshal(versions.Kvs[i].Value.([]byte), &version)
		if err != nil {
			return -1
		}
		if version.ParticipantId != participantID {
			continue
		}
		if version.Order > order {
			order = version.Order
		}
	}
	return order
}

func RetrieveProviderConsumerPact(ctx context.Context,
	in *brokerpb.GetProviderConsumerVersionPactRequest) (*brokerpb.GetProviderConsumerVersionPactResponse, int32, error) {
	if in == nil || len(in.ProviderId) == 0 || len(in.ConsumerId) == 0 || len(in.Version) == 0 {
		PactLogger.Errorf(nil, "pact retrieve request failed: invalid params.")
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Request format invalid."),
		}, -1, nil
	}
	tenant := GetDefaultTenantProject()
	// Get provider microservice
	provider, err := serviceUtil.GetService(ctx, tenant, in.ProviderId)
	if err != nil {
		PactLogger.Errorf(err, "pact retrieve failed, providerId is %s: query provider failed.", in.ProviderId)
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Query provider failed."),
		}, -1, err
	}
	if provider == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, providerId is %s: provider not exist.", in.ProviderId)
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Provider does not exist."),
		}, -1, nil
	}
	// Get consumer microservice
	consumer, err := serviceUtil.GetService(ctx, tenant, in.ConsumerId)
	if err != nil {
		PactLogger.Errorf(err, "pact retrieve failed, consumerId is %s: query consumer failed.", in.ConsumerId)
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Query consumer failed."),
		}, -1, err
	}
	if consumer == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, consumerId is %s: consumer not exist.", in.ConsumerId)
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Consumer does not exist."),
		}, -1, nil
	}
	// Get provider participant
	//providerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, provider.AppId, provider.ServiceName)
	providerParticipant, err := GetParticipant(ctx, tenant, provider.AppId, provider.ServiceName)
	if err != nil || providerParticipant == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, provider participant %s cannot be searched.", in.ProviderId)
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "Provider participant cannot be searched."),
		}, -1, err
	}
	// Get consumer participant
	//consumerParticipantKey := apt.GenerateBrokerParticipantKey(tenant, consumer.AppId, consumer.ServiceName)
	consumerParticipant, err := GetParticipant(ctx, tenant, consumer.AppId, consumer.ServiceName)
	if err != nil || consumerParticipant == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, consumer participant %s cannot be searched.", in.ConsumerId)
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "consumer participant cannot be searched."),
		}, -1, err
	}
	// Get or create version
	//versionKey := apt.GenerateBrokerVersionKey(tenant, in.Version, consumerParticipant.ID)
	version, err := GetVersion(ctx, tenant, in.Version, consumerParticipant.Id)
	if err != nil || version == nil {
		PactLogger.Errorf(nil, "pact retrieve failed, version cannot be searched.")
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "version cannot be searched."),
		}, -1, err
	}
	// Get all pactversions and filter using the provider participant id
	pactVersionKey := util.StringJoin([]string{
		GetBrokerPactVersionKey(tenant),
		strconv.Itoa(int(version.Id))},
		"/")
	pactVersions, err := Store().PactVersion().Search(ctx,
		client.WithPrefix(),
		client.WithStrKey(pactVersionKey))

	if err != nil {
		return nil, -1, err
	}
	if len(pactVersions.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPact] No pact version found, sorry")
		return nil, -1, nil
	}
	pactIDs := make(map[int32]int32)
	for i := 0; i < len(pactVersions.Kvs); i++ {
		pactVersion := &brokerpb.PactVersion{}
		err = json.Unmarshal(pactVersions.Kvs[i].Value.([]byte), pactVersion)
		if err != nil {
			return nil, -1, err
		}
		// Obviously true, but checking it anyways
		if pactVersion.VersionId == version.Id {
			pactid := pactVersion.PactId
			pactIDs[pactid] = pactid
		}
	}
	if len(pactIDs) == 0 {
		PactLogger.Errorf(nil, "pact retrieve failed, pact cannot be found.")
		return &brokerpb.GetProviderConsumerVersionPactResponse{
			Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be found."),
		}, -1, err
	}
	pactKey := util.StringJoin([]string{
		GetBrokerPactKey(tenant),
		strconv.Itoa(int(consumerParticipant.Id)),
		strconv.Itoa(int(providerParticipant.Id))},
		"/")
	pacts, err := Store().Pact().Search(ctx,
		client.WithStrKey(pactKey),
		client.WithPrefix())

	if err != nil {
		return nil, -1, err
	}
	if len(pacts.Kvs) == 0 {
		PactLogger.Info("[RetrieveProviderPact] No pact version found, sorry")
		return nil, -1, nil
	}
	for i := 0; i < len(pacts.Kvs); i++ {
		pactObj := &brokerpb.Pact{}
		err = json.Unmarshal(pacts.Kvs[i].Value.([]byte), pactObj)
		if err != nil {
			return nil, -1, err
		}
		if _, ok := pactIDs[pactObj.Id]; ok {
			//PactLogger.Infof("pact retrieve succeeded, found pact: %s", string(pactObj.Content))
			return &brokerpb.GetProviderConsumerVersionPactResponse{
				Response: pb.CreateResponse(pb.ResponseSuccess, "pact found."),
				Pact:     pactObj.Content,
			}, pactObj.Id, nil
		}
	}
	PactLogger.Errorf(nil, "pact retrieve failed, pact cannot be found.")
	return &brokerpb.GetProviderConsumerVersionPactResponse{
		Response: pb.CreateResponse(scerr.ErrInternal, "pact cannot be found."),
	}, -1, nil
}

func DeletePactData(ctx context.Context,
	in *brokerpb.BaseBrokerRequest) (*pb.Response, error) {
	//tenant := util.ParseTenantProject(ctx)
	allPactKey := GetBrokerRootKey() //GetBrokerVerificationKey("default") //util.StringJoin([]string{ apt.GetRootKey(), apt.REGISTRY_PACT_ROOT_KEY }, "/")

	_, err := client.Instance().Do(ctx,
		client.DEL, client.WithStrKey(allPactKey), client.WithPrefix())
	if err != nil {
		return pb.CreateResponse(scerr.ErrInternal, "error deleting pacts."), err
	}
	return pb.CreateResponse(pb.ResponseSuccess, "deleting pacts Succeed."), nil
}
