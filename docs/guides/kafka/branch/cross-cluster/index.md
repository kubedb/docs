---
title: Branch Kafka Across Clusters
menu:
  docs_{{ .version }}:
    identifier: kafka-branch-cross-cluster
    name: Cross-Cluster
    parent: guides-kafka-branch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch Kafka Across Clusters

Cross-cluster branching creates an independent Kafka deployment on another Kubernetes cluster from a point-in-time snapshot of an existing deployment. Use it to prepare development, testing, or data analysis environments on a separate cluster using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `KafkaVersion` `3.9.0` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Have existing OCM hub registration for managed clusters `source` and `target`, with Courier agents enabled. Configure kubectl contexts `hub`, `source`, and `target`.
- Both clusters must access **the same network-reachable Ceph backend** with compatible CSI identities, pool settings, and credentials. Matching StorageClass names alone are insufficient.
- Allow capacity for three source and three target database pods, each requesting 1200Mi of memory, plus operators and any action Jobs.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl --context source create namespace demo
kubectl --context target create namespace dev
```

Check the existing OCM registration and storage classes:

```shell
kubectl --context hub get managedclusters source target
kubectl --context source get storageclass ceph-rbd
kubectl --context target get storageclass ceph-rbd
```

## Deploy the source database

Save the following as `kafka-source.yaml`. It creates a three-node combined broker/controller cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-source
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
kubectl --context source -n demo apply -f kafka-source.yaml
kubectl --context source -n demo wait kafka/kafka-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it kafka-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Create a single-partition topic replicated to all three nodes. `flush.messages=1` makes this small demonstration flush each topic record to disk; it is not a general production tuning recommendation.

```shell
kafka-topics.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --create --topic customers --partitions 1 --replication-factor 3 --config flush.messages=1
```

Start the producer:

```shell
kafka-console-producer.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
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
kafka-get-offsets.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
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

## Create the cross-cluster Branch

Save the following as `kafka-cross.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-cross
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-cross
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
    clusterName: target
  deletionPolicy: Delete
```

```shell
kubectl --context source -n demo apply -f kafka-cross.yaml
kubectl --context source -n demo wait branch/kafka-cross --for=condition=Ready --timeout=60m
kubectl --context source -n demo get branch kafka-cross
```

The Branch is created on `source`. Courier sends the target resource manifests through the OCM hub; the target KubeDB provisioner creates the database using the delivered cloned PVCs. `target.clusterName` is the OCM ManagedCluster name.

```shell
kubectl --context hub -n source get branchworks
kubectl --context hub -n target get manifestworks
kubectl --context target -n dev get kafka kafka-cross
```

## Verify the copied data

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it kafka-cross-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
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

### Check the target cluster

```shell
kubectl --context target -n dev get pods,pvc -l app.kubernetes.io/instance=kafka-cross
```

The target has three combined nodes. Inspect the copied topic:

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it kafka-cross-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-topics.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --describe --topic customers
```

Expect `PartitionCount: 1`, `ReplicationFactor: 3`, and three in-sync replicas after recovery.

Run `exit` to return to your terminal.

## Verify independent writes

### Change the branch

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it kafka-cross-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Start the producer, enter the following record, then press **Ctrl+D**:

```shell
kafka-console-producer.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
  --producer.config /opt/kafka/config/clientauth.properties --topic customers --request-required-acks all
```

```text
BRANCH-ONLY
```

```shell
kafka-get-offsets.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
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

### Change and inspect the source

From your terminal, open a pod shell:

```shell
kubectl --context source -n demo exec -it kafka-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Start the producer, enter the following record, then press **Ctrl+D**:

```shell
kafka-console-producer.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
  --producer.config /opt/kafka/config/clientauth.properties --topic customers --request-required-acks all
```

```text
SOURCE-ONLY
```

```shell
kafka-get-offsets.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-source-pods.demo.svc.cluster.local:9092 \
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

The source contains `SOURCE-ONLY` but no `BRANCH-ONLY` record.

### Check the branch again

From your terminal, open a pod shell:

```shell
kubectl --context target -n dev exec -it kafka-cross-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-cross-pods.dev.svc.cluster.local:9092 \
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

The branch contains `BRANCH-ONLY` but no `SOURCE-ONLY` record. Each topic now has six records, with different final records.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl --context source -n demo delete branch kafka-cross --wait=true --timeout=30m
kubectl --context source -n demo delete kafka kafka-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/kafka/branch/refresh.md).
- [Reset the branch root password](/docs/guides/kafka/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/kafka/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/kafka/branch/customization/index.md).
