#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-filter-api
  make build && mv build/dp-filter-api $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
