---
title: GitOps YugabyteDB using the KubeDB GitOps Operator
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-gitops-guide
    name: GitOps YugabyteDB
    parent: yugabytedb-gitops
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps YugabyteDB using the KubeDB GitOps Operator

> **Not verified on a cluster.** The YugabyteDB GitOps operator has never run against a real API server, webhooks or ops-manager. What was checked: its Go tests pass (21 cases with fake objects, covering the diff to ops request mapping and the immutable field checks), and the build and vet are clean. It lives on the branch `yugabytedb` of the `kubedb/gitops` repository, and its API type (`gitops.kubedb.com/v1alpha1` `YugabyteDB`) is on the branch `yugabytedb-autoscaler` of `apimachinery`, not yet merged into `yugabytedb`.

With the GitOps operator you keep a `YugabyteDB` manifest of the API group `gitops.kubedb.com/v1alpha1` in Git (for example synced by ArgoCD or FluxCD). The operator creates the real `YugabyteDB` for it, and when you change the manifest it works out the difference and creates the matching `YugabyteDBOpsRequest` instead of editing the running database directly.

Setting up the GitOps tool and the repository is the same as for the other databases; follow the [Etcd GitOps guide](/docs/guides/etcd/gitops/gitops.md) for it. Install KubeDB with the GitOps CRDs (`--set kubedb-crd-manager.installGitOpsCRDs=true`) and the YugabyteDB feature gate (`--set featureGates.YugabyteDB=true`).

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: YugabyteDB
metadata:
  name: yb-gitops
  namespace: demo
spec:
  version: 2025.2.8
  replicationFactor: 3
  topology:
    master:
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
          containers:
          - name: yb-master
            resources:
              requests: {cpu: 500m, memory: 1Gi}
              limits: {memory: 1Gi}
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 2Gi}}
    tserver:
      replicas: 3
      podTemplate:
        spec:
          podPlacementPolicy:
            name: spread-across-zones
          containers:
          - name: yb-tserver
            resources:
              requests: {cpu: "1", memory: 2Gi}
              limits: {memory: 2Gi}
      storage:
        accessModes: [ReadWriteOnce]
        resources: {requests: {storage: 5Gi}}
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/gitops/yugabytedb.yaml
```

## What a manifest change becomes

| Change in the manifest | Ops request |
|---|---|
| `spec.version` | `UpdateVersion` |
| `spec.topology.tserver.replicas`, `spec.topology.readReplica.replicas` | `HorizontalScaling` |
| container `resources` of master, tserver or readReplica | `VerticalScaling` |
| per role `storage` size | `VolumeExpansion` (a smaller size is rejected) |
| `spec.configuration`, `spec.ysql.connectionManager` | `Reconfigure` |
| `spec.tls`, including `nodeToNode`, `clientToNode` and `webserver` | `ReconfigureTLS` |
| `spec.authSecret` | `RotateAuth` |
| `storageClassName` of the volumes | `StorageMigration` (roles that ask for different classes are rejected) |
| `spec.monitor`, `spec.placement`, pod template changes that are not resources | `Restart` |
| `spec.serviceTemplates`, `spec.healthChecker`, `spec.deletionPolicy`, `spec.halted` | patched directly, no ops request |

The number of masters follows `spec.replicationFactor`, so it is never a scaling difference.

## What is rejected

The reconciler refuses a change to these fields with a message, because they cannot change after creation:

- `spec.replicationFactor`
- `spec.ysql.enabled`, `spec.ycql.enabled`
- `spec.disableSecurity`
- per role `dataVolumes` and `storageType`
- `spec.topology.readReplica.placementUUID`
- adding or removing the `readReplica` block

## Not covered

`Rewind`, `Clone`, `FinalizeUpgrade`, `RollbackUpgrade` and `spec.pitr` are not mapped by the GitOps operator according to its source. An upgrade done through GitOps uses the default `finalize` mode; this page does not describe how to request a manual finalize through Git.

## Cleaning up

```bash
$ kubectl delete yb -n demo yb-gitops
```
