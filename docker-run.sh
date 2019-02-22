#!/usr/bin/env bash
make docker

docker run -d --name injector -p 8443:443 --mount type=bind,src=/Users/mchandramouli/src/go/src/github.com/expediadotcom/haystack-kube-sidecar-injector/sample,dst=/etc/mutator expediadotcom/haystack-kube-sidecar-injector:latest -logtostderr

docker logs -f $(docker ps -f name=injector -q)

