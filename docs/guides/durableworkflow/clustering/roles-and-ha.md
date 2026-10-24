---
title: Roles and High Availability
menu:
  docs_{{ .version }}:
    identifier: dw-roles-ha-clustering
    name: Roles and High Availability
    parent: dw-clustering-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Roles and High Availability

A DurableWorkflow cluster is a set of stateless roles. Each role is one PetSet, and `spec.topology` sets its replicas and pod template.

| Role               | Purpose                                                              | Engines  |
|--------------------|----------------------------------------------------------------------|----------|
| `frontend`         | The API gateway that clients, the UI and the CLI talk to              | both     |
| `history`          | Owns the history shards and runs the workflow state machines          | both     |
| `matching`         | Owns the task queues                                                  | both     |
| `worker`           | The system workers of the cluster itself                              | both     |
| `internalFrontend` | A frontend for the cluster's own system worker, needed for authorization | Temporal |

If you leave a role out of `spec.topology`, it runs with 1 replica. `internalFrontend` runs only when you set it.

## Ports

The ports come from the `DurableWorkflowVersion`, not from the `DurableWorkflow`.

| Role             | Temporal: gRPC / membership / HTTP | Cadence: TChannel / gRPC |
|------------------|------------------------------------|--------------------------|
| frontend         | 7233 / 6933 / 7243                 | 7933 / 7833              |
| internalFrontend | 7236 / 6936 / 7246                 | not available            |
| history          | 7234 / 6934 / none                 | 7934 / 7834              |
| matching         | 7235 / 6935 / none                 | 7935 / 7835              |
| worker           | 7239 / 6939 / none                 | 7939 / none              |
| metrics, all roles | 9090                             | 9090                     |
| web UI           | 8080                               | 8088                     |

Clients use the frontend Service (`<name>`), which exposes the gRPC port, and also the HTTP port for Temporal.

## History shards

`spec.numHistoryShards` is the number of history shards. The default is `16` and comes from the `DurableWorkflowVersion`. **It is immutable.** Choose it at creation time, because it is the ceiling of the scale of the cluster: a history pod without a shard has no work. The webhook warns, and does not reject, when `topology.history.replicas` exceeds `numHistoryShards`.

## High availability

A cluster is **highly available** when every enabled role has at least 2 replicas, and the UI has at least 2 if it is enabled. The webhook of the ops requests uses this definition to warn you before an operation that would cause downtime.

On a highly available cluster:

- every role with at least 2 replicas has a PodDisruptionBudget with `maxUnavailable: 1`, which also protects the cluster during node drains;
- the pods of a role are spread over the nodes;
- the PetSets use the `OnDelete` update strategy, so only a [DurableWorkflowOpsRequest](/docs/guides/durableworkflow/concepts/durableworkflow-ops-request/index.md) replaces pods, and one at a time;
- [Restart, UpdateVersion, Reconfigure, scaling, RotateAuth and certificate rotation](/docs/guides/durableworkflow/update-version/overview.md) run without downtime.

On a cluster that is not highly available the same operations still work, but a role with one replica is unavailable while its pod restarts. The `UpdateVersion` webhook warns you about it.

Readiness means able to serve, not that a port is open. For Temporal, the pods use the gRPC health protocol, so a replaced history pod becomes ready only when it owns shards again. For Cadence, the frontend runs a CLI check and the other roles check the port.

A sample with sizes:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-sized
  namespace: demo
spec:
  version: temporal-1.32.1
  numHistoryShards: 512
  topology:
    frontend:
      replicas: 3
      podTemplate:
        spec:
          containers:
          - name: durableworkflow
            resources:
              requests:
                cpu: 500m
                memory: 512Mi
              limits:
                memory: 1Gi
    history:
      replicas: 3
      podTemplate:
        spec:
          containers:
          - name: durableworkflow
            resources:
              requests:
                cpu: "1"
                memory: 2Gi
              limits:
                memory: 2Gi
    matching:
      replicas: 2
    worker:
      replicas: 2
    internalFrontend:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  deletionPolicy: WipeOut
```

The container of every role is named `durableworkflow`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/clustering/temporal-sized.yaml
durableworkflow.kubedb.com/temporal-sized created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Next Steps

- [Scale](/docs/guides/durableworkflow/scaling/overview.md) the roles.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
