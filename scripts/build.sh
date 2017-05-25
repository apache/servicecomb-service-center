export OLD_PATH=$(pwd)
export BUILD_TEMP_PATH=src/github.com/servicecomb/service-center
cd ..
rm -rf src
mkdir -p  $BUILD_TEMP_PATH
cp -r $OLD_PATH/*  $BUILD_TEMP_PATH
export GOPATH=$(pwd)
cd $BUILD_TEMP_PATH
CGO_ENABLED=0 GO_EXTLINK_ENABLED=0 go build --ldflags '-w -extldflags "-static"' -o servicecenter
