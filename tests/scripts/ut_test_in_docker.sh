#!/bin/bash

set -x
set +e
export c_name="etcd"
docker rm -f $c_name
set -e
red=`tput setaf 1`
green=`tput setaf 2`
reset=`tput sgr0`
echo "${green}Starting Unit Testing for Service Center${reset}"

echo "${green}Starting etcd in docker${reset}"
docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 40010:40010 -p 23800:23800 -p 2379:2379 --name etcd quay.io/coreos/etcd etcd -name etcd0 -advertise-client-urls http://127.0.0.1:2379,http://127.0.0.1:40010 -listen-client-urls http://0.0.0.0:2379,http://0.0.0.0:40010 -initial-advertise-peer-urls http://127.0.0.1:23800 -listen-peer-urls http://0.0.0.0:23800 -initial-cluster-token etcd-cluster-1 -initial-cluster etcd0=http://127.0.0.1:23800 -initial-cluster-state new
while ! nc -z 127.0.0.1 2379; do
  echo "Waiting Etcd to launch on 2379..."
  sleep 1
done
echo "${green}Etcd is running......${reset}"

echo "${green}Preparing the env for UT....${reset}"
./scripts/prepare_env_ut.sh

echo "${green}Running UT for Service-Center server"
bash -x ./scripts/ut.sh server
  
if [ $? == 0 ]; then
	echo "${green}All the unit test passed..${reset}"
	echo "${green}Coverage is created in the file ./coverage.txt${reset}"
else
	echo "${red}Some or all the unit test failed..please check the logs for more details.${reset}"
	exit 1
fi

echo "${green}Service-Center finished${reset}"

echo "${green}Cleaning up the etcd docker container${reset}"
docker rm -f $c_name
