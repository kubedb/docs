---
title: MS SQL Server Quickstart
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-quickstart
    name: Quickstart
    parent: guides-mssqlserver
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MS SQL Server QuickStart

This tutorial will show you how to use KubeDB to run an MS SQL Server database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mysql/quickstart/images/mysql-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/guides/mssqlserver/quickstart/](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mssqlserver/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  5d20h
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available MSSQLServerVersion

When you have installed KubeDB, it has created `MSSQLServerVersion` crd for all supported MSSQLServer versions. Check it by using the following command,

```bash
$ kubectl get mssqlserverversions
NAME        VERSION   DB_IMAGE                                                DEPRECATED   AGE
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                2d3h

```

## Create a MSSQLServer database

KubeDB implements a `MSSQLServer` CRD to define the specification of a MSSQLServer database. Below is the `MSSQLServer` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver-quickstart
  namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: Delete

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/quickstart/mssqlserver-quickstart.yaml
mssqlserver.kubedb.com/mssqlserver-quickstart created
```

Here,

- `spec.version` is the name of the MSSQLServerVersion CRD where the docker images are specified. In this tutorial, a MSSQLServer `2022` database is going to be created.
- `spec.storageType` specifies the type of storage that will be used for MSSQLServer database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MSSQLServer database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MSSQLServer` crd or which resources KubeDB should keep or delete when you delete `MSSQLServer` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified in storage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `MSSQLServer` objects using Kubernetes api. When a `MSSQLServer` object is created, KubeDB operator will create a new PetSet and a Service with the matching MSSQLServer object name. KubeDB operator will also create a governing service for PetSets with the name `<MSSQLServerName>-pods`, if one is not already present.

```bash
$ kubectl dba describe ms -n demo mssqlserver-quickstart 
Name:         mssqlserver-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         MSSQLServer
Metadata:
  Creation Timestamp:  2024-05-02T13:42:30Z
  Finalizers:
    kubedb.com
  Generation:        2
  Resource Version:  191795
  UID:               af908d5e-31ba-4ac5-9d9b-b49f697fceab
Spec:
  Auth Secret:
    Name:  mssqlserver-quickstart-auth
  Coordinator:
    Resources:
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Placement Policy:
    Name:  default
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  mssql
        Resources:
          Limits:
            Memory:  1536Mi
          Requests:
            Cpu:     500m
            Memory:  1536Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Add:
              NET_BIND_SERVICE
            Drop:
              ALL
          Run As Group:     10001
          Run As Non Root:  true
          Run As User:      10001
          Seccomp Profile:
            Type:  RuntimeDefault
      Init Containers:
        Name:  mssql-init
        Resources:
          Limits:
            Memory:  512Mi
          Requests:
            Cpu:     200m
            Memory:  512Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     10001
          Run As Non Root:  true
          Run As User:      10001
          Seccomp Profile:
            Type:  RuntimeDefault
      Security Context:
        Fs Group:  10001
  Replicas:        1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Termination Policy:    Delete
  Version:               2022-cu12
Status:
  Conditions:
    Last Transition Time:  2024-05-02T13:42:30Z
    Message:               The KubeDB operator has started the provisioning of MSSQL: demo/mssqlserver-quickstart
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-05-02T13:42:50Z
    Message:               All replicas are ready for MSSQL demo/mssqlserver-quickstart
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-05-02T13:43:10Z
    Message:               database demo/mssqlserver-quickstart is accepting connection
    Observed Generation:   2
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-05-02T13:43:10Z
    Message:               database demo/mssqlserver-quickstart is ready
    Observed Generation:   2
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-05-02T13:43:10Z
    Message:               The MSSQL: demo/mssqlserver-quickstart is successfully provisioned.
    Observed Generation:   2
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready


$ kubectl get petset -n demo mssqlserver-quickstart
NAME                     AGE
mssqlserver-quickstart   13m


$ kubectl get pvc -n demo
NAME                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-mssqlserver-quickstart-0   Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            standard       15m


$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                STORAGECLASS   REASON   AGE
pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            Delete           Bound    demo/data-mssqlserver-quickstart-0   standard                15m


kubectl get service -n demo
NAME                          TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
mssqlserver-quickstart        ClusterIP   10.96.128.61   <none>        1433/TCP   15m
mssqlserver-quickstart-pods   ClusterIP   None           <none>        1433/TCP   15m

```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MSSQLServer object:

```yaml
$ kubectl get ms -n demo mssqlserver-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MSSQLServer","metadata":{"annotations":{},"name":"mssqlserver-quickstart","namespace":"demo"},"spec":{"replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"Delete","version":"2022-cu12"}}
  creationTimestamp: "2024-05-02T13:42:30Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: mssqlserver-quickstart
  namespace: demo
  resourceVersion: "191795"
  uid: af908d5e-31ba-4ac5-9d9b-b49f697fceab
spec:
  authSecret:
    name: mssqlserver-quickstart-auth
  coordinator:
    resources: {}
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podPlacementPolicy:
    name: default
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
        - name: mssql
          resources:
            limits:
              memory: 1536Mi
            requests:
              cpu: 500m
              memory: 1536Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              add:
                - NET_BIND_SERVICE
              drop:
                - ALL
            runAsGroup: 10001
            runAsNonRoot: true
            runAsUser: 10001
            seccompProfile:
              type: RuntimeDefault
      initContainers:
        - name: mssql-init
          resources:
            limits:
              memory: 512Mi
            requests:
              cpu: 200m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
            runAsGroup: 10001
            runAsNonRoot: true
            runAsUser: 10001
            seccompProfile:
              type: RuntimeDefault
      securityContext:
        fsGroup: 10001
  replicas: 1
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: Delete
  version: 2022-cu12
status:
  conditions:
    - lastTransitionTime: "2024-05-02T13:42:30Z"
      message: 'The KubeDB operator has started the provisioning of MSSQL: demo/mssqlserver-quickstart'
      observedGeneration: 1
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2024-05-02T13:42:50Z"
      message: All replicas are ready for MSSQL demo/mssqlserver-quickstart
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2024-05-02T13:43:10Z"
      message: database demo/mssqlserver-quickstart is accepting connection
      observedGeneration: 2
      reason: AcceptingConnection
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2024-05-02T13:43:10Z"
      message: database demo/mssqlserver-quickstart is ready
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: Ready
    - lastTransitionTime: "2024-05-02T13:43:10Z"
      message: 'The MSSQL: demo/mssqlserver-quickstart is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready

```

## Connect with MSSQLServer database

KubeDB operator has created a new Secret called `mssqlserver-quickstart-auth` *(format: {mssqlserver-object-name}-auth)* for storing the sa password for `mssqlserver`. This secret contains a `username` key which contains the *username* for MSSQLServer SA and a `password` key which contains the *password* for MSSQLServer SA user.

If you want to use an existing secret please specify that when creating the MSSQLServer object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `sa` as value of `username` and a strong password for the sa user. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `mssqlserver-quickstart-auth` secret holds username and password

```bash
$ kubectl get pods -n demo mssqlserver-quickstart-0 -oyaml | grep podIP
  podIP: 10.244.0.168

$ kubectl get secret -n demo mssqlserver-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
sa

$ kubectl get secret -n demo mssqlserver-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
axgXHj4oRIVQ1ocK
```
we will exec into the pod `mysql-quickstart-0` and connect to the database using username and password

```bash
$ kubectl exec -it -n demo mssqlserver-quickstart-0 -- bash
Defaulted container "mssql" out of: mssql, mssql-init (init)
mssql@mssqlserver-quickstart-0:/$ /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P "axgXHj4oRIVQ1ocK"
1> select name from sys.databases
2> go
name                                                                                                                            
--------------------------------------------------------------------------------------------------------------------------------
master                                                                                                                          
tempdb                                                                                                                          
model                                                                                                                           
msdb                                                                                                                            
kubedb_system                                                                                                                   

(5 rows affected)
1> 


```
You can also connect with database management tools like SSMS
You can also use the external-ip of the service. You can also port forward your service to connect.




## Database TerminationPolicy

This field is used to regulate the deletion process of the related resources when `MSSQLServer` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `terminationPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete ms -n demo mssqlserver-quickstart
Error from server (BadRequest): admission webhook "mysql.validators.kubedb.com" denied the request: mysql "mysql-quickstart" can't be halted. To delete, change spec.terminationPolicy
```

Now, run `kubectl edit ms -n demo mssqlserver-quickstart` to set `spec.terminationPolicy` to `Halt` (which deletes the mssqlserver object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `TerminationPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `MSSQLServer` object `terminationPolicy` to `Halt`.

When the [TerminationPolicy](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy) is set to `halt` and the MSSQLServer object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `terminationPolicy` to `halt` in existing database using `edit` command for testing.

At first, run `kubectl edit ms -n demo mssqlserver-quickstart` to set `spec.terminationPolicy` to `Halt`. Then delete the mysql object,

```bash
$ kubectl delete ms -n demo mssqlserver-quickstart
mssqlserver.kubedb.com "mssqlserver-quickstart" deleted
```

Now, run the following command to get mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get petset,svc,secret,pvc -n demo 
NAME                                 TYPE                       DATA   AGE
secret/dbm-login-secret              kubernetes.io/basic-auth   1      56m
secret/master-key-secret             kubernetes.io/basic-auth   1      56m
secret/mssqlserver-quickstart-auth   kubernetes.io/basic-auth   2      56m

NAME                                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mssqlserver-quickstart-0   Bound    pvc-0e6a361e-9195-4d6b-8042-e90ec98d8288   1Gi        RWO            standard       4m17s

```

From the above output, you can see that all mssqlserver resources(`PetSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your mssqlserver again using these resources.

>You can also set the `terminationPolicy` to `Halt`(deprecated). It's behavior same as `halt` and right now `halt` is replaced by `Halt`.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `MSSQLServer` object `terminationPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [TerminationPolicy](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy) is set to `Delete` and the MSSQLServer object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `terminationPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete ms -n demo mssqlserver-quickstart 
mssqlserver.kubedb.com "mssqlserver-quickstart" deleted
```

Now, run the following command to get all mssql resources in `demo` namespaces,

```bash
$ kubectl get petset,svc,secret,pvc -n demo
NAME                                 TYPE                       DATA   AGE
secret/mssqlserver-quickstart-auth   kubernetes.io/basic-auth   2      58m

```

From the above output, you can see that all mssqlserver resources(`PetSet`, `Service`, `PVCs` etc.) are deleted except `Secret`. You can initialize your mssqlserver using `snapshots`(if previously taken) and `secret`.

>If you don't set the terminationPolicy then the kubeDB set the TerminationPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `MSSQLServer` database and relevant resources without any tracking by setting `terminationPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `MSSQLServer` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `terminationPolicy` is set to `WipeOut`.

Suppose, we have a database with `terminationPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete ms -n demo mssqlserver-quickstart
mssqlserver.kubedb.com "mssqlserver-quickstart" deleted
```

Now, run the following command to get all mssqlserver resources in `demo` namespaces,

```bash
$ kubectl get petset,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all mssqlserver resources are deleted. there is no option to recreate/reinitialize your database if `terminationPolicy` is set to `WipeOut`.

>Be careful when you set the `terminationPolicy` to `WipeOut`. Because there is no option to trace the database resources if once deleted the database.

## Database Halted

If you want to delete MSSQLServer resources(`PetSet`,`Service`, etc.) without deleting the `MSSQLServer` object, `PVCs` and `Secret` you have to set the `spec.halted` to `true`. KubeDB operator will be able to delete the MSSQLServer related resources except the `MSSQLServer` object, `PVCs` and `Secret`.

Suppose we have a database running `mssqlserver-quickstart` in our cluster. Now, we are going to set `spec.halted` to `true` in `MSSQLServer`  object by running `kubectl edit -n demo mssqlserver-quickstart` command.

Run the following command to get MSSQLServer resources,

```bash
$ kubectl get ms,petset,svc,secret,pvc -n demo
NAME                                            VERSION     STATUS   AGE
mssqlserver.kubedb.com/mssqlserver-quickstart   2022-cu12   Ready    12m

NAME                                 TYPE                       DATA   AGE
secret/dbm-login-secret              kubernetes.io/basic-auth   1      64m
secret/master-key-secret             kubernetes.io/basic-auth   1      64m
secret/mssqlserver-quickstart-auth   kubernetes.io/basic-auth   2      64m

NAME                                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-mssqlserver-quickstart-0   Bound    pvc-0e6a361e-9195-4d6b-8042-e90ec98d8288   1Gi        RWO            standard       12m
```

From the above output , you can see that `MSSQLServer` object, `PVCs`, `Secret` are still alive. Then you can recreate your `MSSQLServer` with same configuration.

>When you set `spec.halted` to `true` in `MSSQLServer` object then the `terminationPolicy` is also set to `Halt` by KubeDB operator.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mssqlserver/mssqlserver-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete mssqlserver -n demo mssqlserver-quickstart 


kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular MSSQLServer crd when you delete the crd. For more details about termination policy, please visit [here](/docs/guides/mysql/concepts/database/index.md#specterminationpolicy).

## Next Steps

- Initialize [MSSQLServer with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MSSQLServer database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Detail concepts of [MSSQLServer object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MSSQLServerVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
