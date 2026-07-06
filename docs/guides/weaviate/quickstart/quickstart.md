---
title: Weaviate Quickstart
menu:
  docs_{{ .version }}:
    identifier: weaviate-quickstart-overview
    name: Overview
    parent: weaviate-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Weaviate

This tutorial will show you how to use KubeDB to run a [Weaviate](https://weaviate.io/) vector database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the Weaviate feature gate during installation.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will need to provide a `StorageClass` in the Weaviate CR specification. Check the available `StorageClass` in your cluster using the following command:

```bash
$ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  38h
longhorn               driver.longhorn.io      Delete          Immediate              true                   6m27s
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   6m23s
```

Here, we have the `longhorn` StorageClass in our cluster. It supports volume expansion, which is required by some of the day-2 operations (such as Volume Expansion and Storage Autoscaling) shown in later guides.

## Find Available WeaviateVersion

When you install KubeDB, it creates a `WeaviateVersion` CR for each supported Weaviate version. Let's check the available `WeaviateVersion`s:

```bash
$ kubectl get weaviateversions
NAME     VERSION   DB_IMAGE                                  DEPRECATED   AGE
1.33.1   1.33.1    ghcr.io/appscode-images/weaviate:1.33.1                34h
```

Notice the `DEPRECATED` column. `true` means that the `WeaviateVersion` is deprecated for the current KubeDB version and KubeDB will not work for that version.

In this tutorial, we will use the `1.33.1` WeaviateVersion CR to create a Weaviate cluster.

## Create a Weaviate Database

KubeDB implements a `Weaviate` CRD to define the specification of a Weaviate database.

Below is the `Weaviate` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      containers:
        - name: weaviate
          securityContext:
            runAsNonRoot: false
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 500m
              memory: 1Gi
  deletionPolicy: WipeOut
  healthChecker:
    periodSeconds: 10
    timeoutSeconds: 10
    failureThreshold: 3
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/quickstart/weaviate-sample.yaml
weaviate.kubedb.com/weaviate-sample created
```

Here,

- `spec.version` is the name of the `WeaviateVersion` CR where the Docker images are specified. In this tutorial, a Weaviate `1.33.1` cluster is created.
- `spec.replicas` specifies the number of Weaviate nodes that form the cluster. A multi-node cluster lets Weaviate replicate collections across nodes for high availability.
- `spec.storageType` can be `Durable` or `Ephemeral`. `Durable` uses a PersistentVolumeClaim; `Ephemeral` uses an `emptyDir`.
- `spec.storage` specifies the size and StorageClass of the PVC that will be dynamically allocated to store data for each Weaviate pod. This storage spec will be passed to the PetSet created by the KubeDB operator to run database pods.
- `spec.deletionPolicy` specifies what KubeDB should do when a user tries to delete the `Weaviate` CR. Deletion policy `DoNotTerminate` prevents deletion of this object if the admission webhook is enabled.

> **Note:** `spec.storage` section is used to create PVC for the database pods. Specify only `requests`, not `limits` — a PVC does not resize automatically.

Now, let's watch the progress of creating the `Weaviate` cluster:

```bash
$ kubectl get weaviate -n demo weaviate-sample -w
NAME              TYPE                  VERSION   STATUS         AGE
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Provisioning   6s
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Provisioning   2m
weaviate-sample   kubedb.com/v1alpha2   1.33.1    Ready          5m21s
```

## Describe Weaviate

Let's describe the `Weaviate` object to see its current state:

```bash
$ kubectl describe weaviate -n demo weaviate-sample
Name:         weaviate-sample
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Weaviate
Metadata:
  Creation Timestamp:  2026-06-30T17:23:54Z
  Finalizers:
    kubedb.com/weaviate
  Generation:        3
  Resource Version:  61772
  UID:               ac548e5a-1d08-4ca6-946e-cdd0a34a2d92
Spec:
  Auth Secret:
    Active From:    2026-06-30T17:23:54Z
    API Group:      
    Kind:           
    Name:           weaviate-sample-auth
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  weaviate
        Resources:
          Limits:
            Cpu:     500m
            Memory:  1Gi
          Requests:
            Cpu:     500m
            Memory:  1Gi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Non Root:  false
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
  Replicas:  3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  longhorn
  Storage Type:          Durable
  Version:               1.33.1
Status:
  Conditions:
    Last Transition Time:  2026-06-30T17:23:54Z
    Message:               The KubeDB operator has started the provisioning of Weaviate demo/weaviate-sample
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2026-06-30T17:28:43Z
    Message:               All replicas are ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2026-06-30T17:28:54Z
    Message:               The Weaviate: demo/weaviate-sample is accepting client requests
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2026-06-30T17:28:55Z
    Message:               The Weaviate: demo/weaviate-sample is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2026-06-30T17:28:55Z
    Message:               The Weaviate demo/weaviate-sample is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

## Find Underlying Kubernetes Resources

KubeDB operator creates a PetSet, PVCs, Services, and Secrets for the Weaviate database. Let's check them:

```bash
$ kubectl get petset -n demo weaviate-sample
NAME              AGE
weaviate-sample   5m19s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
NAME                READY   STATUS    RESTARTS   AGE
weaviate-sample-0   1/1     Running   0          62s
weaviate-sample-1   1/1     Running   0          55s
weaviate-sample-2   1/1     Running   0          44s

$ kubectl get pvc -n demo
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-weaviate-sample-0   Bound    pvc-b8b6d9e6-634f-4ead-b1fd-1bfe549976e4   1Gi        RWO            longhorn       <unset>                 5m18s
data-weaviate-sample-1   Bound    pvc-4e9329c0-8a3d-4402-919c-afa4fe2144c9   1Gi        RWO            longhorn       <unset>                 56s
data-weaviate-sample-2   Bound    pvc-a846947c-212f-4aea-92a7-c8f88ae7f463   1Gi        RWO            longhorn       <unset>                 45s

$ kubectl get service -n demo
NAME                   TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                                         AGE
weaviate-sample        ClusterIP   10.43.98.199   <none>        8080/TCP,50051/TCP,7102/TCP,7103/TCP,8300/TCP   5m23s
weaviate-sample-pods   ClusterIP   None           <none>        8080/TCP,50051/TCP,7102/TCP,7103/TCP,8300/TCP   5m23s
```

KubeDB creates two services for a Weaviate cluster:

- `weaviate-sample` — the primary `ClusterIP` service used by clients. It exposes the REST API (`8080`, or `8443` when TLS is enabled), the gRPC API (`50051`), and the internal cluster ports.
- `weaviate-sample-pods` — a headless (governing) service used for stable per-pod DNS within the cluster.

The internal ports are: `8080` (HTTP REST), `50051` (gRPC), `8300` (raft consensus), `7102` (gossip), and `7103` (cluster data).

## Verify Weaviate YAML Output

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created and is able to accept client connections. Run the following command to see the modified Weaviate object:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  creationTimestamp: "2026-06-30T17:23:54Z"
  finalizers:
  - kubedb.com/weaviate
  generation: 3
  name: weaviate-sample
  namespace: demo
  resourceVersion: "61772"
  uid: ac548e5a-1d08-4ca6-946e-cdd0a34a2d92
spec:
  authSecret:
    activeFrom: "2026-06-30T17:23:54Z"
    apiGroup: ""
    kind: ""
    name: weaviate-sample-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 3
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
      - name: weaviate
        resources:
          limits:
            cpu: 500m
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop:
            - ALL
          runAsNonRoot: false
          seccompProfile:
            type: RuntimeDefault
      podPlacementPolicy:
        name: default
      securityContext: {}
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: longhorn
  storageType: Durable
  version: 1.33.1
status:
  conditions:
  - lastTransitionTime: "2026-06-30T17:23:54Z"
    message: The KubeDB operator has started the provisioning of Weaviate demo/weaviate-sample
    observedGeneration: 1
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2026-06-30T17:28:43Z"
    message: All replicas are ready
    observedGeneration: 3
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2026-06-30T17:28:54Z"
    message: 'The Weaviate: demo/weaviate-sample is accepting client requests'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2026-06-30T17:28:55Z"
    message: 'The Weaviate: demo/weaviate-sample is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2026-06-30T17:28:55Z"
    message: The Weaviate demo/weaviate-sample is successfully provisioned.
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  phase: Ready
```

## Connect to Weaviate

By default, KubeDB enables API-key authentication for the Weaviate cluster and stores the generated key in a Secret named `<database-name>-auth`. Let's check it:

```bash
$ kubectl get secret -n demo weaviate-sample-auth -o yaml
```
```yaml
apiVersion: v1
data:
  AUTHENTICATION_APIKEY_ALLOWED_KEYS: dnpXU2ppUkdOTkVaRXl0Ug==
  AUTHENTICATION_APIKEY_ENABLED: dHJ1ZQ==
  AUTHENTICATION_APIKEY_USERS: YWRtaW4=
kind: Secret
metadata:
  annotations:
    kubedb.com/auth-active-from: "2026-06-30T17:23:54Z"
  creationTimestamp: "2026-06-30T17:23:54Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: weaviate-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: weaviates.kubedb.com
  name: weaviate-sample-auth
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Weaviate
    name: weaviate-sample
    uid: ac548e5a-1d08-4ca6-946e-cdd0a34a2d92
  resourceVersion: "60964"
  uid: ca1779b9-b1ba-4635-a769-6cee4f998384
type: Opaque
```

The Secret stores the standard Weaviate API-key environment variables: `AUTHENTICATION_APIKEY_ENABLED`, `AUTHENTICATION_APIKEY_ALLOWED_KEYS` (the API key itself), and `AUTHENTICATION_APIKEY_USERS` (the bound user, `admin`).

Now, let's connect to the Weaviate cluster using port forwarding. In one terminal, start the port-forward:

```bash
$ kubectl port-forward -n demo svc/weaviate-sample 8080:8080
Forwarding from 127.0.0.1:8080 -> 8080
```

In another terminal, export the API key and call the REST API:

```bash
$ export WEAVIATE_API_KEY=$(kubectl get secret -n demo weaviate-sample-auth -o jsonpath='{.data.AUTHENTICATION_APIKEY_ALLOWED_KEYS}' | base64 -d)

$ curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/v1/.well-known/ready \
  -H "Authorization: Bearer $WEAVIATE_API_KEY"
200
```

Check the cluster nodes — all three should report `HEALTHY`:

```bash
$ curl -s http://localhost:8080/v1/nodes -H "Authorization: Bearer $WEAVIATE_API_KEY" | jq
{
  "nodes": [
    {"name": "weaviate-sample-0", "status": "HEALTHY", "version": "1.33.1", "gitHash": "c87f308", "batchStats": {"queueLength": 0, "ratePerSecond": 0}, "shards": null},
    {"name": "weaviate-sample-1", "status": "HEALTHY", "version": "1.33.1", "gitHash": "c87f308", "batchStats": {"queueLength": 0, "ratePerSecond": 0}, "shards": null},
    {"name": "weaviate-sample-2", "status": "HEALTHY", "version": "1.33.1", "gitHash": "c87f308", "batchStats": {"queueLength": 0, "ratePerSecond": 0}, "shards": null}
  ]
}
```

Let's create a collection (class) and then read the schema back:

```bash
$ curl -s -X POST http://localhost:8080/v1/schema \
  -H "Authorization: Bearer $WEAVIATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"class":"Article","vectorizer":"none"}'
{"class":"Article","invertedIndexConfig":{...},"multiTenancyConfig":{"enabled":false},...}

$ curl -s http://localhost:8080/v1/schema -H "Authorization: Bearer $WEAVIATE_API_KEY" | jq '.classes[].class'
"Article"
```

The collection was created and is served by the cluster.

## AppBinding

KubeDB creates an AppBinding CR that holds the necessary information to connect with the database.

```bash
$ kubectl get appbinding -n demo
NAME              TYPE                  VERSION   AGE
weaviate-sample   kubedb.com/weaviate   1.33.1    5m20s

$ kubectl get appbinding -n demo weaviate-sample -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  creationTimestamp: "2026-06-30T17:23:57Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: weaviate-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: weaviates.kubedb.com
  name: weaviate-sample
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Weaviate
    name: weaviate-sample
    uid: ac548e5a-1d08-4ca6-946e-cdd0a34a2d92
  resourceVersion: "60960"
  uid: 73731c82-f346-4424-be73-05e2ea6e3ebd
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Weaviate
    name: weaviate-sample
    namespace: demo
  clientConfig:
    service:
      name: weaviate-sample
      port: 8080
      scheme: http
  secret:
    name: weaviate-sample-auth
  type: kubedb.com/weaviate
  version: 1.33.1
```

You can use this AppBinding to connect with the Weaviate cluster from external applications, including KubeStash for backup and restore.

## Database DeletionPolicy

This field regulates the deletion process of the related resources when the `Weaviate` object is deleted. The available options are:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB prevents deletion of the database using admission webhooks. If you try to delete it, you will get an error:

```bash
$ kubectl patch -n demo weaviate/weaviate-sample -p '{"spec":{"deletionPolicy":"DoNotTerminate"}}' --type="merge"
weaviate.kubedb.com/weaviate-sample patched

$ kubectl delete weaviate -n demo weaviate-sample
The Weaviate "weaviate-sample" is invalid: spec.deletionPolicy: Invalid value: "weaviate-sample": Can not delete as deletionPolicy is set to "DoNotTerminate"
```

**Halt:**

When `deletionPolicy` is set to `Halt`, KubeDB deletes the `Weaviate` object and its pods but keeps the `PVCs` and `Secrets` intact. This allows you to recreate the database later using the same data.

**Delete:**

When `deletionPolicy` is set to `Delete`, KubeDB deletes the `Weaviate` object, pods, and `PVCs` but keeps the `Secrets`. This allows you to restore the database from a previously taken backup.

**WipeOut:**

When `deletionPolicy` is set to `WipeOut`, KubeDB deletes all resources of this database (pods, PVCs, Secrets, snapshots, etc.). There is no option to recreate the database once deleted with this policy.

```bash
$ kubectl patch -n demo weaviate/weaviate-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
weaviate.kubedb.com/weaviate-sample patched
```

> Be careful when using `WipeOut` — there is no way to recover the database after deletion.

## Next Steps

- Use [custom configuration](/docs/guides/weaviate/configuration/using-config-file.md) for your Weaviate cluster.
- Encrypt traffic with [TLS](/docs/guides/weaviate/tls/overview.md).
- Run day-2 operations: [Restart](/docs/guides/weaviate/restart/restart.md), [Reconfigure](/docs/guides/weaviate/reconfigure/reconfigure.md), [Volume Expansion](/docs/guides/weaviate/volume-expansion/volume-expansion.md), [Vertical Scaling](/docs/guides/weaviate/scaling/vertical-scaling/vertical-scaling.md), [Horizontal Scaling](/docs/guides/weaviate/scaling/horizontal-scaling/horizontal-scaling.md), [Rotate Authentication](/docs/guides/weaviate/rotate-auth/rotate-auth.md), and [Storage Migration](/docs/guides/weaviate/storage-migration/storage-migration.md).
- Automatically scale your cluster with the [Compute Autoscaler](/docs/guides/weaviate/autoscaler/compute/compute-autoscale.md) and [Storage Autoscaler](/docs/guides/weaviate/autoscaler/storage/storage-autoscale.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviate -n demo weaviate-sample
kubectl delete ns demo
```
