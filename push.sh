#! /bin/bash

SHA1=$1

# Exit on any error
set -e

#Check this commit is tagging
TAG=`git name-rev --name-only --tags $SHA1`

echo $TAG

if [ ! $TAG == 'undefined' ] ; then

echo "Building and Pushing images"
docker tag goloom/loom:$SHA1 goloom/images:$TAG
docker push goloom/images:$TAG 

docker build -t goloom/loom:$TAG-ubuntu14.04 -f Dockerfile.ubuntu14.04 . 
docker push goloom/loom:$TAG-ubuntu14.04

fi


