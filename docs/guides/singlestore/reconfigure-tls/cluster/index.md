---
title: Reconfigure SingleStore TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-reconfigure-tls-cluster
    name: Reconfigure TLS/SSL Encryption
    parent: guides-sdb-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure SingleStore TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing SingleStore database via a SingleStoreOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes Cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.6.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Add TLS to a SingleStore Cluster

Here, We are going to create a SingleStore database without TLS and then reconfigure the database to use TLS.
> **Note:** Steps for reconfiguring TLS of SingleStore `Standalone` is same as SingleStore `Cluster`.

### Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

### Deploy SingleStore without TLS

In this section, we are going to deploy a SingleStore Cluster database without TLS. In the next few sections we will reconfigure TLS using `SingleStoreOpsRequest` CRD. Below is the YAML of the `SingleStore` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sample-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "700m"
              requests:
                memory: "2Gi"
                cpu: "700m"
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 1
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "700m"
                requests:
                  memory: "2Gi"
                  cpu: "700m"                      
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  deletionPolicy: WipeOut
```

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-tls/cluster/examples/sample-sdb.yaml
singlestore.kubedb.com/sample-sdb created
```

Now, wait until `sample-sdb` has status `Ready`. i.e,

```bash
$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
sample-sdb   kubedb.com/v1alpha2   8.7.10    Ready    38m

```

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1188
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show variables like '%ssl%';
+---------------------------------+------------+
| Variable_name                   | Value      |
+---------------------------------+------------+
| default_user_require_ssl        | OFF        |
| exporter_ssl_ca                 |            |
| exporter_ssl_capath             |            |
| exporter_ssl_cert               |            |
| exporter_ssl_key                |            |
| exporter_ssl_key_passphrase     | [redacted] |
| have_openssl                    | OFF        |
| have_ssl                        | OFF        |
| jwks_ssl_ca_certificate         |            |
| node_replication_ssl_only       | OFF        |
| openssl_version                 | 805306480  |
| processlist_rpc_json_max_size   | 2048       |
| ssl_ca                          |            |
| ssl_capath                      |            |
| ssl_cert                        |            |
| ssl_cipher                      |            |
| ssl_fips_mode                   | OFF        |
| ssl_key                         |            |
| ssl_key_passphrase              | [redacted] |
| ssl_last_reload_attempt_time    |            |
| ssl_last_successful_reload_time |            |
+---------------------------------+------------+
21 rows in set (0.00 sec)
```

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=memsql/O=kubedb"
Generating a RSA private key
...........................................................................+++++
........................................................................................................+++++
writing new private key to './ca.key'
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls sdb-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/sdb-ca created
```

Now, we are going to create an `Issuer` using the `sdb-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: sdb-issuer
  namespace: demo
spec:
  ca:
    secretName: sdb-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-tls/cluster/examples/issuer.yaml
issuer.cert-manager.io/sdb-issuer created
```

### Create SingleStoreOpsRequest

In order to add TLS to the database, we have to create a `SingleStoreOpsRequest` CRO with our created issuer. Below is the YAML of the `SingleStoreOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-sdb
  tls:
    issuerRef:
      name: sdb-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - singlestore
          organizationalUnits:
            - client
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/singlestore/concepts/singlestore.md#spectls).

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-tls/cluster/examples/sdbops-add-tls.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-add-tls created

```

#### Verify TLS Enabled Successfully

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CRO,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                                                  TYPE             STATUS       AGE
singlestoreopsrequest.ops.kubedb.com/sdbops-add-tls   ReconfigureTLS   Successful   2m45s

```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded.

Now, we are going to connect to the database for verifying the `SingleStore` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ ls etc/memsql/certs/
ca.crt	client.crt  client.key	server.crt  server.key
[memsql@sample-sdb-aggregator-0 /]$ 
[memsql@sample-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 90
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show variables like '%ssl%';
+---------------------------------+------------------------------+
| Variable_name                   | Value                        |
+---------------------------------+------------------------------+
| default_user_require_ssl        | OFF                          |
| exporter_ssl_ca                 |                              |
| exporter_ssl_capath             |                              |
| exporter_ssl_cert               |                              |
| exporter_ssl_key                |                              |
| exporter_ssl_key_passphrase     | [redacted]                   |
| have_openssl                    | ON                           |
| have_ssl                        | ON                           |
| jwks_ssl_ca_certificate         |                              |
| node_replication_ssl_only       | OFF                          |
| openssl_version                 | 805306480                    |
| processlist_rpc_json_max_size   | 2048                         |
| ssl_ca                          | /etc/memsql/certs/ca.crt     |
| ssl_capath                      |                              |
| ssl_cert                        | /etc/memsql/certs/server.crt |
| ssl_cipher                      |                              |
| ssl_fips_mode                   | OFF                          |
| ssl_key                         | /etc/memsql/certs/server.key |
| ssl_key_passphrase              | [redacted]                   |
| ssl_last_reload_attempt_time    |                              |
| ssl_last_successful_reload_time |                              |
+---------------------------------+------------------------------+
21 rows in set (0.00 sec)
```

We can see from the above output that, `have_ssl` is set to `ture`. So, database TLS is enabled successfully to this database.

## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ openssl x509 -in /etc/memsql/certs/server.crt -inform  PEM -enddate -nameopt RFC2253 -noout
notAfter=Jan  6 06:56:55 2025 GMT

```

So, the certificate will expire on this time `Jan  6 06:56:55 2025 GMT`.

### Create SingleStoreOpsRequest

Now we are going to increase it using a SingleStoreOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-sdb
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-tls/cluster/examples/sdbops-rotate-tls.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-rotate-tls created
```

#### Verify Certificate Rotated Successfully

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CRO,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                TYPE             STATUS       AGE
sdbops-rotate-tls   ReconfigureTLS   Successful   4m14s

```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded. Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ openssl x509 -in /etc/memsql/certs/server.crt -inform  PEM -enddate -nameopt RFC2253 -noout
notAfter=Jan  6 07:15:47 2025 GMT

```

As we can see from the above output, the certificate has been rotated successfully.

## Update Certificate

Now, we are going to update the server certificate.

- Let's describe the server certificate `sample-sdb-server-cert`
```bash
 $ kubectl describe certificate -n demo sample-sdb-server-cert
Name:         sample-sdb-server-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sample-sdb
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=singlestores.kubedb.com
Annotations:  <none>
API Version:  cert-manager.io/v1
Kind:         Certificate
Metadata:
  Creation Timestamp:  2024-10-08T06:56:55Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Singlestore
    Name:                  sample-sdb
    UID:                   5e42538e-c631-4583-9f47-328742e6d938
  Resource Version:        4965452
  UID:                     65c6936b-1bd0-413d-a96d-edf0cff17897
Spec:
  Common Name:  sample-sdb
  Dns Names:
    *.sample-sdb-pods.demo.svc
    *.sample-sdb-pods.demo.svc.cluster.local
    *.sample-sdb.demo.svc
    localhost
    sample-sdb
    sample-sdb.demo.svc
  Ip Addresses:
    127.0.0.1
  Issuer Ref:
    Group:      cert-manager.io
    Kind:       Issuer
    Name:       sdb-issuer
  Secret Name:  sample-sdb-server-cert
  Usages:
    digital signature
    key encipherment
    server auth
    client auth
Status:
  Conditions:
    Last Transition Time:  2024-10-08T06:56:56Z
    Message:               Certificate is up to date and has not expired
    Observed Generation:   1
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2025-01-06T07:15:47Z
  Not Before:              2024-10-08T07:15:47Z
  Renewal Time:            2024-12-07T07:15:47Z
  Revision:                23
Events:
  Type    Reason     Age                    From                                       Message
  ----    ------     ----                   ----                                       -------
  Normal  Generated  23m                    cert-manager-certificates-key-manager      Stored new private key in temporary Secret resource "sample-sdb-server-cert-48d82"
  Normal  Requested  23m                    cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-msv5z"
  Normal  Issuing    23m                    cert-manager-certificates-trigger          Issuing certificate as Secret does not exist
  Normal  Requested  7m39s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-qpmbp"
  Normal  Requested  7m38s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-2cldn"
  Normal  Requested  7m34s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-qtm4z"
  Normal  Requested  7m33s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-5tflq"
  Normal  Requested  7m29s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-qzd6h"
  Normal  Requested  7m28s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-q6bd7"
  Normal  Requested  7m12s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-jd2cx"
  Normal  Requested  7m11s                  cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-74dr5"
  Normal  Requested  7m7s                   cert-manager-certificates-request-manager  Created new CertificateRequest resource "sample-sdb-server-cert-4k2wf"
  Normal  Reused     5m7s (x22 over 7m39s)  cert-manager-certificates-key-manager      Reusing private key stored in existing Secret resource "sample-sdb-server-cert"
  Normal  Issuing    5m7s (x23 over 23m)    cert-manager-certificates-issuing          The certificate has been successfully issued
  Normal  Requested  5m7s (x13 over 7m6s)   cert-manager-certificates-request-manager  (combined from similar events): Created new CertificateRequest resource "sample-sdb-server-cert-qn8g9"

```

We want to add `subject` and `emailAddresses` in the spec of server sertificate.

### Create SingleStoreOpsRequest

Below is the YAML of the `SingleStoreOpsRequest` CRO that we are going to create ton update the server certificate,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-update-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-sdb
  tls:
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      emailAddresses:
      - "kubedb@appscode.com"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the changes that we want in certificate objects.
- `spec.tls.certificates[].alias` specifies the certificate type which is one of these: `server`, `client`.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-tls/cluster/examples/sdbops-update-tls.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-update-tls created

```

#### Verify certificate is updated successfully

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CRO,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                TYPE             STATUS       AGE
sdbops-update-tls   ReconfigureTLS   Successful   3m24s


```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded.

Now, Let's exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ openssl x509 -in /etc/memsql/certs/server.crt -inform PEM  -subject -email -nameopt RFC2253 -noout
subject=CN=sample-sdb,O=kubedb:server
kubedb@appscode.com
```

We can see from the above output that, the subject name and email address match with the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a SingleStoreOpsRequest.

### Create SingleStoreOpsRequest

Below is the YAML of the `SingleStoreOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-sdb
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-sdb` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-tls/cluster/examples/sdbops-remove-tls.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-remove-tls created
```

#### Verify TLS Removed Successfully

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CRO,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                TYPE             STATUS       AGE
sdbops-remove-tls   ReconfigureTLS   Successful   27m
```

We can see from the above output that the `SingleStoreOpsRequest` has succeeded. If we describe the `SingleStoreOpsRequest` we will get an overview of the steps that were followed.

Now, Let's exec into the database and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo sample-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sample-sdb-aggregator-0 /]$ ls etc/memsql/
memsql_exporter.cnf  memsqlctl.hcl
[memsql@sample-sdb-aggregator-0 /]$ 
[memsql@sample-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 840
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show variables like '%ssl%';
+---------------------------------+------------+
| Variable_name                   | Value      |
+---------------------------------+------------+
| default_user_require_ssl        | OFF        |
| exporter_ssl_ca                 |            |
| exporter_ssl_capath             |            |
| exporter_ssl_cert               |            |
| exporter_ssl_key                |            |
| exporter_ssl_key_passphrase     | [redacted] |
| have_openssl                    | OFF        |
| have_ssl                        | OFF        |
| jwks_ssl_ca_certificate         |            |
| node_replication_ssl_only       | OFF        |
| openssl_version                 | 805306480  |
| processlist_rpc_json_max_size   | 2048       |
| ssl_ca                          |            |
| ssl_capath                      |            |
| ssl_cert                        |            |
| ssl_cipher                      |            |
| ssl_fips_mode                   | OFF        |
| ssl_key                         |            |
| ssl_key_passphrase              | [redacted] |
| ssl_last_reload_attempt_time    |            |
| ssl_last_successful_reload_time |            |
+---------------------------------+------------+
21 rows in set (0.00 sec)

singlestore> exit
Bye
```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sdb -n demo --all
$ kubectl delete issuer -n demo --all
$ kubectl delete singlestoreopsrequest -n demo --all
$ kubectl delete ns demo
```
