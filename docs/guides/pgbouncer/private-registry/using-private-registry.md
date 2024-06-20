---
title: Run PgBouncer using Private Registry
menu:
  docs_{{ .version }}:
    identifier: pb-using-private-registry-private-registry
    name: Quickstart
    parent: pb-private-registry-pgbouncer
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using private Docker registry

KubeDB supports using private Docker registry. This tutorial will show you how to run KubeDB managed PgBouncer using private Docker images.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgbouncer](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgbouncer) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Private Docker Registry

- You will need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories). In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For pgbouncer, push `SERVER_IMAGE`, `EXPORTER_IMAGE` of following PgBouncerVersions, where `deprecated` is not true, to your private registry.

  ```bash
  $ kubectl get pgbouncerversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.server.image,EXPORTER_IMAGE:.spec.exporter.image,DEPRECATED:.spec.deprecated
  NAME     VERSION   SERVER_IMAGE              EXPORTER_IMAGE                     DEPRECATED
  1.17.0   1.17.0    kubedb/pgbouncer:1.17.0   kubedb/pgbouncer_exporter:v0.1.1   false
  ```

  Docker hub repositories:

- [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
- [kubedb/pgbouncer](https://hub.docker.com/r/kubedb/pgbouncer)
- [kubedb/pgbouncer_exporter](https://hub.docker.com/r/kubedb/pgbouncer_exporter)

## Create ImagePullSecret

ImagePullSecrets is a type of a Kubernetes Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

Run the following command, substituting the appropriate uppercase values to create an image pull secret for your private Docker registry:

```bash
$ kubectl create secret generic -n demo docker-registry myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of Kubernetes.

> Note; If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to appropriate value.
Follow the steps to [install KubeDB operator](/docs/setup/README.md) properly in cluster so that to points to the DOCKER_REGISTRY you wish to pull images from.

## Create PgBouncerVersion CRD

KubeDB uses images specified in PgBouncerVersion crd for pgbouncer server, and prometheus metrics exporter. You have to create a PgBouncerVersion crd specifying images from your private registry. Then, you have to point this PgBouncerVersion crd in `spec.version` field of Postgres object. For more details about PgBouncerVersion crd, please visit [here](/docs/guides/pgbouncer/concepts/catalog.md).

Here, is an example of PgBouncerVersion crd. Replace `<YOUR_PRIVATE_REGISTRY>` with your private registry.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PgBouncerVersion
metadata:
  name: "1.17.0"
spec:
  exporter:
    image: PRIVATE_REGISTRY/pgbouncer_exporter:v0.1.1
  pgBouncer:
    image: PRIVATE_REGISTRY/pgbouncer:1.17.0
  version: 1.17.0
```

Now, create the PgBouncerVersion crd,

```bash
$ kubectl apply -f pvt-pgbouncerversion.yaml
pgbouncerversion.kubedb.com/pvt-1.17.0 created
```

## Deploy PgBouncer from Private Registry

While deploying PgBouncer from private repository, you have to add `myregistrykey` secret in PgBouncer `spec.podTemplate.spec.imagePullSecrets` and specify `pvt-1.17.0` in `spec.version` field.

Below is the PgBouncer object we will create in this tutorial

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PgBouncer
metadata:
  name: pvt-reg-pgbouncer
  namespace: demo
spec:
  version: "1.17.0"
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
      namespace: demo
  connectionPool:
    maxClientConnections: 20
    reservePoolSize: 5
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
```

Now run the command to create this pgbouncer server:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgbouncer/private-registry/pvt-reg-pgbouncer.yaml
pgbouncer.kubedb.com/pvt-reg-pgbouncer created
```

To check if the images pulled successfully from the repository, see if the PgBouncer is in Running state:

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=pvt-reg-pgbouncer"
NAME                 READY     STATUS    RESTARTS   AGE
pvt-reg-pgbouncer-0   1/1       Running   0          3m
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pb/pvt-reg-pgbouncer

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Monitor your PgBouncer with KubeDB using [built-in Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Monitor your PgBouncer with KubeDB using [Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
