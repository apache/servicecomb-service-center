### Service-Center Release

#### Making Release

To make a release of Service-center please follow the below steps

```
## Clone the code in a proper GOPATH

git clone https://github.com/apache/incubator-servicecomb-service-center.git $GOPATH/src/github.com/apache/incubator-servicecomb-service-center
cd $GOPATH/src/github.com/apache/incubator-servicecomb-service-center

## Download all the dependency
go get github.com/FiloSottile/gvt
gvt restore

```

#### Linux Release

```
# bash -x scripts/release/make_release.sh OS_NAME VERSION_NUMBER PACKAGE_NUMBER
bash -x scripts/release/make_release.sh linux 1.0.0 1.0.0-m1
```

#### Windows Release

```
# bash -x scripts/release/make_release.sh OS_NAME VERSION_NUMBER PACKAGE_NUMBER
bash -x scripts/release/make_release.sh windows 1.0.0 1.0.0-m1
```


