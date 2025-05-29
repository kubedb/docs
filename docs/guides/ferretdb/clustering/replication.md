---
title: FerretDB Replication Guide
menu:
  docs_{{ .version }}:
    identifier: fr-replication-guide-clustering
    name: Replication Guide
    parent: fr-clustering-ferretdb
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - FerretDB Replication

This tutorial will show you how to use KubeDB to run a FerretDB Replication.

## Before You Begin

Before proceeding:

- Read [ferretdb replication concept](/docs/guides/ferretdb/clustering/replication-concept.md) to learn about FerretDB Replication clustering.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/ferretdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ferretdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy FerretDB Replication

To deploy a MongoDB Replication with both primary and secondary server, user have to specify `spec.server` option in `FerretDB` CRD.

The following is an example of a `FerretDB` object which creates FerretDB with two primary server and two secondary server.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  name: ferret
  namespace: demo
spec:
  version: "2.0.0"
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
  server:
    primary:
      replicas: 2
    secondary:
      replicas: 2
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ferretdb/clustering/fr-replication.yaml
ferretdb.kubedb.com/ferret created
```

Here,

- `spec.server` represents the configuration for primary and secondary server.
- `spec.server.primary.replicas` denotes the number of members in primary server.
- `spec.server.secondary.replicas` denotes the number of members in sercondary server.
- `spec.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. So, each members will have a pod of this storage configuration. You can specify any StorageClass available in your cluster with appropriate resource requests.

KubeDB operator watches for `FerretDB` objects using Kubernetes api. When a `FerretDB` object is created, KubeDB operator will create a new `Postgres` for the backend of this FerretDB object. Then it will create PetSet and a Service with the matching FerretDB object name. This service will always point to the primary of the replication servers. Another secondary service will be created as well for secondary server. KubeDB operator will also create two governing service for PetSets with the name `<ferretdb-name>-pods` and `<ferretdb-name>-secondary-pods` for primary and secondary servers.

```bash
$ kubectl describe fr -n demo ferret
Name:         ferret
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         FerretDB
Metadata:
  Creation Timestamp:  2025-04-10T08:52:39Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  430129
  UID:               4f47b22a-1cc8-4fec-b75d-fc94fcdfdc6c
Spec:
  Auth Secret:
    Name:           ferret-auth
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Server:
    Primary:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  ferretdb
            Resources:
              Limits:
                Memory:  1Gi
              Requests:
                Cpu:     500m
                Memory:  1Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Group:     1000
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        2
    Secondary:
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  ferretdb
            Resources:
              Limits:
                Memory:  1Gi
              Requests:
                Cpu:     500m
                Memory:  1Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Group:     1000
              Run As Non Root:  true
              Run As User:      1000
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  1000
      Replicas:        2
  Ssl Mode:            disabled
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:  1Gi
  Storage Type:   Durable
  Version:        2.0.0
Status:
  Conditions:
    Last Transition Time:  2025-04-10T08:52:39Z
    Message:               The KubeDB operator has started the provisioning of FerretDB: demo/ferret
    Observed Generation:   2
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-04-10T08:54:20Z
    Message:               All replicas are ready for FerretDB demo/ferret
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-04-10T08:54:31Z
    Message:               The FerretDB: demo/ferret is accepting client requests.
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-04-10T08:54:31Z
    Message:               The FerretDB: demo/ferret is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-04-10T08:54:31Z
    Message:               The FerretDB: demo/ferret is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```
Lets see what also created with this FerretDB,

```bash

$ kubectl get fr -n demo
NAME     NAMESPACE   VERSION   STATUS   AGE
ferret   demo        2.0.0     Ready    45m

$ kubectl get pg -n demo
NAME                VERSION           STATUS   AGE
ferret-pg-backend   17.4-documentdb   Ready    45m

$ kubectl get petset -n demo
NAME                AGE
ferret              43m
ferret-pg-backend   44m
ferret-secondary    43m

$ kubectl get svc -n demo
NAME                        TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
ferret                      ClusterIP   10.43.199.240   <none>        27017/TCP                    46m
ferret-pg-backend           ClusterIP   10.43.51.247    <none>        5432/TCP,2379/TCP            46m
ferret-pg-backend-pods      ClusterIP   None            <none>        5432/TCP,2380/TCP,2379/TCP   46m
ferret-pg-backend-standby   ClusterIP   10.43.121.236   <none>        5432/TCP                     46m
ferret-pods                 ClusterIP   None            <none>        27017/TCP                    46m
ferret-secondary            ClusterIP   10.43.15.156    <none>        27017/TCP                    46m
ferret-secondary-pods       ClusterIP   None            <none>        27017/TCP                    46m


$ kubectl get pvc -n demo
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-ferret-pg-backend-0   Bound    pvc-e26fdc37-a7e4-4567-bc2e-d2e5e4e52ad9   1Gi        RWO            longhorn       <unset>                 46m
data-ferret-pg-backend-1   Bound    pvc-470d3d5a-b00e-4100-bb8e-8bad63369f7b   1Gi        RWO            longhorn       <unset>                 45m
data-ferret-pg-backend-2   Bound    pvc-5a02705b-b389-413d-98bb-969b5f0b9128   1Gi        RWO            longhorn       <unset>                 45m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-470d3d5a-b00e-4100-bb8e-8bad63369f7b   1Gi        RWO            Delete           Bound    demo/data-ferret-pg-backend-1   longhorn       <unset>                          46m
pvc-5a02705b-b389-413d-98bb-969b5f0b9128   1Gi        RWO            Delete           Bound    demo/data-ferret-pg-backend-2   longhorn       <unset>                          46m
pvc-e26fdc37-a7e4-4567-bc2e-d2e5e4e52ad9   1Gi        RWO            Delete           Bound    demo/data-ferret-pg-backend-0   longhorn       <unset>                          46m

$ kubectl get secret -n demo
NAME                        TYPE                       DATA   AGE
ferret                      Opaque                     1      48m
ferret-auth                 kubernetes.io/basic-auth   2      46m
ferret-backend-connection   Opaque                     2      46m
ferret-pg-backend-auth      kubernetes.io/basic-auth   2      48m
ferretdb-ca                 kubernetes.io/tls          2      2d4h
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified FerretDB object:

```yaml
$ kubectl get mg -n demo mgo-replicaset -o yaml
apiVersion: kubedb.com/v1alpha2
kind: FerretDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"FerretDB","metadata":{"annotations":{},"name":"ferret","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","server":{"primary":{"replicas":2},"secondary":{"replicas":2}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"version":"2.0.0"}}
  creationTimestamp: "2025-04-10T08:52:39Z"
  finalizers:
    - kubedb.com
  generation: 3
  name: ferret
  namespace: demo
  resourceVersion: "430129"
  uid: 4f47b22a-1cc8-4fec-b75d-fc94fcdfdc6c
spec:
  authSecret:
    name: ferret-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  server:
    primary:
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          containers:
            - name: ferretdb
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
              securityContext:
                allowPrivilegeEscalation: false
                capabilities:
                  drop:
                    - ALL
                runAsGroup: 1000
                runAsNonRoot: true
                runAsUser: 1000
                seccompProfile:
                  type: RuntimeDefault
          podPlacementPolicy:
            name: default
          securityContext:
            fsGroup: 1000
      replicas: 2
    secondary:
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          containers:
            - name: ferretdb
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 1Gi
              securityContext:
                allowPrivilegeEscalation: false
                capabilities:
                  drop:
                    - ALL
                runAsGroup: 1000
                runAsNonRoot: true
                runAsUser: 1000
                seccompProfile:
                  type: RuntimeDefault
          podPlacementPolicy:
            name: default
          securityContext:
            fsGroup: 1000
      replicas: 2
  sslMode: disabled
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  storageType: Durable
  version: 2.0.0
status:
  conditions:
    - lastTransitionTime: "2025-04-10T08:52:39Z"
      message: 'The KubeDB operator has started the provisioning of FerretDB: demo/ferret'
      observedGeneration: 2
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2025-04-10T08:54:20Z"
      message: All replicas are ready for FerretDB demo/ferret
      observedGeneration: 3
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2025-04-10T08:54:31Z"
      message: 'The FerretDB: demo/ferret is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2025-04-10T08:54:31Z"
      message: 'The FerretDB: demo/ferret is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2025-04-10T08:54:31Z"
      message: 'The FerretDB: demo/ferret is successfully provisioned.'
      observedGeneration: 3
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `ferret-auth` *(format: {ferretdb-object-name}-auth)* for storing the password for `postgres` superuser managed by postgres as FerretDB don't manage authentication itself. This secret contains a `username` key which contains the *username* for FerretDB superuser and a `password` key which contains the *password* for FerretDB superuser.

If you want to use custom or existing secret please specify that when creating the FerretDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/ferretdb/concepts/ferretdb.md#specauthsecret).

## Redundancy and Data Availability

Now, you can connect to this database through [mongo-shell](https://www.mongodb.com/try/download/shell) locally). In this tutorial, we will insert document on the primary server, and we will see if the data becomes available on secondary server.

At first, insert data inside primary server.

```bash
$ kubectl get secrets -n demo ferret-auth -o jsonpath='{.data.\username}' | base64 -d
postgres

$ kubectl get secrets -n demo ferret-auth -o jsonpath='{.data.\password}' | base64 -d
7.KCC5Z97cUlOdyt

$ kubectl port-forward svc/ferret -n demo 27017
Forwarding from 127.0.0.1:27017 -> 27017
Forwarding from [::1]:27017 -> 27017
Handling connection for 27017
Handling connection for 27017
```

Now in another terminal

```bash
$ mongosh 'mongodb://postgres:UxV5a35kURSFE(;5@localhost:27017/ferretdb'
Current Mongosh Log ID:	67f793c162ee47b21d6b140a
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.4.2
Using MongoDB:		7.0.77
Using Mongosh:		2.4.2

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/

------
   The server generated these startup warnings when booting
   2025-04-10T09:47:49.075Z: Powered by FerretDB v2.0.0-1-g7fb2c9a8 and DocumentDB 0.102.0 (PostgreSQL 17.4).
   2025-04-10T09:47:49.075Z: Please star 🌟 us on GitHub: https://github.com/FerretDB/FerretDB and https://github.com/microsoft/documentdb.
   2025-04-10T09:47:49.075Z: The telemetry state is undecided. Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.
------

ferretdb> show dbs
kubedb_system  0 B

ferretdb> use newdb
switched to db newdb

newdb> db.movie.insert({"name":"batman"});
{
  acknowledged: true,
  insertedIds: { '0': ObjectId('67f793ef62ee47b21d6b140b') }
}

newdb> db.movie.find().pretty()
[ { _id: ObjectId('67f793ef62ee47b21d6b140b'), name: 'batman' } ]

newdb> exit
```

Now, check the redundancy and data availability in secondary server.

KubeDB create `ferret-secondary` service to connect with the FerretDB secondary server. Basically this service is connected with standby service `ferret-pg-backend-standby` created for backend Postgres by KubeDB operator.

So we have to port-forward the secondary service.

```bash
$ kubectl port-forward svc/ferret-secondary -n demo 27018:27017
Forwarding from 127.0.0.1:27018 -> 27017
Forwarding from [::1]:27018 -> 27017
```

Now in another terminal

```bash
$ mongosh 'mongodb://postgres:UxV5a35kURSFE(;5@localhost:27018/ferretdb'
Current Mongosh Log ID:	67f793c162ee47b21d6b140a
Connecting to:		mongodb://<credentials>@localhost:27017/ferretdb?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.4.2
Using MongoDB:		7.0.77
Using Mongosh:		2.4.2

For mongosh info see: https://www.mongodb.com/docs/mongodb-shell/

------
   The server generated these startup warnings when booting
   2025-04-10T09:47:49.075Z: Powered by FerretDB v2.0.0-1-g7fb2c9a8 and DocumentDB 0.102.0 (PostgreSQL 17.4).
   2025-04-10T09:47:49.075Z: Please star 🌟 us on GitHub: https://github.com/FerretDB/FerretDB and https://github.com/microsoft/documentdb.
   2025-04-10T09:47:49.075Z: The telemetry state is undecided. Read more about FerretDB telemetry and how to opt out at https://beacon.ferretdb.com.
------

ferretdb> show dbs
kubedb_system  0 B

ferretdb> use newdb
switched to db newdb

newdb> db.movie.find().pretty()
[ { _id: ObjectId('67f793ef62ee47b21d6b140b'), name: 'batman' } ]

newdb> exit
```
So we can see data inserted in primary server also exist is secondary server.

### Automatic Failover

FerretDB's backend Postgres is managed by KubeDB. FerretDB just connect with the KubeDB managed backend Postgres Primary and Standby service. So KubeDB managed FerretDB won't have to take any responsibility in case any failover required in backend postgres.
KubeDB will automatically manage the failover process of backend Postgres.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete fr -n demo ferret
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/ferretdb/monitoring/using-prometheus-operator.md).
- Monitor your FerretDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/ferretdb/monitoring/using-builtin-prometheus.md).
- Detail concepts of [FerretDB object](/docs/guides/ferretdb/concepts/ferretdb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).