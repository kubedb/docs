---
title: Run Weaviate using Private Registry
menu:
  docs_{{ .version }}:
    identifier: weaviate-using-private-registry
    name: Quickstart
    parent: weaviate-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Private Docker Registry

KubeDB supports using private Docker registry. This tutorial will show you how to run KubeDB managed Weaviate database using private Docker images.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/private-registry](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/private-registry) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Private Docker Registry

You will need a Docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories). In this tutorial we will use a private repository on [Docker Hub](https://hub.docker.com/).

You need to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For Weaviate, push `DB_IMAGE` of the following `WeaviateVersion`s, where `deprecated` is not true, to your private registry.

```bash
$ kubectl get weaviateversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,DEPRECATED:.spec.deprecated
NAME      VERSION   DB_IMAGE                        DEPRECATED
1.25.0    1.25.0    kubedb/weaviate:1.25.0          <none>
1.28.0    1.28.0    kubedb/weaviate:1.28.0          <none>
1.30.0    1.30.0    kubedb/weaviate:1.30.0          <none>
1.33.1    1.33.1    kubedb/weaviate:1.33.1          <none>
```

Docker hub repository:
- [kubedb/weaviate](https://hub.docker.com/r/kubedb/weaviate)

## Create ImagePullSecret

`ImagePullSecrets` is a type of Kubernetes Secret whose purpose is to pull private images from a Docker registry. It allows you to specify the URL of the Docker registry, credentials for logging in, and the image name of your private Docker image.

Run the following command, substituting the appropriate uppercase values, to create an image pull secret for your private Docker registry:

```bash
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of Kubernetes.

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to the appropriate values. Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in your cluster so that it points to the `DOCKER_REGISTRY` you wish to pull images from.

## Create WeaviateVersion CRD

KubeDB uses images specified in `WeaviateVersion` CRD for the database. You have to create a `WeaviateVersion` CRD specifying images from your private registry. Then, you have to point this `WeaviateVersion` CRD in `spec.version` field of the `Weaviate` object. For more details about `WeaviateVersion` CRD, please visit [here](/docs/guides/weaviate/concepts/catalog.md).

Here is an example of a `WeaviateVersion` CRD. Replace `PRIVATE_REGISTRY` with your private registry:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: WeaviateVersion
metadata:
  name: "1.33.1-private"
spec:
  db:
    image: PRIVATE_REGISTRY/weaviate:1.33.1
  version: "1.33.1"
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/private-registry/weaviateversion.yaml
weaviateversion.catalog.kubedb.com/1.33.1-private created
```

## Deploy Weaviate from Private Registry

While deploying `Weaviate` from private registry, you have to add `myregistrykey` secret in `spec.podTemplate.spec.imagePullSecrets` and specify `1.33.1-private` in `spec.version` field.

Below is the YAML for Weaviate crd we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: pvt-reg-weaviate
  namespace: demo
spec:
  version: "1.33.1-private"
  replicas: 3
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

Now run the command to create this Weaviate object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/private-registry/pvt-reg-weaviate.yaml
weaviate.kubedb.com/pvt-reg-weaviate created
```

To check if the images pulled successfully from the registry, wait for the Weaviate to go into `Ready` state:

```bash
$ kubectl get weaviate -n demo pvt-reg-weaviate -w
NAME               VERSION          STATUS         AGE
pvt-reg-weaviate   1.33.1-private   Provisioning   5s
pvt-reg-weaviate   1.33.1-private   Ready          1m
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo pvt-reg-weaviate
kubectl delete ns demo
```