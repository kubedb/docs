---
title: Reconfigure Elasticsearch TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: es-reconfigure-tls
    name: Reconfigure Elasticsearch TLS/SSL Encryption
    parent: es-reconfigure-tls-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Elasticsearch TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing Elasticsearch database via a ElasticsearchOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/Elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/Elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a Elasticsearch database

Here, We are going to create a Elasticsearch without TLS and then reconfigure the database to use TLS.

### Deploy Elasticsearch without TLS

In this section, we are going to deploy a Elasticsearch topology cluster without TLS. In the next few sections we will reconfigure TLS using `ElasticsearchOpsRequest` CRD. Below is the YAML of the `Elasticsearch` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-demo
  namespace: demo
spec:
  deletionPolicy: WipeOut
  enableSSL: true
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: local-path
  storageType: Durable
  version: xpack-8.11.1


```

Let's create the `Elasticsearch` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/Elasticsearch.yaml
Elasticsearch.kubedb.com/es-demo created
```

Now, wait until `es-demo` has status `Ready`. i.e,

```bash
$ kubectl get es -n demo -w
NAME      VERSION        STATUS   AGE
es-demo   xpack-8.11.1   Ready    26h

```

Now, we can exec one hazelcast pod and verify configuration that the TLS is disabled.
```bash
$ kubectl exec -n demo es-demo-0 -- \
                                        cat /usr/share/elasticsearch/config/elasticsearch.yml | grep -A 2 -i xpack.security

Defaulted container "elasticsearch" out of: elasticsearch, init-sysctl (init), config-merger (init)
xpack.security.enabled: true

xpack.security.transport.ssl.enabled: true
xpack.security.transport.ssl.verification_mode: certificate
xpack.security.transport.ssl.key: certs/transport/tls.key
xpack.security.transport.ssl.certificate: certs/transport/tls.crt
xpack.security.transport.ssl.certificate_authorities: [ "certs/transport/ca.crt" ]

xpack.security.http.ssl.enabled: false

```
Here, transport TLS is enabled but HTTP TLS is disabled. So, internal node to node communication is encrypted but communication from client to node is not encrypted.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in Elasticsearch. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

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
$ kubectl create secret tls es-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/es-ca created
```

Now, Let's create an `Issuer` using the `Elasticsearch-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: es-issuer
  namespace: demo
spec:
  ca:
    secretName: es-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/Elasticsearch-issuer.yaml
issuer.cert-manager.io/es-issuer created
```

### Create ElasticsearchOpsRequest

In order to add TLS to the Elasticsearch, we have to create a `ElasticsearchOpsRequest` CRO with our created issuer. Below is the YAML of the `ElasticsearchOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: es-demo
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: es-issuer
    certificates:
    - alias: http
      subject:
        organizations:
        - kubedb.com
      emailAddresses:
      - abc@kubedb.com
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `es-demo` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Elasticsearch.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

> **Note:** For combined Elasticsearch, you just need to refer Elasticsearch combined object in `databaseRef` field. To learn more about combined Elasticsearch, please visit [here](/docs/guides/elasticsearch/clustering/combined-cluster/index.md).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/Elasticsearch-add-tls.yaml
Elasticsearchopsrequest.ops.kubedb.com/add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CRO,

```bash
$ kubectl get Elasticsearchopsrequest -n demo
NAME      TYPE             STATUS       AGE
add-tls   ReconfigureTLS   Successful   73m
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo add-tls
Name:         add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-28T05:16:12Z
  Generation:          1
  Resource Version:    884868
  UID:                 2fa3b86a-4cfa-4e51-8cde-c5d7508c3eb0
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-demo
  Tls:
    Certificates:
      Alias:  http
      Email Addresses:
        abc@kubedb.com
      Subject:
        Organizations:
          kubedb.com
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       es-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-11-28T05:16:12Z
    Message:               Elasticsearch ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-11-28T05:16:20Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-11-28T05:16:20Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2025-11-28T05:16:20Z
    Message:               issue condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssueCondition
    Last Transition Time:  2025-11-28T05:16:20Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-11-28T05:16:32Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-0
    Last Transition Time:  2025-11-28T05:16:32Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-0
    Last Transition Time:  2025-11-28T05:16:32Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-0
    Last Transition Time:  2025-11-28T05:17:42Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-28T05:16:57Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-1
    Last Transition Time:  2025-11-28T05:16:57Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-1
    Last Transition Time:  2025-11-28T05:16:57Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-1
    Last Transition Time:  2025-11-28T05:17:22Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-2
    Last Transition Time:  2025-11-28T05:17:22Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-2
    Last Transition Time:  2025-11-28T05:17:22Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-2
    Last Transition Time:  2025-11-28T05:17:47Z
    Message:               Successfully restarted all the nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-11-28T05:17:51Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

Now, Let's exec into a Elasticsearch broker pod and verify the configuration that the TLS is enabled.

```bash
$ kubectl exec -n demo es-demo-0 -- \
                                  cat /usr/share/elasticsearch/config/elasticsearch.yml | grep -A 2 -i xpack.security

Defaulted container "elasticsearch" out of: elasticsearch, init-sysctl (init), config-merger (init)
xpack.security.enabled: true

xpack.security.transport.ssl.enabled: true
xpack.security.transport.ssl.verification_mode: certificate
xpack.security.transport.ssl.key: certs/transport/tls.key
xpack.security.transport.ssl.certificate: certs/transport/tls.crt
xpack.security.transport.ssl.certificate_authorities: [ "certs/transport/ca.crt" ]

xpack.security.http.ssl.enabled: true
xpack.security.http.ssl.key:  certs/http/tls.key
xpack.security.http.ssl.certificate: certs/http/tls.crt
xpack.security.http.ssl.certificate_authorities: [ "certs/http/ca.crt" ]

```

We can see from the above output that,  `xpack.security.http.ssl.enabled: true` which means TLS is enabled for HTTP communication.

## Rotate Certificate

Now we are going to rotate the certificate of this cluster. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -n demo es-demo-0 -- /bin/sh -c '\
                                                      openssl s_client -connect localhost:9200 -showcerts < /dev/null 2>/dev/null | \
                                                      sed -ne "/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p" > /tmp/server.crt && \
                                                      openssl x509 -in /tmp/server.crt -noout -enddate'
Defaulted container "elasticsearch" out of: elasticsearch, init-sysctl (init), config-merger (init)
notAfter=Feb 26 05:16:15 2026 GMT

```

So, the certificate will expire on this time `Feb 26 05:16:17 2026 GMT`.

### Create ElasticsearchOpsRequest

Now we are going to increase it using a ElasticsearchOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: es-demo
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `es-demo`.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our cluster.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this Elasticsearch cluster.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/esops-rotate.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CRO,

```bash
$ kubectl get Elasticsearchopsrequest -n demo esops-rotate
NAME           TYPE             STATUS       AGE
esops-rotate   ReconfigureTLS   Successful   85m

```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo esops-rotate
Name:         esops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-28T07:02:38Z
  Generation:          1
  Resource Version:    893511
  UID:                 43503dc9-ddeb-4569-a8a9-b10a96feeb60
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-demo
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-11-28T07:02:38Z
    Message:               Elasticsearch ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-11-28T07:02:41Z
    Message:               successfully add issuing condition to all the certificates
    Observed Generation:   1
    Reason:                IssueCertificatesSucceeded
    Status:                True
    Type:                  IssueCertificatesSucceeded
    Last Transition Time:  2025-11-28T07:02:46Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-11-28T07:02:46Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2025-11-28T07:02:47Z
    Message:               issue condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssueCondition
    Last Transition Time:  2025-11-28T07:02:47Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-11-28T07:02:56Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-0
    Last Transition Time:  2025-11-28T07:02:56Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-0
    Last Transition Time:  2025-11-28T07:02:56Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-0
    Last Transition Time:  2025-11-28T07:04:06Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-28T07:03:21Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-1
    Last Transition Time:  2025-11-28T07:03:21Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-1
    Last Transition Time:  2025-11-28T07:03:21Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-1
    Last Transition Time:  2025-11-28T07:03:46Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-2
    Last Transition Time:  2025-11-28T07:03:46Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-2
    Last Transition Time:  2025-11-28T07:03:46Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-2
    Last Transition Time:  2025-11-28T07:04:11Z
    Message:               Successfully restarted all the nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-11-28T07:04:15Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```



As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb-updated"
.+........+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*........+.....+......+...+.+..............+....+..+.+...+......+.....+.........+............+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*.......+........+.......+...+......+.....+..........+..+.........+......+....+...+..+....+..+.......+............+...+..+...+.+............+..+................+.....+................+.....+.+........+.+.....+.........................+........+......+....+...........+.+....................+.+..+......+......+...+...+...+......+.+...+.........+.....+.......+...+..+.............+.....+.+..............+......+.+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
..+........+...+...............+...+....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*...+...+...+...................+.....+.+......+.....+.........+....+...+.....+...+.......+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++*....+...+..+............+....+..+...+..........+.........+......+.........+...........+....+..+.+..+.......+.....+++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++

```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls es-new-ca \
                                       --cert=ca.crt \
                                       --key=ca.key \
                                       --namespace=demo
secret/es-new-ca created

```

Now, Let's create a new `Issuer` using the `mongo-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: es-new-issuer
  namespace: demo
spec:
  ca:
    secretName: es-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/Elasticsearch-new-issuer.yaml
issuer.cert-manager.io/es-new-issuer created
```

### Create ElasticsearchOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `ElasticsearchOpsRequest` CRO with the newly created issuer. Below is the YAML of the `ElasticsearchOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-update-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: es-demo
  tls:
    issuerRef:
      name: es-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `es-demo` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our Elasticsearch.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/Elasticsearch-update-tls-issuer.yaml
Elasticsearchpsrequest.ops.kubedb.com/esops-update-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CRO,

```bash
$ kubectl get Elasticsearchopsrequests -n demo esops-update-issuer
NAME                  TYPE             STATUS       AGE
esops-update-issuer   ReconfigureTLS   Successful   6m28s
```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo esops-update-issuer
Name:         esops-update-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-28T09:32:41Z
  Generation:          1
  Resource Version:    905680
  UID:                 9abdfdc1-2c7e-4d1d-b226-029c0e6d99fc
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-demo
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       es-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-11-28T09:32:41Z
    Message:               Elasticsearch ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-11-28T09:32:49Z
    Message:               get certificate; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetCertificate
    Last Transition Time:  2025-11-28T09:32:49Z
    Message:               ready condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  ReadyCondition
    Last Transition Time:  2025-11-28T09:32:49Z
    Message:               issue condition; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IssueCondition
    Last Transition Time:  2025-11-28T09:32:49Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                CertificateSynced
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2025-11-28T09:33:00Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-0
    Last Transition Time:  2025-11-28T09:33:00Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-0
    Last Transition Time:  2025-11-28T09:33:00Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-0
    Last Transition Time:  2025-11-28T09:35:31Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-28T09:33:25Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-1
    Last Transition Time:  2025-11-28T09:33:25Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-1
    Last Transition Time:  2025-11-28T09:33:25Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-1
    Last Transition Time:  2025-11-28T09:33:50Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-2
    Last Transition Time:  2025-11-28T09:33:50Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-2
    Last Transition Time:  2025-11-28T09:33:50Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-2
    Last Transition Time:  2025-11-28T09:34:15Z
    Message:               Successfully restarted all the nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-11-28T09:34:21Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                     Age    From                         Message
  ----     ------                                                     ----   ----                         -------
  Normal   PauseDatabase                                              6m47s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-demo
  Warning  get certificate; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issue condition; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  issue condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issue condition; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  issue condition; ConditionStatus:True
  Warning  get certificate; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  get certificate; ConditionStatus:True
  Warning  ready condition; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  ready condition; ConditionStatus:True
  Warning  issue condition; ConditionStatus:True                      6m39s  KubeDB Ops-manager Operator  issue condition; ConditionStatus:True
  Normal   CertificateSynced                                          6m39s  KubeDB Ops-manager Operator  Successfully synced all certificates
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-0        6m28s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:True; PodName:es-demo-0  6m28s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-0
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-0         6m28s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:False                    6m23s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     6m8s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-1        6m3s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:True; PodName:es-demo-1  6m3s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-1
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-1         6m3s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:False                    5m58s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     5m43s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-2        5m38s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:True; PodName:es-demo-2  5m38s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-2
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-2         5m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:False                    5m33s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     5m18s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                               5m13s  KubeDB Ops-manager Operator  Successfully restarted all the nodes
  Normal   ResumeDatabase                                             5m7s   KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-demo
  Normal   ResumeDatabase                                             5m7s   KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-demo
  Normal   Successful                                                 5m7s   KubeDB Ops-manager Operator  Successfully Reconfigured TLS
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-0        5m7s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:True; PodName:es-demo-0  5m7s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-0
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-0         5m7s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:False                    5m2s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     4m47s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-1        4m42s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:True; PodName:es-demo-1  4m42s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-1
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-1         4m42s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:False                    4m37s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     4m22s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-2        4m17s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:True; PodName:es-demo-2  4m17s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-2
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-2         4m17s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:False                    4m12s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     3m57s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                               3m52s  KubeDB Ops-manager Operator  Successfully restarted all the nodes

```

Now, Let's exec into a Elasticsearch node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo es-demo-0 -- bash
elasticsearch@es-demo-0:~$ openssl x509 -in /usr/share/elasticsearch/config/certs/http/..2025_11_28_09_34_24.3912740802/tls.crt -noout -issuer
issuer=CN = ca-updated, O = kubedb-updated
elasticsearch@es-demo-0:~$ openssl x509 -in /usr/share/elasticsearch/config/certs/transport/..2025_11_28_09_34_24.2105953641/tls.crt -noout -issuer
issuer=CN = ca-updated, O = kubedb-updated

```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a ElasticsearchOpsRequest.

### Create ElasticsearchOpsRequest

Below is the YAML of the `ElasticsearchOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: es-demo
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `es-demo` cluster.
- `spec.type` specifies that we are performing `ReconfigureTLS` on Elasticsearch.
- `spec.tls.remove` specifies that we want to remove tls from this cluster.

Let's create the `ElasticsearchOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Elasticsearch/reconfigure-tls/esops-remove.yaml
Elasticsearchopsrequest.ops.kubedb.com/esops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `ElasticsearchOpsRequest` to be `Successful`.  Run the following command to watch `ElasticsearchOpsRequest` CRO,

```bash
$ kubectl get Elasticsearchopsrequest -n demo esops-remove
NAME           TYPE             STATUS       AGE
esops-remove   ReconfigureTLS   Successful   3m16s

```

We can see from the above output that the `ElasticsearchOpsRequest` has succeeded. If we describe the `ElasticsearchOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe Elasticsearchopsrequest -n demo esops-remove
Name:         esops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         ElasticsearchOpsRequest
Metadata:
  Creation Timestamp:  2025-11-28T10:42:00Z
  Generation:          1
  Resource Version:    911280
  UID:                 7eefbe63-1fcc-4ca3-bb5d-65ec22d7fd9a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  es-demo
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2025-11-28T10:42:00Z
    Message:               Elasticsearch ops request is reconfiguring TLS
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  ReconfigureTLS
    Last Transition Time:  2025-11-28T10:42:14Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-0
    Last Transition Time:  2025-11-28T10:42:14Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-0
    Last Transition Time:  2025-11-28T10:42:14Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-0
    Last Transition Time:  2025-11-28T10:43:24Z
    Message:               create es client; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient
    Last Transition Time:  2025-11-28T10:42:34Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-1
    Last Transition Time:  2025-11-28T10:42:34Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-1
    Last Transition Time:  2025-11-28T10:42:34Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-1
    Last Transition Time:  2025-11-28T10:43:09Z
    Message:               pod exists; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  PodExists--es-demo-2
    Last Transition Time:  2025-11-28T10:43:09Z
    Message:               create es client; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  CreateEsClient--es-demo-2
    Last Transition Time:  2025-11-28T10:43:09Z
    Message:               evict pod; ConditionStatus:True; PodName:es-demo-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--es-demo-2
    Last Transition Time:  2025-11-28T10:43:29Z
    Message:               Successfully restarted all the nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2025-11-28T10:43:33Z
    Message:               Successfully reconfigured TLS
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                     Age    From                         Message
  ----     ------                                                     ----   ----                         -------
  Normal   PauseDatabase                                              3m43s  KubeDB Ops-manager Operator  Pausing Elasticsearch demo/es-demo
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-0        3m29s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:True; PodName:es-demo-0  3m29s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-0
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-0         3m29s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-0
  Warning  create es client; ConditionStatus:False                    3m24s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     3m14s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-1        3m9s   KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:True; PodName:es-demo-1  3m9s   KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-1
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-1         3m9s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-1
  Warning  create es client; ConditionStatus:False                    3m4s   KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     2m39s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Warning  pod exists; ConditionStatus:True; PodName:es-demo-2        2m34s  KubeDB Ops-manager Operator  pod exists; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:True; PodName:es-demo-2  2m34s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True; PodName:es-demo-2
  Warning  evict pod; ConditionStatus:True; PodName:es-demo-2         2m34s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:es-demo-2
  Warning  create es client; ConditionStatus:False                    2m29s  KubeDB Ops-manager Operator  create es client; ConditionStatus:False
  Warning  create es client; ConditionStatus:True                     2m19s  KubeDB Ops-manager Operator  create es client; ConditionStatus:True
  Normal   RestartNodes                                               2m14s  KubeDB Ops-manager Operator  Successfully restarted all the nodes
  Normal   ResumeDatabase                                             2m10s  KubeDB Ops-manager Operator  Resuming Elasticsearch demo/es-demo
  Normal   ResumeDatabase                                             2m10s  KubeDB Ops-manager Operator  Successfully resumed Elasticsearch demo/es-demo
  Normal   Successful                                                 2m10s  KubeDB Ops-manager Operator  Successfully Reconfigured TLS

```

Now, Let's exec into one of the broker node and find out that TLS is disabled or not.

```bash
$ kubectl exec -n demo es-demo-0 -- \
       cat /usr/share/elasticsearch/config/elasticsearch.yml | grep -A 2 -i xpack.security

Defaulted container "elasticsearch" out of: elasticsearch, init-sysctl (init), config-merger (init)
xpack.security.enabled: true

xpack.security.transport.ssl.enabled: true
xpack.security.transport.ssl.verification_mode: certificate
xpack.security.transport.ssl.key: certs/transport/tls.key
xpack.security.transport.ssl.certificate: certs/transport/tls.crt
xpack.security.transport.ssl.certificate_authorities: [ "certs/transport/ca.crt" ]

xpack.security.http.ssl.enabled: false

```

So, we can see from the above that, `xpack.security.http.ssl.enabled` is set to `false` which means TLS is disabled for HTTP layer. Also, the transport layer TLS settings are removed from the `elasticsearch.yml` file.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete elasticsearchopsrequest -n demo add-tls esops-remove esops-rotate esops-update-issuer
kubectl delete Elasticsearch -n demo es-demo
kubectl delete issuer -n demo es-issuer es-new-issuer
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Different Elasticsearch topology clustering modes [here](/docs/guides/elasticsearch/clustering/_index.md).
- Monitor your Elasticsearch database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

