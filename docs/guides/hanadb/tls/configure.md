---
title: Configure HanaDB TLS
menu:
  docs_{{ .version }}:
    identifier: hanadb-tls-configure
    name: Configure TLS
    parent: hanadb-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS in HanaDB

This guide shows how to create a HanaDB System Replication cluster with TLS enabled from the beginning.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- Install cert-manager.
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/tls).

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/tls/issuer.yaml
```

## Deploy HanaDB with TLS

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-cluster
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 2
  storageType: Durable
  topology:
    mode: SystemReplication
    systemReplication:
      replicationMode: fullsync
      operationMode: logreplay_readaccess
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: hdb-ca-issuer
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/tls/hanadb-cluster.yaml
kubectl wait -n demo hanadb/hanadb-cluster --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Verify

Check the HanaDB object and generated certificate Secrets:

```bash
kubectl get hanadb -n demo hanadb-cluster
kubectl get secret -n demo -l app.kubernetes.io/instance=hanadb-cluster
```

## Cleanup

```bash
kubectl delete hanadb -n demo hanadb-cluster
kubectl delete issuer -n demo hdb-ca-issuer
kubectl delete ns demo
```
