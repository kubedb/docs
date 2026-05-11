---
title: Run DB2 using Private Registry
menu:
  docs_{{ .version }}:
    identifier: db2-using-private-registry
    name: Using Private Registry
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

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/db2](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/db2) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Private Docker Registry

You need to have a private Docker registry to host the DB2 images. You can use:

- A private repository in [Docker Hub](https://docs.docker.com/docker-hub/repos/#private-repositories)
- A self-hosted [Docker Registry](https://docs.docker.com/registry/)
- A cloud provider's container registry (AWS ECR, Google GCR, etc.)

You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For DB2, push the `DB_IMAGE` and `COORDINATOR_IMAGE` of the following DB2Versions (where `deprecated` is not true) to your private registry.

```bash
$ kubectl get db2versions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,COORDINATOR_IMAGE:.spec.coordinator.image,DEPRECATED:.spec.deprecated
NAME         VERSION   DB_IMAGE                 COORDINATOR_IMAGE                          DEPRECATED
11.5.8.0       11.5.8.0    kubedb/db2:11.5.8.0        ghcr.io/kubedb/db2-coordinator:v0.5.0-ubi  <none>
```

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

Verify the image pull secret is created:

```bash
$ kubectl get secret -n demo myregistrykey
NAME             TYPE                             DATA   AGE
myregistrykey    kubernetes.io/dockercfg         1      5m
```

## Install KubeDB Operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate values. Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that it points to the DOCKER_REGISTRY you wish to pull images from.

## Create DB2Version CRD

KubeDB uses images specified in DB2Version CRD for the database container and the coordinator container for health checking. You have to create a DB2Version CRD specifying images from your private registry. Then, you have to point this DB2Version CRD in `spec.version` field of DB2 object. For more details about DB2Version CRD, please visit [here](/docs/guides/db2/concepts/catalog.md).

Here is an example of DB2Version CRD. Replace `PRIVATE_REGISTRY` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: DB2Version
metadata:
  name: "11.5.8.0"
spec:
  version: "11.5.8.0"
  db:
    image: PRIVATE_REGISTRY/db2:11.5.9
  coordinator:
    image: PRIVATE_REGISTRY/db2-coordinator:v0.5.0-ubi
  deprecated: false
```

Now, create the DB2Version CRD:

```bash
$ kubectl apply -f pvt-db2version.yaml
db2version.catalog.kubedb.com/11.5.9 created
```

Verify the DB2Version is created:

```bash
$ kubectl get db2version
NAME        VERSION   DB_IMAGE                           COORDINATOR_IMAGE                               DEPRECATED
11.5.8.0      11.5.8.0    PRIVATE_REGISTRY/db2:11.5.8.0        PRIVATE_REGISTRY/db2-coordinator:v0.5.0-ubi     false
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
  version: "11.5.8.0"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
  deletionPolicy: Delete
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

Wait for the DB2 instance to be ready:

```bash
$ kubectl get db2 -n demo pvt-reg-db2
NAME           VERSION   STATUS    AGE
pvt-reg-db2    11.5.8.0    Running   3m
```

Check the pod logs to verify the database is ready:

## Verify Image Pull

To verify that KubeDB has successfully pulled the private images from your registry, you can inspect the pod events:

```bash
$ kubectl describe pod -n demo pvt-reg-db2-0
```

Look for events like:
```
Successfully pulled image "PRIVATE_REGISTRY/db2:11.5.9" in XXms
Successfully pulled image "PRIVATE_REGISTRY/db2-coordinator:v0.5.0-ubi" in XXms
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo db2/pvt-reg-db2 -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo db2/pvt-reg-db2

$ kubectl delete -n demo db2version 11.5.8.0

$ kubectl delete secret -n demo myregistrykey

$ kubectl delete ns demo
```

## Next Steps

- Learn how to use KubeDB to run DB2 [here](/docs/guides/db2/README.md).
- Learn about [custom RBAC](/docs/guides/db2/custom-rbac/using-custom-rbac.md) for DB2.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

