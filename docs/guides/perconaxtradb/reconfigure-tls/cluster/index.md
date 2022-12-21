---
title: Reconfigure PerconaXtraDB TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-reconfigure-tls-cluster
    name: Reconfigure PerconaXtraDB TLS/SSL Encryption
    parent: guides-perconaxtradb-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Reconfigure PerconaXtraDB TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing PerconaXtraDB database via a PerconaXtraDBOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes Cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.6.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Add TLS to a PerconaXtraDB Cluster

Here, We are going to create a PerconaXtraDB database without TLS and then reconfigure the database to use TLS.
> **Note:** Steps for reconfiguring TLS of PerconaXtraDB `Standalone` is same as PerconaXtraDB `Cluster`.

### Deploy PerconaXtraDB without TLS

In this section, we are going to deploy a PerconaXtraDB Cluster database without TLS. In the next few sections we will reconfigure TLS using `PerconaXtraDBOpsRequest` CRD. Below is the YAML of the `PerconaXtraDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: PerconaXtraDB
metadata:
  name: sample-pxc
  namespace: demo
spec:
  version: "8.0.26"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Let's create the `PerconaXtraDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure-tls/cluster/examples/sample-pxc.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

Now, wait until `sample-pxc` has status `Ready`. i.e,

```bash
$ kubectl get perconaxtradb -n demo
NAME             VERSION   STATUS   AGE
sample-pxc       8.0.26    Ready    9m17s
```

```bash
$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-pxc-auth -o jsonpath='{.data.\password}' | base64 -d
U6(h_pYrekLZ2OOd

$ kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb  -- bash
root@sample-pxc-0:/  mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 108
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]>  show variables like '%ssl%';
+---------------------+-----------------------------+
| Variable_name       | Value                       |
+---------------------+-----------------------------+
| have_openssl        | YES                         |
| have_ssl            | DISABLED                    |
| ssl_ca              |                             |
| ssl_capath          |                             |
| ssl_cert            |                             |
| ssl_cipher          |                             |
| ssl_crl             |                             |
| ssl_crlpath         |                             |
| ssl_key             |                             |
| version_ssl_library | OpenSSL 1.1.1f  31 Mar 2020 |
+---------------------+-----------------------------+
10 rows in set (0.001 sec)

```

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=perconaxtradb/O=kubedb"
Generating a RSA private key
...........................................................................+++++
........................................................................................................+++++
writing new private key to './ca.key'
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls px-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/px-ca created
```

Now, we are going to create an `Issuer` using the `px-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: px-issuer
  namespace: demo
spec:
  ca:
    secretName: px-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}//docs/guides/perconaxtradb/reconfigure-tls/cluster/examples/issuer.yaml
issuer.cert-manager.io/px-issuer created
```

### Create PerconaXtraDBOpsRequest

In order to add TLS to the database, we have to create a `PerconaXtraDBOpsRequest` CRO with our created issuer. Below is the YAML of the `PerconaXtraDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-pxc
  tls:
    requireSSL: true
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: px-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `requireSSL` specifies that the clients connecting to the server are required to use secured connection.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/perconaxtradb/concepts/perconaxtradb/#spectls).

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure-tls/cluster/examples/pxops-add-tls.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CRO,

```bash
$ kubectl get perconaxtradbopsrequest --all-namespaces
NAMESPACE   NAME            TYPE             STATUS       AGE
demo        pxops-add-tls   ReconfigureTLS   Successful   6m6s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded.

Now, we are going to connect to the database for verifying the `PerconaXtraDB` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb -- bash
root@sample-pxc-0:/ ls /etc/mysql/certs/client
ca.crt  tls.crt  tls.key
root@sample-pxc-0:/ ls /etc/mysql/certs/server
ca.crt  tls.crt  tls.key
root@sample-pxc-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 58
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]> show variables like '%ssl%';
+---------------------+---------------------------------+
| Variable_name       | Value                           |
+---------------------+---------------------------------+
| have_openssl        | YES                             |
| have_ssl            | YES                             |
| ssl_ca              | /etc/mysql/certs/server/ca.crt  |
| ssl_capath          | /etc/mysql/certs/server         |
| ssl_cert            | /etc/mysql/certs/server/tls.crt |
| ssl_cipher          |                                 |
| ssl_crl             |                                 |
| ssl_crlpath         |                                 |
| ssl_key             | /etc/mysql/certs/server/tls.key |
| version_ssl_library | OpenSSL 1.1.1f  31 Mar 2020     |
+---------------------+---------------------------------+
10 rows in set (0.005 sec)

PerconaXtraDB [(none)]> show variables like '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.005 sec)

PerconaXtraDB [(none)]> quit;
Bye
```

We can see from the above output that, `have_ssl` is set to `ture`. So, database TLS is enabled successfully to this database.

> Note: Add or Update reconfigure TLS with with `RequireSSL=true` will create downtime of the database while `PerconaXtraDBOpsRequest` is in `Progressing` status.

## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb -- bash
root@sample-pxc-0:/ apt update
root@sample-pxc-0:/ apt install openssl
root@sample-pxc-0:/ openssl x509 -in /etc/mysql/certs/client/tls.crt -inform  PEM -enddate -nameopt RFC2253 -noout
notAfter=Apr 13 05:18:43 2022 GMT
```

So, the certificate will expire on this time `Apr 13 05:18:43 2022 GMT`.

### Create PerconaXtraDBOpsRequest

Now we are going to increase it using a PerconaXtraDBOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-pxc
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure-tls/cluster/examples/pxops-rotate-tls.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-rotate-tls created
```

#### Verify Certificate Rotated Successfully

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CRO,

```bash
$ kubectl get perconaxtradbopsrequest --all-namespaces
NAMESPACE   NAME               TYPE             STATUS       AGE
demo        pxops-rotate-tls   ReconfigureTLS   Successful    3m
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb -- bash
root@sample-pxc-0:/ apt update
root@sample-pxc-0:/ apt install openssl
root@sample-pxc-0:/# openssl x509 -in /etc/mysql/certs/client/tls.crt -inform  PEM -enddate -nameopt RFC2253 -noout
notAfter=Apr 13 06:04:50 2022 GMT
```

As we can see from the above output, the certificate has been rotated successfully.

## Update Certificate

Now, we are going to update the server certificate.

- Let's describe the server certificate `sample-pxc-server-cert`
```bash
$ kubectl describe certificate -n demo sample-pxc-server-cert
Name:         sample-pxc-server-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=sample-pxc
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=perconaxtradbs.kubedb.com
Annotations:  <none>
API Version:  cert-manager.io/v1
Kind:         Certificate
Metadata:
  Creation Timestamp:  2022-01-13T05:18:42Z
  Generation:          1
  ...
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  PerconaXtraDB
    Name:                  sample-pxc
    UID:                   ed8f45c7-7caf-4890-8a9c-b8437b6ca48b
  Resource Version:        241340
  UID:                     3343e971-395d-46df-9536-47194eb96dcc
Spec:
  Common Name:  sample-pxc.demo.svc
  Dns Names:
    *.sample-pxc-pods.demo.svc
    *.sample-pxc-pods.demo.svc.cluster.local
    *.sample-pxc.demo.svc
    localhost
    sample-pxc
    sample-pxc.demo.svc
  Ip Addresses:
    127.0.0.1
  Issuer Ref:
    Group:      cert-manager.io
    Kind:       Issuer
    Name:       px-issuer
  Secret Name:  sample-pxc-server-cert
  Subject:
    Organizations:
      kubedb:server
  Usages:
    digital signature
    key encipherment
    server auth
    client auth
Status:
  Conditions:
    Last Transition Time:  2022-01-13T05:18:43Z
    Message:               Certificate is up to date and has not expired
    Observed Generation:   1
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2022-04-13T06:04:50Z
  Not Before:              2022-01-13T06:04:50Z
  Renewal Time:            2022-03-14T06:04:50Z
  Revision:                6
Events:
  Type    Reason     Age                From          Message
  ----    ------     ----               ----          -------
  Normal  Requested  22m                cert-manager  Created new CertificateRequest resource "sample-pxc-server-cert-8tnj5"
  Normal  Requested  22m                cert-manager  Created new CertificateRequest resource "sample-pxc-server-cert-fw6sk"
  Normal  Requested  22m                cert-manager  Created new CertificateRequest resource "sample-pxc-server-cert-cvphm"
  Normal  Requested  20m                cert-manager  Created new CertificateRequest resource "sample-pxc-server-cert-nvhp6"
  Normal  Requested  19m                cert-manager  Created new CertificateRequest resource "sample-pxc-server-cert-p5287"
  Normal  Reused     19m (x5 over 22m)  cert-manager  Reusing private key stored in existing Secret resource "sample-pxc-server-cert"
  Normal  Issuing    19m (x6 over 65m)  cert-manager  The certificate has been successfully issued
```

We want to add `subject` and `emailAddresses` in the spec of server sertificate.

### Create PerconaXtraDBOpsRequest

Below is the YAML of the `PerconaXtraDBOpsRequest` CRO that we are going to create ton update the server certificate,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-update-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-pxc
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

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the changes that we want in certificate objects.
- `spec.tls.certificates[].alias` specifies the certificate type which is one of these: `server`, `client`, `metrics-exporter`.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure-tls/cluster/examples/pxops-update-tls.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-update-tls created
```

#### Verify certificate is updated successfully

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CRO,

```bash
$ kubectl get perconaxtradbopsrequest -n demo
Every 2.0s: kubectl get perconaxtradbopsrequest -n demo
NAME                  TYPE             STATUS        AGE
pxops-update-tls   ReconfigureTLS     Successful      7m

```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded.

Now, Let's exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ kubectl exec -it -n demo sample-pxc-0  -c perconaxtradb -- bash
root@sample-pxc-0:/ apt update
root@sample-pxc-0:/ apt install openssl
root@sample-pxc-0:/ openssl x509 -in /etc/mysql/certs/server/tls.crt -inform PEM  -subject -email -nameopt RFC2253 -noout
subject=CN=sample-pxc.demo.svc,O=kubedb:server
kubedb@appscode.com
```

We can see from the above output that, the subject name and email address match with the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a PerconaXtraDBOpsRequest.

### Create PerconaXtraDBOpsRequest

Below is the YAML of the `PerconaXtraDBOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PerconaXtraDBOpsRequest
metadata:
  name: pxops-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: sample-pxc
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `sample-pxc` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `PerconaXtraDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/perconaxtradb/reconfigure-tls/cluster/examples/pxops-remove-tls.yaml
perconaxtradbopsrequest.ops.kubedb.com/pxops-remove-tls created
```

#### Verify TLS Removed Successfully

Let's wait for `PerconaXtraDBOpsRequest` to be `Successful`.  Run the following command to watch `PerconaXtraDBOpsRequest` CRO,

```bash
$ kubectl get perconaxtradbopsrequest --all-namespaces
NAMESPACE   NAME               TYPE             STATUS       AGE
demo        pxops-remove-tls   ReconfigureTLS   Successful   6m27s
```

We can see from the above output that the `PerconaXtraDBOpsRequest` has succeeded. If we describe the `PerconaXtraDBOpsRequest` we will get an overview of the steps that were followed.

Now, Let's exec into the database and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo sample-pxc-0 -c perconaxtradb  -- bash
root@sample-pxc-0:/  mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the PerconaXtraDB monitor.  Commands end with ; or \g.
Your PerconaXtraDB connection id is 108
Server version: 8.0.26-PerconaXtraDB-1:8.0.26+maria~focal perconaxtradb.org binary distribution

Copyright (c) 2000, 2018, Oracle, PerconaXtraDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

PerconaXtraDB [(none)]>  show variables like '%ssl%';
+---------------------+-----------------------------+
| Variable_name       | Value                       |
+---------------------+-----------------------------+
| have_openssl        | YES                         |
| have_ssl            | DISABLED                    |
| ssl_ca              |                             |
| ssl_capath          |                             |
| ssl_cert            |                             |
| ssl_cipher          |                             |
| ssl_crl             |                             |
| ssl_crlpath         |                             |
| ssl_key             |                             |
| version_ssl_library | OpenSSL 1.1.1f  31 Mar 2020 |
+---------------------+-----------------------------+
10 rows in set (0.001 sec)

```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete perconaxtradb -n demo --all
$ kubectl delete issuer -n demo --all
$ kubectl delete perconaxtradbopsrequest -n demo --all
$ kubectl delete ns demo
```
