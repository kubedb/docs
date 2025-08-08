---
title: Run Redis using Private Registry
menu:
  docs_{{ .version }}:
    identifier: rd-using-private-registry-private-registry
    name: Quickstart
    parent: rd-private-registry-redis
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using private Docker registry

KubeDB operator supports using private Docker registry. This tutorial will show you how to use KubeDB to run Redis server using private Docker images.

## Before You Begin

- Read [concept of Redis Version Catalog](/docs/guides/redis/concepts/catalog.md) to learn detail concepts of `RedisVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).  In this tutorial we will use private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Docker hub account](https://hub.docker.com/r/kubedb/) into your private registry. For redis, push `DB_IMAGE`, `TOOLS_IMAGE`, `EXPORTER_IMAGE` of following RedisVersions, where `deprecated` is not true, to your private registry.

```bash
$ kubectl get redisversions -n kube-system  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,INITCONTAINER_IMAGE:.spec.initContainer.image,DB_IMAGE:.spec.db.image,EXPORTER_IMAGE:.spec.exporter.image
NAME           VERSION   INITCONTAINER_IMAGE                DB_IMAGE                                        EXPORTER_IMAGE
4.0.11         4.0.11    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/kubedb/redis:4.0.11                     ghcr.io/kubedb/redis_exporter:1.66.0
5.0.14         5.0.14    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:5.0.14-bullseye   ghcr.io/kubedb/redis_exporter:1.66.0
6.0.20         6.0.20    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:6.0.20-bookworm   ghcr.io/kubedb/redis_exporter:1.66.0
6.2.14         6.2.14    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:6.2.14-bookworm   ghcr.io/kubedb/redis_exporter:1.66.0
6.2.16         6.2.16    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:6.2.16-bookworm   ghcr.io/kubedb/redis_exporter:1.66.0
7.0.14         7.0.14    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.0.14-bookworm   ghcr.io/kubedb/redis_exporter:1.66.0
7.0.15         7.0.15    ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.0.15-bookworm   ghcr.io/kubedb/redis_exporter:1.66.0
7.2.3          7.2.3     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.2.3-bookworm    ghcr.io/kubedb/redis_exporter:1.66.0
7.2.4          7.2.4     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.2.4-bookworm    ghcr.io/kubedb/redis_exporter:1.66.0
7.2.6          7.2.6     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.2.6-bookworm    ghcr.io/kubedb/redis_exporter:1.66.0
7.4.0          7.4.0     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.4.0-bookworm    ghcr.io/kubedb/redis_exporter:1.66.0
7.4.1          7.4.1     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/redis:7.4.1-bookworm    ghcr.io/kubedb/redis_exporter:1.66.0
valkey-7.2.5   7.2.5     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/valkey:7.2.5            ghcr.io/kubedb/redis_exporter:1.66.0
valkey-7.2.9   7.2.9     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/valkey:7.2.9            ghcr.io/kubedb/redis_exporter:1.66.0
valkey-8.0.3   8.0.3     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/valkey:8.0.3            ghcr.io/kubedb/redis_exporter:1.66.0
valkey-8.1.1   8.1.1     ghcr.io/kubedb/redis-init:0.12.0   ghcr.io/appscode-images/valkey:8.1.1            ghcr.io/kubedb/redis_exporter:1.66.0
```

  Docker hub repositories:

  - [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
  - [kubedb/redis](https://hub.docker.com/r/kubedb/redis)
  - [kubedb/redis_exporter](https://hub.docker.com/r/kubedb/redis_exporter)

`Note`: While using Valkey as the DB image, take initContainer version greater than or equal to 0.10.0

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: RedisVersion
  metadata:
    name: 6.2.14
  spec:
    db:
      image: PRIVATE_DOCKER_REGISTRY:6.0.20
    exporter:
      image: PRIVATE_DOCKER_REGISTRY:1.9.0
    initContainer:
      image: PRIVATE_DOCKER_REGISTRY:0.12.0
    podSecurityPolicies:
      databasePolicyName: redis-db
    version: 6.0.20
  ```

## Create ImagePullSecret

ImagePullSecrets is a type of Kubernetes Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

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

## Deploy Redis server from Private Registry

While deploying `Redis` from private repository, you have to add `myregistrykey` secret in `Redis` `spec.imagePullSecrets`.
Below is the Redis CRD object we will create.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-pvt-reg
  namespace: demo
spec:
  version: 6.2.14
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

Now run the command to deploy this `Redis` object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/private-registry/demo-2.yaml
redis.kubedb.com/redis-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `Redis` is in running state:

```bash
$ kubectl get pods -n demo -w
NAME              READY     STATUS              RESTARTS   AGE
redis-pvt-reg-0   0/1       Pending             0          0s
redis-pvt-reg-0   0/1       Pending             0          0s
redis-pvt-reg-0   0/1       ContainerCreating   0          0s
redis-pvt-reg-0   1/1       Running             0          2m


$ kubectl get rd -n demo
NAME            VERSION   STATUS    AGE
redis-pvt-reg   6.2.14    Running   40s
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo rd/redis-pvt-reg -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo rd/redis-pvt-reg

kubectl patch -n demo drmn/redis-pvt-reg -p '{"spec":{"wipeOut":true}}' --type="merge"
kubectl delete -n demo drmn/redis-pvt-reg

kubectl delete ns demo
```

```bash
$ kubectl patch -n demo rd/redis-pvt-reg -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/redis-pvt-reg patched

$ kubectl delete -n demo rd/redis-pvt-reg
redis.kubedb.com "redis-pvt-reg" deleted

$ kubectl delete -n demo secret myregistrykey
secret "myregistrykey" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Monitor your Redis server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
