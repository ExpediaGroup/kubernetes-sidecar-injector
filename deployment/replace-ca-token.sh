#!/bin/bash
set -euo pipefail
set -o errexit
set -o nounset

export CA_BUNDLE=$(kubectl get configmap -n kube-system extension-apiserver-authentication -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')

sed -e "s|\${CA_BUNDLE}|${CA_BUNDLE}|g"
