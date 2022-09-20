---
title: Run PostgreSQL using Private Registry
menu:
  docs_{{ .version }}:
    identifier: pg-using-private-registry-private-registry
    name: Quickstart
    parent: pg-private-registry-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using private Docker registry

KubeDB supports using private Docker registry. This tutorial will show you how to run KubeDB managed PostgreSQL database using private Docker images.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/postgres](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/postgres) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Private Docker Registry

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories). In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For postgres, push `DB_IMAGE`, `TOOLS_IMAGE`, `EXPORTER_IMAGE` of following PostgresVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get postgresversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,TOOLS_IMAGE:.spec.tools.image,EXPORTER_IMAGE:.spec.exporter.image,DEPRECATED:.spec.deprecated
  NAME       VERSION   DB_IMAGE                   TOOLS_IMAGE                      EXPORTER_IMAGE                    DEPRECATED
  10.2       10.2      kubedb/postgres:10.2       kubedb/postgres-tools:10.2       kubedb/operator:0.8.0             true
  10.2-v1    10.2      kubedb/postgres:10.2-v2    kubedb/postgres-tools:10.2-v2    kubedb/postgres_exporter:v0.4.6   true
  10.2-v2    10.2      kubedb/postgres:10.2-v3    kubedb/postgres-tools:10.2-v3    kubedb/postgres_exporter:v0.4.7   <none>
  10.2-v3    10.2      kubedb/postgres:10.2-v4    kubedb/postgres-tools:10.2-v3    kubedb/postgres_exporter:v0.4.7   <none>
  10.2-v4    10.2      kubedb/postgres:10.2-v5    kubedb/postgres-tools:10.2-v3    kubedb/postgres_exporter:v0.4.7   <none>
  10.2-v5    10.2      kubedb/postgres:10.2-v6    kubedb/postgres-tools:10.2-v3    kubedb/postgres_exporter:v0.4.7   <none>
  10.6       10.6      kubedb/postgres:10.6       kubedb/postgres-tools:10.6       kubedb/postgres_exporter:v0.4.7   <none>
  10.6-v1    10.6      kubedb/postgres:10.6-v1    kubedb/postgres-tools:10.6       kubedb/postgres_exporter:v0.4.7   <none>
  10.6-v2    10.6      kubedb/postgres:10.6-v2    kubedb/postgres-tools:10.6       kubedb/postgres_exporter:v0.4.7   <none>
  10.6-v3    10.6      kubedb/postgres:10.6-v3    kubedb/postgres-tools:10.6       kubedb/postgres_exporter:v0.4.7   <none>
  11.1       11.1      kubedb/postgres:11.1       kubedb/postgres-tools:11.1       kubedb/postgres_exporter:v0.4.7   <none>
  11.1-v1    11.1      kubedb/postgres:11.1-v1    kubedb/postgres-tools:11.1       kubedb/postgres_exporter:v0.4.7   <none>
  11.1-v2    11.1      kubedb/postgres:11.1-v2    kubedb/postgres-tools:11.1       kubedb/postgres_exporter:v0.4.7   <none>
  11.1-v3    11.1      kubedb/postgres:11.1-v3    kubedb/postgres-tools:11.1       kubedb/postgres_exporter:v0.4.7   <none>
  11.2       11.2      kubedb/postgres:11.2       kubedb/postgres-tools:11.2       kubedb/postgres_exporter:v0.4.7   <none>
  11.2-v1    11.2      kubedb/postgres:11.2-v1    kubedb/postgres-tools:11.2       kubedb/postgres_exporter:v0.4.7   <none>
  9.6        9.6       kubedb/postgres:9.6        kubedb/postgres-tools:9.6        kubedb/operator:0.8.0             true
  9.6-v1     9.6       kubedb/postgres:9.6-v2     kubedb/postgres-tools:9.6-v2     kubedb/postgres_exporter:v0.4.6   true
  9.6-v2     9.6       kubedb/postgres:9.6-v3     kubedb/postgres-tools:9.6-v3     kubedb/postgres_exporter:v0.4.7   <none>
  9.6-v3     9.6       kubedb/postgres:9.6-v4     kubedb/postgres-tools:9.6-v3     kubedb/postgres_exporter:v0.4.7   <none>
  9.6-v4     9.6       kubedb/postgres:9.6-v5     kubedb/postgres-tools:9.6-v3     kubedb/postgres_exporter:v0.4.7   <none>
  9.6-v5     9.6       kubedb/postgres:9.6-v6     kubedb/postgres-tools:9.6-v3     kubedb/postgres_exporter:v0.4.7   <none>
  9.6.7      9.6.7     kubedb/postgres:9.6.7      kubedb/postgres-tools:9.6.7      kubedb/operator:0.8.0             true
  9.6.7-v1   9.6.7     kubedb/postgres:9.6.7-v2   kubedb/postgres-tools:9.6.7-v2   kubedb/postgres_exporter:v0.4.6   true
  9.6.7-v2   9.6.7     kubedb/postgres:9.6.7-v3   kubedb/postgres-tools:9.6.7-v3   kubedb/postgres_exporter:v0.4.7   <none>
  9.6.7-v3   9.6.7     kubedb/postgres:9.6.7-v4   kubedb/postgres-tools:9.6.7-v3   kubedb/postgres_exporter:v0.4.7   <none>
  9.6.7-v4   9.6.7     kubedb/postgres:9.6.7-v5   kubedb/postgres-tools:9.6.7-v3   kubedb/postgres_exporter:v0.4.7   <none>
  9.6.7-v5   9.6.7     kubedb/postgres:9.6.7-v6   kubedb/postgres-tools:9.6.7-v3   kubedb/postgres_exporter:v0.4.7   <none>
  ```

  Docker hub repositories:

- [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
- [kubedb/postgres](https://hub.docker.com/r/kubedb/postgres)
- [kubedb/postgres-tools](https://hub.docker.com/r/kubedb/postgres-tools)
- [kubedb/postgres_exporter](https://hub.docker.com/r/kubedb/postgres_exporter)

```bash
```

## Create ImagePullSecret

ImagePullSecrets is a type of a Kubernetes Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

Run the following command, substituting the appropriate uppercase values to create an image pull secret for your private Docker registry:

```bash
$ kubectl create secret generic -n demo docker-registry myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of Kubernetes.

> Note; If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate value.
Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that to points to the DOCKER_REGISTRY you wish to pull images from.

## Create PostgresVersion CRD

KubeDB uses images specified in PostgresVersion crd for database, backup and exporting prometheus metrics. You have to create a PostgresVersion crd specifying images from your private registry. Then, you have to point this PostgresVersion crd in `spec.version` field of Postgres object. For more details about PostgresVersion crd, please visit [here](/docs/guides/postgres/concepts/catalog.md).

Here, is an example of PostgresVersion crd. Replace `<YOUR_PRIVATE_REGISTRY>` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PostgresVersion
metadata:
  name: "13.2"
spec:
  coordinator:
    image: PRIVATE_REGISTRY/pg-coordinator:v0.1.0
  db:
    image: PRIVATE_REGISTRY/postgres:13.2-alpine
  distribution: PostgreSQL
  exporter:
    image: PRIVATE_REGISTRY/postgres-exporter:v0.9.0
  initContainer:
    image: PRIVATE_REGISTRY/postgres-init:0.1.0
  podSecurityPolicies:
    databasePolicyName: postgres-db
  stash:
    addon:
      backupTask:
        name: postgres-backup-13.1
      restoreTask:
        name: postgres-restore-13.1
  version: "13.2"
```

Now, create the PostgresVersion crd,

```bash
$ kubectl apply -f pvt-postgresversion.yaml
postgresversion.kubedb.com/pvt-10.2 created
```

## Deploy PostgreSQL database from Private Registry

While deploying PostgreSQL from private repository, you have to add `myregistrykey` secret in Postgres `spec.podTemplate.spec.imagePullSecrets` and specify `pvt-10.2` in `spec.version` field.

Below is the Postgres object we will create in this tutorial

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: pvt-reg-postgres
  namespace: demo
spec:
  version: "13.2"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
```

Now run the command to create this Postgres object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/private-registry/pvt-reg-postgres.yaml
postgres.kubedb.com/pvt-reg-postgres created
```

To check if the images pulled successfully from the repository, see if the PostgreSQL is in Running state:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=pvt-reg-postgres"
NAME                 READY     STATUS    RESTARTS   AGE
pvt-reg-postgres-0   1/1       Running   0          3m
```

## Snapshot

You can specify `imagePullSecret` for Snapshot objects in `spec.podTemplate.spec.imagePullSecrets` field of Snapshot object. If you are using scheduled backup, you can also provide `imagePullSecret` in `backupSchedule.podTemplate.spec.imagePullSecrets` field of Postgres crd. KubeDB also reuses `imagePullSecret` for Snapshot object from `spec.podTemplate.spec.imagePullSecrets` field of Postgres crd.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo pg/pvt-reg-postgres -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo pg/pvt-reg-postgres

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
