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

# Configure TLS/SSL in MariaDB

`KubeDB` supports providing TLS/SSL encryption (via, `requireSSL` mode) for `MariaDB`. This tutorial will show you how to use `KubeDB` to deploy a `MariaDB` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

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
apiVersion: cert-manager.io/v1
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
  name: md-standalone-tls
  namespace: demo
spec:
  version: "10.5.23"
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
  deletionPolicy: WipeOut
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
mariadb.kubedb.com/md-standalone-tls created
```

**Wait for the database to be ready:**

Now, wait for `MariaDB` going on `Running` state and also wait for `StatefulSet` and its pod to be created and going to `Running` state,

```bash
$ kubectl get mariadb -n demo md-standalone-tls
NAME             VERSION   STATUS   AGE
md-standalone-tls   10.5.23    Ready    5m48s

$ kubectl get sts -n demo md-standalone-tls
NAME             READY   AGE
md-standalone-tls   1/1     7m5s
```

**Verify tls-secrets created successfully:**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` Ops Manager. Default tls-secret name formed as _{mariadb-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep md-standalone-tls
md-standalone-tls-archiver-cert             kubernetes.io/tls                     3      7m53s
md-standalone-tls-auth                      kubernetes.io/basic-auth              2      7m54s
md-standalone-tls-metrics-exporter-cert     kubernetes.io/tls                     3      7m53s
md-standalone-tls-metrics-exporter-config   Opaque                                1      7m54s
md-standalone-tls-server-cert               kubernetes.io/tls                     3      7m53s
md-standalone-tls-token-7hhg2
```

**Verify MariaDB Standalone configured with TLS/SSL:**

Now, we are going to connect to the database for verifying the `MariaDB` server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo md-standalone-tls-0 -- bash

root@md-standalone-tls-0:/ ls /etc/mysql/certs/client
ca.crt  tls.crt  tls.key
root@md-standalone-tls-0:/ ls /etc/mysql/certs/server
ca.crt  tls.crt  tls.key

root@md-standalone-tls-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 64
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

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
$ kubectl exec -it -n demo md-standalone-tls-0 -- bash
root@md-standalone-tls-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 92
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE USER 'new_user'@'localhost' IDENTIFIED BY '1234' REQUIRE SSL;
Query OK, 0 rows affected (0.028 sec)

MariaDB [(none)]> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.000 sec)

MariaDB [(none)]> exit
Bye

#  accessing the database server with newly created user
root@md-standalone-tls-0:/ mysql -unew_user -p1234
ERROR 1045 (28000): Access denied for user 'new_user'@'localhost' (using password: YES)

# accessing the database server newly created user with certificates
root@md-standalone-tls-0:/ mysql -unew_user -p1234 --ssl-ca=/etc/mysql/certs/server/ca.crt  --ssl-cert=/etc/mysql/certs/server/tls.crt --ssl-key=/etc/mysql/certs/server/tls.key
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 116
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> exit 
Bye
```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/client/` directory.

## Deploy MariaDB Cluster with TLS/SSL configuration

Now, we are going to deploy a `MariaDB` Cluster with TLS/SSL configuration. Below is the YAML for MariaDB cluster that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: md-cluster-tls
  namespace: demo
spec:
  version: "10.5.23"
  replicas: 3
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
  deletionPolicy: WipeOut
```

**Deploy MariaDB Cluster:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/tls/configure/examples/tls-cluster.yaml
mariadb.kubedb.com/md-cluster-tls created
```

**Wait for the database to be ready :**

Now, wait for `MariaDB` going on `Running` state and also wait for `StatefulSet` and its pods to be created and going to `Running` state,

```bash
$ kubectl get mariadb -n demo md-cluster-tls
NAME             VERSION   STATUS   AGE
md-cluster-tls   10.5.23    Ready    2m49s

$ kubectl get pod -n demo | grep md-cluster-tls
md-cluster-tls-0   1/1     Running   0          3m29s
md-cluster-tls-1   1/1     Running   0          3m9s
md-cluster-tls-2   1/1     Running   0          2m49s
```

**Verify tls-secrets created successfully :**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` Ops Manager. Default tls-secret name formed as _{mariadb-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep md-cluster-tls
md-cluster-tls-archiver-cert             kubernetes.io/tls                     3      6m20s
md-cluster-tls-auth                      kubernetes.io/basic-auth              2      6m22s
md-cluster-tls-metrics-exporter-cert     kubernetes.io/tls                     3      6m20s
md-cluster-tls-metrics-exporter-config   Opaque                                1      6m21s
md-cluster-tls-server-cert               kubernetes.io/tls                     3      6m21s
md-cluster-tls-token-nrs75
```

**Verify MariaDB Cluster configured with TLS/SSL:**

Now, we are going to connect to the database for verifying the `MariaDB` server has configured with TLS/SSL encryption.

Let's exec into the first pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo md-cluster-tls-0 -- bash

root@md-cluster-tls-0:/ ls /etc/mysql/certs/client
ca.crt  tls.crt  tls.key
root@md-cluster-tls-0:/ ls /etc/mysql/certs/server
ca.crt  tls.crt  tls.key

root@md-cluster-tls-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 64
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

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

Now let's check for the second database server,

```bash
$ kubectl exec -it -n demo md-cluster-tls-1 -- bash
root@md-cluster-tls-1:/ ls /etc/mysql/certs/client
ca.crt  tls.crt  tls.key
root@md-cluster-tls-1:/ ls /etc/mysql/certs/server
ca.crt  tls.crt  tls.key
root@md-cluster-tls-1:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 34
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

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
10 rows in set (0.001 sec)

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
$ kubectl exec -it -n demo md-cluster-tls-0 -- bash
root@md-cluster-tls-0:/ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 92
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> CREATE USER 'new_user'@'localhost' IDENTIFIED BY '1234' REQUIRE SSL;
Query OK, 0 rows affected (0.028 sec)

MariaDB [(none)]> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.000 sec)

MariaDB [(none)]> exit
Bye

#  accessing the database server with newly created user
root@md-cluster-tls-0:/ mysql -unew_user -p1234
ERROR 1045 (28000): Access denied for user 'new_user'@'localhost' (using password: YES)

# accessing the database server newly created user with certificates
root@md-cluster-tls-0:/ mysql -unew_user -p1234 --ssl-ca=/etc/mysql/certs/server/ca.crt  --ssl-cert=/etc/mysql/certs/server/tls.crt --ssl-key=/etc/mysql/certs/server/tls.key
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MariaDB connection id is 116
Server version: 10.5.23-MariaDB-1:10.5.23+maria~focal mariadb.org binary distribution

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MariaDB [(none)]> exit 
Bye
```

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/client/` directory.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete  mariadb demo  md-standalone-tls
mariadb.kubedb.com "md-standalone-tls" deleted
$ kubectl delete  mariadb demo  md-cluster-tls
mariadb.kubedb.com "md-cluster-tls" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```