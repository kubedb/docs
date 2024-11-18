---
title: Reconfigure ZooKeeper TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: zk-reconfigure-tls-zookeeper
    name: Reconfigure ZooKeeper TLS/SSL Encryption
    parent: zk-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure ZooKeeper TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing ZooKeeper database via a ZooKeeperOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/zookeeper](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/zookeeper) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a ZooKeeper database

Here, We are going to create a ZooKeeper without TLS and then reconfigure the database to use TLS.

### Deploy ZooKeeper without TLS

In this section, we are going to deploy a ZooKeeper ensemble without TLS. In the next few sections we will reconfigure TLS using `ZooKeeperOpsRequest` CRD. Below is the YAML of the `ZooKeeper` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: zk-quickstart
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"

```

Let's create the `ZooKeeper` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zookeeper.yaml
zookeeper.kubedb.com/zk-quickstart created
```

Now, wait until `zk-quickstart` has status `Ready`. i.e,

```bash
$ watch kubectl get zookeeper -n demo
NAME              TYPE                    VERSION   STATUS    AGE
zk-quickstart     kubedb.com/v1alpha2     3.8.3     Ready     60s
```

Now, we can exec one zookeeper broker pod and verify configuration that the TLS is disabled.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ cat ../conf/zoo.cfg
4lw.commands.whitelist=*
dataDir=/data
tickTime=2000
initLimit=10
syncLimit=2
clientPort=2181
globalOutstandingLimit=1000
preAllocSize=65536
snapCount=10000
commitLogCount=500
snapSizeLimitInKb=4194304
maxCnxns=0
maxClientCnxns=60
minSessionTimeout=4000
maxSessionTimeout=40000
autopurge.snapRetainCount=3
autopurge.purgeInterval=1
quorumListenOnAllIPs=false
admin.serverPort=8080
authProvider.1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider
reconfigEnabled=true
standaloneEnabled=false
dynamicConfigFile=/data/zoo.cfg.dynamic
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ 
```

We can verify from the above output that TLS is disabled for this Ensemble.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in ZooKeeper. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls zookeeper-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/zookeeper-ca created
```

Now, Let's create an `Issuer` using the `zookeeper-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: zk-issuer
  namespace: demo
spec:
  ca:
    secretName: zookeeper-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zookeeper-issuer.yaml
issuer.cert-manager.io/zk-issuer created
```

### Create ZooKeeperOpsRequest

In order to add TLS to the zookeeper, we have to create a `ZooKeeperOpsRequest` CRO with our created issuer. Below is the YAML of the `ZooKeeperOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zkops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    issuerRef:
      name: zookeeper-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - zookeeper
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `zk-quickstart` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on zookeeper.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/zookeeper/concepts/zookeeper.md#spectls).

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zookeeper-add-tls.yaml
zookeeperopsrequest.ops.kubedb.com/zkops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CRO,

```bash
$ kubectl get zookeeperopsrequest -n demo
NAME            TYPE             STATUS       AGE
zkops-add-tls   ReconfigureTLS   Successful   4m36s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe zookeeperopsrequest -n demo zkops-add-tls 
Name:         zkops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T05:46:18Z
  Generation:          1
  Resource Version:    2118117
  UID:                 aa25e2b8-2583-4757-b3f7-b053fc21819f
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       zookeeper-ca-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T05:46:18Z
    Message:               ZooKeeper ops-request has started to reconfigure tls for zookeeper nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T05:46:31Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-04T05:46:26Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-04T05:46:26Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-04T05:46:26Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-04T05:46:36Z
    Message:               successfully reconciled the ZooKeeper with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T05:48:56Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T05:46:41Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T05:46:41Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T05:46:46Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T05:47:26Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T05:47:26Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T05:48:16Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T05:48:16Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T05:48:56Z
    Message:               Successfully completed reconfigureTLS for zookeeper.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now, Let's exec into a zookeeper ensemble pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ cat ../conf/zoo.cfg
4lw.commands.whitelist=*
dataDir=/data
tickTime=2000
initLimit=10
syncLimit=2
clientPort=2181
globalOutstandingLimit=1000
preAllocSize=65536
snapCount=10000
commitLogCount=500
snapSizeLimitInKb=4194304
maxCnxns=0
maxClientCnxns=60
minSessionTimeout=4000
maxSessionTimeout=40000
autopurge.snapRetainCount=3
autopurge.purgeInterval=1
quorumListenOnAllIPs=false
admin.serverPort=8080
authProvider.1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider
reconfigEnabled=true
standaloneEnabled=false
dynamicConfigFile=/data/zoo.cfg.dynamic
secureClientPort=2182
serverCnxnFactory=org.apache.zookeeper.server.NettyServerCnxnFactory
authProvider.x509=org.apache.zookeeper.server.auth.X509AuthenticationProvider
ssl.keyStore.location=/var/private/ssl/server.keystore.jks
ssl.keyStore.password=fdjk2dgffqn9
ssl.trustStore.location=/var/private/ssl/server.truststore.jks
ssl.trustStore.password=fdjk2dgffqn9
sslQuorum=true
ssl.quorum.keyStore.location=/var/private/ssl/server.keystore.jks
ssl.quorum.keyStore.password=fdjk2dgffqn9
ssl.quorum.trustStore.location=/var/private/ssl/server.truststore.jks
ssl.quorum.trustStore.password=fdjk2dgffqn9
ssl.quorum.hostnameVerification=false
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ 
```

We can see from the above output that, keystore location is `/var/private/ssl/server.keystore.jks` which means that TLS is enabled.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ openssl x509 -in /var/private/ssl/tls.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb  2 12:53:30 2025 GMT
```

So, the certificate will expire on this time `Feb 2 12:53:30 2025 GMT`.

### Create ZooKeeperOpsRequest

Now we are going to increase it using a ZooKeeperOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zkops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `zk-quickstart`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this zookeeper cluster.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zkops-rotate.yaml
zookeeperopsrequest.ops.kubedb.com/zkops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CRO,

```bash
$ kubectl get zookeeperopsrequests -n demo zkops-rotate
NAME            TYPE             STATUS       AGE
zkops-rotate    ReconfigureTLS   Successful   4m4s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe zookeeperopsrequest -n demo zkops-rotate
Name:         zkops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T13:10:03Z
  Generation:          1
  Resource Version:    2153555
  UID:                 a1886cd3-784b-4523-936c-a510327d6129
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T13:10:03Z
    Message:               ZooKeeper ops-request has started to reconfigure tls for zookeeper nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T13:10:16Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-04T13:10:11Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-04T13:10:11Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-04T13:10:11Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-04T13:10:22Z
    Message:               successfully reconciled the ZooKeeper with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T13:12:42Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T13:10:27Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T13:10:27Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T13:10:32Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T13:11:07Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T13:11:07Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T13:11:52Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T13:11:52Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T13:12:42Z
    Message:               Successfully completed reconfigureTLS for zookeeper.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  2m57s  KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/zkops-rotate
  Normal   Starting                                                  2m57s  KubeDB Ops-manager Operator  Pausing ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                2m57s  KubeDB Ops-manager Operator  Successfully paused ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zkops-rotate
  Warning  get certificate; ConditionStatus:True                     2m49s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m49s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m49s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                     2m49s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m49s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m49s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                         2m49s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                     2m44s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m44s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m44s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                     2m44s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m44s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m44s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                         2m44s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                             2m38s  KubeDB Ops-manager Operator  successfully reconciled the ZooKeeper with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-0    2m33s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-0  2m33s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  running pod; ConditionStatus:False                        2m28s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-1    113s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-1  113s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-2    68s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-2  68s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-2
  Normal   RestartNodes                                              18s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  18s    KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                18s    KubeDB Ops-manager Operator  Successfully resumed ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zkops-rotate
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ openssl x509 -in /var/private/ssl/tls.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 2 13:12:42 2025 GMT
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
$ kubectl create secret tls zookeeper-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/zookeeper-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: zk-new-issuer
  namespace: demo
spec:
  ca:
    secretName: zookeeper-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zookeeper-new-issuer.yaml
issuer.cert-manager.io/zk-new-issuer created
```

### Create ZooKeeperOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `ZooKeeperOpsRequest` CRO with the newly created issuer. Below is the YAML of the `ZooKeeperOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zkops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    issuerRef:
      name: zk-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `zk-quickstart` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our zookeeper.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zookeeper-update-tls-issuer.yaml
zookeeperpsrequest.ops.kubedb.com/zkops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CRO,

```bash
$ kubectl get zookeeperopsrequests -n demo zkops-update-issuer
NAME                  TYPE             STATUS       AGE
zkops-update-issuer   ReconfigureTLS   Successful   8m6s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe zookeeperopsrequest -n demo zkops-update-issuer
Name:         zkops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T13:27:25Z
  Generation:          1
  Resource Version:    2155331
  UID:                 399cae54-a6ab-4848-93ff-5dba09a128d7
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       zk-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T13:27:25Z
    Message:               ZooKeeper ops-request has started to reconfigure tls for zookeeper nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T13:27:35Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-04T13:27:30Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-04T13:27:30Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-04T13:27:30Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-04T13:27:40Z
    Message:               successfully reconciled the ZooKeeper with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T13:30:00Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T13:27:45Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T13:27:45Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T13:27:50Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T13:28:30Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T13:28:30Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T13:29:20Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T13:29:20Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T13:30:00Z
    Message:               Successfully completed reconfigureTLS for zookeeper.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  2m53s  KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/zkops-update-issuer
  Warning  get certificate; ConditionStatus:True                     2m48s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m48s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m48s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                     2m48s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m48s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m48s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                         2m48s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                     2m43s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m43s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m43s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                     2m43s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True               2m43s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                   2m43s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                         2m43s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                             2m38s  KubeDB Ops-manager Operator  successfully reconciled the ZooKeeper with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-0    2m33s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-0  2m33s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  running pod; ConditionStatus:False                        2m28s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-1    108s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-1  108s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-2    58s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-2  58s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-2
  Normal   RestartNodes                                              18s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  18s    KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                18s    KubeDB Ops-manager Operator  Successfully resumed ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zkops-update-issuer
```

Now, Let's exec into a zookeeper node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ keytool -list -v -keystore /var/private/ssl/server.keystore.jks -storepass fdjk2dgffqn9 | grep 'Issuer'
Issuer: O=kubedb-updated, CN=ca-updated
Issuer: O=kubedb-updated, CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a ZooKeeperOpsRequest.

### Create ZooKeeperOpsRequest

Below is the YAML of the `ZooKeeperOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ZooKeeperOpsRequest
metadata:
  name: zkops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: zk-quickstart
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `zk-quickstart` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on ZooKeeper.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `ZooKeeperOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/zookeeper/reconfigure-tls/zkops-remove.yaml
zookeeperopsrequest.ops.kubedb.com/zkops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `ZooKeeperOpsRequest` to be `Successful`.  Run the following command to watch `ZooKeeperOpsRequest` CRO,

```bash
$ kubectl get zookeeperopsrequest -n demo zkops-remove
NAME           TYPE             STATUS        AGE
zkops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `ZooKeeperOpsRequest` has succeeded. If we describe the `ZooKeeperOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe zookeeperopsrequest -n demo zkops-remove
Name:         zkops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ZooKeeperOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T13:39:19Z
  Generation:          1
  Resource Version:    2156556
  UID:                 8f669fe1-169f-4446-9d12-bf959216e2e0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  zk-quickstart
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T13:39:19Z
    Message:               ZooKeeper ops-request has started to reconfigure tls for zookeeper nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T13:39:27Z
    Message:               successfully reconciled the ZooKeeper with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T13:41:42Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T13:39:32Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T13:39:32Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-0
    Last Transition Time:  2024-11-04T13:39:37Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T13:40:22Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T13:40:22Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-1
    Last Transition Time:  2024-11-04T13:41:02Z
    Message:               get pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T13:41:02Z
    Message:               evict pod; ConditionStatus:True; PodName:zk-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--zk-quickstart-2
    Last Transition Time:  2024-11-04T13:41:42Z
    Message:               Successfully completed reconfigureTLS for zookeeper.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  2m26s  KubeDB Ops-manager Operator  Start processing for ZooKeeperOpsRequest: demo/zkops-remove
  Normal   Starting                                                  2m26s  KubeDB Ops-manager Operator  Pausing ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                2m26s  KubeDB Ops-manager Operator  Successfully paused ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zkops-remove
  Normal   UpdatePetSets                                             2m18s  KubeDB Ops-manager Operator  successfully reconciled the ZooKeeper with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-0    2m13s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-0  2m13s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-0
  Warning  running pod; ConditionStatus:False                        2m8s   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-1    83s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-1  83s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-1
  Warning  get pod; ConditionStatus:True; PodName:zk-quickstart-2    43s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:zk-quickstart-2
  Warning  evict pod; ConditionStatus:True; PodName:zk-quickstart-2  43s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:zk-quickstart-2
  Normal   RestartNodes                                              3s     KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  3s     KubeDB Ops-manager Operator  Resuming ZooKeeper database: demo/zk-quickstart
  Normal   Successful                                                3s     KubeDB Ops-manager Operator  Successfully resumed ZooKeeper database: demo/zk-quickstart for ZooKeeperOpsRequest: zkops-remove
```

Now, Let's exec into one of the broker node and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo zk-quickstart-0 -- bash
Defaulted container "zookeeper" out of: zookeeper, zookeeper-init (init)
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ cat ../conf/zoo.cfg
4lw.commands.whitelist=*
dataDir=/data
tickTime=2000
initLimit=10
syncLimit=2
clientPort=2181
globalOutstandingLimit=1000
preAllocSize=65536
snapCount=10000
commitLogCount=500
snapSizeLimitInKb=4194304
maxCnxns=0
maxClientCnxns=60
minSessionTimeout=4000
maxSessionTimeout=40000
autopurge.snapRetainCount=3
autopurge.purgeInterval=1
quorumListenOnAllIPs=false
admin.serverPort=8080
authProvider.1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider
reconfigEnabled=true
standaloneEnabled=false
dynamicConfigFile=/data/zoo.cfg.dynamic
zookeeper@zk-quickstart-0:/apache-zookeeper-3.8.3-bin$ 
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete opsrequest zkops-add-tls zkops-remove zkops-rotate zkops-update-issuer
kubectl delete zookeeper -n demo zk-quickstart
kubectl delete issuer -n demo zk-issuer zk-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [ZooKeeper object](/docs/guides/zookeeper/concepts/zookeeper.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

