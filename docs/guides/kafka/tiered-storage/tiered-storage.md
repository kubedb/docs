---
title: Tiered Storage
menu:
  docs_{{ .version }}:
    identifier: kf-tiered-storage-guides-docs
    name: Overview
    parent: kf-tiered-storage-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Tiered Storage

This tutorial will show you how to use KubeDB to run a [Tiered Storage](https://kafka.apache.org/41/operations/tiered-storage/). Kafka Tiered Storage is a feature that separates hot data and cold data by storing recent Kafka log segments on local broker disks and automatically offloading older segments to remote object storage (like S3, GCS, or Azure Blob).

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/install/_index.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [examples/kafka/tiered-storage/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/tiered-storage) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> This tutorial will only work from version 4.0.0 onwards.

## Create Secret for S3

Before creating a Kafka cluster with S3 tiered storage, you need to create a secret containing the AWS access key and secret key. Here's an example secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: aws-secret
  namespace: demo
type: Opaque
stringData:
  accessKeyId: YOUR_ACCESS_KEY_ID
  secretAccessKey: YOUR_SECRET_ACCESS_KEY
```

Apply the secret:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tiered-storage/kafka-s3-tiered-secret.yaml
secret/aws-secret created
```

## Create a Kafka Tiered Storage with S3 compatible storage

Here is an example Kafka CR that uses Tiered Storage with S3 compatible storage:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod-tiered
  namespace: demo
spec:
  version: 4.0.0
  tieredStorage:
    provider: s3
    s3:
      bucket: kafka
      endpoint: http://minio.demo.svc.cluster.local:80
      region: us-east-1
      secretName: aws-secret
      prefix: tiered-storage-demo/
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Here,

- `spec.tieredStorage` specifies the tiered storage configuration for the Kafka cluster.
  - `spec.tieredStorage.provider` specifies the tiered storage provider. Here, it is set to `s3`.
  - `spec.tieredStorage.s3` specifies the S3 compatible storage configuration.
    - `spec.tieredStorage.s3.bucket` specifies the S3 bucket name.
    - `spec.tieredStorage.s3.endpoint` specifies the S3 endpoint URL.
    - `spec.tieredStorage.s3.region` specifies the S3 region.
    - `spec.tieredStorage.s3.secretName` specifies the name of the secret that contains the S3 access key and secret key.
    - `spec.tieredStorage.s3.prefix` specifies the prefix for the S3 objects.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tiered-storage/kafka-s3-tiered.yaml
kafka.kubedb.com/kafka-prod-tiered created
```

```bash
$ kubectl get kafka -n demo -w
NAME                TYPE            VERSION   STATUS   AGE
kafka-prod-tiered   kubedb.com/v1   4.0.0     Provisioning   2s
kafka-prod-tiered   kubedb.com/v1   4.0.0     Provisioning   4s
.
.
kafka-prod-tiered   kubedb.com/v1   4.0.0     Ready          112s
```

Exec one of the broker pods and run the following command to create a tiered storage enabled topic and insert some data into it:

```bash
$ kubectl exec -n demo -it kafka-prod-tiered-broker-0 -- bash
root@kafka-prod-tiered-broker-0:/# kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create \
  --config remote.storage.enable=true \
  --config retention.ms=-1 \
  --config segment.bytes=1048576 \
  --config retention.bytes=104857600 \
  --config local.retention.bytes=1 \
  --partitions 1 \
  --replication-factor 1 \
  --topic topic1 \
  --command-config config/clientauth.properties

topic1 created

root@kafka-prod-tiered-broker-0:/# kafka-producer-perf-test.sh \
  --producer-props bootstrap.servers=localhost:9092 \
  --topic topic1 \
  --num-records 10000 \
  --record-size 512 \
  --throughput 1000 \
  --producer.config config/clientauth.properties

4998 records sent, 999.2 records/sec (0.49 MB/sec), 13.5 ms avg latency, 526.0 ms max latency.
10000 records sent, 999.3 records/sec (0.49 MB/sec), 8.51 ms avg latency, 526.00 ms max latency, 4 ms 50th, 50 ms 95th, 92 ms 99th, 92 ms 99.9th.
```

here, we created a topic with `local.retention.bytes=1` which will force kafka to offload segments to the remote tiered storage as soon as possible. You can check the S3 bucket to see the offloaded segments.

In this example, we are using an S3-compatible storage (MinIO). You can verify the offloaded segments using the `mc` (MinIO Client) command:
to the consumer transparently.

Verify Offloaded Segments in MinIO by running:

```bash
mc ls --recursive local/kafka
```

Example output:

```bash
[2026-02-19 23:42:14 +06] 2.0KiB STANDARD tiered-storage-demo/topic1-FzanGxsCRj6eR4xkkImQ9g/0/00000000000000000000-1qL0uiXzTrWBm-07BoNlnw.indexes
[2026-02-19 23:42:14 +06]1016KiB STANDARD tiered-storage-demo/topic1-FzanGxsCRj6eR4xkkImQ9g/0/00000000000000000000-1qL0uiXzTrWBm-07BoNlnw.log
[2026-02-19 23:42:14 +06]   736B STANDARD tiered-storage-demo/topic1-FzanGxsCRj6eR4xkkImQ9g/0/00000000000000000000-1qL0uiXzTrWBm-07BoNlnw.rsm-manifest
...
```

These files confirm that older Kafka log segments have been offloaded to MinIO (remote storage).

Inside the broker pod:

```bash
ls -lh /var/log/kafka/0/topic1-0
```

Example output:

```bash
total 104K
-rw-r--r-- 1 kafka kafka 10M Feb 19 17:42 00000000000000009844.index
-rw-r--r-- 1 kafka kafka 81K Feb 19 17:42 00000000000000009844.log
-rw-r--r-- 1 kafka kafka 56  Feb 19 17:42 00000000000000009844.snapshot
-rw-r--r-- 1 kafka kafka 10M Feb 19 17:42 00000000000000009844.timeindex
-rw-r--r-- 1 kafka kafka 8   Feb 19 17:41 leader-epoch-checkpoint
-rw-r--r-- 1 kafka kafka 43  Feb 19 17:41 partition.metadata
```

Notice that older segments are no longer present locally — they exist only in remote storage.

To consume data that has been offloaded, run the following inside the broker pod:

```bash
kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic topic1 \
  --from-beginning \
  --timeout-ms 15000 \
  --consumer.config config/clientauth.properties
```

Since the requested offsets are no longer available on local disk, Kafka must retrieve them from remote storage.

What Happens Internally

1. The consumer requests offset `0`.
2. The broker checks local storage for the required segment.
3. The segment is not found locally.
4. The broker fetches the segment from MinIO (remote storage).
5. The broker uses the remote log index cache.
6. The data is served to the consumer transparently.

The entire process is handled automatically, and the consumer is unaware whether the data came from local or remote storage.


> **Note**: You can set `local.retention.ms` instead of `local.retention.bytes` to offload segments based on time.

## Create a Kafka Tiered Storage with Azure compatible storage

Here is an example Kafka CR that uses Tiered Storage with Azure compatible storage:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod-tiered
  namespace: demo
spec:
  version: 4.0.0
  tieredStorage:
    provider: azure
    azure:
      container: kafka
      secretName: azure-secret
      prefix: tiered-storage-demo/
      storageAccount: demo-account
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Here,

- `spec.tieredStorage` specifies the tiered storage configuration for the Kafka cluster.
  - `spec.tieredStorage.provider` specifies the tiered storage provider. Here, it is set to `azure`.
  - `spec.tieredStorage.azure` specifies the azure compatible storage configuration.
    - `spec.tieredStorage.azure.container` specifies the azure container name.
    - `spec.tieredStorage.azure.secretName` specifies the name of the secret that contains the azure storage account key.
    - `spec.tieredStorage.azure.prefix` specifies the prefix for the azure blobs.
    - `spec.tieredStorage.azure.storageAccount` specifies the azure storage account name.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tiered-storage/kafka-azure-tiered.yaml
kafka.kubedb.com/kafka-prod-tiered created
```

```bash
$ kubectl get kafka -n demo -w
NAME                TYPE            VERSION   STATUS   AGE
kafka-prod-tiered   kubedb.com/v1   4.0.0     Provisioning   2s
kafka-prod-tiered   kubedb.com/v1   4.0.0     Provisioning   4s
.
.
kafka-prod-tiered   kubedb.com/v1   4.0.0     Ready          112s
```

## Create a Kafka Tiered Storage with GCS compatible storage

Here is an example Kafka CR that uses Tiered Storage with GCS compatible storage:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod-tiered
  namespace: demo
spec:
  version: 4.0.0
  tieredStorage:
    provider: gcs
    gcs:
      bucket: test-bucket
      secretName: gcs-secret
      prefix: tiered-storage-demo/
  topology:
    broker:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 3
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

Here,

- `spec.tieredStorage` specifies the tiered storage configuration for the Kafka cluster.
  - `spec.tieredStorage.provider` specifies the tiered storage provider. Here, it is set to `gcs`.
  - `spec.tieredStorage.gcs` specifies the gcs compatible storage configuration.
    - `spec.tieredStorage.gcs.bucket` specifies the gcs bucket name.
    - `spec.tieredStorage.gcs.secretName` specifies the name of the secret that contains the gcs service account key.
    - `spec.tieredStorage.gcs.prefix` specifies the prefix for the gcs objects.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/tiered-storage/kafka-gcs-tiered.yaml
kafka.kubedb.com/kafka-prod-tiered created
```

```bash
$ kubectl get kafka -n demo -w
NAME                TYPE            VERSION   STATUS   AGE
kafka-prod-tiered   kubedb.com/v1   4.0.0     Provisioning   2s
kafka-prod-tiered   kubedb.com/v1   4.0.0     Provisioning   4s
.
.
kafka-prod-tiered   kubedb.com/v1   4.0.0     Ready          112s
```

## Next Steps

- [Quickstart Kafka](/docs/guides/kafka/quickstart/kafka/index.md) with KubeDB Operator.
- [Quickstart ConnectCluster](/docs/guides/kafka/connectcluster/quickstart.md) with KubeDB Operator.
- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [ConnectCluster object](/docs/guides/kafka/concepts/connectcluster.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

