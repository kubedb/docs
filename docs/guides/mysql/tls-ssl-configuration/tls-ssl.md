---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: my-tls-encryption
    name: MySQL TLS/SSL Configuration
    parent: my-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Configure TLS/SSL in MySQL

`KubeDB` supports providing TLS/SSL encryption (via, `requireSSL` mode) for `MySQL`. This tutorial will show you how to use `KubeDB` to deploy a `MySQL` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `cert-manger operator` (v0.12.0 or later) to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here]().

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```console
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/mysql](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mysql) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

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
apiVersion: cert-manager.io/v1beta1
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
kubectl apply -f ./docs/examples/my/day-2-operations/mysql/issuer.yaml
issuer.cert-manager.io/mysql-issuer created
```

### Deploy MySQL Standalone with TLS/SSL configuration

Here, our issuer `mysql-issuer`  is ready to deploy a `MySQL` standalone with TLS/SSL configuration. Below is the YAML for MySQL Standalone that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-standalone-tls
  namespace: demo
spec:
  version: "8.0.21"
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
      apiGroup: cert-manager.io/v1beta1
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

Here,

- `spec.requireSSL` specifies the SSL/TLS client connection to the server is required.  

- `spec.tls.issuerRef` refers to the `mysql-issuer` issuer.

- `spec.tls.certificates` gives you a lots of option to configure so that the certificate will be renewed and kept up to date. 
You can found more details from [here](/docs/concepts/databases/mysql.md#tls)

**Deploy MySQL Standalone :**

Let’s create the `MySQL` cr we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/tls-ssl/tls-standalone.yaml
mysql.kubedb.com/my-standalone-tls created
```

**Wait for the database to be ready :**

Now, watch `MySQL` is going to `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```console
$ watch -n 3 kubectl get my -n demo my-standalone-tls
Every 3.0s: kubectl get my -n demo my-standalone-tls            suaas-appscode: Thu Aug 13 18:12:39 2020

NAME                VERSION   STATUS    AGE
my-standalone-tls   8.0.21    Running   7m5s

$ watch -n 3 kubectl get sts -n demo my-standalone-tls
Every 3.0s: kubectl get sts -n demo my-standalone-tls            suaas-appscode: Thu Aug 13 18:12:59 2020

NAME                READY   AGE
my-standalone-tls   1/1     7m15s

$ watch -n 3 kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-standalone-tls
Every 3.0s: kubectl get pod -n demo -l kubedb.com/kind=MySQL...  suaas-appscode: Thu Aug 13 18:13:19 2020

NAME                  READY   STATUS    RESTARTS   AGE
my-standalone-tls-0   1/1     Running   0          7m35s
```

**Verify tls-secrets created successfully :**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` enterprise operator. Default tls-secret name formed as _{mysql-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```console
$ kubectl get secrets -n demo | grep "my-standalone-tls"
my-standalone-tls-archiver-cert             kubernetes.io/tls                     3      33m
my-standalone-tls-auth                      Opaque                                2      33m
my-standalone-tls-metrics-exporter-cert     kubernetes.io/tls                     3      33m
my-standalone-tls-metrics-exporter-config   Opaque                                1      33m
my-standalone-tls-server-cert               kubernetes.io/tls                     3      33m
my-standalone-tls-token-rkjd2               kubernetes.io/service-account-token   3      33m
```

**Verify MySQL Standalone configured with TLS/SSL :**

Now, we are going to connect to the database for verifying the `MySQL` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```console
$ kubectl exec -it -n  demo  my-standalone-tls-0 -- bash
# ls /etc/mysql/certs/
ca.crt  client.crt  client.key  server.crt  server.key

# mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 356
Server version: 5.7.29 MySQL Community Server (GPL)

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql>  SHOW VARIABLES LIKE '%ssl%';
+--------------------+-----------------------------+
| Variable_name      | Value                       |
+--------------------+-----------------------------+
| admin_ssl_ca       |                             |
| admin_ssl_capath   |                             |
| admin_ssl_cert     |                             |
| admin_ssl_cipher   |                             |
| admin_ssl_crl      |                             |
| admin_ssl_crlpath  |                             |
| admin_ssl_key      |                             |
| have_openssl       | YES                         |
| have_ssl           | YES                         |
| mysqlx_ssl_ca      |                             |
| mysqlx_ssl_capath  |                             |
| mysqlx_ssl_cert    |                             |
| mysqlx_ssl_cipher  |                             |
| mysqlx_ssl_crl     |                             |
| mysqlx_ssl_crlpath |                             |
| mysqlx_ssl_key     |                             |
| ssl_ca             | /etc/mysql/certs/ca.crt     |
| ssl_capath         | /etc/mysql/certs            |
| ssl_cert           | /etc/mysql/certs/server.crt |
| ssl_cipher         |                             |
| ssl_crl            |                             |
| ssl_crlpath        |                             |
| ssl_fips_mode      | OFF                         |
| ssl_key            | /etc/mysql/certs/server.key |
+--------------------+-----------------------------+
24 rows in set (0.00 sec)

mysql> SHOW VARIABLES LIKE '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.01 sec)

mysql> exit
Bye
```

The above output shows that `MySQL` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in the `/etc/ mysql/certs/` directory for client and server.

**Verify secure connection for SSL required user :**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```console
# creating SSL required user
$ kubectl exec -it -n  demo  my-standalone-tls-0 -- bash

root@mysql-tls-0:/# mysql -uroot -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 10
Server version: 5.7.29 MySQL Community Server (GPL)

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE USER 'mysql_user'@'localhost' IDENTIFIED BY 'pass' REQUIRE SSL;
Query OK, 0 rows affected (0.00 sec)

mysql> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye

# accessing the database server with newly created user
root@mysql-tls-0:/# mysql -umysql_user -ppass
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)

# accessing the database server newly created user with ssl-mode=disable
root@mysql-tls-0:/# mysql -umysql_user -ppass --ssl-mode=disabled
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)

# accessing the database server newly created user with certificates
root@mysql-tls-0:/# mysql -umysql_user -ppass --ssl-ca=/etc/mysql/certs/ca.crt  --ssl-cert=/etc/mysql/certs/client.crt --ssl-key=/etc/mysql/certs/client.key
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 47
Server version: 5.7.29 MySQL Community Server (GPL)

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

You are enforcing ssl conection via unix socket. Please consider
switching ssl off as it does not make connection via unix socket
any more secure.
mysql> exit
Bye
```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/` directory.

## Deploy MySQL Group Replication with TLS/SSL configuration

Now, we are going to deploy a `MySQL` group replication with TLS/SSL configuration. Below is the YAML for MySQL group replication that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: my-group-tls
  namespace: demo
spec:
  version: "8.0.21"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      name: "dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"
      baseServerID: 100
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
      apiGroup: cert-manager.io/v1beta1
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

**Deploy MySQL group replication :**

```console
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/day-2-operations/mysql/tls-ssl/tls-group.yaml
mysql.kubedb.com/my-group-tls created
```

**Wait for the database to be ready :**

Now, watch `MySQL` is going to `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```console
$ watch -n 3 kubectl get my -n demo my-group-tls
Every 3.0s: kubectl get my -n demo my-group-tls                 suaas-appscode: Thu Aug 13 19:02:15 2020

NAME           VERSION   STATUS    AGE
my-group-tls   8.0.21    Running   9m41s

$ watch -n 3 kubectl get sts -n demo my-group-tls
Every 3.0s: kubectl get sts -n demo my-group-tls                suaas-appscode: Thu Aug 13 19:02:42 2020

NAME           READY   AGE
my-group-tls   3/3     9m51s

$ watch -n 3 kubectl get pod -n demo -l kubedb.com/kind=MySQL,kubedb.com/name=my-group-tls
Every 3.0s: kubectl get pod -n demo -l kubedb.com/kind=MySQ...  suaas-appscode: Thu Aug 13 19:03:02 2020

NAME             READY   STATUS    RESTARTS   AGE
my-group-tls-0   2/2     Running   0          10m
my-group-tls-1   2/2     Running   0          4m4s
my-group-tls-2   2/2     Running   0          2m3s
```

**Verify tls-secrets created successfully :**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` enterprise operator. Default tls-secret name formed as _{mysql-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```console
$ kubectl get secrets -n demo | grep "my-group-tls"
my-group-tls-archiver-cert                  kubernetes.io/tls                     3      13m
my-group-tls-auth                           Opaque                                2      13m
my-group-tls-metrics-exporter-cert          kubernetes.io/tls                     3      13m
my-group-tls-metrics-exporter-config        Opaque                                1      13m
my-group-tls-server-cert                    kubernetes.io/tls                     3      13m
my-group-tls-token-49sjm                    kubernetes.io/service-account-token   3      13m
```

**Verify MySQL Standalone configured to TLS/SSL :**

Now, we are going to connect to the database for verifying the `MySQL` group replication has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```console
$ kubectl exec -it -n  demo  my-group-tls-0 -c mysql -- bash
root@my-group-0:/# ls /etc/mysql/certs/
ca.crt  client.crt  client.key  server.crt  server.key

root@my-group-0:/# mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 241
Server version: 5.7.29-log MySQL Community Server (GPL)

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

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

The above output shows that `MySQL` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in the `/etc/ mysql/certs/` directory for client and server.

**Verify secure connection for SSL required user :**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```console
# creating ssl required user
$ kubectl exec -it -n  demo  my-group-tls-0 -c mysql -- bash

root@my-group-0:/# mysql -uroot -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 360
Server version: 5.7.29-log MySQL Community Server (GPL)

Copyright (c) 2000, 2020, Oracle and/or its affiliates. All rights reserved.

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

```console
kubectl delete my -n demo  my-standalone-tls
kubectl delete my -n demo  my-group-tls
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/concepts/databases/mysql.md).