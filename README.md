[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaDotCom/haystack/blob/master/LICENSE)

Table of Contents
=================

* [Table of Contents](#table-of-contents)
  * [Build and deployment](#build-and-deployment)
     * [Dependencies](#dependencies)
     * [Build and run locally](#build-and-run-locally)
     * [Build and run with docker](#build-and-run-with-docker)
     * [Build and deploy in Kubernetes](#build-and-deploy-in-kubernetes)
        * [Build](#build)
        * [Deploy using Kubectl](#deploy-using-kubectl)
        * [Deploy using Helm](#deploy-using-helm)
        * [Label the namespace](#label-the-namespace)
        * [Test the webhook](#test-the-webhook)


# Build and deployment

## Dependencies

* Install `dep` and `goimports`

```bash
go get -u github.com/golang/dep/cmd/dep
go get golang.org/x/tools/cmd/goimports
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
docker run -d --name injector -p 8443:443 --mount type=bind,src=/Users/mchandramouli/src/go/src/github.com/mchandramouli/haystack-kube-sidecar-injector/sample,dst=/etc/mutator mageshcmouli/haystack-kube-sidecar-injector:latest -logtostderr
```

* Send a sample request

```bash
curl -kvX POST --header "Content-Type: application/json" -d @sample/admission-request.json https://localhost:8443/mutate
```

## Build and deploy in Kubernetes

### Build

To build and push docker container

```bash
make release
```

### Deploy using Kubectl
To deploy and test this in [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)

```bash
./deployment/kubectl/deploy.sh
``` 

The command above does the following steps

* Creates a key pair, certificate request and gets it signed by Kubernetes CA. Uploads the signed certificate and private key to Kubernetes as a secret using [deployment/kubectl/create-server-cert.sh](deployment/kubectl/create-server-cert.sh)
* Exports Kubernetes CA file and creates a yaml file to register mutating webhook using [deployment/kubectl/replace-ca-token.sh](deployment/kubectl/replace-ca-token.sh)
* Uploads a config map to be used by haystack agent as config file [deployment/kubectl/haystack-agent-configmap.yaml](deployment/kubectl/haystack-agent-configmap.yaml)
* Uploads a config map with the container and volume spec to be injected as side car [deployment/kubectl/sidecar-configmap.yaml](deployment/kubectl/sidecar-configmap.yaml)
* Uploads a deployment spec for `haystack-kube-sidecar-injector` [deployment/kubectl/sidecar-injector-deployment.yaml](deployment/kubectl/sidecar-injector-deployment.yaml). This spec uses `sidecar-configmap` from previous step and `server certificate` from first step
* Uploads a service spec for sidecar-injector deployment [deployment/kubectl/sidecar-injector-service.yaml](deployment/kubectl/sidecar-injector-service.yaml)
* Uploads a spec to register the mutating webhook that was generated in step 2 [deployment/kubectl/generated-mutatingwebhook.yaml](deployment/kubectl/generated-mutatingwebhook.yaml)

After deployment, one can check the service running by

```bash
kubectl get pod

NAME                                                        READY     STATUS    RESTARTS   AGE
haystack-kube-sidecar-injector-deployment-5b5874466-k4gnk   1/1       Running   0          1m

```

### Deploy using Helm

Follow the steps mentioned below to install the helm chart

1. install the helm client based on the instructions given [here](https://docs.helm.sh/using_helm/#installing-helm)
2. configure helm to point to kubernetes cluster
```console
$ minikube start
$ helm init
```
3. move to the directory where the code is cloned
4. run the following command
```console
$ helm install --name haystack-agent-webhook ./deployment/helm
```

The following table lists the configurable parameters of the helm chart and
their default values.

| Parameter                   | Description                                                                                | Default         |
|:----------------------------|:-------------------------------------------------------------------------------------------|:----------------|
| `image.repository`          | Container image to use                                                                     | `mageshcmouli/haystack-kube-sidecar-injector`      |
| `image.tag`                 | Container image tag to deploy                                                              |  `latest`      |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.

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


