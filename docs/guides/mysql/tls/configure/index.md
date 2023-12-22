---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-tls-configure
    name: MySQL TLS/SSL Configuration
    parent: guides-mysql-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in MySQL

`KubeDB` supports providing TLS/SSL encryption (via, `requireSSL` mode) for `MySQL`. This tutorial will show you how to use `KubeDB` to deploy a `MySQL` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/mysql/tls/configure/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Deploy MySQL database with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a MySQL standalone and a group replication that will be configured with these certificates by `KubeDB` operator.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mysql/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls my-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/my-ca created
```

Now, we are going to create an `Issuer` using the `my-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: mysql-issuer
  namespace: demo
spec:
  ca:
    secretName: my-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/issuer.yaml
issuer.cert-manager.io/mysql-issuer created
```

## Deploy MySQL with TLS/SSL configuration

<ul class="nav nav-tabs" id="definationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="gr-tab" data-toggle="tab" href="#groupReplication" role="tab" aria-controls="groupReplication" aria-selected="true">Group Replication</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="ic-tab" data-toggle="tab" href="#innodbCluster" role="tab" aria-controls="innodbCluster" aria-selected="false">Innodb Cluster</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="sc-tab" data-toggle="tab" href="#semisync" role="tab" aria-controls="semisync" aria-selected="false">Semi sync </a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="st-tab" data-toggle="tab" href="#standAlone" role="tab" aria-controls="standAlone" aria-selected="false">Stand Alone</a>
  </li>
</ul>


<div class="tab-content" id="definationTabContent">
  <div class="tab-pane fade show active" id="groupReplication" role="tabpanel" aria-labelledby="gr-tab">

Now, we are going to deploy a `MySQL` group replication with TLS/SSL configuration. Below is the YAML for MySQL group replication that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: some-mysql
  namespace: demo
spec:
  version: "8.0.35"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
  terminationPolicy: WipeOut
```

**Deploy MySQL group replication:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/group-replication.yaml
mysql.kubedb.com/some-mysql created
```

  </div>

  <div class="tab-pane fade" id="innodbCluster" role="tabpanel" aria-labelledby="ic-tab">

Now, we are going to deploy a `MySQL` Innodb with TLS/SSL configuration. Below is the YAML for MySQL innodb cluster that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: some-mysql
  namespace: demo
spec:
  version: "8.0.31-innodb"
  replicas: 3
  topology:
    mode: InnoDBCluster
    innoDBCluster:
      router:
        replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  terminationPolicy: WipeOut
```

**Deploy MySQL Innodb Cluster:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/innodb.yaml
mysql.kubedb.com/some-mysql created
```
  </div>

  <div class="tab-pane fade " id="semisync" role="tabpanel" aria-labelledby="sc-tab">
Now, we are going to deploy a `MySQL` Semi sync cluster with TLS/SSL configuration. Below is the YAML for MySQL semi-sync cluster that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: some-mysql
  namespace: demo
spec:
  version: "8.0.35"
  replicas: 3
  topology:
    mode: SemiSync
    semiSync:
      sourceWaitForReplicaCount: 1
      sourceTimeout: 24h
      errantTransactionRecoveryPolicy: PseudoTransaction
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  terminationPolicy: WipeOut
```

**Deploy MySQL Semi-sync:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/semi-sync.yaml
mysql.kubedb.com/some-mysql created
```

  </div>


  <div class="tab-pane fade" id="standAlone" role="tabpanel" aria-labelledby="st-tab">

Now, we are going to deploy a stand alone `MySQL` with TLS/SSL configuration. Below is the YAML for stand alone MySQL that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: some-mysql
  namespace: demo
spec:
  version: "8.0.35"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mysql-issuer
    certificates:
      - alias: server
        subject:
          organizations:
            - kubedb:server
        dnsNames:
          - localhost
        ipAddresses:
          - "127.0.0.1"
  terminationPolicy: WipeOut
```

**Deploy Stand Alone MySQL:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/standalone.yaml
mysql.kubedb.com/some-mysql created
```
  </div>

</div>



**Wait for the database to be ready :**

Now, watch `MySQL` is going to `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get my -n demo some-mysql
Every 3.0s: kubectl get my -n demo some-mysql                 suaas-appscode: Thu Aug 13 19:02:15 2020

NAME           VERSION   STATUS    AGE
some-mysql   8.0.35    Running   9m41s

$ watch -n 3 kubectl get sts -n demo some-mysql
Every 3.0s: kubectl get sts -n demo some-mysql                suaas-appscode: Thu Aug 13 19:02:42 2020

NAME           READY   AGE
some-mysql   3/3     9m51s

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=some-mysql
Every 3.0s: kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com  suaas-appscode: Thu Aug 13 19:03:02 2020

NAME             READY   STATUS    RESTARTS   AGE
some-mysql-0   2/2     Running   0          10m
some-mysql-1   2/2     Running   0          4m4s
some-mysql-2   2/2     Running   0          2m3s
```

**Verify tls-secrets created successfully :**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` Ops Manager. Default tls-secret name formed as _{mysql-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep "some-mysql"
some-mysql-client-cert                    kubernetes.io/tls                     3      13m
some-mysql-auth                           Opaque                                2      13m
some-mysql-metrics-exporter-cert          kubernetes.io/tls                     3      13m
some-mysql-metrics-exporter-config        Opaque                                1      13m
some-mysql-server-cert                    kubernetes.io/tls                     3      13m
some-mysql-token-49sjm                    kubernetes.io/service-account-token   3      13m
```

**Verify MySQL Standalone configured to TLS/SSL:**

Now, we are going to connect to the database for verifying the `MySQL` group replication has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n  demo  some-mysql-0 -c mysql -- bash
root@my-group-0:/# ls /etc/mysql/certs/
ca.crt  client.crt  client.key  server.crt  server.key

root@my-group-0:/# mysql -u${MYSQL_ROOT_USERNAME} -p{MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 27
Server version: 8.0.23 MySQL Community Server - GPL

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW VARIABLES LIKE '%ssl%';
+---------------------------------------------------+-----------------------------+
| Variable_name                                     | Value                       |
+---------------------------------------------------+-----------------------------+
| admin_ssl_ca                                      |                             |
| admin_ssl_capath                                  |                             |
| admin_ssl_cert                                    |                             |
| admin_ssl_cipher                                  |                             |
| admin_ssl_crl                                     |                             |
| admin_ssl_crlpath                                 |                             |
| admin_ssl_key                                     |                             |
| group_replication_recovery_ssl_ca                 |                             |
| group_replication_recovery_ssl_capath             |                             |
| group_replication_recovery_ssl_cert               |                             |
| group_replication_recovery_ssl_cipher             |                             |
| group_replication_recovery_ssl_crl                |                             |
| group_replication_recovery_ssl_crlpath            |                             |
| group_replication_recovery_ssl_key                |                             |
| group_replication_recovery_ssl_verify_server_cert | OFF                         |
| group_replication_recovery_use_ssl                | ON                          |
| group_replication_ssl_mode                        | REQUIRED                    |
| have_openssl                                      | YES                         |
| have_ssl                                          | YES                         |
| mysqlx_ssl_ca                                     |                             |
| mysqlx_ssl_capath                                 |                             |
| mysqlx_ssl_cert                                   |                             |
| mysqlx_ssl_cipher                                 |                             |
| mysqlx_ssl_crl                                    |                             |
| mysqlx_ssl_crlpath                                |                             |
| mysqlx_ssl_key                                    |                             |
| ssl_ca                                            | /etc/mysql/certs/ca.crt     |
| ssl_capath                                        | /etc/mysql/certs            |
| ssl_cert                                          | /etc/mysql/certs/server.crt |
| ssl_cipher                                        |                             |
| ssl_crl                                           |                             |
| ssl_crlpath                                       |                             |
| ssl_fips_mode                                     | OFF                         |
| ssl_key                                           | /etc/mysql/certs/server.key |
+---------------------------------------------------+-----------------------------+
34 rows in set (0.02 sec)


mysql> SHOW VARIABLES LIKE '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

The above output shows that the `MySQL` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in the `/etc/ mysql/certs/` directory for client and server.

**Verify secure connection for SSL required user:**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```bash
# creating SSL required user
$ kubectl exec -it -n  demo  some-mysql-0 -c mysql -- bash

root@my-group-0:/# mysql -uroot -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 27
Server version: 8.0.23 MySQL Community Server - GPL

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE USER 'mysql_user'@'localhost' IDENTIFIED BY 'pass' REQUIRE SSL;
Query OK, 0 rows affected (0.01 sec)

mysql> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye

# accessing database server with newly created user
root@my-group-0:/# mysql -umysql_user -ppass
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)

# accessing the database server newly created user with ssl-mode=disable
root@my-group-0:/# mysql -umysql_user -ppass --ssl-mode=disabled
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)

# accessing the database server newly created user with certificates
root@my-group-0:/# mysql -umysql_user -ppass --ssl-ca=/etc/mysql/certs/ca.crt  --ssl-cert=/etc/mysql/certs/client.crt --ssl-key=/etc/mysql/certs/client.key
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 384
Server version: 5.7.29-log MySQL Community Server (GPL)

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

You are enforcing ssl connection via unix socket. Please consider
switching ssl off as it does not make connection via unix socket
any more secure.
mysql> exit
Bye
```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/` directory.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete my -n demo  my-standalone-tls
kubectl delete my -n demo  some-mysql
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).