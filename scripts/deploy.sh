#!/bin/sh

# Deploy a new version of the setlist searcher to the k8s cluster on gke.
# This script requires kubectl to be setup properly.

# exit when any command fails
set -e

# keep track of the last executed command
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG
# echo an error message before exiting
trap 'echo "\"${last_command}\" command filed with exit code $?."' EXIT

IMAGE=gcr.io/setlist-searcher/setlist-search

# Set the tag to the last git revision. So we need to commit changes before
# pushing.
TAG=`git rev-parse HEAD | cut -c 1-10`

docker build -t setlist-search .
docker tag setlist-search ${IMAGE}:${TAG}
docker push ${IMAGE}:${TAG}
kubectl set image deployment/setlist-search-deployment setlist-search=${IMAGE}:${TAG}
