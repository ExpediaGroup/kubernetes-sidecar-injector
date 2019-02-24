[![Build Status](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector.svg?branch=master)](https://travis-ci.org/ExpediaDotCom/haystack-kube-sidecar-injector)
[![License](https://img.shields.io/badge/license-Apache%20License%202.0-blue.svg)](https://github.com/ExpediaDotCom/haystack/blob/master/LICENSE)

Kubernetes Mutating Webhook
===========

This [mutating webhook](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#mutatingadmissionwebhook) was to inject [Haystack](http://expediadotcom.github.io/haystack/)'s agent as a sidecar in a kubernetes pod so applications can ship trace data to Haystack server. 

Though this was primarily written to inject [haystack-agent](https://github.com/ExpediaDotCom/haystack-agent) as a sidecar, one can use this to inject any container as a sidecar in a pod.

## Developing

If one is interested in contributing to this codebase, please read the [developer documentation](DEVELOP.md) on how to build and test this codebase.

## Using this webhook



