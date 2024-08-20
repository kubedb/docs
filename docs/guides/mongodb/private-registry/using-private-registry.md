---
title: Run MongoDB using Private Registry
menu:
  docs_{{ .version }}:
    identifier: mg-using-private-registry-private-registry
    name: Quickstart
    parent: mg-private-registry-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run MongoDB database using private Docker images.

## Before You Begin

- Read [concept of MongoDB Version Catalog](/docs/guides/mongodb/concepts/catalog.md) to learn detail concepts of `MongoDBVersion` object.

- you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images into your private registry. For mongodb, push `DB_IMAGE`, `TOOLS_IMAGE`, `EXPORTER_IMAGE` of following MongoDBVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get mongodbversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,INITCONTAINER_IMAGE:.spec.initContainer.image,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image
  NAME             VERSION   INITCONTAINER_IMAGE            DB_IMAGE                                 EXPORTER_IMAGE
  3.4.17-v1        3.4.17    kubedb/mongodb-init:4.1-v7     mongo:3.4.17                             kubedb/mongodb_exporter:v0.20.4
  3.4.22-v1        3.4.22    kubedb/mongodb-init:4.1-v7     mongo:3.4.22                             kubedb/mongodb_exporter:v0.32.0
  3.6.13-v1        3.6.13    kubedb/mongodb-init:4.1-v7     mongo:3.6.13                             kubedb/mongodb_exporter:v0.32.0
  4.4.26         3.6.8     kubedb/mongodb-init:4.1-v7     mongo:3.6.8                              kubedb/mongodb_exporter:v0.32.0
  4.0.11-v1        4.0.11    kubedb/mongodb-init:4.1-v7     mongo:4.0.11                             kubedb/mongodb_exporter:v0.32.0
  4.0.3-v1         4.0.3     kubedb/mongodb-init:4.1-v7     mongo:4.0.3                              kubedb/mongodb_exporter:v0.32.0
  4.4.26         4.0.5     kubedb/mongodb-init:4.1-v7     mongo:4.0.5                              kubedb/mongodb_exporter:v0.32.0
  4.4.26        4.1.13    kubedb/mongodb-init:4.2-v7     mongo:4.1.13                             kubedb/mongodb_exporter:v0.32.0
  4.1.4-v1         4.1.4     kubedb/mongodb-init:4.1.4-v7   mongo:4.1.4                              kubedb/mongodb_exporter:v0.32.0
  4.1.7-v3         4.1.7     kubedb/mongodb-init:4.2-v7     mongo:4.1.7                              kubedb/mongodb_exporter:v0.32.0
  4.4.26            4.4.26     kubedb/mongodb-init:4.2-v7     mongo:4.4.26                              kubedb/mongodb_exporter:v0.32.0
  4.4.26            4.4.26     kubedb/mongodb-init:4.2-v7     mongo:4.4.26                              kubedb/mongodb_exporter:v0.32.0
  5.0.2            5.0.2     kubedb/mongodb-init:4.2-v7     mongo:5.0.2                              kubedb/mongodb_exporter:v0.32.0
  5.0.3            5.0.3     kubedb/mongodb-init:4.2-v7     mongo:5.0.3                              kubedb/mongodb_exporter:v0.32.0
  percona-3.6.18   3.6.18    kubedb/mongodb-init:4.1-v7     percona/percona-server-mongodb:3.6.18    kubedb/mongodb_exporter:v0.32.0
  percona-4.0.10   4.0.10    kubedb/mongodb-init:4.1-v7     percona/percona-server-mongodb:4.0.10    kubedb/mongodb_exporter:v0.32.0
  percona-4.2.7    4.2.7     kubedb/mongodb-init:4.2-v7     percona/percona-server-mongodb:4.2.7-7   kubedb/mongodb_exporter:v0.32.0
  percona-4.4.10   4.4.10    kubedb/mongodb-init:4.2-v7     percona/percona-server-mongodb:4.4.10    kubedb/mongodb_exporter:v0.32.0
  ```

  Docker hub repositories:

  - [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
  - [kubedb/mongo](https://hub.docker.com/r/kubedb/mongo)
  - [kubedb/mongo-tools](https://hub.docker.com/r/kubedb/mongo-tools)
  - [kubedb/mongodb_exporter](https://hub.docker.com/r/kubedb/mongodb_exporter)


## Install KubeDB operator from Private Registry

If you want to install KubeDB operator with some private registry images, set the flags `--registry` and `--imagePullSecret` to appropriate value, when installing the operator.
Follow the steps [install KubeDB operator](/docs/setup/README.md) properly. The list configuration arguments of the helm installation command will be found [here](https://github.com/kubedb/installer/tree/v2022.10.18/charts/kubedb#configuration).


## Use DB related images from Private Registry

- Update KubeDB catalog for private Docker registry. Ex:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MongoDBVersion
metadata:
  name: "4.4.26"
  labels:
    app: kubedb
spec:
  version: "4.4.26"
  db:
    image: "PRIVATE_DOCKER_REGISTRY/mongo:4.4.26"
  exporter:
    image: "PRIVATE_DOCKER_REGISTRY/percona-mongodb-exporter:v0.8.0"
  initContainer:
    image: "PRIVATE_DOCKER_REGISTRY/mongodb-init:4.2"
  podSecurityPolicies:
    databasePolicyName: mongodb-db
  replicationModeDetector:
    image: "PRIVATE_DOCKER_REGISTRY/replication-mode-detector:v0.3.2"
```

### Create ImagePullSecret

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

DOCKER_REGISTRY_SERVER value will be `docker.io` for docker hub.

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of Kubernetes.

NB: If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

### Create Demo namespace

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created
```

### Deploy MongoDB

While deploying `MongoDB` from private repository, you have to add `myregistrykey` secret in `MongoDB` `spec.imagePullSecrets`.
Below is the MongoDB CRD object we will create.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mgo-pvt-reg
  namespace: demo
spec:
  version: 4.4.26
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

Now run the command to deploy this `MongoDB` object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/private-registry/replicaset.yaml
mongodb.kubedb.com/mgo-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `MongoDB` is in running state:

```bash
$ kubectl get pods -n demo 
NAME            READY     STATUS              RESTARTS   AGE
mgo-pvt-reg-0   1/1       Running             0          5m


$ kubectl get mg -n demo
NAME          VERSION   STATUS    AGE
mgo-pvt-reg   4.4.26     Ready     38s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mgo-pvt-reg -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mgo-pvt-reg

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/stash/overview/index.md) MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
