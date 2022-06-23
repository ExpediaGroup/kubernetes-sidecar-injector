![example branch parameter](https://github.com/ExpediaGroup/kubernetes-sidecar-injector/actions/workflows/deploy.yaml/badge.svg?branch=main)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaGroup/kubernetes-sidecar-injector/blob/main/LICENSE)

## Contributing

Code contributions are always welcome.

* Open an issue in the repo with defect/enhancements
* We can also be reached @ https://gitter.im/expedia-haystack/Lobby
* Fork, make the changes, build and test it locally
* Issue a PR- watch the PR build in [deploy](https://github.com/ExpediaGroup/kubernetes-sidecar-injector/actions)
* Once merged to main, GitHub Actions will build and release the container with latest tag


## Dependencies

* Ensure [GOROOT, GOPATH and GOBIN](https://www.programming-books.io/essential/go/d6da4b8481f94757bae43be1fdfa9e73-gopath-goroot-gobin) environment variables are set correctly.

## Build and run using an IDE (JetBrains)
Run the included [`go build kubernetes-sidecar-injector`](.run/go build kubernetes-sidecar-injector.run.xml) `Go Build` job.

## Build and deploy in Kubernetes

### Deploy using Kubectl

To deploy and test this in [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation/)

* run the command

```bash
make kind-install
```
(To understand the command above - check the [README](README.md) file)

After deployment, one can check the service running by

```bash
kubectl get pods -n sidecar-injector

NAME                                           READY   STATUS    RESTARTS   AGE
kubernetes-sidecar-injector-78648d458b-7cv7l   1/1     Running   0          32m
```

### Test the webhook

Run the following command to deploy a sample `echo-server`. Note, this [deployment spec carries an annotation](sample/chart/echo-server/templates/deployment.yaml#L16) `sidecar-injector.expedia.com/inject: "haystack-agent"` that triggers injection of `haystack-agent` sidecar defined in [sidecar-configmap.yaml](sample/chart/echo-server/templates/sidecar-configmap.yaml) file.

```bash
make install-sample-container
```

One can then run the following command to confirm the sidecar has been injected

```bash
kubectl get pod -n sample

NAME                                                        READY     STATUS             RESTARTS   AGE
echo-server-deployment-849b87649d-9x95k                     2/2       Running            0          4m
```

Note the **2 containers** in the echo-server pod instead of one. 

### Clean up webhook

Run the following commands to delete and cleanup the deployed webhook

```
helm delete -n sidecar-injector kubernetes-sidecar-injector
helm delete -n sample sample-echo-server-sidecar-injector
```






