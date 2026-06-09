---
title: Update Oracle Version
menu:
  docs_{{ .version }}:
    identifier: oracle-version-upgrading
    name: Version Upgrading
    parent: oracle-update-version
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Oracle Version

This guide will show you how to use `KubeDB` ops-manager operator to update the version of `Oracle` cr.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Oracle](/docs/guides/oracle/concepts/oracle.md)
  - [OracleOpsRequest](/docs/guides/oracle/concepts/opsrequest.md)
  - [Updating Overview](/docs/guides/oracle/update-version/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/oracle/update-version/versionupgrading/yamls](/docs/guides/oracle/update-version/versionupgrading/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Version Updating on Oracle

Here, we are going to deploy a `Oracle` instance using a supported version by `KubeDB` provisioner. Then we are going to apply update-ops-request on it.

#### Prepare Oracle

At first, we are going to deploy a Oracle instance using a supported `Oracle` version. In the next two sections, we are going to find out the supported versions and version update constraints.

**Find supported OracleVersion:**

When you have installed `KubeDB`, it has created `OracleVersion` CR for all supported `Oracle` versions. Let's check the supported versions:

```bash
$ kubectl get oracleversion
NAME       VERSION   DB_IMAGE                      DEPRECATED   AGE
1.7.4      1.7.4     oracle/oracle:v1.7.4                       5m
1.8.4      1.8.4     oracle/oracle:v1.8.4                       5m
1.9.7      1.9.7     oracle/oracle:v1.9.7                       5m
1.10.1     1.10.1    oracle/oracle:v1.10.1                      5m
1.11.5     1.11.5    oracle/oracle:v1.11.5                      5m
1.12.6     1.12.6    oracle/oracle:v1.12.6                      5m
1.13.4     1.13.4    oracle/oracle:v1.13.4                      5m
1.14.1     1.14.1    oracle/oracle:v1.14.1                      5m
1.15.1     1.15.1    oracle/oracle:v1.15.1                      5m
1.16.1     1.16.1    oracle/oracle:v1.16.1                      5m
1.17.0     1.17.0    oracle/oracle:v1.17.0                      5m
1.18.0     1.18.0    oracle/oracle:v1.18.0                      5m
```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `Oracle`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `OracleVersion` for the `Oracle` instance that we will update from this version to another. In the next section, we are going to verify version update constraints.

**Check update Constraints:**

Oracle supports rolling version updates. You can update from any currently running version to a newer patch or minor version. Major version jumps should follow the Oracle upstream upgrade notes. For example, you can update directly from `1.17.0` to `1.18.0`.

Let's get one of the `oracleversion` YAMLs:

```bash
$ kubectl get oracleversion 1.17.0 -o yaml | kubectl neat
apiVersion: catalog.kubedb.com/v1alpha1
kind: OracleVersion
metadata:
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
  name: "1.17.0"
spec:
  coordinator:
    image: ghcr.io/kubedb/oracle-coordinator:v0.10.0
  db:
    image: oracle/oracle:v1.17.0
  exporter:
    image: container-registry.oracle.com/database/observability-exporter:2.2.1
  version: "21.3.0"
```

**Deploy Oracle Instance:**

In this section, we are going to deploy a Oracle instance. Then, in the next section, we will update the version of the database using `UpdateVersion`. Below is the YAML of the `Oracle` cr that we are going to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sample
  namespace: demo
spec:
  version: "21.3.0"
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` cr we have shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/update-version/versionupgrading/yamls/oracle.yaml
oracle.kubedb.com/oracle-sample created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `Oracle` objects using Kubernetes API. When a `Oracle` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc.

Now, watch `Oracle` is going to `Running` state and also watch `PetSet` and its pod is created and going to `Running` state:

```bash
$ watch -n 3 kubectl get oracle -n demo
Every 3.0s: kubectl get oracle -n demo

NAME              VERSION   STATUS   AGE
oracle-sample     1.17.0    Ready    3m42s

$ watch -n 3 kubectl get petset -n demo oracle-sample
Every 3.0s: kubectl get petset -n demo oracle-sample

NAME              READY   AGE
oracle-sample     3/3     4m17s

$ watch -n 3 kubectl get pod -n demo
Every 3.0s: kubectl get pods -n demo

NAME                READY   STATUS    RESTARTS   AGE
oracle-sample-0     1/1     Running   0          4m55s
oracle-sample-1     1/1     Running   0          4m12s
oracle-sample-2     1/1     Running   0          3m38s
```

Let's verify the `Oracle`, the `PetSet` and its `Pod` image version:

```bash
$ kubectl get oracle -n demo oracle-sample -o=jsonpath='{.spec.version}{"\n"}'
1.17.0

$ kubectl get petset -n demo oracle-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
oracle/oracle:v1.17.0

$ kubectl get pod -n demo oracle-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
oracle/oracle:v1.17.0
```

We are ready to apply version updating on this `Oracle` instance.

#### UpdateVersion

Here, we are going to update `Oracle` from version `1.17.0` to `1.18.0`.

**Create OracleOpsRequest:**

To update the instance, you have to create a `OracleOpsRequest` cr with your desired version that is supported by `KubeDB`. Below is the YAML of the `OracleOpsRequest` cr that we are going to create:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: qdops-update-version
  namespace: demo
spec:
  type: UpdateVersion
  updateVersion:
    targetVersion: "1.18.0"
  databaseRef:
    name: oracle-sample
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `oracle-sample` Oracle database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies the expected version `1.18.0` after updating.

Let's create the `OracleOpsRequest` cr we have shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/oracle/update-version/versionupgrading/yamls/update_version.yaml
oracleopsrequest.ops.kubedb.com/qdops-update-version created
```

**Verify Oracle version updated successfully:**

If everything goes well, `KubeDB` ops-manager operator will update the image of `Oracle`, `PetSet`, and its `Pod`.

At first, we will wait for `OracleOpsRequest` to be successful. Run the following command to watch `OracleOpsRequest` cr:

```bash
$ watch -n 3 kubectl get OracleOpsRequest -n demo qdops-update-version
Every 3.0s: kubectl get OracleOpsRequest -n demo qdops-update-version

NAME                     TYPE            STATUS       AGE
qdops-update-version     UpdateVersion   Successful   3m12s
```

We can see from the above output that the `OracleOpsRequest` has succeeded. If we describe the `OracleOpsRequest`, we shall see that the `Oracle`, `PetSet`, and its `Pod` have updated with a new image.

```bash
$ kubectl describe OracleOpsRequest -n demo qdops-update-version
Name:         qdops-update-version
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         OracleOpsRequest
Spec:
  Database Ref:
    Name:  oracle-sample
  Type:    UpdateVersion
  Update Version:
    Target Version:  1.18.0
Status:
  Conditions:
    Last Transition Time:  2026-05-01T10:00:00Z
    Message:               Oracle ops request is updating version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2026-05-01T10:00:05Z
    Message:               Successfully updated PetSets update strategy type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-05-01T10:00:30Z
    Message:               Successfully updated pod images
    Observed Generation:   1
    Reason:                UpdatePetSetImage
    Status:                True
    Type:                  UpdatePetSetImage
    Last Transition Time:  2026-05-01T10:02:45Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason            Age    From                        Message
  ----    ------            ----   ----                        -------
  Normal  PauseDatabase     3m12s  KubeDB Enterprise Operator  Pausing Oracle demo/oracle-sample
  Normal  PauseDatabase     3m12s  KubeDB Enterprise Operator  Successfully paused Oracle demo/oracle-sample
  Normal  Updating          3m12s  KubeDB Enterprise Operator  Updating PetSets
  Normal  Updating          3m12s  KubeDB Enterprise Operator  Successfully Updated PetSets
  Normal  UpdatePetSetImage 1m10s  KubeDB Enterprise Operator  Successfully Updated pod images
  Normal  ResumeDatabase    1m10s  KubeDB Enterprise Operator  Resuming Oracle demo/oracle-sample
  Normal  ResumeDatabase    1m10s  KubeDB Enterprise Operator  Successfully resumed Oracle demo/oracle-sample
  Normal  Successful        1m10s  KubeDB Enterprise Operator  Successfully Updated Database
```

Now, we are going to verify whether the `Oracle`, `PetSet` and its `Pod` have updated with the new image. Let's check:

```bash
$ kubectl get oracle -n demo oracle-sample -o=jsonpath='{.spec.version}{"\n"}'
1.18.0

$ kubectl get petset -n demo oracle-sample -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
oracle/oracle:v1.18.0

$ kubectl get pod -n demo oracle-sample-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
oracle/oracle:v1.18.0
```

You can see above that our `Oracle` has been updated with the new version. It verifies that we have successfully updated our Oracle instance.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo oracle-sample
kubectl delete OracleOpsRequest -n demo qdops-update-version
```