#!/bin/bash

# build base image

DIR=$(cd `dirname $0` && pwd)
(cd $DIR && docker build --rm -t myodc/playground-base .)
