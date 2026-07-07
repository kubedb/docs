---
title: HanaDB Quickstart
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-quickstart-guide
    name: Quickstart
    parent: guides-hanadb-quickstart
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# HanaDB QuickStart

This tutorial shows you how to run a standalone [SAP HANA](https://www.sap.com/products/technology-platform/hana.html)
database using KubeDB.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/quickstart) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` command-line tool configured to communicate with it.
- Install the KubeDB **Provisioner** and **Ops-manager** operators in your cluster following the steps
  [here](/docs/setup/README.md), with the HanaDB feature and catalog enabled.
- SAP HANA is resource and disk heavy. A standalone instance can take **20–30 minutes** to become
  `Ready`. Make sure your cluster has enough CPU, memory, and storage.

To keep things isolated, this tutorial uses a separate namespace called `demo`:

```bash
kubectl create ns demo
```
namespace/demo created

## Check Available StorageClass

```bash
kubectl get storageclass
```
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  19d

## Find Available HanaDBVersion

KubeDB maintains a `HanaDBVersion` CRD with all supported SAP HANA versions and their images:

```bash
kubectl get hanadbversions
```
NAME     VERSION   DB_IMAGE                                               DEPRECATED   AGE
2.0.76   2.0.76    docker.io/saplabs/hanaexpress:2.00.076.00.20240701.1                31h
2.0.82   2.0.82    docker.io/saplabs/hanaexpress:2.00.082.00.20250528.1                6d13h
2.0.88   2.0.88    docker.io/saplabs/hanaexpress:2.00.088.00.20251110.1                31h

## Create a HanaDB Database

KubeDB implements a `HanaDB` CRD to define the specification of a HANA database. Below is a minimal
standalone `HanaDB` object:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-quickstart
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: Delete
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/quickstart/standalone.yaml
```
hanadb.kubedb.com/hanadb-quickstart created

Here,

- `spec.version` is the name of the `HanaDBVersion` CRD, which specifies the container images.
- `spec.replicas` is the number of database pods. For a standalone database this is `1`.
- `spec.storageType` is `Durable` (uses a PVC) or `Ephemeral` (uses an `emptyDir`).
- `spec.storage` defines the PVC for the HANA data volume.
- `spec.deletionPolicy` controls what happens to data when the `HanaDB` object is deleted.

When a `HanaDB` object is created, the KubeDB operator creates a `PetSet`, a primary `Service`, a
governing (headless) `Service`, an authentication `Secret`, and an `AppBinding` with the matching name.

> Because `hanadb` is also a short name registered by the KubeDB GitOps API, prefer the fully qualified
> resource `hanadb.kubedb.com` (or short name `hdb`) in `kubectl` commands.

## Wait for the Database to be Ready

```bash
kubectl get hanadb.kubedb.com -n demo hanadb-quickstart -w
```
NAME                VERSION   STATUS         AGE
hanadb-quickstart   2.0.82    Provisioning   2m
hanadb-quickstart   2.0.82    Provisioning   18m
hanadb-quickstart   2.0.82    Ready          19m

When `status.phase` becomes `Ready`, the database is ready for traffic. Let's look at the details with
`kubectl describe`:

```bash
kubectl describe hanadb.kubedb.com -n demo hanadb-quickstart
```
Name:         hanadb-quickstart
Namespace:    demo
API Version:  kubedb.com/v1alpha2
Kind:         HanaDB
Metadata:
  Finalizers:
    kubedb.com
Spec:
  Auth Secret:
    Name:           hanadb-quickstart-auth
  Deletion Policy:  Delete
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Pod Template:
    Spec:
      Containers:
        Name:  hanadb
        Resources:
          Limits:
            Cpu:     4
            Memory:  10Gi
          Requests:
            Cpu:     2
            Memory:  8Gi
        Security Context:
          Run As Group:     79
          Run As Non Root:  true
          Run As User:      12000
      Security Context:
        Fs Group:  12000
  Replicas:        1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         64Gi
    Storage Class Name:  local-path
  Storage Type:          Durable
  Version:               2.0.82
Status:
  Conditions:
    Reason:  DatabaseProvisioningStartedSuccessfully
    Status:  True
    Type:    ProvisioningStarted
    Reason:  AllReplicasReady
    Status:  True
    Type:    ReplicaReady
    Reason:  AcceptingConnection
    Status:  True
    Type:    AcceptingConnection
    Reason:  AllReplicasReady
    Status:  True
    Type:    Ready
    Reason:  DatabaseSuccessfullyProvisioned
    Status:  True
    Type:    Provisioned
  Phase:     Ready

Note that KubeDB filled in sensible defaults — for example the default resources on the `hanadb`
container and the `12000:79` security context derived from the `HanaDBVersion`.

## Check Resources Created by KubeDB

```bash
kubectl get hanadb.kubedb.com,pods,pvc,svc -n demo -l app.kubernetes.io/instance=hanadb-quickstart
```
NAME                              VERSION   STATUS   AGE
hanadb.kubedb.com/hanadb-quickstart   2.0.82    Ready    19m

NAME                      READY   STATUS    RESTARTS   AGE
pod/hanadb-quickstart-0   1/1     Running   0          6m30s

NAME                                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-hanadb-quickstart-0   Bound    pvc-83118378-1265-4843-b653-a89568615269   64Gi       RWO            local-path     6m31s

NAME                             TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)               AGE
service/hanadb-quickstart        ClusterIP   10.43.188.77   <none>        39017/TCP             6m33s
service/hanadb-quickstart-pods   ClusterIP   None           <none>        39001/TCP,39017/TCP   6m33s

- `hanadb-quickstart` (port `39017`) is the primary SQL `Service`.
- `hanadb-quickstart-pods` is the governing headless `Service` (nameserver port `39001`, SQL port `39017`).

## Connect to the HanaDB Database

KubeDB stores the `SYSTEM` user credentials in the `hanadb-quickstart-auth` Secret:

```bash
kubectl get secret -n demo hanadb-quickstart-auth -o jsonpath='{.type}'
```
kubernetes.io/basic-auth

```bash
kubectl get secret -n demo hanadb-quickstart-auth -o go-template='{{range $k,$v := .data}}{{$k}}{{"\n"}}{{end}}'
```
password
password.json
username

Read the password into a shell variable (avoid pasting real passwords into shared terminals):

```bash
HANA_PASSWORD="$(kubectl get secret hanadb-quickstart-auth -n demo -o jsonpath='{.data.password}' | base64 -d)"
```

Run a query with `hdbsql` from inside the database pod. Source the HANA environment first:

```bash
kubectl exec -n demo hanadb-quickstart-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$HANA_PASSWORD' 'SELECT 1 AS HELLO FROM DUMMY'"
```
HELLO
1
1 row selected (overall time 3143 usec; server time 158 usec)

List the databases inside the HANA instance:

```bash
kubectl exec -n demo hanadb-quickstart-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$HANA_PASSWORD' \"SELECT DATABASE_NAME, ACTIVE_STATUS FROM SYS.M_DATABASES\""
```
DATABASE_NAME,ACTIVE_STATUS
"SYSTEMDB","YES"
"HXE","YES"
"KUBEDB_HEALTH_CHECK","YES"
3 rows selected

Here, `SYSTEMDB` is the HANA system database, `HXE` is the tenant database, and `KUBEDB_HEALTH_CHECK`
is the tenant database KubeDB uses for its periodic write probe.

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo hanadb.kubedb.com/hanadb-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
```

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-quickstart
```

```bash
kubectl delete ns demo
```

## Next Steps

- Deploy a [System Replication cluster](/docs/guides/hanadb/clustering/system-replication.md).
- Apply [custom configuration](/docs/guides/hanadb/configuration/using-config-file.md).
- Set up [monitoring](/docs/guides/hanadb/monitoring/using-builtin-prometheus.md) and [TLS](/docs/guides/hanadb/tls/overview.md).
- Detailed concepts of the [HanaDB object](/docs/guides/hanadb/concepts/hanadb.md).

> ## ⚠️ Legal Notice
>
> SAP® and SAP HANA® are registered trademarks of SAP SE. KubeDB is not affiliated with, endorsed by,
> or sponsored by SAP SE. KubeDB provides only orchestration and management tooling and does not
> distribute any SAP HANA software or binaries. Users must provide their own SAP HANA container images
> and hold valid SAP licenses.
