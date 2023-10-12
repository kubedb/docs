---
title: MySQL Remote Replica Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-clustering-remote-replica
    name: MySQL Remote Replica Guide
    parent: guides-mysql-clustering
    weight: 21
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - MySQL Remote Replica

This tutorial will show you how to use KubeDB to provision a MySQL Remote Replica from a kubedb managed mysql instance. Remote replica can used in in or across cluster


## Before You Begin

Before proceeding:

- Read [mysql replication concept](/docs/guides/mysql/clustering/overview/index.md) to learn about MySQL Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/clustering/remote-replica/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/clustering/group-replication/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).
## Remote Replica

The remote replica allows you to replicate data from an KubeDB managed MySQL server to a read-only mysql server.The whole process  uses MySQL asynchronous replication to keep up-to-date the replica with  source server.
It's useful to use remote replica to scale of read-intensive workloads, can be a workaround for your  BI and analytical workloads and can be geo-replicated.

## Deploy Mysql server

The following is an example `MySQL` object which creates a MySQL Group replicated instance.we will create a tls secure instance since were planing to replicated across cluster

Lets start with creating a secret first to access to database and we will deploy a tls secured instance since were replication across cluster

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/remote-replica/yamls/issuer.yaml
issuer.cert-manager.io/mysql-issuer created
```

### Create Auth Secret

```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cm9vdA==
kind: Secret
metadata:
  name: mysql-singapore-auth
  namespace: demo
type: kubernetes.io/basic-auth
```
```bash 
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/remote-replica/yamls/mysql-singapore-auth.yaml
secret/mysql-singapore-auth created
```
## Deploy MySQL with TLS/SSL configuration

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-singapore
  namespace: demo
spec:
  authSecret:
    name: mysql-singapore-auth
  version: "8.0.31"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "linode-block-storage"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
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

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/remote-replica/yamls/mysql-singapore.yaml
mysql.kubedb.com/mysql created
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created

```bash
$ kubectl get mysql -n demo
NAME              VERSION   STATUS   AGE
mysql-singapore   8.0.31    Ready    22h
```

## Connect with MySQL database

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo mysql-singapore-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mysql-singapore-auth -o jsonpath='{.data.\password}' | base64 -d
pass
```

The operator creates a standalone mysql server for the newly created `MySQL` object.

Now you can connect to the database using the above info. Ignore the warning message. It is happening for using password in the command.


##  Data Insertion

Let's insert some data to the newly created mysql server . we can use the primary service or governing service to connect with the database  
> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# create a database on primary
$ kubectl exec -it -n demo mysql-singapore-0 -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table
$ kubectl exec -it -n demo mysql-singapore-0 -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$  kubectl exec -it -n demo mysql-singapore-0 -c mysql -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from primary
$ kubectl exec -it -n demo mysql-singapore-0 -c mysql -- mysql -u root --password='pass' --host=mysql-singapore-0.mysql-singapore-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

# Exposing to outside world
For Now we will expose our mysql with ingress with to outside world
```bash
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
$ helm upgrade -i ingress-nginx ingress-nginx/ingress-nginx  \
                                      --namespace demo --create-namespace \
                                      --set tcp.3306="demo/mysql-singapore:3306"
```
Let's apply the ingress yaml thats refers to `mysql-singpore` service
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: mysql-singapore
  namespace: demo  
spec:
  ingressClassName: nginx
  rules:
  - host: mysql-singapore.something.org
    http:
      paths:
      - backend:
          service:
            name: mysql-singapore
            port:
              number: 3306
        path: /
        pathType: Prefix
```
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/remote-replica/yamls/mysql-ingress.yaml
ingress.networking.k8s.io/mysql-singapore created
$ kubectl get ingress -n demo
NAME              CLASS   HOSTS                           ADDRESS          PORTS   AGE
mysql-singapore   nginx   mysql-singapore.something.org   172.104.37.147   80      22h
```
Now will be able to communicate from another cluster to our source database
# Prepare for Remote Replica
We wil use the [kubedb_plugin](somelink) for generating configuration for remote replica. It will create the appbinding and and necessary secrets to connect with source server
```bash
$ kubectl dba remote-config mysql -n demo mysql-singapore -uremote -ppass -d 172.104.37.147 -y
home/mehedi/go/src/kubedb.dev/yamls/mysql/mysql-singapore-remote-config.yaml
```
#  Create  Remote Replica
We have prepared another cluster in london region for replicating across cluster. follow the installation instruction [above](/docs/README.md).

### create sourceRef 

We will apply the generated config from kubeDB plugin to create the source refs and secrets for it

```bash
$ kubectl apply -f  /home/mehedi/go/src/kubedb.dev/yamls/bank_abc/mysql/mysql-singapore-remote-config.yaml

secret/mysql-singapore-remote-replica-auth created
secret/mysql-singapore-client-cert-remote created
appbinding.appcatalog.appscode.com/mysql-singapore created

$ kubectl get appbinding -n  demo
NAME              TYPE               VERSION   AGE
mysql-singapore   kubedb.com/mysql   8.0.31    4m17s
```

### create remote replica auth 
we will need to use the same auth secrets for remote replicas as well since operations like clone also replicated the auth-secrets from source server
```yaml
apiVersion: v1
data:
  password: cGFzcw==
  username: cm9vdA==
kind: Secret
metadata:
  name: mysql-london-auth
  namespace: demo
type: kubernetes.io/basic-auth
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/remote-replica/yamls/mysql-london-auth.yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: mysql-london
  namespace: demo
spec:
  authSecret:
    name: mysql-london-auth
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
    disableWriteCheck: true
  version: "8.0.31"
  replicas: 1
  topology:
    mode: RemoteReplica
    remoteReplica:
      sourceRef:
        name: mysql-singapore
        namespace: demo
  storageType: Durable
  storage:
    storageClassName: "linode-block-storage"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  terminationPolicy: WipeOut
```
Here,

- `spec.topology` contains the information about the mysql server.
- `spec.topology.mode` we are defining the server will be working a `Remote Replica`.
- `spec.topology.remoteReplica.sourceref` we are referring to source to read. The  mysql instance we previously created.
- `spec.terminationPolicy` specifies what KubeDB should do when a user try to delete the operation of MySQL CR. *Wipeout* means that the database will be deleted without restrictions. It can also be "Halt", "Delete" and "DoNotTerminate". Learn More about these [HERE](https://kubedb.com/docs/latest/guides/mysql/concepts/database/#specterminationpolicy).
```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clustering/remote-replica/mysql-london.yaml
mysql.kubedb.com/mysql-london created
```

Now we will be able to see kubedb will provision a Remote Replica from the source mysql instance. Lets checkout out the statefulSet , pvc , pv and services associated with it
.
KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified `MySQL` object:
```bash
$ kubectl get mysql -n demo 
NAME           VERSION   STATUS   AGE
mysql-london   8.0.31    Ready    7m17s
```

##  Validate Remote Replica
Since both source and replica database are in the ready state. we can validate Remote Replica is working properly by checking the replication status 

```bash
$ kubectl exec -it -n demo mysql-london-0 -c mysql -- mysql -u root --password='pass' --host=mysql-london-0.mysql-london-pods.demo -e "show slave status\G" 
mysql: [Warning] Using a password on the command line interface can be insecure.
*************************** 1. row ***************************
               Slave_IO_State: Waiting for source to send event
                  Master_Host: 172.104.37.147
                  Master_User: remote
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: binlog.000001
          Read_Master_Log_Pos: 4698131
               Relay_Log_File: mysql-london-0-relay-bin.000007
                Relay_Log_Pos: 1415154
        Relay_Master_Log_File: binlog.000001
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
            ....           
```

# Read Data
In the previous step we have inserted into the primary pod. In the next step we will read from secondary pods to determine whether the data has been successfully copied to the secondary pods.
```bash
# read from secondary-1
$ kubectl exec -it -n demo mysql-london-0 -c mysql -- mysql -u root --password='pass'  --host=mysql-london-0.mysql-london-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Write on Secondary Should Fail

Only, primary member preserves the write permission. No secondary can write data.

## Automatic Failover

To test automatic failover, we will force the primary Pod to restart. Since the primary member (`Pod`) becomes unavailable, the rest of the members will elect a new primary for these group. When the old primary comes back, it will join the group as a secondary member.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# delete the primary Pod mysql-london-0
$ kubectl delete pod mysql-london-0 -n demo
pod "mysql-london-0" deleted

# check the new primary ID
$ kubectl exec -it -n demo mysql-london-0 -c mysql -- mysql -u root --password='pass' --host=mysql-london-0.mysql-london-pods.demo -e "show slave status\G" 
mysql: [Warning] Using a password on the command line interface can be insecure.
*************************** 1. row ***************************
               Slave_IO_State: Waiting for source to send event
                  Master_Host: mysql.demo.svc
                  Master_User: root
                  Master_Port: 3306
                Connect_Retry: 60
              Master_Log_File: binlog.000002
          Read_Master_Log_Pos: 214789
               Relay_Log_File: mysql-london-0-relay-bin.000002
                Relay_Log_Pos: 186366
        Relay_Master_Log_File: binlog.000002
             Slave_IO_Running: Yes
            Slave_SQL_Running: Yes
        ...

# read data after recovery
$ kubectl exec -it -n demo mysql-london-0 -c mysql -- mysql -u root --password='pass' --host=mysql-read-2.mysql-read-pods.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  7 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Cleaning up

Clean what you created in this tutorial.

```bash
kubectl delete -n demo my/mysql-singapore
kubectl delete -n demo my/mysql-london
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
