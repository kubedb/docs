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
NAME                TYPE                  VERSION   STATUS   AGE
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Provisioning   2s
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Provisioning   4s
.
.
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Ready          112s
```

Exec one of the broker pods and run the following command to create a tiered storage enabled topic and insert some data into it:

```bash
$ kubectl exec -n demo -it kafka-prod-tiered-broker-0 -- bash
root@kafka-prod-tiered-broker-0:/# kafka-topics.sh --bootstrap-server localhost:9092 --create --config remote.storage.enable=true --config retention.ms=-1 --config segment.bytes=1048576 \            --config retention.bytes=104857600 --config local.retention.bytes=1 --partitions 1 --replication-factor 1 --topic topic1 --command-config config/clientauth.properties
topic1 created
root@kafka-prod-tiered-broker-0:/#  kafka-producer-perf-test.sh --producer-props bootstrap.servers=localhost:9092 --topic topic1  --num-records 10000 --record-size 512 --throughput 1000 --producer.config config/clientauth.properties
4998 records sent, 999.2 records/sec (0.49 MB/sec), 13.5 ms avg latency, 526.0 ms max latency.
10000 records sent, 999.3 records/sec (0.49 MB/sec), 8.51 ms avg latency, 526.00 ms max latency, 4 ms 50th, 50 ms 95th, 92 ms 99th, 92 ms 99.9th.
```

here, we created a topic with `local.retention.bytes=1` which will force kafka to offload segments to the remote tiered storage as soon as possible. You can check the S3 bucket to see the offloaded segments.

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
NAME                TYPE                  VERSION   STATUS   AGE
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Provisioning   2s
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Provisioning   4s
.
.
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Ready          112s
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
NAME                TYPE                  VERSION   STATUS   AGE
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Provisioning   2s
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Provisioning   4s
.
.
kafka-prod-tiered   kubedb.com/v1alpha2   4.0.0     Ready          112s
```

## Next Steps

- Use [kubedb cli](/docs/guides/kafka/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
