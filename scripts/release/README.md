### Service-Center Release

#### Making Release

To make a release of Service-center please follow the below steps

```
## Clone the code in a proper GOPATH

git clone https://github.com/apache/servicecomb-service-center.git $GOPATH/src/github.com/apache/servicecomb-service-center
cd $GOPATH/src/github.com/apache/servicecomb-service-center

## Download all the dependency
curl https://glide.sh/get | sh
glide install
```

##### Note: [bower](https://www.npmjs.com/package/bower) should be installed in this machine

##### Note: We have used `go version go1.11` for making this release

#### Linux Release

```
# bash -x scripts/release/make_release.sh OS_NAME VERSION_NUMBER PACKAGE_NUMBER
bash -x scripts/release/make_release.sh linux 1.1.0 1.1.0
```

#### Windows Release

```
# bash -x scripts/release/make_release.sh OS_NAME VERSION_NUMBER PACKAGE_NUMBER
bash -x scripts/release/make_release.sh windows 1.1.0 1.1.0
```

#### Mac OS Release

```
# bash -x scripts/release/make_release.sh OS_NAME VERSION_NUMBER PACKAGE_NUMBER
bash -x scripts/release/make_release.sh mac 1.1.0 1.1.0
```

