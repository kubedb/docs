---
title: Reconfigure Cassandra TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: cas-reconfigure-tls-cassandra
    name: Reconfigure Cassandra TLS/SSL Encryption
    parent: cas-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Cassandra TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Cassandra database via a CassandraOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Cassandra database

Here, We are going to create a Cassandra without TLS and then reconfigure the database to use TLS.

### Deploy Cassandra without TLS

In this section, we are going to deploy a Cassandra topology cluster without TLS. In the next few sections we will reconfigure TLS using `CassandraOpsRequest` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 3Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/cassandra.yaml
cassandra.kubedb.com/cassandra-prod created
```

Now, wait until `cassandra-prod` has status `Ready`. i.e,

```bash
$ kubectl get cas -n demo -w
NAME             TYPE                  VERSION   STATUS         AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   54s
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   84s
.
.
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready          2m8s

```

Now, we can try to access cqlsh of one cassandra pod without providing ssl flag and verify configuration that the TLS is disabled.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- cqlsh -u admin -p MkyikyIvjFEzzgB6
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> 
```

We can verify from the above output that TLS is disabled for this cluster.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Cassandra. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=cassandra/O=kubedb"
.....+................+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*..+...+...+......+......+........+......+....+..+.......+..+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*...............+.............+..+.+.....+...+.......+.....+....+...+........+....+.........+.....+.+.....+....+...........+.+..+...+...+.+...+...........+....+...+..............+..........+.....+...+..........+......+.........+...+..+...........................+.........+....+...........+....+.....+..................+...+.+...+.....+.+..+....+........+.............+...+..+....+...........+.......+......+...........+.+..+.......+........+...+............+.+.....+.+.....+.........+......+.+........+.+.....+.+....................+...+.......+...+......+...........+..........+............+.....+.+.....+.+.....+.+.........+........+...+....+.....+.........+.........+...+.......+.....+.......+........+.......+.....................+.....+....+...+...+...............+.....+...+....+..+...............+....+..+..........+.....+.......+...+.........+.........+..+...+.+...+..+.+.....+...+...........................+....+.....+...+......+.+...+...+............+..+...................+............+..+......+.+.....+......+.......+........+....+........+......+.+...........+...+.+...+............+......+..+..........+..+.+..+............+....+.........+..+.+............+.....+.......+...+...........+.+........................+......+...+.....+...+.......+..+................+.........+...+......+......+...........+.............+..............+.+...........+.+..+.......+.....+.........+......+...+.......+...+...........+....+.....+...+...+......+.+..+......+.......+..+.......+...+........+.......+..+....+.........+......+..+....+...+...........+..........+...........+.+...............+...............+..+......+...................+..+...+.......+...+.....+...+...+.......+...+......+...+.....+.......+.....+...............+.........+......+.........+....+..+...+.+..+.........+...+...+.............+..+...+..........+.....+..........+.........+..+.+.....+....+.........+..+...+....+......+..+.........+......+.......+...+...+..+.......+..+.........+.+.....+......+...+......+..........+.....+...............+..................+.+............+........+....+...+........+.+.....+.........+....+........+...+....+...+..............+.+...+......+...+......+............+.........+...+..+.+..+......+......+......+...+.+..............+.+...+...+........+....+.....+............+...+.+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
........+....+.........+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*..+....+......+......+..+......+.+........+......+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*..+.....+..................+...+....+...........+...+...................+......+...............+...........+....+......+........+...+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls cassandra-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/cassandra-ca created
```

Now, Let's create an `Issuer` using the `cassandra-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: cassandra-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: cassandra-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/cassandra-issuer.yaml
issuer.cert-manager.io/cas-issuer created
```

### Create CassandraOpsRequest

In order to add TLS to the cassandra, we have to create a `CassandraOpsRequest` CRO with our created issuer. Below is the YAML of the `CassandraOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    issuerRef:
      name: cassandra-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - cassandra
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on cassandra.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/cassandra/concepts/cassandra.md#spectls).

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/cassandra-add-tls.yaml
cassandraopsrequest.ops.kubedb.com/casops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME             TYPE             STATUS       AGE
casops-add-tls   ReconfigureTLS   Successful   3m34s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed.

```bash
$  kubectl describe cassandraopsrequest -n demo casops-add-tls 
Name:         casops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-29T10:45:39Z
  Generation:          1
  Resource Version:    89339
  UID:                 9099287b-41b6-4ff7-8a78-e383063f99df
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          cassandra
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       cassandra-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-07-29T10:45:39Z
    Message:               Cassandra ops-request has started to reconfigure tls for cassandra nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-07-29T10:45:55Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-07-29T10:45:50Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-07-29T10:45:50Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2025-07-29T10:45:50Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-07-29T10:46:03Z
    Message:               successfully reconciled the Cassandra with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-29T10:48:48Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-29T10:46:08Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T10:46:08Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T10:46:13Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-29T10:46:48Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T10:46:48Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T10:48:48Z
    Message:               Successfully completed reconfigureTLS for cassandra.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           3m52s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-add-tls
  Normal   Starting                                                           3m52s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         3m52s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-add-tls
  Warning  get certificate; ConditionStatus:True                              3m41s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m41s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m41s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                              3m41s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m41s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m41s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                  3m41s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                              3m36s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m36s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m36s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                              3m36s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m36s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m36s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                  3m36s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                      3m28s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    3m23s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  3m23s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 3m18s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    2m43s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  2m43s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    2m3s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  2m3s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    83s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  83s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       43s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           43s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         43s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-add-tls
```

Now, Let's try to access cqlsh of a cassandra pod without and with ssl flag and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- cqlsh -u admin -p MkyikyIvjFEzzgB6
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connection error: ('Unable to connect to any servers', {'127.0.0.1:9042': ConnectionShutdown('Connection to 127.0.0.1:9042 was closed')})
command terminated with exit code 1

$ kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- cqlsh -u admin -p MkyikyIvjFEzzgB6 --ssl
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> exit
```

We can see from the above output that, cqlsh is only accessable by using ssl flag which means that TLS is enabled.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- keytool -list -v -keystore /opt/cassandra/ssl/keystore.jks -storepass 'Yd33L.bUW(EdUCaV' | grep -E 'Valid from|Alias name'
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
Alias name: ca
Valid from: Tue Jul 29 10:40:48 GMT 2025 until: Wed Jul 29 10:40:48 GMT 2026
Alias name: certificate
Valid from: Tue Jul 29 10:45:45 GMT 2025 until: Mon Oct 27 10:45:45 GMT 2025
```

So, the certificate will expire on this time `Wed Jul 29 10:40:48 GMT 2026`.

### Create CassandraOpsRequest

Now we are going to increase it using a CassandraOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `cassandra-prod`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this cassandra cluster.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/casops-rotate.yaml
cassandraopsrequest.ops.kubedb.com/casops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME            TYPE             STATUS       AGE
casops-rotate   ReconfigureTLS   Successful   3m22s

```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed.

```bash
$kubectl describe cassandraopsrequest -n demo casops-rotate
Name:         casops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-29T11:06:02Z
  Generation:          1
  Resource Version:    91847
  UID:                 c6aa8d77-0a27-4a55-a48b-018345941b53
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-prod
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-07-29T11:06:02Z
    Message:               Cassandra ops-request has started to reconfigure tls for cassandra nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-07-29T11:06:15Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-07-29T11:06:10Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-07-29T11:06:10Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2025-07-29T11:06:10Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-07-29T11:06:25Z
    Message:               successfully reconciled the Cassandra with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-29T11:09:10Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-29T11:06:30Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T11:06:30Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T11:06:35Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-29T11:07:10Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T11:07:10Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T11:09:10Z
    Message:               Successfully completed reconfigureTLS for cassandra.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           3m44s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-rotate
  Normal   Starting                                                           3m44s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         3m44s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate
  Warning  get certificate; ConditionStatus:True                              3m36s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m36s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m36s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                              3m36s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m36s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m36s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                  3m36s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                              3m31s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m31s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m31s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                              3m31s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m31s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m31s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                  3m31s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                      3m21s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    3m16s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  3m16s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 3m11s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    2m36s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  2m36s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    116s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  116s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    76s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  76s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       36s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           36s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         36s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- keytool -list -v -keystore /opt/cassandra/ssl/keystore.jks -storepass 'Yd33L.bUW(EdUCaV' | grep -E 'Valid from|Alias name'

Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
Alias name: ca
Valid from: Tue Jul 29 10:40:48 GMT 2025 until: Wed Jul 29 10:40:48 GMT 2026
Alias name: certificate
Valid from: Tue Jul 29 11:09:11 GMT 2025 until: Mon Oct 27 11:09:11 GMT 2025
```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$  openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=cassandra-updated/O=kubedb-updated"
....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.......+.....+..........+...+...+..+...+....+............+...........+....+........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+..........+..+.......+...+.....+......+.......+...+..+....+.....+.............+..+.+.....+.......+..+.+...+....................+.........+...+..........+.......................+.....................+.+........+....+..+...+.......+.........+..+...+.+......+..+.............+........+......+......+.......+...........+.+.....+................+...+......+........+.......+...+........+...+....+.....+............+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
.....+.....+....+.....+...+....+........+.+..+.......+........+...+.......+........+......+.+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.....+...+........+.........+....+......+...+..+....+..+....+........+............+.+...+............+.........+.....+...+...+.........+.+...+..+.......+........+......................+.....+..........+...+..+......+.+.........+......+....................+.+...+.....+......+.+..............+...+.+..+....+.........+......+......+........+......+....+..+....+......+..+............+.+.................+...+....+...+............+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls cassandra-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/cassandra-new-ca created
```

Now, Let's create a new `Issuer` using the `cassandra-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: cas-new-issuer
  namespace: demo
spec:
  ca:
    secretName: cassandra-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/cassandra-new-issuer.yaml
issuer.cert-manager.io/cas-new-issuer created
```

### Create CassandraOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `CassandraOpsRequest` CRO with the newly created issuer. Below is the YAML of the `CassandraOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    issuerRef:
      name: cas-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cassandra.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/cassandra-update-tls-issuer.yaml
cassandrapsrequest.ops.kubedb.com/casops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$ kubectl get cassandraopsrequests -n demo casops-update-issuer 
NAME                   TYPE             STATUS       AGE
casops-update-issuer   ReconfigureTLS   Successful   3m44s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed.

```bash
$  kubectl describe cassandraopsrequest -n demo casops-update-issuer
Name:         casops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-29T11:17:26Z
  Generation:          1
  Resource Version:    93272
  UID:                 9df190c4-3a99-45b2-99f2-bf976eb62f36
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-prod
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       cas-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-07-29T11:17:26Z
    Message:               Cassandra ops-request has started to reconfigure tls for cassandra nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-07-29T11:17:39Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-07-29T11:17:34Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-07-29T11:17:34Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2025-07-29T11:17:34Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-07-29T11:17:47Z
    Message:               successfully reconciled the Cassandra with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-29T11:20:38Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-29T11:17:53Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T11:17:53Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T11:17:58Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-29T11:18:33Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T11:18:33Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T11:20:38Z
    Message:               Successfully completed reconfigureTLS for cassandra.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           4m10s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-update-issuer
  Normal   Starting                                                           4m10s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         4m10s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-update-issuer
  Warning  get certificate; ConditionStatus:True                              4m2s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        4m2s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            4m2s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                              4m2s   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        4m2s   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            4m2s   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                  4m2s   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                              3m57s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m57s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m57s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                              3m57s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                        3m57s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                            3m57s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                  3m57s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                      3m49s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    3m43s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  3m43s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 3m38s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    3m3s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  3m3s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    2m23s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  2m23s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    99s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  98s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       58s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           58s    KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         58s    KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-update-issuer
```

Now, Let's exec into a cassandra node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- keytool -list -v -keystore /opt/cassandra/ssl/keystore.jks -storepass 'Yd33L.bUW(EdUCaV' | grep 'Issuer'
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
Issuer: O=kubedb-updated, CN=cassandra-updated
Issuer: O=kubedb-updated, CN=cassandra-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a CassandraOpsRequest.

### Create CassandraOpsRequest

Below is the YAML of the `CassandraOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: casops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: cassandra-prod
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `cassandra-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Cassandra.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/reconfigure-tls/casops-remove.yaml
cassandraopsrequest.ops.kubedb.com/casops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CRO,

```bash
$  kubectl get cassandraopsrequest -n demo casops-remove
NAME            TYPE             STATUS       AGE
casops-remove   ReconfigureTLS   Successful   4m12s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed.

```bash
$  kubectl describe cassandraopsrequest -n demo casops-remove
Name:         casops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-29T11:24:47Z
  Generation:          1
  Resource Version:    94172
  UID:                 b4a3cc00-5b2e-4a0b-bc07-057aa8c02937
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  cassandra-prod
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-07-29T11:24:47Z
    Message:               Cassandra ops-request has started to reconfigure tls for cassandra nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-07-29T11:24:58Z
    Message:               successfully reconciled the Cassandra with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-29T11:27:43Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-07-29T11:25:03Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T11:25:03Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-29T11:25:08Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-29T11:25:43Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T11:25:43Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-29T11:27:43Z
    Message:               Successfully completed reconfigureTLS for cassandra.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           6m33s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/casops-remove
  Normal   Starting                                                           6m33s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         6m33s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-remove
  Normal   UpdatePetSets                                                      6m22s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    6m17s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  6m17s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 6m12s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    5m37s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  5m37s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    4m57s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  4m57s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    4m17s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  4m17s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartNodes                                                       3m37s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                           3m37s  KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         3m37s  KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: casops-remove
```

Now, Let's try to access cqlsh of one cassandra pod without providing ssl flag and verify configuration that the TLS is disabled.

```bash
$  kubectl exec -it -n demo cassandra-prod-rack-r0-0 -- cqlsh -u admin -p MkyikyIvjFEzzgB6
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> 

```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete opsrequest casops-add-tls casops-remove casops-rotate casops-update-issuer
kubectl delete cassandra -n demo cassandra-prod
kubectl delete issuer -n demo cas-issuer cas-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Different Cassandra topology clustering modes [here](/docs/guides/cassandra/clustering/_index.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Cassandra database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

