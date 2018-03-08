#!/usr/bin/env bash

# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "${green}Building Service-center${reset}"

go build -o service-center github.com/apache/incubator-servicecomb-service-center
if [ $? == 0 ]; then
	echo "${green}Service-center built successfully..${reset}"
else
	echo "${red}Service-center build failed, please check the logs for more details.${reset}"
	exit 1
fi

echo "${green}Starting etcd in docker${reset}"
set +e
docker rm -f etcd
kill -9 $(ps aux | grep 'service-center' | awk '{print $2}')
set -e
docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 40010:40010 -p 23800:23800 -p 2379:2379 --name etcd quay.io/coreos/etcd etcd -name etcd0 -advertise-client-urls http://127.0.0.1:2379,http://127.0.0.1:40010 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:40010 -initial-advertise-peer-urls http://127.0.0.1:23800 -listen-peer-urls http://0.0.0.0:23800 -initial-cluster-token etcd-cluster-1 -initial-cluster etcd0=http://127.0.0.1:23800 -initial-cluster-state new
while ! nc -z 127.0.0.1 2379; do
  echo "Waiting Etcd to launch on 2379..."
  sleep 1
done
echo "${green}Etcd is running......${reset}"

echo "${green}Running the service-center for IT....${reset}"

cp -r etc/conf conf
./service-center > start-sc.log 2>&1 &

if [ $? == 0 ]; then
	echo "${green}Service-center is running now..${reset}"
else
	echo "${red}Service-center has some errors while running, please check the logs for more details.${reset}"
	exit 1
fi

echo "${green}Starting the integration test now....${reset}"

cd integration
go test -v

if [ $? == 0 ]; then
	echo "${green}All the integration test passed..${reset}"
else
	echo "${red}Some or all the integration test failed..please check the logs for more details.${reset}"
	set +e
	docker rm -f etcd
	kill -9 $(ps aux | grep 'service-center' | awk '{print $2}')
	set -e
	exit 1
fi
echo "Cleaning the env"
set +e
docker rm -f etcd
kill -9 $(ps aux | grep 'service-center' | awk '{print $2}')
set -e
