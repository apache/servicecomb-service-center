#! /bin/bash -e

SCRIPT_DIR=$(cd $(dirname $0); pwd)
BASE_DIR=${SCRIPT_DIR}/../../../

mkdir -p $SCRIPT_DIR/service-center

cd $BASE_DIR

# make CGO_ENABLED=0 since busybox will not execute if it is dynamically linked
export CGO_ENABLED=0

# buils service-center
go build -o service-center

#copy service-center executable in the build-image/service-center
cp ./service-center $SCRIPT_DIR/service-center

# give permission 
chmod 755 $SCRIPT_DIR/service-center/service-center

# copy the conf folder to build-image/service-center 
cp -r $BASE_DIR/etc/conf $SCRIPT_DIR/service-center

#go to the script directory
cd $SCRIPT_DIR

# store the etcd image name  and version
BASE_IMAGE=quay.io/coreos/etcd
BASE_IMAGE_VERSION=v3.1.0

cp Dockerfile.tmpl Dockerfile

sed -i "s|{{.BASE_IMAGE}}|${BASE_IMAGE}|g" Dockerfile
sed -i "s|{{.BASE_IMAGE_VERSION}}|${BASE_IMAGE_VERSION}|g" Dockerfile

docker build -t developement/servicecomb/service-center:latest .
docker save developement/servicecomb/service-center:latest |gzip >service-center-dev.tgz

# remove the service-center directory from the build-image path
rm -rf service-center Dockerfile 
