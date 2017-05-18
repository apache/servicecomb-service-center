#!/bin/bash

# //////////////////////////////////////////////////////////////////////////// #
#                                                                              #
#                              Global Variables                                #
#                                                                              #
# //////////////////////////////////////////////////////////////////////////// #
set -x
set +e
export c_name="etcd_sc"
docker rm -f $c_name
set -e
export BUILD_DIR=$(cd $(dirname $0); pwd)
export PROJECT_DIR=$(cd $BUILD_DIR/../../;pwd)

# //////////////////////////////////////////////////////////////////////////// #
#                                                                              #
#                              Functions                                       #
#                                                                              #
# //////////////////////////////////////////////////////////////////////////// #

# //////////////////////////////////////////////////////////////////////////// #
#                                                                              #
#                              Executions                                      #
#                                                                              #
# //////////////////////////////////////////////////////////////////////////// #
echo "ut test hsa-service-center"
bash $PROJECT_DIR/tests/scripts/setup_dependency.sh
export GIT_SSL_NO_VERIFY=1
git config --global http.sslverify false
go install github.com/onsi/ginkgo/ginkgo
go install github.com/onsi/gomega


docker run -d -v /usr/share/ca-certificates/:/etc/ssl/certs -p 40010:40010 -p 23800:23800 -p 23790:23790 \
  --name $c_name 10.162.197.95:5000/cse/etcd:v3.1.0 \
 etcd -name etcd0 \
 -advertise-client-urls http://${HostIP}:23790,http://${HostIP}:40010 \
 -listen-client-urls http://0.0.0.0:23790,http://0.0.0.0:40010 \
 -initial-advertise-peer-urls http://${HostIP}:23800 \
 -listen-peer-urls http://0.0.0.0:23800 \
 -initial-cluster-token etcd-cluster-1 \
 -initial-cluster etcd0=http://${HostIP}:23800 \
 -initial-cluster-state new
while ! nc -z 127.0.0.1 23790; do
  echo "Waiting Etcd to launch on 23790..."
  sleep 1
done
ut_root="$PROJECT_DIR/src/github.com/servicecomb/service-center/server/service"
export GOPATH=${PROJECT_DIR}

if [ ! -d "${ut_root}/conf/" ]; then
    mkdir -p ${ut_root}/conf/
fi

export APP_ROOT=${ut_root}
export SSL_ROOT=$PROJECT_DIR/deployment/release/cse-service-center/etc/ssl
export CIPHER_ROOT=$PROJECT_DIR/deployment/release/cse-service-center/etc/cipher
export registry_address="127.0.0.1:23790"
ssl_mode="${CSE_SSL_MODE:-"0"}"
#sed -i "s/\${ssl_passphase}/00000001000000014DCF0FE2FEDEE0D92A0018D45F5A8631D45C4905A7420F6111A2947E05A4FBBD/g" $PROJECT_DIR/deployment/release/cse-service-center/etc/ssl/ssl.properties

cp $PROJECT_DIR/deployment/release/cse-service-center/conf/* ${ut_root}/conf/
sed -i s/"ssl_mode.*=.*$"/"ssl_mode = $ssl_mode"/g ${ut_root}/conf/app.conf
sed -i s@"manager_cluster.*=.*$"@"manager_cluster = \"$registry_address\""@g ${ut_root}/conf/app.conf

cp -r ${ut_root}/conf ${ut_root}/microservice/

ginkgo -r -v -cover ${ut_root}
if [ $? == 0 ]; then
	echo "test service-center ok"
else
	echo "test service-center failed"

	exit 1
fi

echo "finished build service-center"

docker rm -f $c_name


