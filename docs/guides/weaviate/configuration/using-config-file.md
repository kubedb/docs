---
title: Run Weaviate with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: weaviate-using-config-file
    name: Config File
    parent: weaviate-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Weaviate. This tutorial will show you how to use KubeDB to run a Weaviate database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Weaviate allows configuring the database via a YAML configuration file named `weaviate.yaml`. KubeDB takes advantage of this feature to allow users to provide their custom configuration. To know more about configuring Weaviate, see [here](https://weaviate.io/developers/weaviate/configuration).

At first, you have to create a config file named `weaviate.yaml` with your desired configuration. Then create a Secret with this configuration file and provide its name in `spec.configuration.secretName`. The operator reads this Secret and mounts it into the Weaviate pods automatically.

In this tutorial, we will configure `query_defaults.limit` and disable anonymous access via a custom config file.

## Custom Configuration

At first, let's create a `weaviate.yaml` file with custom settings:

```yaml
authentication:
  anonymous_access:
    enabled: false
query_defaults:
  limit: 25
persistence:
  data_path: /var/lib/weaviate
```

Now, create a Secret with this configuration file:

```bash
$ kubectl create secret generic -n demo weaviate-config \
  --from-file=weaviate.yaml=./weaviate.yaml
secret/weaviate-config created
```

Verify the Secret has the configuration file:

```yaml
$ kubectl get secret -n demo weaviate-config -o yaml
apiVersion: v1
data:
  weaviate.yaml: YXV0aGVudGljYXRpb246CiAgYW5vbnltb3VzX2FjY2Vzczo...
kind: Secret
metadata:
  name: weaviate-config
  namespace: demo
```

Now, create the `Weaviate` CR specifying `spec.configuration.secretName` field:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/configuration/weaviate-configuration.yaml
weaviate.kubedb.com/custom-weaviate created
```

Below is the YAML for the `Weaviate` CR we just created:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: custom-weaviate
  namespace: demo
spec:
  version: "1.33.1"
  replicas: 3
  configuration:
    secretName: weaviate-config
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create the necessary PVC, StatefulSet, services, and secrets. If everything goes well, we will see that a pod with the name `custom-weaviate-0` has been created.

Check that the StatefulSet's pod is running:

```bash
$ kubectl get pod -n demo custom-weaviate-0
NAME                 READY   STATUS    RESTARTS   AGE
custom-weaviate-0    1/1     Running   0          2m
```

Now, wait for the `Weaviate` CR to go into `Ready` state:

```bash
$ kubectl get weaviate -n demo custom-weaviate
NAME              VERSION   STATUS   AGE
custom-weaviate   1.33.1    Ready    3m
```

We can verify the Weaviate database is running with our custom configuration by querying the meta endpoint:

```bash
$ kubectl port-forward -n demo pod/custom-weaviate-0 8080:8080 &
$ export WEAVIATE_API_KEY=$(kubectl get secret -n demo custom-weaviate-auth -o jsonpath='{.data.api-key}' | base64 -d)

$ curl -H "Authorization: Bearer $WEAVIATE_API_KEY" http://localhost:8080/v1/meta | jq '.version'
"1.33.1"
```

Anonymous access is now disabled, and the query default limit is set to `25` as configured.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo custom-weaviate
kubectl delete secret -n demo weaviate-config
kubectl delete ns demo
```