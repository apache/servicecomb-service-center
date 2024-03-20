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

set -x
set +e
db_name=$1
docker rm -f "$db_name"
set -e

ut_for_dir() {
    local name=$1
    echo "${green}Running  UT for Service-Center, skip package $name${reset}"
    bash -x ./scripts/ut.sh "$name"
}

echo "${green}Starting Unit Testing for Service Center${reset}"

if [ "${db_name}" == "etcd" ];then
  echo "${green}Starting etcd in docker${reset}"
  docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 40010:40010 -p 23800:23800 -p 2379:2379 --name etcd quay.io/coreos/etcd etcd -name etcd0 -advertise-client-urls http://127.0.0.1:2379,http://127.0.0.1:40010 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:40010 -initial-advertise-peer-urls http://127.0.0.1:23800 -listen-peer-urls http://0.0.0.0:23800 -initial-cluster-token etcd-cluster-1 -initial-cluster etcd0=http://127.0.0.1:23800 -initial-cluster-state new
  while ! nc -z 127.0.0.1 2379; do
    echo "Waiting Etcd to launch on 2379..."
    sleep 1
  done
  echo "${green}Etcd is running......${reset}"
elif [ ${db_name} == "mongo" ];then
  while ! nc -z 127.0.0.1 27017; do
    echo "Waiting mongo to launch on 27017..."
    sleep 1
  done
  echo "${green}mongodb is running......${reset}"
elif [ ${db_name} == "local" ];then
  echo "${green}Starting etcd in docker${reset}"
  docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 40010:40010 -p 23800:23800 -p 2379:2379 --name etcd quay.io/coreos/etcd etcd -name etcd0 -advertise-client-urls http://127.0.0.1:2379,http://127.0.0.1:40010 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:40010 -initial-advertise-peer-urls http://127.0.0.1:23800 -listen-peer-urls http://0.0.0.0:23800 -initial-cluster-token etcd-cluster-1 -initial-cluster etcd0=http://127.0.0.1:23800 -initial-cluster-state new
  while ! nc -z 127.0.0.1 2379; do
    echo "Waiting Etcd to launch on 2379..."
    sleep 1
  done
  echo "${green}Etcd is running......${reset}"
else
  echo "${db_name} non-existent"
	exit 1
fi

echo "${green}Preparing the env for UT....${reset}"
./scripts/prepare_env_ut.sh

if [ ${db_name} == "etcd" ];then
  export TEST_MODE=etcd
  [ $? == 0 ] && ut_for_dir 'datasource/mongo'
elif [ ${db_name} == "mongo" ];then
  export TEST_MODE=mongo
  [ $? == 0 ] && ut_for_dir 'datasource/etcd\|datasource/schema'
elif [ ${db_name} == "local" ];then
  export TEST_MODE=local
  [ $? == 0 ] && ut_for_dir 'datasource/etcd\|datasource/schema'
else
  echo "${db_name} non-existent"
	exit 1
fi

ret=$?

if [ ${ret} == 0 ]; then
	echo "${green}All the unit test passed..${reset}"
	echo "${green}Coverage is created in the file ./coverage.txt${reset}"
else
	echo "${red}Some or all the unit test failed..please check the logs for more details.${reset}"
	exit 1
fi

echo "${green}Service-Center finished${reset}"

echo "${green}Cleaning up the $db_name docker container${reset}"
docker rm -f "$db_name"
