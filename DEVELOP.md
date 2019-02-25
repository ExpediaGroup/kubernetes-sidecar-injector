[![Build Status](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector.svg?branch=master)](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaDotCom/haystack/blob/master/LICENSE)

## Contributing

Code contributions are always welcome.

* Open an issue in the repo with defect/enhancements
* We can also be reached @ https://gitter.im/expedia-haystack/Lobby
* Fork, make the changes, build and test it locally
* Issue a PR - watch the PR build in [travis-ci](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector)
* Once merged to master, travis-ci will build and release the container with latest tag


## Dependencies

* Install `dep` and `goimports`

```bash
go get -u github.com/golang/dep/cmd/dep
go get golang.org/x/tools/cmd/goimports
go get -u golang.org/x/lint/golint
```

* Ensure [GOROOT, GOPATH and GOBIN](https://www.programming-books.io/essential/go/d6da4b8481f94757bae43be1fdfa9e73-gopath-goroot-gobin) environment variables are set correctly.

## Build and run locally

* Build

```bash
make build
```

* Run

```bash
./haystack-kube-sidecar-injector -port=8443 -certFile=sample/certs/cert.pem  -keyFile=sample/certs/key.pem -sideCar=sample/sidecar.yaml -logtostderr
```

* Send a sample request

```bash
curl -kvX POST --header "Content-Type: application/json" -d @sample/admission-request.json https://localhost:8443/mutate
```

## Build and run with docker

* Build

```bash
make docker
```

* Run

```bash
docker run -d --name injector -p 8443:443 --mount type=bind,src=/Users/mchandramouli/src/go/src/github.com/expediadotcom/haystack-kube-sidecar-injector/sample,dst=/etc/mutator expediadotcom/haystack-kube-sidecar-injector:latest -logtostderr
```

* Send a sample request

```bash
curl -kvX POST --header "Content-Type: application/json" -d @sample/admission-request.json https://localhost:8443/mutate
```

## Build and deploy in Kubernetes

### Deploy using Kubectl

To deploy and test this in [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
* build docker container locally with the fixes
* update `deployment/kubectl/sidecar-injector-deployment.yaml` file to use the local docker image
* run the command

```bash
./deployment/kubectl/deploy.sh
``` 
(To understand the command above - check the [README](README.md) file)

After deployment, one can check the service running by

```bash
kubectl get pod

NAME                                                        READY     STATUS    RESTARTS   AGE
haystack-kube-sidecar-injector-deployment-5b5874466-k4gnk   1/1       Running   0          1m

```

### Label the namespace

Before deploying a pod to see the side car being injected, one needs to do one additional step.  

[Registration spec of this mutating webhook](deployment/mutatingwebhook-template.yaml#L22) specifies that this webhook be called only for pods deployed in namespaces with a label `haystack-sidecar-injector: enabled`

Following spec applies this label to `default` namespace

```bash
kubectl apply -f sample/namespace-label.yaml
```

### Test the webhook

One can run the following command to deploy a sample `echo-server`. Note, this [deployment spec carries an annotation](sample/echo-server.yaml#L12) `haystack-kube-sidecar-injector.expedia.com/inject: "yes"` that triggers injection of the sidecar.

```bash
kubectl apply -f sample/echo-server.yaml
```

One can then run the following command to confirm the sidecar has been injected

```bash
kubectl get pod

NAME                                                        READY     STATUS             RESTARTS   AGE
echo-server-deployment-849b87649d-9x95k                     2/2       Running            0          4m
haystack-kube-sidecar-injector-deployment-cc4648b7f-bdk2v   1/1       Running            0          6m
```

Note the **2 containers** in the echo-server pod instead of one. 


