set +e
rm -rf servicecomb-service-center-1.0.0-m1-linux-amd64
rm -rf servicecomb-service-center-1.0.0-m1-linux-amd64.tar.gz
set -e
mkdir -p servicecomb-service-center-1.0.0-m1-linux-amd64

## Make the Binary
export GIT_COMMIT=$(git log  --pretty=format:'%h' -n 1)
export BUILD_NUMBER=$BUILD_NUMBER
GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.BUILD_TAG=$(date +%Y%m%d%H%M%S).$BUILD_NUMBER.$GIT_COMMIT'"
GO_LDFLAGS="${GO_LDFLAGS} -X 'github.com/apache/incubator-servicecomb-service-center/version.VERSION=$BUILD_NUMBER'"
go build --ldflags "${GO_LDFLAGS}"

## Copy the necessary file
cp -r incubator-servicecomb-service-center servicecomb-service-center-1.0.0-m1-linux-amd64
cp -r scripts/release/conf/* servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r LICENSE servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r NOTICE servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r DISCLAIMER servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r README.md servicecomb-service-center-1.0.0-m1-linux-amd64/

## Tar the release
tar -czvf servicecomb-service-center-1.0.0-m1-linux-amd64.tar.gz servicecomb-service-center-1.0.0-m1-linux-amd64
set +e
rm -rf servicecomb-service-center-1.0.0-m1-linux-amd64
rm -rf incubator-servicecomb-service-center

##DONE

