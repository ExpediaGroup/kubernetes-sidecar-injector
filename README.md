[![Build Status](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector.svg?branch=master)](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaDotCom/haystack/blob/master/LICENSE)

Kubernetes Mutating Webhook
===========

This [mutating webhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) was developed to inject [Haystack](http://expediadotcom.github.io/haystack/)'s agent as a sidecar to a Kubernetes pod so applications can ship trace data to Haystack server. 

Though this was primarily written to inject [haystack-agent](https://github.com/ExpediaDotCom/haystack-agent) as a sidecar, __one can use this to inject any container as a sidecar in a pod__.

## Developing

If one is interested in contributing to this codebase, please read the [developer documentation](DEVELOP.md) on how to build and test this codebase.

## Using this webhook

We have provided two ways to deploy this webhook. Using [Helm](https://helm.sh/) and using [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/). Deployment files are in `deployment/helm` and `deployment/kubectl` respectively. 

### How to enable sidecar injection using this webhook

1. One can simply deploy this mutating webhook by cloning this repository and running the following command (needs kubectl installed and configured to point to the kubernets cluster or minikube)

    ```bash
    ./deployment/kubectl/deploy.sh
    ```
    
    or using helm
    
    ```bash
    helm init
    helm install --name haystack-agent-webhook ./deployment/helm
    ```
    
2. Apply the label `haystack-sidecar-injector=enabled` in the namespaces where the sidecar injection should be considered. [This sample](sample/namespace-label.yaml) file applies the label mentioned to _default_ namespace

3. Add the annotation `haystack-kube-sidecar-injector.expedia.com/inject: yes` in pod spec to inject the side car. [This sample spec](sample/echo-server.yaml#L12) shows the annotation added to a pod spec. 

### Kubectl deployment files

Lets go over the files in the __deployment/kubectl__ folder.

1.  __sidecar-configmap.yaml__:  This file contains two _configmap_ entries.  First one, _haystack-sidecar-configmap_ contains the haystack-agent sidecar container to be injected and the second one, _haystack-agent-conf-configmap_ contains a configuration file that is used by haystack-agent sidecar. 

    _Though this file carries haystack-agent, one can replace this with_ __any sidecar to be injected__. 

2. __sidecar-injector-deployment.yaml__: This file deploys _haystack-kube-sidecar-injector_ pod and _haystack-kube-sidecar-injector-svc_ service. This is the mutating webhook admission controller service. This is invoked by kebernetes while creating a new pod with the pod spec that is being created. That allows this webhook to inspect and make a decision on whether to inject the sidecar or not. This webhook checks for two conditions to determine whether to inject a sidecar or not
    1. __Namespace check__:  Sidecar injection will be attempted _only_ if the the pod is being created in a namespace with the label `haystack-sidecar-injector=enabled` __and__  the namespace is NOT `kube-system` or `kube-public`
    2. __Annotation check__: Sidecar inkection will be attempted _only_ if the pod being created carries an annotation `haystack-kube-sidecar-injector.expedia.com/inject: yes`

3. __create-server-cert.sh__: Mutating webhook admission controllers need to listen on `https (TLS)`. This script generates a key, a certificate request and gets that request signed by Kubernetes CA. i.e., produces a signed certificate and deploys it as a kubernets secret to be used by the service defined in #2

4. __mutatingwebhook-template.yaml__: This file registers the mutating webhook admission controller. This spec carries the CA file that will validate the server certificate used by the service. This file is a template and the `caBundle` field in it is populated by the script `replace-ca-token.sh` file

5. __deploy.sh__: This is a simple bash script that deploys the webhook by executing the scripts / deployment specifications mentioned above.  

### Helm deployment files

Files in __deployment/helm/templates__ are the same as the files in kubectl folder and provide the same functionality.

### Addendum

#### Injecting env variables in the sidecar

At times one may have to pass additional information to the sidecar from the pod spec. For example, a pod specific `api-key` to be used by the sidecar. To allow that, this webhook looks for special annotations with prefix `haystack-kube-sidecar-injector.expedia.com` in the pod spec and adds the annotation key-value as environment variables to the sidecar. 

For example, this [sample pod specification](sample/echo-server.yaml#L13) has the following annotation 

  ```yaml
  haystack-kube-sidecar-injector.expedia.com/some-api-key: "6feab492-fc9b-4c38-b50d-3791718c8203"
  ```

and this will cause this webhook to inject

  ```yaml
  some-api-key: "6feab492-fc9b-4c38-b50d-3791718c8203"
  ```

as an environment variable in the sidecar.

