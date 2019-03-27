#!/bin/sh

# Deploy a new version of the setlist searcher to the k8s cluster on gke.
# This script requires kubectl to be setup properly.

# exit when any command fails
set -e

IMAGE=gcr.io/setlist-searcher/setlist-search

# Set the tag to the last git revision. So we need to commit changes before
# pushing.
TAG=`git rev-parse HEAD | cut -c 1-10`

echo "Building the container image...\n"
docker build -t setlist-search .
echo "Tagging the image with version ${TAG}"
docker tag setlist-search ${IMAGE}:${TAG}
docker push ${IMAGE}:${TAG}
kubectl set image deployment/setlist-search-deployment setlist-search=${IMAGE}:${TAG}
