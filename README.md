# setlist-search
Website for searching through setlists using a regex like syntax.

# Deployment

Make sure you have `gcloud` installed and setup with `setlist-searcher` as the
default project.

## Build the Docker image.

First we build the image.

`docker build -t setlist-search .`

Next we tag it for gcr.

`docker tag setlist-search gcr.io/setlist-searcher/setlist-search`

Finally we push the image to gcr.

`docker push gcr.io/setlist-searcher/setlist-search`

## Run on GKE

Create the cluster.

`gcloud container clusters create searchphish-cluster --num-nodes=2 --zone=us-west1-a`

Get the credentials so we can use kubectl

`gcloud container clusters get-credentials searchphish-cluster`

Get the admin password.

`gcloud container clusters describe searchphish-cluster --zone us-west1-a | grep password`

Setup traefik.

`kubectl apply -f manifests/traefik.yaml --username=admin --password=mrYpEl4UevMG8iU7`

Setup the setlist search backend.

`kubectl apply -f manifests/setlist-search.yaml`

Setup the ingress rule.

`kubectl apply -f manifests/ingress.yaml`
