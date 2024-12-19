---
title: Reconfigure Kafka TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: kf-reconfigure-tls-kafka
    name: Reconfigure Kafka TLS/SSL Encryption
    parent: kf-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Kafka TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Kafka database via a KafkaOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/kafka](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/kafka) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Kafka database

Here, We are going to create a Kafka without TLS and then reconfigure the database to use TLS.

### Deploy Kafka without TLS

In this section, we are going to deploy a Kafka topology cluster without TLS. In the next few sections we will reconfigure TLS using `KafkaOpsRequest` CRD. Below is the YAML of the `Kafka` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kafka.yaml
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

Now, we can exec one kafka broker pod and verify configuration that the TLS is disabled.

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'ssl.keystore'
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=null sensitive=false synonyms={}
  ssl.keystore.password=null sensitive=true synonyms={}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=null sensitive=false synonyms={}
  ssl.keystore.password=null sensitive=true synonyms={}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
```

We can verify from the above output that TLS is disabled for this cluster.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Kafka. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls kafka-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/kafka-ca created
```

Now, Let's create an `Issuer` using the `kafka-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: kf-issuer
  namespace: demo
spec:
  ca:
    secretName: kafka-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kafka-issuer.yaml
issuer.cert-manager.io/kf-issuer created
```

### Create KafkaOpsRequest

In order to add TLS to the kafka, we have to create a `KafkaOpsRequest` CRO with our created issuer. Below is the YAML of the `KafkaOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    issuerRef:
      name: kf-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - kafka
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on kafka.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/kafka/concepts/kafka.md#spectls).

Let's create the `KafkaOpsRequest` CR we have shown above,

> **Note:** For combined kafka, you just need to refer kafka combined object in `databaseRef` field. To learn more about combined kafka, please visit [here](/docs/guides/kafka/clustering/combined-cluster/index.md).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kafka-add-tls.yaml
kafkaopsrequest.ops.kubedb.com/kfops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CRO,

```bash
$ kubectl get kafkaopsrequest -n demo
NAME            TYPE             STATUS       AGE
kfops-add-tls   ReconfigureTLS   Successful   4m36s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-add-tls 
Name:         kfops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-07-31T06:36:27Z
  Generation:          1
  Resource Version:    158448
  UID:                 9c95ef81-2db8-4740-9708-60618ab57db5
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   kafka-prod
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          kafka
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       kf-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-31T06:36:27Z
    Message:               Kafka ops-request has started to reconfigure tls for kafka nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-31T06:36:36Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-07-31T06:36:36Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-07-31T06:36:36Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-07-31T06:36:37Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-07-31T06:38:45Z
    Message:               successfully reconciled the Kafka with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-31T06:38:50Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T06:38:50Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T06:39:06Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T06:39:10Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T06:39:10Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T06:39:25Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T06:39:30Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T06:39:35Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T06:39:45Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T06:39:50Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T06:39:50Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T06:40:05Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T06:40:10Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-07-31T06:40:11Z
    Message:               Successfully completed reconfigureTLS for kafka.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   4m59s  KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-add-tls
  Normal   Starting                                                                   4m59s  KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 4m59s  KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-add-tls
  Warning  get certificate; ConditionStatus:True                                      4m51s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m50s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m50s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                      4m49s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m49s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m49s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                          4m49s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                      4m44s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m44s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m44s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                      4m43s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m43s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m43s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                          4m43s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                              2m41s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             2m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           2m36s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  2m31s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   2m21s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             2m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           2m16s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  2m11s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   2m1s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 116s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:False; PodName:kafka-prod-broker-0              116s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 111s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               111s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      106s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       101s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 96s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               96s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      91s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       81s    KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               76s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   76s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 76s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-add-tls
```

Now, Let's exec into a kafka broker pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'ssl.keystore'
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=/var/private/ssl/server.keystore.jks sensitive=false synonyms={STATIC_BROKER_CONFIG:ssl.keystore.location=/var/private/ssl/server.keystore.jks}
  ssl.keystore.password=null sensitive=true synonyms={STATIC_BROKER_CONFIG:ssl.keystore.password=null}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=/var/private/ssl/server.keystore.jks sensitive=false synonyms={STATIC_BROKER_CONFIG:ssl.keystore.location=/var/private/ssl/server.keystore.jks}
  ssl.keystore.password=null sensitive=true synonyms={STATIC_BROKER_CONFIG:ssl.keystore.password=null}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
```

We can see from the above output that, keystore location is `/var/private/ssl/server.keystore.jks` which means that TLS is enabled.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
$ $ kubectl exec -it -n demo kafka-prod-broker-0 -- keytool -list -v -keystore /var/private/ssl/server.keystore.jks -storepass wt6f5pwxpg84 | grep -E 'Valid from|Alias name'
Alias name: ca
Valid from: Wed Jul 31 06:11:30 UTC 2024 until: Thu Jul 31 06:11:30 UTC 2025
Alias name: certificate
Valid from: Wed Jul 31 06:36:31 UTC 2024 until: Tue Oct 29 06:36:31 UTC 2024
```

So, the certificate will expire on this time `Tue Oct 29 06:36:31 UTC 2024`.

### Create KafkaOpsRequest

Now we are going to increase it using a KafkaOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `kafka-prod`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this kafka cluster.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kfops-rotate.yaml
kafkaopsrequest.ops.kubedb.com/kfops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CRO,

```bash
$ kubectl get kafkaopsrequests -n demo kfops-rotate
NAME            TYPE             STATUS       AGE
kfops-rotate    ReconfigureTLS   Successful   4m4s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-rotate
Name:         kfops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-07-31T07:02:10Z
  Generation:          1
  Resource Version:    161186
  UID:                 d1e6f412-3771-4963-8384-2c31bab3a057
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-31T07:02:10Z
    Message:               Kafka ops-request has started to reconfigure tls for kafka nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-31T07:02:18Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-07-31T07:02:18Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-07-31T07:02:18Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-07-31T07:02:18Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-07-31T07:03:59Z
    Message:               successfully reconciled the Kafka with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-31T07:04:05Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T07:04:05Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T07:04:20Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T07:04:25Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T07:04:25Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T07:04:40Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T07:04:45Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T07:04:45Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T07:05:20Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T07:05:25Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T07:05:25Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T07:05:35Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T07:05:40Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-07-31T07:05:40Z
    Message:               Successfully completed reconfigureTLS for kafka.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                     Age    From                         Message
  ----     ------                                                                     ----   ----                         -------
  Normal   Starting                                                                   5m7s   KubeDB Ops-manager Operator  Start processing for KafkaOpsRequest: demo/kfops-rotate
  Normal   Starting                                                                   5m7s   KubeDB Ops-manager Operator  Pausing Kafka databse: demo/kafka-prod
  Normal   Successful                                                                 5m7s   KubeDB Ops-manager Operator  Successfully paused Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-rotate
  Warning  get certificate; ConditionStatus:True                                      4m59s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m59s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m59s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                      4m59s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m59s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m59s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                          4m59s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                      4m53s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m53s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m53s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                      4m53s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                                4m53s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                    4m53s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                          4m53s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                              3m18s  KubeDB Ops-manager Operator  successfully reconciled the Kafka with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0             3m12s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0           3m12s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0  3m7s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0   2m57s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1             2m52s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1           2m52s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1  2m47s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-controller-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1   2m37s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0                 2m32s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0               2m32s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0      2m27s  KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-0
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0       117s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
  Warning  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1                 112s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1               112s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1      107s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:False; PodName:kafka-prod-broker-1
  Warning  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1       102s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
  Normal   RestartNodes                                                               97s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                   97s    KubeDB Ops-manager Operator  Resuming Kafka database: demo/kafka-prod
  Normal   Successful                                                                 97s    KubeDB Ops-manager Operator  Successfully resumed Kafka database: demo/kafka-prod for KafkaOpsRequest: kfops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo kafka-prod-broker-0 -- keytool -list -v -keystore /var/private/ssl/server.keystore.jks -storepass wt6f5pwxpg84 | grep -E 'Valid from|Alias name'
Alias name: ca
Valid from: Wed Jul 31 06:11:30 UTC 2024 until: Thu Jul 31 06:11:30 UTC 2025
Alias name: certificate
Valid from: Wed Jul 31 07:05:40 UTC 2024 until: Tue Oct 29 07:05:40 UTC 2024
```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls kafka-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/kafka-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: kf-new-issuer
  namespace: demo
spec:
  ca:
    secretName: kafka-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kafka-new-issuer.yaml
issuer.cert-manager.io/kf-new-issuer created
```

### Create KafkaOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `KafkaOpsRequest` CRO with the newly created issuer. Below is the YAML of the `KafkaOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    issuerRef:
      name: kf-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our kafka.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kafka-update-tls-issuer.yaml
kafkapsrequest.ops.kubedb.com/kfops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CRO,

```bash
$ kubectl get kafkaopsrequests -n demo kfops-update-issuer
NAME                  TYPE             STATUS       AGE
kfops-update-issuer   ReconfigureTLS   Successful   8m6s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-update-issuer
Name:         kfops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-07-31T07:33:37Z
  Generation:          1
  Resource Version:    163574
  UID:                 d81c7a63-199b-4c45-b9c0-a4a93fed3c10
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       kf-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-31T07:33:37Z
    Message:               Kafka ops-request has started to reconfigure tls for kafka nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-31T07:33:43Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-07-31T07:33:43Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-07-31T07:33:44Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-07-31T07:33:44Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-07-31T07:35:49Z
    Message:               successfully reconciled the Kafka with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-31T07:35:54Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T07:35:54Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T07:36:09Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T07:36:14Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T07:36:14Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T07:36:34Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T07:36:39Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T07:36:39Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T07:37:19Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T07:37:24Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T07:37:24Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T07:38:04Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T07:38:09Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-07-31T07:38:09Z
    Message:               Successfully completed reconfigureTLS for kafka.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now, Let's exec into a kafka node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it kafka-prod-broker-0 -- bash
kafka@kafka-prod-broker-0:~$ keytool -list -v -keystore /var/private/ssl/server.keystore.jks -storepass wt6f5pwxpg84 | grep 'Issuer'
Issuer: O=kubedb-updated, CN=ca-updated
Issuer: O=kubedb-updated, CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a KafkaOpsRequest.

### Create KafkaOpsRequest

Below is the YAML of the `KafkaOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: KafkaOpsRequest
metadata:
  name: kfops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: kafka-prod
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `kafka-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Kafka.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `KafkaOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/kafka/reconfigure-tls/kfops-remove.yaml
kafkaopsrequest.ops.kubedb.com/kfops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `KafkaOpsRequest` to be `Successful`.  Run the following command to watch `KafkaOpsRequest` CRO,

```bash
$ kubectl get kafkaopsrequest -n demo kfops-remove
NAME           TYPE             STATUS        AGE
kfops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `KafkaOpsRequest` has succeeded. If we describe the `KafkaOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe kafkaopsrequest -n demo kfops-remove
Name:         kfops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         KafkaOpsRequest
Metadata:
  Creation Timestamp:  2024-07-31T09:34:09Z
  Generation:          1
  Resource Version:    171329
  UID:                 c21b5c15-8fc0-43b5-9b46-6d1a98c9422d
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  kafka-prod
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-07-31T09:34:09Z
    Message:               Kafka ops-request has started to reconfigure tls for kafka nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-07-31T09:34:17Z
    Message:               successfully reconciled the Kafka with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-07-31T09:34:22Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T09:34:22Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T09:34:32Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-0
    Last Transition Time:  2024-07-31T09:34:37Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T09:34:37Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T09:34:47Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-controller-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-controller-1
    Last Transition Time:  2024-07-31T09:34:52Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T09:34:52Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T09:35:32Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-0
    Last Transition Time:  2024-07-31T09:35:37Z
    Message:               get pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T09:35:37Z
    Message:               evict pod; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T09:38:47Z
    Message:               check pod running; ConditionStatus:True; PodName:kafka-prod-broker-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--kafka-prod-broker-1
    Last Transition Time:  2024-07-31T09:38:52Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-07-31T09:38:52Z
    Message:               Successfully completed reconfigureTLS for kafka.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now, Let's exec into one of the broker node and find out that TLS is disabled or not.

```bash
$$ kubectl exec -it -n demo kafka-prod-broker-0 -- kafka-configs.sh --bootstrap-server localhost:9092 --command-config /opt/kafka/config/clientauth.properties --describe --entity-type brokers --all | grep 'ssl.keystore'
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=null sensitive=false synonyms={}
  ssl.keystore.password=null sensitive=true synonyms={}
  ssl.keystore.type=JKS sensitive=false synonyms={DEFAULT_CONFIG:ssl.keystore.type=JKS}
  ssl.keystore.certificate.chain=null sensitive=true synonyms={}
  ssl.keystore.key=null sensitive=true synonyms={}
  ssl.keystore.location=null sensitive=false synonyms={}
  ssl.keystore.password=null sensitive=true synonyms={}
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete opsrequest kfops-add-tls kfops-remove kfops-rotate kfops-update-issuer
kubectl delete kafka -n demo kafka-prod
kubectl delete issuer -n demo kf-issuer kf-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Kafka object](/docs/guides/kafka/concepts/kafka.md).
- Different Kafka topology clustering modes [here](/docs/guides/kafka/clustering/_index.md).
- Monitor your Kafka database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/kafka/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Kafka database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/kafka/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

