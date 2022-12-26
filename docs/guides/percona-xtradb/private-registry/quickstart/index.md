---
title: Run PerconaXtraDB using Private Registry
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-privateregistry-quickstart
    name: Quickstart
    parent: guides-perconaxtradb-privateregistry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy PerconaXtraDB from private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run PerconaXtraDB database using private Docker images.

## Before You Begin

- Read [concept of PerconaXtraDB Version Catalog](/docs/guides/percona-xtradb/concepts/perconaxtradb-version) to learn detail concepts of `PerconaXtraDBVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/u/kubedb) into your private registry. For perconaxtradb, push `DB_IMAGE`, `EXPORTER_IMAGE`, `INITCONTAINER_IMAGE` of following PerconaXtraDBVersions, where `deprecated` is not true, to your private registry.

```bash
$ kubectl get perconaxtradbversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image,INITCONTAINER_IMAGE:.spec.initContainer.image,DEPRECATED:.spec.deprecated
NAME     VERSION   DB_IMAGE                                EXPORTER_IMAGE                 INITCONTAINER_IMAGE                DEPRECATED
8.0.26   8.0.26    percona/percona-xtradb-cluster:8.0.26   prom/mysqld-exporter:v0.13.0   kubedb/percona-xtradb-init:0.2.0   <none>
8.0.28   8.0.28    percona/percona-xtradb-cluster:8.0.28   prom/mysqld-exporter:v0.13.0   kubedb/percona-xtradb-init:0.2.0   <none>
```

Docker hub repositories:

- [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
- [kubedb/perconaxtradb](https://hub.docker.com/r/percona/percona-xtradb-cluster)
- [kubedb/mysqld-exporter](https://hub.docker.com/r/kubedb/mysqld-exporter)

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: PerconaXtraDBVersion
  metadata:
    name: 8.0.26
  spec:
    db:
      image: PRIVATE_REGISTRY/mysql:8.0.26
    exporter:
      image: PRIVATE_REGISTRY/mysqld-exporter:v0.11.0
    initContainer:
      image: PRIVATE_REGISTRY/busybox
    podSecurityPolicies:
      databasePolicyName: perconaxtra-db
    version: 8.0.26
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
$ kubectl create secret docker-registry -n demo pxregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/pxregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of Kubernetes.

NB: If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate value. Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that to points to the DOCKER_REGISTRY you wish to pull images from.

## Deploy PerconaXtraDB database from Private Registry

While deploying `PerconaXtraDB` from private repository, you have to add `pxregistrykey` secret in `PerconaXtraDB` `spec.imagePullSecrets`.
Below is the PerconaXtraDB CRD object we will create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: px-pvt-reg
  namespace: demo
spec:
  version: "8.0.26"
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
      - name: pxregistrykey
  terminationPolicy: WipeOut
```

Now run the command to deploy this `PerconaXtraDB` object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/private-registry/quickstart/examples/demo.yaml
perconaxtradb.kubedb.com/px-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `PerconaXtraDB` is in running state:

```bash
$ kubectl get pods -n demo
NAME              READY     STATUS    RESTARTS   AGE
px-pvt-reg-0   1/1       Running   0          56s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo px-pvt-reg
perconaxtradb.kubedb.com "px-pvt-reg" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
