---
title: Run HanaDB with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: hanadb-using-config-file
    name: Config File
    parent: hanadb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports user-provided SAP HANA configuration. This tutorial shows how to run HanaDB with a custom `global.ini` file.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl` to communicate with it. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install the KubeDB CLI on your workstation and the KubeDB operator in your cluster by following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB supports custom HanaDB configuration through a user-provided `global.ini` file. The `spec.configuration.secretName` field lets you provide this configuration without manually mounting any volume into the pod.

To apply custom configuration, you create a Kubernetes Secret containing your custom config file and provide its name in `spec.configuration.secretName`. The operator reads this Secret internally and applies the configuration automatically.

In this tutorial, you configure `global.ini` with a custom memory allocation limit.

## Custom Configuration

Create a Secret that contains a custom `global.ini` file:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hanadb-configuration
  namespace: demo
stringData:
  global.ini: |
    [memorymanager]
    global_allocation_limit = 8589934592
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/configuration/hanadb-configuration.yaml
secret/hanadb-configuration created
```

Verify the Secret has the configuration file.

```yaml
$ kubectl get secret -n demo hanadb-configuration -o yaml
apiVersion: v1
data:
  global.ini: <base64-encoded-content>
kind: Secret
metadata:
  name: hanadb-configuration
  namespace: demo
```

Create a HanaDB object with `spec.configuration.secretName` set to the Secret name.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: custom-hanadb
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  configuration:
    secretName: hanadb-configuration
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/configuration/custom-hanadb.yaml
hanadb.kubedb.com/custom-hanadb created
```

Wait for the HanaDB instance to become ready.

```bash
$ kubectl get hanadb -n demo custom-hanadb
NAME            VERSION   STATUS   AGE
custom-hanadb   2.0.82    Ready    5m
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo custom-hanadb-0
NAME              READY   STATUS    RESTARTS   AGE
custom-hanadb-0   1/1     Running   0          5m
```

Check whether the database started with the custom configuration by running `hdbsql` inside the pod.

```bash
$ kubectl exec -it -n demo custom-hanadb-0 -- hdbsql \
  -u SYSTEM -p <password> \
  "SELECT KEY, VALUE FROM SYS.M_INIFILE_CONTENTS WHERE FILE_NAME = 'global.ini' AND KEY = 'global_allocation_limit'"
KEY                       VALUE
global_allocation_limit   8589934592
```

This guide covers initial custom configuration during provisioning.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/custom-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/custom-hanadb

kubectl delete -n demo secret hanadb-configuration
kubectl delete ns demo
```
