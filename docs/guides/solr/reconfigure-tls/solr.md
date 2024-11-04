---
title: Reconfiguring TLS/SSL
menu:
  docs_{{ .version }}:
    identifier: sl-reconfigure-tls-solr
    name: Reconfigure Solr TLS/SSL Encryption
    parent: sl-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Solr TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Solr database via a SolrOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/Solr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Solr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Solr database

Here, We are going to create a Solr without TLS and then reconfigure the database to use TLS.

### Deploy Solr without TLS

In this section, we are going to deploy a Solr topology cluster without TLS. In the next few sections we will reconfigure TLS using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  enableSSL: true
  deletionPolicy: DoNotTerminate
  version: 9.6.1
  zookeeperRef:
    name: zoo-com
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    data:
      replicas: 2
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/solr/clustering/yamls/topology.yaml
solr.kubedb.com/solr-cluster created
```

Now, wait until `solr-cluster` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME           TYPE                  VERSION   STATUS   AGE
solr-cluster   kubedb.com/v1alpha2   9.6.1     Ready    148m
```

Now, we can exec one Solr broker pod and verify configuration that the TLS is disabled.

```bash
$ kubectl exec -it -n demo solr-cluster-data-0 -- env | grep SSL
Defaulted container "solr" out of: solr, init-solr (init)
```

We can verify from the above output that TLS is disabled for this cluster.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Solr. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca /O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls solr-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/solr-ca created
```

Now, Let's create an `Issuer` using the `Solr-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: solr-ca-issuer
  namespace: demo
spec:
  ca:
    secretName: solr-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/tls/sl-issuer.yaml
issuer.cert-manager.io/solr-ca-issuer created
```

### Create SolrOpsRequest

In order to add TLS to the Solr, we have to create a `SolrOpsRequest` CRO with our created issuer. Below is the YAML of the `SolrOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-add-tls
  namespace: demo
spec:
  apply: IfReady
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      name: solr-ca-issuer
      kind: Issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  databaseRef:
    name: solr-cluster
  type: ReconfigureTLS
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `solr-cluster` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Solr.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/solr/concepts/solr.md#spectls).

Let's create the `SolrOpsRequest` CR we have shown above,

> **Note:** For combined Solr, you just need to refer solr combined object in `databaseRef` field. To learn more about combined solr, please visit [here](/docs/guides/solr/clustering/combined_cluster.md).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure-tls/add-tls.yaml
Solropsrequest.ops.kubedb.com/slops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CRO,

```bash
$ kubectl get Solropsrequest -n demo
NAME            TYPE             STATUS       AGE
slops-add-tls   ReconfigureTLS   Successful   4m36s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe slops -n demo slops-add-tls
Name:         slops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T09:05:19Z
  Generation:          1
  Resource Version:    1533152
  UID:                 4f057ed5-33be-4753-85ce-a16e2915c6f3
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
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
      Kind:       ClusterIssuer
      Name:       self-signed-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T09:05:19Z
    Message:               Solr ops-request has started to reconfigure tls for solr nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T09:05:32Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-04T09:05:27Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-04T09:05:27Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-04T09:05:27Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-04T09:05:38Z
    Message:               successfully reconciled the Solr with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T09:08:13Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T09:05:43Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T09:05:43Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T09:05:48Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T09:06:33Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T09:06:33Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T09:07:23Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T09:07:23Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T09:08:13Z
    Message:               Successfully completed reconfigureTLS for solr.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now, Let's exec into a Solr broker pod and verify the configuration that the TLS is enabled.

```bash
 $ kubectl exec -it -n demo solr-cluster-data-0 -- env | grep -i ssl
Defaulted container "solr" out of: solr, init-solr (init)
JAVA_OPTS= -Djavax.net.ssl.trustStore=/var/solr/etc/truststore.p12 -Djavax.net.ssl.trustStorePassword=Ni5tEgfjahzS53D3 -Djavax.net.ssl.keyStore=/var/solr/etc/keystore.p12 -Djavax.net.ssl.keyStorePassword=Ni5tEgfjahzS53D3 -Djavax.net.ssl.keyStoreType=PKCS12 -Djavax.net.ssl.trustStoreType=PKCS12
SOLR_SSL_KEY_STORE_PASSWORD=Ni5tEgfjahzS53D3
SOLR_SSL_TRUST_STORE=/var/solr/etc/truststore.p12
SOLR_SSL_KEY_STORE=/var/solr/etc/keystore.p12
SOLR_SSL_WANT_CLIENT_AUTH=false
SOLR_SSL_ENABLED=true
SOLR_SSL_TRUST_STORE_PASSWORD=Ni5tEgfjahzS53D3
SOLR_SSL_NEED_CLIENT_AUTH=false
```

We can see from the above output that, keystore location is `/var/solr/etc/keystore.p12` which means that TLS is enabled.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
$ $ kubectl exec -it -n demo solr-cluster-data-0 -- keytool -list -v -keystore /var/solr/etc/keystore.p12 -storepass Ni5tEgfjahzS53D3 | grep -E 'Valid from|Alias name'
Alias name: 1
Valid from: Mon Nov 04 09:05:23 UTC 2024 until: Sun Feb 02 09:05:23 UTC 2025
Valid from: Thu Aug 15 05:59:09 UTC 2024 until: Fri Aug 15 05:59:09 UTC 2025

```

So, the certificate will expire on this time `Sun Feb 02 09:05:23 UTC 2025`.

### Create SolrOpsRequest

Now we are going to increase it using a SolrOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: solr-cluster
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `solr-cluster`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this Solr cluster.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure-tls/rotate-tls.yaml
Solropsrequest.ops.kubedb.com/slops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CRO,

```bash
$ kubectl get slops -n demo slops-rotate
NAME           TYPE             STATUS       AGE
slops-rotate   ReconfigureTLS   Successful   32m
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe slops -n demo slops-rotate
Name:         slops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T12:20:18Z
  Generation:          1
  Resource Version:    1550013
  UID:                 0a9e1d2c-f322-4f7d-8344-43440456331b
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T12:20:18Z
    Message:               Solr ops-request has started to reconfigure tls for solr nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T12:20:31Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-04T12:20:26Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-04T12:20:26Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-04T12:20:26Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-04T12:20:37Z
    Message:               successfully reconciled the Solr with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T12:23:07Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T12:20:42Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T12:20:42Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T12:20:47Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T12:21:32Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T12:21:32Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T12:22:22Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T12:22:22Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T12:23:07Z
    Message:               Successfully completed reconfigureTLS for solr.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                               Age   From                         Message
  ----     ------                                                               ----  ----                         -------
  Normal   Starting                                                             33m   KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/rotate-tls
  Normal   Starting                                                             33m   KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                                                           33m   KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: rotate-tls
  Warning  get certificate; ConditionStatus:True                                33m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          33m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              33m   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                33m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          33m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              33m   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                    33m   KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                33m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          33m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              33m   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                33m   KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          33m   KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              33m   KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                    33m   KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                        33m   KubeDB Ops-manager Operator  successfully reconciled the Solr with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0       33m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0     33m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  running pod; ConditionStatus:False                                   32m   KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-data-0           32m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0         32m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0    31m   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0  31m   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Normal   RestartNodes                                                         30m   KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                             30m   KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                                           30m   KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: rotate-tls
```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo solr-cluster-data-0 -- keytool -list -v -keystore /var/solr/etc/keystore.p12 -storepass Ni5tEgfjahzS53D3 | grep -E 'Valid from|Alias name'
Defaulted container "solr" out of: solr, init-solr (init)
Alias name: 1
Valid from: Mon Nov 04 12:23:07 UTC 2024 until: Sun Feb 02 12:23:07 UTC 2025
Valid from: Thu Aug 15 05:59:09 UTC 2024 until: Fri Aug 15 05:59:09 UTC 2025
```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated /O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls Solr-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/solr-new-ca created
```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: sl-new-issuer
  namespace: demo
spec:
  ca:
    secretName: solr-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure-tls/sl-new-issuer.yaml
issuer.cert-manager.io/sl-new-issuer created
```

### Create SolrOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `SolrOpsRequest` CRO with the newly created issuer. Below is the YAML of the `SolrOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: solr-cluster
  tls:
    issuerRef:
      name: sl-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `solr-cluster` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our Solr.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Solr/reconfigure-tls/sl-update-issuer.yaml
solrpsrequest.ops.kubedb.com/slops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CRO,

```bash
$ kubectl get solropsrequests -n demo slops-update-issuer
NAME                  TYPE             STATUS       AGE
slops-update-issuer   ReconfigureTLS   Successful   8m6s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe slops -n demo slops-update-issuer 
Name:         slops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T13:04:09Z
  Generation:          1
  Resource Version:    1553891
  UID:                 aa1a5101-8daa-4a0e-b640-c6ba8c20a431
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       sl-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T13:04:09Z
    Message:               Solr ops-request has started to reconfigure tls for solr nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T13:04:22Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2024-11-04T13:04:17Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2024-11-04T13:04:17Z
    Message:               check ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CheckReadyCondition
    Last Transition Time:  2024-11-04T13:04:17Z
    Message:               issuing condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssuingCondition
    Last Transition Time:  2024-11-04T13:04:27Z
    Message:               successfully reconciled the Solr with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T13:07:02Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T13:04:32Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T13:04:32Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T13:04:37Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T13:05:22Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T13:05:22Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T13:06:12Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T13:06:12Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T13:07:02Z
    Message:               Successfully completed reconfigureTLS for solr.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                               Age    From                         Message
  ----     ------                                                               ----   ----                         -------
  Normal   Starting                                                             3m52s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-update-issuer
  Normal   Starting                                                             3m52s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                                                           3m52s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: slops-update-issuer
  Warning  get certificate; ConditionStatus:True                                3m44s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          3m44s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              3m44s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                3m44s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          3m44s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              3m44s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                    3m44s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  get certificate; ConditionStatus:True                                3m39s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          3m39s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              3m39s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                                3m39s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  check ready condition; ConditionStatus:True                          3m39s  KubeDB Ops-manager Operator  check ready condition; ConditionStatus:True
  Warning  issuing condition; ConditionStatus:True                              3m39s  KubeDB Ops-manager Operator  issuing condition; ConditionStatus:True
  Normal   CertificateSynced                                                    3m39s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Normal   UpdatePetSets                                                        3m34s  KubeDB Ops-manager Operator  successfully reconciled the Solr with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0       3m29s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0     3m29s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  running pod; ConditionStatus:False                                   3m24s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-data-0           2m39s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0         2m39s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0    109s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0  109s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Normal   RestartNodes                                                         59s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                             59s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                                           59s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: slops-update-issuer
```

Now, Let's exec into a Solr node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo solr-cluster-data-0 -- bash
Defaulted container "solr" out of: solr, init-solr (init)
solr@solr-cluster-data-0:/opt/solr-9.6.1$ keytool -list -v -keystore /var/solr/etc/keystore.p12 -storepass Ni5tEgfjahzS53D3 | grep 'Issuer'
Issuer: O=kubedb-updated, CN="ca-updated "
Issuer: O=kubedb-updated, CN="ca-updated "

```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a SolrOpsRequest.

### Create SolrOpsRequest

Below is the YAML of the `SolrOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: slops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: solr-cluster
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `solr-cluster` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Solr.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure-tls/remove-tls.yaml
solropsrequest.ops.kubedb.com/slops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CRO,

```bash
$ kubectl get solropsrequest -n demo slops-remove
NAME           TYPE             STATUS        AGE
slops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe slops -n demo slops-remove
Name:         slops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-04T13:15:15Z
  Generation:          1
  Resource Version:    1555016
  UID:                 a98301fe-af47-4554-9de9-bf6be3041dc3
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  solr-cluster
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2024-11-04T13:15:15Z
    Message:               Solr ops-request has started to reconfigure tls for solr nodes
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2024-11-04T13:15:23Z
    Message:               successfully reconciled the Solr with tls configuration
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-04T13:17:58Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-04T13:15:28Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T13:15:28Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-overseer-0
    Last Transition Time:  2024-11-04T13:15:33Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-04T13:16:13Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T13:16:13Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-data-0
    Last Transition Time:  2024-11-04T13:17:08Z
    Message:               get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T13:17:08Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-cluster-coordinator-0
    Last Transition Time:  2024-11-04T13:17:58Z
    Message:               Successfully completed reconfigureTLS for solr.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                               Age    From                         Message
  ----     ------                                                               ----   ----                         -------
  Normal   Starting                                                             6m3s   KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/slops-remove
  Normal   Starting                                                             6m3s   KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr-cluster
  Normal   Successful                                                           6m3s   KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr-cluster for SolrOpsRequest: slops-remove
  Normal   UpdatePetSets                                                        5m55s  KubeDB Ops-manager Operator  successfully reconciled the Solr with tls configuration
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0       5m50s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0     5m50s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
  Warning  running pod; ConditionStatus:False                                   5m45s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-data-0           5m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0         5m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
  Warning  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0    4m10s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0  4m10s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
  Normal   RestartNodes                                                         3m20s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                             3m20s  KubeDB Ops-manager Operator  Resuming Solr database: demo/solr-cluster
  Normal   Successful                                                           3m20s  KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr-cluster for SolrOpsRequest: slops-remove
```

Now, Let's exec into one of the broker node and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo solr-cluster-data-0 -- env | grep -i ssl
Defaulted container "solr" out of: solr, init-solr (init)
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete opsrequest slops-add-tls slops-remove slops-rotate slops-update-issuer
kubectl delete solr -n demo solr-cluster
kubectl delete issuer -n demo sl-issuer sl-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md)
- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

