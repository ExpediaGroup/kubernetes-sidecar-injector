#!/bin/bash
set -euo pipefail
set -o errexit
set -o nounset
# echo commands
#set -x

file-check-and-apply() {
    if [[ ! -f $1 ]]; then
        echo "*** ERROR: File not found: $1  ***"
        exit 1
    fi

    kubectl apply -f $1
}

BASEDIR=`dirname $0`

# ensure kubectl
if [[ ! "$(command -v kubectl)" ]]; then
    echo "kubectl not found"
    exit 1
fi

# create server cert and key and inject it as
#

# This script uses k8s' CertificateSigningRequest API to a generate a
# certificate signed by k8s CA suitable for use with webhook
# services. This requires permissions to create and approve CSR. See
# https://kubernetes.io/docs/tasks/tls/managing-tls-in-a-cluster for
# detailed explanation and additional instructions.
#
# The server key/cert k8s CA cert are stored in a k8s secret.
#
# namespace: default
# secret-name: kubernetes-sidecar-injector-certs
# service-name: kubernetes-sidecar-injector-svc
${BASEDIR}/create-server-cert.sh -p kubernetes-sidecar-injector-certs -n default -s kubernetes-sidecar-injector-svc

# create deployment spec for mutating-webhook.
# This script uses a template and populates the template with the CA file from k8s
cat ${BASEDIR}/mutatingwebhook-template.yaml | ${BASEDIR}/replace-ca-token.sh > ${BASEDIR}/generated-mutatingwebhook.yaml

# deploy the config map used by the injected sidecar
# deploy the config map with the side care spec used by webhook
file-check-and-apply ${BASEDIR}/sidecar-configmap.yaml

# deploy the sidecar injector webhook
# deploy the sidecar injector service
file-check-and-apply ${BASEDIR}/sidecar-injector-deployment.yaml

# deploy the sidecar injecting webhook
file-check-and-apply ${BASEDIR}/generated-mutatingwebhook.yaml


