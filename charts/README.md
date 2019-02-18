# haystack kube sidecar Helm Chart
This helm chart creates a [Haystack agent mutating webhook](https://github.com/expediadotcom/haystack-agent).

## Prerequisites
* Kubernetes 1.10

## Chart Details

This chart will do the following:

* Deploy haystack agent mutation webhook to inject haystack agent sidecar along with apps

## Installing the Chart

To install the chart, use the following:
1. clone this github repo code
2. install the helm client based on the instructions given [here](https://docs.helm.sh/using_helm/#installing-helm)
3. configure helm to point to kubernetes cluster
```console
$ minikube start
$ helm init
```
3. move to the directory where the code is cloned
4. run the following command
```console
$ helm install --name haystack-agent-webhook ./charts
```

## Configuration

The following table lists the configurable parameters of the docker-registry chart and
their default values.

| Parameter                   | Description                                                                                | Default         |
|:----------------------------|:-------------------------------------------------------------------------------------------|:----------------|
| `image.repository`          | Container image to use                                                                     | `registry`      |
| `image.tag`                 | Container image tag to deploy                                                                 `0.4.3`      |

Specify each parameter using the `--set key=value[,key=value]` argument to
`helm install`.
