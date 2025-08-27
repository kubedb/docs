---
title: Configuring clickhouse Using Config File
menu:
  docs_{{ .version }}:
    identifier: ch-configuration-using-config-file
    name: Configure Using Config File
    parent: ch-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for ClickHouse. This tutorial will show you how to use KubeDB to run a ClickHouse with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

ClickHouse allows configuring via configuration file. The default configuration file for ClickHouse deployed by `KubeDB` can be found in `/etc/clickhouse-server/config.xml`. When `spec.configSecret` is set to clickhouse, KubeDB operator will get the secret and after that it will validate the values of the secret and then will keep the validated customizable configurations from the user and merge it with the remaining default config. After all that this secret will be mounted to clickhouse for use it as the configuration file.

> To learn available configuration option of ClickHouse see [Configuration Options](https://clickhouse.com/docs/operations/configuration-files).

At first, you have to create a secret with your configuration file contents as the value of this key `clickhouse.yaml`. Then, you have to specify the name of this secret in `spec.configSecret.name` section while creating clickhouse CRO.

## Custom Configuration

At first, create `clickhouse.yaml` file containing required configuration settings.

```bash
$ cat clickhouse-config.yaml
profiles:
      default:
        max_query_size: 200000
```

Now, create the secret with this configuration file.

```bash
➤ kubectl create secret generic -n demo clickhouse-configuration --from-file=./clickhouse-config.yaml
secret/clickhouse-configuration created
```

Verify the secret has the configuration file.

```bash
➤ kubectl get secret -n demo clickhouse-configuration -oyaml
apiVersion: v1
data:
  clickhouse.yaml: cHJvZmlsZXM6CiAgZGVmYXVsdDoKICAgIG1heF9xdWVyeV9zaXplOiAxNTAwMDA=
kind: Secret
metadata:
  creationTimestamp: "2025-08-20T12:05:24Z"
  name: clickhouse-configuration
  namespace: demo
  resourceVersion: "199185"
  uid: a3439cc2-af41-441a-ad07-56572c86b9c2
type: Opaque

➤ echo cHJvZmlsZXM6CiAgZGVmYXVsdDoKICAgIG1heF9xdWVyeV9zaXplOiAxNTAwMDA= | base64 -d
profiles:
  default:
    max_query_size: 200000
```

Now, create clickhouse crd specifying `spec.configSecret` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: ch-standalone
  namespace: demo
spec:
  version: 24.4.1
  configSecret:
    name: clickhouse-configuration
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
➤ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/configuration/ch-custom-config-standalone.yaml
clickhouse.kubedb.com/ch-standalone created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `ch-standalone-0` has been created.

Check that the petset's pod is running

```bash
➤ kubectl get pod -n demo
NAME              READY   STATUS    RESTARTS   AGE
ch-standalone-0   1/1     Running   0          21m

```

Now, we will check if the clickhouse has started with the custom configuration we have provided.

Now, you can exec into the clickhouse pod and find if the custom configuration is there,

```bash
➤ kubectl exec -it -n demo ch-standalone-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@ch-standalone-0:/$ cd /etc/clickhouse-server/conf.d
clickhouse@ch-standalone-0:/etc/clickhouse-server/conf.d$ ls
clickhouse-config.yaml	server-config.yaml
clickhouse@ch-standalone-0:/etc/clickhouse-server/conf.d$ cat clickhouse-config.yaml 
profiles:
      default:
        max_query_size: 200000
clickhouse@ch-standalone-0:/etc/clickhouse-server/conf.d$ exit
exit

```

As we can see from the configuration of running clickhouse, the value of `max_query_size` has been set to our desired value successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete ch -n demo ch-standalone
kubectl delete secret -n demo clickhouse-configuration 
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Detail concepts of [ClickHouseVersion object](/docs/guides/clickhouse/concepts/clickhouseversion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
