---
title: Horizontal Scaling Hazelcast
menu:
  docs_{{ .version }}:
    identifier: hazelcast-horizontal-scaling
    name: Horizontal Scaling
    parent: hz-horizontal-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Horizontal Scale Hazelcast

This guide will give an overview on how KubeDB Ops-manager operator scales up or down `Hazelcast` cluster members.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Community and Enterprise operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Hazelcast](/docs/guides/hazelcast/concepts/hazelcast.md)
  - [HazelcastOpsRequest](/docs/guides/hazelcast/concepts/hazelcast-opsrequest.md)
  - [Horizontal Scaling Overview](/docs/guides/hazelcast/scaling/horizontal-scaling/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/hazelcast](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hazelcast) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Apply Horizontal Scaling on Hazelcast

Here, we are going to deploy a `Hazelcast` database using a supported version by `KubeDB` operator. Then we are going to apply horizontal scaling on it.

### Prepare Hazelcast Database

Now, we are going to deploy a `Hazelcast` database with version `5.5.2`.

### Deploy Hazelcast

Before deploying hazelcast we need to create license secret since we are running enterprise version of hazelcast.


```bash
kubectl create secret generic hz-license-key -n demo --from-literal=licenseKey=TrialLicense#10Nodes#eyJhbGxvd2VkTmF0aXZlTWVtb3J5U2l6ZSI6MTAwLCJhbGxvd2VkTnVtYmVyT2ZOb2RlcyI6MTAsImFsbG93ZWRUaWVyZWRTdG9yZVNpemUiOjAsImFsbG93ZWRUcGNDb3JlcyI6MCwiY3JlYXRpb25EYXRlIjoxNzQ4ODQwNDc3LjYzOTQ0NzgxNiwiZXhwaXJ5RGF0ZSI6MTc1MTQxNDM5OS45OTk5OTk5OTksImZlYXR1cmVzIjpbMCwyLDMsNCw1LDYsNyw4LDEwLDExLDEzLDE0LDE1LDE3LDIxLDIyXSwiZ3JhY2VQZXJpb2QiOjAsImhhemVsY2FzdFZlcnNpb24iOjk5LCJvZW0iOmZhbHNlLCJ0cmlhbCI6dHJ1ZSwidmVyc2lvbiI6IlY3In0=.6PYD6i-hejrJ5Czgc3nYsmnwF7mAI-78E8LFEuYp-lnzXh_QLvvsYx4ECD0EimqcdeG2J5sqUI06okLD502mCA==
secret/hz-license-key created
```

In this section, we are going to deploy a Hazelcast database. Then, in the next section we will scale the database using `HazelcastOpsRequest` CRD. Below is the YAML of the `Hazelcast` CR that we are going to create,

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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/scaling/horizontal-scaling/hazelcast.yaml
hazelcast.kubedb.com/hz-prod created
```

Now, wait until `hz-prod` has status `Ready`. i.e,

```bash
$ kubectl get hz -n demo
NAME      TYPE                  VERSION   STATUS   AGE
hz-prod   kubedb.com/v1alpha2   5.5.2     Ready    4m
```

Let's check the number of member nodes this database has from the Hazelcast object, number of pods the Statefulset have,

```bash
$ kubectl get hazelcast -n demo hz-prod -o json | jq '.spec.replicas'
3

$ kubectl get statefulset -n demo hz-prod -o json | jq '.spec.replicas'
3

$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=hz-prod" | wc -l
4
```

You can see from all the above outputs that the database has 3 member nodes.

We are now ready to apply the `HazelcastOpsRequest` CR to scale this database.

## Scale Up Members

Here, we are going to scale up the member nodes of the database to meet the desired number of member nodes after scaling.

### Create HazelcastOpsRequest

In order to scale up the member nodes of the database, we have to create a `HazelcastOpsRequest` CR with our desired number of members. Below is the YAML of the `HazelcastOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-scale-up
  namespace: demo
spec:
  databaseRef:
    name: hz-prod
  type: HorizontalScaling
  horizontalScaling:
    hazelcast: 4
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `hz-prod` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.hazelcast` specifies the desired number of member nodes after scaling.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/scaling/horizontal-scaling/hz-hscale-up.yaml
hazelcastopsrequest.ops.kubedb.com/hz-hscale-up created
```

### Verify hazelcast node scaled up successfully

If everything goes well, `KubeDB` Enterprise operator will update the number of member nodes in `Hazelcast` object and related `StatefulSets` and `Pods`.

Let's wait for `HazelcastOpsRequest` to be `Successful`. Run the following command to watch `HazelcastOpsRequest` CR,

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME                   TYPE                STATUS       AGE
hazelcast-scale-up     HorizontalScaling   Successful   2m5s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded. If we describe the `HazelcastOpsRequest` we will get an overview of the steps that were followed to scale the database.

```bash
$ kubectl describe hazelcastopsrequest -n demo hazelcast-scale-up
Name:         hazelcast-scale-up
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         HazelcastOpsRequest
Metadata:
  Creation Timestamp:  2025-08-19T10:35:27Z
  Generation:          1
  Resource Version:    5472886
  UID:                 38184783-1a3a-41ca-9847-d46a71435e32
Spec:
  Apply:  IfReady
  Database Ref:
    Name:  hz-prod
  Horizontal Scaling:
    Hazelcast:  4
  Type:         HorizontalScaling
Status:
  Conditions:
    Last Transition Time:  2025-08-19T10:35:27Z
    Message:               Hazelcast ops-request has started to horizontally scaling the nodes
    Observed Generation:   1
    Reason:                HorizontalScaling
    Status:                True
    Type:                  HorizontalScaling
    Last Transition Time:  2025-08-19T10:36:00Z
    Message:               ScaleUp hz-prod nodes
    Observed Generation:   1
    Reason:                HorizontalScale
    Status:                True
    Type:                  HorizontalScale
    Last Transition Time:  2025-08-19T10:35:40Z
    Message:               patch stateful set; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  PatchStatefulSet
    Last Transition Time:  2025-08-19T10:35:58Z
    Message:               is node in cluster; ConditionStatus:True
    Observed Generation:   1
    Status:                True
    Type:                  IsNodeInCluster
    Last Transition Time:  2025-08-19T10:36:00Z
    Message:               Successfully completed horizontally scale Hazelcast cluster
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                    Age    From                         Message
  ----     ------                                    ----   ----                         -------
  Normal   Starting                                  3m2s   KubeDB Ops-manager Operator  Start processing for HazelcastOpsRequest: demo/hazelcast-scale-up
  Normal   Starting                                  3m2s   KubeDB Ops-manager Operator  Pausing Hazelcast databse: demo/hz-prod
  Normal   Successful                                3m2s   KubeDB Ops-manager Operator  Successfully paused Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hazelcast-scale-up
  Warning  patch stateful set; ConditionStatus:True  2m49s  KubeDB Ops-manager Operator  patch stateful set; ConditionStatus:True
  Warning  is node in cluster; ConditionStatus:True  2m31s  KubeDB Ops-manager Operator  is node in cluster; ConditionStatus:True
  Normal   HorizontalScale                           2m29s  KubeDB Ops-manager Operator  ScaleUp hz-prod nodes
  Normal   Starting                                  2m29s  KubeDB Ops-manager Operator  Resuming Hazelcast database: demo/hz-prod
  Normal   Successful                                2m29s  KubeDB Ops-manager Operator  Successfully resumed Hazelcast database: demo/hz-prod for HazelcastOpsRequest: hazelcast-scale-up
```

Now, we are going to verify the number of member nodes this database has from the Hazelcast object, number of pods the Stateful have,

```bash
$ kubectl get hazelcast -n demo hz-prod -o json | jq '.spec.replicas'
4

$ kubectl get statefulset -n demo hz-prod -o json | jq '.spec.replicas'
4

$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=hz-prod" | wc -l
5
```

From all the above outputs we can see that the number of member nodes are 4. That means we have successfully scaled up the member nodes of the Hazelcast database.

## Scale Down Members

Here, we are going to scale down the member nodes of the database to meet the desired number of member nodes after scaling.

### Create HazelcastOpsRequest

In order to scale down the member nodes of the database, we have to create a `HazelcastOpsRequest` CR with our desired number of members. Below is the YAML of the `HazelcastOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HazelcastOpsRequest
metadata:
  name: hazelcast-scale-down
  namespace: demo
spec:
  databaseRef:
    name: hz-prod
  type: HorizontalScaling
  horizontalScaling:
    hazelcast: 2
```

Here,

- `spec.databaseRef.name` specifies that we are performing horizontal scaling operation on `hz-prod` database.
- `spec.type` specifies that we are performing `HorizontalScaling` on our database.
- `spec.horizontalScaling.hazelcast` specifies the desired number of member nodes after scaling.

Let's create the `HazelcastOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hazelcast/scaling/horizontal-scaling/hz-hscale-down.yaml
hazelcastopsrequest.ops.kubedb.com/hz-hscale-down created
```

### Verify Member nodes scaled down successfully

If everything goes well, `KubeDB` Enterprise operator will update the number of member nodes in `Hazelcast` object and related `PetSets` and `Pods`.

Let's wait for `HazelcastOpsRequest` to be `Successful`. Run the following command to watch `HazelcastOpsRequest` CR,

```bash
$ kubectl get hazelcastopsrequest -n demo
NAME             TYPE                STATUS       AGE
hz-hscale-down   HorizontalScaling   Successful   2m38s
```

We can see from the above output that the `HazelcastOpsRequest` has succeeded.

Now, we are going to verify the number of member nodes this database has from the Hazelcast object, number of pods the PetSet have,

```bash
$ kubectl get hazelcast -n demo hz-prod -o json | jq '.spec.replicas'
2

$ kubectl get statefulset -n demo hz-prod -o json | jq '.spec.replicas'
2

$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=hz-prod" | wc -l
3
```

From all the above outputs we can see that the number of member nodes are 3. That means we have successfully scaled down the member nodes of the Hazelcast database.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete hazelcastopsrequest -n demo hz-hscale-up hz-hscale-down
kubectl delete hazelcast -n demo hz-prod
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Hazelcast object](/docs/guides/hazelcast/concepts/hazelcast.md).
- Monitor your Hazelcast database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/hazelcast/monitoring/prometheus-operator.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
