---
title: Initialize PerconaXtraDB using Script Source
menu:
  docs_{{ .version }}:
    identifier: pxc-script-source-initialization
    name: Using Script
    parent: pxc-initialization-perconaxtradb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize PerconaXtraDB with Script

KubeDB supports PerconaXtraDB database initialization. This tutorial will show you how to use KubeDB to initialize a PerconaXtraDB cluster from a script.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/percona-xtradb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/percona-xtradb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Initialization Scripts

PerconaXtraDB supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use an `init.sql` script to create a TABLE `kubedb_table` in the `test` database.

We will use a ConfigMap as the script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

At first, we will create a ConfigMap from an `init.sql` file. Then, we will provide this ConfigMap as a script source in `init.script` of the PerconaXtraDB CRD spec.

Let's create a ConfigMap with the initialization script:

```bash
$ kubectl create configmap -n demo pxc-init-script \
--from-literal=init.sql="$(curl -fsSL https://raw.githubusercontent.com/kubedb/percona-xtradb-init-scripts/master/init.sql)"
configmap/pxc-init-script created
```

## Create PerconaXtraDB with Script Source

Following YAML describes the PerconaXtraDB object with `init.script`:

```yaml
apiVersion: kubedb.com/v1
kind: PerconaXtraDB
metadata:
  name: script-pxc
  namespace: demo
spec:
  version: "8.4.3"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: pxc-init-script
  deletionPolicy: WipeOut
```

Here,

- `init.script` specifies the scripts used to initialize the database when it is being created. Scripts are executed alphabetically. The `*.sql`, `*.sql.gz`, and `*.sh` scripts stored in the root folder of the volume source will be executed. Scripts inside child folders are skipped.

VolumeSource provided in `init.script` will be mounted in the Pod and will be executed while creating PerconaXtraDB.

Now, let's create the PerconaXtraDB CRD using the YAML shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/percona-xtradb/initialization/script-pxc.yaml
perconaxtradb.kubedb.com/script-pxc created
```

Now, wait until PerconaXtraDB goes in `Ready` state. Verify that the cluster is in `Ready` state using the following command:

```bash
$ kubectl get perconaxtradb -n demo script-pxc
NAME         VERSION   STATUS   AGE
script-pxc   8.4.3     Ready    3m
```

You can use `kubectl dba describe` command to view which resources have been created by KubeDB for this PerconaXtraDB object:

```bash
$ kubectl dba describe perconaxtradb -n demo script-pxc
Name:         script-pxc
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1
Kind:         PerconaXtraDB
Metadata:
  Creation Timestamp:  2026-07-01T11:40:55Z
  Finalizers:
    kubedb.com
  Generation:        5
  Resource Version:  1101867
  UID:               634db392-1587-45aa-a614-93f5b5732138
Spec:
  Allowed Schemas:
    Namespaces:
      From:  Same
  Auth Secret:
    Active From:  2026-07-01T11:40:57Z
    API Group:    
    Kind:         Secret
    Name:         script-pxc-auth
  Auto Ops:
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Init:
    Initialized:  true
    Script:
      Config Map:
        Name:  pxc-init-script
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  perconaxtradb
        Resources:
          Limits:
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
        Name:      px-coordinator
        Resources:
          Limits:
            Memory:  256Mi
          Requests:
            Cpu:     200m
            Memory:  256Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  px-init
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  256Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     1001
          Run As Non Root:  true
          Run As User:      1001
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            1001
      Service Account Name:  script-pxc
  Replicas:                  3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  local-path
  Storage Type:          Durable
  System User Secrets:
    Monitor User Secret:
      API Group:  
      Kind:       
      Name:       script-pxc-monitor
    Replication User Secret:
      API Group:  
      Kind:       
      Name:       script-pxc-replication
  Version:        8.4.3
Status:
  Conditions:
    Last Transition Time:  2026-07-01T11:40:55Z
    Message:               The KubeDB operator has started the provisioning of PerconaXtraDB: demo/script-pxc
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2026-07-01T11:42:06Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2026-07-01T11:43:49Z
    Message:               database script-pxc/demo is accepting connection
    Observed Generation:   5
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2026-07-01T11:43:49Z
    Message:               database script-pxc/demo is ready
    Observed Generation:   5
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2026-07-01T11:43:42Z
    Message:               The PerconaXtraDB: demo/script-pxc is successfully provisioned.
    Observed Generation:   4
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     4
  Phase:                   Ready
Events:
  Type    Reason        Age   From             Message
  ----    ------        ----  ----             -------
  Normal  PhaseChanged  17m   KubeDB Operator  Phase changed from  to Provisioning.
  Normal  Successful    17m   KubeDB Operator  Successfully created governing service
  Normal  Successful    17m   KubeDB Operator  Successfully created Service
  Normal  Successful    17m   KubeDB Operator  Successfully created PetSet demo/script-pxc
  Normal  Successful    17m   KubeDB Operator  Successfully created PerconaXtraDB
  Normal  Successful    17m   KubeDB Operator  Successfully created appbinding
  Normal  PhaseChanged  15m   KubeDB Operator  Phase changed from Provisioning to Ready.
```

## Verify Initialization

Now let's connect to our PerconaXtraDB cluster to verify that the database has been initialized successfully.

**Connection Information:**

- Host name/address: you can use any of these
  - Service: `script-pxc.demo`
  - Pod IP: (`$ kubectl get pods script-pxc-0 -n demo -o yaml | grep podIP`)
- Port: `3306`

- Username: Run the following command to get the *username*:

  ```bash
  $ kubectl get secret -n demo script-pxc-auth -o jsonpath='{.data.username}' | base64 -d
  root
  ```

- Password: Run the following command to get the *password*:

  ```bash
  $ kubectl get secret -n demo script-pxc-auth -o jsonpath='{.data.password}' | base64 -d
    nsTqGdVwR!~DA(t
  ```

Now, connect to the PerconaXtraDB cluster and run the following query to confirm initialization:

```bash
$ kubectl exec -it -n demo script-pxc-0 -- mysql -u root --password='nsTqGdVwR!~DA(ta' -e "SHOW TABLES FROM mysql;"
Defaulted container "perconaxtradb" out of: perconaxtradb, px-coordinator, px-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
+------------------------------------------------------+
| Tables_in_mysql                                      |
+------------------------------------------------------+
| columns_priv                                         |
| component                                            |
| db                                                   |
| default_roles                                        |
| engine_cost                                          |
| func                                                 |
| general_log                                          |
| global_grants                                        |
| gtid_executed                                        |
| help_category                                        |
| help_keyword                                         |
| help_relation                                        |
| help_topic                                           |
| innodb_index_stats                                   |
| innodb_table_stats                                   |
| kubedb_table                                         |
| ndb_binlog_index                                     |
| password_history                                     |
| plugin                                               |
| procs_priv                                           |
| proxies_priv                                         |
| replication_asynchronous_connection_failover         |
| replication_asynchronous_connection_failover_managed |
| replication_group_configuration_version              |
| replication_group_member_actions                     |
| role_edges                                           |
| server_cost                                          |
| servers                                              |
| slave_master_info                                    |
| slave_relay_log_info                                 |
| slave_worker_info                                    |
| slow_log                                             |
| tables_priv                                          |
| time_zone                                            |
| time_zone_leap_second                                |
| time_zone_name                                       |
| time_zone_transition                                 |
| time_zone_transition_type                            |
| user                                                 |
| wsrep_cluster                                        |
| wsrep_cluster_members                                |
| wsrep_streaming_log                                  |
+------------------------------------------------------+

```

We can see the TABLE `kubedb_table` in `test` database which was created through initialization.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo pxc/script-pxc -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo pxc/script-pxc

$ kubectl delete -n demo configmap/pxc-init-script
$ kubectl delete ns demo
```

## Next Steps

- Want to setup a PerconaXtraDB cluster? Check the [clustering guide](/docs/guides/percona-xtradb/clustering/overview/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
