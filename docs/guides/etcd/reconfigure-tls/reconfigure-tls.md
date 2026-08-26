---
title: Reconfigure Etcd TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: etcd-reconfigure-tls-guide
    name: Reconfigure TLS/SSL Encryption
    parent: etcd-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Etcd TLS/SSL (Transport Encryption)

KubeDB supports adding, removing, updating and rotating the TLS configuration of a running `Etcd` cluster through an `EtcdOpsRequest`. This tutorial walks through all four.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manager`](https://cert-manager.io/docs/installation/) v1.0.0 or later in your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` Provisioner and Ops-manager operators in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the operators were installed with `--set featureGates.Etcd=true`.

- You should be familiar with the following `KubeDB` concepts:
    - [Etcd](/docs/guides/etcd/concepts/etcd.md)
    - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
    - [Reconfigure TLS Overview](/docs/guides/etcd/reconfigure-tls/overview.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to an Etcd cluster

Here, we are going to deploy an `Etcd` cluster **without** TLS and then add TLS to it with an `EtcdOpsRequest`.

### Deploy Etcd without TLS

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-cluster
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcd.yaml
etcd.kubedb.com/etcd-cluster created
```

Wait until the cluster is `Ready`:

```bash
$ kubectl get etcd -n demo etcd-cluster
NAME           VERSION   STATUS   AGE
etcd-cluster   3.6.4     Ready    3m
```

Verify that TLS is off — every advertised URL uses the `http` scheme, and there are no `--cert-file`/`--peer-cert-file` flags:

```bash
$ kubectl get pod -n demo etcd-cluster-0 -o jsonpath='{.spec.containers[0].args}' | tr ',' '\n'
["--name=$(POD_NAME)"
"--data-dir=/var/lib/etcd/data"
"--initial-advertise-peer-urls=$(ETCD_MEMBER_PEER_URL)"
"--listen-peer-urls=http://0.0.0.0:2380"
"--listen-client-urls=http://0.0.0.0:2379"
"--advertise-client-urls=$(ETCD_MEMBER_CLIENT_URL)"
"--listen-metrics-urls=http://0.0.0.0:2381"]
```

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- etcdctl --endpoints=http://127.0.0.1:2379 endpoint health
http://127.0.0.1:2379 is healthy: successfully committed proposal: took = 1.9ms
```

### Create Issuer/ClusterIssuer

Now we need an `Issuer` for cert-manager to sign the etcd certificates with.

- Generate a CA certificate:

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=etcd/O=kubedb"
```

- Create a secret from it:

```bash
$ kubectl create secret tls etcd-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/etcd-ca created
```

- Create the `Issuer`:

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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcd-issuer.yaml
issuer.cert-manager.io/etcd-ca-issuer created
```

### Create EtcdOpsRequest

To add TLS, create an `EtcdOpsRequest` of type `ReconfigureTLS` that names the issuer:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-cluster
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: etcd-ca-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb
          organizationalUnits:
            - server
```

Here,

- `spec.databaseRef.name` refers to the `Etcd` object we want to reconfigure.
- `spec.type: ReconfigureTLS` tells the Ops-manager operator that this request is about TLS.
- `spec.tls.issuerRef` names the `Issuer` to sign the certificates. Supplying it is what adds TLS to a cluster that has none.
- `spec.tls.certificates` is optional. Here it only customises the subject of the `server` certificate; the `client` and `peer` aliases get the defaults. You do not need to list an alias just to have it issued.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcdops-add-tls.yaml
etcdopsrequest.ops.kubedb.com/etcdops-add-tls created
```

### Verify TLS enabled successfully

Let's watch the ops request. It moves `Pending` → `Progressing` → `Successful`:

```bash
$ watch kubectl get etcdopsrequest -n demo etcdops-add-tls
NAME              TYPE             STATUS       AGE
etcdops-add-tls   ReconfigureTLS   Successful   4m
```

`kubectl describe` shows the individual steps as conditions. The condition types come straight from the operator's state machine, so they are a reliable way to see where a long-running request currently is:

```bash
$ kubectl describe etcdopsrequest -n demo etcdops-add-tls
Name:         etcdops-add-tls
Namespace:    demo
Kind:         EtcdOpsRequest
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   etcd-cluster
  Type:     ReconfigureTLS
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       etcd-ca-issuer
Status:
  Conditions:
    Type:                  Running
    Status:                True
    Reason:                Running
    Message:               Etcd ops request is reconfiguring TLS

    Type:                  CertificateSynced
    Status:                True
    Reason:                CertificateSynced
    Message:               Successfully synced all the etcd certificates

    Type:                  UpdateEtcdPetSet
    Status:                True
    Reason:                UpdateEtcdPetSet
    Message:               Successfully re-rendered the petset with the new TLS configuration

    Type:                  RestartEtcdPods
    Status:                True
    Reason:                RestartEtcdPods
    Message:               Successfully restarted the etcd members with the new TLS configuration

    Type:                  UpdateDatabase
    Status:                True
    Reason:                UpdateDatabase
    Message:               Successfully updated the Etcd TLS configuration

    Type:                  Successful
    Status:                True
    Reason:                Successful
    Message:               Successfully reconfigured the etcd TLS configuration
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason           Message
  ----    ------           -------
  Normal  PauseDatabase    Pausing Etcd demo/etcd-cluster
  Normal  Successful       Successfully restarted the etcd members with the new TLS configuration
  Normal  ResumeDatabase   Successfully resumed Etcd demo/etcd-cluster
  Normal  Successful       Successfully reconfigured the etcd TLS configuration
```

While the request is running, the `Etcd` object itself carries a `Paused` condition — that is the operator's way of telling the Provisioner operator to keep its hands off the `PetSet`. It is removed again when the request finishes.

The certificates now exist:

```bash
$ kubectl get certificate -n demo
NAME                        READY   SECRET                      AGE
etcd-cluster-client-cert    True    etcd-cluster-client-cert    4m
etcd-cluster-peer-cert      True    etcd-cluster-peer-cert      4m
etcd-cluster-server-cert    True    etcd-cluster-server-cert    4m
```

And `spec.tls` has been written back onto the `Etcd` object:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.tls.issuerRef}'
{"apiGroup":"cert-manager.io","kind":"Issuer","name":"etcd-ca-issuer"}
```

The members have been re-rendered with the TLS flags and the schemes have flipped to `https`:

```bash
$ kubectl get pod -n demo etcd-cluster-0 -o jsonpath='{.spec.containers[0].args}' | tr ',' '\n'
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

Note that `--listen-metrics-urls` stays on plain `http` at port `2381` — that dedicated listener is what keeps the kubelet readiness probe working now that the client port demands a client certificate.

Finally, confirm from inside a member that only a mutually authenticated client gets through:

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint health
127.0.0.1:2379 is unhealthy: failed to commit proposal: context deadline exceeded
Error: unhealthy cluster

$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=https://127.0.0.1:2379 \
      --cacert=/var/run/etcd/tls/client/ca.crt \
      --cert=/var/run/etcd/tls/client/tls.crt \
      --key=/var/run/etcd/tls/client/tls.key \
      endpoint health
https://127.0.0.1:2379 is healthy: successfully committed proposal: took = 2.0ms
```

And that all three members are back in the quorum after the rolling restart:

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=https://etcd-cluster.demo.svc:2379 \
      --cacert=/var/run/etcd/tls/client/ca.crt \
      --cert=/var/run/etcd/tls/client/tls.crt \
      --key=/var/run/etcd/tls/client/tls.key \
      endpoint health --cluster -w table
+----------------------------------------------------------+--------+-------+-------+
|                         ENDPOINT                         | HEALTH | TOOK  | ERROR |
+----------------------------------------------------------+--------+-------+-------+
| https://etcd-cluster-0.etcd-cluster-pods.demo.svc...:2379 |   true | 2.1ms |       |
| https://etcd-cluster-1.etcd-cluster-pods.demo.svc...:2379 |   true | 2.3ms |       |
| https://etcd-cluster-2.etcd-cluster-pods.demo.svc...:2379 |   true | 1.8ms |       |
+----------------------------------------------------------+--------+-------+-------+
```

## Rotate Certificates

Rotating asks cert-manager to re-issue every existing certificate with the same issuer and the same specs. Use it when a certificate is nearing expiry, or when you suspect a key has been exposed.

First, note the current expiry so you can tell the rotation apart afterwards:

```bash
$ kubectl get secret -n demo etcd-cluster-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
notBefore=Aug 15 06:12:00 2026 GMT
notAfter=Nov 13 06:12:00 2026 GMT
```

### Create EtcdOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-cluster
  tls:
    rotateCertificates: true
```

Here, `spec.tls.rotateCertificates: true` is the **only** TLS field set. Combining it with `issuerRef`, `certificates` or `remove` is rejected by the webhook with `only one TLS reconfiguration operation is allowed at a time`. It also requires TLS to already be enabled — otherwise the webhook rejects it with `rotateCertificates requires TLS to already be enabled with issuerRef on Etcd`.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcdops-rotate.yaml
etcdopsrequest.ops.kubedb.com/etcdops-rotate created
```

### Verify certificates rotated successfully

```bash
$ watch kubectl get etcdopsrequest -n demo etcdops-rotate
NAME             TYPE             STATUS       AGE
etcdops-rotate   ReconfigureTLS   Successful   3m
```

A rotation adds one extra step compared to adding TLS — the operator has to put cert-manager's `Issuing` condition on every `Certificate` first, which is recorded as the `IssueCertificatesSucceeded` condition:

```bash
$ kubectl get etcdopsrequest -n demo etcdops-rotate -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
Running                       True
IssueCertificatesSucceeded    True
CertificateSynced             True
UpdateEtcdPetSet              True
RestartEtcdPods               True
UpdateDatabase                True
Successful                    True
```

The certificate now carries a new validity window:

```bash
$ kubectl get secret -n demo etcd-cluster-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -dates
notBefore=Aug 15 06:41:00 2026 GMT
notAfter=Nov 13 06:41:00 2026 GMT
```

Since the issuer — and therefore the CA — did not change, existing clients that trust the CA keep working across the rotation. Only the leaf certificates changed.

## Change Issuer/ClusterIssuer

You can migrate the cluster onto a different CA by pointing `spec.tls.issuerRef` at another `Issuer`/`ClusterIssuer`. This is the same "issue/update" operation as adding TLS.

### Create a new Issuer

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb"

$ kubectl create secret tls etcd-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/etcd-new-ca created
```

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: etcd-new-issuer
  namespace: demo
spec:
  ca:
    secretName: etcd-new-ca
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcd-new-issuer.yaml
issuer.cert-manager.io/etcd-new-issuer created
```

### Create EtcdOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-cluster
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: etcd-new-issuer
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcdops-update-issuer.yaml
etcdopsrequest.ops.kubedb.com/etcdops-update-issuer created
```

### Verify the issuer changed successfully

```bash
$ watch kubectl get etcdopsrequest -n demo etcdops-update-issuer
NAME                    TYPE             STATUS       AGE
etcdops-update-issuer   ReconfigureTLS   Successful   5m
```

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.tls.issuerRef}'
{"apiGroup":"cert-manager.io","kind":"Issuer","name":"etcd-new-issuer"}
```

Check the issuer recorded on a certificate secret:

```bash
$ kubectl get secret -n demo etcd-cluster-server-cert -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -noout -issuer
issuer=CN = ca-updated, O = kubedb
```

> **Plan for the CA change.** Every member is restarted with the new CA in its `--trusted-ca-file`/`--peer-trusted-ca-file`, so the cluster ends up internally consistent. External clients, however, still trust the *old* CA and will fail to verify the server after this completes. Distribute the new `ca.crt` to your applications as part of the same change window. Within a single request the members are rolled one at a time with quorum checks in between, so the cluster itself stays available.

## Remove TLS from the cluster

To turn TLS off entirely, set `remove: true`.

### Create EtcdOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcdops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: etcd-cluster
  tls:
    remove: true
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure-tls/etcdops-remove.yaml
etcdopsrequest.ops.kubedb.com/etcdops-remove-tls created
```

### Verify TLS removed successfully

```bash
$ watch kubectl get etcdopsrequest -n demo etcdops-remove-tls
NAME                 TYPE             STATUS       AGE
etcdops-remove-tls   ReconfigureTLS   Successful   4m
```

Removing TLS skips the certificate-issuing steps entirely — there is nothing to sync — so the condition list is shorter:

```bash
$ kubectl get etcdopsrequest -n demo etcdops-remove-tls -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
Running               True
UpdateEtcdPetSet      True
RestartEtcdPods       True
UpdateDatabase        True
Successful            True
```

`spec.tls` is gone from the `Etcd` object:

```bash
$ kubectl get etcd -n demo etcd-cluster -o jsonpath='{.spec.tls}'

```

The now-orphaned `Certificate` objects are cleaned up by the operator:

```bash
$ kubectl get certificate -n demo
No resources found in demo namespace.
```

And the members are back on plain `http`:

```bash
$ kubectl exec -it -n demo etcd-cluster-0 -c etcd -- \
    etcdctl --endpoints=http://127.0.0.1:2379 endpoint health
http://127.0.0.1:2379 is healthy: successfully committed proposal: took = 1.9ms
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete etcdopsrequest -n demo etcdops-add-tls etcdops-rotate etcdops-update-issuer etcdops-remove-tls
$ kubectl delete etcd -n demo etcd-cluster
$ kubectl delete issuer -n demo etcd-ca-issuer etcd-new-issuer
$ kubectl delete secret -n demo etcd-ca etcd-new-ca
$ kubectl delete ns demo
```

## Next Steps

- Deploy an etcd cluster with TLS from the start: [Run Etcd with TLS/SSL](/docs/guides/etcd/tls/configure-ssl.md).
- Monitor your etcd cluster with [builtin Prometheus](/docs/guides/etcd/monitoring/using-builtin-prometheus.md) or the [Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).
- Detail concepts of [EtcdOpsRequest object](/docs/guides/etcd/concepts/etcdopsrequest.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
