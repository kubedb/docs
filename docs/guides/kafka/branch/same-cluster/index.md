---
title: Branch Kafka into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: kafka-branch-same-cluster
    name: Same-Cluster
    parent: guides-kafka-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch Kafka into Another Namespace

Same-cluster branching creates an independent Kafka deployment from a point-in-time snapshot within the same Kubernetes cluster. Place the branch in another namespace to prepare development, testing, or data analysis environments using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

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
kubectl -n demo apply -f kafka-source.yaml
kubectl -n demo wait kafka/kafka-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-source-0 -c kafka -- bash
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

## Create the Branch

Save the following as `kafka-dev.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-dev
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
    name: kafka-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f kafka-dev.yaml
kubectl -n demo wait branch/kafka-dev --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-dev
```

Courier creates `kafka/kafka-dev` and its cloned PVCs in `dev`; KubeDB provisions the target database.

## Verify the copied data

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-dev-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
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
kubectl -n dev get pods,pvc -l app.kubernetes.io/instance=kafka-dev
```

The target has three combined nodes. Inspect the copied topic:

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-dev-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-topics.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --describe --topic customers
```

Expect `PartitionCount: 1`, `ReplicationFactor: 3`, and three in-sync replicas after recovery.

Run `exit` to return to your terminal.

## Verify independent writes

### Change the branch

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-dev-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Start the producer, enter the following record, then press **Ctrl+D**:

```shell
kafka-console-producer.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
  --producer.config /opt/kafka/config/clientauth.properties --topic customers --request-required-acks all
```

```text
BRANCH-ONLY
```

```shell
kafka-get-offsets.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
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
kubectl -n demo exec -it kafka-source-0 -c kafka -- bash
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
kubectl -n dev exec -it kafka-dev-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-get-offsets.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-dev-pods.dev.svc.cluster.local:9092 \
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
kubectl -n demo delete branch kafka-dev --wait=true --timeout=30m
kubectl -n demo delete kafka kafka-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/kafka/branch/refresh.md).
- [Reset the branch root password](/docs/guides/kafka/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/kafka/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/kafka/branch/customization/index.md).
