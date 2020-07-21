---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: my-tls-encryption
    name: TLS/SSL (Transport Encryption)
    parent: my-tls
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

# Run MySQL with TLS/SSL (Transport Encryption)

`KubeDB` supports providing TLS/SSL encryption (via, `requireSSL` mode) for `MySQL`. This tutorial will show you how to use `KubeDB` (both community and enterprise operator) to run a `MySQL` database with TLS/SSL configuration.

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


## Create Issuer/ ClusterIssuer

We are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating you ca-certificates using openssl.

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mongo/O=kubedb"
```

- Now create a ca-secret using the certificate files you have just generated.

```bash
kubectl create secret tls my-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/my-ca created
```

Now, create an `Issuer` using the `ca-secret` you have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1alpha2
kind: Issuer
metadata:
  name: mysql-issuer
  namespace: demo
spec:
  ca:
    secretName: my-ca
```

Apply the `YAML` file:

```bash
kubectl apply -f ./docs/examples/my/day-2-operations/issuer.yaml
issuer.cert-manager.io/mysql-issuer created
```

## TLS/SSL configuration in MySQL Standalone

Below is the YAML for MySQL Standalone that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: mysql-quickstart
  namespace: demo
spec:
  version: "5.7.29"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/builtin
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io/v1alpha2
      kind: Issuer
      name: mysql-issuer
    certificate:
      organization:
      - kubedb:server
      dnsNames:
      - localhost
      - "127.0.0.1"
  terminationPolicy: WipeOut
```

Here,

- [`spec.requireSSL`](/docs/concepts/databases/mysql.md#requireSSL) specifies SSL/TLS client connections to the server are required or not.

- `spec.tls.issuerRef.apiGroup` specifies the group name of the resource being referenced. The value for `Issuer` or `ClusterIssuer` is `cert-manager.io` (cert-manager v0.12.0 and later).

- `spec.tls.issuerRef.kind` specifies the type of resource being referenced. KubeDB supports both of `Issuer` and `ClusterIssuer` as values for this field.

- `spec.tls.issuerRef.name` is the name of resource (`Issuer` or `ClusterIssuer`) being referenced.

- `spec.tls.certificate` gives you a lots of option to configure so that the certificate will be renewed and kept up to date. More details [here](/docs/concepts/databases/mysql.md#tls)

**Deploy MySQL Standalone :**

```console
$ kubectl apply -f ./docs/examples/day-2-operations/tls-standalone.yaml
mysql.kubedb.com/mysql-quickstart created
```

Now, wait until `mysql-quickstart` has status `Running`. i.e,

```console
$ kubectl get my -n demo mysql-quickstart
NAME               VERSION   STATUS    AGE
mysql-quickstart   5.7.29    Running   3m31s
```

**Verify tls-secrets created :**

If everything goes well, you can see that our tls-secrets will be created which contains server and client certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by KubeDB enterprise operator according to the following format,

- _{mysql-object-name}-server-cert_
- _{mysql-object-name}-exporter-cert_
- _{mysql-object-name}-client-cert_

Let's check tls-secret,

```console
$ kubectl get secrets -n demo | grep "mysql-quickstart"
mysql-quickstart-auth            Opaque                                2      16m
mysql-quickstart-client-cert     kubernetes.io/tls                     3      16m
mysql-quickstart-exporter-cert   kubernetes.io/tls                     3      16m
mysql-quickstart-server-cert     kubernetes.io/tls                     3      16m
mysql-quickstart-token-wjn22     kubernetes.io/service-account-token   3      16m
```

**Verify MySQL Standalone configured to TLS/SSL :**

Now, we are going to connect to the database for verifying the `MySQL` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```
$ kubectl exec -it -n  demo  mysql-quickstart-0 -c mysql -- sh
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
+---------------+-----------------------------+
| Variable_name | Value                       |
+---------------+-----------------------------+
| have_openssl  | YES                         |
| have_ssl      | YES                         |
| ssl_ca        | /etc/mysql/certs/ca.crt     |
| ssl_capath    | /etc/mysql/certs            |
| ssl_cert      | /etc/mysql/certs/server.crt |
| ssl_cipher    |                             |
| ssl_crl       |                             |
| ssl_crlpath   |                             |
| ssl_key       | /etc/mysql/certs/server.key |
+---------------+-----------------------------+
9 rows in set (0.00 sec)

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
$ kubectl exec -it -n  demo  mysql-quickstart-0 -c mysql -- bash
root@mysql-quickstart-0:/# mysql -uroot -p${MYSQL_ROOT_PASSWORD}
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

# accessing database server with newly created user
root@mysql-quickstart-0:/# mysql -umysql_user -ppass
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)


root@mysql-quickstart-0:/# mysql -umysql_user -ppass --ssl-mode=disabled
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)


root@mysql-quickstart-0:/# mysql -umysql_user -ppass --ssl-ca=/etc/mysql/certs/ca.crt  --ssl-cert=/etc/mysql/certs/client.crt --ssl-key=/etc/mysql/certs/client.key
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

## TLS/SSL configuration in MySQL GroupReplication

Below is the YAML for MySQL group replication that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha1
kind: MySQL
metadata:
  name: mysql-quickstart
  namespace: demo
spec:
  version: "5.7.29"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  monitor:
    agent: prometheus.io/builtin
  requireSSL: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io/v1alpha2
      kind: Issuer
      name: mysql-issuer
    certificate:
      organization:
      - kubedb:server
      dnsNames:
      - localhost
      - "127.0.0.1"
  terminationPolicy: WipeOut
```

**Deploy MySQL group replication :**

```console
$ kubectl apply -f ./docs/examples/day-2-operations/tls-group.yaml
mysql.kubedb.com/my-group created
```

Now, wait until `my-group` has the status `Running`. i.e,

```console
$ kubectl get my -n demo my-group
NAME               VERSION   STATUS    AGE
mysql-quickstart   5.7.29    Running   3m31s
```

**Verify tls-secrets created :**

If everything goes well, you can see that our tls-secrets will be created which contains server and client certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by KubeDB enterprise operator according to the following format,

- _{mysql-object-name}-server-cert_
- _{mysql-object-name}-exporter-cert_
- _{mysql-object-name}-client-cert_

Let's check tls-secret,

```console
$ kubectl get secrets -n demo | grep "my-group"
my-group-auth            Opaque                                2      3m8s
my-group-client-cert     kubernetes.io/tls                     3      3m7s
my-group-exporter-cert   kubernetes.io/tls                     3      3m7s
my-group-server-cert     kubernetes.io/tls                     3      3m7s
my-group-token-zj7pq     kubernetes.io/service-account-token   3      3m8s
```

**Verify MySQL Standalone configured to TLS/SSL :**

Now, we are going to connect to the database for verifying the `MySQL` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```console
$ kubectl exec -it -n  demo  my-group-0 -c mysql -- bash
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
| ssl_ca                                            | /etc/mysql/certs/ca.crt     |
| ssl_capath                                        | /etc/mysql/certs            |
| ssl_cert                                          | /etc/mysql/certs/server.crt |
| ssl_cipher                                        |                             |
| ssl_crl                                           |                             |
| ssl_crlpath                                       |                             |
| ssl_key                                           | /etc/mysql/certs/server.key |
+---------------------------------------------------+-----------------------------+
19 rows in set (0.00 sec)

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
kubectl exec -it -n  demo  my-group-0 -c mysql -- bash
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


root@my-group-0:/# mysql -umysql_user -ppass --ssl-mode=disabled
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'mysql_user'@'localhost' (using password: YES)


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

You are enforcing ssl conection via unix socket. Please consider
switching ssl off as it does not make connection via unix socket
any more secure.
mysql> exit
Bye
```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/` directory.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```console
kubectl delete my -n demo  mysql-quickstart
kubectl delete my -n demo  my-group
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/concepts/databases/mysql.md).