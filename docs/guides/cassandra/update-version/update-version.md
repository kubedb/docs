---
title: Update Version of Cassandra
menu:
  docs_{{ .version }}:
    identifier: cas-update-version-cassandra
    name: Cassandra Update Version
    parent: cas-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Cassandra

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Cassandra` Topology.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Cassandra](/docs/guides/cassandra/concepts/cassandra.md)
    - [CassandraOpsRequest](/docs/guides/cassandra/concepts/cassandraopsrequest.md)
    - [Updating Overview](/docs/guides/cassandra/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/cassandra](/docs/examples/cassandra) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare Cassandra

Now, we are going to deploy a `Cassandra` replicaset database with version `4.1.8`.

### Deploy Cassandra

In this section, we are going to deploy a Cassandra topology cluster. Then, in the next section we will update the version using `CassandraOpsRequest` CRD. Below is the YAML of the `Cassandra` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-prod
  namespace: demo
spec:
  version: 4.1.8
  topology:
    rack:
      - name: r0
        replicas: 2
        podTemplate:
          spec:
            containers:
              - name: cassandra
                resources:
                  limits:
                    memory: 2Gi
                    cpu: 2
                  requests:
                    memory: 1Gi
                    cpu: 1
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Cassandra` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/update-version/cassandra.yaml
cassandra.kubedb.com/cassandra-prod created
```

Now, wait until `cassandra-prod` created has status `Ready`. i.e,

```bash
$  kubectl get cas -n demo -w
NAME             TYPE                  VERSION   STATUS         AGE
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   45s
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Provisioning   82s
.
.
cassandra-prod   kubedb.com/v1alpha2   5.0.3     Ready          106s

```

We are now ready to apply the `CassandraOpsRequest` CR to update.

### update Cassandra Version

Here, we are going to update `Cassandra` from `4.1.8` to `5.0.3`.

#### Create CassandraOpsRequest:

In order to update the version, we have to create a `CassandraOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `CassandraOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: CassandraOpsRequest
metadata:
  name: cassandra-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: cass
  updateVersion:
    targetVersion: 5.0.3
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `cassandra-prod` Cassandra.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `5.0.3`.

> **Note:** If you want to update combined Cassandra, you just refer to the `Cassandra` combined object name in `spec.databaseRef.name`. To create a combined Cassandra, you can refer to the [Cassandra Combined](/docs/guides/cassandra/clustering/combined-cluster/index.md) guide.

Let's create the `CassandraOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/update-version/update-version.yaml
cassandraopsrequest.ops.kubedb.com/cassandra-update-version created
```

#### Verify Cassandra version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Cassandra` object and related `PetSets` and `Pods`.

Let's wait for `CassandraOpsRequest` to be `Successful`.  Run the following command to watch `CassandraOpsRequest` CR,

```bash
$ kubectl get cassandraopsrequest -n demo
NAME                       TYPE            STATUS        AGE
cassandra-update-version   UpdateVersion   Successful    2m6s
```

We can see from the above output that the `CassandraOpsRequest` has succeeded. If we describe the `CassandraOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl describe cassandraopsrequest -n demo cassandra-update-version
Name:         cassandra-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         CassandraOpsRequest
Metadata:
  Creation Timestamp:  2025-07-22T06:15:30Z
  Generation:          1
  Resource Version:    124398
  UID:                 03d0f9ef-fbbc-48bb-b1cb-9c38d5b127ce
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   cassandra-prod
  Timeout:  5m
  Type:     UpdateVersion
  Update Version:
    Target Version:  5.0.3
Status:
  Conditions:
    Last Transition Time:  2025-07-22T06:15:30Z
    Message:               Cassandra ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2025-07-22T06:15:38Z
    Message:               successfully reconciled the Cassandra with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2025-07-22T06:18:23Z
    Message:               Successfully Restarted Cassandra nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-07-22T06:15:43Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-22T06:15:43Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-0
    Last Transition Time:  2025-07-22T06:15:48Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-07-22T06:16:23Z
    Message:               get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-22T06:16:23Z
    Message:               evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--cassandra-prod-rack-r0-1
    Last Transition Time:  2025-07-22T06:18:23Z
    Message:               Successfully completed update cassandra version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                             Age    From                         Message
  ----     ------                                                             ----   ----                         -------
  Normal   Starting                                                           6m51s  KubeDB Ops-manager Operator  Start processing for CassandraOpsRequest: demo/cassandra-update-version
  Normal   Starting                                                           6m51s  KubeDB Ops-manager Operator  Pausing Cassandra databse: demo/cassandra-prod
  Normal   Successful                                                         6m51s  KubeDB Ops-manager Operator  Successfully paused Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-update-version
  Normal   UpdatePetSets                                                      6m43s  KubeDB Ops-manager Operator  successfully reconciled the Cassandra with updated version
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    6m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  6m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  running pod; ConditionStatus:False                                 6m33s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    5m58s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  5m58s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0    5m18s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0  5m18s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-0
  Warning  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1    4m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Warning  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1  4m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:cassandra-prod-rack-r0-1
  Normal   RestartPods                                                        3m58s  KubeDB Ops-manager Operator  Successfully Restarted Cassandra nodes
  Normal   Starting                                                           3m58s  KubeDB Ops-manager Operator  Resuming Cassandra database: demo/cassandra-prod
  Normal   Successful                                                         3m58s  KubeDB Ops-manager Operator  Successfully resumed Cassandra database: demo/cassandra-prod for CassandraOpsRequest: cassandra-update-version
```

Now, we are going to verify whether the `Cassandra` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get cas -n demo cassandra-prod -o=jsonpath='{.spec.version}{"\n"}'
5.0.3

$ kubectl get petset -n demo cassandra-prod-broker -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/cassandra-kraft:3.9.0@sha256:e251d3c0ceee0db8400b689e42587985034852a8a6c81b5973c2844e902e6d11

$ kubectl get petset -n demo cassandra-prod-rack-r0 -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/cassandra-management:5.0.3@sha256:ef296c7ce02b438f3af43bd07457ca44881c845c6eeef631989b4ed7351b7243

```

You can see from above, our `Cassandra` has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete cassandraopsrequest -n demo cassandra-update-version
kubectl delete cas -n demo cassandra-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Different Cassandra topology clustering modes [here](/docs/guides/cassandra/clustering/_index.md).
- Monitor your Cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).

[//]: # (- Monitor your Cassandra database with KubeDB using [out-of-the-box builtin-Prometheus]&#40;/docs/guides/cassandra/monitoring/using-builtin-prometheus.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
