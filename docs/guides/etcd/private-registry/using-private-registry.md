---
title: Run Etcd using Private Registry
menu:
  docs_{{ .version }}:
    identifier: etcd-private-registry-quickstart
    name: Quickstart
    parent: etcd-private-registry
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Using a private Docker registry

The KubeDB operator supports using a private Docker registry. This tutorial shows you how to use
KubeDB to run an `Etcd` cluster from private Docker images.

## Before You Begin

- Read the [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) concept to learn the details of
  the `EtcdVersion` catalog object.

- You need a Kubernetes cluster with `kubectl` configured to talk to it. If you do not already have a
  cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- You will also need a private Docker [registry](https://docs.docker.com/registry/) or
  [private repository](https://docs.docker.com/docker-hub/repos/#private-repositories).

- You have to push the required images into your private registry. For etcd, push the `DB_IMAGE` of
  every `EtcdVersion` where `deprecated` is not `true`:

  ```bash
  $ kubectl get etcdversions -o=custom-columns=NAME:.metadata.name,VERSION:.spec.version,DB_IMAGE:.spec.db.image,DEPRECATED:.spec.deprecated
  NAME     VERSION   DB_IMAGE                                  DEPRECATED
  3.5.21   3.5.21    ghcr.io/appscode-images/etcd:v3.5.21      <none>
  3.6.4    3.6.4     ghcr.io/appscode-images/etcd:v3.6.4       <none>
  ```

  `EtcdVersion` also carries a `spec.exporter.image`. For etcd it is set to the same etcd image,
  because **KubeDB does not deploy a separate metrics exporter sidecar** — etcd serves its own
  Prometheus metrics natively on port `2381`. There is no extra image to mirror for monitoring.

- Update the KubeDB catalog to point at your private Docker registry. For example:

  ```yaml
  apiVersion: catalog.kubedb.com/v1alpha1
  kind: EtcdVersion
  metadata:
    name: 3.6.4
  spec:
    version: 3.6.4
    db:
      image: PRIVATE_REGISTRY/etcd:v3.6.4
    exporter:
      image: PRIVATE_REGISTRY/etcd:v3.6.4
    securityContext:
      runAsUser: 1000
    updateConstraints:
      allowlist:
        - '>= 3.6.4, <= 3.7.0'
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo`:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: the YAML files used in this tutorial are stored in
> [docs/examples/etcd/private-registry](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd/private-registry)
> of the [kubedb/docs](https://github.com/kubedb/docs) repository.

## Create ImagePullSecret

An `imagePullSecret` is a Kubernetes `Secret` whose sole purpose is to pull private images from a
Docker registry. It holds the URL of the registry and the credentials to log into it.

Run the following command, substituting the appropriate uppercase values, to create an image pull
secret for your private Docker registry:

```bash
$ kubectl create secret docker-registry -n demo myregistrykey \
  --docker-server=DOCKER_REGISTRY_SERVER \
  --docker-username=DOCKER_USER \
  --docker-email=DOCKER_EMAIL \
  --docker-password=DOCKER_PASSWORD
secret/myregistrykey created
```

If you wish to follow other ways to pull private images, see the
[official Kubernetes docs](https://kubernetes.io/docs/concepts/containers/images/).

## Install the KubeDB operator

When installing the KubeDB operator, set the `--docker-registry` and `--image-pull-secret` flags to
the appropriate values so that the operator's own images also come from your registry. Follow the
steps to [install the KubeDB operator](/docs/setup/README.md).

etcd support is an alpha feature gate, so it must be enabled explicitly as well — pass
`--set featureGates.Etcd=true` to the KubeDB Helm chart.

## Deploy Etcd from the private registry

While deploying an `Etcd` cluster from a private repository, add the `myregistrykey` secret to
`spec.podTemplate.spec.imagePullSecrets`. Below is the `Etcd` object we will create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-pvt-reg
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      imagePullSecrets:
        - name: myregistrykey
      containers:
        - name: etcd
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 250m
              memory: 512Mi
  deletionPolicy: WipeOut
```

Now run the command to deploy this `Etcd` object:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/private-registry/etcd-pvt-reg.yaml
etcd.kubedb.com/etcd-pvt-reg created
```

To check whether the images were pulled successfully, watch the pods come up. KubeDB brings the
members up one ordinal at a time, waiting for each to become Ready before starting the next:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-pvt-reg -w
NAME             READY   STATUS              RESTARTS   AGE
etcd-pvt-reg-0   0/1     ContainerCreating   0          18s
etcd-pvt-reg-0   1/1     Running             0          32s
etcd-pvt-reg-1   1/1     Running             0          58s
etcd-pvt-reg-2   1/1     Running             0          84s

$ kubectl get etcd -n demo etcd-pvt-reg
NAME           VERSION   STATUS   AGE
etcd-pvt-reg   3.6.4     Ready    2m5s
```

If the images could not be pulled, the pods stay in `ImagePullBackOff` and the reason is visible in
the pod events:

```bash
$ kubectl describe pod -n demo etcd-pvt-reg-0
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo etcd/etcd-pvt-reg -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
etcd.kubedb.com/etcd-pvt-reg patched

$ kubectl delete -n demo etcd/etcd-pvt-reg
etcd.kubedb.com "etcd-pvt-reg" deleted

$ kubectl delete -n demo secret myregistrykey
secret "myregistrykey" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Run etcd with [custom RBAC resources](/docs/guides/etcd/custom-rbac/using-custom-rbac.md).
- Back up and restore your etcd cluster with
  [KubeStash](/docs/guides/etcd/backup/kubestash/overview/index.md).
- Detail concepts of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.
- Detail concepts of the [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) object.
- Monitor your etcd cluster with KubeDB using the
  [out-of-the-box Prometheus operator](/docs/guides/etcd/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
