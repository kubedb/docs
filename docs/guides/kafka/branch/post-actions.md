---
title: Running Kafka Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: kafka-branch-post-actions
    name: Post-Actions
    parent: guides-kafka-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Kafka Branch Post-Actions

Post-actions prepare a Kafka branch before it becomes ready for use. They can create development topics or publish initialization records. Kafka records are immutable: this guide creates a branch-only topic and appends a marker rather than rewriting existing messages.

## Execution and readiness

Each `postActions` entry creates a Job in the target namespace. This example uses one entry and one Job. Courier waits for target database readiness, then runs the action; the Branch becomes `Ready` only after it succeeds. A configured password reset would run before the action.

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

Save the following as `kafka-pa-source.yaml`. It creates a three-node combined broker/controller cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-pa-source
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
kubectl -n demo apply -f kafka-pa-source.yaml
kubectl -n demo wait kafka/kafka-pa-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Seed and inspect the source data

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-pa-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Create a single-partition topic replicated to all three nodes. `flush.messages=1` makes this small demonstration flush each topic record to disk; it is not a general production tuning recommendation.

```shell
kafka-topics.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --create --topic customers --partitions 1 --replication-factor 3 --config flush.messages=1
```

Start the producer:

```shell
kafka-console-producer.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
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
kafka-get-offsets.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
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

## Create a Branch with post-actions

Save the following as `kafka-pa-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-pa-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-pa-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-pa-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
  postActions:
  - image: apache/kafka:3.9.0
    jobDefaults:
      imagePullPolicy: IfNotPresent
      backoffLimit: 3
      activeDeadlineSeconds: 900
      ttlSecondsAfterFinished: 3600
    jobTemplate:
      metadata:
        labels:
          branch-action: kafka-pa-branch-prepare
      spec:
        args:
        - bash
        - -ec
        - |
          # Build the action pod's own SASL client configuration.
          printf 'security.protocol=SASL_PLAINTEXT\nsasl.mechanism=PLAIN\nsasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="%s" password="%s";\n' "$KAFKA_USER" "$KAFKA_PASSWORD" > /tmp/client.properties
          # First operation: create a topic used only by the branch.
          /opt/kafka/bin/kafka-topics.sh --bootstrap-server kafka-pa-branch-pods.dev.svc.cluster.local:9092 --command-config /tmp/client.properties --create --if-not-exists --topic post-action-ran --partitions 1 --replication-factor 3
          # Second operation: append a marker only after topic creation succeeds.
          printf '%s\n' 'POST-ACTION-APPLIED' | /opt/kafka/bin/kafka-console-producer.sh --bootstrap-server kafka-pa-branch-pods.dev.svc.cluster.local:9092 --producer.config /tmp/client.properties --topic customers --request-required-acks all
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f kafka-pa-branch.yaml
kubectl -n demo wait branch/kafka-pa-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-pa-branch
```

Courier creates `kafka/kafka-pa-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

The `apache/kafka` image executes the supplied Bash arguments; the KubeDB broker image starts a server and is unsuitable for this Job. `bash -ec` stops on a failed command. The action creates the topic first, then publishes `POST-ACTION-APPLIED`. Re-running it keeps the existing topic and can append another marker.

The action pod does not have the broker pod’s generated client file, so it builds one from Courier’s credentials. The explicit bootstrap address uses the actual headless Service in `dev`; do not substitute the injected bare `KAFKA_HOST`/`KAFKA_BOOTSTRAP` values while they refer to a Service that does not exist.

## Inspect the action Job

The Branch passes through `ActionsRunning`; the phase can finish before you observe it. Find its Job by Courier’s labels:

```shell
kubectl -n dev get jobs -l courier.kubedb.com/branch=kafka-pa-branch,courier.kubedb.com/branch-namespace=demo
kubectl -n dev logs job/kafka-pa-branch-post-action -c actions
kubectl -n dev get pods -l branch-action=kafka-pa-branch-prepare
```

Expect Job status `Complete` with `1/1` completions. The log reports creation of `post-action-ran`.

```shell
kubectl -n demo get branch kafka-pa-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
kubectl -n demo get branch kafka-pa-branch -o jsonpath='{.status.resources}{"\n"}'
```

Expect `PostActionsCompleted=True` and `Ready=True`. The resource references identify the target auth Secret, cloned PVCs, and the action Job.

## Verify that only the branch was changed

### Inspect the branch

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it kafka-pa-branch-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-topics.sh --bootstrap-server kafka-pa-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --list
```

The list includes `customers` and `post-action-ran` (operator health topics may also appear). The copied topic now has six records:

```shell
kafka-get-offsets.sh --bootstrap-server kafka-pa-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:6
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-pa-branch-pods.dev.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 6 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
POST-ACTION-APPLIED
```

Run `exit` to return to your terminal.

### Inspect the source

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-pa-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

```shell
kafka-topics.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --list
```

`post-action-ran` must be absent. The source still has only its original five messages:

```shell
kafka-get-offsets.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-pa-source-pods.demo.svc.cluster.local:9092 \
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

## Connection environment

| Variable | Value used by the action |
|---|---|
| `KAFKA_HOST` | Injected database name; use the actual headless Service in this example |
| `KAFKA_PORT` | 9092 |
| `KAFKA_BOOTSTRAP` | Injected host:port; this example supplies the headless Service explicitly |
| `KAFKA_USER` | Username from the target auth Secret |
| `KAFKA_PASSWORD` | Password from the target auth Secret |

See the common [action configuration and Job defaults](/docs/operatormanual/branching/post-actions.md#post-action-field-reference) for the feature fields. The image must execute its supplied arguments; the Branch action template does not expose a container `command` override.

## Use post-actions with refresh or another cluster

With [refresh](/docs/guides/kafka/branch/refresh.md), actions run again against each new copy before readiness. Generation `1` uses a Job named `kafka-pa-branch-post-action-1`; older generation Jobs can be removed by Courier. Design actions for repetition and use logs before cleanup. For [cross-cluster branching](/docs/guides/kafka/branch/cross-cluster/index.md), the Job runs on the target cluster; inspect it there and use the target namespace in any explicit connection address.

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch kafka-pa-branch --wait=true --timeout=30m
kubectl -n demo delete kafka kafka-pa-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
