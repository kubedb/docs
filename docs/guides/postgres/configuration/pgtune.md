---
title: Run PostgreSQL with Auto Configuration Tuning
menu:
  docs_{{ .version }}:
    identifier: pg-auto-tuning-pgtune
    name: Auto Tuning (pgtune)
    parent: pg-configuration
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run PostgreSQL with Auto Configuration Tuning

Getting good PostgreSQL settings usually means knowing a lot about parameters like `shared_buffers`, `work_mem`, or `effective_cache_size`. KubeDB can do this for you. Instead of writing a configuration file by hand, you tell KubeDB a few simple things about your workload and how much CPU/memory the database gets, and it automatically calculates a sensible configuration for you.

This auto-tuning is powered by [pgtune](https://github.com/gregs1104/pgtune), a well-known tool for generating PostgreSQL configuration.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

To enable auto-tuning, you add a `spec.configuration.tuning` section to your `Postgres` object with three fields:

- the kind of workload you run (`profile`),
- the maximum number of connections you need (`maxConnections`),
- the kind of disk your storage uses (`storageType`).

KubeDB then looks at the **CPU and memory you gave the database** (in `spec.podTemplate`) and calculates a matching configuration. The result is written into a Kubernetes Secret that the database loads on startup — you don't have to manage any config file yourself.

This is an easier alternative to providing a full custom configuration file (see [Using a Config File](/docs/guides/postgres/configuration/using-config-file.md)).

## Tuning Fields

#### profile

Pick the profile that best matches your application.

| Profile   | Use it for                                                        |
|-----------|-------------------------------------------------------------------|
| `web`     | Web applications with many simple, short queries.                 |
| `oltp`    | Transaction-heavy applications with lots of concurrent writes.    |
| `dw`      | Data warehouse / analytics with large, complex queries.           |
| `mixed`   | A general-purpose mix of the above.                               |
| `desktop` | Development or desktop usage where the database isn't the only app.|

#### storageType

Tell KubeDB what kind of disk backs your storage so it can tune disk-related settings.

| Storage Type | Use it for                          |
|--------------|-------------------------------------|
| `ssd`        | Solid-state drives.                 |
| `hdd`        | Traditional spinning hard disks.    |
| `san`        | Network-attached storage (SAN).     |

#### maxConnections

The maximum number of concurrent connections your application needs (for example `200`). Memory-related settings are calculated around this number.

> **Important:** The CPU and memory you set in `spec.podTemplate.spec.containers[postgres].resources` directly drive the calculated values. KubeDB uses the resource `requests` (and falls back to `limits` if requests are not set). Give the database more memory and you'll get larger buffers and caches; give it more CPU (4 cores or more) and KubeDB also enables parallel query workers.

## Deploy PostgreSQL with Tuning

Below is the YAML of a 3-replica `Postgres` with auto-tuning enabled. Notice the `tuning` block and the `podTemplate` resources that feed it.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-ha-tuning
  namespace: demo
spec:
  version: "18.3"
  replicas: 3
  configuration:
    tuning:
      profile: web
      maxConnections: 200
      storageType: ssd
  podTemplate:
    spec:
      containers:
      - name: postgres
        resources:
          requests:
            cpu: 2
            memory: 2Gi
          limits:
            memory: 3Gi
  storageType: Durable
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 3Gi
  deletionPolicy: WipeOut
```

Let's create it:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/configuration/pg-tuning.yaml
postgres.kubedb.com/pg-ha-tuning created
```

Now wait for the database to become `Ready`:

```bash
$ kubectl get pg -n demo pg-ha-tuning
NAME           VERSION   STATUS   AGE
pg-ha-tuning   18.3      Ready    69s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=pg-ha-tuning
NAME             READY   STATUS    RESTARTS   AGE
pg-ha-tuning-0   2/2     Running   0          63s
pg-ha-tuning-1   2/2     Running   0          57s
pg-ha-tuning-2   2/2     Running   0          51s
```

## See the Generated Configuration

KubeDB stores the generated configuration in a Secret owned by the database. The secret name has a random suffix, so it will be different in your cluster. You can find it by listing the secrets for this database:

```bash
$ kubectl get secret -n demo | grep pg-ha-tuning
pg-ha-tuning-auth     kubernetes.io/basic-auth   2      52s
pg-ha-tuning-eba1da   Opaque                     1      52s
```

The `Opaque` secret (`pg-ha-tuning-eba1da` here) holds the tuned configuration under the `pgtune.conf` key. Let's print it:

```bash
$ kubectl get secret -n demo pg-ha-tuning-eba1da -o jsonpath='{.data.pgtune\.conf}' | base64 -d
# Tuned by KubeDB
# https://kubedb.com

# DB Version: 17
# OS Type: linux
# DB Type: web
# Total Memory (RAM): 2 GB
# CPUs num: 2
# Connections num: 200
# Data Storage: ssd

max_connections = 200
shared_buffers = 512MB
effective_cache_size = 1536MB
maintenance_work_mem = 128MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 2520kB
huge_pages = off
min_wal_size = 1GB
max_wal_size = 4GB
```

The comment block at the top shows exactly what inputs KubeDB used (database type, total memory, CPUs, connections and storage), and the values below are calculated from them.

> **Note:** In this example the database has 2 CPUs. Parallel-query settings (such as `max_parallel_workers`) are only added when the database has 4 CPUs or more, so they don't appear here. Give the database more CPU/memory and you'll see different values.

Finally, let's confirm the running database actually uses these values. We'll `exec` into a pod and use the [SHOW](https://www.postgresql.org/docs/current/sql-show.html) command:

```bash
$ kubectl exec -it -n demo pg-ha-tuning-0 -c postgres -- psql -U postgres -c "SHOW shared_buffers; SHOW max_connections;"
 shared_buffers
----------------
 512MB
(1 row)

 max_connections
-----------------
 200
(1 row)
```

The values match what KubeDB generated. 🎉

## Combining with Custom Configuration

Auto-tuning works alongside the existing ways of providing custom configuration:

- Tuned values are stored under the `pgtune.conf` key.
- Your own custom configuration (via `spec.configuration.secretName` or `spec.configuration.inline`) is stored separately under the `inline.conf` key.

This means you can let KubeDB handle the general tuning while still overriding specific parameters yourself.

If you want to turn auto-tuning off, set `disableAutoTune: true`:

```yaml
spec:
  configuration:
    tuning:
      profile: web
      maxConnections: 200
      storageType: ssd
      disableAutoTune: true
```

When disabled, KubeDB removes the `pgtune.conf` key and stops generating tuned values.

## Cleaning Up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pg/pg-ha-tuning
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn how to provide a full custom configuration file in [Using a Config File](/docs/guides/postgres/configuration/using-config-file.md).
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Want to setup a PostgreSQL cluster? Check how to [configure a Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md).
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
