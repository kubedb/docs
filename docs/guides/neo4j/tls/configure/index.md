---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: neo4j-tls-configure
    name: Neo4j TLS/SSL Configuration
    parent: neo4j-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in Neo4j

`KubeDB` supports TLS/SSL for Neo4j client and intra-cluster communication. This guide shows how to deploy a TLS-enabled Neo4j cluster and verify encrypted connection using `neo4j+s`.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.4.0 or later.
- Install KubeDB following [the setup guide](/docs/setup/README.md).

```bash
kubectl create ns demo
```
namespace/demo created

### Create Issuer

Create a CA secret and an `Issuer` that KubeDB will use to generate Neo4j certificates.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ca.key -out ca.crt -subj "/CN=neo4j/O=kubedb"
```

```bash
kubectl create secret tls neo4j-ca --cert=ca.crt --key=ca.key -n demo
```
secret/neo4j-ca created

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: neo4j-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: neo4j-ca
EOF
```
issuer.cert-manager.io/neo4j-ca-issuer created

## Deploy Neo4j with TLS

Now apply a `Neo4j` object with TLS configuration:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: tls-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-ca-issuer
    bolt:
      mode: TLS
    cluster:
      mode: mTLS
  deletionPolicy: WipeOut
```

Here,

- `spec.tls.issuerRef` points to the certificate issuer.
- `spec.tls.bolt.mode: TLS` enables encrypted Bolt client traffic.
- `spec.tls.cluster.mode: mTLS` enables mutual TLS for inter-server traffic.

Apply the CR:

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: tls-neo4j
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-ca-issuer
    bolt:
      mode: TLS
    cluster:
      mode: mTLS
  deletionPolicy: WipeOut
EOF
```
neo4j.kubedb.com/tls-neo4j created

## Verify TLS

Check database readiness and generated TLS secret:

```bash
kubectl wait --for=condition=Ready neo4j/tls-neo4j -n demo --timeout=600s
```
neo4j.kubedb.com/tls-neo4j condition met

```bash
kubectl get secret -n demo tls-neo4j-server-cert
```
NAME                   TYPE                DATA   AGE
tls-neo4j-server-cert  kubernetes.io/tls   3      2m

Get Neo4j credentials:

```bash
kubectl get secret -n demo tls-neo4j-auth -o jsonpath='{.data.username}' | base64 -d && echo
```
neo4j

```bash
kubectl get secret -n demo tls-neo4j-auth -o jsonpath='{.data.password}' | base64 -d && echo
```
8pyn3zno7QbX4iQn

Connect using `neo4j+s` and verify TLS session:

```bash
kubectl exec -it -n demo tls-neo4j-0 -- bash
```
neo4j@tls-neo4j-0:~$ cypher-shell -a neo4j+s://tls-neo4j-0.demo.svc.cluster.local:7687 -u neo4j -p '8pyn3zno7QbX4iQn'
Connected to Neo4j using Bolt protocol version 6.0 at neo4j+s://tls-neo4j-0.demo.svc.cluster.local:7687 as user neo4j.
Type :help for a list of available commands or :exit to exit the shell.
Note that Cypher queries must end with a semicolon.

The successful `neo4j+s` connection confirms the cluster is accepting encrypted, certificate-validated Bolt traffic.

## Cleaning up

```bash
kubectl patch -n demo neo4j/tls-neo4j -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
```
neo4j.kubedb.com/tls-neo4j patched

```bash
kubectl delete -n demo neo4j/tls-neo4j
```
neo4j.kubedb.com "tls-neo4j" deleted

```bash
kubectl delete ns demo
```
namespace "demo" deleted

## Next Steps

- Learn Neo4j CR fields from [Neo4j concept doc](/docs/guides/neo4j/concepts/neo4j.md).

