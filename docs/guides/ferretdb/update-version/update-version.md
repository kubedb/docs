---
title: Updating FerretDB
menu:
  docs_{{ .version }}:
    identifier: fr-updating-ferretdb
    name: Update version
    parent: fr-updating
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# update version of FerretDB

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `FerretDB`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [FerretDB](/docs/guides/ferretdb/concepts/ferretdb.md)
    - [FerretDBOpsRequest](/docs/guides/ferretdb/concepts/opsrequest.md)
    - [Updating Overview](/docs/guides/ferretdb/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/ferretdb](/docs/examples/ferretdb) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare FerretDB

Now, we are going to deploy a `FerretDB` =with version `1.18.0`.

### Deploy FerretDB:

In this section, we are going to deploy a FerretDB. Then, in the next section we will update the version  using `FerretDBOpsRequest` CRD. Below is the YAML of the `FerretDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: fr-update
  namespace: demo
spec:
  version: "1.18.0"
  backend:
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/update-version/fr-update.yaml
ferretdb.kubedb.com/fr-update created
```

Now, wait until `fr-update` created has status `Ready`. i.e,

```bash
$ kubectl get fr -n demo
 NAME        TYPE                  VERSION   STATUS   AGE
 fr-update   kubedb.com/v1alpha2   1.18.0    Ready    26s
```

We are now ready to apply the `FerretDBOpsRequest` CR to update this FerretDB.

### update FerretDB Version

Here, we are going to update `FerretDB` from `1.18.0` to `1.23.0`.

#### Create FerretDBOpsRequest:

In order to update the FerretDB, we have to create a `FerretDBOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `FerretDBOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: FerretDBOpsRequest
metadata:
  name: ferretdb-version-update
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: fr-update
  updateVersion:
    targetVersion: 1.23.0
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `fr-update` FerretDB.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our FerretDB.
- `spec.updateVersion.targetVersion` specifies the expected version of the FerretDB `1.23.0`.


Let's create the `FerretDBOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/update-version/frops-update.yaml
ferretdbopsrequest.ops.kubedb.com/ferretdb-version-update created
```

#### Verify FerretDB version updated successfully :

If everything goes well, `KubeDB` Ops-manager operator will update the image of `FerretDB` object and related `PetSets` and `Pods`.

Let's wait for `FerretDBOpsRequest` to be `Successful`.  Run the following command to watch `FerretDBOpsRequest` CR,

```bash
$ watch kubectl get ferretdbopsrequest -n demo
Every 2.0s: kubectl get ferretdbopsrequest -n demo
NAME                      TYPE                STATUS       AGE
ferretdb-version-update   UpdateVersion       Successful   93s
```

We can see from the above output that the `FerretDBOpsRequest` has succeeded. If we describe the `FerretDBOpsRequest` we will get an overview of the steps that were followed to update the FerretDB.

```bash
$ kubectl describe ferretdbopsrequest -n demo ferretdb-version-update
Name:         ferretdb-version-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         FerretDBOpsRequest
Metadata:
  Creation Timestamp:  2024-10-21T05:06:17Z
  Generation:          1
  Resource Version:    324860
  UID:                 30d486a6-a8fe-4d82-a8b3-f13e299ef035
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  fr-update
  Type:    UpdateVersion
  Update Version:
    Target Version:  1.23.0
Status:
  Conditions:
    Last Transition Time:  2024-10-21T05:06:17Z
    Message:               FerretDB ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2024-10-21T05:06:25Z
    Message:               successfully reconciled the FerretDB with updated version
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-10-21T05:06:30Z
    Message:               get pod; ConditionStatus:True; PodName:fr-update-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--fr-update-0
    Last Transition Time:  2024-10-21T05:06:30Z
    Message:               evict pod; ConditionStatus:True; PodName:fr-update-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--fr-update-0
    Last Transition Time:  2024-10-21T05:06:35Z
    Message:               check pod running; ConditionStatus:True; PodName:fr-update-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodRunning--fr-update-0
    Last Transition Time:  2024-10-21T05:06:40Z
    Message:               Successfully Restarted FerretDB pods
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-10-21T05:06:40Z
    Message:               Successfully updated FerretDB
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2024-10-21T05:06:40Z
    Message:               Successfully updated FerretDB version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                        Age   From                         Message
  ----     ------                                                        ----  ----                         -------
  Normal   Starting                                                      59s   KubeDB Ops-manager Operator  Start processing for FerretDBOpsRequest: demo/ferretdb-version-update
  Normal   Starting                                                      59s   KubeDB Ops-manager Operator  Pausing FerretDB database: demo/fr-update
  Normal   Successful                                                    59s   KubeDB Ops-manager Operator  Successfully paused FerretDB database: demo/fr-update for FerretDBOpsRequest: ferretdb-version-update
  Normal   UpdatePetSets                                                 51s   KubeDB Ops-manager Operator  successfully reconciled the FerretDB with updated version
  Warning  get pod; ConditionStatus:True; PodName:fr-update-0            46s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:fr-update-0
  Warning  evict pod; ConditionStatus:True; PodName:fr-update-0          46s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:fr-update-0
  Warning  check pod running; ConditionStatus:True; PodName:fr-update-0  41s   KubeDB Ops-manager Operator  check pod running; ConditionStatus:True; PodName:fr-update-0
  Normal   RestartPods                                                   36s   KubeDB Ops-manager Operator  Successfully Restarted FerretDB pods
  Normal   Starting                                                      36s   KubeDB Ops-manager Operator  Resuming FerretDB database: demo/fr-update
  Normal   Successful                                                    36s   KubeDB Ops-manager Operator  Successfully resumed FerretDB database: demo/fr-update for FerretDBOpsRequest: ferretdb-version-update
```

Now, we are going to verify whether the `FerretDB` and the related `PetSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get fr -n demo fr-update -o=jsonpath='{.spec.version}{"\n"}'                                                                                          
1.23.0

$ kubectl get petset -n demo fr-update -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'                                                               
ghcr.io/appscode-images/ferretdb:1.23.0

$ kubectl get pods -n demo fr-update-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'                                                                           
ghcr.io/appscode-images/ferretdb:1.23.0
```

You can see from above, our `FerretDB` has been updated with the new version. So, the update process is successfully completed.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete fr -n demo fr-update
kubectl delete ferretdbopsrequest -n demo ferretdb-version-update
```