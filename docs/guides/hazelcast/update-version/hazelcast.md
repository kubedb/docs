---
title: Updating Hazelcast Database
menu:
  docs_{{ .version }}:
    identifier: hz-update-version
    name: Update Version
    parent: update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Hazelcast

This guide will show you how to use `KubeDB` Enterprise operator to update the version of `Hazelcast`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)
  - [updating Overview](/docs/guides/hazelcast/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/hazelcast](/docs/examples/hazelcast) directory of [kubedb/docs](https://github.com/kube/docs) repository.

### Prepare Hazelcast Database

Now, we are going to deploy a `Hazelcast` database with version `5.2.2`.

### Deploy Hazelcast:

Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.


```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey=TrialLicense#10Nodes#eyJhbGxvd2VkTmF0aXZlTWVtb3J5U2l6ZSI6MTAwLCJhbGxvd2VkTnVtYmVyT2ZOb2RlcyI6MTAsImFsbG93ZWRUaWVyZWRTdG9yZVNpemUiOjAsImFsbG93ZWRUcGNDb3JlcyI6MCwiY3JlYXRpb25EYXRlIjoxNzQ4ODQwNDc3LjYzOTQ0NzgxNiwiZXhwaXJ5RGF0ZSI6MTc1MTQxNDM5OS45OTk5OTk5OTksImZlYXR1cmVzIjpbMCwyLDMsNCw1LDYsNyw4LDEwLDExLDEzLDE0LDE1LDE3LDIxLDIyXSwiZ3JhY2VQZXJpb2QiOjAsImhhemVsY2FzdFZlcnNpb24iOjk5LCJvZW0iOmZhbHNlLCJ0cmlhbCI6dHJ1ZSwidmVyc2lvbiI6IlY3In0=.6PYD6i-hejrJ5Czgc3nYsmnwF7mAI-78E8LFEuYp-lnzXh_QLvvsYx4ECD0EimqcdeG2J5sqUI06okLD502mCA==
secret/hz-license-key created
```

In this section, we are going to deploy a Hazelcast database. Then, in the next section we will update the version of the database using `HazelcastOpsRequest` CRD. Below is the YAML of the `Hazelcast` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hz-prod
  namespace: demo
spec:
  deletionPolicy: WipeOut
  licenseSecret:
    name: hz-license-key
  replicas: 3
  version: 5.5.2
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
```

Let's create the `Hazelcast` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/update-version/hazelcast.yaml
hazelcast.kubedb.com/hz-prod created
```

Now, wait until `hz-prod` has status `Ready`. i.e,

```bash
$ kubectl get hazelcast -n demo
NAME      TYPE                  VERSION   STATUS   AGE
hz-prod   kubedb.com/v1alpha2   5.5.2     Ready    3h4m
```

We are now ready to apply the `HazelcastOpsRequest` CR to update this database.

### Update Hazelcast Version

Here, we are going to update `Hazelcast` from `5.2.2` to `5.5.6`.

#### Create HazelcastOpsRequest:

In order to update the database, we have to create a `HazelcastOpsRequest` CR with our desired version that is supported by `KubeDB`. Below is the YAML of the `HazelcastOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hzops-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef:
    name: hz-prod
  updateVersion:
    targetVersion: 5.5.6
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `hz-prod` Hazelcast database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `5.5.6`.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/update-version/update-version.yaml
hazelcastopsrequest.ops.kubedb.com/hzops-update-version created
```

#### Verify Hazelcast version updated successfully:

If everything goes well, `KubeDB` Enterprise operator will update the image of `Hazelcast` object and related `PetSets` and `Pods`.

Let's wait for `HazelcastOpsRequest` to be `Successful`. Run the following command to watch `HazelcastOpsRequest` CR,

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME                   TYPE            STATUS       AGE
hzops-update-version   UpdateVersion   Successful   3m2s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed to update the database.

```bash
$ kubectl describe hazelcastopsrequest -n demo hzops-update-version
Name:         hzops-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T08:27:38Z
  Generation:          1
  Resource Version:    5455422
  UID:                 ecb686fb-895f-4fb8-b182-22ebc4d77a3a
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hz-prod
  Type:    UpdateVersion
  Update Version:
    Target Version:  5.5.6
Status:
  Conditions:
    Last Transition Time:  2025-08-19T08:27:38Z
    Message:               Hazelcast ops-request has started to update version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2025-08-19T08:27:51Z
    Message:               successfully reconciled the Hazelcast with updated version
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-19T08:30:41Z
    Message:               Successfully Restarted Hazelcast nodes
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2025-08-19T08:28:01Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-0
    Last Transition Time:  2025-08-19T08:28:01Z
    Message:               evict pod; ConditionStatus:True; PodName:hz-prod-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hz-prod-0
    Last Transition Time:  2025-08-19T08:28:11Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2025-08-19T08:28:51Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-1
    Last Transition Time:  2025-08-19T08:28:51Z
    Message:               evict pod; ConditionStatus:True; PodName:hz-prod-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hz-prod-1
    Last Transition Time:  2025-08-19T08:29:41Z
    Message:               get pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--hz-prod-2
    Last Transition Time:  2025-08-19T08:29:41Z
    Message:               evict pod; ConditionStatus:True; PodName:hz-prod-2
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--hz-prod-2
    Last Transition Time:  2025-08-19T08:30:41Z
    Message:               Successfully updated hazelcast version
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                              Age    From                         Message
  ----     ------                                              ----   ----                         -------
  Normal   Starting                                            3m18s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-update-version
  Normal   Starting                                            3m18s  KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hz-prod
  Normal   Successful                                          3m18s  KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-update-version
  Normal   UpdateStatefulSets                                  3m5s   KubeDB Ops-manager Operator  successfully reconciled the Hazelcast with updated version
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-0    2m55s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  evict pod; ConditionStatus:True; PodName:hz-prod-0  2m55s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hz-prod-0
  Warning  running pod; ConditionStatus:False                  2m45s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-1    2m5s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  evict pod; ConditionStatus:True; PodName:hz-prod-1  2m5s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hz-prod-1
  Warning  get pod; ConditionStatus:True; PodName:hz-prod-2    75s    KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:hz-prod-2
  Warning  evict pod; ConditionStatus:True; PodName:hz-prod-2  75s    KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:hz-prod-2
  Normal   RestartPods                                         15s    KubeDB Ops-manager Operator  Successfully Restarted Hazelcast nodes
  Normal   Starting                                            15s    KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hz-prod
  Normal   Successful                                          15s    KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hzops-update-version
```

Now, we are going to verify whether the `Hazelcast` and the related `StatefulSets` their `Pods` have the new version image. Let's check,

```bash
$ kubectl get hazelcast -n demo hz-prod -o=jsonpath='{.spec.version}{"\n"}'
5.5.6
```

```bash
$ kubectl get statefulset -n demo hz-prod -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/hazelcast:5.5.6@sha256:abc123def456...
```

```bash
$ kubectl get pods -n demo hz-prod-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/hazelcast:5.5.6@sha256:abc123def456...
```

Let's also verify that the database is ready to take connections:

```bash
$ kubectl get hazelcast -n demo hz-quickstart
NAME      TYPE                  VERSION   STATUS   AGE
hz-prod   kubedb.com/v1alpha2   5.5.6     Ready    3h14m
```

You can see from the above outputs that the `Hazelcast` object and related resources have been updated with the new version `5.5.6`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequest -n demo hzops-update-version
kubectl delete hazelcast -n demo hz-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Monitor your Hazelcast database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
