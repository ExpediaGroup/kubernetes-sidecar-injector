#!/usr/bin/env bash
docker run -d --name injector -p 8443:443 --mount type=bind,src=/Users/mchandramouli/src/go/src/github.com/mchandramouli/haystack-kube-sidecar-injector/sample,dst=/etc/mutator mchandramouli/haystack-kube-sidecar-injector:1.0 -logtostderr=true 

docker logs -f $(docker ps -f name=injector -q)

