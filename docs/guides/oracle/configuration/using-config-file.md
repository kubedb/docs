---
title: Run Oracle with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: oracle-using-config-file
    name: Config File
    parent: oracle-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Oracle. This tutorial will show you how to use KubeDB to run a Oracle database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/oracle/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Oracle allows configuring the database via a YAML configuration file named `production.yaml`. When the Oracle Docker image starts, it merges configuration from the default `config.yaml` with any `production.yaml` file present. KubeDB takes advantage of this feature to allow users to provide their custom configuration. To know more about configuring Oracle, see [here](https://oracle.tech/documentation/guides/configuration/).

At first, you have to create a config file named `production.yaml` with your desired configuration. Then create a Secret with this configuration file and provide its name in `spec.configuration.secretName`. The operator reads this Secret and mounts it into the Oracle pods automatically.

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
$ kubectl create secret generic -n demo oracle-config \
  --from-file=production.yaml=./production.yaml
secret/oracle-config created
```

Verify the Secret has the configuration file:

```yaml
$ kubectl get secret -n demo oracle-config -o yaml
apiVersion: v1
data:
  production.yaml: bG9nX2xldmVsOiBJTkZPCnNlcnZpY2U6CiAgbWF4X3JlcXVlc3Rfc2l6ZV9tYjogNjQK...
kind: Secret
metadata:
  name: oracle-config
  namespace: demo
```

Now, create the `Oracle` CR specifying `spec.configuration.secretName` field:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/configuration/oracle-configuration.yaml
oracle.kubedb.com/custom-oracle created
```

Below is the YAML for the `Oracle` CR we just created:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: custom-oracle
  namespace: demo
spec:
  version: "21.3.0"
  replicas: 3
  configuration:
    secretName: oracle-config
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create the necessary PVC, StatefulSet, services, and secrets. If everything goes well, we will see that a pod with the name `custom-oracle-0` has been created.

Check that the StatefulSet's pod is running:

```bash
$ kubectl get pod -n demo custom-oracle-0
NAME               READY   STATUS    RESTARTS   AGE
custom-oracle-0    1/1     Running   0          2m
```

Now, wait for the `Oracle` CR to go into `Ready` state:

```bash
$ kubectl get oracle -n demo custom-oracle
NAME            VERSION   STATUS   AGE
custom-oracle   1.17.0    Ready    3m
```

We can check that the Oracle database is running with our custom configuration by accessing the telemetry endpoint:

```bash
$ kubectl port-forward -n demo pod/custom-oracle-0 1521:1521 &
$ curl http://localhost:1521/telemetry | jq '.result.app.max_request_size_mb'
64
```

The output confirms the database is using our custom `max_request_size_mb` value of `64`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo custom-oracle
kubectl delete secret -n demo oracle-config
kubectl delete ns demo
```