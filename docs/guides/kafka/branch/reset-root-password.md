---
title: Resetting a Kafka Branch Root Password
menu:
  docs_{{ .version }}:
    identifier: kafka-branch-reset-root-password
    name: ResetRootPassword
    parent: guides-kafka-branch
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Resetting a Kafka Branch Root Password

Set `spec.resetRootPassword: true` to give the branch a separate administrator SASL password. Rotating branch credentials reduces the risk of exposing or reusing production credentials in development and test environments. The source retains its original password.

This independent tutorial deploys a source and a same-cluster branch in another namespace, compares their passwords, and verifies which credentials each deployment accepts.

## Execution order

Courier first copies the source credential so the target can start with the cloned data. KubeDB Ops Manager then processes a target `KafkaOpsRequest` with type `RotateAuth`. Courier waits for rotation to succeed before completing password reset, running any post-actions, and reporting the Branch `Ready`. Different Secret names alone do not establish different passwords.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation) with KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `KafkaVersion` `3.9.0` is available in your KubeDB catalog.
- Use Ceph RBD storage supporting copy-on-write snapshots and clones. These three-member examples use compatible VolumeSnapshot and VolumeGroupSnapshot classes, with the corresponding CSI APIs, controllers, and driver support. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- Install KubeDB Ops Manager and the `KafkaOpsRequest` CRD; password reset uses a `RotateAuth` request on the target.
- Allow capacity for three source and three target database pods, each requesting 1200Mi of memory, plus operators and any action Jobs.
- Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.

The source uses namespace `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the names in this tutorial and create namespaces if they do not already exist.

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `kafka-reset-source.yaml`. It creates a three-node combined broker/controller cluster:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-reset-source
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
kubectl -n demo apply -f kafka-reset-source.yaml
kubectl -n demo wait kafka/kafka-reset-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

## Record the source credential

Before creating the Branch, record the decoded source password for comparison and the later rejection check.

Discover the auth Secret referenced by `kafka-reset-source`:

```shell
kubectl -n demo get kafka kafka-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
kafka-reset-source-auth
```

Use the returned Secret name below if it differs. Decode its username and password:

```shell
kubectl -n demo get secret kafka-reset-source-auth -o jsonpath='{.data.username}' | base64 -d
kubectl -n demo get secret kafka-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

The generated username in this example is `admin`. Enter the decoded password when the client prompts.

## Seed and inspect the source data

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it kafka-reset-source-0 -c kafka -- bash
```

Run the following Kafka commands **inside this Bash shell**. The operator has written the broker pod’s SASL settings to `/opt/kafka/config/clientauth.properties`. Use the headless Service for bootstrap; Kafka advertises the individual broker addresses after connection.

Create a single-partition topic replicated to all three nodes. `flush.messages=1` makes this small demonstration flush each topic record to disk; it is not a general production tuning recommendation.

```shell
kafka-topics.sh --bootstrap-server kafka-reset-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --create --topic customers --partitions 1 --replication-factor 3 --config flush.messages=1
```

Start the producer:

```shell
kafka-console-producer.sh --bootstrap-server kafka-reset-source-pods.demo.svc.cluster.local:9092 \
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
kafka-get-offsets.sh --bootstrap-server kafka-reset-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-reset-source-pods.demo.svc.cluster.local:9092 \
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

## Create a Branch with password reset enabled

Save the following as `kafka-reset-branch.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: kafka-reset-branch
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Kafka
      name: kafka-reset-source
    namespace: demo
  target:
    namespace: dev
    name: kafka-reset-branch
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1200Mi
      limits:
        memory: 1200Mi
  resetRootPassword: true
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f kafka-reset-branch.yaml
kubectl -n demo wait branch/kafka-reset-branch --for=condition=Ready --timeout=60m
kubectl -n demo get branch kafka-reset-branch
```

Courier creates `kafka/kafka-reset-branch` and its cloned PVCs in `dev`; KubeDB provisions the target database.

Inspect the initial generation’s rotation request in `dev`:

```shell
kubectl -n dev get kafkaopsrequest kafka-reset-branch-branch-reset-0 -o custom-columns='NAME:.metadata.name,TYPE:.spec.type,PHASE:.status.phase'
```

```text
NAME                                  TYPE         PHASE
kafka-reset-branch-branch-reset-0   RotateAuth   Successful
```

Check the Branch conditions:

```shell
kubectl -n demo get branch kafka-reset-branch -o jsonpath='{range .status.conditions[*]}{.type}{"\t"}{.status}{"\n"}{end}'
```

Expect `RootPasswordReset=True` and `Ready=True`. The following checks establish that the intended credentials actually authenticate.

## Compare source and branch passwords

Discover both Secret names:

```shell
kubectl -n demo get kafka kafka-reset-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
kubectl -n dev get kafka kafka-reset-branch -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
kafka-reset-source-auth
kafka-reset-branch-auth
```

Use the returned names if they differ. Read each password separately:

```shell
kubectl -n demo get secret kafka-reset-source-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
source-example-password
```

```shell
kubectl -n dev get secret kafka-reset-branch-auth -o jsonpath='{.data.password}' | base64 -d
```

```text
branch-example-password
```

The displayed strings are illustrative. Your source password must match the value recorded before branching, while the branch password must differ.

## Verify authentication through the Services

Open the source pod shell from your terminal:

```shell
kubectl -n demo exec -it kafka-reset-source-0 -c kafka -- bash
```

### Connect with the branch credential

Inside the pod, create a separate client file. Replace `BRANCH_PASSWORD` with the decoded branch password; keep the source’s operator-generated file unchanged:

```shell
printf '%s\n' \
  'security.protocol=SASL_PLAINTEXT' \
  'sasl.mechanism=PLAIN' \
  'sasl.jaas.config=org.apache.kafka.common.security.plain.PlainLoginModule required username="admin" password="BRANCH_PASSWORD";' \
  > /tmp/branch-client.properties
```

Authenticate to the branch and read its copied records:

```shell
kafka-broker-api-versions.sh --bootstrap-server kafka-reset-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /tmp/branch-client.properties
```

Successful output lists the branch brokers and their API versions. Read the five records:

```shell
kafka-get-offsets.sh --bootstrap-server kafka-reset-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /tmp/branch-client.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-reset-branch-pods.dev.svc.cluster.local:9092 \
  --consumer.config /tmp/branch-client.properties --topic customers --from-beginning --max-messages 5 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
```

### Confirm the source credential is rejected by the branch

Use the source pod’s original SASL configuration against the **same branch bootstrap address**:

```shell
kafka-broker-api-versions.sh --bootstrap-server kafka-reset-branch-pods.dev.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties
```

```text
SaslAuthenticationException: Authentication failed: Invalid username or password
```

### Confirm the source credential still works

Use that same source client file against the source bootstrap address:

```shell
kafka-get-offsets.sh --bootstrap-server kafka-reset-source-pods.demo.svc.cluster.local:9092 \
  --command-config /opt/kafka/config/clientauth.properties --topic customers --time latest
```

```text
customers:0:5
```

```shell
kafka-console-consumer.sh --bootstrap-server kafka-reset-source-pods.demo.svc.cluster.local:9092 \
  --consumer.config /opt/kafka/config/clientauth.properties --topic customers --from-beginning --max-messages 5 --timeout-ms 20000
```

```text
{"id": 1, "full_name": "Customer 1", "email": "customer1@example.com"}
{"id": 2, "full_name": "Customer 2", "email": "customer2@example.com"}
{"id": 3, "full_name": "Customer 3", "email": "customer3@example.com"}
{"id": 4, "full_name": "Customer 4", "email": "customer4@example.com"}
{"id": 5, "full_name": "Customer 5", "email": "customer5@example.com"}
```

All five original records remain. Remove the temporary branch client file and leave the pod shell:

```shell
rm /tmp/branch-client.properties
exit
```

## Use password reset with other branch features

On each [refresh](/docs/guides/kafka/branch/refresh.md), the cloned data brings back the source credential and Courier performs a new rotation before readiness. Re-read the branch Secret after refresh. Configured [post-actions](/docs/guides/kafka/branch/post-actions.md) run after password reset and receive branch credentials. For another cluster, use the same setting on the [cross-cluster Branch](/docs/guides/kafka/branch/cross-cluster/index.md) and inspect the rotation request on the target.

See the common [password-reset field reference](/docs/operatormanual/branching/reset-root-password.md#field-reference).

## Cleanup

Delete the Branch on the source cluster first and let Courier finish removing its target and snapshots. Then delete the source database:

```shell
kubectl -n demo delete branch kafka-reset-branch --wait=true --timeout=30m
kubectl -n demo delete kafka kafka-reset-source --wait=true --timeout=30m
```

Keep Courier, the storage backend, and the OCM control plane when used available until Branch deletion completes. The commands leave the namespaces in place.
