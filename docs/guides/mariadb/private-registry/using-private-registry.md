---
title: Run MariaDB using Private Registry
menu:
  docs_{{ .version }}:
    identifier: my-using-private-registry-private-registry
    name: Quickstart
    parent: my-private-registry-mariadb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy MariaDB from private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run MariaDB database using private Docker images.

## Before You Begin

- Read [concept of MariaDB Version Catalog](/docs/guides/mariadb/concepts/catalog.md) to learn detail concepts of `MariaDBVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For mariadb, push `DB_IMAGE`, `EXPORTER_IMAGE`, `REPLICATION_MODE_DETECTOR_IMAGE`(only required for Group Replication), `INITCONTAINER_IMAGE` of following MariaDBVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get mariadbversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image,REPLICATION_MODE_DETECTOR_IMAGE:.spec.replicationModeDetector.image,INITCONTAINER_IMAGE:.spec.initContainer.image,DEPRECATED:.spec.deprecated
  NAME        VERSION   DB_IMAGE                   EXPORTER_IMAGE                     REPLICATION_MODE_DETECTOR_IMAGE                          INITCONTAINER_IMAGE   DEPRECATED
  5           5         kubedb/mariadb:5           kubedb/operator:0.8.  0            kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1     kubedb/busybox        true
  5-v1        5         kubedb/mariadb:5-v1        kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  5.7         5.7       kubedb/mariadb:5.7         kubedb/operator:0.8.  0            kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1     kubedb/busybox        true
  5.7-v1      5.7       kubedb/mariadb:5.7-v1      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  5.7-v2      5.7.25    kubedb/mariadb:5.7-v2      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  5.7-v3      5.7.25    kubedb/mariadb:5.7.25      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  5.7-v4      5.7.29    kubedb/mariadb:5.7.29      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  5.7.25      5.7.25    kubedb/mariadb:5.7.25      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  5.7.25-v1   5.7.25    kubedb/mariadb:5.7.25-v1   kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  5.7.29      5.7.29    kubedb/mariadb:5.7.29      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  5.7.31      5.7.31    kubedb/mariadb:5.7.31      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  8           8         kubedb/mariadb:8           kubedb/operator:0.8.  0            kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1     kubedb/busybox        true
  8-v1        8         kubedb/mariadb:8-v1        kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  8.0         8.0       kubedb/mariadb:8.0         kubedb/operator:0.8.  0            kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1     kubedb/busybox        true
  8.0-v1      8.0.3     kubedb/mariadb:8.0-v1      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  8.0-v2      8.0.14    kubedb/mariadb:8.0-v2      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  8.0-v3      8.0.20    kubedb/mariadb:8.0.20      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  8.0.14      8.0.14    kubedb/mariadb:8.0.14      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  8.0.14-v1   8.0.14    kubedb/mariadb:8.0.14-v1   kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  8.0.20      8.0.20    kubedb/mariadb:8.0.20      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  8.0.21      8.0.21    kubedb/mariadb:8.0.21      kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  8.0.3       8.0.3     kubedb/mariadb:8.0.3       kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        true
  8.0.3-v1    8.0.3     kubedb/mariadb:8.0.3-v1    kubedb/mariadbd-exporter:v0.  11.0   kubedb/mariadb-replication-mode-detector:v0.1.0-beta.1   kubedb/  busybox        <none>
  ```

  Docker hub repositories:
  - [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
  - [kubedb/mariadb](https://hub.docker.com/r/kubedb/mariadb)
  - [kubedb/mariadb-tools](https://hub.docker.com/r/kubedb/mariadb-tools)
  - [kubedb/mariadbd-exporter](https://hub.docker.com/r/kubedb/mariadbd-exporter)

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: MariaDBVersion
  metadata:
    name: "8.0.21"
    labels:
      app: kubedb
  spec:
    version: "8.0.21"
    db:
      image: "PRIVATE_DOCKER_REGISTRY/mariadb:8.0.21"
    exporter:
      image: "PRIVATE_DOCKER_REGISTRY/mariadbd-exporter:v0.11.0"
    initContainer:
      image: "PRIVATE_DOCKER_REGISTRY/busybox"
    replicationModeDetector:
      image: "PRIVATE_DOCKER_REGISTRY/mariadb-replication-mode-detector:v0.1.0-beta.1"
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
   ```

## Create ImagePullSecret

ImagePullSecrets is a type of a Kubernete Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

Run the following command, substituting the appropriate uppercase values to create an image pull secret for your private Docker registry:

```bash
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of kubernetes.

NB: If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate value. Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that to points to the DOCKER_REGISTRY you wish to pull images from.

## Deploy MariaDB database from Private Registry

While deploying `MariaDB` from private repository, you have to add `myregistrykey` secret in `MariaDB` `spec.imagePullSecrets`.
Below is the MariaDB CRD object we will create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: mariadb-pvt-reg
  namespace: demo
spec:
  version: "8.0.21"
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

Now run the command to deploy this `MariaDB` object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mariadb/private-registry/demo-2.yaml
mariadb.kubedb.com/mariadb-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `MariaDB` is in running state:

```bash
$ kubectl get pods -n demo
NAME              READY     STATUS    RESTARTS   AGE
mariadb-pvt-reg-0   1/1       Running   0          56s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mariadb/mariadb-pvt-reg -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mariadb/mariadb-pvt-reg

kubectl patch -n demo drmn/mariadb-pvt-reg -p '{"spec":{"wipeOut":true}}' --type="merge"
kubectl delete -n demo drmn/mariadb-pvt-reg

kubectl delete ns demo
```

## Next Steps

- Initialize [MariaDB with Script](/docs/guides/mariadb/initialization/using-script.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mariadb/monitoring/using-prometheus-operator.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mariadb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb.md).
- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
