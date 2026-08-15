---
title: Restart Etcd
menu:
  docs_{{ .version }}:
    identifier: etcd-restart-details
    name: Restart Cluster
    parent: etcd-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Restart Etcd

KubeDB supports restarting an Etcd cluster through an `EtcdOpsRequest`. Restarting is useful if
some member pods are stuck, if a node has to be drained, or if you simply want the members to come
up fresh. This tutorial shows you how to use it.

Unlike a plain `kubectl delete pod`, this restart is **Raft aware**. Evicting the current Raft
leader forces an election, and an election makes the whole cluster unavailable for writes until a
new leader is elected. So the operator:

1. Resolves which pod currently holds the Raft leadership.
2. Restarts every **follower** first, one at a time, waiting for each pod to become `Ready` **and**
   for the cluster to regain quorum before moving to the next one.
3. Restarts the **leader last**, and only after transferring its leadership to another healthy
   voting member with etcd's `MoveLeader` RPC — a transfer is near instantaneous, an election is
   not.

A single member cluster (`spec.replicas: 1`) has nowhere to hand the leadership to, so it is simply
restarted like any other pod.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure the KubeDB Provisioner and Ops-manager operators are installed with `featureGates.Etcd=true`.

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Etcd

In this section, we are going to deploy a 3 member Etcd cluster using KubeDB.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-quickstart
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Etcd` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/restart/etcd.yaml
etcd.kubedb.com/etcd-quickstart created
```

Now, wait until `etcd-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo
NAME              VERSION   STATUS   AGE
etcd-quickstart   3.6.4     Ready    3m

$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-quickstart
NAME                READY   STATUS    RESTARTS   AGE
etcd-quickstart-0   1/1     Running   0          3m
etcd-quickstart-1   1/1     Running   0          2m
etcd-quickstart-2   1/1     Running   0          2m
```

## Apply Restart opsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: etcd-quickstart
  timeout: 5m
  apply: Always
```

- `spec.type` specifies the type of the ops request. Here it is `Restart`.
- `spec.databaseRef` holds the name of the `Etcd` object. The database must be in the same namespace as the ops request.
- `spec.timeout` bounds how long a single step (evicting a pod, waiting for it to be ready, waiting for quorum) may keep retrying before the request is failed.
- The meaning of `spec.timeout` & `spec.apply` fields is described [here](/docs/guides/etcd/concepts/etcdopsrequest.md#spectimeout).

> Note: Restarting a single member cluster and a multi member cluster is done exactly the same way. All you need is to point `spec.databaseRef.name` at the corresponding `Etcd` object.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/restart/ops.yaml
etcdopsrequest.ops.kubedb.com/etcd-restart created
```

## Track the progress

Now the Ops-manager operator will pause the `Etcd` object, restart the member pods one at a time —
followers first, leader last — and finally resume the object.

```bash
$ kubectl get etcdopsrequest -n demo
NAME           TYPE      STATUS       AGE
etcd-restart   Restart   Progressing  35s
```

While it runs, you can watch the pods being recreated one at a time. The cluster keeps quorum
throughout, because the operator never evicts the next pod until the previous one is back in the
Raft quorum:

```bash
$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-quickstart -w
NAME                READY   STATUS        RESTARTS   AGE
etcd-quickstart-0   1/1     Terminating   0          8m
etcd-quickstart-1   1/1     Running       0          8m
etcd-quickstart-2   1/1     Running       0          7m
```

`kubectl describe` gives you the step-by-step view, including the events the operator records for
each retryable step:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-restart
Name:         etcd-restart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-02-10T09:22:57Z
  Generation:          1
  Resource Version:    1072309
  UID:                 6091d9fa-1c2b-4734-bdd1-1ace91460bea
Spec:
  Apply:  Always
  Database Ref:
    Name:   etcd-quickstart
  Timeout:  5m
  Type:     Restart
Status:
  Conditions:
    Last Transition Time:  2026-02-10T09:22:57Z
    Message:               Restart is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-02-10T09:23:10Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-quickstart-0
  Observed Generation:     1
  Phase:                   Progressing
Events:
  Type     Reason                                                             Age   From                         Message
  ----     ------                                                             ----  ----                         -------
  Normal   PauseDatabase                                                      35s   KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-quickstart
  Warning  check pod ready; ConditionStatus:False; PodName:etcd-quickstart-0  30s   KubeDB Ops-manager Operator  check pod ready; ConditionStatus:False; PodName:etcd-quickstart-0
  Warning  check pod ready; ConditionStatus:True; PodName:etcd-quickstart-0   5s    KubeDB Ops-manager Operator  check pod ready; ConditionStatus:True; PodName:etcd-quickstart-0
```

Once every member has been recreated, the request reaches `Successful`:

```bash
$ kubectl get etcdopsrequest -n demo
NAME           TYPE      STATUS       AGE
etcd-restart   Restart   Successful   4m
```

Here is the full object after a successful run. In this example the Raft leader happened to be
`etcd-quickstart-1`, so that pod was restarted **last**, after its leadership had been moved away:

```bash
$ kubectl get etcdopsrequest -n demo etcd-restart -o yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"EtcdOpsRequest","metadata":{"annotations":{},"name":"etcd-restart","namespace":"demo"},"spec":{"apply":"Always","databaseRef":{"name":"etcd-quickstart"},"timeout":"5m","type":"Restart"}}
  creationTimestamp: "2026-02-10T09:22:57Z"
  generation: 1
  name: etcd-restart
  namespace: demo
  resourceVersion: "1072471"
  uid: 6091d9fa-1c2b-4734-bdd1-1ace91460bea
spec:
  apply: Always
  databaseRef:
    name: etcd-quickstart
  maxRetries: 1
  timeout: 5m
  type: Restart
status:
  conditions:
  - lastTransitionTime: "2026-02-10T09:22:57Z"
    message: Restart is in progress
    observedGeneration: 1
    reason: Running
    status: "True"
    type: Running
  - lastTransitionTime: "2026-02-10T09:23:35Z"
    message: check pod ready; ConditionStatus:True; PodName:etcd-quickstart-0
    observedGeneration: 1
    status: "True"
    type: CheckPodReady--etcd-quickstart-0
  - lastTransitionTime: "2026-02-10T09:23:40Z"
    message: etcd cluster healthy; ConditionStatus:True; PodName:etcd-quickstart-0
    observedGeneration: 1
    status: "True"
    type: EtcdClusterHealthy--etcd-quickstart-0
  - lastTransitionTime: "2026-02-10T09:24:15Z"
    message: check pod ready; ConditionStatus:True; PodName:etcd-quickstart-2
    observedGeneration: 1
    status: "True"
    type: CheckPodReady--etcd-quickstart-2
  - lastTransitionTime: "2026-02-10T09:24:20Z"
    message: etcd cluster healthy; ConditionStatus:True; PodName:etcd-quickstart-2
    observedGeneration: 1
    status: "True"
    type: EtcdClusterHealthy--etcd-quickstart-2
  - lastTransitionTime: "2026-02-10T09:24:30Z"
    message: move leader; ConditionStatus:True; PodName:etcd-quickstart-1
    observedGeneration: 1
    status: "True"
    type: MoveLeader--etcd-quickstart-1
  - lastTransitionTime: "2026-02-10T09:25:05Z"
    message: check pod ready; ConditionStatus:True; PodName:etcd-quickstart-1
    observedGeneration: 1
    status: "True"
    type: CheckPodReady--etcd-quickstart-1
  - lastTransitionTime: "2026-02-10T09:25:10Z"
    message: etcd cluster healthy; ConditionStatus:True; PodName:etcd-quickstart-1
    observedGeneration: 1
    status: "True"
    type: EtcdClusterHealthy--etcd-quickstart-1
  - lastTransitionTime: "2026-02-10T09:25:15Z"
    message: Successfully restarted all the etcd members
    observedGeneration: 1
    reason: RestartEtcdPods
    status: "True"
    type: RestartEtcdPods
  - lastTransitionTime: "2026-02-10T09:25:16Z"
    message: Successfully restarted the etcd cluster
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

Reading the conditions:

- **`Running`** is set as soon as the request leaves the `Pending` phase.
- **`RestartEtcdPods`** tracks the whole rolling sequence. It is first written with
  `status: "False"` when the sequence starts and flips to `"True"` once every member has been
  recreated.
- **`MoveLeader--<pod>`** is the leadership handoff. It appears for exactly one pod — whichever one
  held the Raft leadership when the request started — and always before that pod is evicted. On a
  single member cluster it never appears, because there is no other voter to hand the leadership
  to.
- **`CheckPodReady--<pod>`** and **`EtcdClusterHealthy--<pod>`** are the two gates the operator
  waits on after evicting each pod: the pod must be recreated and `Ready`, and the cluster must
  have quorum again (`healthy members >= N/2 + 1`), before the next member is touched.
- **`Successful`** is the terminal condition, written after the `Etcd` object is resumed and handed
  back to the Provisioner operator.

> Note: the per-pod conditions above are recorded by the operator's retry bookkeeping — a step that
> completes on its very first attempt may not leave a condition behind at all. So the exact set of
> `EvictPod--<pod>` / `CheckPodReady--<pod>` / `EtcdClusterHealthy--<pod>` entries you see varies
> between runs. `RestartEtcdPods` and `Successful` are always there.

If you want to transfer the Raft leadership on its own, without restarting anything, use the
dedicated `MoveLeader` ops request type instead — it records an `EtcdLeaderMoved` condition and
optionally takes `spec.moveLeader.newLeader` to name the target member.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-restart
kubectl delete etcd -n demo etcd-quickstart
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Detail concepts of [EtcdOpsRequest object](/docs/guides/etcd/concepts/etcdopsrequest.md).
- Change the etcd tuning knobs of a running cluster with [Reconfigure](/docs/guides/etcd/reconfigure/reconfigure.md).
- Rotate the etcd credentials with [Rotate Authentication](/docs/guides/etcd/rotate-authentication/rotateauth.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
