---
title: Reconfigure ClickHouse TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: ch-reconfigure-tls-clickhouse
    name: Reconfigure ClickHouse TLS/SSL Encryption
    parent: ch-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure ClickHouse TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing ClickHouse database via a ClickHouseOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/clickhouse](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/clickhouse) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a ClickHouse database

Here, We are going to create a ClickHouse without TLS and then reconfigure the database to use TLS.

### Deploy ClickHouse without TLS

In this section, we are going to deploy a ClickHouse topology cluster without TLS. In the next few sections we will reconfigure TLS using `ClickHouseOpsRequest` CRD. Below is the YAML of the `ClickHouse` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-prod
  namespace: demo
spec:
  version: 24.4.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      - name: appscode-cluster
        shards: 2
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: clickhouse
                resources:
                  limits:
                    memory: 4Gi
                  requests:
                    cpu: 500m
                    memory: 2Gi
            initContainers:
              - name: clickhouse-init
                resources:
                  limits:
                    memory: 1Gi
                  requests:
                    cpu: 500m
                    memory: 1Gi
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `ClickHouse` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-cluster.yaml
clickhouse.kubedb.com/clickhouse-prod created
```

Now, wait until `clickhouse-prod` has status `Ready`. i.e,

```bash
➤ kubectl get clickhouse -n demo -w
NAME              TYPE                  VERSION   STATUS         AGE
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   3s
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Provisioning   53s
.
.
clickhouse-prod   kubedb.com/v1alpha2   24.4.1    Ready          2m6s

```

Now, we can try to connect clickhouse tls port using openssl and verify configuration that the TLS is disabled.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ openssl s_client -connect localhost:9440
129329482888512:error:0200206F:system library:connect:Connection refused:../crypto/bio/b_sock2.c:110:
129329482888512:error:2008A067:BIO routines:BIO_connect:connect error:../crypto/bio/b_sock2.c:111:
129329482888512:error:0200206F:system library:connect:Connection refused:../crypto/bio/b_sock2.c:110:
129329482888512:error:2008A067:BIO routines:BIO_connect:connect error:../crypto/bio/b_sock2.c:111:
connect:errno=111
```

We can verify from the above output that TLS is disabled for this cluster.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in ClickHouse. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
➤ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=clickhouse/O=kubedb"
............+..+.+...+..........................+....+..+............+.+..+...+.+...+..+....+.....+.+......+...+..+...+....+.....+...+....+......+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*......+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.....+......+..+.+....................+....+......+..+...+......+.+..+.......+.....+...+.......+...+.....+.+..+...+.+......+........+..........+......+........+...+.........+...............+...+...+....+.........+......+...+......+..+.........+...+.+......+.....+.+...+..+...............+.......+...+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
....+.............+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*......+....+............+.....+....+.....+.+.....+....+...........+...+.......+.........+.....+....+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*..........+....+...........+.........+............+.+..+.+.....+.........+..........+..+.............+......+..+....+.........+.....+.........+......+.+...+.....+.+......+......+........+.......+........+......+.+...+.........+..+...+..........+..+....+..+...+.+........+.+..+....+......+........................+........+...+.......+..+................+...+.........+...........+....+...+...+..............+......+......+...+.+.....+.+......+..............+......+......................+........+...+.......+..............+.+..................+..+...+....+..+......+.........+.+.....+............+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
➤ kubectl create secret tls clickhouse-ca \
           --cert=ca.crt \
           --key=ca.key \
           --namespace=demo
secret/clickhouse-ca created
```

Now, Let's create an `Issuer` using the `clickhouse-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: clickhouse-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: clickhouse-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-issuer.yaml
issuer.cert-manager.io/clickhouse-ca-issuer created
```

### Create ClickHouseOpsRequest

In order to add TLS to the clickhouse, we have to create a `ClickHouseOpsRequest` CRO with our created issuer. Below is the YAML of the `ClickHouseOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: clickhouse-ca-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on clickhouse.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/clickhouse/concepts/clickhouse.md#spectls).

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-add-tls.yaml
clickhouseopsrequest.ops.kubedb.com/chops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CRO,

```bash
➤ kubectl get clickhouseopsrequests -n demo
NAME            TYPE             STATUS       AGE
chops-add-tls   ReconfigureTLS   Successful   4m17s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-add-tls 
Name:         chops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T05:48:41Z
  Generation:          1
  Resource Version:    767064
  UID:                 96f9ad99-6de9-411b-8853-06b4db6149bd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  10m
  Tls:
    Certificates:
      Alias:  server
      Dns Names:
        localhost
      Ip Addresses:
        127.0.0.1
      Subject:
        Organizations:
          kubedb:server
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       clickhouse-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-25T05:48:41Z
    Message:               ClickHouse ops-request has started to reconfigure tls for ClickHouse nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-25T05:48:54Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-08-25T05:48:49Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-08-25T05:48:49Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2025-08-25T05:48:49Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-08-25T05:48:59Z
    Message:               successfully reconciled the ClickHouse with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T05:48:59Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T05:51:14Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-25T05:49:04Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T05:49:04Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T05:49:09Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-25T05:49:34Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T05:49:34Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T05:50:14Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T05:50:14Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T05:50:34Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T05:50:34Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T05:51:14Z
    Message:               Successfully completed reconfigureTLS for ClickHouse.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             5m58s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-add-tls
  Normal   Starting                                                                             5m58s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           5m58s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-add-tls
  Warning  get certificate; ConditionStatus:True                                                5m50s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m50s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m50s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                                5m50s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m50s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m50s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                    5m50s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                                5m45s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m45s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m45s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                                5m45s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m45s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m45s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                    5m45s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  reconcile; ConditionStatus:True                                                      5m40s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      5m40s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      5m40s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        5m40s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    5m35s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  5m35s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   5m30s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    5m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  5m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    4m25s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  4m25s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    4m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  4m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         3m25s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                             3m25s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           3m25s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-add-tls
```

Now, Let's try to connect clickhouse tls port by using openssl and verify configuration that the TLS is enabled.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ openssl s_client -connect localhost:9440
CONNECTED(00000003)
Can't use SSL_get_servername
depth=1 CN = clickhouse, O = kubedb
verify error:num=19:self signed certificate in certificate chain
verify return:1
depth=1 CN = clickhouse, O = kubedb
verify return:1
depth=0 O = kubedb:server, CN = clickhouse-prod
verify return:1
```
We can see from the above output that, tls port is accessible by using openssl which means that TLS is enabled.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ openssl x509 -enddate -noout -in /etc/clickhouse-server/certs/server.crt
notAfter=Nov 23 06:13:38 2025 GMT
```

So, the certificate will expire on this time `Nov 23 06:13:38 2025 GMT`.

### Create ClickHouseOpsRequest

Now we are going to increase it using a ClickHouseOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    rotateCertificates: true
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `clickhouse-prod`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this clickhouse cluster.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-rotate-tls.yaml
clickhouseopsrequest.ops.kubedb.com/chops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CRO,

```bash
➤ kubectl get clickhouseopsrequest -n demo 
NAME           TYPE             STATUS       AGE
chops-rotate   ReconfigureTLS   Successful   3m48s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-rotate 
Name:         chops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T06:13:12Z
  Generation:          1
  Resource Version:    771370
  UID:                 a6731048-7d45-40f4-a524-ceff8d4b5a3c
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  10m
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-25T06:13:12Z
    Message:               ClickHouse ops-request has started to reconfigure tls for ClickHouse nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-25T06:13:22Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-08-25T06:13:17Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-08-25T06:13:17Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2025-08-25T06:13:17Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-08-25T06:13:27Z
    Message:               successfully reconciled the ClickHouse with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T06:13:27Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T06:16:13Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-25T06:13:33Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T06:13:33Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T06:13:38Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-25T06:14:13Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T06:14:13Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T06:14:53Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T06:14:53Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T06:15:33Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T06:15:33Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T06:16:13Z
    Message:               Successfully completed reconfigureTLS for ClickHouse.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             4m30s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-rotate
  Warning  get certificate; ConditionStatus:True                                                4m25s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                4m25s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              4m25s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                                4m25s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                4m25s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              4m25s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                    4m25s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                                4m20s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                4m20s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              4m20s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                                4m20s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                4m20s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              4m20s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                    4m20s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  reconcile; ConditionStatus:True                                                      4m15s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      4m15s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      4m15s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        4m15s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    4m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  4m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   4m4s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    3m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  3m29s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    2m49s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  2m49s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    2m9s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  2m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         89s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                             89s    KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           89s    KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ openssl x509 -enddate -noout -in /etc/clickhouse-server/certs/server.crt
notAfter=Nov 23 06:22:03 2025 GMT
```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$  openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=clickhouse-updated/O=kubedb-updated"
....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.......+.....+..........+...+...+..+...+....+............+...........+....+........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+..........+..+.......+...+.....+......+.......+...+..+....+.....+.............+..+.+.....+.......+..+.+...+....................+.........+...+..........+.......................+.....................+.+........+....+..+...+.......+.........+..+...+.+......+..+.............+........+......+......+.......+...........+.+.....+................+...+......+........+.......+...+........+...+....+.....+............+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
.....+.....+....+.....+...+....+........+.+..+.......+........+...+.......+........+......+.+..+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.....+...+........+.........+....+......+...+..+....+..+....+........+............+.+...+............+.........+.....+...+...+.........+.+...+..+.......+........+......................+.....+..........+...+..+......+.+.........+......+....................+.+...+.....+......+.+..............+...+.+..+....+.........+......+......+........+......+....+..+....+......+..+............+.+.................+...+....+...+............+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls clickhouse-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/clickhouse-new-ca created
```

Now, Let's create a new `Issuer` using the `clickhouse-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: ch-new-issuer
  namespace: demo
spec:
  ca:
    secretName: clickhouse-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-new-issuer.yaml
issuer.cert-manager.io/ch-new-issuer created
```

### Create ClickHouseOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `ClickHouseOpsRequest` CRO with the newly created issuer. Below is the YAML of the `ClickHouseOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    issuerRef:
      name: ch-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our clickhouse.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-update-issuer.yaml
clickhouseopsrequest.ops.kubedb.com/chops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CRO,

```bash
➤ kubectl get clickhouseopsrequests -n demo
NAME                  TYPE             STATUS       AGE
chops-update-issuer   ReconfigureTLS   Successful   4m49s
```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-update-issuer 
Name:         chops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T06:35:13Z
  Generation:          1
  Resource Version:    775305
  UID:                 be08c6b1-8d3f-42b9-ae27-a26c7e9807e8
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  10m
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       ch-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-25T06:35:13Z
    Message:               ClickHouse ops-request has started to reconfigure tls for ClickHouse nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-25T06:35:26Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-08-25T06:35:21Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-08-25T06:35:21Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2025-08-25T06:35:21Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2025-08-25T06:35:31Z
    Message:               successfully reconciled the ClickHouse with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T06:35:31Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T06:38:16Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-25T06:35:36Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T06:35:36Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T06:35:41Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-25T06:36:16Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T06:36:16Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T06:36:56Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T06:36:56Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T06:37:36Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T06:37:36Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T06:38:16Z
    Message:               Successfully completed reconfigureTLS for ClickHouse.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             5m29s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-update-issuer
  Normal   Starting                                                                             5m29s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           5m29s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-update-issuer
  Warning  get certificate; ConditionStatus:True                                                5m21s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m21s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m21s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                                5m21s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m21s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m21s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                    5m21s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                                5m16s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m16s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m16s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                                5m16s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                                                5m16s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                                              5m16s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                                    5m16s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  reconcile; ConditionStatus:True                                                      5m11s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      5m11s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      5m11s  KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        5m11s  KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    5m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  5m6s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   5m1s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    4m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  4m26s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    3m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  3m46s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    3m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  3m6s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         2m26s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                             2m26s  KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           2m26s  KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-update-issuer
```

Now, Let's exec into a clickhouse node and find out the ca subject to see if it matches the one we have provided.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ openssl x509 -in /etc/clickhouse-server/certs/server.crt -noout -issuer
issuer=CN = clickhouse-updated, O = kubedb-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a ClickHouseOpsRequest.

### Create ClickHouseOpsRequest

Below is the YAML of the `ClickHouseOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ClickHouseOpsRequest
metadata:
  name: chops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: clickhouse-prod
  tls:
    remove: true
  timeout: 10m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `clickhouse-prod` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on ClickHouse.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `ClickHouseOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/clickhouse/reconfigure-tls/clickhouse-remove-tls.yaml 
clickhouseopsrequest.ops.kubedb.com/chops-remove-tls created
```

#### Verify TLS Removed Successfully

Let's wait for `ClickHouseOpsRequest` to be `Successful`.  Run the following command to watch `ClickHouseOpsRequest` CRO,

```bash
➤ kubectl get clickhouseopsrequest -n demo chops-remove-tls 
NAME               TYPE             STATUS       AGE
chops-remove-tls   ReconfigureTLS   Successful   3m42s

```

We can see from the above output that the `ClickHouseOpsRequest` has succeeded. If we describe the `ClickHouseOpsRequest` we will get an overview of the steps that were followed.

```bash
➤ kubectl describe clickhouseopsrequest -n demo chops-remove-tls 
Name:         chops-remove-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ClickHouseOpsRequest
Metadata:
  Creation Timestamp:  2025-08-25T06:43:46Z
  Generation:          1
  Resource Version:    776585
  UID:                 e6802ac5-7207-4cee-9964-585eb96c9fdd
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   clickhouse-prod
  Timeout:  10m
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-08-25T06:43:46Z
    Message:               ClickHouse ops-request has started to reconfigure tls for ClickHouse nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-08-25T06:43:54Z
    Message:               successfully reconciled the ClickHouse with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-08-25T06:43:54Z
    Message:               reconcile; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  Reconcile
    Last Transition Time:  2025-08-25T06:46:34Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-08-25T06:43:59Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T06:43:59Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-0
    Last Transition Time:  2025-08-25T06:44:04Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-25T06:44:34Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T06:44:34Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-0-1
    Last Transition Time:  2025-08-25T06:45:14Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T06:45:14Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-0
    Last Transition Time:  2025-08-25T06:45:54Z
    Message:               get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T06:45:54Z
    Message:               evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--clickhouse-prod-appscode-cluster-shard-1-1
    Last Transition Time:  2025-08-25T06:46:34Z
    Message:               Successfully completed reconfigureTLS for ClickHouse.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                                               Age    From                         Message
  ----     ------                                                                               ----   ----                         -------
  Normal   Starting                                                                             4m14s  KubeDB Ops-manager Operator  Start processing for ClickHouseOpsRequest: demo/chops-remove-tls
  Normal   Starting                                                                             4m14s  KubeDB Ops-manager Operator  Pausing ClickHouse databse: demo/clickhouse-prod
  Normal   Successful                                                                           4m14s  KubeDB Ops-manager Operator  Successfully paused ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-remove-tls
  Warning  reconcile; ConditionStatus:True                                                      4m6s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      4m6s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Warning  reconcile; ConditionStatus:True                                                      4m6s   KubeDB Ops-manager Operator  reconcile; ConditionStatus:True
  Normal   UpdatePetSets                                                                        4m6s   KubeDB Ops-manager Operator  successfully reconciled the ClickHouse with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0    4m1s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0  4m1s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-0
  Warning  running pod; ConditionStatus:False                                                   3m56s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1    3m26s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1  3m26s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-0-1
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0    2m46s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0  2m46s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-0
  Warning  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1    2m6s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Warning  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1  2m6s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:clickhouse-prod-appscode-cluster-shard-1-1
  Normal   RestartNodes                                                                         86s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                                             86s    KubeDB Ops-manager Operator  Resuming ClickHouse database: demo/clickhouse-prod
  Normal   Successful                                                                           86s    KubeDB Ops-manager Operator  Successfully resumed ClickHouse database: demo/clickhouse-prod for ClickHouseOpsRequest: chops-remove-tls
```

Now, Let's try to connect clickhouse tls port using openssl and verify configuration that the TLS is disabled.

```bash
➤ kubectl exec -it -n demo clickhouse-prod-appscode-cluster-shard-0-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-prod-appscode-cluster-shard-0-0:/$ openssl s_client -connect localhost:9440
128501305447744:error:0200206F:system library:connect:Connection refused:../crypto/bio/b_sock2.c:110:
128501305447744:error:2008A067:BIO routines:BIO_connect:connect error:../crypto/bio/b_sock2.c:111:
128501305447744:error:0200206F:system library:connect:Connection refused:../crypto/bio/b_sock2.c:110:
128501305447744:error:2008A067:BIO routines:BIO_connect:connect error:../crypto/bio/b_sock2.c:111:
connect:errno=111 
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete chops-add-tls chops-remove-tls chops-rotate chops-update-issuer
kubectl delete clickhouse -n demo clickhouse-prod
kubectl delete issuer -n demo clickhouse-ca-issuer ch-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ClickHouse object](/docs/guides/clickhouse/concepts/clickhouse.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

