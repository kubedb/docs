---
title: Run PgBouncer with Custom Configuration File
menu:
  docs_{{ .version }}:
    identifier: pb-configuration-usingconfigfile
    name: Config File
    parent: pb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for PgBouncer. This tutorial will show you how to use KubeDB to run a PgBouncer with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You will need a PostgreSQL server for PgBouncer to connect to. You can prepare one by following the [PgBouncer quickstart](/docs/guides/pgbouncer/quickstart/quickstart.md) tutorial. In this tutorial, we will use a Postgres named `quick-postgres` in the `demo` namespace.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created

  $ kubectl get ns demo
  NAME    STATUS   AGE
  demo    Active   5s
  ```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/pgbouncer/configuration/using-config-file/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

PgBouncer is configured via an ini-style configuration file, `pgbouncer.ini`. KubeDB lets you provide a custom `pgbouncer.ini` file through `spec.configuration.secretName`. The operator reads this Secret internally and merges its content with the configuration generated from `spec.connectionPool`; settings from the custom config file take precedence over the generated defaults.

In this tutorial, we will configure [`auth_type`](https://www.pgbouncer.org/config.html#auth_type) and [`max_client_conn`](https://www.pgbouncer.org/config.html#max_client_conn) via a custom config file. We will use a Secret as the configuration source.

## Custom Configuration

At first, let's create `pgbouncer.ini` file setting `auth_type` and `max_client_conn` parameters.

```bash
cat <<EOF > pgbouncer.ini
[pgbouncer]
auth_type = scram-sha-256
max_client_conn = 100
EOF

$ cat pgbouncer.ini
[pgbouncer]
auth_type = scram-sha-256
max_client_conn = 100
```

Now, create a Secret with this configuration file.

```bash
$ kubectl create secret generic -n demo pb-configuration --from-file=./pgbouncer.ini
secret/pb-configuration created
```

Verify the Secret has the configuration file.

```yaml
$ kubectl get secret -n demo pb-configuration -o yaml
apiVersion: v1
stringData:
  pgbouncer.ini: |
    [pgbouncer]
    auth_type = scram-sha-256
    max_client_conn = 100
kind: Secret
metadata:
  name: pb-configuration
  namespace: demo
  ...
```

Now, create PgBouncer crd specifying `spec.configuration.secretName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/pgbouncer/configuration/using-config-file/examples/pb-custom.yaml
pgbouncer.kubedb.com/sample-pgbouncer created
```

Below is the YAML for the PgBouncer crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: sample-pgbouncer
  namespace: demo
spec:
  version: "1.18.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
      namespace: demo
  connectionPool:
    port: 5432
  configuration:
    secretName: pb-configuration
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-pgbouncer-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                 READY   STATUS    RESTARTS   AGE
sample-pgbouncer-0   1/1     Running   0          45s

$ kubectl get pgbouncer -n demo
NAME               TYPE            VERSION   STATUS   AGE
sample-pgbouncer   kubedb.com/v1   1.18.0    Ready    71s
```

We can see the PgBouncer is in `Ready` phase so it can accept connections.

Now, we will check if PgBouncer has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sample-pgbouncer-0 -- bash
pgbouncer@sample-pgbouncer-0:/$ cat etc/config/pgbouncer.ini
[databases]
postgres = host=quick-postgres.demo.svc port=5432 dbname=postgres

[pgbouncer]
max_client_conn = 100
...
auth_type = scram-sha-256
...
pgbouncer@sample-pgbouncer-0:/$ exit
exit
```

We can see that the values of `auth_type` and `max_client_conn` are the same as we provided in the custom config file.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete pgbouncer -n demo sample-pgbouncer
pgbouncer.kubedb.com "sample-pgbouncer" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
