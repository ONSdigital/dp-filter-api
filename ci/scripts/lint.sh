#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-filter-api
  pip install poetry
  make -C sdk/python install-dev
  make lint
popd