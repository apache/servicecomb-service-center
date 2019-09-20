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

package main

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"syscall"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/pkg/syssig"
	"github.com/apache/servicecomb-service-center/syncer/samples/multi-servicecenters/servicecenter/hello-server/servicecenter"
	"gopkg.in/yaml.v2"
)

func main() {
	// 加载配置文件
	conf, err := loadConfig("./conf/microservice.yaml")
	if err != nil {
		log.Fatal("load config file faild: %s", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopCh := make(chan struct{})
	err = syssig.AddSignalsHandler(func() {
		log.Info("close stop chanel")
		close(stopCh)
	}, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	if err != nil {
		log.Fatal("listen system signal failed", err)
		return
	}
	go syssig.Run(ctx)
	go start(ctx, conf)
	defer servicecenter.Stop(ctx)

	<-stopCh
}

func start(ctx context.Context, conf *servicecenter.Config)  {
	err := servicecenter.Start(ctx, conf)
	if err != nil {
		log.Fatal("register and discovery from servicecenter failed", err)
		return
	}

	if conf.Provider != nil {
		// check health
		msg, err := checkHealth(conf.Provider.Name)
		if err != nil {
			log.Error("call provider failed", err)
			return
		}
		log.Infof("call provider success, repay is '%s'", msg)
	}

	if conf.Service.Instance != nil {
		http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			log.Info("request from user")

			msg, err := login(r, conf.Provider.Name)
			if err != nil {
				log.Error("call provider failed", err)
				return
			}

			log.Info("call provider success")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(msg))
		})
		go func() {
			err = http.ListenAndServe(conf.Service.Instance.ListenAddress, nil)
			if err != nil {
				log.Fatal("start service failed", err)
			}
		}()
	}
}

func checkHealth(providerName string) (string, error) {
	return callProvider(context.Background(), http.MethodGet, "http://"+providerName+"/checkHealth", nil)
}

func login(r *http.Request, providerName string) (string, error) {
	bytes, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	return callProvider(r.Context(), http.MethodPost, "http://"+providerName+"/login", bytes)
}

func callProvider(ctx context.Context, method, addr string, body []byte) (string, error) {
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	resp, err := servicecenter.Do(ctx, method, addr, header, body)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(data))
	}
	return string(data), nil
}

func loadConfig(filePath string) (*servicecenter.Config, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	conf := servicecenter.DefaultConfig()

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		return nil, err
	}
	return conf, conf.Verify()
}
