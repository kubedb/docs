---
title: Refreshing a Kafka Branch
menu:
  docs_{{ .version }}:
    identifier: kafka-branch-refresh
    name: Refresh
    parent: guides-kafka-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Refreshing a Kafka Branch

Set `spec.schedule.cron` to periodically replace branch data with a fresh source snapshot. This guide deploys its own source and scheduled branch, changes both sides, and verifies what a completed refresh preserves and replaces.

## What happens during refresh

1. Courier advances the refresh generation and takes fresh source snapshots.
2. It imports snapshot references into the target namespace, pauses the target, and replaces its cloned volumes.
3. KubeDB restarts the target using the new volumes.
4. After the target and any configured preparation complete, the Branch returns to `Ready`; Courier releases the previous generation’s snapshots.

**Refresh discards branch-only writes and temporarily interrupts the target.** It replaces data rather than merging changes. The source remains independent. If configured, [password reset](/docs/guides/kafka/branch/reset-root-password.md) and [post-actions](/docs/guides/kafka/branch/post-actions.md) run again on each refresh.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `KafkaVersion` `3.9.0` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Allow capacity for three source and three target database pods, each requesting 1200Mi of memory, plus operators and any action Jobs.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `kafka-refresh-source.yaml`. It creates a three-node combined broker/controller cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-refresh-source
  namespace: demo
spec:
  version: 3.9.0
  replicas: 3
  podTemplate:
    spec:
      containers:
      - name: kafka
        resources:
          requests:
            cpu: 500m
            memory: 1200Mi
          limits:
            memory: 1200Mi
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```shell
kubectl -n demo apply -f kafka-refresh-source.yaml
kubectl -n demo wait kafka/kafka-refresh-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-refresh-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Create a single-partition topic replicated to all three nodes. `flush.messages=1` makes this small demonstration flush each topic record to disk; it is not a general production tuning recommendation.

```shell
kafka-topics.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --create --topic customers --partitions 1 --replication-factor 3 --config flush.messages=1
```

Start the producer:

```shell
kafka-console-producer.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --producer.config /opt/kafka/config/clientauth.properties --topic customers --request-required-acks all
```

Paste these five messages, one per line, then press **Ctrl+D** to leave the producer and return to the pod shell:

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
```

Inspect both the end offset and the actual messages:

```shell
kafka-get-offsets.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 5 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
```

Offsets alone do not prove that the cloned messages are readable. The bounded consumer should return all five records above.

Run `exit` to return to your terminal.

## Create a scheduled Branch

The example runs on ten-minute clock boundaries, such as `10:00`, `10:10`, and `10:20`; it is not a ten-minute timer from creation.

Save the following as `kafka-refresh-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-refresh-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-refresh-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-refresh-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
  schedule:
    cron: '*/10 * * * *'
  historyLimit:
    success: 3
    failed: 2
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f kafka-refresh-branch.yaml
kubectl -n demo wait branch/kafka-refresh-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-refresh-branch
```

Courier creates `kafka/kafka-refresh-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

### Check the initial copy

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-refresh-branch-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 5 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
```

Run `exit` to return to your terminal.

### Record the current generation and schedule

```shell
kubectl -n demo get branch kafka-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime,NEXT:.status.nextRefreshTime'
```

Record the generation and successful completion timestamp before making the following edits. Initial generation `0` can appear as `<none>`. Make both edits before the next scheduled run; if one starts while you are editing, finish the edits and wait for a subsequent generation. Runs that are still active can cause scheduled ticks to be skipped.

## Change source and branch data

### Update the source

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-refresh-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Start the producer, enter the following record, then press **Ctrl+D**:

```shell
kafka-console-producer.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --producer.config /opt/kafka/config/clientauth.properties --topic customers --request-required-acks all
```

```text
SOURCE-ONLY
```

```shell
kafka-get-offsets.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 6 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
SOURCE-ONLY
```

Run `exit` to return to your terminal.

### Make a branch-only change

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-refresh-branch-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Start the producer, enter the following record, then press **Ctrl+D**:

```shell
kafka-console-producer.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --producer.config /opt/kafka/config/clientauth.properties --topic customers --request-required-acks all
```

```text
BRANCH-ONLY
```

```shell
kafka-get-offsets.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 6 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
BRANCH-ONLY
```

Run `exit` to return to your terminal.

## Wait for the next completed refresh

```shell
kubectl -n demo get branch kafka-refresh-branch -o custom-columns='NAME:.metadata.name,PHASE:.status.phase,GENERATION:.status.refreshGeneration,LAST-SUCCESS:.status.lastSuccessfulRefreshTime' --watch
```

Illustrative output:

```text
NAME                   PHASE        GENERATION   LAST-SUCCESS
kafka-refresh-branch   Ready        <none>       2026-11-11T10:02:00Z
kafka-refresh-branch   Refreshing   1            2026-11-11T10:02:00Z
kafka-refresh-branch   Ready        1            2026-11-11T10:18:00Z
```

Wait for **a later generation, phase `Ready`, and an advanced `LAST-SUCCESS` timestamp**. An initial `Ready` row or an increased generation alone does not establish completion. Press **Ctrl+C** to stop watching. Recovery duration depends on the engine and available resources.

```shell
kubectl -n demo get branch kafka-refresh-branch -o yaml
```

Inspect `status.history`, `lastSuccessfulRefreshTime`, and `snapshot.generation`. With the group-snapshot prerequisite above, `snapshot.strategy` is `VolumeGroupSnapshot`; per-volume fallback can occur if no compatible group class is available. `FRESHNESS` measures time since successful completion, not the source snapshot’s age.

## Verify the refreshed data

### Check the branch

Open a new pod/client session after recovery; the previous session may have been disconnected.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-refresh-branch-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-refresh-branch-pods.dev.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 6 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
SOURCE-ONLY
```

Run `exit` to return to your terminal.

The branch now has the source’s six records, including `SOURCE-ONLY`. Its earlier `BRANCH-ONLY` record is gone.

### Confirm the source remains unchanged by refresh

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-refresh-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-refresh-source-pods.demo.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 6 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
SOURCE-ONLY
```

Run `exit` to return to your terminal.

The source retains its own update and never receives the branch-only change.

## Apply the same strategy across clusters

Use the [cross-cluster tutorial](/docs/guides/kafka/branch/cross-cluster/index.md) and add the same schedule to its Branch. Manage the original Branch on the source cluster, inspect the database on the target, and wait for a later completed generation before checking data.

See the common [schedule and history fields](/docs/operatormanual/branching/refresh.md#schedule-and-history-field-reference) and [refresh status reference](/docs/operatormanual/branching/refresh.md#refresh-status-reference) for configuration details.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch kafka-refresh-branch --wait=true --timeout=30m
kubectl -n demo delete kafka kafka-refresh-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
