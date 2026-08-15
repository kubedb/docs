---
title: Etcd Compute Resource Autoscaling
menu:
  docs_{{ .version }}:
    identifier: etcd-autoscaling-compute-description
    name: Autoscale Compute Resources
    parent: etcd-autoscaling-compute
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resources of an Etcd Cluster

This guide shows you how to use `KubeDB` to autoscale the compute resources (cpu and memory) of
an etcd cluster.

## Before You Begin

- At first, you need a Kubernetes cluster, and the `kubectl` command-line tool must be configured
  to communicate with your cluster.

- Install the `KubeDB` Provisioner, Ops-manager and Autoscaler operators in your cluster
  following the steps [here](/docs/setup/README.md).

- etcd support is behind an alpha feature gate. Make sure the operators are installed with
  `Etcd=true` enabled (Helm value `featureGates.Etcd=true`), otherwise the `Etcd` and
  `EtcdAutoscaler` CRDs are not reconciled.

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation).

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdAutoscaler](/docs/guides/etcd/concepts/etcdautoscaler.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [Compute Resource Autoscaling Overview](/docs/guides/etcd/autoscaler/compute/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout
this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in
> [docs/examples/etcd](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/etcd)
> folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Etcd

In this section, we deploy a three member etcd cluster with version `3.6.4`. In the next section
we set up autoscaling for it using the `EtcdAutoscaler` CRD. Below is the YAML of the `Etcd` CR
that we are going to create.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-autoscale
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
  podTemplate:
    spec:
      containers:
        - name: etcd
          resources:
            requests:
              cpu: "500m"
              memory: "1Gi"
            limits:
              cpu: "1"
              memory: "2Gi"
  deletionPolicy: WipeOut
```

Let's create the `Etcd` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/autoscaling/compute/etcd.yaml
etcd.kubedb.com/etcd-autoscale created
```

Now, wait until `etcd-autoscale` has the status `Ready`.

```bash
$ kubectl get etcd -n demo
NAME             VERSION   STATUS   AGE
etcd-autoscale   3.6.4     Ready    2m
```

Let's check the resources of the `etcd` container in one of the member pods,

```bash
$ kubectl get pod -n demo etcd-autoscale-0 -o json | jq '.spec.containers[] | select(.name=="etcd") | .resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

And the resources recorded on the `Etcd` object itself,

```bash
$ kubectl get etcd -n demo etcd-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "1",
    "memory": "2Gi"
  },
  "requests": {
    "cpu": "500m",
    "memory": "1Gi"
  }
}
```

You can see from the outputs above that the resources are the ones we assigned while deploying
the etcd cluster. We are now ready to apply the `EtcdAutoscaler` CRO.

### Compute Resource Autoscaling

#### Create EtcdAutoscaler Object

To set up compute resource autoscaling for this etcd cluster, we create an `EtcdAutoscaler` CRO
with the desired configuration. Below is the YAML of the `EtcdAutoscaler` object that we are
going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: EtcdAutoscaler
metadata:
  name: etcd-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: etcd-autoscale
  opsRequestOptions:
    apply: IfReady
    timeout: 5m
  compute:
    etcd:
      trigger: "On"
      podLifeTimeThreshold: 5m
      resourceDiffPercentage: 20
      minAllowed:
        cpu: 600m
        memory: 1.2Gi
      maxAllowed:
        cpu: 2
        memory: 4Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource autoscaling on the
  `etcd-autoscale` database.
- `spec.compute.etcd.trigger` specifies that compute resource autoscaling is enabled for the
  `etcd` container. The default is `Off`.
- `spec.compute.etcd.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the
  pods before a vertical scaling is initiated. The default is `15m`.
- `spec.compute.etcd.resourceDiffPercentage` specifies the minimum resource difference, in
  percent, that is worth acting on. The default is `50`. If the difference between the current
  and the recommended resources is smaller than this, the Autoscaler operator ignores the update.
- `spec.compute.etcd.minAllowed` specifies the minimum resources the recommender may recommend.
  The default is no minimum.
- `spec.compute.etcd.maxAllowed` specifies the maximum resources the recommender may recommend.
  The default is no maximum. Setting this is strongly recommended so that a transient spike
  cannot recommend an unschedulable pod.
- `spec.compute.etcd.controlledResources` specifies which resources are controlled by the
  autoscaler. The default is `["cpu", "memory"]`.
- `spec.compute.etcd.containerControlledValues` specifies which resource values are controlled,
  `RequestsAndLimits` or `RequestsOnly`. The default is `RequestsAndLimits`.
- `spec.opsRequestOptions` holds the options that are copied onto the `EtcdOpsRequest` the
  autoscaler creates. See
  [spec.timeout](/docs/guides/etcd/concepts/etcdopsrequest.md#spectimeout) and
  [spec.apply](/docs/guides/etcd/concepts/etcdopsrequest.md#specapply).

> **Note:** `spec.compute` also has an `etcd`-sibling field, `nodeTopology`, which snaps the
> recommendation to the node groups of a `NodeTopology` object. Keep in mind that the resulting
> `EtcdOpsRequest` carries no placement fields in its spec — see the
> [caveat in the overview](/docs/guides/etcd/autoscaler/compute/overview.md#a-note-on-node-topology).

Let's create the `EtcdAutoscaler` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/autoscaling/compute/etcd-compute-autoscaler.yaml
etcdautoscaler.autoscaling.kubedb.com/etcd-compute-autoscaler created
```

#### Verify autoscaling is set up successfully

Let's check that the `etcdautoscaler` resource was created successfully,

```bash
$ kubectl get etcdautoscaler -n demo
NAME                      AGE
etcd-compute-autoscaler   3m

$ kubectl describe etcdautoscaler etcd-compute-autoscaler -n demo
Name:         etcd-compute-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         EtcdAutoscaler
Metadata:
  Creation Timestamp:  2026-02-11T09:12:44Z
  Generation:          1
  Resource Version:    41822
  UID:                 1b0f2d6c-8c30-4b0b-9d4a-2d2c9d51f0a1
Spec:
  Compute:
    Etcd:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  4Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  1.2Gi
      Pod Life Time Threshold:   5m0s
      Resource Diff Percentage:  20
      Trigger:                   On
  Database Ref:
    Name:  etcd-autoscale
  Ops Request Options:
    Apply:    IfReady
    Timeout:  5m0s
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              0
        Weight:             10000
      Reference Timestamp:  2026-02-11T09:15:00Z
      Total Weight:         0.8733542386168607
    First Sample Start:     2026-02-11T09:12:41Z
    Last Sample Start:      2026-02-11T09:18:06Z
    Last Update Time:       2026-02-11T09:18:38Z
    Memory Histogram:
      Bucket Weights:
        Index:              11
        Weight:             10000
      Reference Timestamp:  2026-02-11T09:18:00Z
      Total Weight:         0.7827734162991002
    Ref:
      Container Name:     etcd
      Vpa Object Name:    etcd-autoscale
    Total Samples Count:  6
    Version:              v3
  Conditions:
    Last Transition Time:  2026-02-11T09:19:07Z
    Message:               Successfully created EtcdOpsRequest demo/etcdops-etcd-autoscale-vft8xm
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2026-02-11T09:13:12Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  etcd
        Lower Bound:
          Cpu:     600m
          Memory:  1.2Gi
        Target:
          Cpu:     600m
          Memory:  1.2Gi
        Uncapped Target:
          Cpu:     540m
          Memory:  1073741824
        Upper Bound:
          Cpu:     2
          Memory:  4Gi
    Vpa Name:      etcd-autoscale
Events:            <none>
```

So the `EtcdAutoscaler` resource was created successfully.

In the `Status.Vpas.Recommendation` section you can see the recommendation the operator generated
for the `etcd` container. The Autoscaler operator continuously watches that recommendation and,
when it differs from the running resources by more than `resourceDiffPercentage`, creates an
`EtcdOpsRequest` to move the cluster onto it.

#### Watch the generated EtcdOpsRequest

Let's watch the `etcdopsrequest` objects in the `demo` namespace. After a while you should see an
autoscaler-created one appear. Note the `etcdops-` name prefix — that is how the Autoscaler
operator names the requests it generates.

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME                             TYPE              STATUS        AGE
etcdops-etcd-autoscale-vft8xm    VerticalScaling   Progressing   45s
```

Let's wait for the ops request to become successful.

```bash
$ watch kubectl get etcdopsrequest -n demo
Every 2.0s: kubectl get etcdopsrequest -n demo
NAME                             TYPE              STATUS       AGE
etcdops-etcd-autoscale-vft8xm    VerticalScaling   Successful   3m
```

If we describe the `EtcdOpsRequest`, we get an overview of the steps that were followed to scale
the cluster. Note the owner reference back to the `EtcdAutoscaler` — that is what tells you this
request was generated rather than hand-written.

```bash
$ kubectl describe etcdopsrequest -n demo etcdops-etcd-autoscale-vft8xm
Name:         etcdops-etcd-autoscale-vft8xm
Namespace:    demo
Labels:       app.kubernetes.io/component=database
              app.kubernetes.io/instance=etcd-autoscale
              app.kubernetes.io/managed-by=kubedb.com
              app.kubernetes.io/name=etcds.kubedb.com
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         EtcdOpsRequest
Metadata:
  Creation Timestamp:  2026-02-11T09:19:07Z
  Generation:          1
  Owner References:
    API Version:           autoscaling.kubedb.com/v1alpha1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  EtcdAutoscaler
    Name:                  etcd-compute-autoscaler
    UID:                   1b0f2d6c-8c30-4b0b-9d4a-2d2c9d51f0a1
  Resource Version:        41903
  UID:                     6d2c1a55-0f61-4a3e-9c1b-2b9b0f6e0d77
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  etcd-autoscale
  Timeout:  5m0s
  Type:    VerticalScaling
  Vertical Scaling:
    Etcd:
      Resources:
        Limits:
          Cpu:     2
          Memory:  4Gi
        Requests:
          Cpu:     600m
          Memory:  1.2Gi
    Mode:  Restart
Status:
  Conditions:
    Last Transition Time:  2026-02-11T09:19:07Z
    Message:               Vertical Scaling is in progress
    Observed Generation:   1
    Reason:                Running
    Status:                True
    Type:                  Running
    Last Transition Time:  2026-02-11T09:19:15Z
    Message:               Successfully updated the petset resources
    Observed Generation:   1
    Reason:                UpdateEtcdPetSet
    Status:                True
    Type:                  UpdateEtcdPetSet
    Last Transition Time:  2026-02-11T09:19:16Z
    Message:               Successfully updated the Etcd resources
    Observed Generation:   1
    Reason:                UpdateDatabase
    Status:                True
    Type:                  UpdateDatabase
    Last Transition Time:  2026-02-11T09:21:41Z
    Message:               Successfully restarted the etcd members with the new resources
    Observed Generation:   1
    Reason:                VerticalScale
    Status:                True
    Type:                  VerticalScale
    Last Transition Time:  2026-02-11T09:21:42Z
    Message:               Successfully vertically scaled the etcd cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:                    <none>
```

The generated request uses the default vertical scaling mode, `Restart`. The Ops-manager operator
patches the PetSet template and then rolls the member pods one at a time, moving the Raft
leadership off the leader before restarting it and waiting for quorum in between, so the cluster
keeps serving clients throughout. The mode is not configurable from the `EtcdAutoscaler`; if you
want the resize applied without recreating pods, create an `EtcdOpsRequest` with
`spec.verticalScaling.mode: InPlace` by hand instead — see
[Vertical Scaling](/docs/guides/etcd/scaling/vertical-scaling/vertical-scaling.md).

#### Verify the new resources

Now let's verify from the Pod and from the `Etcd` object that the resources were updated.

```bash
$ kubectl get pod -n demo etcd-autoscale-0 -o json | jq '.spec.containers[] | select(.name=="etcd") | .resources'
{
  "limits": {
    "cpu": "2",
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1.2Gi"
  }
}

$ kubectl get etcd -n demo etcd-autoscale -o json | jq '.spec.podTemplate.spec.containers[0].resources'
{
  "limits": {
    "cpu": "2",
    "memory": "4Gi"
  },
  "requests": {
    "cpu": "600m",
    "memory": "1.2Gi"
  }
}
```

The output above confirms that the compute resources of the etcd cluster were autoscaled.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete etcdautoscaler -n demo etcd-compute-autoscaler
kubectl delete etcd -n demo etcd-autoscale
kubectl delete ns demo
```

## Next Steps

- [Storage Autoscaling of an Etcd cluster](/docs/guides/etcd/autoscaler/storage/storage-autoscale.md).
- Detail concepts of the [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
