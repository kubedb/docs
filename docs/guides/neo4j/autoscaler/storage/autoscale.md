---
title: Autoscale Neo4j Storage
menu:
  docs_{{ .version }}:
    identifier: neo4j-storage-autoscaling-guide
    name: Autoscale Storage
    parent: neo4j-storage-autoscaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscale Neo4j Storage

This guide configures KubeDB to expand Neo4j data volumes automatically. We insert payload-bearing graph records until the demonstration threshold is crossed, observe the generated volume-expansion operation, and verify both the new capacity and the stored data.

## Before You Begin

| Requirement | Details |
|---|---|
| KubeDB | Provisioner, Ops Manager, and Autoscaler operators must be installed. |
| Storage metrics | Install KubeDB with `--set kubedb-autoscaler.storage-metrics-server.enabled=true`. |
| Expandable storage | This example uses `longhorn`; the selected StorageClass must report `ALLOWVOLUMEEXPANSION=true`. |
| Tools | `kubectl`, `base64`, and a POSIX-compatible shell must be available locally. |

See [Neo4jAutoscaler](/docs/guides/neo4j/concepts/autoscaler.md) and the [storage autoscaling overview](/docs/guides/neo4j/autoscaler/storage/overview.md) for background.

Verify storage expansion and the custom metrics API before continuing:

```bash
$ kubectl get storageclass longhorn
NAME       PROVISIONER          RECLAIMPOLICY   ALLOWVOLUMEEXPANSION
longhorn   driver.longhorn.io   Delete          true

$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1 | head
```

## Deploy Neo4j

Create an isolated namespace:

```bash
$ kubectl create namespace demo
namespace/demo created
```

The following manifest creates a three-member Neo4j cluster with the minimum `2Gi` of storage per member. The selected StorageClass must support volume expansion:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-autoscale
  namespace: demo
spec:
  version: "2025.12.1"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  podTemplate:
    spec:
      containers:
        - name: neo4j
          resources:
            requests:
              cpu: 500m
              memory: 2Gi
            limits:
              cpu: 500m
              memory: 2Gi
  deletionPolicy: WipeOut
```

Here, `spec.version` selects an installed `Neo4jVersion`, `replicas: 3` creates a fault-tolerant cluster, and `storageType: Durable` provisions one PVC per member. `deletionPolicy: WipeOut` is convenient for a disposable tutorial but also removes the database-owned PVCs and credentials when the Neo4j resource is deleted.

Apply the same manifest from the examples directory and wait for Neo4j to become ready:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/autoscaler/neo4j.yaml
neo4j.kubedb.com/neo4j-autoscale created

$ kubectl get neo4j -n demo neo4j-autoscale -w
NAME              VERSION     STATUS   AGE
neo4j-autoscale   2025.12.1   Ready    3m
```

Confirm the initial PVC capacities:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=neo4j-autoscale \
    -o custom-columns=NAME:.metadata.name,CAPACITY:.status.capacity.storage
NAME                       CAPACITY
data-neo4j-autoscale-0     2Gi
data-neo4j-autoscale-1     2Gi
data-neo4j-autoscale-2     2Gi
```

## Create the Neo4jAutoscaler

The following policy triggers when a data volume reaches `40%` usage and increases its current size by `50%`:

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: Neo4jAutoscaler
metadata:
  name: neo4j-storage-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: neo4j-autoscale
  opsRequestOptions:
    apply: IfReady
    timeout: 10m
    maxRetries: 3
  storage:
    neo4j:
      trigger: "On"
      usageThreshold: 40
      scalingThreshold: 50
      expansionMode: Online
```

Apply it and confirm that it is active:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/neo4j/autoscaler/storage/neo4j-storage-autoscaler.yaml
neo4jautoscaler.autoscaling.kubedb.com/neo4j-storage-autoscaler created

$ kubectl get neo4jautoscaler -n demo neo4j-storage-autoscaler
NAME                         AGE
neo4j-storage-autoscaler     10s
```

> The `40%` threshold is intentionally low so the tutorial completes quickly. It also remains above the observed usage after the first expansion, preventing this sample workload from immediately triggering another operation. Use a higher threshold, such as `80%`, for a production policy and leave sufficient headroom for traffic spikes and expansion time.

## Insert Graph Data

Retrieve the admin password, create an application database, and add a uniqueness constraint so a batch can be retried safely:

```bash
$ PASS=$(kubectl get secret -n demo neo4j-autoscale-auth \
    -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -u neo4j -p "$PASS" \
    "CREATE DATABASE appdb IF NOT EXISTS WAIT"

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "CREATE CONSTRAINT event_id IF NOT EXISTS
     FOR (e:Event) REQUIRE e.id IS UNIQUE"
```

Insert events in bounded transactions. Every event contains a payload of approximately 1 KiB, so this creates real Neo4j store and transaction-log growth without writing unrelated files into the volume:

```bash
$ START_BATCH=0
$ for batch in $(seq "$START_BATCH" "$((START_BATCH + 99))"); do
    kubectl exec -n demo neo4j-autoscale-0 -- \
      cypher-shell -d appdb -u neo4j -p "$PASS" \
      "UNWIND range(1,2500) AS i
       MERGE (e:Event {id: $((batch * 2500)) + i})
       ON CREATE SET
         e.source = 'storage-autoscaling-demo',
         e.payload = reduce(s = '', n IN range(1,32) | s + randomUUID())"
  done
```

Check the data and filesystem usage:

```bash
$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "MATCH (e:Event) RETURN count(e) AS events"
events
250000

$ kubectl exec -n demo neo4j-autoscale-0 -- df -h /data
Filesystem      Size  Used Avail Use% Mounted on
/dev/longhorn   2.0G  962M  955M  51% /data
```

Actual usage varies because Neo4j store files and the storage backend have their own overhead. If usage is still below `40%`, set `START_BATCH=100` and run another batch. Increase it by `100` for each additional run, and stop inserting once the threshold is crossed.

## Observe Volume Expansion

The Autoscaler creates a `Neo4jOpsRequest` of type `VolumeExpansion` after the storage metric reaches the threshold:

```bash
$ kubectl get neo4jopsrequest -n demo -w
NAME                              TYPE              STATUS        AGE
neoops-neo4j-autoscale-xxxxxx     VolumeExpansion   Progressing   20s
neoops-neo4j-autoscale-xxxxxx     VolumeExpansion   Successful    2m
```

Inspect the operation to see the calculated target and completed steps:

```bash
$ kubectl describe neo4jopsrequest -n demo neoops-neo4j-autoscale-xxxxxx
```

With a `50%` scaling threshold, the target is approximately 50% larger than the usable capacity reported by the storage metrics. A nominal `2Gi` PVC produced `2920Mi` PVCs in this test; the exact value can differ slightly by filesystem and storage backend:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=neo4j-autoscale \
    -o custom-columns=NAME:.metadata.name,CAPACITY:.status.capacity.storage
NAME                       CAPACITY
data-neo4j-autoscale-0     2920Mi
data-neo4j-autoscale-1     2920Mi
data-neo4j-autoscale-2     2920Mi
```

Verify the expanded filesystem and the application data:

```bash
$ kubectl exec -n demo neo4j-autoscale-0 -- df -h /data
Filesystem      Size  Used Avail Use% Mounted on
/dev/longhorn   2.8G  962M  1.8G  35% /data

$ kubectl exec -n demo neo4j-autoscale-0 -- \
    cypher-shell -d appdb -u neo4j -p "$PASS" \
    "MATCH (e:Event) RETURN count(e) AS events"
events
250000
```

## Troubleshooting

- If the custom metrics endpoint is unavailable, verify that the KubeDB storage metrics server is enabled and healthy.
- If no OpsRequest appears, describe the Autoscaler and confirm that `/data` usage is above `usageThreshold`.
- If a PVC remains at its old capacity, confirm `allowVolumeExpansion: true` and inspect PVC events.
- Check the storage backend's health before testing expansion. For example, every Longhorn volume must have enough schedulable replicas; a degraded volume can reject resize requests.
- If a PVC reports `FileSystemResizePending`, the CSI driver requires the volume to be remounted before the filesystem sees the new capacity. Use `expansionMode: Offline` for such drivers; Neo4j pods will be restarted during expansion.

## Cleaning Up

Deleting the example database with `deletionPolicy: WipeOut` also deletes its data volumes:

```bash
$ kubectl delete neo4jautoscaler -n demo neo4j-storage-autoscaler
$ kubectl delete neo4j -n demo neo4j-autoscale
$ kubectl delete namespace demo
```
