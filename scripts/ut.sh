#!/bin/sh
set -e
export COVERAGE_PATH=$(pwd)
cd $1
for d in $(go list ./... | grep -v vendor); do
    cd $HOME/gopath/src/
    cd $d
    set +e
    ginkgo -r -v -cover  
    if [ -f *.coverprofile ]; then
        sed '1d;$d' *.coverprofile >> $HOME/gopath/src/github.com/servicecomb/service-center/coverage.txt
    fi
    set -e
done
