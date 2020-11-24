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

package brokerpb

import (
	"github.com/go-chassis/cari/discovery"
)

type Participant struct {
	Id          int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	AppId       string `protobuf:"bytes,2,opt,name=appId" json:"appId,omitempty"`
	ServiceName string `protobuf:"bytes,3,opt,name=serviceName" json:"serviceName,omitempty"`
}

func (m *Participant) Reset() { *m = Participant{} }

func (m *Participant) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Participant) GetAppId() string {
	if m != nil {
		return m.AppId
	}
	return ""
}

func (m *Participant) GetServiceName() string {
	if m != nil {
		return m.ServiceName
	}
	return ""
}

type Version struct {
	Id            int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Number        string `protobuf:"bytes,2,opt,name=number" json:"number,omitempty"`
	ParticipantId int32  `protobuf:"varint,3,opt,name=participantId" json:"participantId,omitempty"`
	Order         int32  `protobuf:"varint,4,opt,name=order" json:"order,omitempty"`
}

func (m *Version) Reset() { *m = Version{} }

func (m *Version) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Version) GetNumber() string {
	if m != nil {
		return m.Number
	}
	return ""
}

func (m *Version) GetParticipantId() int32 {
	if m != nil {
		return m.ParticipantId
	}
	return 0
}

func (m *Version) GetOrder() int32 {
	if m != nil {
		return m.Order
	}
	return 0
}

type Pact struct {
	Id                    int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	ConsumerParticipantId int32  `protobuf:"varint,2,opt,name=consumerParticipantId" json:"consumerParticipantId,omitempty"`
	ProviderParticipantId int32  `protobuf:"varint,3,opt,name=providerParticipantId" json:"providerParticipantId,omitempty"`
	Sha                   []byte `protobuf:"bytes,4,opt,name=sha,proto3" json:"sha,omitempty"`
	Content               []byte `protobuf:"bytes,5,opt,name=content,proto3" json:"content,omitempty"`
}

func (m *Pact) Reset() { *m = Pact{} }

func (m *Pact) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Pact) GetConsumerParticipantId() int32 {
	if m != nil {
		return m.ConsumerParticipantId
	}
	return 0
}

func (m *Pact) GetProviderParticipantId() int32 {
	if m != nil {
		return m.ProviderParticipantId
	}
	return 0
}

func (m *Pact) GetSha() []byte {
	if m != nil {
		return m.Sha
	}
	return nil
}

func (m *Pact) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

type PactVersion struct {
	Id                    int32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	VersionId             int32 `protobuf:"varint,2,opt,name=versionId" json:"versionId,omitempty"`
	PactId                int32 `protobuf:"varint,3,opt,name=pactId" json:"pactId,omitempty"`
	ProviderParticipantId int32 `protobuf:"varint,4,opt,name=providerParticipantId" json:"providerParticipantId,omitempty"`
}

func (m *PactVersion) Reset() { *m = PactVersion{} }

func (m *PactVersion) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *PactVersion) GetVersionId() int32 {
	if m != nil {
		return m.VersionId
	}
	return 0
}

func (m *PactVersion) GetPactId() int32 {
	if m != nil {
		return m.PactId
	}
	return 0
}

func (m *PactVersion) GetProviderParticipantId() int32 {
	if m != nil {
		return m.ProviderParticipantId
	}
	return 0
}

type Tag struct {
	Name      string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	VersionId int32  `protobuf:"varint,2,opt,name=versionId" json:"versionId,omitempty"`
}

func (m *Tag) Reset() { *m = Tag{} }

func (m *Tag) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Tag) GetVersionId() int32 {
	if m != nil {
		return m.VersionId
	}
	return 0
}

type PublishPactRequest struct {
	ProviderId string `protobuf:"bytes,1,opt,name=providerId" json:"providerId,omitempty"`
	ConsumerId string `protobuf:"bytes,2,opt,name=consumerId" json:"consumerId,omitempty"`
	Version    string `protobuf:"bytes,3,opt,name=version" json:"version,omitempty"`
	Pact       []byte `protobuf:"bytes,4,opt,name=pact,proto3" json:"pact,omitempty"`
}

func (m *PublishPactRequest) Reset() { *m = PublishPactRequest{} }

func (m *PublishPactRequest) GetProviderId() string {
	if m != nil {
		return m.ProviderId
	}
	return ""
}

func (m *PublishPactRequest) GetConsumerId() string {
	if m != nil {
		return m.ConsumerId
	}
	return ""
}

func (m *PublishPactRequest) GetPact() []byte {
	if m != nil {
		return m.Pact
	}
	return nil
}

type PublishPactResponse struct {
	Response *discovery.Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
}

func (m *PublishPactResponse) Reset() { *m = PublishPactResponse{} }

type GetAllProviderPactsRequest struct {
	ProviderId string             `protobuf:"bytes,1,opt,name=providerId" json:"providerId,omitempty"`
	BaseUrl    *BaseBrokerRequest `protobuf:"bytes,2,opt,name=baseUrl" json:"baseUrl,omitempty"`
}

func (m *GetAllProviderPactsRequest) Reset() { *m = GetAllProviderPactsRequest{} }

func (m *GetAllProviderPactsRequest) GetProviderId() string {
	if m != nil {
		return m.ProviderId
	}
	return ""
}

func (m *GetAllProviderPactsRequest) GetBaseUrl() *BaseBrokerRequest {
	if m != nil {
		return m.BaseUrl
	}
	return nil
}

type ConsumerInfo struct {
	Href string `protobuf:"bytes,1,opt,name=href" json:"href,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
}

func (m *ConsumerInfo) GetHref() string {
	if m != nil {
		return m.Href
	}
	return ""
}

func (m *ConsumerInfo) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type Links struct {
	Pacts []*ConsumerInfo `protobuf:"bytes,1,rep,name=pacts" json:"pacts,omitempty"`
}

func (m *Links) Reset() { *m = Links{} }

func (m *Links) GetPacts() []*ConsumerInfo {
	if m != nil {
		return m.Pacts
	}
	return nil
}

type GetAllProviderPactsResponse struct {
	Response *discovery.Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	XLinks   *Links              `protobuf:"bytes,2,opt,name=_links,json=Links" json:"_links,omitempty"`
}

func (m *GetAllProviderPactsResponse) Reset() { *m = GetAllProviderPactsResponse{} }

func (m *GetAllProviderPactsResponse) GetXLinks() *Links {
	if m != nil {
		return m.XLinks
	}
	return nil
}

type GetProviderConsumerVersionPactRequest struct {
	ProviderId string             `protobuf:"bytes,1,opt,name=providerId" json:"providerId,omitempty"`
	ConsumerId string             `protobuf:"bytes,2,opt,name=consumerId" json:"consumerId,omitempty"`
	Version    string             `protobuf:"bytes,3,opt,name=version" json:"version,omitempty"`
	BaseUrl    *BaseBrokerRequest `protobuf:"bytes,4,opt,name=baseUrl" json:"baseUrl,omitempty"`
}

func (m *GetProviderConsumerVersionPactRequest) Reset() { *m = GetProviderConsumerVersionPactRequest{} }

func (m *GetProviderConsumerVersionPactRequest) GetProviderId() string {
	if m != nil {
		return m.ProviderId
	}
	return ""
}

func (m *GetProviderConsumerVersionPactRequest) GetConsumerId() string {
	if m != nil {
		return m.ConsumerId
	}
	return ""
}

func (m *GetProviderConsumerVersionPactRequest) GetBaseUrl() *BaseBrokerRequest {
	if m != nil {
		return m.BaseUrl
	}
	return nil
}

type GetProviderConsumerVersionPactResponse struct {
	Response *discovery.Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Pact     []byte              `protobuf:"bytes,2,opt,name=pact,proto3" json:"pact,omitempty"`
}

func (m *GetProviderConsumerVersionPactResponse) Reset() {
	*m = GetProviderConsumerVersionPactResponse{}
}

func (m *GetProviderConsumerVersionPactResponse) GetPact() []byte {
	if m != nil {
		return m.Pact
	}
	return nil
}

type Verification struct {
	Id               int32  `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Number           int32  `protobuf:"varint,2,opt,name=number" json:"number,omitempty"`
	PactVersionId    int32  `protobuf:"varint,3,opt,name=pactVersionId" json:"pactVersionId,omitempty"`
	Success          bool   `protobuf:"varint,4,opt,name=success" json:"success,omitempty"`
	ProviderVersion  string `protobuf:"bytes,5,opt,name=providerVersion" json:"providerVersion,omitempty"`
	BuildUrl         string `protobuf:"bytes,6,opt,name=buildUrl" json:"buildUrl,omitempty"`
	VerificationDate string `protobuf:"bytes,7,opt,name=verificationDate" json:"verificationDate,omitempty"`
}

func (m *Verification) Reset() { *m = Verification{} }

func (m *Verification) GetId() int32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Verification) GetNumber() int32 {
	if m != nil {
		return m.Number
	}
	return 0
}

func (m *Verification) GetPactVersionId() int32 {
	if m != nil {
		return m.PactVersionId
	}
	return 0
}

func (m *Verification) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *Verification) GetProviderVersion() string {
	if m != nil {
		return m.ProviderVersion
	}
	return ""
}

func (m *Verification) GetBuildUrl() string {
	if m != nil {
		return m.BuildUrl
	}
	return ""
}

func (m *Verification) GetVerificationDate() string {
	if m != nil {
		return m.VerificationDate
	}
	return ""
}

type VerificationSummary struct {
	Successful []string `protobuf:"bytes,1,rep,name=successful" json:"successful,omitempty"`
	Failed     []string `protobuf:"bytes,2,rep,name=failed" json:"failed,omitempty"`
	Unknown    []string `protobuf:"bytes,3,rep,name=unknown" json:"unknown,omitempty"`
}

func (m *VerificationSummary) Reset() { *m = VerificationSummary{} }

func (m *VerificationSummary) GetSuccessful() []string {
	if m != nil {
		return m.Successful
	}
	return nil
}

func (m *VerificationSummary) GetFailed() []string {
	if m != nil {
		return m.Failed
	}
	return nil
}

func (m *VerificationSummary) GetUnknown() []string {
	if m != nil {
		return m.Unknown
	}
	return nil
}

type VerificationDetail struct {
	ProviderName               string `protobuf:"bytes,1,opt,name=providerName" json:"providerName,omitempty"`
	ProviderApplicationVersion string `protobuf:"bytes,2,opt,name=providerApplicationVersion" json:"providerApplicationVersion,omitempty"`
	Success                    bool   `protobuf:"varint,3,opt,name=success" json:"success,omitempty"`
	VerificationDate           string `protobuf:"bytes,4,opt,name=verificationDate" json:"verificationDate,omitempty"`
}

func (m *VerificationDetail) Reset() { *m = VerificationDetail{} }

func (m *VerificationDetail) GetProviderName() string {
	if m != nil {
		return m.ProviderName
	}
	return ""
}

func (m *VerificationDetail) GetProviderApplicationVersion() string {
	if m != nil {
		return m.ProviderApplicationVersion
	}
	return ""
}

func (m *VerificationDetail) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *VerificationDetail) GetVerificationDate() string {
	if m != nil {
		return m.VerificationDate
	}
	return ""
}

type VerificationDetails struct {
	VerificationResults []*VerificationDetail `protobuf:"bytes,1,rep,name=verificationResults" json:"verificationResults,omitempty"`
}

func (m *VerificationDetails) GetVerificationResults() []*VerificationDetail {
	if m != nil {
		return m.VerificationResults
	}
	return nil
}

type VerificationResult struct {
	Success         bool                 `protobuf:"varint,1,opt,name=success" json:"success,omitempty"`
	ProviderSummary *VerificationSummary `protobuf:"bytes,2,opt,name=providerSummary" json:"providerSummary,omitempty"`
	XEmbedded       *VerificationDetails `protobuf:"bytes,3,opt,name=_embedded,json=Embedded" json:"_embedded,omitempty"`
}

func (m *VerificationResult) Reset() { *m = VerificationResult{} }

func (m *VerificationResult) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *VerificationResult) GetProviderSummary() *VerificationSummary {
	if m != nil {
		return m.ProviderSummary
	}
	return nil
}

func (m *VerificationResult) GetXEmbedded() *VerificationDetails {
	if m != nil {
		return m.XEmbedded
	}
	return nil
}

type PublishVerificationRequest struct {
	ProviderId                 string `protobuf:"bytes,1,opt,name=providerId" json:"providerId,omitempty"`
	ConsumerId                 string `protobuf:"bytes,2,opt,name=consumerId" json:"consumerId,omitempty"`
	PactId                     int32  `protobuf:"varint,3,opt,name=pactId" json:"pactId,omitempty"`
	Success                    bool   `protobuf:"varint,4,opt,name=success" json:"success,omitempty"`
	ProviderApplicationVersion string `protobuf:"bytes,5,opt,name=providerApplicationVersion" json:"providerApplicationVersion,omitempty"`
}

func (m *PublishVerificationRequest) Reset() { *m = PublishVerificationRequest{} }

func (m *PublishVerificationRequest) GetProviderId() string {
	if m != nil {
		return m.ProviderId
	}
	return ""
}

func (m *PublishVerificationRequest) GetConsumerId() string {
	if m != nil {
		return m.ConsumerId
	}
	return ""
}

func (m *PublishVerificationRequest) GetPactId() int32 {
	if m != nil {
		return m.PactId
	}
	return 0
}

func (m *PublishVerificationRequest) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *PublishVerificationRequest) GetProviderApplicationVersion() string {
	if m != nil {
		return m.ProviderApplicationVersion
	}
	return ""
}

type PublishVerificationResponse struct {
	Response     *discovery.Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Confirmation *VerificationDetail `protobuf:"bytes,2,opt,name=confirmation" json:"confirmation,omitempty"`
}

func (m *PublishVerificationResponse) GetConfirmation() *VerificationDetail {
	if m != nil {
		return m.Confirmation
	}
	return nil
}

type RetrieveVerificationRequest struct {
	ConsumerId      string `protobuf:"bytes,1,opt,name=consumerId" json:"consumerId,omitempty"`
	ConsumerVersion string `protobuf:"bytes,2,opt,name=consumerVersion" json:"consumerVersion,omitempty"`
}

func (m *RetrieveVerificationRequest) Reset() { *m = RetrieveVerificationRequest{} }

func (m *RetrieveVerificationRequest) GetConsumerId() string {
	if m != nil {
		return m.ConsumerId
	}
	return ""
}

func (m *RetrieveVerificationRequest) GetConsumerVersion() string {
	if m != nil {
		return m.ConsumerVersion
	}
	return ""
}

type RetrieveVerificationResponse struct {
	Response *discovery.Response `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	Result   *VerificationResult `protobuf:"bytes,2,opt,name=result" json:"result,omitempty"`
}

func (m *RetrieveVerificationResponse) Reset() { *m = RetrieveVerificationResponse{} }

func (m *RetrieveVerificationResponse) GetResult() *VerificationResult {
	if m != nil {
		return m.Result
	}
	return nil
}

type BaseBrokerRequest struct {
	HostAddress string `protobuf:"bytes,1,opt,name=hostAddress" json:"hostAddress,omitempty"`
	Scheme      string `protobuf:"bytes,2,opt,name=scheme" json:"scheme,omitempty"`
}

func (m *BaseBrokerRequest) Reset() { *m = BaseBrokerRequest{} }

func (m *BaseBrokerRequest) GetHostAddress() string {
	if m != nil {
		return m.HostAddress
	}
	return ""
}

func (m *BaseBrokerRequest) GetScheme() string {
	if m != nil {
		return m.Scheme
	}
	return ""
}

type BrokerAPIInfoEntry struct {
	Href      string `protobuf:"bytes,1,opt,name=href" json:"href,omitempty"`
	Name      string `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	Title     string `protobuf:"bytes,3,opt,name=title" json:"title,omitempty"`
	Templated bool   `protobuf:"varint,4,opt,name=templated" json:"templated,omitempty"`
}

func (m *BrokerAPIInfoEntry) GetHref() string {
	if m != nil {
		return m.Href
	}
	return ""
}

func (m *BrokerAPIInfoEntry) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *BrokerAPIInfoEntry) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *BrokerAPIInfoEntry) GetTemplated() bool {
	if m != nil {
		return m.Templated
	}
	return false
}

type BrokerHomeResponse struct {
	Response *discovery.Response            `protobuf:"bytes,1,opt,name=response" json:"response,omitempty"`
	XLinks   map[string]*BrokerAPIInfoEntry `protobuf:"bytes,2,rep,name=_links,json=Links" json:"_links,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	Curies   []*BrokerAPIInfoEntry          `protobuf:"bytes,3,rep,name=curies" json:"curies,omitempty"`
}

func (m *BrokerHomeResponse) Reset() { *m = BrokerHomeResponse{} }

func (m *BrokerHomeResponse) GetXLinks() map[string]*BrokerAPIInfoEntry {
	if m != nil {
		return m.XLinks
	}
	return nil
}

func (m *BrokerHomeResponse) GetCuries() []*BrokerAPIInfoEntry {
	if m != nil {
		return m.Curies
	}
	return nil
}
