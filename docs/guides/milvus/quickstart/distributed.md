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

This tutorial will show you how to use KubeDB to provision a **Distributed** [Milvus](https://milvus.io) cluster, where the Milvus roles run as separate, independently scalable workloads.

## Before You Begin

- You need a Kubernetes cluster with the KubeDB operator installed (`--set global.featureGates.Milvus=true`).

- Milvus external dependencies must be available:
  - **Object storage** (`my-release-minio` secret) — mandatory.
  - An **etcd operator** — KubeDB provisions an internal etcd cluster for metadata when `spec.metaStorage` is omitted.

- Create the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Distributed Topology

A distributed Milvus is composed of five roles:

| Role | Purpose | Persistent storage |
| --- | --- | --- |
| `mixcoord` | Coordinator (root/data/query/index coordination) | No |
| `proxy` | Client-facing gateway (gRPC `19530`) | No |
| `datanode` | Data persistence | No |
| `querynode` | Query execution | No (scratch only) |
| `streamingnode` | Streaming / WAL | **Yes** |

**Only `streamingnode` carries a persistent volume.** This is why distributed storage operations (volume expansion, storage migration, storage autoscaling) target `streamingnode`.

## Create a Distributed Milvus

In the manifest below, **only `streamingnode` is specified** under `spec.topology.distributed`. The operator **defaults the other four roles** (`mixcoord`, `datanode`, `querynode`, `proxy`) automatically.

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
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  tls:
    issuerRef:
      name: milvus-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    external:
      mode: mTLS
    internal:
      mode: TLS
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/distributed.yaml
milvus.kubedb.com/milvus-cluster created
```

## Wait for the Cluster to be Ready

Distributed Milvus takes longer than standalone — allow time for all components and the internal etcd to settle.

```bash
$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w
NAME             VERSION   STATUS         AGE
milvus-cluster   2.6.11    Provisioning   20s
milvus-cluster   2.6.11    Ready          3m
```

## Verify the Created Resources

### PetSets — one per role

```bash
$ kubectl get petset -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                           AGE
milvus-cluster-datanode        2m57s
milvus-cluster-mixcoord        2m58s
milvus-cluster-proxy           2m54s
milvus-cluster-querynode       2m56s
milvus-cluster-streamingnode   2m55s
```

The four roles other than `streamingnode` were created even though only `streamingnode` was specified — they are the defaulted distributed components.

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

A primary client service (`milvus-cluster`, gRPC `19530`, backed by the proxy), a metrics stats service (`milvus-cluster-stats`), and a headless governing service per role (`9091`) are created:

```bash
$ kubectl get svc -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
milvus-cluster                 ClusterIP   10.43.221.1   <none>        19530/TCP   3m
milvus-cluster-datanode        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-mixcoord        ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-querynode       ClusterIP   None          <none>        9091/TCP    3m
milvus-cluster-stats           ClusterIP   10.43.95.57   <none>        9091/TCP    3m
milvus-cluster-streamingnode   ClusterIP   None          <none>        9091/TCP    3m
```

### Storage — only on streamingnode

There is exactly one Milvus PVC, for the `streamingnode`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-cluster
NAME                                  STATUS   VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-milvus-cluster-streamingnode-0   Bound    pvc-...   1Gi        RWO            local-path     2m55s
```

### Auth, TLS, and AppBinding

As with standalone, KubeDB auto-generates the auth secret (`milvus-cluster-auth`, user `root`), the TLS certificate secrets, the rendered configuration secret, and an `AppBinding`. Because TLS is enabled, the AppBinding scheme is `https`:

```bash
$ kubectl get secret -n demo | grep milvus-cluster
milvus-cluster-auth          kubernetes.io/basic-auth   2      3m
milvus-cluster-client-cert   kubernetes.io/tls          4      3m
milvus-cluster-d7497a        Opaque                     2      3m
milvus-cluster-server-cert   kubernetes.io/tls          3      3m

$ kubectl get appbinding milvus-cluster -n demo -o jsonpath='{.spec.clientConfig.service}'
{"name":"milvus-cluster","path":"/","port":19530,"scheme":"https"}
```

> These base manifests already include **Prometheus Operator monitoring** and **TLS** — see the [monitoring](/docs/guides/milvus/monitoring/using-prometheus-operator.md) and [TLS](/docs/guides/milvus/tls/guide.md) guides.

## Cleaning up

```bash
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com milvus-cluster -n demo
$ kubectl delete ns demo
```

## Next Steps

- [Monitor](/docs/guides/milvus/monitoring/using-prometheus-operator.md) your Milvus cluster.
- [Horizontally scale](/docs/guides/milvus/scaling/horizontal-scaling/guide.md) the distributed roles.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
