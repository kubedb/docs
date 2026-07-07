---
title: Configure TLS for Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-tls-configure
    name: Configure TLS
    parent: weaviate-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS for Weaviate

This tutorial will show you how to provision a Weaviate cluster with TLS enabled from the start, using KubeDB and cert-manager.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- Install [cert-manager](https://cert-manager.io/docs/installation/) in your cluster — Weaviate TLS is issued through cert-manager.

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Weaviate TLS Overview](/docs/guides/weaviate/tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/tls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/tls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create an Issuer

Weaviate TLS is issued through cert-manager. First, create a self-signed CA and a `Secret` that holds it:

```bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout weaviate-ca.key -out weaviate-ca.crt -subj "/CN=weaviate-ca"
```

```bash
kubectl create secret tls weaviate-ca \
    --cert=weaviate-ca.crt --key=weaviate-ca.key -n demo
```
secret/weaviate-ca created

Now, create an `Issuer` named `weaviate-issuer` that references this CA secret:

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

## Deploy Weaviate with TLS

Now, create a `Weaviate` CR with `spec.tls` referencing the issuer. Setting `clientAuth: true` requires clients to present a valid client certificate (mutual TLS):

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: weaviate-issuer
    clientAuth: true
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/tls/tls.yaml
```
weaviate.kubedb.com/weaviate-sample created

Wait until the cluster becomes `Ready`:

```bash
kubectl get weaviate -n demo
```
NAME              TYPE                  VERSION   STATUS   AGE
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Ready    73s

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
```
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          71s
weaviate-sample-1   1/1     Running   0          59s
weaviate-sample-2   1/1     Running   0          45s

## Verify TLS Resources

KubeDB created cert-manager `Certificate` resources and the corresponding TLS secrets (a `server` and a `client` certificate):

```bash
kubectl get certificate -n demo
```
NAME                          READY   SECRET                        AGE
weaviate-sample-client-cert   True    weaviate-sample-client-cert   73s
weaviate-sample-server-cert   True    weaviate-sample-server-cert   73s

```bash
kubectl get secret -n demo | grep weaviate-sample
```
weaviate-sample-auth          Opaque              3      73s
weaviate-sample-client-cert   kubernetes.io/tls   4      73s
weaviate-sample-d25c86        Opaque              1      73s
weaviate-sample-server-cert   kubernetes.io/tls   3      73s

The `spec.tls` block on the `Weaviate` object reflects the TLS configuration:

```bash
kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.tls}' | jq
```
{
  "certificates": [
    {"alias": "server", "secretName": "weaviate-sample-server-cert"},
    {"alias": "client", "secretName": "weaviate-sample-client-cert"}
  ],
  "clientAuth": true,
  "issuerRef": {
    "apiGroup": "cert-manager.io",
    "kind": "Issuer",
    "name": "weaviate-issuer"
  }
}

With TLS enabled, the REST service is served over HTTPS on port `8443` instead of plain HTTP on `8080`:

```bash
kubectl get svc -n demo weaviate-sample -o jsonpath='{range .spec.ports[*]}{.name}={.port} {end}'
```
https=8443 grpc=50051 gossip=7102 data=7103 raft=8300

You can inspect the issued server certificate:

```bash
kubectl get secret -n demo weaviate-sample-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -subject -issuer -dates
```
subject=CN=weaviate-sample
issuer=CN=weaviate-ca
notBefore=Jun 30 18:07:40 2026 GMT
notAfter=Sep 28 18:07:40 2026 GMT

## Connect over TLS

Because `clientAuth` is enabled, clients must present the client certificate. Extract the certificates from the `client` secret and connect through a port-forward:

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
export WEAVIATE_API_KEY=$(kubectl get secret -n demo weaviate-sample-auth -o jsonpath='{.data.AUTHENTICATION_APIKEY_ALLOWED_KEYS}' | base64 -d)
```

```bash
kubectl port-forward -n demo svc/weaviate-sample 8443:8443
```

# in another terminal
```bash
curl -s -o /dev/null -w "%{http_code}\n" \
    --cacert ca.crt --cert client.crt --key client.key \
    https://localhost:8443/v1/.well-known/ready \
    -H "Authorization: Bearer $WEAVIATE_API_KEY"
```
200

```bash
curl -s --cacert ca.crt --cert client.crt --key client.key \
    https://localhost:8443/v1/nodes \
    -H "Authorization: Bearer $WEAVIATE_API_KEY" | jq '.nodes[] | {name, status}'
```
{"name": "weaviate-sample-0", "status": "HEALTHY"}
{"name": "weaviate-sample-1", "status": "HEALTHY"}
{"name": "weaviate-sample-2", "status": "HEALTHY"}

All three nodes are reachable over the TLS-encrypted, mutually-authenticated REST endpoint.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Add, rotate, or remove TLS on a running cluster: [Reconfigure TLS](/docs/guides/weaviate/reconfigure-tls/reconfigure-tls.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
```

```bash
kubectl delete issuer -n demo weaviate-issuer
```

```bash
kubectl delete secret -n demo weaviate-ca
```

```bash
kubectl delete ns demo
```
