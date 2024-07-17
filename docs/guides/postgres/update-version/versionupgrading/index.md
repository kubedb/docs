---
title: Updating Postgres version
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-updating-version
    name: Update version
    parent: guides-postgres-updating
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---
> New to KubeDB? Please start [here](/docs/README.md).

# update version of Postgres

This guide will show you how to use `KubeDB` ops-manager operator to update the version of `Postgres` cr.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Postgres](/docs/guides/postgres/concepts/postgres.md)
  - [PostgresOpsRequest](/docs/guides/postgres/concepts/opsrequest.md)
  - [updating Overview](/docs/guides/postgres/update-version/overview/index.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/update-version/versionupdating/yamls](/docs/guides/postgres/update-version/versionupgrading/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Apply Version updating on Postgres

Here, we are going to deploy a `Postgres` instance using a supported version by `KubeDB` provisioner. Then we are going to apply update-ops-request on it.

#### Prepare Postgres

At first, we are going to deploy a Postgres using supported `Postgres` version whether it is possible to update from this version to another. In the next two sections, we are going to find out the supported version and version update constraints.

**Find supported PostgresVersion:**

When you have installed `KubeDB`, it has created `PostgresVersion` CR for all supported `Postgres` versions. Let's check support versions,

```bash
$ kubectl get postgresversion
NAME                       VERSION   DISTRIBUTION   DB_IMAGE                               DEPRECATED   AGE
10.16                      10.16     Official       postgres:10.16-alpine                               63s
10.16-debian               10.16     Official       postgres:10.16                                      63s
10.19                      10.19     Official       postgres:10.19-bullseye                             63s
10.19-bullseye             10.19     Official       postgres:10.19-bullseye                             63s
11.11                      11.11     Official       postgres:11.11-alpine                               63s
11.11-debian               11.11     Official       postgres:11.11                                      63s
11.14                      11.14     Official       postgres:11.14-alpine                               63s
11.14-bullseye             11.14     Official       postgres:11.14-bullseye                             63s
11.14-bullseye-postgis     11.14     PostGIS        postgis/postgis:11-3.1                              63s
12.6                       12.6      Official       postgres:12.6-alpine                                63s
12.6-debian                12.6      Official       postgres:12.6                                       63s
12.9                       12.9      Official       postgres:12.9-alpine                                63s
12.9-bullseye              12.9      Official       postgres:12.9-bullseye                              63s
12.9-bullseye-postgis      12.9      PostGIS        postgis/postgis:12-3.1                              63s
13.2                       13.2      Official       postgres:13.2-alpine                                63s
13.2-debian                13.2      Official       postgres:13.2                                       63s
13.5                       13.5      Official       postgres:13.5-alpine                                63s
13.5-bullseye              13.5      Official       postgres:13.5-bullseye                              63s
13.5-bullseye-postgis      13.5      PostGIS        postgis/postgis:13-3.1                              63s
14.1                       14.1      Official       postgres:14.1-alpine                                63s
14.1-bullseye              14.1      Official       postgres:14.1-bullseye                              63s
14.1-bullseye-postgis      14.1      PostGIS        postgis/postgis:14-3.1                              63s
9.6.21                     9.6.21    Official       postgres:9.6.21-alpine                              63s
9.6.21-debian              9.6.21    Official       postgres:9.6.21                                     63s
9.6.24                     9.6.24    Official       postgres:9.6.24-alpine                              63s
9.6.24-bullseye            9.6.24    Official       postgres:9.6.24-bullseye                            63s
timescaledb-2.1.0-pg11     11.11     TimescaleDB    timescale/timescaledb:2.1.0-pg11-oss                63s
timescaledb-2.1.0-pg12     12.6      TimescaleDB    timescale/timescaledb:2.1.0-pg12-oss                63s
timescaledb-2.1.0-pg13     13.2      TimescaleDB    timescale/timescaledb:2.1.0-pg13-oss                63s
timescaledb-2.5.0-pg14.1   14.1      TimescaleDB    timescale/timescaledb:2.5.0-pg14-oss                63s


```

The version above that does not show `DEPRECATED` `true` is supported by `KubeDB` for `Postgres`. You can use any non-deprecated version. Now, we are going to select a non-deprecated version from `PostgresVersion` for `Postgres` Instance that will be possible to update from this version to another version. In the next section, we are going to verify version update constraints.

**Check update Constraints:**

When you are trying to update make sure that from current version the target version update is supported.

| Current Version   | updateable Minor Versions | updateable Major Versions                                                            |
| ----------------- | -------------------------- | ------------------------------------------------------------------------------------- |
| `9.6.21`          | `9.6.24`                   | `10.16`, `11.11`, `12.6`, `13.2 `                                                     |
| `9.6.21-debian`   | `9.6.24-bullseye`          | `12.6-debian`, `13.2-debian`                                                          |
| `9.6.24`          | -                          | `10.19`, `11.14`, `12.9`, `13.5`, `14.1`                                              |
| `9.6.24-bullseye` | -                          | `10.19-bullseye`, `11.14-bullseye`, `12.9-bullseye`, `13.5-bullseye`, `14.1-bullseye` |
| `10.16`           | `10.19`                    | `11.11`, `12.6`, `13.2`                                                               |
| `10.16-debian`    | `10.19-bullseye `          | `11.11-debian`                                                                        |
| `10.19`           | -                          | `11.14`, `12.9`, `13.5`, `14.1`                                                       |
| `10.19-bullseye`  | -                          | `11.14-bullseye`, `12.9-bullseye`, `13.5-bullseye`, `14.1-bullseye`                   |
| `11.11`           | ` 11.14`                   | `12.6`, `13.2`                                                                        |
| `11.11-debian`    | `11.14-bullseye`           | -                                                                                     |
| `11.14`           | -                          | `12.9`, `13.5`, `14.1`                                                                |
| `11.14-bullseye`  | -                          | `12.9-bullseye`, `13.5-bullseye`, `14.1-bullseye`                                     |
| `12.6`            | `12.9`                     | `13.2`                                                                                |
| `12.6-debian`     | `12.9-bullseye`            | `13.2-debian`                                                                         |
| `12.9`            | -                          | `13.5`, `14.1`                                                                        |
| `12.9-bullseye`   | -                          | `13.5-bullseye`, `14.1-bullseye`                                                      |
| `13.2`            | `13.5`                     | -                                                                                     |
| `13.2-debian`     | `13.5-bullseye`            | -                                                                                     |
| `13.5`            | -                          | `14.1`                                                                                |
| `13.5-bullseye`   | -                          | `14.1-bullseye`                                                                       |
| `14.1`            | -                          | -                                                                                     |
| `14.1-bullseye`   | -                          | -                                                                                     |

For Example: If you want to update from 9.6.21 to 14.1. From the table, you can see that you can't directly update from 9.6.21 to 14.1. So what you need to is first update 9.6.21 to 9.6.24. then try to update from 9.6.24 to 14.1.

Let's get one of the `postgresversion` YAML:  
```bash
$ kubectl get postgresversion 13.2 -o yaml | kubectl neat
apiVersion: catalog.kubedb.com/v1alpha1
kind: PostgresVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2021.11.24
    helm.sh/chart: kubedb-catalog-v2021.11.24
  name: "13.13"
spec:
  coordinator:
    image: kubedb/pg-coordinator:v0.8.0
  db:
    image: postgres:13.2-alpine
  distribution: Official
  exporter:
    image: prometheuscommunity/postgres-exporter:v0.9.0
  initContainer:
    image: kubedb/postgres-init:0.4.0
  podSecurityPolicies:
    databasePolicyName: postgres-db
  securityContext:
    runAsAnyNonRoot: false
    runAsUser: 70
  stash:
    addon:
      backupTask:
        name: postgres-backup-13.1
      restoreTask:
        name: postgres-restore-13.1
  version: "13.13"


```


**Deploy Postgres Instance:**

In this section, we are going to deploy a Postgres Instance. Then, in the next section, we will update the version of the database using updating. Below is the YAML of the `Postgres` cr that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg
  namespace: demo
spec:
  version: "11.22"
  replicas: 3
  standbyMode: Hot
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

Let's create the `Postgres` cr we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/update-version/versionupgrading/yamls/postgres.yaml
postgres.kubedb.com/pg created
```

**Wait for the database to be ready:**

`KubeDB` operator watches for `Postgres` objects using Kubernetes API. When a `Postgres` object is created, `KubeDB` operator will create a new PetSet, Services, and Secrets, etc. A secret called `pg-auth` (format: <em>{postgres-object-name}-auth</em>) will be created storing the password for postgres superuser.
Now, watch `Postgres` is going to  `Running` state and also watch `PetSet` and its pod is created and going to `Running` state,

```bash
$ watch -n 3 kubectl get postgres -n demo
Every 3.0s: kubectl get postgres -n demo     
            
NAME   VERSION   STATUS   AGE
pg     11.11     Ready    3m17s

$ watch -n 3 kubectl get sts -n demo pg
Every 3.0s: kubectl get sts -n demo pg                              ac-emon: Tue Nov 30 11:38:12 2021

NAME   READY   AGE
pg     3/3     4m17s

$ watch -n 3 kubectl get pod -n demo
Every 3.0s: kubectl get pods -n demo

Every 3.0s: kubectl get pods -n demo                                ac-emon: Tue Nov 30 11:39:03 2021

NAME   READY   STATUS    RESTARTS   AGE
pg-0   2/2     Running   0          4m55s
pg-1   2/2     Running   0          3m15s
pg-2   2/2     Running   0          3m11s

```

Let's verify the `Postgres`, the `PetSet` and its `Pod` image version,

```bash
$ kubectl get pg -n demo pg -o=jsonpath='{.spec.version}{"\n"}'
11.11

$  kubectl get sts -n demo pg -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
postgres:11.11-alpine

$ kubectl get pod -n demo pg-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
postgres:11.11-alpine
```

We are ready to apply updating on this `Postgres` Instance.

#### UpdateVersion

Here, we are going to update `Postgres` Instance from `11.11` to `13.2`.

**Create PostgresOpsRequest:**

To update the Instance, you have to create a `PostgresOpsRequest` cr with your desired version that supported by `KubeDB`. Below is the YAML of the `PostgresOpsRequest` cr that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: pg-update
  namespace: demo
spec:
  type: UpdateVersion
  updateVersion:
    targetVersion: "13.13"
  databaseRef:
    name: pg
```

Here,

- `spec.databaseRef.name` specifies that we are performing operation on `pg-group` Postgres database.
- `spec.type` specifies that we are going to perform `UpdateVersion` on our database.
- `spec.updateVersion.targetVersion` specifies expected version `13.2` after updating.

Let's create the `PostgresOpsRequest` cr we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/update-version/versionupgrading/yamls/update_version.yaml
postgresopsrequest.ops.kubedb.com/pg-update created
```

**Verify Postgres version updated successfully:**

If everything goes well, `KubeDB` ops-manager operator will update the image of `Postgres`, `PetSet`, and its `Pod`.

At first, we will wait for `PostgresOpsRequest` to be successful.  Run the following command to watch `PostgresOpsRequest` cr,

```bash
$ watch -n 3 kubectl get PostgresOpsRequest -n demo pg-update
Every 3.0s: kubectl get PostgresOpsRequest -n demo pg-update

NAME                         TYPE            STATUS       AGE
pg-update                    UpdateVersion   Successful   3m57s
```

We can see from the above output that the `PostgresOpsRequest` has succeeded. If we describe the `PostgresOpsRequest`, we shall see that the `Postgres`, `PetSet`, and its `Pod` have updated with a new image.

```bash
$ kubectl describe PostgresOpsRequest -n demo pg-update
Name:         pg-update
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         PostgresOpsRequest
Metadata:
  Creation Timestamp:  2021-11-30T07:29:04Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:databaseRef:
          .:
          f:name:
        f:type:
        f:updateVersion:
          .:
          f:targetVersion:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-11-30T07:29:04Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-11-30T07:29:04Z
  Resource Version:  638178
  UID:               d18198d9-0d27-449d-9a1d-edf60c7bdf38
Spec:
  Database Ref:
    Name:  pg
  Type:    UpdateVersion
  UpdateVersion:
    Target Version:  13.2
Status:
  Conditions:
    Last Transition Time:  2021-11-30T07:29:04Z
    Message:               Postgres ops request is update-version database version
    Observed Generation:   1
    Reason:                UpdateVersion
    Status:                True
    Type:                  UpdateVersion
    Last Transition Time:  2021-11-30T07:29:04Z
    Message:               Successfully copied binaries for old postgres version
    Observed Generation:   1
    Reason:                CopiedOldBinaries
    Status:                True
    Type:                  CopiedOldBinaries
    Last Transition Time:  2021-11-30T07:29:04Z
    Message:               Successfully updated petsets update strategy type
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2021-11-30T07:29:10Z
    Message:               Successfully Transferred Leadership to first node before pg-coordinator paused
    Observed Generation:   1
    Reason:                TransferLeaderShipToFirstNodeBeforeCoordinatorPaused
    Status:                True
    Type:                  TransferLeaderShipToFirstNodeBeforeCoordinatorPaused
    Last Transition Time:  2021-11-30T07:29:10Z
    Message:               Successfully Pause Pg-Coordinator
    Observed Generation:   1
    Reason:                PausePgCoordinator
    Status:                True
    Type:                  PausePgCoordinator
    Last Transition Time:  2021-11-30T07:29:50Z
    Message:               Successfully Updated primary Image
    Observed Generation:   1
    Reason:                UpdatePrimaryImage
    Status:                True
    Type:                  UpdatePrimaryImage
    Last Transition Time:  2021-11-30T07:29:52Z
    Message:               Successfully Initialized new data directory
    Observed Generation:   1
    Reason:                DataDirectoryInitialized
    Status:                True
    Type:                  DataDirectoryInitialized
    Last Transition Time:  2021-11-30T07:29:59Z
    Message:               Successfully updated new data directory
    Observed Generation:   1
    Reason:                PgUpdated
    Status:                True
    Type:                  Pgupdated
    Last Transition Time:  2021-11-30T07:29:59Z
    Message:               Successfully Rename new data directory
    Observed Generation:   1
    Reason:                ReplacedDataDirectory
    Status:                True
    Type:                  ReplacedDataDirectory
    Last Transition Time:  2021-11-30T07:30:24Z
    Message:               Successfully Transfer Primary Role to first node
    Observed Generation:   1
    Reason:                TransferPrimaryRoleToDefault
    Status:                True
    Type:                  TransferPrimaryRoleToDefault
    Last Transition Time:  2021-11-30T07:30:29Z
    Message:               Successfully running the primary
    Observed Generation:   1
    Reason:                ResumePgCoordinator
    Status:                True
    Type:                  ResumePgCoordinator
    Last Transition Time:  2021-11-30T07:32:24Z
    Message:               Successfully Updated replica Images
    Observed Generation:   1
    Reason:                UpdateStandbyPodImage
    Status:                True
    Type:                  UpdateStandbyPodImage
    Last Transition Time:  2021-11-30T07:32:24Z
    Message:               Successfully Updated cluster Image
    Observed Generation:   1
    Reason:                UpdatePetSetImage
    Status:                True
    Type:                  UpdatePetSetImage
    Last Transition Time:  2021-11-30T07:32:24Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                                                Age    From                        Message
  ----    ------                                                ----   ----                        -------
  Normal  PauseDatabase                                         3m50s  KubeDB Enterprise Operator  Pausing Postgres demo/pg
  Normal  PauseDatabase                                         3m50s  KubeDB Enterprise Operator  Successfully paused Postgres demo/pg
  Normal  Updating                                              3m50s  KubeDB Enterprise Operator  Updating PetSets
  Normal  Updating                                              3m50s  KubeDB Enterprise Operator  Successfully Updated PetSets
  Normal  TransferLeaderShipToFirstNodeBeforeCoordinatorPaused  3m44s  KubeDB Enterprise Operator  Successfully Transferred Leadership to first node before pg-coordinator paused
  Normal  UpdatePrimaryImage                                    3m4s   KubeDB Enterprise Operator  Successfully Updated primary Image
  Normal  TransferPrimaryRoleToDefault                          2m30s  KubeDB Enterprise Operator  Successfully Transfer Primary Role to first node
  Normal  ResumePgCoordinator                                   2m25s  KubeDB Enterprise Operator  Successfully running the primary
  Normal  UpdateStandbyPodImage                                 30s    KubeDB Enterprise Operator  Successfully Updated replica Images
  Normal  ResumeDatabase                                        30s    KubeDB Enterprise Operator  Resuming PostgreSQL demo/pg
  Normal  ResumeDatabase                                        30s    KubeDB Enterprise Operator  Successfully resumed PostgreSQL demo/pg
  Normal  Successful                                            30s    KubeDB Enterprise Operator  Successfully Updated Database
  Normal  Successful                                            30s    KubeDB Enterprise Operator  Successfully Updated Database

 ```

Now, we are going to verify whether the `Postgres`, `PetSet` and it's `Pod` have updated with new image. Let's check,

```bash
$ kubectl get postgres -n demo pg -o=jsonpath='{.spec.version}{"\n"}'
13.2

$ kubectl get sts -n demo pg -o=jsonpath='{.spec.template.spec.containers[0].image}{"\n"}'
postgres:13.2-alpine

$ kubectl get pod -n demo pg-0 -o=jsonpath='{.spec.containers[0].image}{"\n"}'
postgres:13.2-alpine
```

You can see above that our `Postgres` has been updated with the new version. It verifies that we have successfully updated our Postgres Instance.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete postgres -n demo pg
kubectl delete PostgresOpsRequest -n demo pg-update
```