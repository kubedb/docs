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

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.

- To keep things isolated, this tutorial uses a separate namespace called `demo`:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Milvus configuration is always supplied through a file named **`milvus.yaml`**. Use that exact key in config secrets and in `applyConfig`.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/reconfigure/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/reconfigure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Reconfigure Standalone Milvus

### Deploy Milvus

Deploy a standalone Milvus and wait for it to become `Ready` (see the [standalone quickstart](/docs/guides/milvus/quickstart/standalone.md)):

```bash
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
    restart: "false"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` is the Milvus we are reconfiguring.
- `spec.configuration.configSecret` references a secret whose `milvus.yaml` holds configuration.
- `spec.configuration.applyConfig` merges an inline `milvus.yaml` on top — here the final, effective values for `log` and `queryNode`.
- `spec.configuration.removeCustomConfig: true` discards any previously applied custom configuration first.
- `spec.configuration.restart: "false"` requests the configuration be applied without forcing a restart.

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
    Message:  Successfully completed reconfigure milvus
    Reason:   Successful
    Type:     Successful
  Phase:      Successful
Events:
  Normal  Starting       Pausing Milvus databse: demo/milvus-standalone
  Normal  UpdatePetSets  successfully reconciled the milvus with new configuration
  Normal  Starting       Resuming Milvus database: demo/milvus-standalone
  Normal  Successful     Successfully resumed Milvus database: demo/milvus-standalone for MilvusOpsRequest: reconfigure-1
```

### Verify the New Configuration

The applied values are rendered into the configuration secret's `milvus.yaml`:

```bash
$ CFG=$(kubectl get secret -n demo -o name | grep -oE 'milvus-standalone-[a-f0-9]{6}' | head -1)
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

The `log.level` is now `info`, `log.file.maxAge` is `30`, and `queryNode.gracefulTime` is `500` — exactly the values supplied through `applyConfig`.

## Reconfigure Distributed Milvus

For a distributed Milvus, the flow is identical; only `spec.databaseRef.name` points at the distributed database (`milvus-cluster`). The same `milvus.yaml` configuration is rendered into the configuration secret and propagated to every distributed role (`mixcoord`, `datanode`, `querynode`, `streamingnode`, `proxy`).

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
    restart: "false"
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
$ CFG=$(kubectl get secret -n demo -o name | grep -oE 'milvus-cluster-[a-f0-9]{6}' | head -1)
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

As with standalone, `log.level` is now `info`, `log.file.maxAge` is `30`, and `queryNode.gracefulTime` is `500`.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo reconfigure-1
$ kubectl delete secret -n demo mv-configuration
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone
$ kubectl delete ns demo
```

## Next Steps

- Learn how to [restart](/docs/guides/milvus/restart/guide.md) a Milvus database.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
