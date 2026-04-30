---
title: Run DB2 using Private Registry
menu:
  docs_{{ .version }}:
    identifier: db2-using-private-registry
    name: Quickstart
    parent: db2-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

KubeDB supports using private Docker registries. This tutorial will show you how to run KubeDB managed DB2 database using private Docker images.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/db2](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/db2) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Private Docker Registry

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories). In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For DB2, push `DB_IMAGE` of the following DB2Versions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get db2versions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,DEPRECATED:.spec.deprecated
  NAME         VERSION   DB_IMAGE                   DEPRECATED
  11.5.9       11.5.9    kubedb/db2:11.5.9          <none>
  ```

  Docker hub repositories:

- [kubedb/db2](https://hub.docker.com/r/kubedb/db2)

## Create ImagePullSecret

ImagePullSecrets is a type of a Kubernetes Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

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

## Install KubeDB Operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate values.
Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that it points to the DOCKER_REGISTRY you wish to pull images from.

## Create DB2Version CRD

KubeDB uses images specified in DB2Version CRD for database and exporting prometheus metrics. You have to create a DB2Version CRD specifying images from your private registry. Then, you have to point this DB2Version CRD in `spec.version` field of DB2 object. For more details about DB2Version CRD, please visit [here](/docs/guides/db2/concepts/catalog.md).

Here, is an example of DB2Version CRD. Replace `PRIVATE_REGISTRY` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DB2Version
metadata:
  name: "11.5.9"
spec:
  db:
    image: PRIVATE_REGISTRY/db2:11.5.9
  version: "11.5.9"
```

Now, create the DB2Version CRD:

```bash
$ kubectl apply -f pvt-db2version.yaml
db2version.catalog.kubedb.com/11.5.9 created
```

## Deploy DB2 from Private Registry

While deploying DB2 from private repository, you have to add `myregistrykey` secret in DB2 `spec.podTemplate.spec.imagePullSecrets` and specify your private version in `spec.version` field.

Below is the DB2 object we will create in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: pvt-reg-db2
  namespace: demo
spec:
  version: "11.5.9"
  storageType: Durable
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

Now run the command to create this DB2 object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/private-registry/pvt-reg-db2.yaml
db2.kubedb.com/pvt-reg-db2 created
```

To check if the images pulled successfully from the repository, see if the DB2 is in Running state:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=pvt-reg-db2"
NAME             READY     STATUS    RESTARTS   AGE
pvt-reg-db2-0    1/1       Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo db2/pvt-reg-db2 -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo db2/pvt-reg-db2

kubectl delete ns demo
```
