rm -rf servicecomb-service-center-1.0.0-m1-linux-amd64
mkdir -p servicecomb-service-center-1.0.0-m1-linux-amd64

## Make the Binary
go build -o service-center
cp -r service-center servicecomb-service-center-1.0.0-m1-linux-amd64
cp -r etc/conf servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r LICENCE servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r NOTICE servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r DISCLAIMER servicecomb-service-center-1.0.0-m1-linux-amd64/
cp -r README servicecomb-service-center-1.0.0-m1-linux-amd64/

## Tar the release
tar -czvf servicecomb-service-center-1.0.0-m1-linux-amd64.tar.gz servicecomb-service-center-1.0.0-m1-linux-amd64


