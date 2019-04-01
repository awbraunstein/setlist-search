# setlist-search
Website for searching through setlists using a regex like syntax.

## Deployment

Make sure you have `gcloud` installed and setup with `setlist-searcher` as the
default project.

### Launching/Updating the docker container.

Run the `deploy.sh` script.

`scripts/deploy.sh`

### Run on GKE

Create the cluster.

`gcloud container clusters create searchphish-cluster --num-nodes=2 --zone=us-west1-a`

Get the credentials so we can use kubectl

`gcloud container clusters get-credentials searchphish-cluster`

Get the admin password.

`gcloud container clusters describe searchphish-cluster --zone us-west1-a | grep password`

Setup traefik.

`kubectl apply -f manifests/traefik.yaml --username=admin --password=<PASSWORD>`

Setup the setlist search backend.

`kubectl apply -f manifests/setlist-search.yaml`

Setup the ingress rule.

`kubectl apply -f manifests/ingress.yaml`

Open ports 80 and 443 on your nodes, make your external IPs static, and add your
external IPs as A record entries in your DNS Nameserver.
