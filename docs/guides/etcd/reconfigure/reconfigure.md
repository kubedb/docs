---
title: Reconfigure Etcd Cluster
menu:
  docs_{{ .version }}:
    identifier: etcd-reconfigure-details
    name: Reconfigure Configurations
    parent: etcd-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Etcd Cluster

This guide will show you how to use the `KubeDB` Ops-manager operator to reconfigure a running Etcd
cluster with an `EtcdOpsRequest`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md). Etcd support is behind an alpha feature gate, so make sure both operators are installed with `featureGates.Etcd=true`.

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Reconfigure Overview](/docs/guides/etcd/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this
tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Etcd

In this section we deploy a 3 member `Etcd` cluster that already carries a few tuning knobs, so
that we can watch them change. Below is the YAML of the `Etcd` CR we are going to create:

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
  configuration:
    tuning:
      quotaBackendBytes: 2147483648
      autoCompactionMode: periodic
      autoCompactionRetention: "1h"
      snapshotCount: 10000
  deletionPolicy: WipeOut
```

Here, the backend quota is 2 GiB and etcd compacts the keyspace history every hour.

Let's create the `Etcd` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure/etcd.yaml
etcd.kubedb.com/etcd-quickstart created
```

Now, wait until `etcd-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get etcd -n demo
NAME              VERSION   STATUS   AGE
etcd-quickstart   3.6.4     Ready    3m
```

### Check the current configuration

The tuning knobs are rendered onto the etcd command line, so the authoritative place to look is the
container args of a member pod:

```bash
$ kubectl get pod -n demo etcd-quickstart-0 -o jsonpath='{.spec.containers[?(@.name=="etcd")].args}' | jq .
[
  "--name=$(POD_NAME)",
  "--data-dir=/var/lib/etcd/data",
  "--initial-advertise-peer-urls=$(ETCD_MEMBER_PEER_URL)",
  "--listen-peer-urls=http://0.0.0.0:2380",
  "--listen-client-urls=http://0.0.0.0:2379",
  "--advertise-client-urls=$(ETCD_MEMBER_CLIENT_URL)",
  "--listen-metrics-urls=http://0.0.0.0:2381",
  "--quota-backend-bytes=2147483648",
  "--auto-compaction-mode=periodic",
  "--auto-compaction-retention=1h",
  "--snapshot-count=10000"
]
```

The per-member flags (`--name`, the peer/client URLs) are resolved from the downward API inside the
pod, which is why one PodSpec can serve every ordinal. The last four entries are the tuning knobs
from `spec.configuration.tuning`.

You can also confirm that the cluster itself is healthy before touching anything. Grab the
credentials from the auth Secret first:

```bash
$ export ETCD_PASSWORD=$(kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.password}' | base64 -d)

$ kubectl exec -it -n demo etcd-quickstart-0 -c etcd -- etcdctl \
    --endpoints=http://127.0.0.1:2379 --user=root:$ETCD_PASSWORD endpoint status -w table
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
|        ENDPOINT        |        ID        | VERSION | DB SIZE | IS LEADER | IS LEARNER | RAFT TERM | RAFT INDEX | RAFT APPLIED INDEX | ERRORS |
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
| http://127.0.0.1:2379  | 8e9e05c52164694d | 3.6.4   |   20 kB |     false |      false |         2 |         14 |                 14 |        |
+------------------------+------------------+---------+---------+-----------+------------+-----------+------------+--------------------+--------+
```

> **Note:** `etcdctl endpoint status` reports runtime state (version, backend size, Raft term,
> leadership) — it does not echo the process flags back. To verify a tuning change, read the
> container args as shown above, or the `Etcd` object's `spec.configuration.tuning`.

## Reconfigure using the tuning knobs

Now we will reconfigure this cluster to raise the backend quota to 8 GiB and switch auto compaction
from time based to revision based, keeping the last 1000 revisions.

Note that `autoCompactionRetention` is interpreted differently depending on the mode: it is a
duration (`1h`) in `periodic` mode and a revision count (`1000`) in `revision` mode. When you change
the mode, change the retention value along with it.

Below is the YAML of the `EtcdOpsRequest` CR that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: etcd-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: etcd-quickstart
  configuration:
    tuning:
      quotaBackendBytes: 8589934592
      autoCompactionMode: revision
      autoCompactionRetention: "1000"
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring the `etcd-quickstart` cluster.
- `spec.type` specifies that we are performing a `Reconfigure` operation.
- `spec.configuration.tuning` carries the knobs to apply. Only the knobs named here are touched —
  `snapshotCount` is not in the request, so it keeps its current value of `10000`.

> **Note:** `spec.configuration.applyConfig` is **not** supported for etcd — an ops request that uses
> it is rejected at admission and told to use `spec.configuration.tuning` instead. See the
> [overview](/docs/guides/etcd/reconfigure/overview.md#what-configuration-means-for-etcd) for why.

Let's create the `EtcdOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/reconfigure/etcdops-reconfigure.yaml
etcdopsrequest.ops.kubedb.com/etcd-reconfigure created
```

## This ops request restarts the members

Some KubeDB databases can apply a new configuration to a running server and only restart when the
changed settings demand it. **Etcd cannot.** The tuning knobs are etcd process command line flags,
and etcd has no live-reload path for them, so a `Reconfigure` on etcd is a rolling restart by
default: after re-rendering the PetSet, the operator rolls the member pods using the same leader-aware
sequence as the [Restart](/docs/guides/etcd/restart/restart.md) ops request — followers first, one
at a time, quorum verified between each, the Raft leader last and only after its leadership has been
transferred away.

That is why you will normally see a `RestartEtcdPods` condition on a successful etcd `Reconfigure`.

> Setting `spec.configuration.restart: "false"` does skip the rolling restart, but it does **not**
> give you a live reconfiguration: the running members keep their old flags until something else
> recreates them. Only the desired state on the `Etcd` object and the `PetSet` template change. Use
> it only if you intend to schedule the restart yourself later.

## Verify the new configuration

Let's wait for the `EtcdOpsRequest` to be `Successful`. Run the following command to watch it,

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME               TYPE          STATUS       AGE
etcd-reconfigure   Reconfigure   Successful   3m
```

We can see from the above output that the `EtcdOpsRequest` has succeeded. If we describe it, we get
an overview of the steps that were followed:

```bash
$ kubectl describe etcdopsrequest -n demo etcd-reconfigure
Name:         etcd-reconfigure
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-02-10T08:27:00Z
  Generation:          1
  Resource Version:    1548116
  UID:                 4f3daa11-c41b-4079-a8d8-1040931284ef
Spec:
  Apply:  IfReady
  Configuration:
    Tuning:
      Auto Compaction Mode:       revision
      Auto Compaction Retention:  1000
      Quota Backend Bytes:        8589934592
  Database Ref:
    Name:   etcd-quickstart
  Timeout:  5m
  Type:     Reconfigure
Status:
  Conditions:
    Last Transition Time:  2026-02-10T08:27:00Z
    Message:               Reconfigure is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-02-10T08:27:06Z
    Message:               Successfully updated the Etcd configuration
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-02-10T08:27:08Z
    Message:               Successfully re-rendered the petset with the new configuration
    Observed Generation:   1
    Reason:                UpdateEtcdPetSet
    Status:                True
    Type:                  UpdateEtcdPetSet
    Last Transition Time:  2026-02-10T08:27:45Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-quickstart-0
    Last Transition Time:  2026-02-10T08:27:50Z
    Message:               etcd cluster healthy; ConditionStatus:True; PodName:etcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EtcdClusterHealthy--etcd-quickstart-0
    Last Transition Time:  2026-02-10T08:28:25Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-quickstart-2
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-quickstart-2
    Last Transition Time:  2026-02-10T08:28:35Z
    Message:               move leader; ConditionStatus:True; PodName:etcd-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  MoveLeader--etcd-quickstart-1
    Last Transition Time:  2026-02-10T08:29:10Z
    Message:               check pod ready; ConditionStatus:True; PodName:etcd-quickstart-1
    Observed Generation:   1
    Status:                True
    Type:                  CheckPodReady--etcd-quickstart-1
    Last Transition Time:  2026-02-10T08:29:16Z
    Message:               Successfully restarted the etcd members with the new configuration
    Observed Generation:   1
    Reason:                RestartEtcdPods
    Status:                True
    Type:                  RestartEtcdPods
    Last Transition Time:  2026-02-10T08:29:18Z
    Message:               Successfully reconfigured the etcd cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason            Age    From                         Message
  ----    ------            ----   ----                         -------
  Normal  PauseDatabase     3m     KubeDB Ops-manager Operator  Pausing Etcd demo/etcd-quickstart
  Normal  UpdateDatabase    2m54s  KubeDB Ops-manager Operator  Successfully updated the Etcd configuration
  Normal  UpdateEtcdPetSet  2m52s  KubeDB Ops-manager Operator  Successfully re-rendered the petset with the new configuration
  Normal  RestartEtcdPods   44s    KubeDB Ops-manager Operator  Successfully restarted the etcd members with the new configuration
  Normal  ResumeDatabase    42s    KubeDB Ops-manager Operator  Resuming Etcd demo/etcd-quickstart
  Normal  Successful        42s    KubeDB Ops-manager Operator  Successfully reconfigured the etcd cluster
```

The three phases of the request are visible in the conditions: `UpdateDatabase` (the `Etcd` object's
`spec.configuration.tuning` is patched), `UpdateEtcdPetSet` (the PetSet template is re-rendered with
the new flags) and `RestartEtcdPods` (the members are rolled, leader last).

Now, wait until `etcd-quickstart` is `Ready` again. i.e,

```bash
$ kubectl get etcd -n demo
NAME              VERSION   STATUS   AGE
etcd-quickstart   3.6.4     Ready    12m
```

First, check the desired state on the `Etcd` object:

```bash
$ kubectl get etcd -n demo etcd-quickstart -o jsonpath='{.spec.configuration.tuning}' | jq .
{
  "autoCompactionMode": "revision",
  "autoCompactionRetention": "1000",
  "quotaBackendBytes": 8589934592,
  "snapshotCount": 10000
}
```

Note that `snapshotCount` is still `10000` — the ops request did not name it, so it was left alone.

Then check that the running members actually carry the new flags:

```bash
$ kubectl get pod -n demo etcd-quickstart-0 -o jsonpath='{.spec.containers[?(@.name=="etcd")].args}' | jq .
[
  "--name=$(POD_NAME)",
  "--data-dir=/var/lib/etcd/data",
  "--initial-advertise-peer-urls=$(ETCD_MEMBER_PEER_URL)",
  "--listen-peer-urls=http://0.0.0.0:2380",
  "--listen-client-urls=http://0.0.0.0:2379",
  "--advertise-client-urls=$(ETCD_MEMBER_CLIENT_URL)",
  "--listen-metrics-urls=http://0.0.0.0:2381",
  "--quota-backend-bytes=8589934592",
  "--auto-compaction-mode=revision",
  "--auto-compaction-retention=1000",
  "--snapshot-count=10000"
]
```

`--quota-backend-bytes` has been raised from `2147483648` to `8589934592` and the auto compaction
has switched from `periodic`/`1h` to `revision`/`1000`, so the reconfiguration is complete.

Finally, confirm that the cluster is still healthy and every member came back:

```bash
$ kubectl exec -it -n demo etcd-quickstart-0 -c etcd -- etcdctl \
    --endpoints=http://127.0.0.1:2379 --user=root:$ETCD_PASSWORD endpoint health --cluster -w table
+---------------------------------------------------------------+--------+------------+-------+
|                           ENDPOINT                            | HEALTH |    TOOK    | ERROR |
+---------------------------------------------------------------+--------+------------+-------+
| http://etcd-quickstart-0.etcd-quickstart-pods.demo.svc:2379    |   true | 3.1ms      |       |
| http://etcd-quickstart-1.etcd-quickstart-pods.demo.svc:2379    |   true | 3.4ms      |       |
| http://etcd-quickstart-2.etcd-quickstart-pods.demo.svc:2379    |   true | 2.9ms      |       |
+---------------------------------------------------------------+--------+------------+-------+
```

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdopsrequest -n demo etcd-reconfigure
kubectl delete etcd -n demo etcd-quickstart
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Detail concepts of [EtcdOpsRequest object](/docs/guides/etcd/concepts/etcdopsrequest.md).
- Restart the cluster without changing anything with [Restart](/docs/guides/etcd/restart/restart.md).
- Rotate the etcd credentials with [Rotate Authentication](/docs/guides/etcd/rotate-authentication/rotateauth.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
