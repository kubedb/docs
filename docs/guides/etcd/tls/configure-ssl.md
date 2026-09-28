---
title: Etcd TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: etcd-tls-configure
    name: Configure TLS/SSL
    parent: etcd-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Run Etcd Cluster with TLS/SSL

KubeDB supports providing TLS/SSL encryption for the `Etcd` client API, the member-to-member (Raft) channel and the metrics endpoint. This tutorial will show you how to use KubeDB to run an etcd cluster with TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.0.0 or later in your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the operator was installed with `--set featureGates.Etcd=true`.

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB uses the following CRD fields to enable SSL/TLS encryption in `Etcd`:

- `spec:`
    - `tls:`
        - `issuerRef`
        - `certificates`

Read about the fields in detail in the [Etcd Concept Guide](/docs/guides/etcd/concepts/etcd.md#spectls), and about what each certificate alias secures in the [TLS overview](/docs/guides/etcd/tls/overview.md).

There is no `enableSSL` boolean for etcd — setting `spec.tls` is what enables TLS, and `spec.tls.issuerRef` is required once you do. KubeDB then issues three certificates: `server` (client API), `client` (KubeDB's own connections) and `peer` (Raft). The `metrics-exporter` alias is accepted by the schema but nothing is ever issued for it, because etcd's metrics listener is always plain HTTP.

## Create Issuer/ClusterIssuer

We are going to create an example `Issuer` that will be used throughout this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating your CA certificate using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=etcd/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls etcd-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
```

Now, create an `Issuer` using the `etcd-ca` secret you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: etcd-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: etcd-ca
```

Apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/tls/etcd-issuer.yaml
issuer.cert-manager.io/etcd-ca-issuer created
```

## TLS/SSL encryption in an Etcd cluster

Below is the YAML for an `Etcd` object with TLS enabled:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-tls
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: etcd-ca-issuer
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,

- `spec.tls.issuerRef` refers to the `Issuer` that we created in the previous step. Its presence is what enables TLS.
- `spec.tls.certificates` is omitted, so KubeDB fills in the defaults for all required aliases. You only need to set it if you want to override a secret name, a subject, a duration or extra SANs.

### Deploy the Etcd cluster with TLS/SSL

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/tls/etcd-tls.yaml
etcd.kubedb.com/etcd-tls created
```

Now, wait until `etcd-tls` reaches the `Ready` state:

```bash
$ watch kubectl get etcd -n demo etcd-tls
NAME       VERSION   STATUS   AGE
etcd-tls   3.6.4     Ready    3m
```

### Verify that the certificates were issued

KubeDB creates one cert-manager `Certificate` object for each alias it actually uses — `server`, `client` and `peer`. No `metrics-exporter` certificate is issued, whether or not `spec.monitor` is set.

```bash
$ kubectl get certificate -n demo
NAME                    READY   SECRET                  AGE
etcd-tls-client-cert    True    etcd-tls-client-cert    3m
etcd-tls-peer-cert      True    etcd-tls-peer-cert      3m
etcd-tls-server-cert    True    etcd-tls-server-cert    3m
```

Each `Certificate` has a matching `kubernetes.io/tls` secret:

```bash
$ kubectl get secret -n demo -l app.kubernetes.io/instance=etcd-tls
NAME                   TYPE                       DATA   AGE
etcd-tls-auth          kubernetes.io/basic-auth   2      3m
etcd-tls-client-cert   kubernetes.io/tls          3      3m
etcd-tls-peer-cert     kubernetes.io/tls          3      3m
etcd-tls-server-cert   kubernetes.io/tls          3      3m
```

Let's describe the `server` certificate secret and check the SANs cert-manager recorded on it:

```bash
$ kubectl describe secret -n demo etcd-tls-server-cert
Name:         etcd-tls-server-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=etcd-tls
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=etcds.kubedb.com
              controller.cert-manager.io/fao=true
Annotations:  cert-manager.io/alt-names:
                *.etcd-tls-pods.demo.svc,*.etcd-tls-pods.demo.svc.cluster.local,etcd-tls,etcd-tls.demo,etcd-tls.demo.svc,etcd-tls.demo.svc.cluster.local,...
              cert-manager.io/certificate-name: etcd-tls-server-cert
              cert-manager.io/common-name: etcd-tls
              cert-manager.io/ip-sans: 127.0.0.1
              cert-manager.io/issuer-group: cert-manager.io
              cert-manager.io/issuer-kind: Issuer
              cert-manager.io/issuer-name: etcd-ca-issuer

Type:  kubernetes.io/tls

Data
====
ca.crt:   1159 bytes
tls.crt:  1493 bytes
tls.key:  1704 bytes
```

Notice the wildcard `*.etcd-tls-pods.demo.svc` entry — that is what lets a member present a valid certificate for its own ordinal DNS name when a peer dials it.

### Verify that etcd is actually serving over TLS

etcd speaks gRPC on port `2379`, not HTTP, so `openssl s_client` is not the right tool here the way it is for SQL databases — a successful TLS handshake would tell you nothing about whether etcd itself accepted the connection. Use `etcdctl` instead, from inside a member pod, where the certificates are already mounted.

First, confirm the flags the operator rendered onto the member:

```bash
$ kubectl get pod -n demo etcd-tls-0 -o jsonpath='{.spec.containers[0].args}' | tr ',' '\n'
["--name=$(POD_NAME)"
"--data-dir=/var/lib/etcd/data"
"--initial-advertise-peer-urls=$(ETCD_MEMBER_PEER_URL)"
"--listen-peer-urls=https://0.0.0.0:2380"
"--listen-client-urls=https://0.0.0.0:2379"
"--advertise-client-urls=$(ETCD_MEMBER_CLIENT_URL)"
"--listen-metrics-urls=http://0.0.0.0:2381"
"--cert-file=/var/run/etcd/tls/server/tls.crt"
"--key-file=/var/run/etcd/tls/server/tls.key"
"--trusted-ca-file=/var/run/etcd/tls/server/ca.crt"
"--client-cert-auth=true"
"--peer-cert-file=/var/run/etcd/tls/peer/tls.crt"
"--peer-key-file=/var/run/etcd/tls/peer/tls.key"
"--peer-trusted-ca-file=/var/run/etcd/tls/peer/ca.crt"
"--peer-client-cert-auth=true"]
```

The `https://` scheme on both the client and the peer listeners, plus `--client-cert-auth=true` and `--peer-client-cert-auth=true`, is the whole TLS story for etcd.

Now exec into a member and ask etcd itself:

```bash
$ kubectl exec -it -n demo etcd-tls-0 -c etcd -- sh
```

```bash
# TLS is on, so a plain (non-TLS) call must fail. This is the negative control.
$ etcdctl --endpoints=http://127.0.0.1:2379 endpoint health
{"level":"warn","msg":"retrying of unary invoker failed","error":"context deadline exceeded"}
127.0.0.1:2379 is unhealthy: failed to commit proposal: context deadline exceeded
Error: unhealthy cluster
```

```bash
# With the CA, the client certificate and its key, the same call succeeds.
$ etcdctl \
    --endpoints=https://127.0.0.1:2379 \
    --cacert=/var/run/etcd/tls/client/ca.crt \
    --cert=/var/run/etcd/tls/client/tls.crt \
    --key=/var/run/etcd/tls/client/tls.key \
    endpoint health
https://127.0.0.1:2379 is healthy: successfully committed proposal: took = 2.1ms
```

Because `--client-cert-auth=true` is set, trusting the CA alone is not enough. Dropping `--cert`/`--key` gets you a completed TLS handshake and then a rejected request — which is exactly the behaviour you want to confirm:

```bash
$ etcdctl \
    --endpoints=https://127.0.0.1:2379 \
    --cacert=/var/run/etcd/tls/client/ca.crt \
    endpoint health
{"level":"warn","msg":"retrying of unary invoker failed","error":"rpc error: code = Unavailable desc = connection error: desc = \"transport: authentication handshake failed: remote error: tls: bad certificate\""}
https://127.0.0.1:2379 is unhealthy: failed to commit proposal: context deadline exceeded
Error: unhealthy cluster
```

Finally, check the whole cluster rather than just the local member. Point `etcdctl` at the load-balanced client Service and ask for the per-member status — every member answering over `https://` means the Raft channel came up under mTLS too:

```bash
$ etcdctl \
    --endpoints=https://etcd-tls.demo.svc:2379 \
    --cacert=/var/run/etcd/tls/client/ca.crt \
    --cert=/var/run/etcd/tls/client/tls.crt \
    --key=/var/run/etcd/tls/client/tls.key \
    endpoint status --cluster -w table
+------------------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|                       ENDPOINT                       |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+------------------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| https://etcd-tls-0.etcd-tls-pods.demo.svc.cluster...  | 8e9e05c52164694d |   3.6.4 |   20 kB |      true |      false |         2 |         12 |                 12 |        |
| https://etcd-tls-1.etcd-tls-pods.demo.svc.cluster...  | 91bc3c398fb3c146 |   3.6.4 |   20 kB |     false |      false |         2 |         12 |                 12 |        |
| https://etcd-tls-2.etcd-tls-pods.demo.svc.cluster...  | fd422379fda50e48 |   3.6.4 |   20 kB |     false |      false |         2 |         12 |                 12 |        |
+------------------------------------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

You can also do a round-trip write and read to be sure the TLS session carries real traffic:

```bash
$ etcdctl \
    --endpoints=https://etcd-tls.demo.svc:2379 \
    --cacert=/var/run/etcd/tls/client/ca.crt \
    --cert=/var/run/etcd/tls/client/tls.crt \
    --key=/var/run/etcd/tls/client/tls.key \
    put /kubedb/hello world
OK

$ etcdctl \
    --endpoints=https://etcd-tls.demo.svc:2379 \
    --cacert=/var/run/etcd/tls/client/ca.crt \
    --cert=/var/run/etcd/tls/client/tls.crt \
    --key=/var/run/etcd/tls/client/tls.key \
    get /kubedb/hello
/kubedb/hello
world
```

> If you have enabled etcd RBAC on this cluster (which happens the first time you run a [`RotateAuth` EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)), add `--user=root:<password>` to the commands above. The password lives in the `<db-name>-auth` secret.

## Connecting from your own application

Your application should **not** reuse the `client` certificate: it is issued with `CN=root`, and etcd maps that Common Name onto the etcd `root` user. Issue a separate certificate for your workload from the same `Issuer` and mount it into your pod, then dial the client Service:

```
https://etcd-tls.demo.svc:2379
```

KubeDB also publishes an `AppBinding` named after the database, which records the client endpoint, the auth secret and — when TLS is on — the TLS secret, so operators and backup tooling can discover all of this without hard-coding it:

```bash
$ kubectl get appbinding -n demo etcd-tls -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: etcd-tls
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0t...
    insecureSkipTLSVerify: false
    service:
      name: etcd-tls
      port: 2379
      scheme: etcd
  secret:
    name: etcd-tls-auth
  tlsSecret:
    kind: Secret
    name: etcd-tls-client-cert
  type: kubedb.com/etcd
  version: 3.6.4
```

Here, `spec.clientConfig.caBundle` is the CA of the `client` certificate, copied out of the certificate secret so that a consumer can verify the server without reading the secret itself.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete etcd -n demo etcd-tls
$ kubectl delete issuer -n demo etcd-ca-issuer
$ kubectl delete secret -n demo etcd-ca
$ kubectl delete ns demo
```

## Next Steps

- Add, rotate or remove TLS on a **running** cluster with a [ReconfigureTLS EtcdOpsRequest](/docs/guides/etcd/reconfigure-tls/reconfigure-tls.md).
- Monitor your etcd cluster with [builtin Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md) or the [Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).
- Tune etcd at creation time with [custom configuration](/docs/guides/etcd/custom-configuration/using-config.md).
- Detail concepts of [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
