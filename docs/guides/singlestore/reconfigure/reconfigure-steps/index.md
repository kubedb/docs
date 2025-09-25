---
title: Reconfigure SingleStore Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-reconfigure-reconfigure-steps
    name: Reconfigure OpsRequest
    parent: guides-sdb-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure SingleStore Cluster Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a SingleStore Cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
- [SingleStore](/docs/guides/singlestore/concepts/singlestore.md)
- [SingleStore Cluster](/docs/guides/singlestore/clustering)
- [SingleStoreOpsRequest](/docs/guides/singlestore/concepts/opsrequest.md)
- [Reconfigure Overview](/docs/guides/singlestore/reconfigure/overview)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

Now, we are going to deploy a  `SingleStore` Cluster using a supported version by `KubeDB` operator. Then we are going to apply `SingleStoreOpsRequest` to reconfigure its configuration.

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Deploy SingleStore

At first, we will create `sdb-config.cnf` file containing required configuration settings.

```ini
$ cat sdb-config.cnf 
[server]
max_connections = 250
read_buffer_size = 122880

```

Here, `max_connections` is set to `250`, whereas the default value is `100000`. Likewise, `read_buffer_size` has the deafult value `131072`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo sdb-configuration --from-file=./sdb-config.cnf
secret/sdb-configuration created
```

In this section, we are going to create a SingleStore object specifying `spec.topology.aggreagtor.configSecret` field to apply this custom configuration. Below is the YAML of the `SingleStore` CR that we are going to create,


```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: custom-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      configSecret:
        name: sdb-configuration
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
      configSecret:
        name: sdb-configuration
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

Let's create the `SingleStore` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure/reconfigure-steps/yamls/custom-sdb.yaml
singlestore.kubedb.com/custom-sdb created


Now, wait until `custom-sdb` has status `Ready`. i.e,

```bash
$ kubectl get pod -n demo
NAME                      READY   STATUS    RESTARTS   AGE
custom-sdb-aggregator-0   2/2     Running   0          94s
custom-sdb-aggregator-1   2/2     Running   0          88s
custom-sdb-leaf-0         2/2     Running   0          91s
custom-sdb-leaf-1         2/2     Running   0          86s

$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
custom-sdb   kubedb.com/v1alpha2   8.7.10    Ready    4m29s
```

We can see the database is in ready phase so it can accept conncetion.

Now, we will check if the database has started with the custom configuration we have provided.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# Connceting to the database
$ kubectl exec -it -n demo custom-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@custom-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 208
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
singlestore> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 250   |
+-----------------+-------+
1 row in set (0.00 sec)

# value of `read_buffer_size` is same as provided
singlestore> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 122880 |
+------------------+--------+
1 row in set (0.00 sec)

singlestore> exit
Bye

```

As we can see from the configuration of ready singlestore, the value of `max_connections` has been set to `250` and `read_buffer_size` has been set to `122880`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `max_connections` to `350` and `read_buffer_size` to `132880`.

Now, we will create new file `new-sdb-config.cnf` containing required configuration settings.

#### Create SingleStoreOpsRequest

Now, we will use this secret to replace the previous secret using a `SingleStoreOpsRequest` CR. The `SingleStoreOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-config
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: custom-sdb
  configuration:
    aggregator:
      applyConfig:
        sdb-apply.cnf: |-
          max_connections = 550
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `custom-sdb` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.aggregator.applyConfig` is a map where key supports 1 values, namely `sdb-apply.cnf` for aggregator nodes. You can also specifies `spec.configuration.leaf.applyConfig` which is a map where key supports 1 values, namely `sdb-apply.cnf` for leaf nodes.

Let's create the `SinglestoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure-steps/yamls/reconfigure-using-applyConfig.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-reconfigure-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `SingleStore` object.

Let's wait for `SinglestoreOpsRequest` to be `Successful`.  Run the following command to watch `SinglestoreOpsRequest` CR,

```bash
$ kubectl get singlestoreopsrequest --all-namespaces
NAMESPACE   NAME                        TYPE            STATUS       AGE
demo        sdbops-reconfigure-config   Reconfigure     Successful   10m
```

We can see from the above output that the `SinglestoreOpsRequest` has succeeded. If we describe the `SinglestoreOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe singlestoreopsrequest -n demo sdbops-reconfigure-config
Name:         sdbops-reconfigure-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SinglestoreOpsRequest
Metadata:
  Creation Timestamp:  2024-10-04T10:18:22Z
  Generation:          1
  Resource Version:    2114236
  UID:                 56b37f6d-d8be-49c7-a588-9740863edd2a
Spec:
  Apply:  IfReady
  Configuration:
    Aggregator:
      Apply Config:
        sdb-apply.cnf:  max_connections = 550
  Database Ref:
    Name:  custom-sdb
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-10-04T10:18:22Z
    Message:               Singlestore ops-request has started to expand volume of singlestore nodes.
    Observed Generation:   1
    Reason:                Configuration
    Status:                True
    Type:                  Configuration
    Last Transition Time:  2024-10-04T10:18:28Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-04T10:18:28Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-04T10:19:53Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-04T10:18:33Z
    Message:               get pod; ConditionStatus:True; PodName:custom-sdb-aggregator-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--custom-sdb-aggregator-0
    Last Transition Time:  2024-10-04T10:18:33Z
    Message:               evict pod; ConditionStatus:True; PodName:custom-sdb-aggregator-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--custom-sdb-aggregator-0
    Last Transition Time:  2024-10-04T10:19:08Z
    Message:               check pod ready; ConditionStatus:True; PodName:custom-sdb-aggregator-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--custom-sdb-aggregator-0
    Last Transition Time:  2024-10-04T10:19:13Z
    Message:               get pod; ConditionStatus:True; PodName:custom-sdb-aggregator-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--custom-sdb-aggregator-1
    Last Transition Time:  2024-10-04T10:19:13Z
    Message:               evict pod; ConditionStatus:True; PodName:custom-sdb-aggregator-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--custom-sdb-aggregator-1
    Last Transition Time:  2024-10-04T10:19:48Z
    Message:               check pod ready; ConditionStatus:True; PodName:custom-sdb-aggregator-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--custom-sdb-aggregator-1
    Last Transition Time:  2024-10-04T10:19:53Z
    Message:               Successfully completed the reconfiguring for Singlestore
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>

```

Now let's connect to a singlestore instance and run a memsql internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo custom-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@custom-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 626
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 550   |
+-----------------+-------+
1 row in set (0.00 sec)

singlestore> exit
Bye


```

As we can see from the configuration has changed, the value of `max_connections` has been changed from `250` to `550`. So the reconfiguration of the database is successful.

### Remove Custom Configuration

We can also remove exisiting custom config using `SinglestoreOpsRequest`. Provide `true` to field `spec.configuration.aggregator.removeCustomConfig` and make an Ops Request to remove existing custom configuration.

#### Create SingleStoreOpsRequest

Lets create an `SinglestoreOpsRequest` having `spec.configuration.aggregator.removeCustomConfig` is equal `true`,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SinglestoreOpsRequest
metadata:
  name: sdbops-reconfigure-remove
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: custom-sdb
  configuration:
    aggregator:  
      removeCustomConfig: true
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `custom-sdb` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.aggregator.removeCustomConfig` is a bool field that should be `true` when you want to remove existing custom configuration.

Let's create the `SingleStoreOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/reconfigure/yamls/reconfigure-steps/reconfigure-remove.yaml
singlestoreopsrequest.ops.kubedb.com/sdbops-reconfigure-remove created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `SingleStore` object.

Let's wait for `SingleStoreOpsRequest` to be `Successful`.  Run the following command to watch `SingleStoreOpsRequest` CR,

```bash
$ kubectl get singlestoreopsrequest -n demo
NAME                        TYPE             STATUS       AGE
sdbops-reconfigure-remove   Reconfigure      Successful  5m31s
```

Now let's connect to a singlestore instance and run a singlestore internal command to check the new configuration we have provided.

```bash
$ kubectl exec -it -n demo custom-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@custom-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 166
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> 
singlestore> 
singlestore> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 100000|
+-----------------+-------+
1 row in set (0.00 sec)

singlestore> exit
Bye


```

As we can see from the configuration has changed to its default value. So removal of existing custom configuration using `SingleStoreOpsRequest` is successful.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete singlestore -n demo custom-sdb
$ kubectl delete singlestoreopsrequest -n demo sdbops-reconfigure-config  sdbops-reconfigure-remove
$ kubectl delete ns demo
```
