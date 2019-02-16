Table of Contents
=================

* [Build and deployment](#build-and-deployment)
   * [Build](#build-and-run-locally)
   * [Build and deploy in Kubernetes](#build-and-deploy-in-kubernetes)
      * [Build](#build-1)
      * [Deploy](#deploy)
      * [Label the namespace](#label-the-namespace)
      * [Test the webhook](#test-the-webhook)

## Build and deployment

### Build and run locally

* Install `dep`

```bash
go get -u github.com/golang/dep/cmd/dep
```

* Ensure [GOROOT, GOPATH and GOBIN](https://www.programming-books.io/essential/go/d6da4b8481f94757bae43be1fdfa9e73-gopath-goroot-gobin) environment variables are set correctly.

* Build

```bash
dep ensure
go build
```

* Run

```bash
./haystack-kube-sidecar-injector -port=8443 -certFile=sample/certs/cert.pem  -keyFile=sample/certs/key.pem -sideCar=sample/sidecar.yaml -logtostderr
```

* Send a sample request

```bash
curl -kvX POST --header "Content-Type: application/json" -d @sample/a//localhost:8443/mutate ttps:/
```

### Build and deploy in Kubernetes

#### Build

To build and push docker container

```bash
./build.sh push
```

#### Deploy

To deploy and test this in [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)

```bash
./deployment/deploy.sh
``` 

The command above does the following steps

* Creates a key pair, certificate request and gets it signed by Kubernetes CA. Uploads the signed certificate and private key to Kubernetes as a secret using [deployment/create-server-cert.sh](deployment/create-server-cert.sh)
* Exports Kubernetes CA file and creates a yaml file to register mutating webhook using [deployment/replace-ca-token.sh](deployment/replace-ca-token.sh)
* Uploads a config map to be used by haystack agent as config file [deployment/haystack-agent-configmap.yaml](deployment/haystack-agent-configmap.yaml)
* Uploads a config map with the container and volume spec to be injected as side cat [deployment/sidecar-configmap.yaml](deployment/sidecar-configmap.yaml)
* Uploads a deployment spec for `haystack-kube-sidecar-injector` [deployment/sidecar-injector-deployment.yaml](deployment/sidecar-injector-deployment.yaml). This spec use `sidecar-configmap` from previous step and `server certificate` from first step
* Uploads a service spec for sidecar-injector deployment [deployment/sidecar-injector-service.yaml](deployment/sidecar-injector-service.yaml)
* Uploads a spec to register the mutating webhook that was generated in step 2 [deployment/generated-mutatingwebhook.yaml](deployment/generated-mutatingwebhook.yaml)

After deployment, one can check the service running by

```bash
kubectl get pod

NAME                                                        READY     STATUS    RESTARTS   AGE
haystack-kube-sidecar-injector-deployment-5b5874466-k4gnk   1/1       Running   0          1m

```
#### Label the namespace

Before deploying an pod to see the side car being injected, one need to do one additional step.  

[Registration spec of this mutating webhook](deployment/mutatingwebhook-template.yaml#L22) specifies that this webhook be attached to only pods deployed in namespaces with a label `haystack-sidecar-injector: enabled`

Following spec applies this label to `default` namespace

```bash
kubectl apply -f sample/namespace-label.yaml
```

#### Test the webhook

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


