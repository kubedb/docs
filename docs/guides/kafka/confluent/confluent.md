---
title: Confluent Enterprise Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-confluent-kafka-docs
    name: Overview
    parent: kf-confluent-kafka
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Confluent Enterprise Kafka

KubeDB can run [Confluent Server](https://docs.confluent.io/platform/current/installation/docker/image-reference.html) (Confluent's Enterprise Kafka distribution) instead of KubeDB's own Apache Kafka based image. This tutorial will show you how to deploy a Kafka cluster on the `Confluent` distribution and provide it with an enterprise license.

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

> Note: YAML files used in this tutorial are stored in [examples/kafka/confluent/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka/confluent) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Confluent distribution KafkaVersion

Running Confluent Server requires a `KafkaVersion` catalog object whose `spec.distribution` is set to `Confluent`, pointing `spec.db.image` at Confluent's own image directly (KubeDB doesn't rehost Confluent Server, since its license is commercial and doesn't permit redistribution). KubeDB ships a `confluent-8.3.2` `KafkaVersion` by default:

```bash
$ kubectl get kafkaversion confluent-8.3.2 -o yaml
```

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: KafkaVersion
metadata:
  name: confluent-8.3.2
spec:
  distribution: Confluent
  db:
    image: docker.io/confluentinc/cp-server:8.3.2
  version: 8.3.2
  ...
```

See the [KafkaVersion concept](/docs/guides/kafka/concepts/kafkaversion.md#specdistribution) page for details.

## Create License Secret

Confluent Server runs on a 30-day trial without a license, then stops working. Create a Secret containing your enterprise license key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kafka-confluent-license
  namespace: demo
stringData:
  license: <your-confluent-enterprise-license-key>
```

Apply the secret:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/confluent/kafka-confluent-license.yaml
secret/kafka-confluent-license created
```

## Create a Confluent Enterprise Kafka

Here is an example `Kafka` CR that uses the `confluent-8.3.2` `KafkaVersion` and the license Secret created above:

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-confluent
  namespace: demo
spec:
  version: confluent-8.3.2
  license:
    secretName: kafka-confluent-license
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is set to `confluent-8.3.2`, the `Confluent` distribution `KafkaVersion`.
- `spec.license.secretName` points at the Secret holding the enterprise license key. See the [Kafka concept](/docs/guides/kafka/concepts/kafka.md#speclicense) page for details.

Apply the yaml:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/confluent/kafka-confluent.yaml
kafka.kubedb.com/kafka-confluent created
```

KubeDB operator watches for `Kafka` objects using Kubernetes API. When a `Kafka` object is created, KubeDB operator will create a new PetSet running Confluent Server, along with the other Kubernetes resources like Service, Secret etc. required to run the Kafka cluster.

```bash
$ kubectl get kafka -n demo -w
NAME               TYPE            VERSION           STATUS         AGE
kafka-confluent    kubedb.com/v1   confluent-8.3.2   Provisioning   2s
kafka-confluent    kubedb.com/v1   confluent-8.3.2   Provisioning   4s
.
.
kafka-confluent    kubedb.com/v1   confluent-8.3.2   Ready          112s
```

## Cleanup

To clean up the resources created by this tutorial, run the following commands:

```bash
$ kubectl delete -n demo kafka kafka-confluent
$ kubectl delete -n demo secret kafka-confluent-license
$ kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Detail concepts of [KafkaVersion object](/docs/guides/kafka/concepts/kafkaversion.md).
- Deploy your first Kafka database with KubeDB by following the guide [here](/docs/guides/kafka/quickstart/kafka/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
