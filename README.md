## Build and running locally

Install `dep`

```bash
go get -u github.com/golang/dep/cmd/dep
```

Ensure [GOROOT, GOPATH and GOBIN](https://www.programming-books.io/essential/go/d6da4b8481f94757bae43be1fdfa9e73-gopath-goroot-gobin) environment variables are set correctly.

Run 

```bash
dep ensure
go install
$GOBIN/haystack-kube-sidecar-injector --port 8443 --certFile sample/server-cert.pem --keyFile sample/server-key.pem --sideCar=sample/sidecar.yaml -logtostderr=true
```
