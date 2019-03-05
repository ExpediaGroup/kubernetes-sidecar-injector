#!/bin/bash
set -euo pipefail
set -o errexit
set -o nounset

usage() {
    cat <<EOF
Generate k8s ca signed certificate and set the certificate and key as a secret in k8s.

usage: ${0} [OPTIONS]

The following flags are required.

       --s       Service name
       --n       Namespace where service and secret reside.
       --p       Secret name for CA certificate and server certificate/key pair.
EOF
}

SERVICE=
SECRET=
NAMESPACE=
BASEDIR=$(dirname "$0")

while getopts "hs:n:p:" OPTION
do
     case ${OPTION} in
         h)
             usage
             exit 1
             ;;
         n)
             NAMESPACE=${OPTARG}
             ;;
         p)
             SECRET=${OPTARG}
             ;;
         s)
             SERVICE=${OPTARG}
             ;;
         ?)
             usage
             exit
             ;;
     esac
done

[[ -z ${SERVICE} ]] && SERVICE=kubernetes-sidecar-injector-svc
[[ -z ${SECRET} ]] && SECRET=kubernetes-sidecar-injector-certs
[[ -z ${NAMESPACE} ]] && NAMESPACE=default

if [[ ! "$(command -v openssl)" ]]; then
    echo "openssl not found"
    exit 1
fi

CSR_NAME=${SERVICE}.${NAMESPACE}
TMP_DIR=$(mktemp -d)
#TMP_DIR=/Users/mchandramouli/tmp

echo "creating certs in tmpdir ${TMP_DIR} "
cat ${BASEDIR}/csr-template.conf | sed -e "s|\${SERVICE}|${SERVICE}|g" -e "s|\${NAMESPACE}|${NAMESPACE}|g" > ${TMP_DIR}/csr.conf

openssl genrsa -out ${TMP_DIR}/server-key.pem 2048
openssl req -new -key ${TMP_DIR}/server-key.pem -subj "/CN=${SERVICE}.${NAMESPACE}.svc" -out ${TMP_DIR}/server.csr -config ${TMP_DIR}/csr.conf

# clean-up any previously created CSR for our service. Ignore errors if not present.
kubectl delete csr ${CSR_NAME} 2>/dev/null || true

# generate certificate signing request
CSR_STR=`cat ${TMP_DIR}/server.csr | base64 | tr -d '\n'`
cat ${BASEDIR}/csr-template.yaml | sed -e "s|\${CSR_NAME}|${CSR_NAME}|g" -e "s|\${CSR_STR}|${CSR_STR}|g" > ${TMP_DIR}/gen_csr.yaml

# create  server cert/key CSR and  send to k8s API
kubectl create -f ${TMP_DIR}/gen_csr.yaml

# verify CSR has been created
while true; do
    kubectl get csr ${CSR_NAME}
    if [[ "$?" -eq 0 ]]; then
        break
    fi
done

# approve and fetch the signed certificate
kubectl certificate approve ${CSR_NAME}

# verify certificate has been signed
for x in $(seq 10); do
    SERVER_CERT=$(kubectl get csr ${CSR_NAME} -o jsonpath='{.status.certificate}')
    if [[ ${SERVER_CERT} != '' ]]; then
        break
    fi
    sleep 1
done

if [[ ${SERVER_CERT} == '' ]]; then
    echo "ERROR: After approving csr ${CSR_NAME}, the signed certificate did not appear on the resource. Giving up after 10 attempts." >&2
    exit 1
fi
echo ${SERVER_CERT} | openssl base64 -d -A -out ${TMP_DIR}/server-cert.pem


# create the secret with CA cert and server cert/key
kubectl create secret generic ${SECRET} \
        --from-file=key.pem=${TMP_DIR}/server-key.pem \
        --from-file=cert.pem=${TMP_DIR}/server-cert.pem \
        --dry-run -o yaml |
    kubectl -n ${NAMESPACE} apply -f -