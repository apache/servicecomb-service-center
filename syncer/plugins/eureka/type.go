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

package eureka

import (
	"encoding/json"
	"strconv"
)

const (
	UNKNOWN      = "UNKNOWN"
	UP           = "UP"
	DOWN         = "DOWN"
	STARTING     = "STARTING"
	OUTOFSERVICE = "OUTOFSERVICE"
)

type eurekaData struct {
	APPS *Applications `json:"applications"`
}

type Applications struct {
	VersionsDelta string         `json:"versions__delta"`
	AppsHashcode  string         `json:"apps__hashcode"`
	Applications  []*Application `json:"application,omitempty"`
}

type Application struct {
	Name      string      `json:"name"`
	Instances []*Instance `json:"instance"`
}

type InstanceRequest struct {
	Instance *Instance `json:"instance"`
}

type Instance struct {
	InstanceId                    string          `json:"instanceId"`
	HostName                      string          `json:"hostName"`
	APP                           string          `json:"app"`
	IPAddr                        string          `json:"ipAddr"`
	Status                        string          `json:"status"`
	OverriddenStatus              string          `json:"overriddenStatus,omitempty"`
	Port                          *Port           `json:"port,omitempty"`
	SecurePort                    *Port           `json:"securePort,omitempty"`
	CountryId                     int             `json:"countryId,omitempty"`
	DataCenterInfo                *DataCenterInfo `json:"dataCenterInfo"`
	LeaseInfo                     *LeaseInfo      `json:"leaseInfo,omitempty"`
	Metadata                      *MetaData       `json:"metadata,omitempty"`
	HomePageUrl                   string          `json:"homePageUrl,omitempty"`
	StatusPageUrl                 string          `json:"statusPageUrl,omitempty"`
	HealthCheckUrl                string          `json:"healthCheckUrl,omitempty"`
	VipAddress                    string          `json:"vipAddress,omitempty"`
	SecureVipAddress              string          `json:"secureVipAddress,omitempty"`
	IsCoordinatingDiscoveryServer BoolString      `json:"isCoordinatingDiscoveryServer,omitempty"`
	LastUpdatedTimestamp          string          `json:"lastUpdatedTimestamp,omitempty"`
	LastDirtyTimestamp            string          `json:"lastDirtyTimestamp,omitempty"`
	ActionType                    string          `json:"actionType,omitempty"`
}

type DataCenterInfo struct {
	Name     string              `json:"name"`
	Class    string              `json:"@class"`
	Metadata *DataCenterMetadata `json:"metadata,omitempty"`
}

type DataCenterMetadata struct {
	AmiLaunchIndex   string `json:"ami-launch-index,omitempty"`
	LocalHostname    string `json:"local-hostname,omitempty"`
	AvailabilityZone string `json:"availability-zone,omitempty"`
	InstanceId       string `json:"instance-id,omitempty"`
	PublicIpv4       string `json:"public-ipv4,omitempty"`
	PublicHostname   string `json:"public-hostname,omitempty"`
	AmiManifestPath  string `json:"ami-manifest-path,omitempty"`
	LocalIpv4        string `json:"local-ipv4,omitempty"`
	Hostname         string `json:"hostname,omitempty"`
	AmiId            string `json:"ami-id,omitempty"`
	InstanceType     string `json:"instance-type,omitempty"`
}

type LeaseInfo struct {
	RenewalIntervalInSecs  int  `json:"renewalIntervalInSecs,omitempty"`
	DurationInSecs         int  `json:"durationInSecs,omitempty"`
	RegistrationTimestamp  int  `json:"registrationTimestamp,omitempty"`
	LastRenewalTimestamp   int  `json:"lastRenewalTimestamp,omitempty"`
	EvictionDurationInSecs uint `json:"evictionDurationInSecs,omitempty"`
	EvictionTimestamp      int  `json:"evictionTimestamp,omitempty"`
	ServiceUpTimestamp     int  `json:"serviceUpTimestamp,omitempty"`
}

type Port struct {
	Port    int        `json:"$"`
	Enabled BoolString `json:"@enabled"`
}

type MetaData struct {
	Map   map[string]string
	Class string
}

func (s *MetaData) MarshalJSON() ([]byte, error) {
	newMap := make(map[string]string)
	for key, value := range s.Map {
		newMap[key] = value
	}
	if s.Class != "" {
		newMap["@class"] = s.Class
	}
	return json.Marshal(&newMap)
}

func (s *MetaData) UnmarshalJSON(data []byte) error {
	newMap := make(map[string]string)
	err := json.Unmarshal(data, &newMap)
	if err != nil {
		return err
	}

	s.Map = newMap
	if val, ok := s.Map["@class"]; ok {
		s.Class = val
		delete(s.Map, "@class")
	}
	return nil
}

type BoolString string

func (b *BoolString) Set(value bool) {
	str := strconv.FormatBool(value)
	*b = BoolString(str)
}

func (b BoolString) Bool() bool {
	enabled, err := strconv.ParseBool(string(b))
	if err != nil {
		return false
	}
	return enabled
}
