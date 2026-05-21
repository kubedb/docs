---
title: Configure TLS in Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-tls-description
    name: Configure TLS
    parent: qdrant-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Qdrant with TLS (Transport Encryption)

KubeDB supports providing TLS encryption for Qdrant. This tutorial will show you how to use KubeDB to run a Qdrant cluster with TLS encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.4.0 or later to your cluster to manage your TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [TLS Overview](/docs/guides/qdrant/tls/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/tls](/docs/examples/qdrant/tls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Create Issuer/ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial to enable TLS in Qdrant. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating your CA certificates using openssl:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=qdrant/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated:

```bash
$ kubectl create secret tls qdrant-ca --cert=ca.crt --key=ca.key --namespace=demo
secret/qdrant-ca created
```

Now, create an `Issuer` using the `qdrant-ca` secret you have just created. Below is the YAML of the `Issuer` CR:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: qdrant-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: qdrant-ca
```

Let's create the `Issuer` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/tls/issuer.yaml
issuer.cert-manager.io/qdrant-ca-issuer created
```

## TLS Encryption in Qdrant

Below is the YAML for the Qdrant cluster with TLS enabled:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-tls
  namespace: demo
spec:
  version: "1.17.0"
  mode: Distributed
  replicas: 3
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: qdrant-ca-issuer
      kind: Issuer
    client: true
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,
- `spec.tls.issuerRef` refers to the `qdrant-ca-issuer` issuer that we created in the previous step.
- `spec.tls.client` (optional, default `false`): Enables TLS for client-to-server communication. When set to `true`, clients must connect using TLS.

### Deploy Qdrant Cluster

Let's create the `Qdrant` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/tls/tls-qdrant.yaml
qdrant.kubedb.com/qdrant-tls created
```

Now, wait until `qdrant-tls` has status `Ready`:

```bash
$ watch -n 3 kubectl get qdrant -n demo qdrant-tls
Every 3.0s: kubectl get qdrant -n demo qdrant-tls

NAME          VERSION   STATUS   AGE
qdrant-tls    1.17.0    Ready    7m

$ watch -n 3 kubectl get pods -n demo -l app.kubernetes.io/instance=qdrant-tls
Every 3.0s: kubectl get pods -n demo -l app.kubernetes.io/instance=qdrant-tls

NAME              READY   STATUS    RESTARTS   AGE
qdrant-tls-0      1/1     Running   0          7m
qdrant-tls-1      1/1     Running   0          2m
qdrant-tls-2      1/1     Running   0          117s
```

### Verify TLS Configuration

Now, let's verify the TLS certificates were created for the Qdrant database:

```bash
$ kubectl get secrets -n demo | grep qdrant-tls
qdrant-tls-160bbc          Opaque                     1      7m
qdrant-tls-auth            Opaque                     2      7m
qdrant-tls-client-cert     kubernetes.io/tls          4      7m
qdrant-tls-server-cert     kubernetes.io/tls          3      7m
```

The `qdrant-tls-client-cert` secret contains the client TLS certificate. Let's inspect it:

```bash
$ kubectl describe secret -n demo qdrant-tls-client-cert
Name:         qdrant-tls-client-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=qdrant-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=qdrants.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
              cert-manager.io/certificate-name: qdrant-tls-client-cert
              cert-manager.io/common-name: qdrant
              cert-manager.io/ip-sans:
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: qdrant-ca-issuer
              cert-manager.io/uri-sans:

Type:  kubernetes.io/tls

Data
====
ca.crt:            1151 bytes
tls-combined.pem:  2811 bytes
tls.crt:           1131 bytes
tls.key:           1679 bytes
```

We can also verify that the TLS configuration has been applied inside the Qdrant pod:

```bash
$ kubectl exec -n demo qdrant-tls-0 -- cat /qdrant/config/config.yaml
Defaulted container "qdrant" out of: qdrant, update-raft-state (init)
cluster:
  enabled: true
  p2p:
    port: 6335
log_level: INFO
service:
  enable_tls: true
  verify_https_client_certificate: true
tls:
  ca_cert: /tls/ca.pem
  cert: /tls/cert.pem
  key: /tls/key.pem

$ kubectl exec -n demo qdrant-tls-0 -- ls /tls/
Defaulted container "qdrant" out of: qdrant, update-raft-state (init)
ca.crt
ca.pem
cert.pem
client.crt
client.key
key.pem
```

The TLS certificates are mounted at `/tls/` inside the container, and the Qdrant config shows `service.enable_tls: true`.

### Connect to Qdrant with TLS

Extract the CA certificate, client certificate, and client key from the secret to your local machine:

```bash
kubectl get secret -n demo qdrant-tls-client-cert -o jsonpath='{.data.ca\.crt}' | base64 -d > ca.crt
kubectl get secret -n demo qdrant-tls-client-cert -o jsonpath='{.data.tls\.crt}' | base64 -d > tls.crt
kubectl get secret -n demo qdrant-tls-client-cert -o jsonpath='{.data.tls\.key}' | base64 -d > tls.key
```

Then, port-forward the Qdrant service and connect using TLS:

```bash
$ kubectl port-forward -n demo svc/qdrant-tls 6333:6333 &
Forwarding from 127.0.0.1:6333 -> 6333
```

Get the API key from the auth secret:

```bash
$ kubectl get secret -n demo qdrant-tls-auth -o jsonpath='{.data.api-key}' | base64 -d
GuBrzentGdAcZuqh
```

Now you can connect to the Qdrant cluster using TLS:

```bash
$ curl --cacert ca.crt --cert tls.crt --key tls.key -H "api-key: GuBrzentGdAcZuqh" \
  'https://localhost:6333/collections'
{"result":{"collections":[{"name":"KubeDBHealthCheckCollection"}]},"status":"ok","time":3.63e-6}
```

> Without the TLS certificates or the API key, the connection will be rejected.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-tls
kubectl delete issuer -n demo qdrant-ca-issuer
kubectl delete secret -n demo qdrant-ca
rm ca.crt tls.crt tls.key
```

## Next Steps

- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
