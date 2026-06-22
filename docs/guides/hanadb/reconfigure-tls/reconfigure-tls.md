---
title: Reconfigure HanaDB TLS
menu:
  docs_{{ .version }}:
    identifier: hanadb-reconfigure-tls-guide
    name: Reconfigure TLS
    parent: hanadb-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure HanaDB TLS

This guide shows how to add or rotate TLS for HanaDB using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- Install cert-manager.
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/reconfigure-tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure-tls).

## Create Issuer

Generate a CA certificate and key:

```bash
openssl req -x509 -nodes -days 365 \
  -newkey rsa:2048 \
  -keyout ca.key \
  -out ca.crt \
  -subj "/CN=HanaDB/O=kubedb"
```

Create the CA Secret:

```bash
kubectl create secret tls hdb-ca \
  --cert=ca.crt \
  --key=ca.key \
  --namespace=demo
```

Create an Issuer:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: hdb-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: hdb-ca
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure-tls/issuer.yaml
```

## Deploy HanaDB

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure-tls/hanadb-cluster.yaml
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Add TLS

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hanadb-cluster
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure-tls/hdbops-add-tls.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-add-tls --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Rotate TLS Certificates

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: hanadb-cluster
  tls:
    rotateCertificates: true
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure-tls/hdbops-rotate-tls.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-rotate-tls --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
```

## Cleanup

```bash
kubectl delete hdbops -n demo --all
kubectl delete hanadb -n demo hanadb-cluster
kubectl delete ns demo
```
