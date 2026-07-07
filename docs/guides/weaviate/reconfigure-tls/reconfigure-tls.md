---
title: Reconfigure Weaviate TLS
menu:
  docs_{{ .version }}:
    identifier: weaviate-reconfigure-tls-cluster
    name: Reconfigure TLS
    parent: weaviate-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Weaviate TLS (Transport Encryption)

This guide will show you how to use the `KubeDB` Ops Manager to add TLS to a running Weaviate cluster, rotate its certificates, update its issuer, and finally remove TLS.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- Install [cert-manager](https://cert-manager.io/docs/installation/) in your cluster — Weaviate TLS is issued through cert-manager.

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Reconfigure TLS Overview](/docs/guides/weaviate/reconfigure-tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/reconfigure-tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure-tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate without TLS

Deploy a Weaviate cluster without TLS and wait for it to become `Ready`. The REST service is served over plain HTTP on port `8080`:

```bash
kubectl get svc -n demo weaviate-sample -o jsonpath='{range .spec.ports[*]}{.name}={.port} {end}'
```
http=8080 grpc=50051 gossip=7102 data=7103 raft=8300

## Create an Issuer

Weaviate TLS is issued through cert-manager. First, create a CA secret and an `Issuer` named `weaviate-issuer` in the `demo` namespace:

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout weaviate-ca.key -out weaviate-ca.crt -subj "/CN=weaviate-ca"
```

```bash
kubectl create secret tls weaviate-ca \
    --cert=weaviate-ca.crt --key=weaviate-ca.key -n demo
```
secret/weaviate-ca created

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: weaviate-issuer
  namespace: demo
spec:
  ca:
    secretName: weaviate-ca
```

```bash
kubectl apply -f issuer.yaml
```
issuer.cert-manager.io/weaviate-issuer created

```bash
kubectl get issuer -n demo
```
NAME              READY   AGE
weaviate-issuer   True    3s

## Add TLS to the Cluster

Now, create a `ReconfigureTLS` OpsRequest that points at the issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: weaviate-sample
  tls:
    issuerRef:
      name: weaviate-issuer
      kind: Issuer
      apiGroup: cert-manager.io
  timeout: 5m
  apply: IfReady
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure-tls/add-tls.yaml
```
weaviateopsrequest.ops.kubedb.com/weaviate-add-tls created

The Ops Manager issues the certificates and restarts the pods.

```bash
kubectl get weaviateopsrequest -n demo weaviate-add-tls
```
NAME               TYPE             STATUS       AGE
weaviate-add-tls   ReconfigureTLS   Successful   2m

The `status.conditions` show the certificates being synced and the pods restarted:

```bash
kubectl get weaviateopsrequest -n demo weaviate-add-tls -o yaml
```
...
status:
  conditions:
  - message: Weaviate ops-request has started to reconfigure tls for Weaviate nodes
    reason: ReconfigureTLS
    status: "True"
    type: ReconfigureTLS
  - message: get certificate; ConditionStatus:True
    status: "True"
    type: GetCertificate
  - message: Successfully synced all certificates
    reason: CertificateSynced
    status: "True"
    type: CertificateSynced
  - message: successfully reconciled the Weaviate with tls configuration
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: Successfully restarted all nodes
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - message: Successfully completed reconfigureTLS for Weaviate.
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

Verify that the REST service now serves HTTPS on port `8443` and the certificates were created:

```bash
kubectl get svc -n demo weaviate-sample -o jsonpath='{range .spec.ports[*]}{.name}={.port} {end}'
```
https=8443 grpc=50051 gossip=7102 data=7103 raft=8300

```bash
kubectl get certificate -n demo
```
NAME                          READY   SECRET                        AGE
weaviate-sample-client-cert   True    weaviate-sample-client-cert   84s
weaviate-sample-server-cert   True    weaviate-sample-server-cert   84s

The cluster requires client certificate authentication (mTLS) by default. You can connect like this:

```bash
kubectl get secret -n demo weaviate-sample-client-cert -o jsonpath='{.data.ca\.crt}'  | base64 -d > ca.crt
```

```bash
kubectl get secret -n demo weaviate-sample-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > client.crt
```

```bash
kubectl get secret -n demo weaviate-sample-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > client.key
```

```bash
kubectl port-forward -n demo svc/weaviate-sample 8443:8443
```

# in another terminal
```bash
curl -s -o /dev/null -w "%{http_code}\n" --cacert ca.crt --cert client.crt --key client.key \
    https://localhost:8443/v1/.well-known/ready -H "Authorization: Bearer <api-key>"
```
200

## Rotate Certificates

To re-issue the certificates (for example, before they expire), create a `ReconfigureTLS` OpsRequest with `rotateCertificates: true`. This example also disables client-certificate authentication by setting `clientAuth: false`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: weaviate-sample
  tls:
    clientAuth: false
    rotateCertificates: true
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure-tls/rotate-certificate.yaml
```
weaviateopsrequest.ops.kubedb.com/wvops-rotate created

```bash
kubectl get weaviateopsrequest -n demo wvops-rotate
```
NAME           TYPE             STATUS       AGE
wvops-rotate   ReconfigureTLS   Successful   2m

Verify that the server certificate has a newer validity window:

```bash
kubectl get secret -n demo weaviate-sample-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -subject -dates
```
subject=CN=weaviate-sample
notBefore=Jun 30 17:52:42 2026 GMT
notAfter=Sep 28 17:52:42 2026 GMT

## Update the Issuer

You can switch the cluster to a different cert-manager issuer. First, create the new CA secret and the `weaviate-new-issuer`:

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout weaviate-new-ca.key -out weaviate-new-ca.crt -subj "/CN=weaviate-new-ca"
```

```bash
kubectl create secret tls weaviate-new-ca \
    --cert=weaviate-new-ca.crt --key=weaviate-new-ca.key -n demo
```
secret/weaviate-new-ca created

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: weaviate-new-issuer
  namespace: demo
spec:
  ca:
    secretName: weaviate-new-ca
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure-tls/weaviate-new-issuer.yaml
```
issuer.cert-manager.io/weaviate-new-issuer created

Now, create a `ReconfigureTLS` OpsRequest that points at the new issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: weaviate-sample
  tls:
    issuerRef:
      name: weaviate-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure-tls/update-issuer.yaml
```
weaviateopsrequest.ops.kubedb.com/wvops-update-issuer created

```bash
kubectl get weaviateopsrequest -n demo wvops-update-issuer
```
NAME                  TYPE             STATUS       AGE
wvops-update-issuer   ReconfigureTLS   Successful   2m

Verify that the server certificate is now signed by the new CA:

```bash
kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.tls.issuerRef.name}'
```
weaviate-new-issuer

```bash
kubectl get secret -n demo weaviate-sample-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -issuer
```
issuer=CN=weaviate-new-ca

## Remove TLS

Finally, to disable TLS, create a `ReconfigureTLS` OpsRequest with `remove: true`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: wvops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: weaviate-sample
  tls:
    remove: true
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure-tls/remove-tls.yaml
```
weaviateopsrequest.ops.kubedb.com/wvops-remove created

```bash
kubectl get weaviateopsrequest -n demo wvops-remove
```
NAME           TYPE             STATUS       AGE
wvops-remove   ReconfigureTLS   Successful   2m

Verify that the service is back to plain HTTP on port `8080`, the `spec.tls` field is cleared, and the certificate secrets are gone:

```bash
kubectl get svc -n demo weaviate-sample -o jsonpath='{range .spec.ports[*]}{.name}={.port} {end}'
```
http=8080 grpc=50051 gossip=7102 data=7103 raft=8300

```bash
kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.tls}'
```

```bash
kubectl get secret -n demo | grep weaviate-sample-.*cert
```
# (no cert secrets)

TLS has been added, rotated, re-issued with a new CA, and finally removed — all without recreating the database.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Provision a cluster with TLS from the start: [Weaviate TLS](/docs/guides/weaviate/tls/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviateopsrequest -n demo weaviate-add-tls wvops-rotate wvops-update-issuer wvops-remove
```

```bash
kubectl delete weaviate -n demo weaviate-sample
```

```bash
kubectl delete issuer -n demo weaviate-issuer weaviate-new-issuer
```

```bash
kubectl delete ns demo
```
