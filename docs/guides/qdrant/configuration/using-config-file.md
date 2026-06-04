---
title: Run Qdrant with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: qdrant-using-config-file
    name: Config File
    parent: qdrant-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Qdrant. This tutorial will show you how to use KubeDB to run a Qdrant database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/configuration](/docs/examples/qdrant/configuration) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Overview

KubeDB supports three ways to provide custom configuration for Qdrant:

| Method | Field | Priority |
|--------|-------|----------|
| **Config Secret** | `spec.configuration.secretName` | Medium |
| **Inline Config** | `spec.configuration.inline` | Highest |
| **Default Config** | (built into the Docker image) | Lowest |

The priority order is: **Inline Config > Config Secret > Default Config**. When multiple configuration sources specify the same key, the value from the higher-priority source takes precedence. Inline config values are applied last, overriding any values from the config Secret or defaults.

To know more about configuring Qdrant, see [here](https://qdrant.tech/documentation/guides/configuration/).

In this tutorial, we will configure `log_level` and `service.max_request_size_mb` using both a config Secret and inline configuration.

## Custom Configuration via Config Secret

At first, create a Secret with your custom configuration. Below is the YAML of the `Secret` that we are going to create:

```yaml
apiVersion: v1
stringData:
  config.yaml: |
    log_level: DEBUG
    service:
      max_request_size_mb: 64
kind: Secret
metadata:
  name: qdrant-configuration
  namespace: demo
type: Opaque
```

Let's create the `Secret` we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/configuration/configuration-secret.yaml
secret/qdrant-configuration created
```

Verify the Secret has the configuration file:

```yaml
$ kubectl get secret -n demo qdrant-configuration -o yaml
apiVersion: v1
data:
  config.yaml: bG9nX2xldmVsOiBERUJVRwpzZXJ2aWNlOgogIG1heF9yZXF1ZXN0X3NpemVfbWI6IDY0Cg==
kind: Secret
metadata:
  creationTimestamp: "2026-05-19T06:39:07Z"
  name: qdrant-configuration
  namespace: demo
  resourceVersion: "3834858"
  uid: 9cad78b6-e0e3-4e3e-b999-943c01bfb09c
type: Opaque
```

Now, create the `Qdrant` CR specifying `spec.configuration.secretName` field:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/configuration/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

Below is the YAML for the `Qdrant` CR we just created:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  configuration:
    secretName: qdrant-configuration
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create the necessary PVC, PetSet, services, and secrets. Let's check the status:

```bash
$ kubectl get qdrant -n demo
NAME            VERSION   STATUS   AGE
qdrant-sample   1.17.0    Ready    68s
```

Check that all pods are running:

```bash
$ kubectl get pod -n demo
NAME              READY   STATUS    RESTARTS   AGE
qdrant-sample-0   1/1     Running   0          61s
qdrant-sample-1   1/1     Running   0          57s
qdrant-sample-2   1/1     Running   0          42s
```

Now, let's verify that the custom configuration has been applied by checking the config file inside the pod:

```bash
$ kubectl exec -n demo qdrant-sample-0 -- cat /qdrant/config/config.yaml
log_level: DEBUG
service:
  max_request_size_mb: 64
```

The output confirms the database is running with our custom `log_level` and `max_request_size_mb` values.

As noted in the [Overview](#overview), inline configuration has the highest priority. If both a config Secret and inline config specify the same key, the inline value takes precedence.

## Inline Configuration

You can also provide custom configuration inline within the `Qdrant` CR using `spec.configuration.inline`. This is useful for simple config changes without creating a separate Secret.

Below is an example YAML of a `Qdrant` CR with inline configuration:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  configuration:
    inline:
      log_level: DEBUG
      max_request_size_mb: "64"
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

> **Note:** The `inline` field is a `map[string]string`, so values must be strings. To set the config key `max_request_size_mb` to `64`, write `max_request_size_mb: "64"`. The config Secret method (shown above) supports full nested YAML structure.

When both `spec.configuration.secretName` and `spec.configuration.inline` are set, the inline values override the corresponding keys from the config Secret. Keys not specified in inline retain their values from the config Secret.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete secret -n demo qdrant-configuration
kubectl delete ns demo
```