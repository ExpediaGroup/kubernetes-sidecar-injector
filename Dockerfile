FROM golang:1.18 as build
RUN go install golang.org/x/lint/golint@latest
WORKDIR /build
COPY . ./
RUN make release

FROM scratch
WORKDIR /
COPY --from=build /build/kubernetes-sidecar-injector /

ENTRYPOINT ["/kubernetes-sidecar-injector"]
