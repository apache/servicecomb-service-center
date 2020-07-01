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

	"github.com/apache/servicecomb-service-center/pkg/util"
)

const (
	BrokerRootKey             = "cse-pact"
	BrokerParticipantKey      = "participant"
	BrokerVersionKey          = "version"
	BrokerPactKey             = "pact"
	BrokerPactVersionKey      = "pact-version"
	BrokerPactTagKey          = "pact-tag"
	BrokerPactVerificationKey = "verification"
	BrokerPactLatest          = "latest"
)

// GetBrokerRootKey returns url (/cse-pact)
func GetBrokerRootKey() string {
	return util.StringJoin([]string{
		"",
		BrokerRootKey,
	}, "/")
}

//GetBrokerLatestKey returns  pact related keys
func GetBrokerLatestKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerPactLatest,
		tenant,
	}, "/")
}

//GetBrokerParticipantKey returns the participant root key
func GetBrokerParticipantKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerParticipantKey,
		tenant,
	}, "/")
}

//GenerateBrokerParticipantKey returns the participant key
func GenerateBrokerParticipantKey(tenant string, appID string, serviceName string) string {
	return util.StringJoin([]string{
		GetBrokerParticipantKey(tenant),
		appID,
		serviceName,
	}, "/")
}

//GetBrokerVersionKey returns th root version key
func GetBrokerVersionKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerVersionKey,
		tenant,
	}, "/")
}

//GenerateBrokerVersionKey returns the version key
func GenerateBrokerVersionKey(tenant string, number string, participantID int32) string {
	return util.StringJoin([]string{
		GetBrokerVersionKey(tenant),
		number,
		strconv.Itoa(int(participantID)),
	}, "/")
}

//GetBrokerPactKey returns the pact root key
func GetBrokerPactKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerPactKey,
		tenant,
	}, "/")
}

//GenerateBrokerPactKey returns the pact key
func GenerateBrokerPactKey(tenant string, consumerParticipantID int32,
	providerParticipantID int32, sha []byte) string {
	return util.StringJoin([]string{
		GetBrokerPactKey(tenant),
		strconv.Itoa(int(consumerParticipantID)),
		strconv.Itoa(int(providerParticipantID)),
		string(sha),
	}, "/")
}

//GetBrokerPactVersionKey returns the pact version root key
func GetBrokerPactVersionKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerPactVersionKey,
		tenant,
	}, "/")
}

//GenerateBrokerPactVersionKey returns the pact version root key
func GenerateBrokerPactVersionKey(tenant string, versionID int32, pactID int32) string {
	return util.StringJoin([]string{
		GetBrokerPactVersionKey(tenant),
		strconv.Itoa(int(versionID)),
		strconv.Itoa(int(pactID)),
	}, "/")
}

//GetBrokerTagKey returns the broker tag root key
func GetBrokerTagKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerPactTagKey,
		tenant,
	}, "/")
}

//GenerateBrokerTagKey returns the broker tag key
func GenerateBrokerTagKey(tenant string, versionID int32) string {
	return util.StringJoin([]string{
		GetBrokerTagKey(tenant),
		strconv.Itoa(int(versionID)),
	}, "/")
}

//GetBrokerVerificationKey returns the verification root key
func GetBrokerVerificationKey(tenant string) string {
	return util.StringJoin([]string{
		GetBrokerRootKey(),
		BrokerPactVerificationKey,
		tenant,
	}, "/")
}

//GenerateBrokerVerificationKey returns he verification key
func GenerateBrokerVerificationKey(tenant string, pactVersionID int32, number int32) string {
	return util.StringJoin([]string{
		GetBrokerVerificationKey(tenant),
		strconv.Itoa(int(pactVersionID)),
		strconv.Itoa(int(number)),
	}, "/")
}

//GetBrokerLatestParticipantIDKey returns the latest participant ID
func GetBrokerLatestParticipantIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BrokerParticipantKey,
	}, "/")
}

//GetBrokerLatestVersionIDKey returns latest version ID
func GetBrokerLatestVersionIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BrokerVersionKey,
	}, "/")
}

//GetBrokerLatestPactIDKey returns latest pact ID
func GetBrokerLatestPactIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BrokerPactKey,
	}, "/")
}

//GetBrokerLatestPactVersionIDKey returns lated pact version ID
func GetBrokerLatestPactVersionIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BrokerPactVersionKey,
	}, "/")
}

//GetBrokerLatestVerificationIDKey returns the lastest verification ID
func GetBrokerLatestVerificationIDKey() string {
	return util.StringJoin([]string{
		GetBrokerLatestKey("default"),
		BrokerPactVerificationKey,
	}, "/")
}
