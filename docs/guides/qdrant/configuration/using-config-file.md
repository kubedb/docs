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

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Qdrant allows configuring the database via a YAML configuration file named `production.yaml`. When the Qdrant Docker image starts, it merges configuration from the default `config.yaml` with any `production.yaml` file present. KubeDB takes advantage of this feature to allow users to provide their custom configuration. To know more about configuring Qdrant, see [here](https://qdrant.tech/documentation/guides/configuration/).

At first, you have to create a config file named `production.yaml` with your desired configuration. Then create a Secret with this configuration file and provide its name in `spec.configSecret.name`. The operator reads this Secret and mounts it into the Qdrant pods automatically.

In this tutorial, we will configure `log_level` and `service.max_request_size_mb` via a custom config file.

## Custom Configuration

At first, let's create a `production.yaml` file with custom settings:

```yaml
log_level: INFO
service:
  max_request_size_mb: 64
storage:
  performance:
    max_search_threads: 4
```

Now, create a Secret with this configuration file:

```bash
$ kubectl create secret generic -n demo qdrant-config \
  --from-file=production.yaml=./production.yaml
secret/qdrant-config created
```

Verify the Secret has the configuration file:

```yaml
$ kubectl get secret -n demo qdrant-config -o yaml
apiVersion: v1
data:
  production.yaml: bG9nX2xldmVsOiBJTkZPCnNlcnZpY2U6CiAgbWF4X3JlcXVlc3Rfc2l6ZV9tYjogNjQK...
kind: Secret
metadata:
  name: qdrant-config
  namespace: demo
```

Now, create the `Qdrant` CR specifying `spec.configSecret.name` field:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/configuration/qdrant-configuration.yaml
qdrant.kubedb.com/custom-qdrant created
```

Below is the YAML for the `Qdrant` CR we just created:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: custom-qdrant
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  configSecret:
    name: qdrant-config
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create the necessary PVC, Petset, services, and secrets. If everything goes well, we will see that a pod with the name `custom-qdrant-0` has been created.

Check that the Petset's pod is running:

```bash
$ kubectl get pod -n demo custom-qdrant-0
NAME               READY   STATUS    RESTARTS   AGE
custom-qdrant-0    1/1     Running   0          2m
```

Now, wait for the `Qdrant` CR to go into `Ready` state:

```bash
$ kubectl get qdrant -n demo custom-qdrant
NAME            VERSION   STATUS   AGE
custom-qdrant   1.17.0    Ready    3m
```

We can check that the Qdrant database is running with our custom configuration by accessing the telemetry endpoint:

```bash
$ kubectl port-forward -n demo pod/custom-qdrant-0 6333:6333 &
$ curl http://localhost:6333/telemetry | jq '.result.app.max_request_size_mb'
64
```

The output confirms the database is using our custom `max_request_size_mb` value of `64`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo custom-qdrant
kubectl delete secret -n demo qdrant-config
kubectl delete ns demo
```