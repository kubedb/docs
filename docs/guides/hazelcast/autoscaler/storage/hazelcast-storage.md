---
title: Hazelcast  Autoscaling
menu:
  docs_{{ .version }}:
    identifier: hz-storage-auto-scaling-
    name:  Cluster
    parent: hz-storage-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Autoscaling of a Hazelcast  Cluster

This guide will show you how to use `KubeDB` to autoscale the storage of a Hazelcast  cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- Install Prometheus from [here](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack)

- You must have a `StorageClass` that supports volume expansion.

- You should be familiar with the following `KubeDB` concepts:
    - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
    - [HazelcastAutoscaler](/docs/guides/hazelcast/concepts/hazelcastautoscaler.md)
    - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)
    - [Storage Autoscaling Overview](/docs/guides/hazelcast/autoscaler/storage/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/hazelcast](/docs/examples/hazelcast) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Storage Autoscaling of  Cluster

At first verify that your cluster has a storage class, that supports volume expansion. Let's check,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER            RECLAIMPOLICY   VOLUMEBINDINGMODE   ALLOWVOLUMEEXPANSION   AGE
standard (default)   kubernetes.io/gce-pd   Delete          Immediate           true                   2m49s
```

We can see from the output the `standard` storage class has `ALLOWVOLUMEEXPANSION` field as true. So, this storage class supports volume expansion. We can use it.

Now, we are going to deploy a `Hazelcast`  using a supported version by `KubeDB` operator. Then we are going to apply `HazelcastAutoscaler` to set up autoscaling.

#### Deploy Hazelcast 

In this section, we are going to deploy a Hazelcast  cluster with version `5.5.2`.  Then, in the next section we will set up autoscaling for this cluster using `HazelcastAutoscaler` CRD. Below is the YAML of the `Hazelcast` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Hazelcast
metadata:
  name: hazelcast-dev
  namespace: demo
spec:
  replicas: 2
  version: 5.5.2
  licenseSecret:
    name: hz-license-key
  podTemplate:
    spec:
      containers:
        - name: hazelcast
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: longhorn
  storageType: Durable
  deletionPolicy: WipeOut
```

Let's create the `Hazelcast` CRO we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/autoscaler/hazelcast.yaml
hazelcast.kubedb.com/hazelcast-dev created
```

Now, wait until `hazelcast-dev` has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo -w
NAME             TYPE                    VERSION   STATUS         AGE
hazelcast-dev    kubedb.com/v1alpha2     5.5.2    Provisioning   0s
hazelcast-dev    kubedb.com/v1alpha2     5.5.2     Provisioning   24s
.
.
hazelcast-dev    kubedb.com/v1alpha2     5.5.2     Ready          92s
```

Let's check volume size from statefulset, and from the persistent volume,

```bash
$ kubectl get statefulset -n demo hazelcast-dev -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1Gi"

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                           STORAGECLASS          REASON     AGE
pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51   1Gi        RWO            Delete           Bound    demo/hazelcast-dev-data-hazelcast-dev-0         standard              <unset>    40s
pvc-f068d245-718b-4561-b452-f3130bb260f6   1Gi        RWO            Delete           Bound    demo/hazelcast-dev-data-hazelcast-dev-1         standard              <unset>    35s
```

You can see the statefulset has 1GB storage, and the capacity of all the persistent volume is also 1GB.

We are now ready to apply the `HazelcastAutoscaler` CRO to set up storage autoscaling for this cluster.

### Storage Autoscaling

Here, we are going to set up storage autoscaling using a HazelcastAutoscaler Object.

#### Create HazelcastAutoscaler Object

In order to set up vertical autoscaling for this  cluster, we have to create a `HazelcastAutoscaler` CRO with our desired configuration. Below is the YAML of the `HazelcastAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: HazelcastAutoscaler
metadata:
  name: hz-storage-autoscaler-
  namespace: demo
spec:
  databaseRef:
    name: hazelcast-dev
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  storage:
    hazelcast:
      trigger: "On"
      expansionMode: "Online"
      usageThreshold: 1
      scalingThreshold: 50

```

Here,

- `spec.clusterRef.name` specifies that we are performing vertical scaling operation on `hazelcast-dev` cluster.
- `spec.storage.node.trigger` specifies that storage autoscaling is enabled for this cluster.
- `spec.storage.node.usageThreshold` specifies storage usage threshold, if storage usage exceeds `1%` then storage autoscaling will be triggered.
- `spec.storage.node.scalingThreshold` specifies the scaling threshold. Storage will be scaled to `50%` of the current amount.
- It has another field `spec.storage.node.expansionMode` to set the opsRequest volumeExpansionMode, which support two values: `Online` & `Offline`. Default value is `Online`.

Let's create the `HazelcastAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/autoscaling/storage/hazelcast-storage-autoscaler.yaml
hazelcastautoscaler.autoscaling.kubedb.com/hz-storage-autoscaler created
```

#### Storage Autoscaling is set up successfully

Let's check that the `hazelcastautoscaler` resource is created successfully,

```bash
NAME                             AGE
hz-storage-autoscaler   8s


$ kubectl describe hazelcastautoscaler -n demo hz-storage-autoscaler
Name:         hz-storage-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         HazelcastAutoscaler
Metadata:
  Creation Timestamp:  2025-08-20T05:52:49Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  Hazelcast
    Name:                  hazelcast-dev
    UID:                   ad17f549-4b10-4064-99fe-578894872a92
  Resource Version:        5637400
  UID:                     9b6ba215-73d3-4b6d-bff7-495c01449185
Spec:
  Database Ref:
    Name:  hazelcast-dev
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m0s
  Storage:
    Hazelcast:
      Expansion Mode:  Online
      Scaling Rules:
        Applies Upto:     
        Threshold:        50pc
      Scaling Threshold:  50
      Trigger:            On
      Usage Threshold:    1
Status:
  Conditions:
    Last Transition Time:  2025-08-20T05:53:07Z
    Message:               Successfully created HazelcastOpsRequest demo/hzops-hazelcast-dev-a89pwf
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
Events:                    <none>

```
So, the `hazelcastautoscaler` resource is created successfully.

Now, for this demo, we are going to manually fill up the persistent volume to exceed the `usageThreshold` using `dd` command to see if storage autoscaling is working or not.

Let's exec into the cluster pod and fill the cluster volume using the following commands:

```bash
 $ kubectl exec -it -n demo hazelcast-dev-0 -- bash
hazelcast@hazelcast-dev-0:~$ df -h /data/hazelcast
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51  974M  168K  958M   1% /data/hazelcast
hazelcast@hazelcast-dev-0:~$ dd if=/dev/zero of=/data/hazelcast/file.img bs=600M count=1
1+0 records in
1+0 records out
629145600 bytes (629 MB, 600 MiB) copied, 7.44144 s, 84.5 MB/s
hazelcast@hazelcast-dev-0:~$ df -h /data/hazelcast
Filesystem                                              Size  Used Avail Use% Mounted on
/dev/standard/pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51  974M  601M  358M  63% /data/hazelcast
```

So, from the above output we can see that the storage usage is 63%, which exceeded the `usageThreshold` 1%.

Let's watch the `hazelcastopsrequest` in the demo namespace to see if any `hazelcastopsrequest` object is created. After some time you'll see that a `hazelcastopsrequest` of type `VolumeExpansion` will be created based on the `scalingThreshold`.

```bash
$ watch kubectl get hazelcastopsrequest -n demo
Every 2.0s: kubectl get hazelcastopsrequest -n demo
NAME                        TYPE              STATUS        AGE
hzops-hazelcast-dev-a89pwf   VolumeExpansion   Progressing   111s
```

Let's wait for the ops request to become successful.

```bash
$ kubectl get hazelcastopsrequest -n demo 
NAME                        TYPE              STATUS        AGE
hzops-hazelcast-dev-sa4thn  VolumeExpansion   Successful    97s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed to expand the volume of the cluster.

```bash
$ kubectl describe hzops -n demo hzops-hazelcast-dev-a89pwf
Name:         hzops-hazelcast-dev-a89pwf
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=hazelcast-dev
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=hazelcasts.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-20T05:53:07Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  HazelcastAutoscaler
    Name:                  hz-storage-autoscaler-
    UID:                   9b6ba215-73d3-4b6d-bff7-495c01449185
  Resource Version:        5638392
  UID:                     4146ba75-2d77-42a4-813c-160c5a008595
Spec:
  Apply:  IfReady
  Database Ref:
    Name:   hazelcast-dev
  Timeout:  5m0s
  Type:     VolumeExpansion
  Volume Expansion:
    Hazelcast:  1531054080
    Mode:       Online
Status:
  Conditions:
    Last Transition Time:  2025-08-20T05:53:07Z
    Message:               Hazelcast ops-request has started to expand volume of hazelcast nodes.
    Observed Generation:   1
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2025-08-20T05:53:37Z
    Message:               successfully deleted the statefulSets with orphan propagation policy
    Observed Generation:   1
    Reason:                OrphanStatefulSetPods
    Status:                True
    Type:                  OrphanStatefulSetPods
    Last Transition Time:  2025-08-20T05:53:17Z
    Message:               get statefulset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetStatefulset
    Last Transition Time:  2025-08-20T05:53:17Z
    Message:               delete statefulset; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  DeleteStatefulset
    Last Transition Time:  2025-08-20T05:58:17Z
    Message:               successfully updated PVC sizes
    Observed Generation:   1
    Reason:                VolumeExpansionSucceeded
    Status:                True
    Type:                  VolumeExpansionSucceeded
    Last Transition Time:  2025-08-20T05:53:47Z
    Message:               get pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetPvc
    Last Transition Time:  2025-08-20T05:53:47Z
    Message:               patch pvc; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchPvc
    Last Transition Time:  2025-08-20T05:58:07Z
    Message:               compare storage; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  CompareStorage
    Last Transition Time:  2025-08-20T05:58:28Z
    Message:               successfully reconciled the Hazelcast resources
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2025-08-20T05:58:48Z
    Message:               StatefulSet is recreated
    Observed Generation:   1
    Reason:                ReadyStatefulSets
    Status:                True
    Type:                  ReadyStatefulSets
    Last Transition Time:  2025-08-20T05:58:38Z
    Message:               get stateful set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  GetStatefulSet
    Last Transition Time:  2025-08-20T05:58:38Z
    Message:               Successfully completed volumeExpansion for Hazelcast
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   Starting                                  6m46s  KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hzops-hazelcast-dev-a89pwf
  Warning  get statefulset; ConditionStatus:True     6m36s  KubeDB Ops-manager Operator  get statefulset; ConditionStatus:True
  Warning  delete statefulset; ConditionStatus:True  6m36s  KubeDB Ops-manager Operator  delete statefulset; ConditionStatus:True
  Warning  get statefulset; ConditionStatus:True     6m26s  KubeDB Ops-manager Operator  get statefulset; ConditionStatus:True
  Normal   OrphanStatefulSetPods                     6m16s  KubeDB Ops-manager Operator  successfully deleted the statefulSets with orphan propagation policy
  Warning  get pvc; ConditionStatus:True             6m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True           6m6s   KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    5m56s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True             5m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             5m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             4m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     3m46s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  patch pvc; ConditionStatus:True           3m36s  KubeDB Ops-manager Operator  patch pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:False    3m26s  KubeDB Ops-manager Operator  compare storage; ConditionStatus:False
  Warning  get pvc; ConditionStatus:True             3m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             3m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m56s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m46s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m36s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m26s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m16s  KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             2m6s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             116s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  get pvc; ConditionStatus:True             106s   KubeDB Ops-manager Operator  get pvc; ConditionStatus:True
  Warning  compare storage; ConditionStatus:True     106s   KubeDB Ops-manager Operator  compare storage; ConditionStatus:True
  Normal   VolumeExpansionSucceeded                  96s    KubeDB Ops-manager Operator  successfully updated PVC sizes
  Normal   UpdateStatefulSets                        85s    KubeDB Ops-manager Operator  successfully reconciled the Hazelcast resources
  Warning  get stateful set; ConditionStatus:True    75s    KubeDB Ops-manager Operator  get stateful set; ConditionStatus:True
  Normal   ReadyStatefulSets                         75s    KubeDB Ops-manager Operator  StatefulSet is recreated
  Normal   Starting                                  75s    KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hazelcast-dev
  Normal   Successful                                75s    KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hazelcast-dev for HazelcastOpsRequest: hzops-hazelcast-dev-a89pwf
  Warning  get stateful set; ConditionStatus:True    65s    KubeDB Ops-manager Operator  get stateful set; ConditionStatus:True
  Normal   ReadyStatefulSets                         65s    KubeDB Ops-manager Operator  StatefulSet is recreated

```

Now, we are going to verify from the `Statefulset`, and the `Persistent Volume` whether the volume of the  cluster has expanded to meet the desired state, Let's check,

```bash
$ kubectl get statefulset -n demo hazelcast-dev -o json | jq '.spec.volumeClaimTemplates[].spec.resources.requests.storage'
"1531054080"
$ kubectl get pv -n demo
NAME                                       CAPACITY      ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                           STORAGECLASS          REASON     AGE
pvc-129be4b9-f7e8-489e-8bc5-cd420e680f51   1462Mi        RWO            Delete           Bound    demo/hazelcast-dev-data-hazelcast-dev-0         longhorn              <unset>    30m5s
pvc-f068d245-718b-4561-b452-f3130bb260f6   1462Mi        RWO            Delete           Bound    demo/hazelcast-dev-data-hazelcast-dev-1         longhorn              <unset>    30m1s
```

The above output verifies that we have successfully autoscaled the volume of the Hazelcast  cluster.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequests -n demo zops-hazelcast-dev-a89p
kubectl delete hazelcastautoscaler -n demo hz-storage-autoscaler
kubectl delete hz -n demo hazelcast-dev
```

## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Monitor your Hazelcast database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
