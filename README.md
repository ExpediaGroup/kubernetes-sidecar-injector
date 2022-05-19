![example branch parameter](https://github.com/ExpediaGroup/kubernetes-sidecar-injector/actions/workflows/deploy.yaml/badge.svg?branch=master)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaGroup/kubernetes-sidecar-injector/blob/master/LICENSE)

Kubernetes Mutating Webhook
===========

https://hub.docker.com/r/expediagroup/kubernetes-sidecar-injector

This [mutating webhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) was developed to inject sidecars to a Kubernetes pod. 

## Developing

If one is interested in contributing to this codebase, please read the [developer documentation](DEVELOP.md) on how to build and test this codebase.

## Using this webhook

We have provided two ways to deploy this webhook. Using [Helm](https://helm.sh/) and using [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/). Deployment files are in `deployment/helm` and `deployment/kubectl` respectively.

### ConfigMap Sidecar Configuration

NOTE: Applications only have access to sidecars in their own namespaces.

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-sidecar
  namespace: my-app-namespace
data:
  sidecars.yaml: |
    - name: # Sidcar Name
      initContainers:
        - name: # Example 1
          image: # Example 1
      containers:
        - name: # Example 2
          image: # Example 2
      volumes:
        - name: # Example 3
          configMap:
            name: # Example 3
      imagePullSecrets:
        - name: # Example 4
```


### How to enable sidecar injection using this webhook

1. Deploy this mutating webhook by cloning this repository and running the following command (needs kubectl installed and configured to point to the kubernetes cluster or minikube)

```bash
make helm-install
```

2. By default, all namespaces are watched except `kube-system` and `kube-public`. This can be configured in your [helm values](charts/kubernetes-sidecar-injector/values.yaml#L13-L19).

3. Add the annotation ([`sidecar-injector.expedia.com/inject`](charts/kubernetes-sidecar-injector/values.yaml#L9-L10) by default) with ConfigMap sidecar name to inject in pod spec where sidecar needs to be injected. [This sample spec](sample/chart/echo-server/templates/deployment.yaml#L16) shows such an annotation added to a pod spec to inject `haystack-agent`.

4. Create your ConfigMap sidecar configuration

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-app-sidecar
  namespace: {{ .Release.Namespace }}
data:
  sidecars.yaml: |
    - name: busybox
      initContainers:
        - name: busybox
          image: busybox
          command: [ "/bin/sh" ]
          args: [ "-c", "echo '<html><h1>Hi!</h1><html>' >> /work-dir/index.html" ]
          volumeMounts:
            - name: workdir
              mountPath: "/work-dir"
```
## How to use the kubernetes-sidecar-injector Helm repository

You need to add this repository to your Helm repositories:

```
helm repo add kubernetes-sidecar-injector  https://opensource.expediagroup.com/kubernetes-sidecar-injector/
helm repo update
```

### Kind Testing
```shell
make kind
make install-sample-init-container # or make install-sample-container
make follow-logs
```