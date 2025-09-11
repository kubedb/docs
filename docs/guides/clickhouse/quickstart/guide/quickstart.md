---
title: ClickHouse Quickstart
menu:
  docs_{{ .version }}:
    identifier: ch-clickhouse-quickstart-clickhouse
    name: ClickHouse
    parent: ch-quickstart-clickhouse
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ClickHouse QuickStart

This tutorial will show you how to use KubeDB to run a ClickHouse database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/clickhouse/quickstart/images/clickhouse-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/guides/clickhouse/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/clickhouse/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

```bash
➤ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  27h
```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available ClickHouseVersion

When you have installed KubeDB, it has created `ClickHouseVersion` crd for all supported ClickHouse versions. Check it by using the following command,

```bash
➤ kubectl get clickhouseversions
NAME     VERSION   DB_IMAGE                              DEPRECATED   AGE
24.4.1   24.4.1    clickhouse/clickhouse-server:24.4.1                27h
25.7.1   25.7.1    clickhouse/clickhouse-server:25.7.1                27h
```

## Create a ClickHouse database

KubeDB implements a `ClickHouse` CRD to define the specification of a ClickHouse database. Below is the `ClickHouse` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: clickhouse-quickstart
  namespace: demo
spec:
  version: 24.4.1
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/quickstart/yamls/quickstart-v1alpha2.yaml
clickhouse.kubedb.com/clickhouse-quickstart created
```

Here,

- `spec.version` is the name of the ClickHouseVersion CRD where the docker images are specified. In this tutorial, a ClickHouse `24.4.1` database is going to be created.
- `spec.storageType` specifies the type of storage that will be used for ClickHouse database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create ClickHouse database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` or `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `ClickHouse` crd or which resources KubeDB should keep or delete when you delete `ClickHouse` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `ClickHouse` objects using Kubernetes api. When a `ClickHouse` object is created, KubeDB operator will create a new PetSet and a Service with the matching ClickHouse object name. KubeDB operator will also create a governing service for PetSets with the name `kubedb`, if one is not already present.

```bash
➤ kubectl describe clickhouse -n demo clickhouse-quickstart 
Name:         clickhouse-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         ClickHouse
Metadata:
  Creation Timestamp:  2025-09-03T10:06:06Z
  Finalizers:
    kubedb.com/clickhouse
  Generation:        3
  Resource Version:  49211
  UID:               44c534e7-3143-4cc3-a332-5bd0a77dbd25
Spec:
  Auth Secret:
    Name:  clickhouse-quickstart-auth
  Auto Ops:
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  clickhouse
        Resources:
          Limits:
            Memory:  4Gi
          Requests:
            Cpu:     1
            Memory:  4Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Non Root:  true
          Run As User:      101
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  clickhouse-init
        Resources:
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Non Root:  true
          Run As User:      101
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:  101
  Replicas:        1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:  1Gi
  Storage Type:   Durable
  Version:        24.4.1
Status:
  Conditions:
    Last Transition Time:  2025-09-03T10:06:06Z
    Message:               The KubeDB operator has started the provisioning of ClickHouse: demo/clickhouse-quickstart
    Observed Generation:   2
    Reason:                ProvisioningStarted
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-09-03T10:06:16Z
    Message:               All desired replicas are ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-09-03T10:06:39Z
    Message:               The Clickhouse: demo/clickhouse-quickstart is accepting client requests
    Observed Generation:   3
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-09-03T10:06:39Z
    Message:               database demo/clickhouse-quickstart is ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-09-03T10:06:39Z
    Message:               The ClickHouse: demo/clickhouse-quickstart is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>




➤ kubectl get petset -n demo clickhouse-quickstart 
NAME                    AGE
clickhouse-quickstart   2m6s

➤ kubectl get pvc -n demo
NAME                           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-clickhouse-quickstart-0   Bound    pvc-811822f0-13a4-4dbf-8d70-845dd8f01a3f   1Gi        RWO            local-path     <unset>                 2m25


➤ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                                               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-811822f0-13a4-4dbf-8d70-845dd8f01a3f   1Gi        RWO            Delete           Bound    demo/data-clickhouse-quickstart-0                                   local-path     <unset>                          5m13s

➤ kubectl get svc -n demo
NAME                         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
clickhouse-quickstart        ClusterIP   10.43.239.202   <none>        9000/TCP,8123/TCP   8m1s
clickhouse-quickstart-pods   ClusterIP   None            <none>        9000/TCP,8123/TCP   8m1s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified ClickHouse object:

```yaml
➤ kubectl get clickhouse -n demo clickhouse-quickstart -oyaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"ClickHouse","metadata":{"annotations":{},"name":"clickhouse-quickstart","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"version":"24.4.1"}}
  creationTimestamp: "2025-09-03T10:06:06Z"
  finalizers:
    - kubedb.com/clickhouse
  generation: 3
  name: clickhouse-quickstart
  namespace: demo
  resourceVersion: "49211"
  uid: 44c534e7-3143-4cc3-a332-5bd0a77dbd25
spec:
  authSecret:
    name: clickhouse-quickstart-auth
  autoOps: {}
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 3
    periodSeconds: 20
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
        - name: clickhouse
          resources:
            limits:
              memory: 4Gi
            requests:
              cpu: "1"
              memory: 4Gi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsNonRoot: true
            runAsUser: 101
            seccompProfile:
              type: RuntimeDefault
      initContainers:
        - name: clickhouse-init
          resources: {}
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsNonRoot: true
            runAsUser: 101
            seccompProfile:
              type: RuntimeDefault
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 101
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  version: 24.4.1
status:
  conditions:
    - lastTransitionTime: "2025-09-03T10:06:06Z"
      message: 'The KubeDB operator has started the provisioning of ClickHouse: demo/clickhouse-quickstart'
      observedGeneration: 2
      reason: ProvisioningStarted
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2025-09-03T10:06:16Z"
      message: All desired replicas are ready
      observedGeneration: 3
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2025-09-03T10:06:39Z"
      message: 'The Clickhouse: demo/clickhouse-quickstart is accepting client requests'
      observedGeneration: 3
      reason: AcceptingConnection
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2025-09-03T10:06:39Z"
      message: database demo/clickhouse-quickstart is ready
      observedGeneration: 3
      reason: AllReplicasReady
      status: "True"
      type: Ready
    - lastTransitionTime: "2025-09-03T10:06:39Z"
      message: 'The ClickHouse: demo/clickhouse-quickstart is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```

## Connect with ClickHouse database

KubeDB operator has created a new Secret called `clickhouse-quickstart-auth` *(format: {clickhouse-object-name}-auth)* for storing the password for `clickhouse` superuser. This secret contains a `username` key which contains the *username* for ClickHouse superuser and a `password` key which contains the *password* for ClickHouse superuser.

If you want to use an existing secret please specify that when creating the ClickHouse object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `clickhouse-quickstart-auth` secret holds username and password

```bash
➤ kubectl get pods clickhouse-quickstart-0 -n demo -o yaml | grep podIP
  podIP: 10.42.0.116

➤ kubectl get secrets -n demo clickhouse-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
admin

$ kubectl get secrets -n demo clickhouse-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
hHWEDm8Cr2PnNJD8
```
we will exec into the pod `clickhouse-quickstart-0` and connect to the database using username and password

```bash
➤ kubectl exec -it -n demo clickhouse-quickstart-0 -- bash
Defaulted container "clickhouse" out of: clickhouse, clickhouse-init (init)
clickhouse@clickhouse-quickstart-0:/$ clickhouse-client -uadmin --password="hHWEDm8Cr2PnNJD8"
ClickHouse client version 24.4.1.2088 (official build).
Connecting to localhost:9000 as user admin.
Connected to ClickHouse server version 24.4.1.

Warnings:
 * Delay accounting is not enabled, OSIOWaitMicroseconds will not be gathered. You can enable it using `echo 1 > /proc/sys/kernel/task_delayacct` or by using sysctl.
 * Effective user of the process (clickhouse) does not match the owner of the data (root).

clickhouse-quickstart-0.clickhouse-quickstart-pods.demo.svc.cluster.local :) show databases

SHOW DATABASES

Query id: 338babd1-90f7-46d2-8f7b-85b83288e2c1

   ┌─name───────────────┐
1. │ INFORMATION_SCHEMA │
2. │ default            │
3. │ information_schema │
4. │ kubedb_system      │
5. │ system             │
   └────────────────────┘

5 rows in set. Elapsed: 0.001 sec.

```
## spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `ClickHouse` crd or which resources KubeDB should keep or delete when you delete `ClickHouse` crd. KubeDB provides following four deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete ClickHouse crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ----------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet               |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

Run the following command to get ClickHouse resources,

```bash
➤ kubectl get ch,sts,secret,svc,pvc -n demo
NAME                                          TYPE                  VERSION   STATUS   AGE
clickhouse.kubedb.com/clickhouse-quickstart   kubedb.com/v1alpha2   24.4.1    Ready    12m

NAME                                  TYPE                       DATA   AGE
secret/clickhouse-quickstart-auth     kubernetes.io/basic-auth   2      12m
secret/clickhouse-quickstart-config   Opaque                     1      12m

NAME                                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/clickhouse-quickstart        ClusterIP   10.43.239.202   <none>        9000/TCP,8123/TCP   12m
service/clickhouse-quickstart-pods   ClusterIP   None            <none>        9000/TCP,8123/TCP   12m

NAME                                                 STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-clickhouse-quickstart-0   Bound    pvc-811822f0-13a4-4dbf-8d70-845dd8f01a3f   1Gi        RWO            local-path     <unset>                 12m
```

From the above output , you can see that `ClickHouse` object, `PVCs`, `Secret` are still alive. Then you can recreate your `ClickHouse` with same configuration.

>When you set `spec.halted` to `true` in `ClickHouse` object then the `deletionPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo clickhouse/clickhouse-quickstart
kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular ClickHouse crd when you delete the crd. For more details about termination policy

## Next Steps


- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).