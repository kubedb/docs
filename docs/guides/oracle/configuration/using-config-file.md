---
title: Run Oracle with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-configuration-using-config-file
    name: Config File
    parent: guides-oracle-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Oracle. This tutorial will show you how to run an Oracle database with a user provided configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/configuration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Make sure the `orclcred` secret exists in the `demo` namespace before deploying.

## Overview

Oracle allows configuring the database via a configuration file. When KubeDB provisions an Oracle database, each `KEY = value` entry from the supplied configuration is applied to the running database with an `ALTER SYSTEM SET KEY=value SCOPE=SPFILE;` statement, after which the instance is bounced so the new settings take effect. Typical tunables include `PROCESSES`, `SGA_TARGET`, `PGA_AGGREGATE_TARGET`, etc.

To provide a custom configuration for an Oracle database, you set `spec.configuration`. There are three options, which can be combined:

- **Secret** — point `spec.configuration.secretName` at a Kubernetes `Secret` whose key is **`oracle.cnf`**.
- **Inline** — put one or more configuration files directly under `spec.configuration.inline`.
- **Both** — supply a Secret *and* inline configuration. Inline values override the ones from the Secret.

> Note: for Secret-based configuration, the Secret key **must** be `oracle.cnf`. Inline configuration keys can use any file name.

When you provide inline configuration, you may include multiple files if needed. KubeDB applies inline files in lexicographical order by key name, so prefix file names with a priority number, such as `1-sga.cnf` and `2-backup-files.txt`, to control the processing order.

## Custom Configuration via a Secret

At first, create a Secret with your Oracle configuration. The Secret key must be `oracle.cnf`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oracle-custom-config
  namespace: demo
type: Opaque
stringData:
  oracle.cnf: |
    PROCESSES = 800
```

Let's create the Secret,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/configuration/oracle-custom-config-secret.yaml
secret/oracle-custom-config created
```

Now, create an `Oracle` CR that references this Secret through `spec.configuration.secretName`. The following is a **Standalone** example:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: standalone-cus-conf
  namespace: demo
spec:
  configuration:
    secretName: oracle-custom-config
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Here,

- `spec.configuration.secretName` is the name of the Secret (with key `oracle.cnf`) that holds the custom configuration.

Create the database,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/configuration/standalone-cus-conf.yaml
oracle.kubedb.com/standalone-cus-conf created
```

For a **DataGuard** cluster the configuration is supplied the same way — only `mode`, `replicas`, and the name change:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: dataguard-cus-conf
  namespace: demo
spec:
  configuration:
    secretName: oracle-custom-config
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: DataGuard
  storageType: Durable
  replicas: 3
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

## Custom Configuration via Inline

You can also provide the configuration inline, without creating a Secret. The following **Standalone** example sets `SGA_TARGET=1G` directly:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: standalone-inline-conf
  namespace: demo
spec:
  configuration:
    inline:
      1-sga.cnf: |
        SGA_TARGET=1G
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Here,

- `spec.configuration.inline` holds one or more configuration files directly in the `Oracle` CR. Inline file names can be any valid key name and are processed in lexicographical order.

## Combining Secret and Inline Configuration

Both sources can be used together. When the same parameter is set in both places, the **inline** value wins. The following **DataGuard** example takes `PROCESSES = 800` from the Secret and additionally sets `SGA_TARGET=5G` inline:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: dataguard-inline-conf
  namespace: demo
spec:
  configuration:
    secretName: oracle-custom-config
    inline:
      1-sga.cnf: |
        SGA_TARGET=5G
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: DataGuard
  storageType: Durable
  replicas: 3
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

## Verify the Configuration

Once the database is `Ready` (and the pod has printed the `DATABASE IS READY TO USE!!!` banner), let's verify that our custom configuration was applied. We will use the `standalone-cus-conf` database that references the `oracle-custom-config` Secret (`PROCESSES = 800`).

First, KubeDB projects the configuration file into the database pod at `/etc/config/oracle.cnf`,

```bash
$ kubectl exec -n demo standalone-cus-conf-0 -c oracle -- cat /etc/config/oracle.cnf
PROCESSES = 800
```

Now, let's connect to the database and confirm that the `processes` parameter has actually been set to `800`,

```bash
$ kubectl get secret -n demo standalone-cus-conf-auth -o jsonpath='{.data.password}' | base64 -d
# (use the printed password below)

$ kubectl exec -n demo standalone-cus-conf-0 -c oracle -- bash -lc \
    "echo -e 'SHOW PARAMETER processes;\nexit;' | sqlplus -s sys/<password>@localhost:1521/ORCL as sysdba"

NAME                                 TYPE        VALUE
------------------------------------ ----------- ------------------------------
aq_tm_processes                      integer     1
db_writer_processes                  integer     2
gcs_server_processes                 integer     0
global_txn_processes                 integer     1
job_queue_processes                  integer     308
log_archive_max_processes            integer     30
processes                            integer     800
```

The `processes` parameter is now `800`, confirming that our custom configuration was applied successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo oracle/standalone-cus-conf -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo standalone-cus-conf
kubectl delete secret -n demo oracle-custom-config
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Learn how to [reconfigure](/docs/guides/oracle/reconfigure/reconfigure.md) a running Oracle database with an `OracleOpsRequest`.
- Initialize [Oracle with Script](/docs/guides/oracle/initialization/script_source.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
