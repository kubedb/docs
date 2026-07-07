---
title: Oracle Volume Expansion
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-volume-expansion-details
    name: Volume Expansion
    parent: guides-oracle-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Oracle Volume Expansion

This guide will show you how to use `KubeDB` Ops-manager operator to expand the volume of an Oracle database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You must have a `StorageClass` that supports volume expansion (i.e. its provisioner sets `allowVolumeExpansion: true`). This tutorial uses `longhorn`.

```bash
kubectl get storageclass
```
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  12d
longhorn               driver.longhorn.io      Delete          Immediate              true                   12d

> **Note:** `local-path` has `ALLOWVOLUMEEXPANSION: false`, so it cannot be used for volume expansion. Use a storage class such as `longhorn` that supports it.

- To keep things isolated, this tutorial uses a separate namespace called `demo`.

```bash
kubectl create ns demo
```
namespace/demo created

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/volume-expansion](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/volume-expansion) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Create an Oracle Container Registry token, if you haven't created one already, by following the instructions in the guide below: [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important)

## Deploy Oracle

In this section, we are going to deploy an Oracle standalone database with `10Gi` of storage on the `longhorn` storage class. Below is the YAML of the `Oracle` CR,

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
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR and wait until it is `Ready`.

Once ready, let's check the size of the PersistentVolumeClaim used by the database,

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=oracle-sa-sample
```
NAME                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-oracle-sa-sample-0   Bound    pvc-7c115ba3-ed65-4437-992c-aa8b789b0019   10Gi       RWO            longhorn       8m37s

## Expand Volume

Here, we are going to expand the volume of the database to `12Gi`.

### Create OracleOpsRequest

In order to expand the volume of the database, we have to create an `OracleOpsRequest` CR with our desired volume size. Below is the YAML,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: standalone-volume-expention
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: oracle-sa-sample
  volumeExpansion:
    mode: "Offline"
    node: 12Gi
```

Here,

- `spec.type` specifies that we are performing a `VolumeExpansion` operation.
- `spec.databaseRef.name` specifies the database `oracle-sa-sample`.
- `spec.volumeExpansion.node` specifies the desired size of the database node's PVC.
- `spec.volumeExpansion.mode` specifies the expansion mode. `Offline` recreates the pod after the PVC is resized; `Online` resizes the PVC without recreating the pod (the underlying storage class must support online expansion). For a DataGuard cluster, you can also set `spec.volumeExpansion.observer` to resize the observer's PVC.

Let's create the `OracleOpsRequest`,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/volume-expansion/standalone-volume-expention.yaml
```
oracleopsrequest.ops.kubedb.com/standalone-volume-expention created

### Verify the volume expanded

Let's wait for the `OracleOpsRequest` to become `Successful`,

```bash
kubectl get oracleopsrequest -n demo standalone-volume-expention
```
NAME                          TYPE              STATUS       AGE
standalone-volume-expention   VolumeExpansion   Successful   61s

```bash
kubectl describe oracleopsrequest -n demo standalone-volume-expention
```
Name:         standalone-volume-expention
Namespace:    demo
...
Status:
  Conditions:
    Last Transition Time:  2026-06-22T20:04:52Z
    Message:               Oracle ops-request has started to expand volume of oracle nodes
    Reason:                VolumeExpansion
    Status:                True
    Type:                  VolumeExpansion
    Last Transition Time:  2026-06-22T20:05:14Z
    Message:               successfully deleted the petSets with orphan propagation policy
    Reason:                OrphanPetSetPods
    Status:                True
    Type:                  OrphanPetSetPods
    Last Transition Time:  2026-06-22T20:05:33Z
    Message:               successfully reconciled the Oracle resources
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-22T20:05:50Z
    Message:               Successfully completed volume expansion for Oracle
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful

The operator has updated the `Oracle` spec and the PetSet `volumeClaimTemplate` to `12Gi`,

```bash
kubectl get oracle -n demo oracle-sa-sample -o jsonpath='{.spec.storage.resources.requests.storage}'
```
12Gi

The capacity reported by the `PersistentVolumeClaim` changes once the **storage provisioner** finishes resizing the underlying volume. While the resize is in progress, the PVC carries a `Resizing` condition and the requested size (`spec.resources.requests.storage`) updates ahead of the reported capacity (`status.capacity.storage`),

```bash
kubectl get pvc data-oracle-sa-sample-0 -n demo -o jsonpath='req={.spec.resources.requests.storage} cap={.status.capacity.storage} cond={.status.conditions[*].type}'
```
req=12Gi cap=10Gi cond=Resizing

> **Note (test environment):** On the single-node longhorn dev cluster used to capture this guide, the `OracleOpsRequest` reached the `Successful` phase and both the `Oracle` spec and the PetSet `volumeClaimTemplate` were updated to `12Gi`, but the longhorn volume remained in the `Resizing` state and the PVC capacity had not yet been reflected as `12Gi` at the time of writing. Volume expansion depends on the CSI driver fully completing the resize; on a production-grade storage class the PVC capacity updates to the new size once resizing completes. Always confirm the final size with `kubectl get pvc -n demo data-oracle-sa-sample-0 -o jsonpath='{.status.capacity.storage}'`.

## Expanding a DataGuard cluster's volume

The same `OracleOpsRequest` works for a DataGuard cluster — point `spec.databaseRef.name` at the DataGuard database. You can also resize the observer's PVC through `spec.volumeExpansion.observer`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: dataguard-volume-expention
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: oracle-dg-sample
  volumeExpansion:
    mode: "Offline"
    node: 12Gi
```

The operator expands the PVC of each DataGuard database pod (and the observer, if specified) to the requested size.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo standalone-volume-expention
kubectl patch -n demo oracle/oracle-sa-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo oracle-sa-sample
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Learn how to [vertically scale](/docs/guides/oracle/scaling/vertical-scaling/vertical-scaling.md) an Oracle database.
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
