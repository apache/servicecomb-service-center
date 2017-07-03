#!/usr/bin/env bash
set -e
set -x
SCRIPT_DIR=$(cd $(dirname $0); pwd)
PROJECT_DIR=$(cd $SCRIPT_DIR/../../;pwd)
export GOPATH=${PROJECT_DIR}
src_path="${PROJECT_DIR}/src/servicecenter"

cp -r ${PROJECT_DIR}/vendor/*  ${PROJECT_DIR}/src
cp -r ${PROJECT_DIR}/tests/vendor/* ${PROJECT_DIR}/src
#cp -r ${PROJECT_DIR}/tests/src/* ${PROJECT_DIR}/src
