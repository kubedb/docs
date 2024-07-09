---
title: Reconfigure MySQL TLS/SSL Encryption
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-reconfigure-tls-readme
    name: Reconfigure MySQL TLS/SSL Encryption
    parent: guides-mysql-reconfigure-tls
    weight: 12
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MySQL TLS/SSL (Transport Encryption)

KubeDB supports reconfigure i.e. add, remove, update and rotation of TLS/SSL certificates for existing MySQL database via a MySQLOpsRequest. This tutorial will show you how to use KubeDB to reconfigure TLS/SSL encryption.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install [`cert-manger`](https://cert-manager.io/docs/installation/) v1.0.0 or later to your cluster to manage your SSL/TLS certificates.

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/mysql/reconfigure-tls/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS to a MySQL database

Here, We are going to create a MySQL database without TLS and then reconfigure the database to use TLS.

### Deploy MySQL without TLS

In this section, we are going to deploy a MySQL database without TLS. In the next few sections we will reconfigure TLS using `MySQLOpsRequest` CRD. Below is the YAML of the `MySQL` CR that we are going to create,

<ul class="nav nav-tabs" id="definationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link" id="st-tab" data-toggle="tab" href="#standAlone" role="tab" aria-controls="standAlone" aria-selected="true">Stand Alone</a>
  </li>

  <li class="nav-item">
    <a class="nav-link active" id="gr-tab" data-toggle="tab" href="#groupReplication" role="tab" aria-controls="groupReplication" aria-selected="false">Group Replication</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="ic-tab" data-toggle="tab" href="#innodbCluster" role="tab" aria-controls="innodbCluster" aria-selected="false">Innodb Cluster</a>
  </li>

  <li class="nav-item">
    <a class="nav-link" id="sc-tab" data-toggle="tab" href="#semisync" role="tab" aria-controls="semisync" aria-selected="false">Semi sync </a>
  </li>

</ul>


<div class="tab-content" id="definationTabContent">
  <div class="tab-pane fade" id="groupReplication" role="tabpanel" aria-labelledby="gr-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
spec:
  version: "8.0.35"
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
  deletionPolicy: Delete
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/group-replication.yaml
mysql.kubedb.com/mysql created
```

  </div>

  <div class="tab-pane fade" id="innodbCluster" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
spec:
  version: "8.0.31-innodb"
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
  deletionPolicy: Delete
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/innodb-cluster.yaml
mysql.kubedb.com/mysql created
```

  </div>

  <div class="tab-pane fade " id="semisync" role="tabpanel" aria-labelledby="sc-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: semi-sync-mysql
  namespace: demo
spec:
  version: "8.0.35"
  replicas: 3
  topology:
    mode: SemiSync
    semiSync:
      sourceWaitForReplicaCount: 1
      sourceTimeout: 23h
      errantTransactionRecoveryPolicy: PseudoTransaction
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/semi-sync.yaml
mysql.kubedb.com/mysql created
```

  </div>


  <div class="tab-pane fade show active" id="standAlone" role="tabpanel" aria-labelledby="st-tab">

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
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
  deletionPolicy: Delete

```

Let's create the `MySQL` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/standalone.yaml
mysql.kubedb.com/mysql created
```
  </div>

</div>



Now, wait until `mysql` has status `Ready`. i.e,

```bash
$ kubectl get my -n demo
NAME    VERSION   STATUS   AGE
mysql   8.0.35    Ready    75s

$ kubectl dba describe mysql mysql -n demo
Name:               mysql
Namespace:          demo
CreationTimestamp:  Mon, 21 Nov 2022 16:18:44 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"mysql","namespace":"demo"},"spec":{"storage":{"accessModes":["R...
Replicas:           1  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

PetSet:          
  Name:               mysql
  CreationTimestamp:  Mon, 21 Nov 2022 16:18:49 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mysql
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:        <none>
  Replicas:           824635546904 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mysql
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.238.135
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.23:3306

Service:        
  Name:         mysql-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.23:3306

Auth Secret:
  Name:         mysql-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mysql
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"mysql","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","version":"8.0.35"}}

    Creation Timestamp:  2022-11-21T10:18:49Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mysql
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
    Name:                            mysql
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mysql
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(mysql.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  mysql-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  mysql-restore-8.0.21
    Secret:
      Name:   mysql-auth
    Type:     kubedb.com/mysql
    Version:  8.0.35

Events:
  Type    Reason         Age   From             Message
  ----    ------         ----  ----             -------
  Normal  Phase Changed  1m    KubeDB Operator  phase changed from  to Provisioning reason:
  Normal  Successful     1m    KubeDB Operator  Successfully created governing service
  Normal  Successful     1m    KubeDB Operator  Successfully created service for primary/standalone
  Normal  Successful     1m    KubeDB Operator  Successfully created PetSet
  Normal  Successful     1m    KubeDB Operator  Successfully created MySQL
  Normal  Successful     1m    KubeDB Operator  Successfully created appbinding
  Normal  Phase Changed  25s   KubeDB Operator  phase changed from Provisioning to Ready reason:

```

Now, we can connect to this database through `mysql-shell` and verify that the TLS is disabled.


```bash
$ kubectl get secrets -n demo mysql-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-auth -o jsonpath='{.data.\password}' | base64 -d
f8EyKG)mNMIMdS~a

$ kubectl exec -it mysql-0 -n demo -- mysql -u root --password='f8EyKG)mNMIMdS~a'  --host=mysql-0.mysql-pods.demo -e "show variables like '%require_secure_transport%';";
Defaulted container "mysql" out of: mysql, mysql-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | OFF   |
+--------------------------+-------+

kubectl exec -it mysql-0 -n demo -- mysql -u root --password='f8EyKG)mNMIMdS~a'  --host=mysql-0.mysql-pods.demo -e "\s;";
Defaulted container "mysql" out of: mysql, mysql-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
--------------
mysql  Ver 8.0.35 for Linux on x86_64 (MySQL Community Server - GPL)

Connection id:		91
Current database:	
Current user:		root@mysql-0.mysql-pods.demo.svc.cluster.local
SSL:			Cipher in use is TLS_AES_256_GCM_SHA384
Current pager:		stdout
Using outfile:		''
Using delimiter:	;
Server version:		8.0.35 MySQL Community Server - GPL
Protocol version:	10
Connection:		mysql-0.mysql-pods.demo via TCP/IP
Server characterset:	utf8mb4
Db     characterset:	utf8mb4
Client characterset:	latin1
Conn.  characterset:	latin1
TCP port:		3306
Binary data as:		Hexadecimal
Uptime:			11 min 44 sec

Threads: 2  Questions: 454  Slow queries: 0  Opens: 185  Flush tables: 3  Open tables: 104  Queries per second avg: 0.644


```

We can verify from the above output that TLS is disabled for this database.

### Create Issuer/ ClusterIssuer

Now, We are going to create an example `Issuer` that will be used to enable SSL/TLS in MySQL. Alternatively, you can follow this [cert-manager tutorial](https://cert-manager.io/docs/configuration/ca/) to create your own `Issuer`.

- Start off by generating a ca certificates using openssl.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca/O=kubedb"
Generating a RSA private key
................+++++
........................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls my-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/my-ca created
```

Now, Let's create an `Issuer` using the `my-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: my-issuer
  namespace: demo
spec:
  ca:
    secretName: my-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/issuer.yaml
issuer.cert-manager.io/my-issuer created
```

### Create MySQLOpsRequest

In order to add TLS to the database, we have to create a `MySQLOpsRequest` CRO with our created issuer. Below is the YAML of the `MySQLOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mysql
  tls:
    requireSSL: true
    issuerRef:
      name: my-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - mysql
          organizationalUnits:
            - client
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mysql` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.
- `spec.tls.certificates` specifies the certificates. You can learn more about this field from [here](/docs/guides/mysql/concepts/database/index.md#spectls).

Let's create the `MySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/myops-add-tls.yaml
mysqlopsrequest.ops.kubedb.com/myops-add-tls created
```

#### Verify TLS Enabled Successfully

Let's wait for `MySQLOpsRequest` to be `Successful`.  Run the following command to watch `MySQLOpsRequest` CRO,

```bash
$ kubectl get mysqlopsrequest -n demo
NAME           TYPE             STATUS        AGE
myops-add-tls   ReconfigureTLS   Successful    91s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mysqlopsrequest -n demo myops-add-tls 
Name:         myops-add-tls
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-11-22T04:09:32Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:certificates:
          f:issuerRef:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-11-22T04:09:32Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-11-22T04:09:34Z
  Resource Version:  715635
  UID:               0bae4203-991b-4377-b38b-981648855638
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mysql
  Tls:
    Certificates:
      Alias:  client
      Subject:
        Organizational Units:
          client
        Organizations:
          mysql
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       my-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2022-11-22T04:09:34Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-add-tls
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-11-22T04:09:42Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2022-11-22T04:10:07Z
    Message:               Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-add-tls
    Observed Generation:   1
    Reason:                SuccessfullyRestartedPetSet
    Status:                True
    Type:                  RestartPetSet
    Last Transition Time:  2022-11-22T04:10:16Z
    Message:               Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-add-tls
    Observed Generation:   1
    Reason:                SuccessfullyReconfiguredTLS
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-11-22T04:10:21Z
    Message:               Controller has successfully reconfigure the MySQL demo/myops-add-tls
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    16m   KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/myops-add-tls
  Normal  Starting    16m   KubeDB Enterprise Operator  Pausing MySQL databse: demo/mysql
  Normal  Successful  16m   KubeDB Enterprise Operator  Successfully paused MySQL database: demo/mysql for MySQLOpsRequest: myops-add-tls
  Normal  Successful  16m   KubeDB Enterprise Operator  Successfully synced all certificates
  Normal  Starting    16m   KubeDB Enterprise Operator  Restarting Pod: demo/mysql-0
  Normal  Successful  16m   KubeDB Enterprise Operator  Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-add-tls
  Normal  Successful  16m   KubeDB Enterprise Operator  Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-add-tls
  Normal  Starting    16m   KubeDB Enterprise Operator  Resuming MySQL database: demo/mysql
  Normal  Successful  16m   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/mysql
  Normal  Successful  16m   KubeDB Enterprise Operator  Controller has Successfully Reconfigured TLS

```

All tls-secret are created by KubeDB ops-manager operator. Default tls-secret name formed as {mysql-object-name}-{cert-alias}-cert.
NAME                          TYPE                       DATA   AGE
my-ca                         kubernetes.io/tls          2      22m
mysql-auth                    kubernetes.io/basic-auth   2      22m
mysql-client-cert             kubernetes.io/tls          3      18m
mysql-metrics-exporter-cert   kubernetes.io/tls          3      18m
mysql-server-cert             kubernetes.io/tls          3      18m




Now, Let's exec into a database primary node and connect  to the  mysql-shell,

```bash
bash-4.4# ls /etc/mysql/certs
ca.crt	client.crt  client.key	server.crt  server.key
bash-4.4# 
bash-4.4# 
bash-4.4# openssl x509 -in /etc/mysql/certs/client.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=CN=root,OU=client,O=mysql
bash-4.4# 
bash-4.4# 
bash-4.4# 
bash-4.4# mysql -uroot -p$MYSQL_ROOT_PASSWORD -h mysql.demo.svc --ssl-ca=/etc/mysql/certs/ca.crt  --ssl-cert=/etc/mysql/certs/client.crt --ssl-key=/etc/mysql/certs/client.key
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 94
Server version: 8.0.35 MySQL Community Server - GPL

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> \s
--------------
mysql  Ver 8.0.35 for Linux on x86_64 (MySQL Community Server - GPL)

Connection id:		94
Current database:	
Current user:		root@10.244.0.1
SSL:			Cipher in use is TLS_AES_256_GCM_SHA384
Current pager:		stdout
Using outfile:		''
Using delimiter:	;
Server version:		8.0.35 MySQL Community Server - GPL
Protocol version:	10
Connection:		mysql.demo.svc via TCP/IP
Server characterset:	utf8mb4
Db     characterset:	utf8mb4
Client characterset:	latin1
Conn.  characterset:	latin1
TCP port:		3306
Binary data as:		Hexadecimal
Uptime:			13 min 42 sec

Threads: 2  Questions: 522  Slow queries: 0  Opens: 167  Flush tables: 3  Open tables: 86  Queries per second avg: 0.635
--------------

mysql> show variables like '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | ON    |
+--------------------------+-------+
1 row in set (0.00 sec)
```

## Rotate Certificate

Now we are going to rotate the certificate of this database. First let's check the current expiration date of the certificate.

```bash
$ kubectl exec -it mysql-0 -n demo -- bash
bash-4.4# openssl x509 -in /etc/mysql/certs/client.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 20 04:09:37 2023 GM
```

So, the certificate will expire on this time `Feb 20 04:09:37 2023 GMT`.

### Create MySQLRequest

Now we are going to increase it using a MysqlOpsRequest. Below is the yaml of the ops request that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mysql
  tls:
    rotateCertificates: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mysql` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.rotateCertificates` specifies that we want to rotate the certificate of this database.

Let's create the `MySQLOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/myops-rotate.yaml
mysqlopsrequest.ops.kubedb.com/myops-rotate created
```

#### Verify Certificate Rotated Successfully

Let's wait for `MySQLOpsRequest` to be `Successful`.  Run the following command to watch `MySQLOpsRequest` CRO,

```bash
$ kubectl get mysqlopsrequest -n demo
Every 2.0s: kubectl get mysqlopsrequest -n demo
NAME           TYPE             STATUS        AGE
myops-rotate    ReconfigureTLS   Successful    112s
```

We can see from the above output that the `MysqlOpsRequest` has succeeded. If we describe the `MysqlOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mysqlopsrequest -n demo myops-rotate
Name:         myops-rotate
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-11-22T04:39:37Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:rotateCertificates:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-11-22T04:39:37Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-11-22T04:39:38Z
  Resource Version:  718328
  UID:               89798e59-9868-46b9-a11e-d87ad4e9bd9f
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mysql
  Tls:
    Rotate Certificates:  true
  Type:                   ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2022-11-22T04:39:38Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-rotate
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-11-22T04:39:45Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2022-11-22T04:39:54Z
    Message:               Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-rotate
    Observed Generation:   1
    Reason:                SuccessfullyRestartedPetSet
    Status:                True
    Type:                  RestartPetSet
    Last Transition Time:  2022-11-22T04:40:08Z
    Message:               Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-rotate
    Observed Generation:   1
    Reason:                SuccessfullyReconfiguredTLS
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-11-22T04:40:13Z
    Message:               Controller has successfully reconfigure the MySQL demo/myops-rotate
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    52s   KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/myops-rotate
  Normal  Starting    52s   KubeDB Enterprise Operator  Pausing MySQL databse: demo/mysql
  Normal  Successful  52s   KubeDB Enterprise Operator  Successfully paused MySQL database: demo/mysql for MySQLOpsRequest: myops-rotate
  Normal  Successful  45s   KubeDB Enterprise Operator  Successfully synced all certificates
  Normal  Starting    36s   KubeDB Enterprise Operator  Restarting Pod: demo/mysql-0
  Normal  Successful  36s   KubeDB Enterprise Operator  Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-rotate
  Normal  Successful  22s   KubeDB Enterprise Operator  Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-rotate
  Normal  Starting    17s   KubeDB Enterprise Operator  Resuming MySQL database: demo/mysql
  Normal  Successful  17s   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/mysql
  Normal  Successful  17s   KubeDB Enterprise Operator  Controller has Successfully Reconfigured TLS

```

Now, let's check the expiration date of the certificate.

```bash
$ kubectl exec -it mysql-0 -n demo bash
openssl x509 -in /etc/mysql/certs/client.crt -inform PEM -enddate -nameopt RFC2253 -noout
notAfter=Feb 20 04:40:08 2023 GMT

```

As we can see from the above output, the certificate has been rotated successfully.

## Change Issuer/ClusterIssuer

Now, we are going to change the issuer of this database.

- Let's create a new ca certificate and key using a different subject `CN=ca-update,O=kubedb-updated`.

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=ca-updated/O=kubedb-updated"
Generating a RSA private key
..............................................................+++++
......................................................................................+++++
writing new private key to './ca.key'
-----
```

- Now we are going to create a new ca-secret using the certificate files that we have just generated.

```bash
$ kubectl create secret tls mysql-new-ca \
     --cert=ca.crt \
     --key=ca.key \
     --namespace=demo
secret/mysql-new-ca created
```

Now, Let's create a new `Issuer` using the `mysql-new-ca` secret that we have just created. The `YAML` file looks like this:

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: my-new-issuer
  namespace: demo
spec:
  ca:
    secretName: mysql-new-ca
```

Let's apply the `YAML` file:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/new-issuer.yaml
issuer.cert-manager.io/my-new-issuer created
```

### Create MySQLOpsRequest

In order to use the new issuer to issue new certificates, we have to create a `MySQLOpsRequest` CRO with the newly created issuer. Below is the YAML of the `MySQLOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mysql
  tls:
    issuerRef:
      name: my-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mysql` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.issuerRef` specifies the issuer name, kind and api group.

Let's create the `MysqlOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/myops-change-issuer.yaml
mysqlopsrequest.ops.kubedb.com/mops-change-issuer created
```

#### Verify Issuer is changed successfully

Let's wait for `MySQLOpsRequest` to be `Successful`.  Run the following command to watch `MySQLOpsRequest` CRO,

```bash
$ kubectl get mysqlopsrequest -n demo
Every 2.0s: kubectl get mongodbopsrequest -n demo
NAME                  TYPE             STATUS        AGE
myops-change-issuer   ReconfigureTLS   Successful   87s
```

We can see from the above output that the `MysqlOpsRequest` has succeeded. If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mysqlopsrequest -n demo myops-change-issuer
Name:         myops-change-issuer
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-11-22T04:56:51Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:issuerRef:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-11-22T04:56:51Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-11-22T04:56:51Z
  Resource Version:  719824
  UID:               bcc96807-7efb-45e9-add8-54f858ed18d4
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mysql
  Tls:
    Issuer Ref:
      API Group:  cert-manager.io
      Kind:       Issuer
      Name:       my-new-issuer
  Type:           ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2022-11-22T04:56:51Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-change-issuer
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-11-22T04:56:57Z
    Message:               Successfully synced all certificates
    Observed Generation:   1
    Reason:                ReconfigureTLS
    Status:                True
    Type:                  CertificateSynced
    Last Transition Time:  2022-11-22T04:57:06Z
    Message:               Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-change-issuer
    Observed Generation:   1
    Reason:                SuccessfullyRestartedPetSet
    Status:                True
    Type:                  RestartPetSet
    Last Transition Time:  2022-11-22T04:57:15Z
    Message:               Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-change-issuer
    Observed Generation:   1
    Reason:                SuccessfullyReconfiguredTLS
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-11-22T04:57:19Z
    Message:               Controller has successfully reconfigure the MySQL demo/myops-change-issuer
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age    From                        Message
  ----    ------      ----   ----                        -------
  Normal  Starting    2m16s  KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/myops-change-issuer
  Normal  Starting    2m16s  KubeDB Enterprise Operator  Pausing MySQL databse: demo/mysql
  Normal  Successful  2m16s  KubeDB Enterprise Operator  Successfully paused MySQL database: demo/mysql for MySQLOpsRequest: myops-change-issuer
  Normal  Successful  2m10s  KubeDB Enterprise Operator  Successfully synced all certificates
  Normal  Starting    2m1s   KubeDB Enterprise Operator  Restarting Pod: demo/mysql-0
  Normal  Successful  2m1s   KubeDB Enterprise Operator  Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-change-issuer
  Normal  Successful  112s   KubeDB Enterprise Operator  Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-change-issuer
  Normal  Starting    108s   KubeDB Enterprise Operator  Resuming MySQL database: demo/mysql
  Normal  Successful  108s   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/mysql
  Normal  Successful  108s   KubeDB Enterprise Operator  Controller has Successfully Reconfigured TLS

```

Now, Let's exec into a database node and find out the ca subject to see if it matches the one we have provided.

```bash
$ `kubectl exec -it mysql-0 -n demo -- bash`
root@mgo-rs-tls-2:/$ openssl x509 -in /etc/mysql/certs/ca.crt -inform PEM -subject -nameopt RFC2253 -noout
subject=O=kubedb-updated,CN=ca-updated
```

We can see from the above output that, the subject name matches the subject name of the new ca certificate that we have created. So, the issuer is changed successfully.

## Remove TLS from the Database

Now, we are going to remove TLS from this database using a MySQLOpsRequest.

### Create MySQLOpsRequest

Below is the YAML of the `MySQLOpsRequest` CRO that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: mysql
  tls:
    remove: true
```

Here,

- `spec.databaseRef.name` specifies that we are performing reconfigure TLS operation on `mysql` database.
- `spec.type` specifies that we are performing `ReconfigureTLS` on our database.
- `spec.tls.remove` specifies that we want to remove tls from this database.

Let's create the `mysqlOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/reconfigure-tls/reconfigure/yamls/myops-remove.yaml
mysqlopsrequest.ops.kubedb.com/mops-remove created
```

#### Verify TLS Removed Successfully

Let's wait for `MySQLOpsRequest` to be `Successful`.  Run the following command to watch `MySQLOpsRequest` CRO,

```bash
$ kubectl get mysqlopsrequest -n demo
Every 2.0s: kubectl get mysql opsrequest -n demo
NAME          TYPE             STATUS        AGE
myops-remove   ReconfigureTLS   Successful    105s
```

We can see from the above output that the `MySQLOpsRequest` has succeeded. If we describe the `MySQLOpsRequest` we will get an overview of the steps that were followed.

```bash
$ kubectl describe mysqlopsrequest -n demo myops-remove
Name:         myops-remove
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MySQLOpsRequest
Metadata:
  Creation Timestamp:  2022-11-22T05:02:52Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:databaseRef:
        f:tls:
          .:
          f:remove:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2022-11-22T05:02:52Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-ops-manager
    Operation:       Update
    Subresource:     status
    Time:            2022-11-22T05:02:52Z
  Resource Version:  720411
  UID:               43adad9c-e9f6-4cd9-a19e-6ba848901c0c
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  mysql
  Tls:
    Remove:  true
  Type:      ReconfigureTLS
Status:
  Conditions:
    Last Transition Time:  2022-11-22T05:02:52Z
    Message:               Controller has started to Progress the MySQLOpsRequest: demo/myops-remove
    Observed Generation:   1
    Reason:                OpsRequestProgressingStarted
    Status:                True
    Type:                  Progressing
    Last Transition Time:  2022-11-22T05:03:08Z
    Message:               Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-remove
    Observed Generation:   1
    Reason:                SuccessfullyRestartedPetSet
    Status:                True
    Type:                  RestartPetSet
    Last Transition Time:  2022-11-22T05:03:18Z
    Message:               Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-remove
    Observed Generation:   1
    Reason:                SuccessfullyReconfiguredTLS
    Status:                True
    Type:                  DBReady
    Last Transition Time:  2022-11-22T05:03:27Z
    Message:               Controller has successfully reconfigure the MySQL demo/myops-remove
    Observed Generation:   1
    Reason:                OpsRequestProcessedSuccessfully
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason      Age   From                        Message
  ----    ------      ----  ----                        -------
  Normal  Starting    90s   KubeDB Enterprise Operator  Start processing for MySQLOpsRequest: demo/myops-remove
  Normal  Starting    90s   KubeDB Enterprise Operator  Pausing MySQL databse: demo/mysql
  Normal  Successful  90s   KubeDB Enterprise Operator  Successfully paused MySQL database: demo/mysql for MySQLOpsRequest: myops-remove
  Normal  Starting    74s   KubeDB Enterprise Operator  Restarting Pod: demo/mysql-0
  Normal  Successful  74s   KubeDB Enterprise Operator  Successfully restarted MySQL pods for MySQLDBOpsRequest: demo/myops-remove
  Normal  Successful  64s   KubeDB Enterprise Operator  Successfully reconfigured MySQL TLS for MySQLOpsRequest: demo/myops-remove
  Normal  Starting    55s   KubeDB Enterprise Operator  Resuming MySQL database: demo/mysql
  Normal  Successful  55s   KubeDB Enterprise Operator  Successfully resumed MySQL database: demo/mysql
  Normal  Successful  55s   KubeDB Enterprise Operator  Controller has Successfully Reconfigured TLS

```

Now, Let's exec into the database primary node and find out that TLS is disabled or not.

```bash
$ kubectl exec -it -n demo mysql-0 -- mysql  -u root -p 'f8EyKG)mNMIMdS~a'

mysql> show variables like '%require_secure_transport%';
+--------------------------+-------+
| Variable_name            | Value |
+--------------------------+-------+
| require_secure_transport | OFF   |
+--------------------------+-------+

```

So, we can see from the above that, output that tls is disabled successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mysql -n demo mysql
kubectl delete issuer -n demo my-issuer my-new-issuer
kubectl delete mysqlopsrequest myops-add-tls myops-remove mops-rotate myps-change-issuer
kubectl delete ns demo
```

## Next Steps

- [Quickstart MySQL](/docs/guides/mysql/quickstart/index.md) with KubeDB Operator.
- Initialize [MySQL with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/index.md) to deploy MySQL with KubeDB.
- Use [kubedb cli](/docs/guides/mysql/cli/index.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
