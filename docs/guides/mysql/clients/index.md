---
title: Connecting to a MySQL Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-clients
    name: Connecting to a MySQL Cluster
    parent: guides-mysql
    weight: 105
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Connecting to a MySQL Cluster

KubeDB creates separate services for primary and secondary replicas. In this tutorial, we are going to show you how to connect your application with primary or secondary replicas using those services.

## Before You Begin

- Read [mysql group replication concept](/docs/guides/mysql/clustering/overview/index.md) to learn about MySQL Group Replication.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

- You need to have a mysql client. If you don't have a mysql client install in your local machine, you can install from [here](https://dev.mysql.com/doc/mysql-installation-excerpt/8.0/en/)

> Note: YAML files used in this tutorial are stored in [guides/mysql/clients/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/clients/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy MySQL Cluster

Here, we are going to deploy a `MySQL` group replication using a supported version by KubeDB operator. Then we are going to connect with the cluster using two separate services.

Below is the YAML of `MySQL` group replication with 3 members (one is a primary member and the two others are secondary members) that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: my-group
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
  terminationPolicy: WipeOut
```

Let's create the MySQL CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/clients/yamls/group-replication.yaml
mysql.kubedb.com/my-group created
```

KubeDB operator watches for `MySQL` objects using Kubernetes API. When a `MySQL` object is created, KubeDB operator will create a new StatefulSet and two separate Services for client connection with the cluster. The services have the following format:

- `<mysql-object-name>` has both read and write operation.
- `<mysql-object-name>-standby` has only read operation.

Now, wait for the `MySQL` is going to `Running` state and also wait for `SatefulSet` and `services` going to the `Ready` state.

```bash
$ watch -n 3 kubectl get my -n demo my-group
Every 3.0s: kubectl get my -n demo my-group                      suaas-appscode: Wed Sep  9 10:54:34 2020

NAME       VERSION   STATUS    AGE
my-group   8.0.23    Running   16m

$ watch -n 3 kubectl get sts -n demo my-group
ery 3.0s: kubectl get sts -n demo my-group                     suaas-appscode: Wed Sep  9 10:53:52 2020

NAME       READY   AGE
my-group   3/3     15m

$ kubectl get service -n demo
my-group           ClusterIP   10.109.133.141   <none>        3306/TCP   31s
my-group-pods      ClusterIP   None             <none>        3306/TCP   31s
my-group-standby   ClusterIP   10.110.47.184    <none>        3306/TCP   31s
```

If you describe the object, you can find more details here,

```bash
$ kubectl dba describe my -n demo my-group
Name:               my-group
Namespace:          demo
CreationTimestamp:  Mon, 15 Mar 2021 18:18:35 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"my-group","namespace":"demo"},"spec":{"replicas":3,"storage":{"...
Replicas:           3  total
Status:             Provisioning
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

StatefulSet:          
  Name:               my-group
  CreationTimestamp:  Mon, 15 Mar 2021 18:18:35 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=my-group
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:        <none>
  Replicas:           824635568200 desired | 3 total
  Pods Status:        2 Running / 1 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         my-group
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.109.133.141
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.15:3306

Service:        
  Name:         my-group-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.15:3306,10.244.0.17:3306,10.244.0.19:3306

Service:        
  Name:         my-group-standby
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.110.47.184
  Port:         standby  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.17:3306

Auth Secret:
  Name:         my-group-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=my-group
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mysqls.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    username:  4 bytes
    password:  16 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MySQL","metadata":{"annotations":{},"name":"my-group","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"WipeOut","topology":{"group":{"name":"dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"},"mode":"GroupReplication"},"version":"8.0.23"}}

    Creation Timestamp:  2021-03-15T12:18:35Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    my-group
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mysqls.kubedb.com
    Name:                            my-group
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    my-group
        Path:    /
        Port:    3306
        Scheme:  mysql
      URL:       tcp(my-group:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  mysql-backup-8.0.21-v1
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  mysql-restore-8.0.21-v1
    Secret:
      Name:   my-group-auth
    Type:     kubedb.com/mysql
    Version:  8.0.23

Events:
  Type    Reason      Age   From             Message
  ----    ------      ----  ----             -------
  Normal  Successful  2m    KubeDB Operator  Successfully created governing service
  Normal  Successful  2m    KubeDB Operator  Successfully created service for primary/standalone
  Normal  Successful  2m    KubeDB Operator  Successfully created service for secondary replicas
  Normal  Successful  2m    KubeDB Operator  Successfully created database auth secret
  Normal  Successful  2m    KubeDB Operator  Successfully created StatefulSet
  Normal  Successful  2m    KubeDB Operator  Successfully created appbinding
  Normal  Successful  2m    KubeDB Operator  Successfully patched StatefulSet
```

Our database cluster is ready to connect.

## How KubeDB distinguish between primary and secondary Replicas

`KubeDB` add a `sidecar` into the pod beside the database container. This sidecar is used to add a label into the pod to distinguish between primary and secondary replicas. The label is added into the pods as follows:

- `mysql.kubedb.com/role:primary` are added for primary.
- `mysql.kubedb.com/role:secondary` are added for secondary.

Let's verify that the `mysql.kubedb.com/role:<primary/secondary>` label are added into the StatefulSet's replicas,

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group -A -o=custom-columns='Name:.metadata.name,Labels:metadata.labels,PodIP:.status.podIP'
Name         Labels                                                                                                                                                                           PodIP
my-group-0   map[controller-revision-hash:my-group-55b9f49f98 app.kubernetes.io/name:mysqls.kubedb.com app.kubernetes.io/instance:my-group mysql.kubedb.com/role:primary statefulset.kubernetes.io/pod-name:my-group-0]     10.244.1.8
my-group-1   map[controller-revision-hash:my-group-55b9f49f98 app.kubernetes.io/name:mysqls.kubedb.com app.kubernetes.io/instance:my-group mysql.kubedb.com/role:secondary statefulset.kubernetes.io/pod-name:my-group-1]   10.244.2.11
my-group-2   map[controller-revision-hash:my-group-55b9f49f98 app.kubernetes.io/name:mysqls.kubedb.com app.kubernetes.io/instance:my-group mysql.kubedb.com/role:secondary statefulset.kubernetes.io/pod-name:my-group-2]   10.244.2.13
```

You can see from the above output that the `my-group-0` pod is selected as a primary member in our existing database cluster. It has the `mysql.kubedb.com/role:primary` label and the podIP is `10.244.1.8`. Besides, the rest of the replicas are selected as a secondary member which has `mysql.kubedb.com/role:secondary` label.

KubeDB creates two separate services(already shown above) to connect with the database cluster. One for connecting to the primary replica and the other for secondaries. The service who is dedicated to connecting to the primary has both permissions of read and write operation and the service who is dedicated to other secondaries has only the permission of read operation.

You can find the service which selects for primary replica have the following selector,

```bash
$ kubectl get svc -n demo my-group -o json | jq '.spec.selector'
{
  "app.kubernetes.io/instance": "my-group",
  "app.kubernetes.io/managed-by": "kubedb.com",
  "app.kubernetes.io/name": "mysqls.kubedb.com",
  "kubedb.com/role": "primary"
}
```

If you get the endpoint of the above service, you will see the podIP of the primary replica,

```bash
$ kubectl get endpoints -n demo my-group
NAME       ENDPOINTS         AGE
my-group   10.244.1.8:3306   5h49m
```

You can also find the service which selects for secondary replicas have the following selector,

```bash
$ kubectl get svc -n demo my-group-replicas -o json | jq '.spec.selector'
{
  "app.kubernetes.io/name": "mysqls.kubedb.com",
  "app.kubernetes.io/instance": "my-group",
  "mysql.kubedb.com/role": "secondary"
}
```

If you get the endpoint of the above service, you will see the podIP of the secondary replicas,

```bash
$ kubectl get endpoints -n demo my-group-replicas
NAME                ENDPOINTS                           AGE
my-group-replicas   10.244.2.11:3306,10.244.2.13:3306   5h53m
```

## Connecting Information

KubeDB operator has created a new Secret called `my-group-auth` **(format: {mysql-object-name}-auth)** for storing the password for `mysql` superuser. This secret contains a `username` key which contains the **username** for MySQL superuser and a `password` key which contains the **password** for MySQL superuser.

Now, you can connect to this database from your terminal using the `mysql` user and password.

```bash
$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo my-group-auth -o jsonpath='{.data.\password}' | base64 -d
RmxLjEomvE6tVj4-
```

You can connect to any of these group members. In that case, you just need to specify the hostname of that member Pod (either PodIP or the fully-qualified-domain-name for that Pod using any of the services) by `--host` flag.

## Connecting with Primary Replica

The primary replica has both the permission of read and write operation. So the clients are able to perform both operations using the service `my-group` which select primary replica(already shown above). First, we will insert data into the database cluster, then we will see whether we insert into the cluster using `my-group` service.

At first, we are going to port-forward the service to connect to the database cluster from the outside of the cluster.

Let's port-forward the `my-group` service using the following command,

```bash
$ kubectl port-forward service/my-group -n demo 8081:3306
Forwarding from 127.0.0.1:8081 -> 3306
Forwarding from [::1]:8081 -> 3306
```

>For testing purpose, we need to have a mysql client to connect with the cluster. If you don't have a client in your local machine, you can install from [here](https://dev.mysql.com/doc/mysql-installation-excerpt/8.0/en/)

**Write Operation :**

```bash
# create a database on cluster
$ mysql -uroot -pRmxLjEomvE6tVj4- --port=8081 --host=127.0.0.1 -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.


# create a table
$ mysql -uroot -pRmxLjEomvE6tVj4- --port=8081 --host=127.0.0.1 -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.


# insert a row
$ mysql -uroot -pRmxLjEomvE6tVj4- --port=8081 --host=127.0.0.1 -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.
```

**Read Operation :**

```bash
# read data from cluster
$ mysql -uroot -pRmxLjEomvE6tVj4- --port=8081 --host=127.0.0.1 -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

You can see from the above output that both write and read operations are performed successfully using primary pod selector service named `my-group`.

## Connection with Secondary Replicas

The secondary replica has only the permission of read operation. So the clients are able to perform only read operation using the service `my-group-replicas` which select only secondary replicas(already shown above). First, we will try to insert data into the database cluster, then we will read existing data from the cluster using `my-group-replicas` service.

At first, we are going to port-forward the service to connect to the database cluster from the outside of the cluster.

Let's port-forward the `my-group-replicas` service using the following command,

```bash
$ kubectl port-forward service/my-group-replicas -n demo 8080:3306
Forwarding from 127.0.0.1:8080 -> 3306
Forwarding from [::1]:8080 -> 3306
```

**Write Operation:**

```bash
# in our database cluster we have created a database and a table named playground and equipment respectively. so we will try to insert data into it.
# insert a row
$ mysql -uroot -pRmxLjEomvE6tVj4- --port=8080 --host=127.0.0.1 -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 3, 'black');"
mysql: [Warning] Using a password on the command line interface can be insecure.
ERROR 1290 (HY000) at line 1: The MySQL server is running with the --super-read-only option so it cannot execute this statement
```

**Read Operation:**

```bash
# read data from cluster
$ mysql -uroot -pRmxLjEomvE6tVj4- --port=8080 --host=127.0.0.1 -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
```

You can see from the above output that only read operations are performed successfully using secondary pod selector service named `my-group-replicas`. No data is inserted by using this service. The error `--super-read-only` indicates that the secondary pod has only read permission.

## Automatic Failover

To test automatic failover, we will force the primary Pod to restart. Since the primary member (`Pod`) becomes unavailable, the rest of the members will elect a new primary for these group. When the old primary comes back, it will join the group as a secondary member.

First, delete the primary pod `my-gorup-0` using the following command,

```bash
$ kubectl delete pod my-group-0 -n demo
pod "my-group-0" deleted
```

Now wait for a few minute to automatically elect the primary replica and also wait for the services endpoint update for new primary and secondary replicas,

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/name=mysqls.kubedb.com,app.kubernetes.io/instance=my-group -A -o=custom-columns='Name:.metadata.name,Labels:metadata.labels,PodIP:.status.podIP'
Name         Labels                                                                                                                                                                           PodIP
my-group-0   map[controller-revision-hash:my-group-55b9f49f98 app.kubernetes.io/name:mysqls.kubedb.com app.kubernetes.io/instance:my-group mysql.kubedb.com/role:secondary statefulset.kubernetes.io/pod-name:my-group-0]   10.244.2.18
my-group-1   map[controller-revision-hash:my-group-55b9f49f98 app.kubernetes.io/name:mysqls.kubedb.com app.kubernetes.io/instance:my-group mysql.kubedb.com/role:secondary statefulset.kubernetes.io/pod-name:my-group-1]   10.244.2.11
my-group-2   map[controller-revision-hash:my-group-55b9f49f98 app.kubernetes.io/name:mysqls.kubedb.com app.kubernetes.io/instance:my-group mysql.kubedb.com/role:primary statefulset.kubernetes.io/pod-name:my-group-2]     10.244.2.13
```

You can see from the above output that `my-group-2` pod is elected as a primary automatically and the others become secondary.

If you get the endpoint of the `my-group` service, you will see the podIP of the primary replica,

```bash
$ kubectl get endpoints -n demo my-group
NAME       ENDPOINTS          AGE
my-group   10.244.2.13:3306   111m
```

If you get the endpoint of the `my-group-replicas` service, you will see the podIP of the secondary replicas,

```bash
$ kubectl get endpoints -n demo my-group-replicas
NAME                ENDPOINTS                           AGE
my-group-replicas   10.244.2.11:3306,10.244.2.18:3306   112m
```

## Cleaning up

Clean what you created in this tutorial.

```bash
$ kubectl patch -n demo my/my-group -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo my/my-group

$ kubectl delete ns demo
```

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLDBVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
