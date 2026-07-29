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

KubeDB supports providing custom configuration for Weaviate. This tutorial will show you how to use KubeDB to run a Weaviate database with a custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB supports providing custom configuration for Weaviate through `spec.configuration`. Weaviate reads its configuration from a single `conf.yaml` file (passed to the server via `--config-file`). KubeDB mounts your configuration at `/weaviate-config/conf.yaml` inside the pods.

There are two ways to supply the configuration:

| Method | Field |
|--------|-------|
| **Config Secret** | `spec.configuration.secretName` |
| **Inline Config**  | `spec.configuration.inline` |

In both cases the configuration is provided under the `conf.yaml` key. KubeDB merges your settings with the cluster-specific values it needs (such as `cluster.hostname` and `persistence.data_path`).

To know more about configuring Weaviate, see the [Weaviate environment/config reference](https://weaviate.io/developers/weaviate/config-refs/env-vars).

## Custom Configuration via Config Secret

At first, create a Secret with your custom configuration under the `conf.yaml` key. Below is the YAML of the `Secret` that we are going to create:

```yaml
apiVersion: v1
stringData:
  conf.yaml: |-
    ---
    authentication:
      anonymous_access:
        enabled: true
      oidc:
        enabled: false
    authorization:
      admin_list:
        enabled: false
      rbac:
        enabled: false

    query_defaults:
      limit: 400
    debug: false
kind: Secret
metadata:
  name: weaviate-custom-config
  namespace: demo
  labels:
    app.kubernetes.io/name: weaviates.kubedb.com
    app.kubernetes.io/instance: weaviate-sample
type: Opaque
```

Let's create the `Secret`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/configuration/weaviate-custom-config-secret.yaml
secret/weaviate-custom-config created
```

Now, create the `Weaviate` CR specifying the `spec.configuration.secretName` field:

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
  configuration:
    secretName: weaviate-custom-config
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/configuration/cus-conf.yaml
weaviate.kubedb.com/weaviate-sample created
```

Now, wait a few minutes. KubeDB operator will create the necessary PVC, PetSet, services, and secrets. Let's check the status:

```bash
$ kubectl get weaviate -n demo
NAME              TYPE                  VERSION   STATUS   AGE
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Ready    66s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          65s
weaviate-sample-1   1/1     Running   0          50s
weaviate-sample-2   1/1     Running   0          38s
```

Now, let's verify that the custom configuration has been applied by checking the config file inside the pod:

```bash
$ kubectl exec -n demo weaviate-sample-0 -c weaviate -- cat /weaviate-config/conf.yaml/conf.yaml
authentication:
  anonymous_access:
    enabled: true
  oidc:
    enabled: false
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
  limit: 400
```

The output confirms the database is running with our custom `query_defaults.limit: 400` and `anonymous_access` settings. KubeDB has merged in the cluster-specific `cluster.hostname` and `persistence.data_path` values.

## Inline Configuration

You can also provide custom configuration inline within the `Weaviate` CR using `spec.configuration.inline`. This is useful for simple config changes without creating a separate Secret. The configuration is still provided under the `conf.yaml` key.

Below is an example YAML of a `Weaviate` CR with inline configuration:

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
  configuration:
    inline:
      conf.yaml: |-
        query_defaults:
          limit: 1000
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/configuration/cus-inline-conf.yaml
weaviate.kubedb.com/weaviate-sample created
```

Wait until the cluster is `Ready`, then verify the inline configuration has been applied:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.configuration}'
{"inline":{"conf.yaml":"query_defaults:\n  limit: 1000"}}

$ kubectl exec -n demo weaviate-sample-0 -c weaviate -- cat /weaviate-config/conf.yaml/conf.yaml
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
  limit: 1000
```

The output confirms the database is running with our inline `query_defaults.limit: 1000` setting.

> **Tip:** You can change the configuration of a running Weaviate cluster (and even reference a replacement config Secret) without recreating it using a `Reconfigure` ops request. See [Reconfigure Weaviate](/docs/guides/weaviate/reconfigure/reconfigure.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete weaviate -n demo weaviate-sample
$ kubectl delete secret -n demo weaviate-custom-config
$ kubectl delete ns demo
```
