---
title: Percona XtraDB Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: px-cluster-guide
    name: Percona XtraDB Cluster Guide
    parent: px-cluster
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# KubeDB - Percona XtraDB Cluster

This tutorial will show you how to use KubeDB to provision a Percona XtraDB Cluster.

## Before You Begin

Before proceeding:

- Read [Percona XtraDB Cluster](/docs/guides/percona-xtradb/overview/overview.md).

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```console
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Sample Percona XtraDB Cluster

To deploy a Percona XtraDB Cluster, specify `.spec.replicas` field in `PerconaXtraDB` object.

The following is an example `PerconaXtraDB` object which creates a Percona XtraDB Cluster with three server (each of them are master that means each server accepts write unlike MySQL Group Replication).

```yaml
apiVersion: kubedb.com/v1alpha1
kind: PerconaXtraDB
metadata:
  name: demo-cluster
  namespace: demo
spec:
  version: "5.7-cluster"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  updateStrategy:
    type: "RollingUpdate"
  terminationPolicy: WipeOut
```

```console
$ kubedb create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/demo-cluster.yaml
perconaxtradb.kubedb.com/my-group created
```

Here,

- `.spec.replicas` specifies the number of required nodes.
- `.spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `PerconaXtraDB` objects using Kubernetes API. When a `PerconaXtraDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching `PerconaXtraDB` object name. KubeDB operator will also create a governing service for the StatefulSet with the name `<percona-xtradb-object-name>-gvr`.

```console
$ kubedb describe px -n demo demo-cluster
Name:         demo-cluster
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha1
Kind:         PerconaXtraDB
Metadata:
  Creation Timestamp:  2019-12-12T14:05:04Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  9704
  Self Link:         /apis/kubedb.com/v1alpha1/namespaces/demo/perconaxtradbs/demo-cluster
  UID:               1a128325-738d-406a-a130-c421a4970892
Spec:
  Database Secret:
    Secret Name:  demo-cluster-auth
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Readiness Probe:
        Exec:
          Command:
            /cluster-check.sh
        Initial Delay Seconds:  30
        Period Seconds:         10
      Resources:
  Replicas:  3
  Service Template:
    Metadata:
    Spec:
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         50Mi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    WipeOut
  Update Strategy:
    Type:   RollingUpdate
  Version:  5.7-cluster
Status:
  Observed Generation:  2$4213139756412538772
  Phase:                Running
Events:
  Type    Reason      Age    From                    Message
  ----    ------      ----   ----                    -------
  Normal  Successful  4m34s  PerconaXtraDB operator  Successfully created Service
  Normal  Successful  7s     PerconaXtraDB operator  Successfully created StatefulSet demo/demo-cluster
  Normal  Successful  7s     PerconaXtraDB operator  Successfully created PerconaXtraDB
  Normal  Successful  6s     PerconaXtraDB operator  Successfully created appbinding
  Normal  Successful  6s     PerconaXtraDB operator  Successfully patched StatefulSet demo/demo-cluster
  Normal  Successful  6s     PerconaXtraDB operator  Successfully patched PerconaXtraDB

$ kubedb get statefulset -n demo
NAME           READY   AGE
demo-cluster   3/3     5m17s

$ kubedb get pvc -n demo
NAME                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-demo-cluster-0   Bound    pvc-afba39ba-1244-4566-be12-bb6ffcb25f31   50Mi       RWO            standard       26m
data-demo-cluster-1   Bound    pvc-b1f64867-b5f4-46e4-b635-ada6fc2cdeaf   50Mi       RWO            standard       4m16s
data-demo-cluster-2   Bound    pvc-11fe5fb3-f58f-4c37-befc-f326e87cef51   50Mi       RWO            standard       2m15s

$ kubedb get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                      STORAGECLASS   REASON   AGE
pvc-11fe5fb3-f58f-4c37-befc-f326e87cef51   50Mi       RWO            Delete           Bound    demo/data-demo-cluster-2   standard                2m40s
pvc-afba39ba-1244-4566-be12-bb6ffcb25f31   50Mi       RWO            Delete           Bound    demo/data-demo-cluster-0   standard                26m
pvc-b1f64867-b5f4-46e4-b635-ada6fc2cdeaf   50Mi       RWO            Delete           Bound    demo/data-demo-cluster-1   standard                4m34s

$ kubedb get service -n demo
NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
demo-cluster       ClusterIP   10.97.246.185   <none>        3306/TCP   7m10s
demo-cluster-gvr   ClusterIP   None            <none>        3306/TCP   7m10s
```

KubeDB operator sets the `.status.phase` to `Running` once the database is successfully created. Run the following command to see the modified `PerconaXtraDB` object:

```yaml
$ kubedb get px -n demo demo-cluster -o yaml
apiVersion: kubedb.com/v1alpha1
kind: PerconaXtraDB
metadata:
  creationTimestamp: "2019-12-12T14:05:04Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: demo-cluster
  namespace: demo
  resourceVersion: "9704"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/demo/perconaxtradbs/demo-cluster
  uid: 1a128325-738d-406a-a130-c421a4970892
spec:
  databaseSecret:
    secretName: demo-cluster-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      readinessProbe:
        exec:
          command:
          - /cluster-check.sh
        initialDelaySeconds: 30
        periodSeconds: 10
      resources: {}
  replicas: 3
  serviceTemplate:
    metadata: {}
    spec: {}
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: WipeOut
  updateStrategy:
    type: RollingUpdate
  version: "5.7-cluster"
status:
  observedGeneration: 2$4213139756412538772
  phase: Running
```

## Connect with Percona XtraDB database

KubeDB operator has created a new Secret called `demo-cluster-auth` **(format: {percona-xtradb-cluster-object-name}-auth)** for storing the password for the superuser. This secret contains a `username` key which contains the **username** for the superuser and a `password` key which contains the **password** for the superuser.

If you want to use an existing secret please specify that when creating the `PerconaXtraDB` object using `.spec.databaseSecret.secretName`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/concepts/databases/percona-xtradb.md#specdatabasesecret).

Now, you can connect to this database from your terminal using the `root` user and password.

```console
$ kubectl get secrets -n demo demo-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo demo-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
LFZAX7DoEg_SMOmL
```

The operator creates a cluster according to the newly created `PerconaXtraDB` object. This cluster has 3 nodes (each of them act as master and accepts writes).

You can connect to any of these cluster nodes. In that case you just need to specify the host name of the corresponding Pod (either PodIP or the fully-qualified-domain-name for that Pod using the governing service named `<percona-xtradb-object-name>-gvr`) by `--host` flag.

```console
# first list the percona-xtradb pods list
$ kubectl get pods -n demo -l kubedb.com/name=demo-cluster
NAME             READY   STATUS    RESTARTS   AGE
demo-cluster-0   1/1     Running   0          23m
demo-cluster-1   1/1     Running   0          21m
demo-cluster-2   1/1     Running   0          19m

# get the governing service
$ kubectl get service demo-cluster-gvr -n demo
NAME               TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
demo-cluster-gvr   ClusterIP   None         <none>        3306/TCP   23m

# list the pods with PodIP
$ kubectl get pods -n demo -l kubedb.com/name=demo-cluster -o jsonpath='{range.items[*]}{.metadata.name} ........... {.status.podIP} ............ {.metadata.name}.demo-cluster-gvr.{.metadata.namespace}{"\\n"}{end}'
demo-cluster-0 ........... 10.244.2.7 ............ demo-cluster-0.demo-cluster-gvr.demo
demo-cluster-1 ........... 10.244.1.5 ............ demo-cluster-1.demo-cluster-gvr.demo
demo-cluster-2 ........... 10.244.2.9 ............ demo-cluster-2.demo-cluster-gvr.demo
```

Now you can connect to the database using the above info.

> Ignore the warning message. It is happening for using password on the command line interface.

```console
# connect to the 1st server
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-0.demo-cluster-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 2nd server
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-1.demo-cluster-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+

# connect to the 3rd server
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "select 1;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+
```

## Check the Cluster Status

Now, you are ready to check newly created group status. Connect and run the following commands from any of the hosts and you will get the same results.

```console
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "show status like 'wsrep%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+-------------------------------------------------+
| Variable_name                    | Value                                           |
+----------------------------------+-------------------------------------------------+
| wsrep_local_state_uuid           | 2cb0d532-1d69-11ea-b1f5-13b666d13d7a            |
...
| wsrep_local_state_comment        | Synced                                          |
...
| wsrep_evs_state                  | OPERATIONAL                                     |
| wsrep_gcomm_uuid                 | 363f5b9b-1d69-11ea-bdca-a622a9b00538            |
...
| wsrep_cluster_size               | 3                                               |
...
| wsrep_cluster_status             | Primary                                         |
| wsrep_connected                  | ON                                              |
...
| wsrep_provider_name              | Galera                                          |
| wsrep_provider_vendor            | Codership Oy <info@codership.com>               |
| wsrep_provider_version           | 3.39(rb3295e6)                                  |
| wsrep_ready                      | ON                                              |
+----------------------------------+-------------------------------------------------+

$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-1.demo-cluster-gvr.demo -e "show status like 'wsrep%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+-------------------------------------------------+
| Variable_name                    | Value                                           |
+----------------------------------+-------------------------------------------------+
| wsrep_local_state_uuid           | 2cb0d532-1d69-11ea-b1f5-13b666d13d7a            |
...
| wsrep_local_state_comment        | Synced                                          |
...
| wsrep_evs_state                  | OPERATIONAL                                     |
| wsrep_gcomm_uuid                 | 696e5ac1-1d69-11ea-af02-8f10d05f9fdf            |
...
| wsrep_cluster_size               | 3                                               |
...
| wsrep_cluster_status             | Primary                                         |
| wsrep_connected                  | ON                                              |
...
| wsrep_provider_name              | Galera                                          |
| wsrep_provider_vendor            | Codership Oy <info@codership.com>               |
| wsrep_provider_version           | 3.39(rb3295e6)                                  |
| wsrep_ready                      | ON                                              |
+----------------------------------+-------------------------------------------------+

$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "show status like 'wsrep%'"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------------------------+-------------------------------------------------+
| Variable_name                    | Value                                           |
+----------------------------------+-------------------------------------------------+
| wsrep_local_state_uuid           | 2cb0d532-1d69-11ea-b1f5-13b666d13d7a            |
...
| wsrep_local_state_comment        | Synced                                          |
...
| wsrep_evs_state                  | OPERATIONAL                                     |
| wsrep_gcomm_uuid                 | 872694a1-1d69-11ea-afc4-268fb8c88de8            |
...
| wsrep_cluster_size               | 3                                               |
...
| wsrep_cluster_status             | Primary                                         |
| wsrep_connected                  | ON                                              |
...
| wsrep_provider_name              | Galera                                          |
| wsrep_provider_vendor            | Codership Oy <info@codership.com>               |
| wsrep_provider_version           | 3.39(rb3295e6)                                  |
| wsrep_ready                      | ON                                              |
+----------------------------------+-------------------------------------------------+
```

Here,

- The value of variable `wsrep_local_state_uuid` in the above 3 tables means the local state is same across the 3 nodes.
- Local state in each node is "Synced"
- Cluster size is 3
- Every node is acting as "Primary"

Let's check the cluster view,

```console
$kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "select * from performance_schema.pxc_cluster_view;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------+--------------------------------------+--------+-------------+---------+
| HOST_NAME      | UUID                                 | STATUS | LOCAL_INDEX | SEGMENT |
+----------------+--------------------------------------+--------+-------------+---------+
| demo-cluster-0 | 363f5b9b-1d69-11ea-bdca-a622a9b00538 | SYNCED |           0 |       0 |
| demo-cluster-1 | 696e5ac1-1d69-11ea-af02-8f10d05f9fdf | SYNCED |           1 |       0 |
| demo-cluster-2 | 872694a1-1d69-11ea-afc4-268fb8c88de8 | SYNCED |           2 |       0 |
+----------------+--------------------------------------+--------+-------------+---------+
```

## Data Availability

In a Percona XtraDB Cluster, you can read/write from/to every node. In this tutorial, we will insert data from, and we will see whether we can get the data from any other nodes.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.
> Don't worry about the warning message. It appears, if you provide password on the command.

```console
# create a database on 'demo-cluster-0'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-0.demo-cluster-gvr.demo -e "CREATE DATABASE playground;"
mysql: [Warning] Using a password on the command line interface can be insecure.

# create a table on 'demo-cluster-1'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-1.demo-cluster-gvr.demo -e "CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));"
mysql: [Warning] Using a password on the command line interface can be insecure.

# insert a row on 'demo-cluster-2'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "INSERT INTO playground.equipment (type, quant, color) VALUES ('slide', 2, 'blue');"
mysql: [Warning] Using a password on the command line interface can be insecure.

# read from 'demo-cluster-0'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-0.demo-cluster-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from 'demo-cluster-1'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-1.demo-cluster-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+

# read from 'demo-cluster-2'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Operating Cluster When Node Fails

To test the cluster behavior during node failure, we will force one of the Primary Pods to restart. When it comes back to the cluster, it becomes the `JOINER` node and one of the existing nodes becomes the `DONOR` node. Then the `JOINER` node becomes `"Synced"` by receiving an IST/SST from the `DONOR` node. Let's see,

```console
kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "select * from performance_schema.pxc_cluster_view;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------+--------------------------------------+--------+-------------+---------+
| HOST_NAME      | UUID                                 | STATUS | LOCAL_INDEX | SEGMENT |
+----------------+--------------------------------------+--------+-------------+---------+
| demo-cluster-0 | 363f5b9b-1d69-11ea-bdca-a622a9b00538 | SYNCED |           0 |       0 |
| demo-cluster-1 | 696e5ac1-1d69-11ea-af02-8f10d05f9fdf | SYNCED |           1 |       0 |
| demo-cluster-2 | 872694a1-1d69-11ea-afc4-268fb8c88de8 | SYNCED |           2 |       0 |
+----------------+--------------------------------------+--------+-------------+---------+

$ kubectl delete pod -n demo demo-cluster-0
pod "demo-cluster-0" deleted

# Let's check the cluster view
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "select * from performance_schema.pxc_cluster_view;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------+--------------------------------------+--------+-------------+---------+
| HOST_NAME      | UUID                                 | STATUS | LOCAL_INDEX | SEGMENT |
+----------------+--------------------------------------+--------+-------------+---------+
| demo-cluster-1 | 696e5ac1-1d69-11ea-af02-8f10d05f9fdf | SYNCED |           0 |       0 |
| demo-cluster-2 | 872694a1-1d69-11ea-afc4-268fb8c88de8 | SYNCED |           1 |       0 |
+----------------+--------------------------------------+--------+-------------+---------+

# Let's check it again
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "select * from performance_schema.pxc_cluster_view;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------+--------------------------------------+--------+-------------+---------+
| HOST_NAME      | UUID                                 | STATUS | LOCAL_INDEX | SEGMENT |
+----------------+--------------------------------------+--------+-------------+---------+
| demo-cluster-0 | 363f5b9b-1d69-11ea-bdca-a622a9b00538 | JOINER |           0 |       0 |
| demo-cluster-1 | 696e5ac1-1d69-11ea-af02-8f10d05f9fdf | DONOR  |           1 |       0 |
| demo-cluster-2 | 872694a1-1d69-11ea-afc4-268fb8c88de8 | SYNCED |           2 |       0 |
+----------------+--------------------------------------+--------+-------------+---------+

# check the logs from 'demo-cluster-2'
# find the lines:
#   "Member 0.0 (demo-cluster-0) requested state transfer from '*any*'. Selected 1.0 (demo-cluster-1)(SYNCED) as donor."
#   "0.0 (demo-cluster-0): State transfer from 1.0 (demo-cluster-1) complete.
#   "Member 0.0 (demo-cluster-0) synced with group."
$ kubectl logs -f demo-cluster-2 -n demo
...
2019-12-13T11:26:47.857079Z 0 [Note] WSREP: Current view of cluster as seen by this node
view (view_id(PRIM,696e5ac1,8)
memb {
	696e5ac1,0
	872694a1,0
	}
joined {
	}
left {
	}
partitioned {
	363f5b9b,0
	}
)
2019-12-13T11:26:47.857106Z 0 [Note] WSREP: Save the discovered primary-component to disk
2019-12-13T11:26:47.865184Z 0 [Note] WSREP: forgetting 363f5b9b (tcp://10.244.2.8:4567)
2019-12-13T11:26:47.865198Z 0 [Note] WSREP: deleting entry tcp://10.244.2.8:4567
2019-12-13T11:26:47.865197Z 0 [Note] WSREP: New COMPONENT: primary = yes, bootstrap = no, my_idx = 1, memb_num = 2
...
2019-12-13T11:26:47.866128Z 0 [Note] WSREP: Quorum results:
	version    = 6,
	component  = PRIMARY,
	conf_id    = 7,
	members    = 2/2 (primary/total),
	act_id     = 17,
	last_appl. = 0,
	protocols  = 0/9/3 (gcs/repl/appl),
	group UUID = 2cb0d532-1d69-11ea-b1f5-13b666d13d7a
...
2019-12-13T11:27:05.861941Z 0 [Note] WSREP: Current view of cluster as seen by this node
view (view_id(PRIM,363f5b9b,9)
memb {
	363f5b9b,0
	696e5ac1,0
	872694a1,0
	}
joined {
	}
left {
	}
partitioned {
	}
)
2019-12-13T11:27:05.862022Z 0 [Note] WSREP: Save the discovered primary-component to disk
2019-12-13T11:27:05.874539Z 0 [Note] WSREP: New COMPONENT: primary = yes, bootstrap = no, my_idx = 2, memb_num = 3
...
2019-12-13T11:27:06.083240Z 0 [Note] WSREP: Quorum results:
	version    = 6,
	component  = PRIMARY,
	conf_id    = 8,
	members    = 2/3 (primary/total),
	act_id     = 17,
	last_appl. = 0,
	protocols  = 0/9/3 (gcs/repl/appl),
	group UUID = 2cb0d532-1d69-11ea-b1f5-13b666d13d7a
2019-12-13T11:27:06.083251Z 0 [Note] WSREP: Flow-control interval: [173, 173]
2019-12-13T11:27:06.083257Z 0 [Note] WSREP: Trying to continue unpaused monitor
2019-12-13T11:27:06.083335Z 2 [Note] WSREP: REPL Protocols: 9 (4, 2)
2019-12-13T11:27:06.083349Z 2 [Note] WSREP: New cluster view: global state: 2cb0d532-1d69-11ea-b1f5-13b666d13d7a:17, view# 9: Primary, number of nodes: 3, my index: 2, protocol version 3
2019-12-13T11:27:06.083356Z 2 [Note] WSREP: Setting wsrep_ready to true
2019-12-13T11:27:06.083364Z 2 [Note] WSREP: Auto Increment Offset/Increment re-align with cluster membership change (Offset: 2 -> 3) (Increment: 2 -> 3)
2019-12-13T11:27:06.083370Z 2 [Note] WSREP: wsrep_notify_cmd is not defined, skipping notification.
2019-12-13T11:27:06.083380Z 2 [Note] WSREP: Assign initial position for certification: 17, protocol version: 4
2019-12-13T11:27:06.085121Z 0 [Note] WSREP: Service thread queue flushed.
2019-12-13T11:27:06.649373Z 0 [Note] WSREP: Member 0.0 (demo-cluster-0) requested state transfer from '*any*'. Selected 1.0 (demo-cluster-1)(SYNCED) as donor.
2019-12-13T11:27:09.255068Z 0 [Note] WSREP: (872694a1, 'tcp://0.0.0.0:4567') turning message relay requesting off
2019-12-13T11:27:19.901794Z 0 [Note] WSREP: 1.0 (demo-cluster-1): State transfer to 0.0 (demo-cluster-0) complete.
2019-12-13T11:27:19.903275Z 0 [Note] WSREP: Member 1.0 (demo-cluster-1) synced with group.
2019-12-13T11:27:25.722838Z 0 [Note] WSREP: 0.0 (demo-cluster-0): State transfer from 1.0 (demo-cluster-1) complete.
2019-12-13T11:27:25.723685Z 0 [Note] WSREP: Member 0.0 (demo-cluster-0) synced with group.

# Let's check it one more time
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "select * from performance_schema.pxc_cluster_view;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----------------+--------------------------------------+--------+-------------+---------+
| HOST_NAME      | UUID                                 | STATUS | LOCAL_INDEX | SEGMENT |
+----------------+--------------------------------------+--------+-------------+---------+
| demo-cluster-0 | 363f5b9b-1d69-11ea-bdca-a622a9b00538 | SYNCED |           0 |       0 |
| demo-cluster-1 | 696e5ac1-1d69-11ea-af02-8f10d05f9fdf | SYNCED |           1 |       0 |
| demo-cluster-2 | 872694a1-1d69-11ea-afc4-268fb8c88de8 | SYNCED |           2 |       0 |
+----------------+--------------------------------------+--------+-------------+---------+
```

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

Now, check the data,

```console
# read data from 'demo-cluster-0'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-0.demo-cluster-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from 'demo-cluster-1'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-1.demo-cluster-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+

# read data from 'demo-cluster-2'
$ kubectl exec -it -n demo demo-cluster-0 -- mysql -u root --password=LFZAX7DoEg_SMOmL --host=demo-cluster-2.demo-cluster-gvr.demo -e "SELECT * FROM playground.equipment;"
mysql: [Warning] Using a password on the command line interface can be insecure.
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  3 | slide |     2 | blue  |
+----+-------+-------+-------+
```

## Cleaning up

Clean what you created in this tutorial.

```console
$ kubectl delete -n demo px/demo-cluster
perconaxtradb.kubedb.com "demo-cluster" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Initialize [PerconaXtraDB with Script](/docs/guides/percona-xtradb/initialization/using-script.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/percona-xtradb/monitoring/using-coreos-prometheus-operator.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-custom-config.md).
- How to use [custom rbac resource](/docs/guides/percona-xtradb/custom-rbac/using-custom-rbac.md) for PerconaXtraDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/snapshot/stash.md).
- Detail concepts of [PerconaXtraDB object](/docs/concepts/databases/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/concepts/catalog/percona-xtradb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
