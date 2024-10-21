# RQ Operator

A kubernetes Operator to manage resourceQuotas for namespaces. The operator is written in Go and use the Operator SDK. A basic guide on how to create operators can be found [here](https://sdk.operatorframework.io/docs/building-operators/golang/).

## Setup

```bash
mkdir -p operator
cd operator

operator-sdk init --domain=server.home --repo gitlab.srv-lab.server.home/homelab/iac/operators/rq-operator

# plugins=go/v4 is required for Apple Silicon
operator-sdk create api --group homelab --version v1alpha1 --kind ResourceQuotaConfig --resource --controller --plugins=go/v4 --namespaced=false

# Webhook
operator-sdk create webhook --group homelab --version v1alpha1 --kind ResourceQuotaConfig --defaulting
```
