---
title: Vertical Scaling Oracle
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-scaling-vertical-scale
    name: Scale Vertically
    parent: guides-oracle-scaling-vertical
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale Oracle

This guide will show you how to use the `KubeDB` Ops-manager operator to update the resources (CPU and memory) of an Oracle database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/scaling) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Create an Oracle Container Registry token, if you haven't created one already, by following the instructions in the guide below: [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important)

## Apply Vertical Scaling on Oracle

Here, we are going to deploy an `Oracle` standalone database using a supported version by `KubeDB` operator. Then we are going to apply vertical scaling on it.

**Deploy Oracle:**

In this section, we are going to deploy an Oracle standalone database. Then, in the next section, we will update the resources of the database using vertical scaling. Below is the YAML of the `Oracle` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sa-sample
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/scaling/standalone-minimal.yaml
```
oracle.kubedb.com/oracle-sa-sample created

Now, wait until `oracle-sa-sample` has status `Ready` and the pod prints the `DATABASE IS READY TO USE!!!` banner.

**Check resources before scaling:**

Let's check the Pod containers' resources of the database. Run the following command to get the resources of the `oracle-sa-sample-0` Pod,

```bash
kubectl get pod -n demo oracle-sa-sample-0 -o json | jq '.spec.containers[] | select(.name=="oracle") | .resources'
```
{
  "limits": {
    "cpu": "4",
    "memory": "10Gi"
  },
  "requests": {
    "cpu": "2",
    "memory": "7Gi"
  }
}

These are the default resources KubeDB assigns to the main `oracle` container. Now, we are going to update these resources using vertical scaling.

### Vertical Scaling

Here, we are going to update the resources of the database to meet the desired resources after scaling.

**Create OracleOpsRequest:**

In order to update the resources of the database, we have to create an `OracleOpsRequest` CR with our desired resources. Below is the YAML of the `OracleOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: standalone-vertical-scaling
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: oracle-sa-sample
  verticalScaling:
    node:
      resources:
        limits:
          memory: "10Gi"
          cpu: "5"
        requests:
          memory: "10Gi"
          cpu: "3"
```

Here,

- `spec.type` specifies that we are performing `VerticalScaling` on our database.
- `spec.databaseRef.name` specifies that we are performing vertical scaling operation on `oracle-sa-sample` database.
- `spec.verticalScaling.node.resources` specifies the desired resources (`requests`/`limits`) of the database node after scaling. KubeDB automatically recomputes the Oracle memory parameters (SGA/PGA) from the new container memory.

Let's create the `OracleOpsRequest` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/scaling/standalone-vertical-scaling.yaml
```
oracleopsrequest.ops.kubedb.com/standalone-vertical-scaling created

**Verify Oracle resources updated successfully:**

If everything goes well, the `OracleOpsRequest` will reach the `Successful` phase. Let's wait for it,

```bash
kubectl get oracleopsrequest -n demo standalone-vertical-scaling
```
NAME                          TYPE              STATUS       AGE
standalone-vertical-scaling   VerticalScaling   Successful   43s

We can see from the following `kubectl describe` output that the scaling completed successfully,

```bash
kubectl describe oracleopsrequest -n demo standalone-vertical-scaling
```
Name:         standalone-vertical-scaling
Namespace:    demo
...
Status:
  Conditions:
    Last Transition Time:  2026-06-22T18:59:05Z
    Message:               Oracle ops-request has started to vertically scaling the Oracle nodes
    Reason:                VerticalScaling
    Status:                True
    Type:                  VerticalScaling
    Last Transition Time:  2026-06-22T18:59:07Z
    Message:               Successfully updated PetSets Resources
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-22T18:59:22Z
    Message:               Successfully Restarted Pods With Resources
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2026-06-22T18:59:23Z
    Message:               Successfully completed the vertical scaling for Oracle
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful

Now, let's verify that the Pod resources have been updated to our desired values,

```bash
kubectl get pod -n demo oracle-sa-sample-0 -o json | jq '.spec.containers[] | select(.name=="oracle") | .resources'
```
{
  "limits": {
    "cpu": "5",
    "memory": "10Gi"
  },
  "requests": {
    "cpu": "3",
    "memory": "10Gi"
  }
}

The resources of the Oracle database have been updated successfully.

## Vertically scaling a DataGuard cluster

The same `OracleOpsRequest` works for a DataGuard cluster — point `spec.databaseRef.name` at the DataGuard database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: dataguard-vertical-scaling
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: oracle-dg-sample
  verticalScaling:
    node:
      resources:
        limits:
          memory: "10Gi"
          cpu: "5"
        requests:
          memory: "10Gi"
          cpu: "3"
```

The operator updates the PetSet resources, recomputes the Oracle memory parameters (SGA/PGA) from the new container memory, and performs a rolling restart across the DataGuard pods.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo standalone-vertical-scaling
kubectl patch -n demo oracle/oracle-sa-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo oracle-sa-sample
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Learn how to [expand the volume](/docs/guides/oracle/volume-expansion/volume-expansion.md) of an Oracle database.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
