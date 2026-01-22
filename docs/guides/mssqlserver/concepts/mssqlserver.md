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
  configuration:
    secretName: mssql-custom-config
  authSecret:
    kind: Secret
    name: mssql-admin-cred
  topology:
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
      leaderElection:
        electionTick: 10
        heartbeatTick: 1
        period: 300ms
        transferLeadershipInterval: 1s
        transferLeadershipTimeout: 1m0s
    mode: AvailabilityGroup
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
  tls:
    certificates:
      - alias: server
        emailAddresses:
          - dev@appscode.com
        secretName: mssqlserver-server-cert
        subject:
          organizationalUnits:
            - server
          organizations:
            - kubedb
      - alias: client
        emailAddresses:
          - abc@appscode.com
        secretName: mssqlserver-client-cert
        subject:
          organizationalUnits:
            - client
          organizations:
            - kubedb
      - alias: endpoint
        secretName: mssqlserver-endpoint-cert
        subject:
          organizationalUnits:
            - endpoint
          organizations:
            - kubedb
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

To learn more about how to set up a SQL Server Availability Group cluster (HA configuration) in KubeDB, please visit [here](/docs/guides/mssqlserver/clustering/ag_cluster.md).

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for `mssqlserver` database. If not set, KubeDB operator creates a new Secret with name `{mssqlserver-name}-auth` that hold _username_ and _password_ for `mssqlserver` database.

If you want to use an existing or custom secret, please specify that when creating the MSSQLServer object using `spec.authSecret.name`. This Secret should contain superuser _username_ as `username` key and superuser _password_ as `password` key. Secrets provided by users are not managed by KubeDB, and therefore, won't be modified or garbage collected by the KubeDB operator.

Example:

```bash
$ kubectl create secret generic mssqlserver-auth -n demo \
             --from-literal=username='sa' \
             --from-literal=password='Pa55w0rd!'
secret/mssqlserver-auth created
```

```bash
$ kubectl get secret -n demo  mssqlserver-auth -oyaml
apiVersion: v1
data:
  password: UGE1NXcwcmQh
  username: c2E=
kind: Secret
metadata:
  creationTimestamp: "2024-10-10T06:47:06Z"
  name: mssqlserver-auth
  namespace: demo
  resourceVersion: "315403"
  uid: dafcce02-b6a2-4e65-bdd1-db6b9b6d4913
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

`spec.init` is an optional section that can be used to initialize a newly created MSSQLServer database. MSSQLServer databases can be initialized from Snapshots.

### spec.monitor

MSSQLServer managed by KubeDB can be monitored with Prometheus operator out-of-the-box.

### spec.configuration

`spec.configuration` is an optional field that allows users to provide custom configuration for MSSQLServer. This field accepts a [`VolumeSource`](https://github.com/kubernetes/api/blob/release-1.11/core/v1/types.go#L47). You can use Kubernetes supported volume source `secret`.

### spec.topology

The spec.topology field is the core of configuring your SQL Server cluster's architecture. It defines the operational mode, high-availability settings, and disaster recovery configurations of the SQL Server cluster. It defines how the cluster should behave, including the databases that should be included in the setup, and the leader election process for managing the primary-secondary roles.
```yaml
spec:
  topology:
    mode: DistributedAG
    availabilityGroup:
      # ... local AG settings ...
    distributedAG:
      # ... DAG settings ...
```
#### spec.topology.mode

The `spec.topology.mode` field determines the mode in which the SQL Server cluster operates. Currently, the supported mode is:  

`AvailabilityGroup`: Configures a standard SQL Server Always On Availability Group within a single Kubernetes cluster. This provides high availability and automatic failover for your databases. In this mode, the KubeDB operator sets up an Availability Group with one primary replica and multiple secondary replicas for high availability. The databases specified in `spec.topology.availabilityGroup.databases` are automatically created and added to the Availability Group. Users do not need to perform these tasks manually.   

`DistributedAG`: Configures a Distributed Availability Group. This mode links two separate AvailabilityGroup clusters, typically in different geographic locations or Kubernetes clusters, to provide a robust disaster recovery solution.



#### spec.topology.availabilityGroup

This section defines the configuration for the local SQL Server Availability Group (AG). It is required for both AvailabilityGroup and DistributedAG modes. It includes details about the databases to be added to the group and the leader election configurations.

##### spec.topology.availabilityGroup.databases

(string[]) An array of database names to be included in the Availability Group. KubeDB will automatically create these databases (if they don't exist) and add them to the AG during cluster initialization. For a DistributedAG in the Secondary role, this field must be empty, as databases will be seeded from the primary site. Users can modify this list later to add databases as needed.

Example:

```yaml
availabilityGroup:
  databases:
    - "sales_db"
    - "inventory_db"
```  
In this example: agdb1 and agdb2 are added to the Availability Group upon cluster setup.

##### spec.topology.availabilityGroup.secondaryAccessMode
(string) Controls how secondary replicas handle incoming connections. Default is Passive.   
We have support for active and passive secondary replicas in Microsoft SQL Server Availability Groups, enabling cost-efficient deployments by supporting passive replicas that avoid licensing costs.

Active/Passive Secondary Replicas:
The secondaryAccessMode field in the MSSQLServer CRD under spec.topology.availabilityGroup allows control over secondary replica connection modes:
- Passive: No client connections (default, ideal for DR or failover without licensing costs).
- ReadOnly: Accepts read-intent connections only.
- All: Allows all read-only connections.

```yaml
spec:
topology:
availabilityGroup:
secondaryAccessMode: Passive | ReadOnly | All
```

T-SQL Mapping:   
- Passive: `SECONDARY_ROLE (ALLOW_CONNECTIONS = NO)`
- ReadOnly: `SECONDARY_ROLE (ALLOW_CONNECTIONS = READ_ONLY)`
- All: `SECONDARY_ROLE (ALLOW_CONNECTIONS = ALL)`


### spec.topology.availabilityGroup.leaderElection

There are five fields under MSSQLServer CRD's `spec.leaderElection`. These values define how fast the leader election can happen.

- `Period`: This is the period between each invocation of `Node.Tick`. It represents the time base for election actions. Default is `100ms`.

- `ElectionTick`: This is the number of `Node.Tick` invocations that must pass between elections. If a follower does not receive any message from the leader during this period, it becomes a candidate and starts an election. It is recommended to set `ElectionTick = 10 * HeartbeatTick` to prevent unnecessary leader switching. Default is `10`.

- `HeartbeatTick`: This defines the interval between heartbeats sent by the leader to maintain its leadership. A leader sends heartbeat messages every `HeartbeatTick` ticks. Default is `1`.

- `TransferLeadershipInterval`: This specifies retry interval to transfer leadership to the healthiest node. Default is `1s`.

- `TransferLeadershipTimeout`: This specifies the  retry timeout for transferring leadership to the healthiest node. Default is `60s`.

You can increase the period and the electionTick if the system has high network latency.


### spec.topology.distributedAG
This section is required when spec.topology.mode is set to DistributedAG. It defines the configuration for the Distributed Availability Group.   

`spec.topology.distributedAG.self`
This object defines the configuration for the local Availability Group's participation in the DAG.
- role: (string) Specifies whether this local AG is the Primary or Secondary in the Distributed AG.
- url: (string) The listener endpoint URL of this local AG (e.g., a LoadBalancer IP and port). This URL must be reachable from the remote site.   

`spec.topology.distributedAG.remote`   
This object defines the connection details for the remote Availability Group that this cluster will connect to.
- name: (string) The actual name of the Availability Group on the remote cluster.
- url: (string) The listener endpoint URL of the remote AG. This URL must be reachable from the SQL Server instances in the local cluster.


### spec.podTemplate

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for MSSQLServer.

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
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

You can check out the full list [here](https://github.com/kmodules/offshoot-api/blob/master/api/v2/types.go#L26C1-L279C1).
Uses of some field of `spec.podTemplate` is described below,

#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field. This can be used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field. This can be used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field. This can be used to provide the reference of the `podPlacementPolicy`. `name` of the podPlacementPolicy is referred under this attribute. This will be used by our Petset controller to place the db pods throughout the region, zone & nodes according to the policy. It utilizes kubernetes affinity & podTopologySpreadContraints feature to do so.
```yaml
spec:
  podPlacementPolicy:
    name: default
```

#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list of containers and their configurations for to the database pod. some of the fields are described below,

##### spec.podTemplate.spec.containers[].name
The `spec.podTemplate.spec.containers[].name` field used to specify the name of the container specified as a `DNS_LABEL`. Each container in a pod must have a unique name (DNS_LABEL). Cannot be updated.

##### spec.podTemplate.spec.containers[].args
`spec.podTemplate.spec.containers[].args` is an optional field. This can be used to provide additional arguments to database installation.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies the environment variables to pass to the MSSQLServer docker image. To know about supported environment variables, please visit [here](https://hub.docker.com/r/microsoft/mssql-server).

Note that, the KubeDB operator does not allow `MSSQL_SA_USERNAME` and `MSSQL_SA_PASSWORD` environment variable to set in `spec.podTemplate.spec.env`. If you want to set the superuser _username_ and _password_, please use `spec.authSecret` instead described earlier.

If you try to set `MSSQL_SA_USERNAME` or `MSSQL_SA_PASSWORD` environment variable in MSSQLServer CR, KubeDB operator will reject the request with following error,

```ini
The MSSQLServer "mssqlserver" is invalid: spec.podTemplate: Invalid value: "mssqlserver": environment variable MSSQL_SA_PASSWORD is forbidden to use in MSSQLServer spec
```

Also, note that KubeDB does not allow to update the environment variables as updating them does not have any effect once the database is created.

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field. This can be used to request compute resources required by containers of the database pods. To learn more, visit [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/).

#### spec.podTemplate.spec.serviceAccountName

`serviceAccountName` is an optional field supported by KubeDB Operator that can be used to specify a custom service account to fine tune role based access control.

If this field is left empty, the KubeDB operator will create a service account name matching MSSQLServer CR name. Role and RoleBinding that provide necessary access permissions will also be generated automatically for this service account.

If a service account name is given, but there's no existing service account by that name, the KubeDB operator will create one, and Role and RoleBinding that provide necessary access permissions will also be generated for this service account.

If a service account name is given, and there's an existing service account by that name, the KubeDB operator will use that existing service account. Since this service account is not managed by KubeDB, users are responsible for providing necessary access permissions manually.


#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). To learn more, see [here](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#nodeselector) .


### spec.tls

`spec.tls` specifies the TLS/SSL configurations for the MSSQLServer. KubeDB uses [cert-manager](https://cert-manager.io/) v1 api to provision and manage TLS certificates.

The following fields are configurable in the `spec.tls` section:

- `issuerRef` is a reference to the `Issuer` or `ClusterIssuer` CR of [cert-manager](https://cert-manager.io/docs/concepts/issuer/) that will be used by `KubeDB` to generate necessary certificates.

  - `apiGroup` is the group name of the resource that is being referenced. Currently, the only supported value is `cert-manager.io`.
  - `kind` is the type of resource that is being referenced. KubeDB supports both `Issuer` and `ClusterIssuer` as values for this field.
  - `name` is the name of the resource (`Issuer` or `ClusterIssuer`) being referenced.


- `clientTLS` This setting determines whether TLS (Transport Layer Security) is enabled for the MS SQL Server.   
  - If set to `true`, the sql server will be provisioned with `TLS`, and you will need to install the [csi-driver-cacerts](https://github.com/kubeops/csi-driver-cacerts) which will be used to add self-signed ca certificates to the OS trusted certificate store (/etc/ssl/certs/ca-certificates.crt).
  - If set to `false`, TLS will not be enabled for SQL Server. However, the Issuer will still be used to configure a TLS-enabled WAL-G proxy server, which is necessary for performing SQL Server backup operations.
  

- `certificates` (optional) are a list of certificates used to configure the server and/or client certificate. It has the following fields:
  - `alias` represents the identifier of the certificate. It has the following possible value:
    - `server` is used for server certificate identification.
    - `client` is used for client certificate identification.
    - `endpoint`: For endpoint certificate identification
    - `exporter` is used for metrics exporter certificate identification.
  - `secretName` (optional) specifies the k8s secret name that holds the certificates.
 This field is optional. If the user does not specify this field, the default secret name will be created in the following format: `<database-name>-<cert-alias>-cert`.

  - `subject` (optional) specifies an `X.509` distinguished name. It has the following possible field,
      - `organizations` (optional) are the list of different organization names to be used on the Certificate.
      - `organizationalUnits` (optional) are the list of different organization unit name to be used on the Certificate.
      - `countries` (optional) are the list of country names to be used on the Certificate.
      - `localities` (optional) are the list of locality names to be used on the Certificate.
      - `provinces` (optional) are the list of province names to be used on the Certificate.
      - `streetAddresses` (optional) are the list of a street address to be used on the Certificate.
      - `postalCodes` (optional) are the list of postal code to be used on the Certificate.
      - `serialNumber` (optional) is a serial number to be used on the Certificate.
        You can find more details from [Here](https://golang.org/pkg/crypto/x509/pkix/#Name)
    - `duration` (optional) is the period during which the certificate is valid.
    - `renewBefore` (optional) is a specifiable time before expiration duration.
    - `dnsNames` (optional) is a list of subject alt names to be used in the Certificate.
    - `ipAddresses` (optional) is a list of IP addresses to be used in the Certificate.
    - `uris` (optional) is a list of URI Subject Alternative Names to be set in the Certificate.
    - `emailAddresses` (optional) is a list of email Subject Alternative Names to be set in the Certificate.
    - `privateKey` (optional) specifies options to control private keys used for the Certificate.
      - `encoding` (optional) is the private key cryptography standards (PKCS) encoding for this certificate's private key to be encoded in. If provided, allowed values are "pkcs1" and "pkcs8" standing for PKCS#1 and PKCS#8, respectively. It defaults to PKCS#1 if not specified.
 

### spec.serviceTemplates

KubeDB creates two different services for each MSSQLServer instance. One of them is a primary service named `<mssqlserver-name>` and points to the MSSQLServer `Primary` pod/node. Another one is a secondary service named `<mssqlserver-name>-secondary` and points to MSSQLServer `secondary` replica pods/nodes.

These `primary` and `secondary` services can be customized using [spec.serviceTemplates](#spec.servicetemplate).

You can provide template for the services using `spec.serviceTemplates`. This will allow you to set the type and other properties of the service. If `spec.serviceTemplates` is not provided, KubeDB will create a `primary` service of type `ClusterIP` with minimal settings.

KubeDB allows following fields to set in `spec.serviceTemplates`:
- `alias` represents the identifier of the service. It has the following possible value:
  - `primary` is used for the primary service identification.
  - `secondary` is used for the secondary service identification.
  - `stats` is used for the exporter service identification.
- metadata:
  - labels
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

See [here](https://github.com/kmodules/offshoot-api/blob/kubernetes-1.21.1/api/v1/types.go#L237) to understand these fields in detail.

## spec.healthChecker
It defines the attributes for the health checker.
- `spec.healthChecker.periodSeconds` specifies how often to perform the health check.
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the probe times out.
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the healthChecker to be considered failed.
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the writeCheck or not.

Know details about KubeDB Health checking from this [blog post](https://appscode.com/blog/post/kubedb-health-checker/).


### spec.deletionPolicy

`deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `MSSQLServer` crd or which resources KubeDB should keep or delete when you delete `MSSQLServer` crd. KubeDB provides following four termination policies:

- DoNotTerminate
- Halt
- Delete (`Default`)
- WipeOut

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, `DoNotTerminate` prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`.

Following table show what KubeDB does when you delete MSSQLServer crd for different termination policies,

| Behavior                            | DoNotTerminate |  Halt   |  Delete  | WipeOut  |
|-------------------------------------| :------------: | :------: | :------: | :------: |
| 1. Block Delete operation           |    &#10003;    | &#10007; | &#10007; | &#10007; |
| 2. Delete PetSet                    |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 3. Delete Services                  |    &#10007;    | &#10003; | &#10003; | &#10003; |
| 4. Delete PVCs                      |    &#10007;    | &#10007; | &#10003; | &#10003; |
| 5. Delete Secrets                   |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 6. Delete Snapshots                 |    &#10007;    | &#10007; | &#10007; | &#10003; |
| 7. Delete Snapshot data from bucket |    &#10007;    | &#10007; | &#10007; | &#10003; |

If you don't specify `spec.deletionPolicy` KubeDB uses `Delete` termination policy by default.

> For more details you can visit [here](https://appscode.com/blog/post/deletion-policy/)

### spec.halted
Indicates that the database is halted and all offshoot Kubernetes resources except PVCs are deleted.

### Configuring Environment Variables for SQL Server on Linux
You can use environment variables to configure SQL Server on Linux containers.
When deploying `Microsoft SQL Server` on Linux using `containers`, you need to specify the `product edition` through the [MSSQL_PID](https://mcr.microsoft.com/en-us/product/mssql/server/about#configuration:~:text=MSSQL_PID%20is%20the,documentation%20here.) environment variable. This variable determines which `SQL Server edition` will run inside the container. The acceptable values for `MSSQL_PID` are:   
`Developer`: This will run the container using the Developer Edition (this is the default if no MSSQL_PID environment variable is supplied)    
`Express`: This will run the container using the Express Edition    
`Standard`: This will run the container using the Standard Edition   
`Enterprise`: This will run the container using the Enterprise Edition   
`EnterpriseCore`: This will run the container using the Enterprise Edition Core   
`<valid product id>`: This will run the container with the edition that is associated with the PID

`ACCEPT_EULA` confirms your acceptance of the [End-User Licensing Agreement](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-environment-variables?view=sql-server-ver17#:~:text=ACCEPT_EULA,SQL%20Server%20image).
For a complete list of environment variables that can be used, refer to the documentation [here](https://learn.microsoft.com/en-us/sql/linux/sql-server-linux-configure-environment-variables?view=sql-server-2017).

Below is an example of how to configure the `MSSQL_PID` and `ACCEPT_EULA` environment variable in the KubeDB MSSQLServer Custom Resource Definition (CRD):
```bash
metadata:
  name: mssqlserver
  namespace: demo
spec:
  podTemplate:
    spec:
      containers:
      - name: mssql
        env:
        - name: ACCEPT_EULA
          value: "Y"
        - name: MSSQL_PID
          value: Enterprise
```
In this example, the SQL Server container will run the Enterprise Edition.

## Next Steps

- Learn how to use KubeDB to run a MSSQLServer database [here](/docs/guides/mssqlserver/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
