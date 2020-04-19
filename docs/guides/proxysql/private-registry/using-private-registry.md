---
title: Run ProxySQL using Private Registry
menu:
  docs_{{ .version }}:
    identifier: proxysql-using-private-registry
    name: Run ProxySQL using Private Registry
    parent: proxysql-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Deploy ProxySQL from private Docker registry

KubeDB operator supports using a private Docker registry. This tutorial will show you how to use KubeDB to run ProxySQL using private Docker images.

## Before You Begin

- Read [concept of ProxySQLVersion Catalog](/docs/concepts/catalog/proxysql.md) to learn detail concepts of `ProxySQLVersion` object.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a docker private [registry](https://docs.docker.com/registry/) or [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories). In this tutorial, we will use a private repository of [docker hub](https://hub.docker.com/).

- You have to push the required images from KubeDB's [Dockerhub account](https://hub.docker.com/r/kubedb/) into your private registry. For proxysql, push `PROXYSQL_IMAGE`, `EXPORTER_IMAGE` of following ProxySQLVersion, where `deprecated` is not true, to your private registry. Currently, KubeDB includes the following ProxySQLVersion object.

  ```console
  $ kubectl get proxysqlversions  -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,PROXYSQL_IMAGE:.spec.proxysql.image,EXPORTER_IMAGE:.spec.exporter.image,DEPRECATED:.spec.deprecated
  NAME     VERSION   PROXYSQL_IMAGE          EXPORTER_IMAGE                   DEPRECATED
  2.0.4    2.0.4     kubedb/proxysql:v2.0.4   kubedb/proxysql-exporter:v1.1.0   <none>
  ```

  Docker hub repositories:

  - [kubedb/operator](https://hub.docker.com/r/kubedb/operator)
  - [kubedb/proxysql](https://hub.docker.com/r/kubedb/proxysql)
  - [kubedb/proxysql-exporter](https://hub.docker.com/r/kubedb/proxysql-exporter)

- Update KubeDB catalog for private Docker registry. Ex:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: ProxySQLVersion
  metadata:
    name: "2.0.4"
    labels:
      app: kubedb
  spec:
    version: "2.0.4"
    db:
      image: "PRIVATE_DOCKER_REGISTRY/proxysql:v2.0.4"
    exporter:
      image: "PRIVATE_DOCKER_REGISTRY/proxysql-exporter:v1.1.0"
  ...
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```console
  $ kubectl create ns demo
  namespace/demo created
  ```

## Create ImagePullSecret

ImagePullSecrets is a type of a Kubernetes Secret whose sole purpose is to pull private images from a Docker registry. It allows you to specify the url of the docker registry, credentials for logging in and the image name of your private docker image.

Run the following command, substituting the appropriate uppercase values to create an image pull secret for your private Docker registry:

```console
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images see [official docs](https://kubernetes.io/docs/concepts/containers/images/) of kubernetes.

> NB: If you are using `kubectl` 1.9.0, update to 1.9.1 or later to avoid this [issue](https://github.com/kubernetes/kubernetes/issues/57427).

## Install KubeDB operator

When installing KubeDB operator, set the flags `--docker-registry` and `--image-pull-secret` to the appropriate value. Follow the steps to [install KubeDB operator](/docs/setup/install.md) properly in the cluster so that to points to the DOCKER_REGISTRY you wish to pull images from.

## Deploy ProxySQL from Private Registry

While deploying `ProxySQL` from private repository, you have to add `myregistrykey` secret in `ProxySQL` `.spec.imagePullSecrets`.
Below is the ProxySQL object we will create.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: ProxySQL
metadata:
  name: proxysql-pvt-reg
  namespace: demo
spec:
  version: "2.0.4"
  replicas: 1
  mode: GroupReplication
  backend:
    ref:
      apiGroup: "kubedb.com"
      kind: MySQL
      name: my-group
    replicas: 3
  updateStrategy:
    type: RollingUpdate
  podTemplate:
    spec:
      imagePullSecrets:
      - name: myregistrykey
```

Now run the command to deploy this `ProxySQL` object:

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/proxysql/private-registry.yaml
proxysql.kubedb.com/proxysql-pvt-reg created
```

To check if the images pulled successfully from the repository, see if the `ProxySQL` is in running state:

```console
$ kubectl get pods -n demo
NAME                 READY     STATUS    RESTARTS   AGE
proxysql-pvt-reg-0   1/1       Running   0          56s
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo proxysql/proxysql-pvt-reg
kubectl delete ns demo
```

## Next Steps

- Monitor your ProxySQL with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/proxysql/monitoring/using-builtin-prometheus.md).
- Monitor your ProxySQL with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/proxysql/monitoring/using-coreos-prometheus-operator.md).
- Use custom config file to configure ProxySQL [here](/docs/guides/proxysql/configuration/using-custom-config.md).
- Detail concepts of ProxySQL CRD [here](/docs/concepts/database-proxy/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/concepts/catalog/proxysql.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
