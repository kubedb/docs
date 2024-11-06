---
title: Update Version of Solr
menu:
  docs_{{ .version }}:
    identifier: sl-update-version-solr
    name: Solr
    parent: sl-update-version
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update version of Solr

This guide will show you how to use `KubeDB` Ops-manager operator to update the version of `Solr` Combined or Topology.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Updating Overview](/docs/guides/solr/update-version/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Solr](/docs/examples/Solr) directory of [kubedb/docs](https://github.com/kube/docs) repository.

## Prepare Solr

Now, we are going to deploy a `Solr` replicaset database with version `3.6.8`.

### Deploy Solr

In this section, we are going to deploy a Solr topology cluster. Then, in the next section we will update the version using `SolrOpsRequest` CRD. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr-cluster
  namespace: demo
spec:
  version: 9.4.1
  zookeeperRef:
    name: zoo
    namespace: demo
  topology:
    overseer:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 1
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    coordinator:
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi

```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/update-version/solr.yaml
solr.kubedb.com/solr-cluster created
```

Now, wait until `solr-cluster` created has status `Ready`. i.e,

```bash
$ kubectl get kf -n demo -w                                                                                                                                           
NAME         TYPE            VERSION   STATUS         AGE
Solr-prod   kubedb.com/v1   3.5.2     Provisioning   0s
Solr-prod   kubedb.com/v1   3.5.2     Provisioning   55s
.
.
Solr-prod   kubedb.com/v1   3.5.2     Ready          119s
```

We are now ready to apply the `SolrOpsRequest` CR to update.

### update Solr Version

Here, we are going to update `Solr` from `9.4.1` to `9.6.1`.

#### Create SolrOpsRequest:

In order to update the version, we have to create a `SolrOpsRequest` CR with your desired version that is supported by `KubeDB`. Below is the YAML of the `SolrOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: solr-update-version
  namespace: demo
spec:
  databaseRef:
    name: solr-cluster
  type: UpdateVersion
  updateVersion:
    targetVersion: 9.6.1
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `solr-cluster` Solr.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version of the database `9.6.1`.

> **Note:** If you want to update combined Solr, you just refer to the `Solr` combined object name in `spec.databaseRef.name`. To create a combined Solr, you can refer to the [Solr Combined](/docs/guides/solr/clustering/combined_cluster.md) guide.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Solr/update-version/update-version-ops.yaml
solropsrequest.ops.kubedb.com/solr-update-version created
```

#### Verify Solr version updated successfully

If everything goes well, `KubeDB` Ops-manager operator will update the image of `Solr` object and related `PetSets` and `Pods`.

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CR,

```bash
$ kubectl get Solropsrequest -n demo
NAME                   TYPE            STATUS        AGE
solr-update-version    UpdateVersion   Successful    2m6s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to update the database version.

```bash
$ kubectl get slops -n demo solr-update-version  -oyaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"ops.kubedb.com/v1alpha1","kind":"SolrOpsRequest","metadata":{"annotations":{},"name":"solr-update-version","namespace":"demo"},"spec":{"databaseRef":{"name":"solr-cluster"},"type":"UpdateVersion","updateVersion":{"targetVersion":"9.6.1"}}}
  creationTimestamp: "2024-11-06T07:10:42Z"
  generation: 1
  name: solr-update-version
  namespace: demo
  resourceVersion: "1753051"
  uid: caf69d71-1894-4da1-931f-a8f7ff8088a7
spec:
  apply: IfReady
  databaseRef:
    name: solr-cluster
  type: UpdateVersion
  updateVersion:
    targetVersion: 9.6.1
status:
  conditions:
  - lastTransitionTime: "2024-11-06T07:10:42Z"
    message: Solr ops-request has started to update version
    observedGeneration: 1
    reason: UpdateVersion
    status: "True"
    type: UpdateVersion
  - lastTransitionTime: "2024-11-06T07:10:50Z"
    message: successfully reconciled the Solr with updated version
    observedGeneration: 1
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - lastTransitionTime: "2024-11-06T07:13:20Z"
    message: Successfully Restarted Solr nodes
    observedGeneration: 1
    reason: RestartPods
    status: "True"
    type: RestartPods
  - lastTransitionTime: "2024-11-06T07:10:55Z"
    message: get pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    observedGeneration: 1
    status: "True"
    type: GetPod--solr-cluster-overseer-0
  - lastTransitionTime: "2024-11-06T07:10:55Z"
    message: evict pod; ConditionStatus:True; PodName:solr-cluster-overseer-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--solr-cluster-overseer-0
  - lastTransitionTime: "2024-11-06T07:11:00Z"
    message: running pod; ConditionStatus:False
    observedGeneration: 1
    status: "False"
    type: RunningPod
  - lastTransitionTime: "2024-11-06T07:11:40Z"
    message: get pod; ConditionStatus:True; PodName:solr-cluster-data-0
    observedGeneration: 1
    status: "True"
    type: GetPod--solr-cluster-data-0
  - lastTransitionTime: "2024-11-06T07:11:40Z"
    message: evict pod; ConditionStatus:True; PodName:solr-cluster-data-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--solr-cluster-data-0
  - lastTransitionTime: "2024-11-06T07:12:25Z"
    message: get pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    observedGeneration: 1
    status: "True"
    type: GetPod--solr-cluster-coordinator-0
  - lastTransitionTime: "2024-11-06T07:12:25Z"
    message: evict pod; ConditionStatus:True; PodName:solr-cluster-coordinator-0
    observedGeneration: 1
    status: "True"
    type: EvictPod--solr-cluster-coordinator-0
  - lastTransitionTime: "2024-11-06T07:13:20Z"
    message: Successfully updated SOlr version
    observedGeneration: 1
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
                                                              61s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/Solr-prod for SolrOpsRequest: Solr-update-version
```

Now, we are going to verify whether the `Solr` and the related `PetSets` and their `Pods` have the new version image. Let's check,

```bash
$ kubectl get sl -n demo solr-cluster -o=jsonpath='{.spec.version}{"\n"}'
9.6.1
~/y/s/ops (main|✚23…) $ kubectl get petset -n demo Solr-cluster-data -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
ghcr.io/appscode-images/daily/solr:9.6.1_20241024@sha256:0996340eff1e59bcac49eb8f96c28f0a3efb061f0e91b2053bfb7dade860c0e4

```

You can see from above, our `Solr` has been updated with the new version. So, the updateVersion process is successfully completed.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solropsrequest -n demo solr-update-version
kubectl delete solr -n demo solr-cluster
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
