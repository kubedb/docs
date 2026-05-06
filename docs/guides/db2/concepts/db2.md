---
title: DB2 CRD
menu:
  docs_{{ .version }}:
    identifier: db2-db2-concepts
    name: DB2
    parent: db2-concepts-db2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DB2

## What is DB2

`DB2` is a Kubernetes `CustomResourceDefinition` (CRD) provided by KubeDB. It lets you run and manage IBM DB2 with Kubernetes-native declarative APIs.

## DB2 Spec

Like all Kubernetes resources, a `DB2` object needs `apiVersion`, `kind`, and `metadata`, and it uses `.spec` to define the desired state.

Below is an example DB2 object with all optional fields.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: db2
  namespace: demo
spec:
  version: "11.5.8.0"
  replicas: 1
  authSecret:
    name: db2-auth
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
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
      imagePullSecrets:
      - name: myregistrykey
      containers:
      - name: db2
        env:
        - name: DB2INSTANCE
          value: db2inst1
        resources:
          requests:
            memory: "256Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "2"
      tolerations:
      - key: dedicated
        operator: Equal
        value: database
        effect: NoSchedule
      volumes:
      - name: custom-volume
        emptyDir: {}
      podPlacementPolicy:
        name: default
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: ClusterIP
      ports:
      - name: db2
        port: 50000
  deletionPolicy: Delete
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
    disableWriteCheck: false
```

### spec.version

`spec.version` is a required field that specifies the name of the [DB2Version](/docs/guides/db2/concepts/catalog.md) CRD where the docker images are specified.

### spec.replicas

`spec.replicas` specifies the number of DB2 replicas (pods) to be created for the DB2 instance. For standalone deployment, this is typically set to 1.

### spec.authSecret

`spec.authSecret` is an optional field that points to a Secret used to hold credentials for DB2 user. If not set, KubeDB operator creates a new Secret named `{db2-object-name}-auth` for storing the password.

We can use this field in 3 modes:

1. Using an external secret:
```yaml
authSecret:
  name: <your-created-auth-secret-name>
  externallyManaged: true
```

2. Specifying the secret name only:
```yaml
authSecret:
  name: <intended-auth-secret-name>
```

3. Let KubeDB do everything (omit this field).

The auth secret should contain `username` and `password` keys with base64-encoded values.

### spec.storageType

`spec.storageType` is an optional field that specifies the type of storage to use for the database. It can be either `Durable` or `Ephemeral`. The default value is `Durable`. If `Ephemeral` is used, KubeDB will create the DB2 instance using [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) volume.

### spec.storage

If `spec.storageType` is set to `Durable`, then `spec.storage` field is required. This field specifies the StorageClass and PVC configuration for persistent storage.

- `spec.storage.storageClassName` is the name of the StorageClass used to provision PVCs
- `spec.storage.accessModes` uses the same conventions as Kubernetes PVCs (typically ReadWriteOnce)
- `spec.storage.resources` specifies the storage capacity requirements

To learn how to configure `spec.storage`, visit: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#persistentvolumeclaims

### spec.podTemplate

KubeDB allows providing a template for DB2 pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for the DB2 database.

KubeDB accepts following fields to set in `spec.podTemplate`:

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

#### spec.podTemplate.spec.tolerations

The `spec.podTemplate.spec.tolerations` is an optional field used to specify the pod's tolerations.

#### spec.podTemplate.spec.volumes

The `spec.podTemplate.spec.volumes` is an optional field used to provide the list of volumes that can be mounted by containers belonging to the pod.

#### spec.podTemplate.spec.podPlacementPolicy

`spec.podTemplate.spec.podPlacementPolicy` is an optional field used to provide the reference of the `podPlacementPolicy`. The `name` of the podPlacementPolicy is referred under this attribute. This will be used by the PetSet controller to place the DB2 pods throughout the region, zone & nodes according to the policy.

```yaml
spec:
  podPlacementPolicy:
    name: default
```

#### spec.podTemplate.spec.nodeSelector

`spec.podTemplate.spec.nodeSelector` is an optional field that specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels.

#### spec.podTemplate.spec.containers

The `spec.podTemplate.spec.containers` can be used to provide the list of containers and their configurations for the database pod.

##### spec.podTemplate.spec.containers[].name

The `spec.podTemplate.spec.containers[].name` field specifies the name of the container as a DNS_LABEL. Each container in a pod must have a unique name.

##### spec.podTemplate.spec.containers[].env

`spec.podTemplate.spec.containers[].env` is an optional field that specifies environment variables to pass to the DB2 containers.

##### spec.podTemplate.spec.containers[].resources

`spec.podTemplate.spec.containers[].resources` is an optional field used to request compute resources (CPU and memory) required by containers of the database pods.

### spec.serviceTemplates

You can provide templates for the services created by KubeDB operator for DB2 through `spec.serviceTemplates`. This allows you to set the type and other properties of the services.

KubeDB allows following fields to set in `spec.serviceTemplates`:

- `alias` represents the identifier of the service (typically `primary`)
- metadata:
    - labels
    - annotations
- spec:
    - type
    - ports
    - clusterIP
    - externalIPs
    - loadBalancerIP
    - externalTrafficPolicy

### spec.deletionPolicy

`deletionPolicy` gives flexibility to control what happens when you delete a DB2 CRD. KubeDB provides following deletion policies:

- DoNotTerminate
- WipeOut
- Halt
- Delete

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes to prevent deletion of the database as long as the policy is set to `DoNotTerminate`.

### spec.healthChecker

`spec.healthChecker` defines the attributes for the health checker that monitors the DB2 instance.

- `spec.healthChecker.periodSeconds` specifies how often to perform the health check (in seconds)
- `spec.healthChecker.timeoutSeconds` specifies the number of seconds after which the health check probe times out
- `spec.healthChecker.failureThreshold` specifies minimum consecutive failures for the health check to be considered failed
- `spec.healthChecker.disableWriteCheck` specifies whether to disable the write check or not

## Next Steps

- Learn how to use KubeDB to run DB2 [here](/docs/guides/db2/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
