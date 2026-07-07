---
title: Reconfigure Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-reconfigure-cluster
    name: Reconfigure
    parent: weaviate-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Weaviate

This guide will show you how to use the `KubeDB` Ops Manager to reconfigure a Weaviate cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Custom Configuration](/docs/guides/weaviate/configuration/using-config-file.md)
  - [Reconfigure Overview](/docs/guides/weaviate/reconfigure/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/reconfigure](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

In this section, we are going to deploy a Weaviate cluster. We will reconfigure it in the next step.

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
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR and wait for it to become `Ready`.

## Prepare Reconfigure Helper Resources

The reconfigure operation in this example references a configuration `Secret` and a backup-credentials `Secret`. Let's create them first.

The new configuration Secret (`conf.yaml` as the key):

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-weaviate-config
  namespace: demo
type: Opaque
stringData:
  conf.yaml: |-
    authorization:
      admin_list:
        enabled: false
      rbac:
        enabled: false
    cluster:
      hostname: $(POD_NAME)
    debug: false
    persistence:
      data_path: /var/lib/weaviate
    query_defaults:
      limit: 200
```

The backup-credentials Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: minio-secret
  namespace: demo
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: minio
  AWS_SECRET_ACCESS_KEY: minio123
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure/new-weaviate-config.yaml
```
secret/new-weaviate-config created

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure/minio-secret.yaml
```
secret/minio-secret created

## Apply Reconfigure OpsRequest

Now, we are going to reconfigure the cluster. The OpsRequest below references the new configuration Secret, applies an inline `applyConfig` (which is merged on top), and sets a backup-credentials Secret.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: weaviate-sample
  timeout: 3m
  apply: Always
  configuration:
    applyConfig:
      conf.yaml: |-
        query_defaults:
          limit: 155
    configSecret:
      name: new-weaviate-config
    backupConfigSecret:
      name: minio-secret
```

- `spec.type` specifies that this is a `Reconfigure` operation.
- `spec.configuration.configSecret` references the `Secret` holding the new `conf.yaml`.
- `spec.configuration.applyConfig` provides inline configuration that is merged on top of the configuration from the Secret. Here it overrides `query_defaults.limit` to `155`.
- `spec.configuration.backupConfigSecret` references a `Secret` holding backup credentials.

Let's create the `WeaviateOpsRequest` CR:

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/reconfigure/ops-request.yaml
```
weaviateopsrequest.ops.kubedb.com/reconfigure created

The Ops Manager prepares the new configuration, updates the PetSet, and restarts the pods one by one.

```bash
kubectl get weaviateopsrequest -n demo reconfigure
```
NAME          TYPE          STATUS       AGE
reconfigure   Reconfigure   Successful   83s

Let's check the `status.conditions` of the `WeaviateOpsRequest`:

```bash
kubectl get weaviateopsrequest -n demo reconfigure -o yaml
```
...
status:
  conditions:
  - message: Weaviate ops-request has started to reconfigure Weaviate nodes
    reason: Reconfigure
    status: "True"
    type: Reconfigure
  - message: Successfully prepared user provided apply configs
    reason: PrepareApplyConfig
    status: "True"
    type: PrepareApplyConfig
  - message: successfully reconciled the Weaviate with new configure
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: GetPod--weaviate-sample-0
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: EvictPod--weaviate-sample-0
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: RunningPod--weaviate-sample-0
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: GetPod--weaviate-sample-1
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: EvictPod--weaviate-sample-1
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: RunningPod--weaviate-sample-1
  - message: Successfully restarted all nodes
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - message: Successfully completed reconfigure Weaviate
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

Now, let's verify that the new configuration has been applied to the `Weaviate` object:

```bash
kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.configuration}' | jq
```
{
  "backupConfigSecret": {
    "name": "minio-secret"
  },
  "inline": {
    "conf.yaml": "query_defaults:\n    limit: 155\n"
  },
  "secretName": "new-weaviate-config"
}

The reconfigure operation has been applied successfully.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Learn how to supply a [custom configuration](/docs/guides/weaviate/configuration/using-config-file.md) at provisioning time.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviateopsrequest -n demo reconfigure
```

```bash
kubectl delete weaviate -n demo weaviate-sample
```

```bash
kubectl delete ns demo
```
