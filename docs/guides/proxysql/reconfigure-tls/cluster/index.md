---
title: Reconfigure ProxySQL TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-reconfigure-tls-cluster
    name: Reconfigure ProxySQL TLS/SSL Encryption
    parent: guides-proxysql-reconfigure-tls
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure ProxySQL TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing ProxySQL via a ProxySQLOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

`ReconfigureTLS` is a very useful ops-request when it comes to reconfiguring TLS settings for proxysql server without entering the admin panel. With this type of ops-request you can `add`, `remove` and `update` TLS configuration for the proxysql server. You can `rotate` the certificates as well.

Below, we are providing some examples for the ops-request. 

## Before You Begin

- At first, you need to have a Kubernetes Cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.6.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

### Prepare MySQL Backend

To test any proxysql functionality we need to have a mysql backend . 

Below, here is the yaml for the KubeDB MySQL backend.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-server
  namespace: demo
spec:
  version: "5.7.41"
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

Let's apply the yaml, 

``` bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/sample-mysql.yaml
mysql.kubedb.com/mysql-server created
```

Let's now wait for the mysql instance to be ready, 

```bash
$ kubectl get mysql -n demo
NAME           VERSION   STATUS   AGE
mysql-server   5.7.41    Ready    3m16s

$ kubectl get pods -n demo
NAME             READY   STATUS    RESTARTS   AGE
mysql-server-0   2/2     Running   0          3m11s
mysql-server-1   2/2     Running   0          113s
mysql-server-2   2/2     Running   0          109s
```

We need a user to test all the ssl functionalities. So let's create one user inside the mysql servers,

```bash
~ $ kubectl exec -it -n demo mysql-server-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
root@mysql-server-0:/# mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 106
Server version: 5.7.41-log MySQL Community Server (GPL)

Copyright (c) 2000, 2021, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create user 'test'@'%' identified by 'pass';
Query OK, 0 rows affected (0.00 sec)

mysql> create database testdb;
Query OK, 1 row affected (0.00 sec)

mysql> grant all privileges on testdb.* to 'test'@'%';
Query OK, 0 rows affected (0.01 sec)

mysql> flush privileges;
Query OK, 0 rows affected (0.00 sec)

mysql> exit
Bye
```

## Deploy ProxySQL without TLS

We are now all set with our backend. Now let's create a KubeDB ProxySQL server. Lets keep the syncUser field true so that we don't need to create the user again. 

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ProxySQL
metadata:
  name: proxy-server
  namespace: demo
spec:
  version: "2.3.2-debian"
  replicas: 3
  backend:
    name: mysql-server
  syncUsers: true
  terminationPolicy: WipeOut
```

``` bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/sample-proxysql.yaml
proxysql.kubedb.com/proxy-server created
```

## Check User and current TLS status

Let's exec into the proxysql pod and see the current status. 

```bash
$ kubectl exec -it -n demo proxy-server-0 -- bash
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 18
Server version: 8.0.32 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> select username, use_ssl from mysql_users;
+----------+---------+
| username | use_ssl |
+----------+---------+
| root     | 0       |
| test     | 0       |
+----------+---------+
2 rows in set (0.000 sec)

MySQL [(none)]> show variables like '%have_ssl%';
+----------------+-------+
| Variable_name  | Value |
+----------------+-------+
| mysql-have_ssl | false |
+----------------+-------+
1 row in set (0.001 sec)

MySQL [(none)]> exit 
Bye
```
We can see that the users have been fetched. Also the mysql-have_ssl variables is set to false. The use_ssl column is also set to 0 which means that there is no need for ssl-ca or cert for connect. 

Let's check it with the follwing command. 

```bash
root@proxy-server-0:/# mysql -utest -ppass -h127.0.0.1 -P6033
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 914
Server version: 8.0.32 (ProxySQL)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> \s
--------------
mysql  Ver 15.1 Distrib 10.5.15-MariaDB, for debian-linux-gnu (x86_64) using  EditLine wrapper

Connection id:		914
Current database:	information_schema
Current user:		test@10.244.0.20
SSL:			Not in use
Current pager:		stdout
Using outfile:		''
Using delimiter:	;
Server:			MySQL
Server version:		8.0.32 (ProxySQL)
Protocol version:	10
Connection:		127.0.0.1 via TCP/IP
Server characterset:	latin1
Db     characterset:	utf8
Client characterset:	latin1
Conn.  characterset:	latin1
TCP port:		6033
Uptime:			1 hour 27 min 36 sec

Threads: 1  Questions: 3  Slow queries: 3
--------------

MySQL [(none)]> exit
Bye
```

## Add TLS with RreconfigureTLS Ops-Request

Now we want to add TLS to our proxysql server and we want the frontend connections to be tls-secured.

### Create Issuer

First we need an issuer for this. We can create one with the following command. Make sure that you have cert-manager running in your cluster and openssl installed. 

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mysql/O=kubedb"

Generating a RSA private key
.......................................+++++
...........................+++++
writing new private key to './ca.key'

```

Let's create the ca-secret with the above created ca.crt and ca.key by using the following command,

```bash
$ kubectl create secret tls proxy-ca \
                        --cert=ca.crt \
                        --key=ca.key \
                        --namespace=demo
secret/proxy-ca created
```

Now create issuer with the following yaml, 

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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/issuer.yaml
issuer.cert-manager.io/proxy-issuer created
```

### Apply ops-request to add TLS

We are all set to go! now lets create an ReconfigureTLS ops-request like below. We have set a desired configuration under the `.spec.tls` section here as you can see. You can checkout the api documentation of this field [here](https://pkg.go.dev/kubedb.dev/apimachinery@v0.29.1/apis/ops/v1alpha1#ProxySQLOpsRequestSpec).

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: recon-tls-add
  namespace: demo
spec:
  type: ReconfigureTLS
  proxyRef:
    name: proxy-server
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
      emailAddresses: 
        - "spike@appscode.com"   
```

Let's apply and wait for the ops-request to be succeeded. 

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/proxyops-add-tls.yaml
proxysqlopsrequest.ops.kubedb.com/recon-tls-add created

$ kubectl get proxysqlopsrequest -n demo
NAME               TYPE             STATUS        AGE
recon-tls-add      ReconfigureTLS   Successful    5m
```

### Check ops-request effects

Following secrets should be created 

```bash
$ kubectl get secrets -n demo | grep cert
proxy-server-server-cert             kubernetes.io/tls                     3      4m53s 
proxy-server-client-cert             kubernetes.io/tls                     3      4m53s 
```

The directory `/var/lib/frontend/` should carry the certificates and other files within the directories as seen below. 
```bash
root@proxy-server-0:/# ls /var/lib/frontend/
client	server

root@proxy-server-0:/# ls /var/lib/frontend/client
ca.crt   tls.crt   tls.key

root@proxy-server-0:/# ls /var/lib/frontend/server
ca.crt   tls.crt   tls.key
```

The `mysql-have_ssl` variables should be true by this time. 

```bash
root@proxy-server-0:/# mysql -uadmin -padmin -h127.0.0.1 -P6032
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 22
Server version: 8.0.32 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> show variables like '%have_ssl%';
+----------------+-------+
| Variable_name  | Value |
+----------------+-------+
| mysql-have_ssl | true  |
+----------------+-------+
1 row in set (0.001 sec)
```

### Activate use_ssl field for the test user

Now our ProxySQL server is ready to serve tls-secured connections. Let's modify our test user to use ssl with an ops-request. You can do this task from the admin panel also. But we like to do it in KubeDB way. 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: activate-ssl
  namespace: demo
spec:
  type: Reconfigure  
  proxyRef:
    name: proxy-server
  configuration:
    mysqlUsers:
      users: 
      - username: test
        use_ssl: 1
      reqType: update
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/proxyops-activate-ssl.yaml
proxysqlopsrequest.ops.kubedb.com/activate-ssl created
```

Let's check the effect from the admin panel. 

```bash
MySQL [(none)]> select username,use_ssl from mysql_users;
+----------+---------+
| username | use_ssl |
+----------+---------+
| root     | 0       |
| test     | 1       |
+----------+---------+
2 rows in set (0.001 sec)
```

### Check TLS secured connections

Now our user is also modified to accept only tls-secured requests. Let's try to connect without TLS. 

```bash
root@proxy-server-0:/# mysql -utest -ppass -h127.0.0.01 -P6033
ERROR 1045 (28000): ProxySQL Error: Access denied for user 'test' (using password: YES). SSL is required
```

We can see that the connection is refused. Now try with the tls certificates.

```bash
root@proxy-server-0:/# mysql -utest -ppass -h127.0.0.01 -P6033 --ssl-ca=/var/lib/frontend/client/ca.crt --ssl-cert=/var/lib/frontend/client/tls.crt --ssl-key=/var/lib/frontend/client/tls.key
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 107
Server version: 8.0.32 (ProxySQL)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [testdb]> \s
--------------
mysql  Ver 15.1 Distrib 10.5.15-MariaDB, for debian-linux-gnu (x86_64) using  EditLine wrapper

Connection id:		107
Current database:	testdb
Current user:		test@10.244.0.23
SSL:			Cipher in use is TLS_AES_256_GCM_SHA384
Current pager:		stdout
Using outfile:		''
Using delimiter:	;
Server:			MySQL
Server version:		8.0.32 (ProxySQL)
Protocol version:	10
Connection:		127.0.0.01 via TCP/IP
Server characterset:	latin1
Db     characterset:	latin1
Client characterset:	latin1
Conn.  characterset:	latin1
TCP port:		6033
Uptime:			8 min 3 sec

Threads: 1  Questions: 7  Slow queries: 7
--------------
```

We can see that the user is successfuly logged in with the tls informations. Also in the `\s` query result , the SSL field has got a cipher name, which means the connection is tls-secured. 

## Rotate Certificate

Now we are going to rotate the certificate for this proxysql. First let's check the current expiration date for current certificate.

```bash
root@proxy-server-0:/# openssl x509 -in /var/lib/frontend/client/tls.crt -inform  PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb  6 08:44:01 2023 GMT
```

Let's look into the server certificate crd. 

```bash
~ $ kubectl describe certificate -n demo proxy-server-server-cert
Name:         proxy-server-server-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=proxy-server
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=proxysqls.kubedb.com
              proxysql.kubedb.com/load-balance=GroupReplication
Annotations:  <none>
API Version:  cert-manager.io/v1
Kind:         Certificate
Metadata:
  Creation Timestamp:  2022-11-08T08:44:01Z
  Generation:          1
  ...
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ProxySQL
    Name:                  proxy-server
    UID:                   b4fa48bc-b6cc-4ce7-beaf-c91987f4e0b5
  Resource Version:        29102
  UID:                     aa69c146-1581-4fce-a160-ad85b4296e4d
Spec:
  Common Name:  proxy-server
  Dns Names:
    *.proxy-server-pods.demo.svc
    *.proxy-server-pods.demo.svc.cluster.local
    *.proxy-server.demo.svc
    localhost
    proxy-server
    proxy-server.demo.svc
  Email Addresses:
    spike@appscode.com
  Ip Addresses:
    127.0.0.1
  Issuer Ref:
    Group:      cert-manager.io
    Kind:       Issuer
    Name:       proxy-issuer
  Secret Name:  proxy-server-server-cert
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
    Last Transition Time:  2022-11-08T08:44:01Z
    Message:               Certificate is up to date and has not expired
    Observed Generation:   1
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2023-02-06T08:44:01Z
  Not Before:              2022-11-08T08:44:01Z
  Renewal Time:            2023-01-07T08:44:01Z
  Revision:                1
Events:
  Type    Reason     Age   From          Message
  ----    ------     ----  ----          -------
  Normal  Issuing    17m   cert-manager  Issuing certificate as Secret does not exist
  Normal  Generated  17m   cert-manager  Stored new private key in temporary Secret resource "proxy-server-server-cert-ksk6g"
  Normal  Requested  17m   cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-9mqjf"
  Normal  Issuing    17m   cert-manager  The certificate has been successfully issued
```

### Apply ops-request to rotate certificate

Now lets apply the follwoing yaml and rotate the certificate of our proxysql server. 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: recon-tls-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  proxyRef:
    name: proxy-server
  tls:
    rotateCertificates: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/proxyops-rotate-tls.yaml
proxysqlopsrequest.ops.kubedb.com/recon-tls-rotate created
```

```bash
$ kubectl get proxysqlopsrequest -n demo
NAME                    TYPE             STATUS      AGE
recon-tls-add         ReconfigureTLS   Successful    15m
recon-tls-rotate      ReconfigureTLS   Successful    5m
```

### Check ops-request effect

Let's check if the expiration time has been updated or not. 

```bash
root@proxy-server-0:/# openssl x509 -in /var/lib/frontend/client/tls.crt -inform  PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb  6 09:05:54 2023 GMT
```

The expiration time has been updated. Now lets check the certificate crd. 

```bash
 $ kubectl describe certificate -n demo proxy-server-server-cert
Name:         proxy-server-server-cert
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=proxy-server
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=proxysqls.kubedb.com
              proxysql.kubedb.com/load-balance=GroupReplication
Annotations:  <none>
API Version:  cert-manager.io/v1
Kind:         Certificate
Metadata:
  Creation Timestamp:  2022-11-08T08:44:01Z
  Generation:          1
  ...
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  ProxySQL
    Name:                  proxy-server
    UID:                   b4fa48bc-b6cc-4ce7-beaf-c91987f4e0b5
  Resource Version:        32254
  UID:                     aa69c146-1581-4fce-a160-ad85b4296e4d
Spec:
  Common Name:  proxy-server
  Dns Names:
    *.proxy-server-pods.demo.svc
    *.proxy-server-pods.demo.svc.cluster.local
    *.proxy-server.demo.svc
    localhost
    proxy-server
    proxy-server.demo.svc
  Email Addresses:
    spike@appscode.com
  Ip Addresses:
    127.0.0.1
  Issuer Ref:
    Group:      cert-manager.io
    Kind:       Issuer
    Name:       proxy-issuer
  Secret Name:  proxy-server-server-cert
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
    Last Transition Time:  2022-11-08T08:44:01Z
    Message:               Certificate is up to date and has not expired
    Observed Generation:   1
    Reason:                Ready
    Status:                True
    Type:                  Ready
  Not After:               2023-02-06T09:05:54Z
  Not Before:              2022-11-08T09:05:54Z
  Renewal Time:            2023-01-07T09:05:54Z
  Revision:                6
Events:
  Type    Reason     Age                   From          Message
  ----    ------     ----                  ----          -------
  Normal  Issuing    23m                   cert-manager  Issuing certificate as Secret does not exist
  Normal  Generated  23m                   cert-manager  Stored new private key in temporary Secret resource "proxy-server-server-cert-ksk6g"
  Normal  Requested  23m                   cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-9mqjf"
  Normal  Requested  4m22s                 cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-s7d6r"
  Normal  Requested  4m22s                 cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-cd5sg"
  Normal  Requested  4m17s                 cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-pbm8q"
  Normal  Requested  2m9s                  cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-4qm6l"
  Normal  Requested  2m2s                  cert-manager  Created new CertificateRequest resource "proxy-server-server-cert-l2xgk"
  Normal  Reused     2m2s (x5 over 4m22s)  cert-manager  Reusing private key stored in existing Secret resource "proxy-server-server-cert"
  Normal  Issuing    2m1s (x6 over 23m)    cert-manager  The certificate has been successfully issued
```

This has also been updated.

So from the above ovservation we can say that the TLS certificate rotation has been succeeded. 

## Update TLS Configuration

Now lets update the certificate information. 

Let's check the current info first. 

```bash
root@proxy-server-0:/# openssl x509 -in /var/lib/proxysql/proxysql-cert.pem -inform PEM  -subject -email -nameopt RFC2253 -noout
subject=CN=proxy-server,O=kubedb:server
spike@appscode.com
```

### Apply ops-request to update TLS

We can see the informations. Suppose we want to update the email address . We want to change it to mikebaker@gmail.com. Let's create a ops-request for that in the following manner. 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: recon-tls-update
  namespace: demo
spec:
  type: ReconfigureTLS
  proxyRef:
    name: proxy-server
  tls:
    certificates:
    - alias: server
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses:
      - "mikebaker@gmail.com"
      certificates:
    - alias: client
      subject:
        organizations:
        - kubedb:server
      dnsNames:
      - localhost
      ipAddresses:
      - "127.0.0.1"
      emailAddresses:
      - "mikebaker@gmail.com"
```

Let's apply and then wait for it to be succeed. 

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/proxyops-update-tls.yaml
proxysqlopsrequest.ops.kubedb.com/recon-tls-update created

$ kubectl get proxysqlopsrequest -n demo
NAME                    TYPE             STATUS        AGE
recon-tls-update      ReconfigureTLS   Successful    5m
recon-tls-add         ReconfigureTLS   Successful    15m
recon-tls-rotate      ReconfigureTLS   Successful    10m
```

Let's check the info now. 

```bash
root@proxy-server-1:/# openssl x509 -in /var/lib/frontend/server/tls.crt -inform PEM  -subject -email -nameopt RFC2253 -noout
subject=CN=proxy-server,O=kubedb:server
mikebaker@gmail.com
```

We can see the email has been successfuly updated. You can configure other field as well. To know more about the .spec.tls field refer to the link [here](https://pkg.go.dev/kubedb.dev/apimachinery@v0.29.1/apis/ops/v1alpha1#TLSSpec) .

## Remove TLS

To remove TLS from a KubeDB ProxySQL instance, all you need to do is apply a similar yaml like below. Just change the `.spec.proxyRef.name` field with your own ProxySQL instance name. 

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ProxySQLOpsRequest
metadata:
  name: recon-tls-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  proxyRef:
    name: proxy-server
  tls:
    remove: true
```

Let's apply and check the effects. 

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/proxysql/reconfigure-tls/cluster/examples/proxyops-remove-tls.yaml
proxysqlopsrequest.ops.kubedb.com/recon-tls-remove created

$ kubectl get proxysqlopsrequest -n demo
NAME                    TYPE             STATUS        AGE
recon-tls-remove      ReconfigureTLS   Successful    3m
recon-tls-update      ReconfigureTLS   Successful    7m
recon-tls-add         ReconfigureTLS   Successful    17m
recon-tls-rotate      ReconfigureTLS   Successful    12m
```

### Check ops-request effect

Let's check the effect. 

```bash
root@proxy-server-1:/# mysql -uadmin -padmin -h127.0.0.1 -P6032
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 25
Server version: 8.0.32 (ProxySQL Admin Module)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> show variables like '%have_ssl%';
+----------------+-------+
| Variable_name  | Value |
+----------------+-------+
| mysql-have_ssl | false |
+----------------+-------+
1 row in set (0.001 sec)
```

The mysql-have_ssl has been set to false by the ops-request. So no more tls-secured frontend connections will be created. 

Let's update the user configuration to use_ssl=0 . Otherwise the user won't be able to connect. 

```bash
MySQL [(none)]> update mysql_users set use_ssl=0 where username='test';
Query OK, 1 row affected (0.001 sec)

MySQL [(none)]> LOAD MYSQL USERS TO RUNTIME;
Query OK, 0 rows affected (0.001 sec)

MySQL [(none)]> ^DBye

root@proxy-server-1:/# mysql -utest -ppass -h127.0.0.1 -P6033
Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 267
Server version: 8.0.32 (ProxySQL)

Copyright (c) 2000, 2018, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [(none)]> 
```

We can see the user has been successfuly connected without the tls information.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete proxysql -n demo --all
$ kubectl delete issuer -n demo --all
$ kubectl delete proxysqlopsrequest -n demo --all
$ kubectl delete ns demo
```