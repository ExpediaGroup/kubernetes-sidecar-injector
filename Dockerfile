FROM golang:1.16 as build
RUN go get -u golang.org/x/lint/golint
WORKDIR /build
COPY . ./
RUN make release

FROM scratch
WORKDIR /
COPY --from=build /build/kubernetes-sidecar-injector /

ENTRYPOINT ["/kubernetes-sidecar-injector"]
