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

package servicecenter

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config of servicecenter
type Config struct {
	Service  *MicroService `yaml:"service"`
	Registry *Registry     `yaml:"registry"`
	Provider *MicroService `yaml:"provider"`
	Tenant   *Tenant       `yaml:"tenant"`
}

// MicroService configuration
type MicroService struct {
	ID       string    `yaml:"-"`
	AppID    string    `yaml:"appId"`
	Name     string    `yaml:"name"`
	Version  string    `yaml:"version"`
	Instance *Instance `yaml:"instance"`
}

// Instance configuration
type Instance struct {
	ID            string `yaml:"-"`
	Hostname      string `yaml:"hostname"`
	Protocol      string `yaml:"protocol"`
	ListenAddress string `yaml:"listenAddress"`
}

// Registry configuration
type Registry struct {
	Address   string   `yaml:"address"`
	Endpoints []string `yaml:"-"`
	MongoDB   *MongoDB `yaml:"mongo"`
}

type MongoDB struct {
	Cluster Cluster `yaml:"cluster"`
}

type Cluster struct {
	Uri string `yaml:"uri"`
}

// Tenant configuration
type Tenant struct {
	Domain  string `yaml:"domain"`
	Project string `yaml:"project"`
}

func DefaultConfig() *Config {
	return &Config{
		Service: &MicroService{
			AppID:   "default",
			Name:    "demo",
			Version: "0.0.1",
			Instance: &Instance{
				Protocol:      "rest",
				ListenAddress: "127.0.0.1:8080",
			},
		},
		Registry: &Registry{
			Address: "http://127.0.0.1:30100",
		},
		Tenant: &Tenant{
			Domain:  "default",
			Project: "default",
		},
	}
}

// Verify Provide config verification
func (c *Config) Verify() error {
	if c.Service == nil {
		return errors.New("microservice is empty")
	}

	if c.Service.Instance == nil && c.Provider != nil {
		return errors.New("microservice must be a consumer or provider")
	}

	if c.Tenant == nil {
		c.Tenant = &Tenant{}
	}

	if c.Tenant.Domain == "" {
		c.Tenant.Domain = "default"
	}

	if c.Tenant.Project == "" {
		c.Tenant.Project = "default"
	}

	if c.Service.Instance != nil {
		if c.Service.Instance.Hostname == "" {
			c.Service.Instance.Hostname, _ = os.Hostname()
		}

		if c.Service.Instance.ListenAddress == "" {
			return fmt.Errorf("instance lister address is empty")
		}

		host, port, err := net.SplitHostPort(c.Service.Instance.ListenAddress)
		if err != nil {
			return fmt.Errorf("instance lister address is wrong: %s", err)
		}
		if host == "" {
			host = "127.0.0.1"
		}
		num, err := strconv.Atoi(port)
		if err != nil || num <= 0 {
			return fmt.Errorf("instance lister port %s is wrong: %s", port, err)
		}
		c.Service.Instance.ListenAddress = host + ":" + port
	}

	if c.Registry == nil || c.Registry.Address == "" {
		return errors.New("registry is empty")
	}

	c.Registry.Endpoints = strings.Split(c.Registry.Address, ",")
	for i := 0; i < len(c.Registry.Endpoints); i++ {
		_, err := url.Parse(c.Registry.Endpoints[i])
		if err != nil {
			return fmt.Errorf("parse registry address faild: %s", err)
		}
	}
	return nil
}
