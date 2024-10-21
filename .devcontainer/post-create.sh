#!/bin/bash
OPERATOR_SDK_VERSION=v1.37.0
ARCH=$(case $(uname -m) in x86_64) echo -n amd64 ;; aarch64) echo -n arm64 ;; *) echo -n $(uname -m) ;; esac)
OS=$(uname | awk '{print tolower($0)}')

### Install the Operator SDK ###
# Adapted from https://sdk.operatorframework.io/docs/installation/

OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_${OS}_${ARCH}
curl -LO $OPERATOR_SDK_DL_URL

chmod +x operator-sdk_${OS}_${ARCH} && sudo mv operator-sdk_${OS}_${ARCH} /usr/local/bin/operator-sdk

### Install the Go tools ###
cd operator
go mod tidy