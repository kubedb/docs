---
title: Milvus Distributed Quickstart
menu:
  docs_{{ .version }}:
    identifier: milvus-quickstart-distributed
    name: Distributed
    parent: milvus-quickstart
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Milvus Distributed Cluster

This tutorial shows how to use KubeDB to provision a **distributed** [Milvus](https://milvus.io) cluster, where the Milvus roles run as separate workloads.

## Before You Begin

- You need a Kubernetes cluster and `kubectl` configured to talk to it.
- Install KubeDB with `--set global.featureGates.Milvus=true`.
- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). That guide installs MinIO, creates the `my-release-minio` secret, and installs the etcd operator required by Milvus.
- This quickstart intentionally uses the smallest working distributed manifest. It does **not** require Prometheus Operator or cert-manager.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Distributed Topology

A distributed Milvus is composed of five roles:

| Role | Purpose | Persistent storage |
| --- | --- | --- |
| `mixcoord` | Coordinator | No |
| `proxy` | Client-facing gateway | No |
| `datanode` | Data persistence | No |
| `querynode` | Query execution | No |
| `streamingnode` | Streaming / WAL | Yes |

Only `streamingnode` carries a persistent volume. This is why distributed storage operations target `streamingnode`.

## Create a Distributed Milvus

In the manifest below, only `streamingnode` is specified under `spec.topology.distributed`. The operator defaults the other four roles automatically.

`distributed.yaml`

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-cluster
  namespace: demo
spec:
  version: "2.6.11"
  objectStorage:
    configSecret:
      name: "my-release-minio"
  topology:
    mode: Distributed
    distributed:
      streamingnode:
        storageType: Durable
        storage:
          accessModes:
            - ReadWriteOnce
          storageClassName: local-path
          resources:
            requests:
              storage: 1Gi
```

This manifest also omits `spec.metaStorage`, so KubeDB creates and manages the etcd metadata cluster through the installed etcd operator.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created
```

## Wait for the Cluster to be Ready

Distributed Milvus takes longer than standalone because multiple components have to settle:

```bash
$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w
NAME             VERSION   STATUS         AGE
milvus-cluster   2.6.11    Provisioning   20s
milvus-cluster   2.6.11    Ready          3m
```

## Verify the Created Resources

### PetSets

```bash
$ kubectl get petset -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                           AGE
milvus-cluster-datanode        2m57s
milvus-cluster-mixcoord        2m58s
milvus-cluster-proxy           2m54s
milvus-cluster-querynode       2m56s
milvus-cluster-streamingnode   2m55s
```

The four roles other than `streamingnode` were created even though only `streamingnode` was specified.

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                             READY   STATUS    RESTARTS   AGE
milvus-cluster-datanode-0        1/1     Running   0          2m58s
milvus-cluster-mixcoord-0        1/1     Running   0          2m59s
milvus-cluster-proxy-0           1/1     Running   0          2m55s
milvus-cluster-querynode-0       1/1     Running   0          2m57s
milvus-cluster-streamingnode-0   1/1     Running   0          2m55s
```

### Services

A primary client service (`milvus-cluster`, gRPC `19530`, backed by the proxy) and a headless governing service per role are created:

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
milvus-cluster                 ClusterIP   10.43.221.1   <none>        19530/TCP   3m
milvus-cluster-datanode        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-mixcoord        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-querynode       ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-streamingnode   ClusterIP   None          <none>        9091/TCP    3m
```

If you later enable [Prometheus Operator monitoring](/docs/guides/milvus/monitoring/using-prometheus-operator.md), KubeDB also creates a `milvus-cluster-stats` service and a `ServiceMonitor`.

### Storage

There is exactly one Milvus PVC, for the `streamingnode`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                                  STATUS   VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-milvus-cluster-streamingnode-0   Bound    pvc-...   1Gi        RWO            local-path     2m55s
```

### Auth Secret and AppBinding

As with standalone, KubeDB auto-generates the auth secret (`milvus-cluster-auth`, user `root`), the rendered configuration secret, and an `AppBinding`. Because this quickstart does not enable TLS, the AppBinding scheme is `http`:

```bash
$ kubectl get secret -n demo | grep milvus-cluster
milvus-cluster-auth    kubernetes.io/basic-auth   2   3m
milvus-cluster-d7497a  Opaque                     2   3m

$ kubectl get appbinding milvus-cluster -n demo -o jsonpath='{.spec.clientConfig.service}'
{"name":"milvus-cluster","path":"/","port":19530,"scheme":"http"}
```

## Use an External etcd Cluster Instead

The default quickstart omits `spec.metaStorage`, so KubeDB manages etcd for you. If you want Milvus to use an external `EtcdCluster` instead:

1. Install the etcd operator.
2. Create an external `EtcdCluster` by following [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#optional-path-create-an-external-etcdcluster).
3. Add `spec.metaStorage.externallyManaged: true` and the external endpoints to your `Milvus` manifest.

## Cleaning up

```bash
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com milvus-cluster -n demo
```

If you want to remove the dependencies too, either delete the whole `demo` namespace or follow the cleanup steps in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#cleanup).

## Next Steps

- [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md) for another cluster.
- [Monitor](/docs/guides/milvus/monitoring/using-prometheus-operator.md) your Milvus cluster.
- [Horizontally scale](/docs/guides/milvus/scaling/horizontal-scaling/guide.md) the distributed roles.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
