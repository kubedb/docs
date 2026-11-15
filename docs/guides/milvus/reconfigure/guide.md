---
title: Reconfigure Milvus
menu:
  docs_{{ .version }}:
    identifier: milvus-reconfigure-guide
    name: Guide
    parent: milvus-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Milvus

This guide will show you how to use the `KubeDB` Ops-manager operator to reconfigure a Milvus database, applying custom configuration through the `milvus.yaml` configuration file.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)
  - [Reconfigure Overview](/docs/guides/milvus/reconfigure/overview.md)

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

- To keep things isolated, this tutorial uses a separate namespace called `demo`:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Milvus configuration is always supplied through a file named **`milvus.yaml`**. Use that exact key in config secrets and in `applyConfig`.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/reconfigure/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/reconfigure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Reconfigure Standalone Milvus

### Deploy Milvus

Deploy the [standalone quickstart](/docs/guides/milvus/quickstart/standalone.md) manifest and wait for it to become `Ready`:

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com -n demo milvus-standalone
NAME                VERSION   STATUS   AGE
milvus-standalone   2.6.11    Ready    2m
```

### Apply the Reconfigure OpsRequest

We will apply a new configuration through a config secret and an inline `applyConfig`. The OpsRequest below first references a config secret (`mv-configuration`), then overrides part of it inline:

`reconfigure-standalone.yaml`

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: mv-configuration
  namespace: demo
type: Opaque
stringData:
  milvus.yaml: |
    log:
      level: debug
      file:
        maxAge: 20
    queryNode:
      gracefulTime: 10
    dataNode:
      segment:
        maxSize: 400
---
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: reconfigure-1
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: milvus-standalone
  configuration:
    removeCustomConfig: true
    configSecret:
      name: mv-configuration
    applyConfig:
      milvus.yaml: |
        log:
          level: info
          file:
            maxAge: 30
        queryNode:
          gracefulTime: 500
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` is the Milvus we are reconfiguring.
- `spec.configuration.configSecret` references a secret whose `milvus.yaml` holds configuration.
- `spec.configuration.applyConfig` merges an inline `milvus.yaml` on top. Keys set in both places take the `applyConfig` value, so `log.level`, `log.file.maxAge` and `queryNode.gracefulTime` end up as `info`, `30` and `500`. Keys set only in the config secret keep the secret's value, so `dataNode.segment.maxSize` becomes `400`.
- `spec.configuration.removeCustomConfig: true` discards any previously applied custom configuration (config secret reference and inline configuration) first. The `configSecret` and `applyConfig` of this same request are still applied.
- `spec.configuration.restart` is omitted, so it defaults to `auto`: because the configuration changes, the pods are restarted. Milvus does not reload these settings while running, so a restart is what makes them take effect. With `restart: "false"`, the new `milvus.yaml` is written to the pods, but the running Milvus keeps its old values until the pods restart.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure/yamls/reconfigure-standalone.yaml
secret/mv-configuration created
milvusopsrequest.ops.kubedb.com/reconfigure-1 created
```

### Watch Progress

```bash
$ kubectl get milvusopsrequest -n demo
NAME            TYPE          STATUS       AGE
reconfigure-1   Reconfigure   Successful   28s
```

```bash
$ kubectl describe milvusopsrequest reconfigure-1 -n demo
...
Status:
  Conditions:
    Message:  Milvus ops-request has started to reconfigure Milvus nodes
    Reason:   Reconfigure
    Type:     Reconfigure
    Message:  Successfully prepared user provided apply configs
    Reason:   PrepareApplyConfig
    Type:     PrepareApplyConfig
    Message:  successfully reconciled the milvus with new configuration
    Reason:   UpdatePetSets
    Type:     UpdatePetSets
    ...
    Message:  Successfully Restarted Pods after reconfiguration
    Reason:   RestartPods
    Type:     RestartPods
    Message:  Successfully completed reconfigure milvus
    Reason:   Successful
    Type:     Successful
  Phase:      Successful
Events:
  Normal  Starting       Pausing Milvus databse: demo/milvus-standalone
  Normal  UpdatePetSets  successfully reconciled the milvus with new configuration
  ...
  Normal  Starting       Resuming Milvus database: demo/milvus-standalone
  Normal  Successful     Successfully resumed Milvus database: demo/milvus-standalone for MilvusOpsRequest: reconfigure-1
```

### Verify the New Configuration

The applied values are rendered into the configuration secret's `milvus.yaml`:

```bash
$ DB_UID=$(kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.metadata.uid}')
$ CFG=milvus-standalone-${DB_UID: -6}
$ kubectl get secret $CFG -n demo -o jsonpath='{.data.milvus\.yaml}' | base64 -d | grep -A3 -E '^log:|^queryNode:'
log:
  file:
    maxAge: 30
    maxBackups: 20
    maxSize: 300
...
  level: info
queryNode:
  gracefulTime: 500
  port: 19536
```

The `log.level` is now `info`, `log.file.maxAge` is `30`, and `queryNode.gracefulTime` is `500` — exactly the values supplied through `applyConfig`. `dataNode.segment.maxSize: 400` from the config secret is rendered too.

The rendered configuration secret is named `<db-name>-<last 6 characters of the Milvus object's UID>`, as computed above. The database records the custom configuration in `spec.configuration`: `secretName` is `mv-configuration`, and `inline` holds the `applyConfig` values under the `milvus.yaml` key:

```bash
$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.configuration}'
```

### Verify the Running Milvus

The rendered secret only shows what Milvus *will* load. Check that the pod was restarted (its `AGE` is reset) and that the running Milvus reads the new file:

```bash
$ kubectl get pods -n demo milvus-standalone-0

$ kubectl exec -n demo milvus-standalone-0 -c milvus -- grep -A3 -E '^log:|^queryNode:' /milvus/configs/milvus.yaml
```

You can also ask Milvus for its effective configuration through the management endpoint on port `9091`. Port-forward it in a separate terminal (`kubectl port-forward svc/milvus-standalone -n demo 9091:9091`), then run:

```bash
$ PASSWORD=$(kubectl get secret milvus-standalone-auth -n demo -o jsonpath='{.data.password}' | base64 -d)
$ curl -s -H "Authorization: Bearer root:${PASSWORD}" http://localhost:9091/api/v1/_cluster/configs | jq .
```

The response lists each effective setting with lowercase keys. After the restart, `log.file.maxage` is `30`, `querynode.gracefultime` is `500` and `datanode.segment.maxsize` is `400`.

## Reconfigure Distributed Milvus

For a distributed Milvus, the flow is identical; only `spec.databaseRef.name` points at the distributed database (`milvus-cluster`). The same `milvus.yaml` configuration is rendered into the configuration secret and propagated to every distributed role (`mixcoord`, `datanode`, `querynode`, `streamingnode`, `proxy`), whose pods are then restarted one at a time.

Deploy the [distributed quickstart](/docs/guides/milvus/quickstart/distributed.md) manifest and wait for it to become `Ready`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w
```

The distributed sample reuses the OpsRequest name `reconfigure-1`, so delete the standalone OpsRequest first. The `mv-configuration` secret has the same content in both samples; if it still exists, `kubectl apply` reports it as `unchanged`.

```bash
$ kubectl delete milvusopsrequest -n demo reconfigure-1
```

`reconfigure-distributed.yaml`

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: mv-configuration
  namespace: demo
type: Opaque
stringData:
  milvus.yaml: |
    log:
      level: debug
      file:
        maxAge: 20
    queryNode:
      gracefulTime: 10
    dataNode:
      segment:
        maxSize: 400
---
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: reconfigure-1
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: milvus-cluster
  configuration:
    removeCustomConfig: true
    configSecret:
      name: mv-configuration
    applyConfig:
      milvus.yaml: |
        log:
          level: info
          file:
            maxAge: 30
        queryNode:
          gracefulTime: 500
  timeout: 5m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/reconfigure/yamls/reconfigure-distributed.yaml
secret/mv-configuration created
milvusopsrequest.ops.kubedb.com/reconfigure-1 created

$ kubectl get milvusopsrequest reconfigure-1 -n demo
NAME            TYPE          STATUS       AGE
reconfigure-1   Reconfigure   Successful   21s
```

The applied configuration is rendered into the cluster's configuration secret and propagated to all roles:

```bash
$ DB_UID=$(kubectl get milvuses.kubedb.com milvus-cluster -n demo -o jsonpath='{.metadata.uid}')
$ CFG=milvus-cluster-${DB_UID: -6}
$ kubectl get secret $CFG -n demo -o jsonpath='{.data.milvus\.yaml}' | base64 -d | grep -A2 -E '^log:|^queryNode:|level:'
log:
  file:
    maxAge: 30
    maxBackups: 20
...
  level: info
queryNode:
  enableDisk: true
  gracefulTime: 500
  port: 21123
```

As with standalone, `log.level` is now `info`, `log.file.maxAge` is `30`, and `queryNode.gracefulTime` is `500`. To check the running Milvus, use the same management-endpoint request through `svc/milvus-cluster` (backed by the proxy), or read `/milvus/configs/milvus.yaml` in any role pod, for example `milvus-cluster-querynode-0`.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo reconfigure-1
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete secret -n demo mv-configuration
$ kubectl delete ns demo
```

## Next Steps

- Learn how to [restart](/docs/guides/milvus/restart/guide.md) a Milvus database.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
