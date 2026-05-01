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

KubeDB supports providing custom configuration for HanaDB. This tutorial will show you how to use KubeDB to run HanaDB with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

SAP HANA allows configuration via the `global.ini` and `indexserver.ini` configuration files. KubeDB takes advantage of the `spec.configuration.secretName` field to allow users to provide their custom configuration without mounting any volume into the Pod.

To apply custom configuration, you create a Kubernetes Secret containing your custom config file and provide its name in `spec.configuration.secretName`. The operator reads this Secret internally and applies the configuration automatically.

In this tutorial, we will configure `indexserver.ini` with a custom `max_memory` parameter.

## Custom Configuration

At first, let's create a custom `global.ini` file:

```ini
[system]
usage = development

[memorymanager]
alloclimit = 16384
```

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo hanadb-configuration \
  --from-file=global.ini=./global.ini
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

Now, create HanaDB CRD specifying `spec.configuration.secretName` field.

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
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/configuration/hanadb-configuration.yaml
hanadb.kubedb.com/custom-hanadb created
```

Now, wait for the HanaDB to be ready.

```bash
$ kubectl get hanadb -n demo custom-hanadb
NAME            VERSION   STATUS   AGE
custom-hanadb   2.0       Ready    5m
```

Check that the pod is running:

```bash
$ kubectl get pod -n demo custom-hanadb-0
NAME              READY   STATUS    RESTARTS   AGE
custom-hanadb-0   1/1     Running   0          5m
```

Now, we will check if the database has started with the custom configuration we have provided. We will `exec` into the pod and use the HDB CLI to check the configuration.

```bash
$ kubectl exec -it -n demo custom-hanadb-0 -- hdbsql \
  -u SYSTEM -p <password> \
  "SELECT KEY, VALUE FROM SYS.M_INIFILE_CONTENTS WHERE FILE_NAME = 'global.ini' AND KEY = 'alloclimit'"
KEY          VALUE
alloclimit   16384
```

## Reconfiguring

If you want to change the configuration, you can update the Secret and then trigger a reconfigure OpsRequest. For more details, see the [Reconfigure](/docs/guides/hanadb/ops-request/overview.md) section.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb/custom-hanadb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo hanadb/custom-hanadb

kubectl delete -n demo secret hanadb-configuration
kubectl delete ns demo
```
