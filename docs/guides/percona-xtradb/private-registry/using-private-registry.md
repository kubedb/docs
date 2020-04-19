---
title: Percona XtraDB using Private Registry Guide
menu:
  docs_{{ .version }}:
    identifier: px-private-registry-guide
    name: Private Registry Guide
    parent: px-private-registry
    weight: 10
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Deploy Percona XtraDB from private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run Percona XtraDB database using private Docker images.

## Before You Begin

- Read about [PerconaXtraDBVersion](/docs/concepts/catalog/percona-xtradb.md) to learn how it is used.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For PerconaXtraDB, push `DB_IMAGE`, `EXPORTER_IMAGE` of following PerconaXtraDBVersions, where `deprecated` is not true, to your private registry.

  ```console
  $ kubectl get perconaxtradbversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image,DEPRECATED:.spec.deprecated
  NAME          VERSION   DB_IMAGE                            EXPORTER_IMAGE                   DEPRECATED
  5.7           5.7       kubedb/percona:5.7                  kubedb/mysqld-exporter:v0.11.0   <none>
  5.7-private   5.7       kubedb/percona-xtradb-cluster:5.7   kubedb/mysqld-exporter:v0.11.0   <none>
  ```

  Docker hub repositories:

  - [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
  - [kubedb/percona-xtradb](https://hub.docker.com/r/kubedb/percona-xtradb)
  - [kubedb/mysqld-exporter](https://hub.docker.com/r/kubedb/mysqld-exporter)

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: PerconaXtraDBVersion
  metadata:
    name: "5.7-private"
    labels:
      app: kubedb
  spec:
    version: "5.7"
    db:
      image: "<private-docker-registry>/percona-xtradb-cluster:5.7"
    exporter:
      image: "<private-docker-registry>/mysqld-exporter:v0.11.0"
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```console
  $ kubectl create ns demo
  namespace/demo created
   ```

## Create ImagePullSecret

ImagePullSecrets is a type of a Kubernetes Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

Run the following command, substituting the appropriate uppercase values to create an image pull secret for your private Docker registry:

```console
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of kubernetes.

> NB: If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate value. Follow the steps to [install KubeDB operator](/docs/setup/install.md) properly in cluster so that it points to the Docker registry you wish to pull images from.

## Deploy Percona XtraDB database from Private Registry

While deploying `PerconaXtraDB` from private repository, you have to add `myregistrykey` secret in the `.spec.imagePullSecrets` field of the `PerconaXtraDB` object.
Below is the `PerconaXtraDB` object we will create.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PerconaXtraDB
metadata:
  name: px-pvt-reg
  namespace: demo
spec:
  version: "5.7-private"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
  updateStrategy:
    type: "RollingUpdate"
  terminationPolicy: WipeOut
```

Now run the command to deploy this `PerconaXtraDB` object:

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/private-registry.yaml
perconaxtradb.kubedb.com/px-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `PerconaXtraDB` is in running state:

```console
$ kubectl get pods -n demo
NAME           READY     STATUS    RESTARTS   AGE
px-pvt-reg-0   1/1       Running   0          56s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo perconaxtradb/px-pvt-reg -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo perconaxtradb/px-pvt-reg

kubectl delete ns demo
```

## Next Steps

- Initialize [PerconaXtraDB with Script](/docs/guides/percona-xtradb/initialization/using-script.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-custom-config.md).
- How to use [custom rbac resource](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) for PerconaXtraDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/snapshot/stash.md).
- Detail concepts of [PerconaXtraDB object](/docs/concepts/databases/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/concepts/catalog/percona-xtradb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
