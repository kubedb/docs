---
title: DocumentDB Compute Autoscaling
menu:
  docs_{{ .version }}:
    identifier: dc-auto-compute
    name: Compute Autoscaling
    parent: dc-auto-scaling
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Autoscaling the Compute Resource of a DocumentDB Cluster

This guide will show you how to use `KubeDB` to auto-scale the compute resources i.e. cpu and memory of a `DocumentDB` cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-Manager and Autoscaler operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `Metrics Server` from [here](https://github.com/kubernetes-sigs/metrics-server#installation)

- You should be familiar with the following `KubeDB` concepts:

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> A DocumentDB exposes the MongoDB wire protocol (port `10260`, TLS) backed by an internal PostgreSQL engine. Every pod runs two containers — `documentdb` (the data plane that the autoscaler tunes) and `documentdb-coordinator`. The `DocumentDBAutoscaler` `spec.compute.documentdb` block targets the `documentdb` container.

## How Compute Autoscaling Works

The `DocumentDBAutoscaler` compute loop is VPA-driven:

1. The Autoscaler operator runs an in-process VerticalPodAutoscaler recommender for the DB's PetSet (named after the DB, `dcdb`). The generated recommendation is published in the autoscaler's own `status.vpas` — this cluster has no standalone `VerticalPodAutoscaler` CRD, so you read the recommendation directly from the `DocumentDBAutoscaler` object.
2. When the recommendation differs from the current request by more than `resourceDiffPercentage` (and the pod is older than `podLifeTimeThreshold`, **or** the current request sits outside the `minAllowed`/`maxAllowed` band), the operator creates a `VerticalScaling` `DocumentDBOpsRequest` named `dcops-dcdb-<rand>`.
3. The Ops-Manager operator applies the new resources by rolling the PetSet pods one at a time.

This guide demonstrates a deterministic **scale-up to the recommendation floor**: the base database requests `500m`/`1Gi`, which is *below* the autoscaler's `minAllowed` of `600m`/`1.5Gi`. The recommendation is therefore capped *up* to `minAllowed`, which guarantees an ops request is created regardless of actual load.

## Deploy DocumentDB Cluster

Here, we are going to deploy a `DocumentDB` cluster with 3 replicas and deliberately low compute resources (`500m`/`1Gi`). Below is the YAML of the `DocumentDB` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: dcdb
  namespace: demo
spec:
  version: 'pg17-0.109.0'
  storageType: Durable
  deletionPolicy: Delete
  replicas: 3
  podTemplate:
    spec:
      containers:
        - name: documentdb
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 500m
              memory: 1Gi
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
```

Let's create the `DocumentDB` CR we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/autoscaler/compute/autoscaling-compute-object.yaml
```
documentdb.kubedb.com/dcdb created

Now, wait until `dcdb` has status `Ready`. i.e,

```bash
kubectl get docdb -n demo
```
NAME   NAMESPACE   VERSION        STATUS   AGE
dcdb   demo        pg17-0.109.0   Ready    113s

Let's check the `documentdb` container's resources of the pod,

```bash
kubectl get pod -n demo dcdb-0 -o jsonpath='{range .spec.containers[?(@.name=="documentdb")]}{.resources}{"\n"}{end}'
```
{"limits":{"cpu":"500m","memory":"1Gi"},"requests":{"cpu":"500m","memory":"1Gi"}}

You can see from the above output that the resources are the same as the ones we assigned while deploying the DocumentDB.

We are now ready to apply the `DocumentDBAutoscaler` CR to set up compute autoscaling for this database.

## Compute Resource Autoscaling

Here, we are going to set up compute resource autoscaling using a `DocumentDBAutoscaler` Object.

#### Create DocumentDBAutoscaler Object

In order to set up compute resource autoscaling for this database cluster, we have to create a `DocumentDBAutoscaler` CR with our desired configuration. Below is the YAML of the `DocumentDBAutoscaler` object that we are going to create,

```yaml
apiVersion: autoscaling.kubedb.com/v1alpha1
kind: DocumentDBAutoscaler
metadata:
  name: dcdb-compute-autoscaler
  namespace: demo
spec:
  databaseRef:
    name: dcdb
  opsRequestOptions:
    timeout: 5m
    apply: IfReady
  compute:
    documentdb:
      trigger: "On"
      podLifeTimeThreshold: 1m
      resourceDiffPercentage: 5
      minAllowed:
        cpu: 600m
        memory: 1.5Gi
      maxAllowed:
        cpu: "2"
        memory: 3Gi
      controlledResources: ["cpu", "memory"]
      containerControlledValues: "RequestsAndLimits"
```

Here,

- `spec.databaseRef.name` specifies that we are performing compute resource scaling operation on the `dcdb` database.
- `spec.compute.documentdb.trigger` specifies that compute autoscaling is enabled for this database.
- `spec.compute.documentdb.podLifeTimeThreshold` specifies the minimum lifetime for at least one of the pods to initiate a vertical scaling.
- `spec.compute.documentdb.resourceDiffPercentage` specifies the minimum resource difference (in percentage) between the current and recommended resources required to trigger an update. The default is 10%.
- `spec.compute.documentdb.minAllowed` specifies the minimum allowed resources for the database. Here it is set **above** the deployed resources, so the recommendation floor forces a scale-up.
- `spec.compute.documentdb.maxAllowed` specifies the maximum allowed resources for the database.
- `spec.compute.documentdb.controlledResources` specifies the resources that are controlled by the autoscaler.
- `spec.compute.documentdb.containerControlledValues` specifies which resource values should be controlled. The default is `RequestsAndLimits`.
- `spec.opsRequestOptions.apply` has two supported values: `IfReady` & `Always`. Use `IfReady` to process the opsRequest only when the database is Ready, and `Always` to process it irrespective of the database state.
- `spec.opsRequestOptions.timeout` specifies the maximum time for each step of the opsRequest.

Let's create the `DocumentDBAutoscaler` CR we have shown above,

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/autoscaler/compute/autoscaling-compute.yaml
```
documentdbautoscaler.autoscaling.kubedb.com/dcdb-compute-autoscaler created

#### Verify Autoscaling is set up successfully

Let's check that the `documentdbautoscaler` resource is created successfully,

```bash
kubectl get documentdbautoscaler -n demo
```
NAME                      AGE
dcdb-compute-autoscaler   11s

```bash
kubectl describe documentdbautoscaler dcdb-compute-autoscaler -n demo
```
Name:         dcdb-compute-autoscaler
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  autoscaling.kubedb.com/v1alpha1
Kind:         DocumentDBAutoscaler
Metadata:
  Creation Timestamp:  2026-06-30T13:58:55Z
  Generation:          1
  Owner References:
    API Version:           kubedb.com/v1alpha2
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  DocumentDB
    Name:                  dcdb
Spec:
  Compute:
    Documentdb:
      Container Controlled Values:  RequestsAndLimits
      Controlled Resources:
        cpu
        memory
      Max Allowed:
        Cpu:     2
        Memory:  3Gi
      Min Allowed:
        Cpu:                     600m
        Memory:                  1.5Gi
      Pod Life Time Threshold:   1m
      Resource Diff Percentage:  5
      Trigger:                   On
  Database Ref:
    Name:  dcdb
  Ops Request Options:
    Apply:        IfReady
    Max Retries:  1
    Timeout:      5m
Status:
  Checkpoints:
    Cpu Histogram:
      Bucket Weights:
        Index:              1
        Weight:             5995
        Index:              2
        Weight:             10000
        Index:              3
        Weight:             7164
      Reference Timestamp:  2026-06-30T14:00:00Z
      Total Weight:         1.1832553241627992
    First Sample Start:     2026-06-30T13:59:04Z
    Last Sample Start:      2026-06-30T14:02:03Z
    Last Update Time:       2026-06-30T14:02:23Z
    Ref:
      Container Name:     documentdb-coordinator
      Vpa Object Name:    dcdb
    Total Samples Count:  11
    Version:              v3
  Conditions:
    Last Transition Time:  2026-06-30T13:59:55Z
    Message:               Successfully created DocumentDBOpsRequest demo/dcops-dcdb-y87ecq
    Observed Generation:   1
    Reason:                CreateOpsRequest
    Status:                True
    Type:                  CreateOpsRequest
  Vpas:
    Conditions:
      Last Transition Time:  2026-06-30T13:59:22Z
      Status:                True
      Type:                  RecommendationProvided
    Recommendation:
      Container Recommendations:
        Container Name:  documentdb-coordinator
        Lower Bound:
          Cpu:     50m
          Memory:  131072k
        Target:
          Cpu:     50m
          Memory:  131072k
        Uncapped Target:
          Cpu:     50m
          Memory:  131072k
        Upper Bound:
          Cpu:           23700m
          Memory:        30735427949
        Container Name:  documentdb
        Lower Bound:
          Cpu:     600m
          Memory:  1536Mi
        Target:
          Cpu:     600m
          Memory:  1536Mi
        Uncapped Target:
          Cpu:     182m
          Memory:  131072k
        Upper Bound:
          Cpu:     2
          Memory:  3Gi
    Vpa Name:      dcdb
Events:            <none>

So, the `documentdbautoscaler` resource is created successfully.

We can verify from the above output that `status.vpas` contains the `RecommendationProvided` condition set to `True`, and `status.vpas[].recommendation.containerRecommendations` holds the actual recommendation. Notice the `documentdb` container `Target` of `600m`/`1536Mi` — the uncapped target (`182m`/`131072k`) was well below the band, so it was floored *up* to `minAllowed`. The `status.conditions` already reports `Successfully created DocumentDBOpsRequest demo/dcops-dcdb-y87ecq`.

The Autoscaler operator continuously watches the recommendation and creates a `DocumentDBOpsRequest` based on it whenever the pod resources need to be scaled up or down.

Let's watch the `documentdbopsrequest` in the demo namespace to see if any `documentdbopsrequest` object is created.

```bash
kubectl get documentdbopsrequest -n demo
```
NAME                TYPE              STATUS        AGE
dcops-dcdb-y87ecq   VerticalScaling   Progressing   13s

Let's wait for the ops request to become successful.

```bash
kubectl get documentdbopsrequest -n demo
```
NAME                TYPE              STATUS       AGE
dcops-dcdb-y87ecq   VerticalScaling   Successful   2m55s

We can see from the above output that the `DocumentDBOpsRequest` has succeeded. If we describe the `DocumentDBOpsRequest` (or print its YAML) we get an overview of the steps that were followed to scale the database.

```bash
kubectl get documentdbopsrequest -n demo dcops-dcdb-y87ecq -o yaml
```
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: dcops-dcdb-y87ecq
  namespace: demo
  ownerReferences:
  - apiVersion: autoscaling.kubedb.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: DocumentDBAutoscaler
    name: dcdb-compute-autoscaler
spec:
  apply: IfReady
  databaseRef:
    name: dcdb
  maxRetries: 1
  timeout: 5m0s
  type: VerticalScaling
  verticalScaling:
    documentdb:
      resources:
        limits:
          cpu: 600m
          memory: 1536Mi
        requests:
          cpu: 600m
          memory: 1536Mi
status:
  conditions:
  - message: Vertical Scaling is in progress
    reason: Running
    status: "True"
    type: Running
  - message: Successfully Set Raft Key OpsRequestProgressing
    reason: SetRaftKeyOpsRequestProgressing
    status: "True"
    type: SetRaftKeyOpsRequestProgressing
  - message: Successfully updated petsets resources
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: VerticalScaleSucceeded
    reason: VerticalScale
    status: "True"
    type: VerticalScale
  - message: Successfully Restarted Read Replicas
    reason: RestartReadReplicas
    status: "True"
    type: RestartReadReplicas
  - message: Successfully Vertically Scaled Database
    reason: Successful
    status: "True"
    type: Successful
  - message: Successfully Unset Raft Key OpsRequestProgressing
    reason: UnsetRaftKeyOpsRequestProgressing
    status: "True"
    type: UnsetRaftKeyOpsRequestProgressing
  observedGeneration: 1
  phase: Successful

Notice that the ops request body carries exactly the floored target (`600m`/`1536Mi`), and the rollout walks the cluster pod by pod (`SetRaftKeyOpsRequestProgressing` → `UpdatePetSets` → per-pod readiness checks → `RestartReadReplicas`) so the DocumentDB cluster stays available throughout.

Now, let's verify from the Pod and the DocumentDB object that the resources of the cluster database have been updated to the desired state.

```bash
kubectl get pod -n demo dcdb-0 -o jsonpath='{range .spec.containers[?(@.name=="documentdb")]}{.resources}{"\n"}{end}'
```
{"limits":{"cpu":"600m","memory":"1536Mi"},"requests":{"cpu":"600m","memory":"1536Mi"}}

```bash
kubectl get docdb -n demo dcdb -o json | jq -c '.spec.podTemplate.spec.containers[] | {name:.name, resources:.resources}'
```
{"name":"documentdb","resources":{"limits":{"cpu":"600m","memory":"1536Mi"},"requests":{"cpu":"600m","memory":"1536Mi"}}}
{"name":"documentdb-coordinator","resources":{"limits":{"memory":"256Mi"},"requests":{"cpu":"200m","memory":"256Mi"}}}

The above output verifies that we have successfully autoscaled the compute resources of the DocumentDB cluster database from `500m`/`1Gi` to `600m`/`1.5Gi`.

Finally, let's confirm the database is healthy over the MongoDB wire protocol:

```bash
PASS=$(kubectl get secret -n demo dcdb-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo dcdb-0 -c documentdb -- mongosh \
    "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'db.runCommand({ ping: 1 })'
```
{ ok: 1 }

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete documentdbautoscaler -n demo dcdb-compute-autoscaler
kubectl delete documentdb -n demo dcdb
kubectl delete ns demo
```

## Next Steps

- Learn how to autoscale the storage of a DocumentDB cluster in the [Storage Autoscaling](/docs/guides/documentdb/autoscaler/storage/index.md) guide.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
