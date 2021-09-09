#!/bin/bash -eux

 cwd=$(pwd)

 pushd $cwd/dp-filter-api
   make lint
 popd