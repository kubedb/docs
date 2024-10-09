---
title: MSSQLServer CRD
menu:
  docs_{{ .version }}:
    identifier: ms-concepts-mssqlserver
    name: MSSQLServer
    parent: ms-concepts
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MSSQLServer

## What is MSSQLServer

`MSSQLServer` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [Microsoft SQL Server](https://learn.microsoft.com/en-us/sql/sql-server/) in a Kubernetes native way. You only need to describe the desired database configuration in a MSSQLServer object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## MSSQLServer Spec

As with all other Kubernetes objects, a MSSQLServer needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

Below is an example `MSSQLServer` object.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: mssqlserver
  namespace: demo
spec:
  authSecret:
    name: mssqlserver-auth
  configSecret:
    name: mssqlserver-custom-config
  topology:
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
    mode: AvailabilityGroup
  internalAuth:
    endpointCert:
      certificates:
        - alias: endpoint
          secretName: mssqlserver-endpoint-cert
          subject:
            organizationalUnits:
              - endpoint
            organizations:
              - kubedb
      issuerRef:
        apiGroup: cert-manager.io
        kind: Issuer
        name: mssqlserver-ca-issuer
  leaderElection:
    electionTick: 10
    heartbeatTick: 1
    period: 300ms
    transferLeadershipInterval: 1s
    transferLeadershipTimeout: 1m0s
  podTemplate:
    metadata:
      annotations:
        passMe: ToDatabasePod
    controller:
      annotations:
        passMe: ToPetSet
    spec:
      serviceAccountName: my-custom-sa
      schedulerName: my-scheduler
      nodeSelector:
        disktype: ssd
      containers:
        - name: mssql
          resources:
            limits:
              memory: 4Gi
            requests:
              cpu: 500m
              memory: 4Gi
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
        - name: mssql-coordinator
          resources:
            limits:
              memory: 256Mi
            requests:
              cpu: 200m
              memory: 256Mi
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
      podPlacementPolicy:
        name: default
      securityContext:
        fsGroup: 10001
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  serviceTemplates:
    - alias: primary
      metadata:
        annotations:
          passMe: ToService
      spec:
        type: LoadBalancer
    - alias: secondary
      metadata:
        annotations:
          passMe: ToReplicaService
      spec:
        type: NodePort
        ports:
          - name:  http
            port:  1433
  tls:
    certificates:
      - alias: server
        secretName: mssqlserver-server-cert
        subject:
          organizationalUnits:
            - server
          organizations:
            - kubedb
        emailAddresses:
          - dev@appscode.com
      - alias: client
        secretName: mssqlserver-client-cert
        subject:
          organizationalUnits:
            - client
          organizations:
            - kubedb
        emailAddresses:
          - abc@appscode.com
    clientTLS: true
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: mssqlserver-ca-issuer
  healthChecker:
    periodSeconds: 15
    timeoutSeconds: 10
    failureThreshold: 2
    disableWriteCheck: false
  monitor:
    agent: prometheus.io/operator
    prometheus:
      serviceMonitor:
        labels:
          release: prometheus
        interval: 10s
  version: 2022-cu12
  deletionPolicy: Halt
```

### spec.version

`spec.version` is a required field that specifies the name of the [MSSQLServerVersion](/docs/guides/mssqlserver/concepts/catalog.md) crd where the docker images are specified. Currently, when you install KubeDB, it creates the following `MSSQLServerVersion` resources,

```bash
$ kubectl get msversion
NAME        VERSION   DB_IMAGE                                                DEPRECATED   AGE
2022-cu12   2022      mcr.microsoft.com/mssql/server:2022-CU12-ubuntu-22.04                2d
2022-cu14   2022      mcr.microsoft.com/mssql/server:2022-CU14-ubuntu-22.04                2d
```
### spec.replicas

`spec.replicas` specifies the total number of primary and secondary nodes in SQL Server Availability Group cluster configuration. One pod is selected as Primary and others act as secondary replicas. KubeDB uses `PodDisruptionBudget` to ensure that majority of the replicas are available during [voluntary disruptions](https://kubernetes.io/docs/concepts/workloads/pods/disruptions/#voluntary-and-involuntary-disruptions).

To learn more about how to setup a SQL Server Availability Group cluster (HA configuration) in KubeDB, please visit [here](/docs/guides/mssqlserver/clustering/ag_cluster.md).

### spec.leaderElection

There are three fields under MSSQLServer CRD's `spec.leaderElection`. These values define how fast the leader election can happen.

- `leaseDurationSeconds`: This is the duration in seconds that non-leader candidates will wait to force acquire leadership. This is measured against time of last observed ack. Default 15 sec.
- `renewDeadlineSeconds`: This is the duration in seconds that the acting master will retry refreshing leadership before giving up. Normally, LeaseDuration \* 2 / 3. Default 10 sec.
- `retryPeriodSeconds`: This is the duration in seconds the LeaderElector clients should wait between tries of actions. Normally, LeaseDuration / 3. Default 2 sec.

If the Cluster machine is powerful, user can reduce the times. But, Do not make it so little, in that case MSSQLServer will restart very often.




WORK FROM HERE. TEST EVERYTHING...................... 





### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `mssqlserver` database. If not set, KubeDB operator creates a new Secret with name `{mssqlserver-name}-auth` that hold _username_ and _password_ for `mssqlserver` database.

If you want to use an existing or custom secret, please specify that when creating the MSSQLServer object using `spec.authSecret.name`. This Secret should contain superuser _username_ as `username` key and superuser _password_ as `password` key. Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator.

Example:

```bash
$ kubectl create secret generic mssqlserver-auth -n demo \
--from-literal=POSTGRES_USER=not@user \
--from-literal=POSTGRES_PASSWORD=not@secret
secret "mssqlserver-auth" created
```

```bash
$ kubectl get secret -n demo p1-auth -o yaml
apiVersion: v1
data:
  POSTGRES_PASSWORD: bm90QHNlY3JldA==
  POSTGRES_USER: bm90QHVzZXI=
kind: Secret
metadata:
  creationTimestamp: 2018-09-03T11:25:39Z
  name: p1-auth
  namespace: demo
  resourceVersion: "1677"
  selfLink: /api/v1/namespaces/demo/secrets/p1-auth
  uid: 15b3e8a1-af6c-11e8-996d-0800270d7bae
type: Opaque
```

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for database. It can be either `Durable` or `Ephemeral`. The default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create MSSQLServer database using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume. In this case, you don't have to specify `spec.storage` field.

### spec.storage

If you don't set `spec.storageType:` to `Ephemeral` then `spec.storage` field is required. This field specifies the StorageClass of PVCs dynamically allocated to store data for the database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs. PVCs don’t necessarily have to request a class. A PVC with its storageClassName set equal to "" is always interpreted to be requesting a PV with no class, so it can only be bound to PVs with no class (no annotation or one set equal to ""). A PVC with no storageClassName is not quite the same and is treated differently by the cluster depending on whether the DefaultStorageClass admission plugin is turned on.
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs when requesting storage with specific access modes.
- `spec.storage.resources` can be used to request specific quantities of storage. This follows the same resource model used by PVCs.

To learn how to configure `spec.storage`, please visit the links below:

- https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.init

`spec.init` is an optional section that can be used to initialize a newly created MSSQLServer database. MSSQLServer databases can be initialized from these three ways:

1. Initialize from Script
2. Initialize from Snapshot

#### Initialize via Script

To initialize a MSSQLServer database using a script (shell script, db migrator, etc.), set the `spec.init.script` section when creating a MSSQLServer object. `script` must have the following information:

- [VolumeSource](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes): Where your script is loaded from.

Below is an example showing how a script from a configMap can be used to initialize a MSSQLServer database.

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: mssqlserver-db
  namespace: demo
spec:
  version: "13.13"
  init:
    script:
      configMap:
        name: ms-init-script
```

In the above example, MSSQLServer will execute provided script once the database is running. For more details tutorial on how to initialize from script, please visit [here](/docs/guides/mssqlserver/initialization/script_source.md).

### spec.monitor

MSSQLServer managed by KubeDB can be monitored with builtin-Prometheus and Prometheus operator out-of-the-box. To learn more,

- [Monitor MSSQLServer with builtin Prometheus](/docs/guides/mssqlserver/monitoring/using-builtin-prometheus.md)
- [Monitor MSSQLServer with Prometheus operator](/docs/guides/mssqlserver/monitoring/using-prometheus-operator.md)

### spec.configSecret

`spec.configSecret` is an optional field that allows users to provide custom configuration for MSSQLServer. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). You can use any Kubernetes supported volume source such as `configMap`, `secret`, `azureDisk` etc. To learn more about how to use a custom configuration file see [here](/docs/guides/mssqlserver/configuration/using-config-file.md).

### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for MSSQLServer database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata
  - annotations (pod's annotation)
- controller
  - annotations (petset's annotation)
- spec:
  - containers
  - volumes
  - podPlacementPolicy
  - serviceAccountName
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the podPlacementPolicy. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.




#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a DNS_LABEL. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the MSSQLServer docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/_/mssqlserver/).

Note that, the KubeDB operator does not allow `POSTGRES_USER` and `POSTGRES_PASSWORD` environment variable to set in `spec.podTemplate.spec.env`. If you want to set the superuser _username_ and _password_, please use `spec.authSecret` instead described earlier.

If you try to set `POSTGRES_USER` or `POSTGRES_PASSWORD` environment variable in MSSQLServer crd, KubeDB operator will reject the request with following error,

```ini
Error from server (Forbidden): error when creating "./mssqlserver.yaml": admission webhook "mssqlserver.validators.kubedb.com" denied the request: environment variable POSTGRES_PASSWORD is forbidden to use in MSSQLServer spec
```

Also, note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the database is created. If you try to update environment variables, KubeDB operator will reject the request with following error,

```ini
Error from server (BadRequest): error when applying patch:
...
for: "./mssqlserver.yaml": admission webhook "mssqlserver.validators.kubedb.com" denied the request: precondition failed for:
...
At least one of the following was changed:
    apiVersion
    kind
    name
    namespace
    spec.secondary
    spec.streaming
    spec.authSecret
    spec.storageType
    spec.storage
    spec.podTemplate.spec.nodeSelector
    spec.init
```

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](http://kubernetes.io/docs/user-guide/compute-resources/).

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator (version 0.13.0 and higher) that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching MSSQLServer crd name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually. Follow the guide [here](/docs/guides/mssqlserver/custom-rbac/using-custom-rbac.md) to grant necessary permissions in this scenario.

#### spec.podTemplate.spec.imagePullSecrets

`spec.podTemplate.spec.imagePullSecrets` is an optional field that points to secrets to be used for pulling docker image if you are using a private docker registry. For more details on how to use private docker registry, please visit [here](/docs/guides/mssqlserver/private-registry/using-private-registry.md).

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .

### spec.serviceTemplate

KubeDB creates two different services for each MSSQLServer instance. One of them is a master service named `<mssqlserver-name>` and points to the MSSQLServer `Primary` pod/node. Another one is a replica service named `<mssqlserver-name>-replicas` and points to MSSQLServer `replica` pods/nodes.

These `master` and `replica` services can be customized using [spec.serviceTemplate](#spec.servicetemplate) and [spec.replicaServiceTemplate](#specreplicaservicetemplate) respectively.

You can provide template for the `master` service using `spec.serviceTemplate`. This will allow you to set the type and other properties of the service. If `spec.serviceTemplate` is not provided, KubeDB will create a `master` service of type `ClusterIP` with minimal settings.

KubeDB allows following fields to set in `spec.serviceTemplate`:

- metadata:
  - annotations
- spec:
  - type
  - ports
  - clusterIP
  - externalIPs
  - loadBalancerIP
  - loadBalancerSourceRanges
  - externalTrafficPolicy
  - healthCheckNodePort
  - sessionAffinityConfig

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in detail.

### spec.replicaServiceTemplate

You can provide template for the `replica` service using `spec.replicaServiceTemplate`. If `spec.replicaServiceTemplate` is not provided, KubeDB will create a `replica` service of type `ClusterIP` with minimal settings.

The fileds of `spec.replicaServiceTemplate` is similar to `spec.serviceTemplate`, that is:

- metadata:
  - annotations
- spec:
  - type
  - ports
  - clusterIP
  - externalIPs
  - loadBalancerIP
  - loadBalancerSourceRanges
  - externalTrafficPolicy
  - healthCheckNodePort

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.16.3/api/v1/types.go#L163) to understand these fields in detail.

### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MSSQLServer` crd or which resources KubeDB should keep or delete when you delete `MSSQLServer` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to provide safety from accidental deletion of database. If admission webhook is enabled, KubeDB prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete MSSQLServer crd for different termination policies,

| Behavior                                 | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
| ---------------------------------------- | :------------: | :------: | :------: | :------: |
| 1. Block Delete operation                |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Create Dormant Database               |    &#10007;    | &#10003; | &#10007; | &#10007; |
| 3. Delete PetSet                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete Services                       |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 5. Delete PVCs                           |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 6. Delete Secrets                        |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshots                      |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 8. Delete Snapshot data from bucket      |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Halt` termination policy by default.

## Next Steps

- Learn how to use KubeDB to run a MSSQLServer database [here](/docs/guides/mssqlserver/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
