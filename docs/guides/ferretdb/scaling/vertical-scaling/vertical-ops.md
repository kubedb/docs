---
title: Vertical Scaling FerretDB
menu:
  docs_{{ .version }}:
    identifier: fr-vertical-scaling-ops
    name: VerticalScaling OpsRequest
    parent: fr-vertical-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale FerretDB

This guide will show you how to use `KubeDB` Ops-manager operator to update the resources of a FerretDB.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md)
    - [FerretDBOpsRequest](/docs/guides/ferretdb/concepts/opsrequest.md)
    - [Vertical Scaling Overview](/docs/guides/ferretdb/scaling/vertical-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ferretdb](/docs/examples/ferretdb) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Apply Vertical Scaling on FerretDB

Here, we are going to deploy a  `FerretDB` using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

### Prepare FerretDB

Now, we are going to deploy a `FerretDB` with version `1.23.0`.

### Deploy FerretDB

In this section, we are going to deploy a FerretDB. Then, in the next section we will update the resources using `FerretDBOpsRequest` CRD. Below is the YAML of the `FerretDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: fr-vertical
  namespace: demo
spec:
  version: "1.23.0"
  replicas: 1
  backend:
    externallyManaged: false
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 500Mi
  deletionPolicy: WipeOut
```

Let's create the `FerretDB` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/scaling/fr-vertical.yaml
ferretdb.kubedb.com/fr-vertical created
```

Now, wait until `fr-vertical` has status `Ready`. i.e,

```bash
$ kubectl get fr -n demo
NAME          TYPE                  VERSION   STATUS   AGE
fr-vertical   kubedb.com/v1alpha2   1.23.0    Ready    17s
```

Let's check the Pod containers resources,

```bash
$ kubectl get pod -n demo fr-vertical-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "memory": "1Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see the Pod has default resources which is assigned by the KubeDB operator.

We are now ready to apply the `FerretDBOpsRequest` CR to update the resources of this ferretdb.

### Vertical Scaling

Here, we are going to update the resources of the ferretdb to meet the desired resources after scaling.

#### Create FerretDBOpsRequest

In order to update the resources of the ferretdb, we have to create a `FerretDBOpsRequest` CR with our desired resources. Below is the YAML of the `FerretDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-scale-vertical
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: fr-vertical
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `fr-vertical` ferretdb.
- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.VerticalScaling.standalone` specifies the desired resources after scaling.
- Have a look [here](/docs/guides/ferretdb/concepts/opsrequest.md) on the respective sections to understand the `timeout` & `apply` fields.

Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/scaling/vertical-scaling/fr-vertical-ops.yaml
ferretdbopsrequest.ops.kubedb.com/ferretdb-scale-vertical created
```

#### Verify FerretDB resources updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the resources of `FerretDB` object and related `PetSet` and `Pods`.

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CR,

```bash
$ kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                      TYPE              STATUS       AGE
ferretdb-scale-vertical   VerticalScaling   Successful   44s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed to scale the ferretdb.

```bash
$ kubectl describe ferretdbopsrequest -n demo ferretdb-scale-vertical
Name:         ferretdb-scale-vertical
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-21T12:25:33Z
  Generation:          1
  Resource Version:    366310
  UID:                 38631646-684f-4c2a-8496-c7b085743243
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   fr-vertical
  Timeout:  5m
  Type:     VerticalScaling
  Vertical Scaling:
    Node:
      Resources:
        Limits:
          Cpu:     1
          Memory:  2Gi
        Requests:
          Cpu:     1
          Memory:  2Gi
Status:
  Conditions:
    Last Transition Time:  2024-10-21T12:25:33Z
    Message:               FerretDB ops-request has started to vertically scaling the FerretDB nodes
    Observed Generation:   1
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2024-10-21T12:25:36Z
    Message:               Successfully paused database
    Observed Generation:   1
    Reason:                DatabasePauseSucceeded
    Status:                True
    Type:                  DatabasePauseSucceeded
    Last Transition Time:  2024-10-21T12:25:37Z
    Message:               Successfully updated PetSets Resources
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-21T12:25:42Z
    Message:               get pod; ConditionStatus:True; PodName:fr-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--fr-vertical-0
    Last Transition Time:  2024-10-21T12:25:42Z
    Message:               evict pod; ConditionStatus:True; PodName:fr-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--fr-vertical-0
    Last Transition Time:  2024-10-21T12:25:47Z
    Message:               check pod running; ConditionStatus:True; PodName:fr-vertical-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--fr-vertical-0
    Last Transition Time:  2024-10-21T12:25:52Z
    Message:               Successfully Restarted Pods With Resources
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-21T12:25:52Z
    Message:               Successfully completed the VerticalScaling for FerretDB
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   Starting                                                        58s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/ferretdb-scale-vertical
  Normal   Starting                                                        58s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/fr-vertical
  Normal   Successful                                                      58s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/fr-vertical for FerretDBOpsRequest: ferretdb-scale-vertical
  Normal   UpdatePetSets                                                   54s   KubeDB Ops-manager Operator  Successfully updated PetSets Resources
  Warning  get pod; ConditionStatus:True; PodName:fr-vertical-0            49s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:fr-vertical-0
  Warning  evict pod; ConditionStatus:True; PodName:fr-vertical-0          49s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:fr-vertical-0
  Warning  check pod running; ConditionStatus:True; PodName:fr-vertical-0  44s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:fr-vertical-0
  Normal   RestartPods                                                     39s   KubeDB Ops-manager Operator  Successfully Restarted Pods With Resources
  Normal   Starting                                                        39s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/fr-vertical
  Normal   Successful                                                      39s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/fr-vertical for FerretDBOpsRequest: ferretdb-scale-vertical
```

Now, we are going to verify from the Pod yaml whether the resources of the ferretdb has updated to meet up the desired state, Let's check,

```bash
$ kubectl get pod -n demo fr-vertical-0 -o json | jq '.spec.containers[].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "1",
    "memory": "2Gi"
  }
}
```

The above output verifies that we have successfully scaled up the resources of the FerretDB.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete fr -n demo fr-vertical
kubectl delete ferretdbopsrequest -n demo ferretdb-scale-vertical
```