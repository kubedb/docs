---
title: Run DocumentDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: dc-configuration-using-config-file
    name: Custom Configuration
    parent: dc-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run DocumentDB with Custom Configuration

KubeDB DocumentDB speaks the **MongoDB wire protocol** (port `10260`, TLS) on top of an
internal **PostgreSQL** storage engine (port `9712`, not exposed). Because the storage engine
is Postgres, you tune a DocumentDB instance with ordinary Postgres-style `key=value` settings
placed under a **`user.conf`** key — exactly the way you would tune KubeDB Postgres (it is
*not* a `mongod.conf`).

KubeDB exposes three ways to supply custom configuration at provision time, and they layer on
top of each other in a fixed precedence:

```text
auto-tuning / built-in defaults  <  configuration.secretName  <  configuration.inline
```

Anything you set **inline** wins over a referenced **Secret**, which in turn wins over the
**auto-tuned / default** values. The operator merges every supplied source and renders the
final files into a per-instance config Secret that is mounted into every pod.

| Source                          | Field                           | Precedence |
| ------------------------------- | ------------------------------- | ---------- |
| Auto-tuning / built-in defaults | `spec.configuration.tuning`     | lowest     |
| Secret                          | `spec.configuration.secretName` | middle     |
| Inline                          | `spec.configuration.inline`     | highest    |

## Before You Begin

- You need a Kubernetes cluster, and the `kubectl` command-line tool must be configured to
  communicate with your cluster. If you do not already have a cluster, you can create one by
  using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo`:

  ```bash
  kubectl create ns demo
  ```
  namespace/demo created

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Configuration via a Secret (cluster)

Create a Secret whose single `user.conf` key carries your Postgres settings, then reference it
from the DocumentDB object with `spec.configuration.secretName`. KubeDB merges `user.conf` into
the rendered server configuration for **every** replica.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: documentdb-custom-config
  namespace: demo
stringData:
  user.conf: |
    max_connections=250
    work_mem=8MB
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb-cls-sample
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 3
  configuration:
    secretName: documentdb-custom-config
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

Apply both:

```bash
kubectl apply -f documentdb-custom-config-secret.yaml
```
secret/documentdb-custom-config created

```bash
kubectl apply -f cluster-config-secret.yaml
```
documentdb.kubedb.com/documentdb-cls-sample created

### Inspect the rendered configuration

The `spec.configuration` block on the object confirms which Secret is wired in:

```bash
kubectl get docdb -n demo documentdb-cls-sample -o jsonpath='{.spec.configuration}'
```
{"secretName":"documentdb-custom-config"}

The Secret holds the `user.conf` that KubeDB feeds into each replica:

```bash
kubectl get secret -n demo documentdb-custom-config -o jsonpath='{.data.user\.conf}' | base64 -d
```
max_connections=250
work_mem=8MB

KubeDB also provisions the cluster's two auth secrets alongside it — `documentdb-cls-sample-auth`
(the MongoDB-compatibility `default_user`) and `documentdb-cls-sample-admin-auth` (the backend
admin):

```bash
kubectl get secret -n demo | grep documentdb-cls-sample
```
documentdb-cls-sample-admin-auth   kubernetes.io/basic-auth   2      34m
documentdb-cls-sample-auth         kubernetes.io/basic-auth   2      34m

### Verify the database is serving

Connect over the MongoDB wire protocol (TLS, port `10260`) with the `default_user` credentials
from `<db>-auth` and ping:

```bash
PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
```
{ ok: 1 }

The primary accepts MongoDB-protocol traffic with the custom configuration applied.

Tear the instance down before the next example:

```bash
kubectl delete docdb -n demo documentdb-cls-sample
```
documentdb.kubedb.com "documentdb-cls-sample" deleted

## Configuration inline

The inline form embeds the same Postgres settings directly in the DocumentDB spec under
`spec.configuration.inline`. Inline values take precedence over a referenced Secret.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb-sa-sample
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  configuration:
    inline:
      user.conf: |
        max_connections=300
        work_mem=16MB
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

```bash
kubectl apply -f standalone-config-inline.yaml
```
documentdb.kubedb.com/documentdb-sa-sample created

On a healthy instance the rendered `user.conf` would show `max_connections=300` /
`work_mem=16MB`, overriding any Secret-supplied values.

## Configuration via auto-tuning

The tuning form lets KubeDB compute Postgres settings for you from a workload profile and the
underlying storage characteristics instead of hand-writing `user.conf`. The operator runs a
`pgtune`-style calculation and renders the result into `pgtune.conf` (which sits at the lowest
precedence, so an explicit Secret or inline `user.conf` still overrides it).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: documentdb-sa-sample
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 1
  configuration:
    tuning:
      profile: oltp          # web | oltp | dw | mixed | desktop
      storageType: ssd       # ssd | hdd | san
      maxConnections: 200
      disableAutoTune: false
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

```bash
kubectl apply -f standalone-config-tuning.yaml
```
documentdb.kubedb.com/documentdb-sa-sample created

On a healthy instance the auto-tuner emits a `pgtune.conf` derived from `profile: oltp`,
`storageType: ssd`, and `maxConnections: 200` (tuned `shared_buffers`, `effective_cache_size`,
`work_mem`, `max_connections=200`, etc.).

> [!NOTE]
> **Standalone provisioning limitation in the test environment.** On the cluster used to
> capture this guide, standalone (`replicas: 1`) DocumentDB instances on version `pg17-0.109.0`
> did not finish bootstrapping: the standalone PetSet is rendered with only the `documentdb`
> container (the `documentdb-coordinator` sidecar that runs `initdb` on the clustered topology
> is absent), so the internal PostgreSQL data directory is never created and port `10260` never
> opens. The inline and tuning YAML above are the intended procedure; the live rendered-config
> inspection was therefore captured on the 3-replica (cluster) topology shown in the first
> section. The configuration mechanics (`user.conf` key, three sources, precedence) are
> identical for standalone and cluster.

## Cleaning Up

```bash
kubectl delete docdb -n demo documentdb-cls-sample --ignore-not-found
kubectl delete docdb -n demo documentdb-sa-sample --ignore-not-found
kubectl delete secret -n demo documentdb-custom-config --ignore-not-found
kubectl delete ns demo
```

## Next Steps

- Apply configuration to a running database with the [Reconfigure](/docs/guides/documentdb/reconfigure/) OpsRequest.
- [Restart](/docs/guides/documentdb/restart/) a DocumentDB database.
