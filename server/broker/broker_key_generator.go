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
	"strconv"

	"github.com/apache/incubator-servicecomb-service-center/pkg/util"
)

const (
	BROKER_ROOT_KEY              = "cse-pact"
	BROKER_PARTICIPANT_KEY       = "participant"
	BROKER_VERSION_KEY           = "version"
	BROKER_PACT_KEY              = "pact"
	BROKER_PACT_VERSION_KEY      = "pact-version"
	BROKER_PACT_TAG_KEY          = "pact-tag"
	BROKER_PACT_VERIFICATION_KEY = "verification"
	BROKER_PACT_LATEST           = "latest"
)

// GetBrokerRootKey returns url (/cse-pact)
func GetBrokerRootKey() string {
	return util.StringJoin([]string{
		"",
		BROKER_ROOT_KEY,
	}, "/")
}

//GetBrokerLatestKey returns  pact related keys
func GetBrokerLatestKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_PACT_LATEST,
		tenant,
	}, "/")
}

//GetBrokerParticipantKey returns the participant root key
func GetBrokerParticipantKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_PARTICIPANT_KEY,
		tenant,
	}, "/")
}

//GenerateBrokerParticipantKey returns the participant key
func GenerateBrokerParticipantKey(tenant string, appId string, serviceName string) string {
	return util.StringJoin([]string{
		GetBrokerParticipantKey(tenant),
		appId,
		serviceName,
	}, "/")
}

//GetBrokerVersionKey returns th root version key
func GetBrokerVersionKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_VERSION_KEY,
		tenant,
	}, "/")
}

//GenerateBrokerVersionKey returns the version key
func GenerateBrokerVersionKey(tenant string, number string, participantId int32) string {
	return util.StringJoin([]string{
		GetBrokerVersionKey(tenant),
		number,
		strconv.Itoa(int(participantId)),
	}, "/")
}

//GetBrokerPactKey returns the pact root key
func GetBrokerPactKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_PACT_KEY,
		tenant,
	}, "/")
}

//GenerateBrokerPactKey returns the pact key
func GenerateBrokerPactKey(tenant string, consumerParticipantId int32,
	providerParticipantId int32, sha []byte) string {
	return util.StringJoin([]string{
		GetBrokerPactKey(tenant),
		strconv.Itoa(int(consumerParticipantId)),
		strconv.Itoa(int(providerParticipantId)),
		string(sha),
	}, "/")
}

//GetBrokerPactVersionKey returns the pact version root key
func GetBrokerPactVersionKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_PACT_VERSION_KEY,
		tenant,
	}, "/")
}

//GenerateBrokerPactVersionKey returns the pact version root key
func GenerateBrokerPactVersionKey(tenant string, versionId int32, pactId int32) string {
	return util.StringJoin([]string{
		GetBrokerPactVersionKey(tenant),
		strconv.Itoa(int(versionId)),
		strconv.Itoa(int(pactId)),
	}, "/")
}

//GetBrokerTagKey returns the broker tag root key
func GetBrokerTagKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_PACT_TAG_KEY,
		tenant,
	}, "/")
}

//GenerateBrokerTagKey returns the broker tag key
func GenerateBrokerTagKey(tenant string, versionId int32) string {
	return util.StringJoin([]string{
		GetBrokerTagKey(tenant),
		strconv.Itoa(int(versionId)),
	}, "/")
}

//GetBrokerVerificationKey returns the verification root key
func GetBrokerVerificationKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BROKER_PACT_VERIFICATION_KEY,
		tenant,
	}, "/")
}

//GenerateBrokerVerificationKey returns he verification key
func GenerateBrokerVerificationKey(tenant string, pactVersionId int32, number int32) string {
	return util.StringJoin([]string{
		GetBrokerVerificationKey(tenant),
		strconv.Itoa(int(pactVersionId)),
		strconv.Itoa(int(number)),
	}, "/")
}

//GetBrokerLatestParticipantIDKey returns the latest participant ID
func GetBrokerLatestParticipantIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BROKER_PARTICIPANT_KEY,
	}, "/")
}

//GetBrokerLatestVersionIDKey returns latest version ID
func GetBrokerLatestVersionIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BROKER_VERSION_KEY,
	}, "/")
}

//GetBrokerLatestPactIDKey returns latest pact ID
func GetBrokerLatestPactIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BROKER_PACT_KEY,
	}, "/")
}

//GetBrokerLatestPactVersionIDKey returns lated pact version ID
func GetBrokerLatestPactVersionIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BROKER_PACT_VERSION_KEY,
	}, "/")
}

//GetBrokerLatestVerificationIDKey returns the lastest verification ID
func GetBrokerLatestVerificationIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BROKER_PACT_VERIFICATION_KEY,
	}, "/")
}
