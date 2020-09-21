#!/bin/bash -eux

export cwd=$(pwd)

pushd $cwd/dp-filter-api
  make audit
popd   