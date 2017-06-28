#!/bin/sh
set -e
export COVERAGE_PATH=$(pwd)
cd $1
for d in $(go list ./... | grep -v vendor); do
    cd $HOME/gopath/src/
    cd $d
    set +e
    go test -cover -covermode atomic -coverprofile coverage.out  
    if [ -f coverage.out ]; then
        sed '1d;$d' coverage.out >> $HOME/gopath/src/github.com/servicecomb/service-center/coverage.txt
    fi
    set -e
done
