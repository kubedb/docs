---
title: Run MySQL using Private Registry
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-private-registry
    name: Private Registry
    parent: guides-mysql
    weight: 35
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy MySQL from private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run MySQL database using private Docker images.

## Before You Begin

- Read [concept of MySQL Version Catalog](/docs/guides/mysql/concepts/catalog/index.md) to learn detail concepts of `MySQLVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For mysql, push `DB_IMAGE`, `EXPORTER_IMAGE`, `REPLICATION_MODE_DETECTOR_IMAGE`(only required for Group Replication), `INITCONTAINER_IMAGE` of following MySQLVersions, where `deprecated` is not true, to your private registry.

```bash
$ kubectl get mysqlversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image,REPLICATION_MODE_DETECTOR_IMAGE:.spec.replicationModeDetector.image,INITCONTAINER_IMAGE:.spec.initContainer.image,DEPRECATED:.spec.deprecated
NAME            VERSION   DB_IMAGE                    EXPORTER_IMAGE                   REPLICATION_MODE_DETECTOR_IMAGE            INITCONTAINER_IMAGE                    DEPRECATED
5.7.35-v1       5.7.35    mysql:5.7.35                kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:5.7-v2               <none>
5.7.36          5.7.36    mysql:5.7.36                kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:5.7-v2               <none>
8.0.17          8.0.17    mysql:8.0.17                kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:8.0.3-v1             <none>
8.0.27          8.0.27    mysql:8.0.27                kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:8.0.26-v1            <none>
8.0.27-innodb   8.0.27    mysql/mysql-server:8.0.27   kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:8.0.26-v1            <none>
8.0.29          8.0.29    mysql:8.0.29                kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:8.0.29_linux_amd64   <none>
8.0.3-v4        8.0.3     mysql:8.0.3                 kubedb/mysqld-exporter:v0.13.1   kubedb/replication-mode-detector:v0.13.0   kubedb/mysql-init:8.0.3-v1             <none>

```

  Docker hub repositories:
  - [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
  - [kubedb/mysql](https://hub.docker.com/r/kubedb/mysql)
  - [kubedb/mysql-tools](https://hub.docker.com/r/kubedb/mysql-tools)
  - [kubedb/mysqld-exporter](https://hub.docker.com/r/kubedb/mysqld-exporter)

- Update KubeDB catalog for private Docker registry. Ex:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  name: 8.0.29
spec:
  coordinator:
    image: PRIVATE_REGISTRY/mysql-coordinator:v0.4.0-2-g49a2d26-dirty_linux_amd64
  db:
    image: PRIVATE_REGISTRY/mysql:8.0.29
  distribution: Official
  exporter:
    image: PRIVATE_REGISTRY/mysqld-exporter:v0.13.1
  initContainer:
    image: PRIVATE_REGISTRY/mysql-init:8.0.29_linux_amd64
  podSecurityPolicies:
    databasePolicyName: mysql-db
  replicationModeDetector:
    image: PRIVATE_REGISTRY/replication-mode-detector:v0.13.0
  stash:
    addon:
      backupTask:
        name: mysql-backup-8.0.21
      restoreTask:
        name: mysql-restore-8.0.21
  updateConstraints:
    denylist:
      groupReplication:
      - < 8.0.29
      standalone:
      - < 8.0.29
  version: 8.0.29
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

## Deploy MySQL database from Private Registry

While deploying `MySQL` from private repository, you have to add `myregistrykey` secret in `MySQL` `spec.imagePullSecrets`.
Below is the MySQL CRD object we will create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-pvt-reg
  namespace: demo
spec:
  version: "8.0.29"
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

Now run the command to deploy this `MySQL` object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/private-registry/yamls/standalone.yaml
mysql.kubedb.com/mysql-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `MySQL` is in running state:

```bash
$ kubectl get pods -n demo
NAME              READY     STATUS    RESTARTS   AGE
mysql-pvt-reg-0   1/1       Running   0          56s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mysql/mysql-pvt-reg -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mysql/mysql-pvt-reg

kubectl patch -n demo drmn/mysql-pvt-reg -p '{"spec":{"wipeOut":true}}' --type="merge"
kubectl delete -n demo drmn/mysql-pvt-reg

kubectl delete ns demo
```

## Next Steps

- Initialize [MySQL with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
