
# Build on the native builder platform and cross-compile for the target, so
# multi-arch builds (linux/amd64, linux/arm64) do not need QEMU emulation.
FROM --platform=$BUILDPLATFORM golang:1.27.4@sha256:205c7636d74e0ffbe3d007162df7d74f859f5a8c8322dfb9f907977686c126cf AS build

# Use the toolchain from the image; never download another one.
ENV GOTOOLCHAIN=local

# Installed for the builder platform (before GOOS/GOARCH are set for the target).
RUN go install honnef.co/go/tools/cmd/staticcheck@v0.8.2

WORKDIR /build

# Download modules in their own layer so it is cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG TARGETOS
ARG TARGETARCH
# `make release` runs clean, vet, lint (staticcheck) and a CGO_ENABLED=0 build;
# GOOS/GOARCH select the target platform for vet, lint and the build.
RUN GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} GOFLAGS=-trimpath make release

FROM scratch
WORKDIR /
COPY --from=build /build/kubernetes-sidecar-injector /

ENTRYPOINT ["/kubernetes-sidecar-injector"]
