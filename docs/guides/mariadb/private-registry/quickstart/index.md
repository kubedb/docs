---
title: Run MariaDB using Private Registry
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-privateregistry-quickstart
    name: Quickstart
    parent: guides-mariadb-privateregistry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy MariaDB from private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run MariaDB database using private Docker images.

## Before You Begin

- Read [concept of MariaDB Version Catalog](/docs/guides/mariadb/concepts/mariadb-version) to learn detail concepts of `MariaDBVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/u/kubedb) into your private registry. For mysql, push `DB_IMAGE`, `EXPORTER_IMAGE`, `INITCONTAINER_IMAGE` of following MariaDBVersions, where `deprecated` is not true, to your private registry.

```bash
$ kubectl get mariadbversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image,INITCONTAINER_IMAGE:.spec.initContainer.image,DEPRECATED:.spec.deprecated
NAME      VERSION   DB_IMAGE                 EXPORTER_IMAGE                   INITCONTAINER_IMAGE   DEPRECATED
10.4.32   10.4.32   kubedb/mariadb:10.4.32   kubedb/mysqld-exporter:v0.11.0   kubedb/busybox        <none>
10.5.23    10.5.23    kubedb/mariadb:10.5.23    kubedb/mysqld-exporter:v0.11.0   kubedb/busybox        <none>
```

Docker hub repositories:

- [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
- [kubedb/mariadb](https://hub.docker.com/r/kubedb/mariadb)
- [kubedb/mysqld-exporter](https://hub.docker.com/r/kubedb/mysqld-exporter)

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: MariaDBVersion
  metadata:
    name: 10.5.23
  spec:
    db:
      image: PRIVATE_REGISTRY/mysql:10.5.23
    exporter:
      image: PRIVATE_REGISTRY/mysqld-exporter:v0.11.0
    initContainer:
      image: PRIVATE_REGISTRY/busybox
    podSecurityPolicies:
      databasePolicyName: maria-db
    version: 10.5.23
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

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of Kubernetes.

NB: If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate value. Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that to points to the DOCKER_REGISTRY you wish to pull images from.

## Deploy MariaDB database from Private Registry

While deploying `MariaDB` from private repository, you have to add `myregistrykey` secret in `MariaDB` `spec.imagePullSecrets`.
Below is the MariaDB CRD object we will create.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: md-pvt-reg
  namespace: demo
spec:
  version: "10.5.23"
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
  deletionPolicy: WipeOut
```

Now run the command to deploy this `MariaDB` object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/private-registry/quickstart/examples/demo.yaml
mariadb.kubedb.com/md-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `MariaDB` is in running state:

```bash
$ kubectl get pods -n demo
NAME              READY     STATUS    RESTARTS   AGE
md-pvt-reg-0   1/1       Running   0          56s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mariadb -n demo md-pvt-reg
mariadb.kubedb.com "md-pvt-reg" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
