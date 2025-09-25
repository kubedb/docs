---
title: Initialize SingleStore using Script
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-initialization-usingscript
    name: Using Script
    parent: guides-sdb-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize SingleStore using Script

This tutorial will show you how to use KubeDB to initialize a SingleStore database with \*.sql, \*.sh and/or \*.sql.gz script.
In this tutorial we will use .sql script stored in GitHub repository [kubedb/singlestore-init-scripts](https://github.com/kubedb/singlestore-init-scripts).

> Note: The yaml files that are used in this tutorial are stored [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/initialization/using-script/example) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs)

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Initialization Scripts

SingleStore supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use `init.sql` script from [singlestore-init-scripts](https://github.com/kubedb/singlestore-init-scripts) git repository to create a TABLE `kubedb_write_check` in `kubedb_test` database.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.sql` file. Then, we will provide this ConfigMap as script source in `init.script` of SingleStore crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo sdb-init-script \
--from-literal=init.sql="$(curl -fsSL https://github.com/kubedb/singlestore-init-scripts/raw/master/init.sql)"
configmap/sdb-init-script created
```

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Create a SingleStore database with Init-Script

Below is the `SingleStore` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-sample
  namespace: demo
spec:
  version: "8.7.10"
  init:
    script:
      configMap:
        name: sdb-init-script
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    kind: Secret
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/singlestore/Initialization/demo-1.yaml
singlestore.kubedb.com/singlestore-init-script created
```

Here,

- `spec.init.script` specifies a script source used to initialize the database before database server starts. The scripts will be executed alphabatically. In this tutorial, a sample .sql script from the git repository `https://github.com/kubedb/singlestore-init-scripts.git` is used to create a test database. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) instead of `ConfigMap`.  The \*.sql, \*sql.gz and/or \*.sh sripts that are stored inside the root folder will be executed alphabatically. The scripts inside child folders will be skipped.

KubeDB operator watches for `SingleStore` objects using Kubernetes api. When a `SingleStore` object is created, KubeDB operator will create a new PetSet and a Service with the matching `SingleStore` object name. KubeDB operator will also create a governing service for PetSets with the name `kubedb`, if one is not already present. No SingleStore specific RBAC roles are required for [RBAC enabled clusters](/docs/setup/README.md#using-yaml).

```yaml
$ kubectl get sdb -n demo sdb-sample -oyaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Singlestore","metadata":{"annotations":{},"name":"sdb-sample","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","init":{"script":{"configMap":{"name":"sdb-init-script"}}},"licenseSecret":{"name":"license-secret"},"storageType":"Durable","topology":{"aggregator":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"600m","memory":"2Gi"},"requests":{"cpu":"600m","memory":"2Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"}},"leaf":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"600m","memory":"2Gi"},"requests":{"cpu":"600m","memory":"2Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"10Gi"}},"storageClassName":"standard"}}},"version":"8.7.10"}}
  creationTimestamp: "2024-10-03T07:00:56Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: sdb-sample
  namespace: demo
  resourceVersion: "124012"
  uid: ccfe9d0e-6f13-4187-b652-4e157a21568e
spec:
  authSecret:
    name: sdb-sample-root-cred
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  init:
    script:
      configMap:
        name: sdb-init-script
  licenseSecret:
    kind: Secret
    name: license-secret
  storageType: Durable
  topology:
    aggregator:
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                cpu: 600m
                memory: 2Gi
              requests:
                cpu: 600m
                memory: 2Gi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          - name: singlestore-coordinator
            resources:
              limits:
                memory: 256Mi
              requests:
                cpu: 200m
                memory: 256Mi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          initContainers:
          - name: singlestore-init
            resources:
              limits:
                memory: 512Mi
              requests:
                cpu: 200m
                memory: 512Mi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          podPlacementPolicy:
            name: default
          securityContext:
            fsGroup: 999
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    leaf:
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                cpu: 600m
                memory: 2Gi
              requests:
                cpu: 600m
                memory: 2Gi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          - name: singlestore-coordinator
            resources:
              limits:
                memory: 256Mi
              requests:
                cpu: 200m
                memory: 256Mi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          initContainers:
          - name: singlestore-init
            resources:
              limits:
                memory: 512Mi
              requests:
                cpu: 200m
                memory: 512Mi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          podPlacementPolicy:
            name: default
          securityContext:
            fsGroup: 999
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: standard
  version: 8.7.10
status:
  conditions:
  - lastTransitionTime: "2024-10-03T07:01:02Z"
    message: 'The KubeDB operator has started the provisioning of Singlestore: demo/sdb-sample'
    observedGeneration: 3
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2024-10-03T07:11:23Z"
    message: All leaf replicas are ready for Singlestore demo/sdb-sample
    observedGeneration: 3
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2024-10-03T07:02:13Z"
    message: database demo/sdb-sample is accepting connection
    observedGeneration: 3
    reason: AcceptingConnection
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2024-10-03T07:02:13Z"
    message: database demo/sdb-sample is ready
    observedGeneration: 3
    reason: AllReplicasReady
    status: "True"
    type: Ready
  - lastTransitionTime: "2024-10-03T07:02:14Z"
    message: 'The Singlestore: demo/sdb-sample is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  phase: Ready
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created.

Now, we will connect to this database and check the data inserted by the initlization script.

```bash
# Connecting to the database
$ kubectl exec -it -n demo sdb-sample-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-sample-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 144
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show databases;
+--------------------+
| Database           |
+--------------------+
| cluster            |
| information_schema |
| kubedb_test        |
| memsql             |
| singlestore_health |
+--------------------+
5 rows in set (0.00 sec)

singlestore> use kubedb_test;
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Database changed

# Showing the inserted `kubedb_write_check`
singlestore> select * from kubedb_write_check;
+----+-------+
| id | name  |
+----+-------+
|  3 | name3 |
|  1 | name1 |
|  2 | name2 |
+----+-------+
3 rows in set (0.02 sec)

singlestore> exit
Bye


```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete sdb -n demo sdb-sample
singlestore.kubedb.com "sdb-sample" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
