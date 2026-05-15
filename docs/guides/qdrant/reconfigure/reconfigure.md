---
title: Reconfigure Qdrant
menu:
  docs_{{ .version }}:
    identifier: qdrant-reconfigure-cluster
    name: Cluster
    parent: qdrant-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Qdrant

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a Qdrant cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Qdrant](/docs/guides/qdrant/concepts/qdrant.md)
  - [QdrantOpsRequest](/docs/guides/qdrant/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/qdrant/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/reconfigure](/docs/examples/qdrant/reconfigure) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prepare Qdrant

Now, we are going to deploy a `Qdrant` cluster with an initial configuration.

### Deploy Qdrant with custom config

Below is the YAML of the configuration `Secret` that we are going to create:

```yaml
apiVersion: v1
stringData:
  config.yaml: |
    log_level: DEBUG
    performance:
      max_search_threads: 4
      update_rate_limit: 100
kind: Secret
metadata:
  name: qdrant-configuration
  namespace: demo
type: Opaque
```

Let's create the `Secret` we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/configuration-secret.yaml
secret/qdrant-configuration created
```

Below is the YAML of the `Qdrant` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  configSecret:
    name: qdrant-configuration
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Qdrant` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Now, wait until `qdrant-sample` has status `Ready`:

```bash
$ kubectl get qdrant -n demo
NAME             VERSION   STATUS   AGE
qdrant-sample    1.17.0    Ready    3m42s
```

## Reconfigure using new config secret

Now we will reconfigure this database to change `max_search_threads` to `8`.

Below is the YAML of the new configuration `Secret` that we are going to create:

```yaml
apiVersion: v1
stringData:
  config.yaml: |
    log_level: DEBUG
    performance:
      max_search_threads: 8
      update_rate_limit: 100
kind: Secret
metadata:
  name: new-qdrant-configuration
  namespace: demo
type: Opaque
```

Let's create the `Secret` we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/new-configuration-secret.yaml
secret/new-qdrant-configuration created
```

### Create QdrantOpsRequest

Now, we will use this secret to replace the previous secret using a `QdrantOpsRequest` CR. Below is the YAML of the `QdrantOpsRequest` that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: qdrant-sample
  configuration:
    configSecret:
      name: new-qdrant-configuration
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `qdrant-sample` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new configuration secret.
- `spec.timeout` specifies the timeout for the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#spectimeout)).
- `spec.apply` specifies when to apply the operation (learn more [here](/docs/guides/qdrant/concepts/opsrequest.md#specapply)).

Let's create the `QdrantOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/reconfigure-using-secret.yaml
qdrantopsrequest.ops.kubedb.com/qdops-reconfigure-config created
```

### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of the `Qdrant` object.

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ kubectl get qdops -n demo
NAME                       TYPE          STATUS       AGE
qdops-reconfigure-config   Reconfigure   Successful   3m
```

## Reconfigure using applyConfig

We can also reconfigure our existing secret by modifying configuration inline using `applyConfig`. Below is the YAML of the `QdrantOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-reconfigure-apply-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: qdrant-sample
  configuration:
    applyConfig:
      config.yaml: |
        log_level: DEBUG
        performance:
          max_search_threads: 6
          update_rate_limit: 100
```

> **Note:** You can modify multiple fields of your current configuration using `applyConfig`. If you don't have any existing config secret, `applyConfig` will create a new secret for you. If a config secret already exists, `applyConfig` will merge the new configuration with the existing one.

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `qdrant-sample` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` contains the inline configuration to apply.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/apply-config.yaml
qdrantopsrequest.ops.kubedb.com/qdops-reconfigure-apply-config created
```

### Verify the new configuration is working

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ kubectl get qdops qdops-reconfigure-apply-config -n demo
NAME                              TYPE          STATUS       AGE
qdops-reconfigure-apply-config    Reconfigure   Successful   5m30s
```

## Remove Custom Configuration

We can also remove existing custom config using `QdrantOpsRequest`. Set `spec.configuration.removeCustomConfig: true` to remove the existing custom configuration.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: qdops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: qdrant-sample
  configuration:
    removeCustomConfig: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/reconfigure/remove-config.yaml
qdrantopsrequest.ops.kubedb.com/qdops-reconfigure-remove created
```

Let's wait for `QdrantOpsRequest` to be `Successful`:

```bash
$ kubectl get qdops qdops-reconfigure-remove -n demo
NAME                       TYPE          STATUS       AGE
qdops-reconfigure-remove   Reconfigure   Successful   97s
```

After this, the `Qdrant` CR will no longer reference a `configSecret` and the database will use its default configuration.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete qdrantopsrequest -n demo qdops-reconfigure-config qdops-reconfigure-apply-config qdops-reconfigure-remove
qdrantopsrequest.ops.kubedb.com "qdops-reconfigure-config" deleted
qdrantopsrequest.ops.kubedb.com "qdops-reconfigure-apply-config" deleted
qdrantopsrequest.ops.kubedb.com "qdops-reconfigure-remove" deleted

$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```