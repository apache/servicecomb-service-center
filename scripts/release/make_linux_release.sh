rm -rf servicecomb-service-center-1.0.0-m1-linux-amd64
mkdir -p servicecomb-service-center-1.0.0-m1-linux-amd64

## Make the Binary
export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
export BUILD_NUMBER=1.0.0
GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
go build --ldflags "${GO_LDFLAGS}"
go build -o service-center
cp -r service-center servicecomb-service-center-1.0.0-m1-linux-amd64
cp -r etc/conf servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r LICENCE servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r NOTICE servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r DISCLAIMER servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r README servicecomb-service-center-1.0.0-m1-linux-amd64/

## Tar the release
tar -czvf servicecomb-service-center-1.0.0-m1-linux-amd64.tar.gz servicecomb-service-center-1.0.0-m1-linux-amd64


