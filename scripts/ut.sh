#!/bin/sh
set -e
export COVERAGE_PATH=$(pwd)
cd $1
for d in $(go list ./... | grep -v vendor); do
    cd $GOPATH/src/$d
    if [ $(ls | grep _test.go | wc -l) -gt 0 ]; then
        go test -cover -covermode atomic -coverprofile coverage.out  
        if [ -f coverage.out ]; then
            sed '1d;$d' coverage.out >> $GOPATH/src/github.com/apache/incubator-servicecomb-service-center/coverage.txt
        fi
    fi
done
