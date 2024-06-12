FROM golang:1.21 as build
RUN go install honnef.co/go/tools/cmd/staticcheck@latest
WORKDIR /build
COPY . ./
RUN make release

FROM scratch
WORKDIR /
COPY --from=build /build/kubernetes-sidecar-injector /

ENTRYPOINT ["/kubernetes-sidecar-injector"]
