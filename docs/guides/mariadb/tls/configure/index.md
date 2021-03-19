---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-tls-configure
    name: MariaDB TLS/SSL Configuration
    parent: guides-mariadb-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Configure TLS/SSL in MariaDB

`KubeDB` supports providing TLS/SSL encryption (via, `requireSSL` mode) for `MariaDB`. This tutorial will show you how to use `KubeDB` to deploy a `MariaDB` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/mariadb/tls/configure/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Deploy MariaDB database with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a MariaDB standalone and a group replication that will be configured with these certificates by `KubeDB` operator.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mariadb/O=kubedb"
Generating a RSA private key
...........................................................................+++++
........................................................................................................+++++
writing new private key to './ca.key'
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls md-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/md-ca created
```

Now, we are going to create an `Issuer` using the `md-ca` secret that hols the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1beta1
kind: Issuer
metadata:
  name: md-issuer
  namespace: demo
spec:
  ca:
    secretName: md-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/tls/configure/examples/issuer.yaml
issuer.cert-manager.io/md-issuer created
```

### Deploy MariaDB Standalone with TLS/SSL configuration

Here, our issuer `md-issuer`  is ready to deploy a `MariaDB` standalone with TLS/SSL configuration. Below is the YAML for MariaDB Standalone that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: "10.5.8"
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
      name: md-issuer
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

- `spec.tls.issuerRef` refers to the `md-issuer` issuer.

- `spec.tls.certificates` gives you a lot of options to configure so that the certificate will be renewed and kept up to date. 
You can found more details from [here](/docs/guides/mariadb/concepts/mariadb/#spectls)

**Deploy MariaDB Standalone:**

Let’s create the `MariaDB` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/tls/configure/examples/tls-standalone.yaml
mariadb.kubedb.com/sample-mariadb created
```

**Wait for the database to be ready:**

Now, wait for `MariaDB` going on `Running` state and also wait for `StatefulSet` and its pod to be created and going to `Running` state,

```bash
$ kubectl get mariadb -n demo sample-mariadb
NAME             VERSION   STATUS   AGE
sample-mariadb   10.5.8    Ready    5m48s

$ kubectl get sts -n demo sample-mariadb
NAME             READY   AGE
sample-mariadb   1/1     7m5s
```

**Verify tls-secrets created successfully:**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` enterprise operator. Default tls-secret name formed as _{mysql-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep sample-mariadb
sample-mariadb-archiver-cert             kubernetes.io/tls                     3      7m53s
sample-mariadb-auth                      kubernetes.io/basic-auth              2      7m54s
sample-mariadb-metrics-exporter-cert     kubernetes.io/tls                     3      7m53s
sample-mariadb-metrics-exporter-config   Opaque                                1      7m54s
sample-mariadb-server-cert               kubernetes.io/tls                     3      7m53s
sample-mariadb-token-7hhg2
```

**Verify MariaDB Standalone configured with TLS/SSL:**

Now, we are going to connect to the database for verifying the `MariaDB` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo sample-mariadb-0 -- bash

root@sample-mariadb-0:/ ls /etc/mysql/certs/client
ca.crt  tls.crt  tls.key
root@sample-mariadb-0:/ ls /etc/mysql/certs/server
ca.crt  tls.crt  tls.key

root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 64
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> show variables like '%ssl%';
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
10 rows in set (0.002 sec)

MariaDB [(none)]> show variables like '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.001 sec)

MariaDB [(none)]> quit;
Bye
```

The above output shows that the `MariaDB` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in `/etc/mysql/certs/client/` and `/etc/mysql/certs/server/` directory for client and server respectively.

**Verify secure connection for SSL required user:**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```bash
# creating SSL required user
$ kubectl exec -it -n  demo  my-standalone-tls-0 -- bash

root@mysql-tls-0:/# mysql -uroot -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 27
Server version: 8.0.23 MariaDB Community Server - GPL

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

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
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 47
Server version: 5.7.29 MariaDB Community Server (GPL)

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




$ kubectl exec -it -n demo sample-mariadb-0 -- bash
root@sample-mariadb-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 92
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE USER 'new_user'@'localhost' IDENTIFIED BY '1234' REQUIRE SSL;
Query OK, 0 rows affected (0.028 sec)

MariaDB [(none)]> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.000 sec)

MariaDB [(none)]> exit
Bye

#  accessing the database server with newly created user
root@sample-mariadb-0:/ mysql -unew_user -p1234
ERROR 1045 (28000): Access denied for user 'new_user'@'localhost' (using password: YES)

# accessing the database server newly created user with certificates
root@sample-mariadb-0:/ mysql -unew_user -p1234 --ssl-ca=/etc/mysql/certs/server/ca.crt  --ssl-cert=/etc/mysql/certs/server/tls.crt --ssl-key=/etc/mysql/certs/server/tls.key
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 116
Server version: 10.5.8-MariaDB-1:10.5.8+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> exit 
Bye
```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/client/` directory.

## Deploy MariaDB Group Replication with TLS/SSL configuration

Now, we are going to deploy a `MariaDB` group replication with TLS/SSL configuration. Below is the YAML for MariaDB group replication that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: my-group-tls
  namespace: demo
spec:
  version: "8.0.23"
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

**Deploy MariaDB group replication:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls/tls-group.yaml
mysql.kubedb.com/my-group-tls created
```

**Wait for the database to be ready :**

Now, watch `MariaDB` is going to `Running` state and also watch `StatefulSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get my -n demo my-group-tls
Every 3.0s: kubectl get my -n demo my-group-tls                 suaas-appscode: Thu Aug 13 19:02:15 2020

NAME           VERSION   STATUS    AGE
my-group-tls   8.0.23    Running   9m41s

$ watch -n 3 kubectl get sts -n demo my-group-tls
Every 3.0s: kubectl get sts -n demo my-group-tls                suaas-appscode: Thu Aug 13 19:02:42 2020

NAME           READY   AGE
my-group-tls   3/3     9m51s

$ watch -n 3 kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group-tls
Every 3.0s: kubectl get pod -n demo -l app.kubernetes.io/name=mysqls.kubedb.com  suaas-appscode: Thu Aug 13 19:03:02 2020

NAME             READY   STATUS    RESTARTS   AGE
my-group-tls-0   2/2     Running   0          10m
my-group-tls-1   2/2     Running   0          4m4s
my-group-tls-2   2/2     Running   0          2m3s
```

**Verify tls-secrets created successfully :**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` enterprise operator. Default tls-secret name formed as _{mysql-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep "my-group-tls"
my-group-tls-client-cert                  kubernetes.io/tls                     3      13m
my-group-tls-auth                           Opaque                                2      13m
my-group-tls-metrics-exporter-cert          kubernetes.io/tls                     3      13m
my-group-tls-metrics-exporter-config        Opaque                                1      13m
my-group-tls-server-cert                    kubernetes.io/tls                     3      13m
my-group-tls-token-49sjm                    kubernetes.io/service-account-token   3      13m
```

**Verify MariaDB Standalone configured to TLS/SSL:**

Now, we are going to connect to the database for verifying the `MariaDB` group replication has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n  demo  my-group-tls-0 -c mysql -- bash
root@my-group-0:/# ls /etc/mysql/certs/
ca.crt  client.crt  client.key  server.crt  server.key

root@my-group-0:/# mysql -u${MYSQL_ROOT_USERNAME} -p{MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 27
Server version: 8.0.23 MariaDB Community Server - GPL

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

The above output shows that the `MariaDB` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in the `/etc/ mysql/certs/` directory for client and server.

**Verify secure connection for SSL required user:**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```bash
# creating SSL required user
$ kubectl exec -it -n  demo  my-group-tls-0 -c mysql -- bash

root@my-group-0:/# mysql -uroot -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 27
Server version: 8.0.23 MariaDB Community Server - GPL

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
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 384
Server version: 5.7.29-log MariaDB Community Server (GPL)

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
kubectl delete my -n demo  my-group-tls
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MariaDB object](/docs/guides/mysql/concepts/database/index.md).