---
title: TLS/SSL (Transport Encryption)
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-tls-configure
    name: ProxySQL TLS/SSL Configuration
    parent: guides-proxysql-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Configure TLS/SSL in ProxySQL Frontend Connections

`KubeDB` supports providing TLS/SSL encryption for `ProxySQL`. This tutorial will show you how to use `KubeDB` to deploy a `ProxySQL` with TLS/SSL configuration.

> While talking about TLS secured connections in `ProxySQL`, we know there are two types of connections in `ProxySQL`. The first one is the client-to-proxy and second one is proxy-to-backend. The first type is refered as frontend  connection and the second one as backend. As for the backend connection, it will be TLS secured automatically if the necessary ca_bundle is provided with the `appbinding`. And as for the frontend connections to be TLS secured, in this tutorial we are going to discuss how to achieve it with KubeDB operator.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Install `KubeDB` community and enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/proxysql/tls/configure/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/proxysql/tls/configure/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).


### Deploy KubeDB MySQL instance as the backend

We need a mysql backend for the proxysql server. So we are creating one with the following yaml. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.36"
  replicas: 3
  topology:
    mode: GroupReplication
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/tls/configure/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created 
```

After applying the above yaml wait for the MySQL to be Ready.

## Deploy ProxySQL with TLS/SSL configuration

As pre-requisite, at first, we are going to create an Issuer/ClusterIssuer. This Issuer/ClusterIssuer is used to create certificates. Then we are going to deploy a `ProxySQL` cluster that will be configured with these certificates by `KubeDB` operator.

### Create Issuer/ClusterIssuer

Now, we are going to create an example `Issuer` that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`. With the following steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=proxysql/O=kubedb"
Generating a RSA private key
...........................................................................+++++
........................................................................................................+++++
writing new private key to './ca.key'
```

- create a secret using the certificate files we have just generated,

```bash
kubectl create secret tls proxy-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/proxy-ca created
```

Now, we are going to create an `Issuer` using the `proxy-ca` secret that holds the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: proxy-issuer
  namespace: demo
spec:
  ca:
    secretName: proxy-ca
```

Let’s create the `Issuer` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/tls/configure/examples/issuer.yaml
issuer.cert-manager.io/proxy-issuer created
```

### Deploy ProxySQL Cluster with TLS/SSL configuration

Here, our issuer `proxy-issuer`  is ready to deploy a `ProxySQL` cluster with TLS/SSL configuration. Below is the YAML for ProxySQL Cluster that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"  
  replicas: 3
  mode: GroupReplication
  backend:
    name: mysql-server
  sycnUsers: true
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: proxy-issuer
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

- `spec.tls.issuerRef` refers to the `proxy-issuer` issuer.

- `spec.tls.certificates` gives you a lot of options to configure so that the certificate will be renewed and kept up to date. 
You can find more details from [here](/docs/guides/proxysql/concepts/proxysql/index.md/#spectls)

**Deploy ProxySQL Cluster:**

Let’s create the `ProxySQL` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/tls/configure/examples/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

**Wait for the database to be ready:**

Now, wait for `ProxySQL` going on `Ready` state and also wait for `StatefulSet` and its pod to be created and going to `Running` state,

```bash
$ kubectl get proxysql -n demo proxy-server
NAME             VERSION       STATUS   AGE
proxy-server   2.3.2-debian    Ready    5m48s

$ kubectl get sts -n demo proxy-server
NAME             READY   AGE
proxy-server     3/3     7m5s
```

**Verify tls-secrets created successfully:**

If everything goes well, you can see that our tls-secrets will be created which contains server, client, exporter certificate. Server tls-secret will be used for server configuration and client tls-secret will be used for a secure connection.

All tls-secret are created by `KubeDB` enterprise operator. Default tls-secret name formed as _{proxysql-object-name}-{cert-alias}-cert_.

Let's check the tls-secrets have created,

```bash
$ kubectl get secrets -n demo | grep proxy-server
proxy-server-auth                    kubernetes.io/basic-auth              2      7m54s
proxy-server-configuration           Opaque                                1      7m54s
proxy-server-monitor                 kubernetes.io/basic-auth              2      7m54s
proxy-server-token-4w4mb             kubernetes.io/service-account-token   3      7m54s
proxy-server-server-cert             kubernetes.io/tls                     3      7m53s 
proxy-server-client-cert             kubernetes.io/tls                     3      7m53s 
```

**Verify ProxySQL Cluster configured with TLS/SSL:**

Now, we are going to connect to the proxysql server for verifying the proxysql server has configured with TLS/SSL encryption.

Let's exec into the pod to verify TLS/SSL configuration,

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash

root@proxy-server-0:/ ls /var/lib/frontend/client
ca.crt  tls.crt  tls.key
root@proxy-server-0:/ ls /var/lib/frontend/server
ca.crt  tls.crt  tls.key

root@proxy-server-0:/ mysql -uadmin -padmin -h127.0.0.1 -P6032 --prompt 'ProxySQLAdmin>'
Welcome to the ProxySQL monitor.  Commands end with ; or \g.
Your ProxySQL connection id is 64
Server version: 2.3.2-debian-ProxySQL-1:2.3.2-debian+maria~focal proxysql.org binary distribution

Copyright (c) 2000, 2018, Oracle, ProxySQL Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

ProxySQLAdmin [(none)]> show variables like '%have_ssl%';
+---------------------+-------------------------+
| Variable_name       | Value                   |
+---------------------+-------------------------+
| mysql-have_ssl      | true                    |
+---------------------+-------------------------+
10 rows in set (0.002 sec)

ProxySQLAdmin [(none)]> quit;
Bye
```

The above output shows that the proxy server is configured to TLS/SSL. You can also see that the `.crt` and `.key` files are stored in `/var/lib/frontend/client/` and `/var/lib/frontend/server/` directory for client and server respectively.

**Verify secure connection for user:**

Now, you can create an user that will be used to connect to the server with a secure connection.

First, lets create the user in the backend mysql server.

```bash
$ kubectl exec -it -n demo mysql-server-0 -- bash 
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
root@mysql-server-0:/# mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 26692
Server version: 5.7.36-log MySQL Community Server (GPL)

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create user 'test'@'%' identified by 'pass';
Query OK, 0 rows affected (0.00 sec)

mysql> grant all privileges on test.* to 'again'@'%';
Query OK, 0 rows affected (0.00 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.00 sec)
```

As we deployed the ProxySQL with `.spec.syncUsers` turned true, the user will automatically be fetched into the proxysql server. 

```bash
ProxySQLAdmin [(none)]> select username,active,use_ssl from mysql_users;
+----------+--------+---------+
| username | active | use_ssl |
+----------+--------+---------+
| root     | 1      | 0       |
| test     | 1      | 0       |
+----------+--------+---------+
2 rows in set (0.001 sec)
```

We need to turn the use_ssl on for tls secured connections.

```bash
ProxySQLAdmin [(none)]> update mysql_users set use_ssl=1 where username='test';
Query OK, 1 row affected (0.000 sec)

ProxySQLAdmin [(none)]> LOAD MYSQL USERS TO RUNTIME;
Query OK, 0 rows affected (0.001 sec)

ProxySQLAdmin [(none)]> SAVE MYSQL USERS TO DISK;
Query OK, 0 rows affected (0.008 sec)
```

Let's connect to the proxysql server with a secure connection,

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/ mysql -utest -ppass -h127.0.0.1 -P6033                                                                                                   
ERROR 1045 (28000): ProxySQL Error: Access denied for user 'test' (using password: YES). SSL is required

root@proxy-server-0:/ mysql -utest -ppass -h127.0.0.1 -P6033 --ssl-ca=/var/lib/frontend/server/ca.crt --ssl-cert=/var/lib/frontend/server/tls.crt --ssl-key=/var/lib/frontend/server/tls.key
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 1573
Server version: 8.0.27 (ProxySQL)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> \s
--------------
mysql  Ver 15.1 Distrib 10.5.15-MariaDB, for debian-linux-gnu (x86_64) using  EditLine wrapper

Connection id:		1573
Current database:	information_schema
Current user:		test@10.244.0.26
SSL:			Cipher in use is TLS_AES_256_GCM_SHA384
Current pager:		stdout
Using outfile:		''
Using delimiter:	;
Server:			MySQL
Server version:		8.0.27 (ProxySQL)
Protocol version:	10
Connection:		127.0.0.1 via TCP/IP
Server characterset:	latin1
Db     characterset:	utf8
Client characterset:	latin1
Conn.  characterset:	latin1
TCP port:		6033
Uptime:			2 hours 30 min 27 sec

Threads: 1  Questions: 12  Slow queries: 12
--------------

```

In the above output section we can see there is cipher in user at the SSL field. Which means the connection is TLS secured. 

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete  proxysql -n demo  proxy-server
$ kubectl delete mysql -n demo mysql-server
$ kubectl delete issuer -n demo --all
$ kubectl delete ns demo
```