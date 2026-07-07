---
title: Reconfigure TLS of Neo4j
menu:
  docs_{{ .version }}:
    identifier: neo4j-reconfigure-tls-cluster
    name: Cluster
    parent: neo4j-reconfigure-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS in Neo4j

KubeDB supports TLS reconfiguration for existing `Neo4j` databases through `Neo4jOpsRequest`. You can add TLS, rotate certificates, change issuer, and remove TLS without recreating the database.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to communicate with the cluster.
- Install `cert-manager` to manage certificates.
- Install KubeDB operator following [the setup guide](/docs/setup/README.md).
- This guide uses namespace `demo`.

```bash
kubectl create ns demo
```
namespace/demo created

This guide assumes a running Neo4j database named `tls-neo4j` in namespace `demo`.

## Create Issuer for TLS Certificates

If you already have an `Issuer`/`ClusterIssuer`, you can skip this section.

Generate a CA certificate and private key:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ca.key -out ca.crt -subj "/CN=neo4j-ca/O=kubedb"
```
Generating a RSA private key
...+++++
...+++++
writing new private key to 'ca.key'
-----

Create a TLS secret from the generated CA files:

```bash
kubectl create secret tls neo4j-ca --cert=ca.crt --key=ca.key -n demo
```
secret/neo4j-ca created

Create an `Issuer` using that secret:

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

```bash
kubectl get issuer -n demo neo4j-ca-issuer
```
NAME              READY   AGE
neo4j-ca-issuer   True    10s

## Add TLS to Neo4j

Use this `Neo4jOpsRequest` to add TLS (or replace issuer settings) for the target database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-ca-issuer
```

Here,

- `spec.type: ReconfigureTLS` selects TLS reconfiguration operation.
- `spec.databaseRef.name` selects the target `Neo4j` database.
- `spec.tls.issuerRef` defines which issuer should sign/re-issue certificates.

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-ca-issuer
EOF
```
neo4jopsrequest.ops.kubedb.com/neo4j-add-tls created

```bash
kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/neo4j-add-tls -n demo --timeout=600s
```
neo4jopsrequest.ops.kubedb.com/neo4j-add-tls condition met

Verify request status, generated certs, and encrypted connection via `neo4j+s`:

```bash
kubectl get neo4jopsrequest -n demo neo4j-add-tls
```
NAME            TYPE             STATUS       AGE
neo4j-add-tls   ReconfigureTLS   Successful   2m

```bash
kubectl get secret -n demo tls-neo4j-server-cert
```
NAME                   TYPE                DATA   AGE
tls-neo4j-server-cert  kubernetes.io/tls   3      2m

```bash
kubectl get secret -n demo tls-neo4j-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
```
notBefore=May 15 07:10:28 2026 GMT
notAfter=Aug 13 07:10:28 2026 GMT

```bash
kubectl get secret -n demo tls-neo4j-auth -o jsonpath='{.data.password}' | base64 -d && echo
```
8pyn3zno7QbX4iQn

```bash
kubectl exec -it -n demo tls-neo4j-0 -- cypher-shell -a neo4j+s://tls-neo4j-0.demo.svc.cluster.local:7687 -u neo4j -p '8pyn3zno7QbX4iQn'
```
Connected to Neo4j using Bolt protocol version 6.0 at neo4j+s://tls-neo4j-0.demo.svc.cluster.local:7687 as user neo4j.
Type :help for a list of available commands or :exit to exit the shell.
Note that Cypher queries must end with a semicolon.

## Rotate TLS Certificates

Before rotation, check current certificate validity window:

```bash
kubectl get secret -n demo tls-neo4j-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
```
notBefore=May 15 07:10:28 2026 GMT
notAfter=Aug 13 07:10:28 2026 GMT

Now rotate certificates and update Bolt TLS mode:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-reconfigure-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    rotateCertificates: true
    bolt:
      mode: mTLS
```

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-reconfigure-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    rotateCertificates: true
    bolt:
      mode: mTLS
EOF
```
neo4jopsrequest.ops.kubedb.com/neo4j-reconfigure-tls created

```bash
kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/neo4j-reconfigure-tls -n demo --timeout=600s
```
neo4jopsrequest.ops.kubedb.com/neo4j-reconfigure-tls condition met

```bash
kubectl get neo4jopsrequest -n demo neo4j-reconfigure-tls
```
NAME                    TYPE             STATUS       AGE
neo4j-reconfigure-tls   ReconfigureTLS   Successful   2m

```bash
kubectl get secret -n demo tls-neo4j-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
```
notBefore=May 15 07:24:11 2026 GMT
notAfter=Aug 13 07:24:11 2026 GMT

The changed certificate timestamp confirms rotation happened successfully.

## Change Issuer for Existing TLS

Create a new CA and issuer:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout new-ca.key -out new-ca.crt -subj "/CN=neo4j-ca-updated/O=kubedb-updated"
```
Generating a RSA private key
...+++++
...+++++
writing new private key to 'new-ca.key'
-----

```bash
kubectl create secret tls neo4j-new-ca --cert=new-ca.crt --key=new-ca.key -n demo
```
secret/neo4j-new-ca created

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: neo4j-new-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: neo4j-new-ca
EOF
```
issuer.cert-manager.io/neo4j-new-ca-issuer created

Apply an OpsRequest that points to the new issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-new-ca-issuer
```

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: neo4j-new-ca-issuer
EOF
```
neo4jopsrequest.ops.kubedb.com/neo4j-change-issuer created

```bash
kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/neo4j-change-issuer -n demo --timeout=600s
```
neo4jopsrequest.ops.kubedb.com/neo4j-change-issuer condition met

```bash
kubectl get secret -n demo tls-neo4j-server-cert -o jsonpath='{.data.ca\.crt}' | base64 -d | openssl x509 -noout -subject
```
subject=CN = neo4j-ca-updated, O = kubedb-updated

## Remove TLS from Neo4j

Use this request to disable TLS from the target database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    remove: true
```

```bash
cat <<'EOF' | kubectl apply -f -
apiVersion: ops.kubedb.com/v1alpha1
kind: Neo4jOpsRequest
metadata:
  name: neo4j-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: tls-neo4j
  tls:
    remove: true
EOF
```
neo4jopsrequest.ops.kubedb.com/neo4j-remove-tls created

```bash
kubectl wait --for=jsonpath='{.status.phase}'=Successful neo4jopsrequest/neo4j-remove-tls -n demo --timeout=600s
```
neo4jopsrequest.ops.kubedb.com/neo4j-remove-tls condition met

```bash
kubectl get neo4jopsrequest -n demo neo4j-remove-tls
```
NAME               TYPE             STATUS       AGE
neo4j-remove-tls   ReconfigureTLS   Successful   1m

## Verify All Requests

```bash
kubectl get neo4jopsrequest -n demo neo4j-reconfigure-tls
```
NAME                    TYPE             STATUS       AGE
neo4j-reconfigure-tls   ReconfigureTLS   Successful   2m

```bash
kubectl get neo4jopsrequest -n demo neo4j-remove-tls
```
NAME               TYPE             STATUS       AGE
neo4j-remove-tls   ReconfigureTLS   Successful   1m

```bash
kubectl get neo4jopsrequest -n demo neo4j-add-tls
```
NAME            TYPE             STATUS       AGE
neo4j-add-tls   ReconfigureTLS   Successful   1m

```bash
kubectl get neo4jopsrequest -n demo neo4j-change-issuer
```
NAME                 TYPE             STATUS       AGE
neo4j-change-issuer  ReconfigureTLS   Successful   1m

## Cleaning up

```bash
kubectl delete neo4jopsrequest -n demo neo4j-add-tls neo4j-reconfigure-tls neo4j-change-issuer neo4j-remove-tls
```
neo4jopsrequest.ops.kubedb.com "neo4j-add-tls" deleted
neo4jopsrequest.ops.kubedb.com "neo4j-reconfigure-tls" deleted
neo4jopsrequest.ops.kubedb.com "neo4j-change-issuer" deleted
neo4jopsrequest.ops.kubedb.com "neo4j-remove-tls" deleted

## Next Steps

- Learn `Neo4jOpsRequest` fields in detail from [Neo4j OpsRequest concept](/docs/guides/neo4j/concepts/opsrequest.md).

