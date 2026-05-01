---
title: Run Milvus with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: milvus-using-config-file
    name: Config File
    parent: milvus-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Milvus. This tutorial will show you how to use KubeDB to run Milvus with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/milvus](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/milvus) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Milvus supports configuration via the `milvus.yaml` file. KubeDB takes advantage of `spec.configuration.secretName` to allow users to provide a custom `milvus.yaml` without mounting any volume into the Pod. The operator reads this Secret internally and applies the configuration automatically.

In this tutorial, we will configure `queryNode.gracefulTime` and `dataNode.segment.maxSize` parameters.

## Custom Configuration

At first, let's create a custom `milvus.yaml` file:

```yaml
queryNode:
  gracefulTime: 5000

dataNode:
  segment:
    maxSize: 512
```

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo milvus-configuration \
  --from-file=milvus.yaml=./milvus.yaml
secret/milvus-configuration created
```

Verify the Secret has the configuration file.

```bash
$ kubectl get secret -n demo milvus-configuration -o yaml
apiVersion: v1
data:
  milvus.yaml: <base64-encoded-content>
kind: Secret
metadata:
  name: milvus-configuration
  namespace: demo
```

Now, create Milvus CRD specifying `spec.configuration.secretName` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: custom-milvus
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: my-release-minio
  configuration:
    secretName: milvus-configuration
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/milvus/configuration/milvus-configuration.yaml
milvus.kubedb.com/custom-milvus created
```

Now, wait for the Milvus to be ready.

```bash
$ kubectl get milvus -n demo custom-milvus
NAME            VERSION   STATUS   AGE
custom-milvus   2.4.0     Ready    3m
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo custom-milvus-0
NAME              READY   STATUS    RESTARTS   AGE
custom-milvus-0   1/1     Running   0          3m
```

Now, we will verify the configuration has been applied. We will `exec` into the pod and check the configuration file.

```bash
$ kubectl exec -it -n demo custom-milvus-0 -- cat /milvus/configs/milvus.yaml | grep -A 5 queryNode
queryNode:
  gracefulTime: 5000
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo milvus/custom-milvus -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo milvus/custom-milvus

kubectl delete -n demo secret milvus-configuration
kubectl delete ns demo
```
