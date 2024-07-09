---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-tls-configure
    name: PerconaXtraDB TLS/SSL Configuration
    parent: guides-perconaxtradb-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configure TLS/SSL in PerconaXtraDB

`KubeDB` supports providing TLS/SSL encryption (via, `requireSSL` mode) for `PerconaXtraDB`. This tutorial will show you how to use `KubeDB` to deploy a `PerconaXtraDB` database with TLS/SSL configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.9.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/percona-xtradb/tls/configure/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/tls/configure/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

### Deploy PerconaXtraDB database with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a PerconaXtraDB standalone and a group replication that will be configured with these certificates by `KubeDB` operator.

### Create Issuer/ClusterIssuer

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/tls/configure/examples/issuer.yaml
issuer.cert-manager.io/px-issuer created
```

## Deploy PerconaXtraDB Cluster with TLS/SSL configuration

Now, we are going to deploy a `PerconaXtraDB` Cluster with TLS/SSL configuration. Below is the YAML for PerconaXtraDB cluster that we are going to create,

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
  requireSSL: true
  tls:
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
  deletionPolicy: WipeOut
```

**Deploy PerconaXtraDB Cluster:**

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/percona-xtradb/tls/configure/examples/tls-cluster.yaml
perconaxtradb.kubedb.com/sample-pxc created
```

**Wait for the database to be ready :**

Now, wait for `PerconaXtraDB` going on `Running` state and also wait for `StatefulSet` and its pods to be created and going to `Running` state,

```bash
$ kubectl get perconaxtradb -n demo sample-pxc
NAME         VERSION   STATUS   AGE
sample-pxc   8.0.26    Ready    3m23s


$ kubectl get pod -n demo | grep sample-pxc
sample-pxc-0   2/2     Running   0          3m32s
sample-pxc-1   2/2     Running   0          3m32s
sample-pxc-2   2/2     Running   0          3m32s
```

**Verify tls-secrets created successfully :**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` Ops Manager. Default tls-secret name formed as _{perconaxtradb-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep sample-pxc
sample-pxc-auth                    kubernetes.io/basic-auth              2      4m18s
sample-pxc-client-cert             kubernetes.io/tls                     3      4m19s
sample-pxc-metrics-exporter-cert   kubernetes.io/tls                     3      4m18s
sample-pxc-monitor                 kubernetes.io/basic-auth              2      4m18s
sample-pxc-replication             kubernetes.io/basic-auth              2      4m18s
sample-pxc-server-cert             kubernetes.io/tls                     3      4m18s
sample-pxc-token-84hrj             kubernetes.io/service-account-token   3      4m19s

```

**Verify PerconaXtraDB Cluster configured with TLS/SSL:**

Now, we are going to connect to the database for verifying the `PerconaXtraDB` server has configured with TLS/SSL encryption.

Let's exec into the first pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ ls /etc/mysql/certs/client
ca.crt	tls.crt  tls.key
bash-4.4$ ls /etc/mysql/certs/server
ca.crt	tls.crt  tls.key
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 78
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like '%ssl%';
+-------------------------------------+---------------------------------+
| Variable_name                       | Value                           |
+-------------------------------------+---------------------------------+
| admin_ssl_ca                        |                                 |
| admin_ssl_capath                    |                                 |
| admin_ssl_cert                      |                                 |
| admin_ssl_cipher                    |                                 |
| admin_ssl_crl                       |                                 |
| admin_ssl_crlpath                   |                                 |
| admin_ssl_key                       |                                 |
| have_openssl                        | YES                             |
| have_ssl                            | YES                             |
| mysqlx_ssl_ca                       |                                 |
| mysqlx_ssl_capath                   |                                 |
| mysqlx_ssl_cert                     |                                 |
| mysqlx_ssl_cipher                   |                                 |
| mysqlx_ssl_crl                      |                                 |
| mysqlx_ssl_crlpath                  |                                 |
| mysqlx_ssl_key                      |                                 |
| performance_schema_show_processlist | OFF                             |
| ssl_ca                              | /etc/mysql/certs/server/ca.crt  |
| ssl_capath                          | /etc/mysql/certs/server         |
| ssl_cert                            | /etc/mysql/certs/server/tls.crt |
| ssl_cipher                          |                                 |
| ssl_crl                             |                                 |
| ssl_crlpath                         |                                 |
| ssl_fips_mode                       | OFF                             |
| ssl_key                             | /etc/mysql/certs/server/tls.key |
+-------------------------------------+---------------------------------+
25 rows in set (0.00 sec)

mysql> show variables like '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.00 sec)

mysql> quit;
Bye

```

Now let's check for the second database server,

```bash
$ kubectl exec -it -n demo sample-pxc-1 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ ls /etc/mysql/certs/client
ca.crt	tls.crt  tls.key
bash-4.4$ ls /etc/mysql/certs/server
ca.crt	tls.crt  tls.key
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 186
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> show variables like '%ssl%';
+-------------------------------------+---------------------------------+
| Variable_name                       | Value                           |
+-------------------------------------+---------------------------------+
| admin_ssl_ca                        |                                 |
| admin_ssl_capath                    |                                 |
| admin_ssl_cert                      |                                 |
| admin_ssl_cipher                    |                                 |
| admin_ssl_crl                       |                                 |
| admin_ssl_crlpath                   |                                 |
| admin_ssl_key                       |                                 |
| have_openssl                        | YES                             |
| have_ssl                            | YES                             |
| mysqlx_ssl_ca                       |                                 |
| mysqlx_ssl_capath                   |                                 |
| mysqlx_ssl_cert                     |                                 |
| mysqlx_ssl_cipher                   |                                 |
| mysqlx_ssl_crl                      |                                 |
| mysqlx_ssl_crlpath                  |                                 |
| mysqlx_ssl_key                      |                                 |
| performance_schema_show_processlist | OFF                             |
| ssl_ca                              | /etc/mysql/certs/server/ca.crt  |
| ssl_capath                          | /etc/mysql/certs/server         |
| ssl_cert                            | /etc/mysql/certs/server/tls.crt |
| ssl_cipher                          |                                 |
| ssl_crl                             |                                 |
| ssl_crlpath                         |                                 |
| ssl_fips_mode                       | OFF                             |
| ssl_key                             | /etc/mysql/certs/server/tls.key |
+-------------------------------------+---------------------------------+
25 rows in set (0.00 sec)

mysql> show variables like '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.00 sec)

mysql> quit;
Bye
```

The above output shows that the `PerconaXtraDB` server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in `/etc/mysql/certs/client/` and `/etc/mysql/certs/server/` directory for client and server respectively.

**Verify secure connection for SSL required user:**

Now, you can create an SSL required user that will be used to connect to the database with a secure connection.

Let's connect to the database server with a secure connection,

```bash
$ kubectl exec -it -n demo sample-pxc-0 -- bash
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
bash-4.4$ mysql -u${MYSQL_ROOT_USERNAME} -p${MYSQL_ROOT_PASSWORD}
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 232
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE USER 'new_user'@'localhost' IDENTIFIED BY '1234' REQUIRE SSL;
Query OK, 0 rows affected (0.01 sec)

mysql> FLUSH PRIVILEGES;
Query OK, 0 rows affected (0.01 sec)

mysql> exit
Bye
bash-4.4$ mysql -unew_user -p1234
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1045 (28000): Access denied for user 'new_user'@'localhost' (using password: YES)
bash-4.4$ mysql -unew_user -p1234 --ssl-ca=/etc/mysql/certs/server/ca.crt  --ssl-cert=/etc/mysql/certs/server/tls.crt --ssl-key=/etc/mysql/certs/server/tls.key
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 242
Server version: 8.0.26-16.1 Percona XtraDB Cluster (GPL), Release rel16, Revision b141904, WSREP version 26.4.3

Copyright (c) 2009-2021 Percona LLC and/or its affiliates
Copyright (c) 2000, 2021, Oracle and/or its affiliates.

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

From the above output, you can see that only using client certificate we can access the database securely, otherwise, it shows "Access denied". Our client certificate is stored in `/etc/mysql/certs/client/` directory.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete  perconaxtradb demo  sample-pxc
perconaxtradb.kubedb.com "sample-pxc" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```