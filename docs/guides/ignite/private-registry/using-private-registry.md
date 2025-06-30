---
title: Run Ignite using Private Registry
menu:
  docs_{{ .version }}:
    identifier: ig-using-private-registry-private-registry
    name: Quickstart
    parent: ig-private-registry-ignite
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run Ignite server using private Docker images.

## Before You Begin

- Read [concept of Ignite Version Catalog](/docs/guides/ignite/concepts/ignite-version.md) to learn detail concepts of `IgniteVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For ignite, push `DB_IMAGE`, `EXPORTER_IMAGE` of following IgniteVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get igniteversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image,DEPRECATED:.spec.deprecated
  NAME     VERSION   DB_IMAGE                                          EXPORTER_IMAGE                                          DEPRECATED
  2.17.0   2.17.0    ghcr.io/appscode-images/ignite:2.17.0             ghcr.io/kubedb/ignite-init:2.17.0-v1                    <none>
  ```

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: IgniteVersion
  metadata:
    name: 2.17.0
  spec:
    db:
      image: PRIVATE_REGISTRY/ignite:2.17.0
    securityContext:
      runAsUser: 70
    version: 2.17.0
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

## Deploy Ignite server from Private Registry

While deploying `Ignite` from private repository, you have to add `myregistrykey` secret in `Ignite` `spec.imagePullSecrets`.
Below is the Ignite CRD object we will create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: ig-pvt-reg
  namespace: demo
spec:
  replicas: 3
  version: "2.17.0"
  podTemplate:
    spec:
      containers:
      - name: ignite
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 250m
            memory: 64Mi
      imagePullSecrets:
      - name: myregistrykey
```

Now run the command to deploy this `Ignite` object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/private-registry/demo-2.yaml
ignite.kubedb.com/ig-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `Ignite` is in running state:

```bash
$ kubectl get pods -n demo -w
NAME                             READY     STATUS              RESTARTS   AGE
ig-pvt-reg-694d4d44df-bwtk8      0/1       ContainerCreating   0          18s
ig-pvt-reg-694d4d44df-tkqc4      0/1       ContainerCreating   0          17s
ig-pvt-reg-694d4d44df-zhj4l      0/1       ContainerCreating   0          17s
ig-pvt-reg-694d4d44df-bwtk8      1/1       Running             0          25s
ig-pvt-reg-694d4d44df-zhj4l      1/1       Running             0          26s
ig-pvt-reg-694d4d44df-tkqc4      1/1       Running             0          27s

$ kubectl get ig -n demo
NAME            VERSION    STATUS    AGE
ig-pvt-reg      2.17.0     Running   59s
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo ig/ig-pvt-reg -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
ignite.kubedb.com/ig-pvt-reg patched

$ kubectl delete -n demo ig/ig-pvt-reg
ignite.kubedb.com "ig-pvt-reg" deleted

$ kubectl delete -n demo secret myregistrykey
secret "myregistrykey" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Monitor your Ignite server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ignite/monitoring/using-prometheus-operator.md).
- Monitor your Ignite server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ignite/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Ignite object](/docs/guides/ignite/concepts/ignite.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
