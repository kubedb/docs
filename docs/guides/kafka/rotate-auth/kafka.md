---
title: Kafka Rotate Authentication Guide
menu:
  docs_{{ .version }}:
    identifier: kf-rotate-auth-kafka
    name: Kafka RotateAuth Guide
    parent: kf-rotate-auth
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Kafka Authentication

KubeDB supports rotating Authentication for existing Kafka via a KafkaOpsRequest. There are two ways to do that.
1. **Operator Generated**: User will not provide any secret. KubeDB operator will generate a random password and update the existing secret with that password.
2. **User Defined**: User can create a `kubernetes.io/basic-auth` type secret with `username` and  `password` and refers this to `KafkaOpsRequest`.

This tutorial will show you how to use KubeDB to rotate authentication credentials.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Create Kafka with Enabling Authentication

In this section, we are going to deploy a Kafka topology cluster with authentication enabled. In the next few sections we will rotate the authentication using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Kafka
metadata:
  name: kafka-prod
  namespace: demo
spec:
  version: 3.9.0
  topology:
    broker:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    controller:
      replicas: 2
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

Let's create the `Kafka` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/rotate-auth/kafka-prod.yaml
kafka.kubedb.com/kafka-prod created
```

Now, wait until `kafka-prod` has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w
NAME          TYPE            VERSION   STATUS         AGE
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   0s
kafka-prod    kubedb.com/v1   3.9.0     Provisioning   9s
.
.
kafka-prod    kubedb.com/v1   3.9.0     Ready          2m10s
```

Now, we can exec one kafka broker pod and verify configuration that authentication is enabled.

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep sasl.enabled.mechanism

  listener.name.local.sasl.enabled.mechanisms=PLAIN sensitive=false synonyms={STATIC_BROKER_CONFIG:listener.name.local.sasl.enabled.mechanisms=PLAIN, STATIC_BROKER_CONFIG:sasl.enabled.mechanisms=PLAIN,SCRAM-SHA-256, DEFAULT_CONFIG:sasl.enabled.mechanisms=GSSAPI}
  sasl.enabled.mechanisms=PLAIN,SCRAM-SHA-256 sensitive=false synonyms={STATIC_BROKER_CONFIG:sasl.enabled.mechanisms=PLAIN,SCRAM-SHA-256, DEFAULT_CONFIG:sasl.enabled.mechanisms=GSSAPI}
  listener.name.local.sasl.enabled.mechanisms=PLAIN sensitive=false synonyms={STATIC_BROKER_CONFIG:listener.name.local.sasl.enabled.mechanisms=PLAIN, STATIC_BROKER_CONFIG:sasl.enabled.mechanisms=PLAIN,SCRAM-SHA-256, DEFAULT_CONFIG:sasl.enabled.mechanisms=GSSAPI}
  sasl.enabled.mechanisms=PLAIN,SCRAM-SHA-256 sensitive=false synonyms={STATIC_BROKER_CONFIG:sasl.enabled.mechanisms=PLAIN,SCRAM-SHA-256, DEFAULT_CONFIG:sasl.enabled.mechanisms=GSSAPI}
```

We can verify from the above output that authentication is enabled for this cluster. By default, KubeDB operator create default credentials for the Kafka cluster. The default credentials are stored in a secret named `<kafka-name>-auth` in the same namespace as the Kafka cluster. You can find the secret by running the following command:

```bash
$ kubectl get kf -n demo kafka-prod -ojson | jq .spec.authSecret.name
"kafka-prod-auth"

$ kubectl get secret -n demo kafka-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin

$ kubectl get secret -n demo kafka-prod-auth -o=jsonpath='{.data.password}' | base64 -d
zvrFXkStB~9A!NTC
```

You will find a new field `.spec.authSecret.activeFrom` in the `Kafka` CR. This field is used to track the active credentials. The value of this field is set the time when the secret (`.spec.authSecret.name`) is active for kafka cluster. The value of this field is updated when the authentication is rotated.

```bash
$ kubectl get kf -n demo kafka-prod -ojsonpath='{.spec.authSecret.activeFrom}'
2025-04-03T08:42:05Z
```

> **Note:** There is another field `.spec.authSecret.rotateAfter` in the `Kafka` CR. This field is used to track the time when the authentication will be rotated. When a user set this field, Recommendation Engine will generate a recommendation `RotateAuth` Ops Request after this time from `.spec.authSecret.activeFrom`(i.e. `activeFrom + rotateAfter`). You need `Recommendation Engine` to be installed in order to use this feature.

### Create RotateAuth KafkaOpsRequest

#### 1. Using operator generated credentials:

In order to rotate authentication to the kafka using operator generated, we have to create a `KafkaOpsRequest` CRO with `RotateAuth` type. Below is the YAML of the `KafkaOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: kafka-prod
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on kafka.

Let's create the `KafkaOpsRequest` CR we have shown above,

> **Note:** For combined kafka, you just need to refer kafka combined object in `databaseRef` field. To learn more about combined kafka, please visit [here](/docs/guides/kafka/clustering/combined-cluster/index.md).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/rotate-auth/kafka-rotate-auth-generated.yaml
kafkaopsrequest.ops.kubedb.com/kfops-rotate-auth-generated created
```

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CRO,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                          TYPE         STATUS       AGE
kfops-rotate-auth-generated   RotateAuth   Successful   3m18s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-rotate-auth-generated 
Name:         kfops-rotate-auth-generated
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2025-05-15T11:11:04Z
  Generation:          1
  Resource Version:    290550
  UID:                 71ff7cec-f895-424c-b14f-9b957ccf9ccd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-05-15T11:11:04Z
    Message:               Kafka ops-request has started to rotate auth for kafka nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-05-15T11:11:07Z
    Message:               Successfully generated new credentials
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-05-15T11:11:15Z
    Message:               successfully reconciled the Kafka with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-05-15T11:11:20Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2025-05-15T11:11:20Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2025-05-15T11:11:55Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2025-05-15T11:12:00Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2025-05-15T11:12:00Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2025-05-15T11:12:35Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2025-05-15T11:12:40Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2025-05-15T11:12:40Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2025-05-15T11:13:15Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2025-05-15T11:13:20Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2025-05-15T11:13:20Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2025-05-15T11:13:55Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2025-05-15T11:14:00Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-05-15T11:14:00Z
    Message:               Successfully completed reconfigure kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m51s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-rotate-auth-generated
  Normal   Starting                                                                   3m51s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 3m51s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-rotate-auth-generated
  Normal   UpdatePetSets                                                              3m40s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             3m35s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           3m35s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  3m30s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   3m     KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             2m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           2m55s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  2m50s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   2m20s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 2m15s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               2m15s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      2m10s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       100s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 95s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               95s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      90s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       60s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               55s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   55s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 55s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-rotate-auth-generated
```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$ kubectl get kf -n demo kafka-prod -ojson | jq .spec.authSecret.name
"kafka-prod-auth"

$ kubectl get secret -n demo kafka-prod-auth -o=jsonpath='{.data.username}' | base64 -d
admin

$ kubectl get secret -n demo kafka-prod-auth -o=jsonpath='{.data.password}' | base64 -d
al9jY2xvYW5pbmc=
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo kafka-prod-auth -o=jsonpath='{.data.username.prev}' | base64 -d
admin
$ kubectl get secret -n demo kafka-prod-auth -o=jsonpath='{.data.password.prev}' | base64 -d
zvrFXkStB~9A!NTC
```

The above output shows that the password has been changed successfully. The previous username & password is stored for rollback purpose.

#### 2. Using user created credentials

At first, we need to create a secret with `kubernetes.io/basic-auth` type using custom `username` and `password`. Below is the command to create a secret with `kubernetes.io/basic-auth` type,

```bash
$ kubectl create secret generic kafka-user-auth -n demo \
          --type=kubernetes.io/basic-auth \
          --from-literal=username=kafka \
          --from-literal=password=kafka-secret
secret/kafka-user-auth created
```

Now create a Kafka Ops Request with `RotateAuth` type. Below is the YAML of the `KafkaOpsRequest` that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: kafka-prod
  authentication:
    secretRef:
      name: kafka-user-auth
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing rotate authentication operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `RotateAuth` on kafka.
- `spec.authentication.secretRef.name` specifies that we are using `kafka-user-auth` secret for authentication.

Let's create the `KafkaOpsRequest` CR we have shown above,

> **Note:** For combined kafka, you just need to refer kafka combined object in `databaseRef` field. To learn more about combined kafka, please visit [here](/docs/guides/kafka/clustering/combined-cluster/index.md).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/rotate-auth/kafka-rotate-auth-user.yaml
kafkaopsrequest.ops.kubedb.com/kfops-rotate-auth-user created
```

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CRO,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME                          TYPE         STATUS       AGE
kfops-rotate-auth-generated   RotateAuth   Successful   83m
kfops-rotate-auth-user        RotateAuth   Successful   2m58s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-rotate-auth-user 
Name:         kfops-rotate-auth-user
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2025-05-15T12:31:13Z
  Generation:          1
  Resource Version:    310786
  UID:                 13513a65-ac25-4667-8a11-80e356500c53
Spec:
  Apply:  IfReady
  Authentication:
    Secret Ref:
      Name:  kafka-user-auth
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Type:     RotateAuth
Status:
  Conditions:
    Last Transition Time:  2025-05-15T12:31:13Z
    Message:               Kafka ops-request has started to rotate auth for kafka nodes
    Observed Generation:   1
    Reason:                RotateAuth
    Status:                True
    Type:                  RotateAuth
    Last Transition Time:  2025-05-15T12:31:16Z
    Message:               Successfully referenced the user provided authSecret
    Observed Generation:   1
    Reason:                UpdateCredential
    Status:                True
    Type:                  UpdateCredential
    Last Transition Time:  2025-05-15T12:31:24Z
    Message:               successfully reconciled the Kafka with new auth credentials and configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-05-15T12:31:29Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2025-05-15T12:31:29Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2025-05-15T12:32:04Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2025-05-15T12:32:09Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2025-05-15T12:32:09Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2025-05-15T12:32:44Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2025-05-15T12:32:49Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2025-05-15T12:32:49Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2025-05-15T12:33:24Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2025-05-15T12:33:29Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2025-05-15T12:33:29Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2025-05-15T12:34:04Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2025-05-15T12:34:09Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-05-15T12:34:09Z
    Message:               Successfully completed reconfigure kafka
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   3m17s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-rotate-auth-user
  Normal   Starting                                                                   3m17s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 3m17s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-rotate-auth-user
  Normal   UpdatePetSets                                                              3m6s   KubeDB Ops-manager Operator  successfully reconciled the Kafka with new auth credentials and configuration
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             3m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           3m1s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  2m56s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   2m26s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             2m21s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           2m21s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  2m16s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   106s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 101s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               101s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      96s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       66s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 61s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               61s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      56s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       26s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               21s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   21s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 21s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-rotate-auth-user
```

#### Verify Password is changed

Now, We can verify that the password has been changed. You can find the secret and its data by running the following command:

```bash
$ kubectl get kf -n demo kafka-prod -ojson | jq .spec.authSecret.name
"kafka-user-auth"

$ kubectl get secret -n demo kafka-user-auth -o=jsonpath='{.data.username}' | base64 -d
kafka

$ kubectl get secret -n demo kafka-user-auth -o=jsonpath='{.data.password}' | base64 -d
kafka-secret
```

Also, there will be two more new keys in the secret that stores the previous credentials. The keys are `username.prev` and `password.prev`. You can find the secret and its data by running the following command:

```bash
$ kubectl get secret -n demo kafka-user-auth -o=jsonpath='{.data.username.prev}' | base64 -d
admin
$ kubectl get secret -n demo kafka-user-auth -o=jsonpath='{.data.password.prev}' | base64 -d
al9jY2xvYW5pbmc=
```

The above output shows that the password has been changed successfully. The previous username & password is stored in the secret for rollback purpose.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete kafkaopsrequest -n demo kfops-rotate-auth-generated kfops-rotate-auth-user
kubectl delete kafka -n demo kafka-prod
kubectl delete secret -n demo kafka-user-auth
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).
- Kafka ConnectCluster with KubeDB [here](/docs/guides/kafka/connectcluster/quickstart.md).
- Kafka Schema Registry with KubeDB [here](/docs/guides/kafka/schemaregistry/overview.md).
- Kafka RestProxy with KubeDB [here](/docs/guides/kafka/restproxy/overview.md).
- Kafka Migration with KubeDB [here](/docs/guides/kafka/migration/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

