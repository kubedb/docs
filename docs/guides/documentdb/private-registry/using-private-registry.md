---
title: Run DocumentDB using Private Registry
menu:
  docs_{{ .version }}:
    identifier: documentdb-using-private-registry
    name: Quickstart
    parent: documentdb-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

KubeDB supports using private Docker registries. This tutorial will show you how to run KubeDB managed DocumentDB database using private Docker images.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Private Docker Registry

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories). In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For DocumentDB, push the `DB_IMAGE` of the following DocumentDBVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get documentdbversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,DEPRECATED:.spec.deprecated
  NAME           VERSION          DB_IMAGE                       DEPRECATED
  pg17-0.109.0     pg17-0.109.0    kubedb/documentdb:5.0.6        <none>
  ```

## Create ImagePullSecret

Run the following command to create an image pull secret for your private Docker registry:

```bash
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

## Install KubeDB Operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate values.
Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that it points to the DOCKER_REGISTRY you wish to pull images from.

## Create DocumentDBVersion CRD

KubeDB uses images specified in DocumentDBVersion CRD for database. You have to create a DocumentDBVersion CRD specifying images from your private registry. Then, point this DocumentDBVersion CRD in `spec.version` field of DocumentDB object. For more details about DocumentDBVersion CRD, please visit [here](/docs/guides/documentdb/concepts/catalog.md).

Here, is an example of DocumentDBVersion CRD. Replace `PRIVATE_REGISTRY` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DocumentDBVersion
metadata:
  name: "pg17-0.109.0"
spec:
  db:
    image: PRIVATE_REGISTRY/documentdb:5.0.6
  version: "pg17-0.109.0"
```

Now, create the DocumentDBVersion CRD:

```bash
$ kubectl apply -f pvt-documentdbversion.yaml
documentdbversion.catalog.kubedb.com/pg17-0.109.0 created
```

## Deploy DocumentDB from Private Registry

While deploying DocumentDB from private repository, you have to add `myregistrykey` secret in DocumentDB `spec.podTemplate.spec.imagePullSecrets` and specify your private version in `spec.version` field.

Below is the DocumentDB object we will create in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: pvt-reg-docdb
  namespace: demo
spec:
  version: "pg17-0.109.0"
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

Now run the command to create this DocumentDB object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/private-registry/pvt-reg-docdb.yaml
documentdb.kubedb.com/pvt-reg-docdb created
```

To check if the images pulled successfully from the repository, see if the DocumentDB is in Running state:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=pvt-reg-docdb"
NAME               READY   STATUS    RESTARTS   AGE
pvt-reg-docdb-0    1/1     Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo documentdb/pvt-reg-docdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo documentdb/pvt-reg-docdb

kubectl delete ns demo
```
