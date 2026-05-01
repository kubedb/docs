---
title: Reconfigure Oracle
menu:
  docs_{{ .version }}:
    identifier: oracle-reconfigure-cluster
    name: Cluster
    parent: oracle-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Oracle

This guide will show you how to use `KubeDB` Enterprise operator to reconfigure a Oracle cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/oracle/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/reconfigure/yamls](/docs/guides/oracle/reconfigure/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Oracle

Now, we are going to deploy a `Oracle` cluster with an initial configuration.

#### Deploy Oracle with custom config

First, we will create a `oracle.yaml` config file containing our initial configuration settings.

```yaml
# oracle.yaml
storage:
  performance:
    max_search_threads: 4
```

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo oracle-configuration --from-file=./oracle.yaml
secret/oracle-configuration created
```

Below is the YAML of the `Oracle` CR that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  configSecret:
    name: oracle-sampleuration
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure/yamls/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

Now, wait until `oracle-sample` has status `Ready`:

```bash
$ kubectl get oracle -n demo
NAME             VERSION   STATUS   AGE
oracle-sample    1.17.0    Ready    3m42s
```

### Reconfigure using new config secret

Now we will reconfigure this database to change `max_search_threads` to `8`.

First, we will create a new `oracle.yaml` file containing the updated configuration:

```yaml
# oracle.yaml
storage:
  performance:
    max_search_threads: 8
```

Then, we will create a new secret with this configuration file:

```bash
$ kubectl create secret generic -n demo new-oracle-configuration --from-file=./oracle.yaml
secret/new-oracle-configuration created
```

#### Create OracleOpsRequest

Now, we will use this secret to replace the previous secret using a `OracleOpsRequest` CR. Below is the YAML of the `OracleOpsRequest` that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: oracle-sample
  configuration:
    configSecret:
      name: new-oracle-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `oracle-sample` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new configuration secret.

Let's create the `OracleOpsRequest` CR we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure/yamls/reconfigure-using-secret.yaml
oracleopsrequest.ops.kubedb.com/qdops-reconfigure-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Enterprise operator will update the `configSecret` of the `Oracle` object.

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ kubectl get qdops -n demo
NAME                       TYPE          STATUS       AGE
qdops-reconfigure-config   Reconfigure   Successful   3m21s
```

### Reconfigure using applyConfig

We can also reconfigure our existing secret by modifying configuration inline using `applyConfig`. Below is the YAML of the `OracleOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-reconfigure-apply-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: oracle-sample
  configuration:
    applyConfig:
      oracle.yaml: |
        storage:
          performance:
            max_search_threads: 6
```

> **Note:** You can modify multiple fields of your current configuration using `applyConfig`. If you don't have any existing config secret, `applyConfig` will create a new secret for you.

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `oracle-sample` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` contains the inline configuration to apply.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure/yamls/apply-config.yaml
oracleopsrequest.ops.kubedb.com/qdops-reconfigure-apply-config created
```

#### Verify the new configuration is working

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ kubectl get qdops qdops-reconfigure-apply-config -n demo
NAME                              TYPE          STATUS       AGE
qdops-reconfigure-apply-config    Reconfigure   Successful   4m59s
```

### Remove Custom Configuration

We can also remove existing custom config using `OracleOpsRequest`. Set `spec.configuration.removeCustomConfig: true` to remove the existing custom configuration.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: oracle-sample
  configuration:
    removeCustomConfig: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/reconfigure/yamls/remove-config.yaml
oracleopsrequest.ops.kubedb.com/qdops-reconfigure-remove created
```

Let's wait for `OracleOpsRequest` to be `Successful`:

```bash
$ kubectl get qdops qdops-reconfigure-remove -n demo
NAME                       TYPE          STATUS       AGE
qdops-reconfigure-remove   Reconfigure   Successful   2m10s
```

After this, the `Oracle` CR will no longer reference a `configSecret` and the database will use its default configuration.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo qdops-reconfigure-config qdops-reconfigure-apply-config qdops-reconfigure-remove
kubectl delete oracle -n demo oracle-sample
kubectl delete ns demo
```