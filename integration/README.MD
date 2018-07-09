# Service-Center Integration Tests
These are high-level tests for the Service-Center that tests all the API's of the Service-Center

These tests assumes that you have docker installed in your system.

## How to run:
This test starts etcd in docker container and then builds and runs the service-center from source present in your GOPATH.
This test assumes you have already downloaded all the dependencies using 'glide install'.

```
bash -x scripts/integration_test.sh 
```
