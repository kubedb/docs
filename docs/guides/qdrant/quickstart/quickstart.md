---
title: Qdrant Quickstart
menu:
  docs_{{ .version }}:
    identifier: qdrant-quickstart-overview
    name: Overview
    parent: qdrant-quickstart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running Qdrant

This tutorial will show you how to use KubeDB to run a Qdrant database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md)  and make sure install with helm command including `--set global.featureGates.Qdrant=true` to ensure Qdrant CRD installation.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will need to provide `StorageClass` in the Qdrant CR specification. Check available `StorageClass` in your cluster using the following command:

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  10d
```

Here, we have `standard` StorageClass in our cluster.

## Find Available QdrantVersion

When you install KubeDB, it creates `QdrantVersion` CRDs for all supported Qdrant versions. Let's check available `QdrantVersion`s:

```bash
$ kubectl get qdrantversions
NAME     VERSION   DB_IMAGE                                       DEPRECATED   AGE
1.15.4   1.15.4    docker.io/qdrant/qdrant:v1.15.4-unprivileged                13d
1.16.2   1.16.2    docker.io/qdrant/qdrant:v1.16.2-unprivileged                13d
1.17.0   1.17.0    docker.io/qdrant/qdrant:v1.17.0-unprivileged                13d
```

Notice the `DEPRECATED` column. `true` means that QdrantVersion is deprecated for the current KubeDB version and KubeDB will not work for that version.

In this tutorial, we will use `1.17.0` QdrantVersion CR to create a Qdrant cluster. To know more about what `QdrantVersion` CR is and why there may be variation in version names, please visit [here](/docs/guides/qdrant/concepts/catalog.md).

## Create a Qdrant Database

KubeDB implements a `Qdrant` CRD to define the specification of a Qdrant database.

Below is the `Qdrant` object created in this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  mode: "Standalone"
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/qdrant-sample.yaml
qdrant.kubedb.com/qdrant-sample created
```

Here,

- `spec.version` is the name of the QdrantVersion CR where Docker images are specified. In this tutorial, a Qdrant `1.17.0` cluster is created.
- `spec.mode` specifies the Qdrant deployment mode `Standalone` or `Distributed`.
- `spec.storage` specifies the size and StorageClass of the PVC that will be dynamically allocated to store data for each Qdrant pod. This storage spec will be passed to the Petset created by KubeDB operator to run database pods.
- `spec.deletionPolicy` specifies what KubeDB should do when a user tries to delete the `Qdrant` CR. Deletion policy `DoNotTerminate` prevents deletion of this object if the admission webhook is enabled.

> **Note:** `spec.storage` section is used to create PVC for the database pods. Specify only `requests`, not `limits` — PVC does not resize automatically.

Now, let's watch the progress of creating the `Qdrant` cluster:

```bash
$ kubectl get qdrant -n demo qdrant-sample -w
NAME             VERSION   STATUS         AGE
qdrant-sample    1.17.0    Provisioning   5s
qdrant-sample    1.17.0    Provisioning   30s
qdrant-sample    1.17.0    Ready          2m
```

## Describe Qdrant

Let's describe the `Qdrant` object to see its current state:

```bash
$ kubectl describe qdrant -n demo qdrant-sample
Name:         qdrant-sample
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Qdrant
Metadata:
  Creation Timestamp:  2026-05-14T08:43:57Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  3259342
  UID:               42df526b-fba5-4d21-aea4-d20ae5f36f30
Spec:
  Auth Secret:
    Active From:    2026-05-14T08:43:57Z
    API Group:      
    Kind:           Secret
    Name:           qdrant-sample-auth
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     10
    Timeout Seconds:    10
  Mode:                 Standalone
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  qdrant
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
        Fs Group:            1000
      Service Account Name:  qdrant-sample
  Replicas:                  1
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Version:               1.17.0
Status:
  Conditions:
    Last Transition Time:  2026-05-14T08:43:57Z
    Message:               The KubeDB operator has started the provisioning of Qdrant: demo qdrant-sample
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2026-05-14T08:44:11Z
    Message:               All desired replicas are ready.
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2026-05-14T08:44:22Z
    Message:               database demo/qdrant-sample is accepting connection
    Observed Generation:   3
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2026-05-14T08:44:22Z
    Message:               database demo/qdrant-sample is ready
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2026-05-14T08:44:23Z
    Message:               The Qdrant: demo/qdrant-sample is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>

```

## Find Underlying Kubernetes Resources

KubeDB operator creates a Petset, PVCs, PVs, and Services for the Qdrant database. Let's check them:

```bash
$ kubectl get petset -n demo qdrant-sample
NAME            AGE
qdrant-sample   2m34s

$ kubectl get pvc -n demo
NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-qdrant-sample-0                      Bound    pvc-0015c0ad-4ddd-404c-9d8b-b9ea1f6cc15f   1Gi        RWO            standard       <unset>      

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                          STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-0015c0ad-4ddd-404c-9d8b-b9ea1f6cc15f   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-0                      standard       <unset>                          4m14s


$ kubectl get service -n demo
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                       AGE
qdrant-sample        ClusterIP   10.43.18.112   <none>        6333/TCP,6334/TCP             5m36s
qdrant-sample-pods   ClusterIP   None           <none>        6335/TCP                      5m36s

```

## Verify Qdrant YAML Output

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created and is able to accept client connections. Run the following command to see the modified Qdrant object:

```bash
$ kubectl get qdrant -n demo qdrant-sample -o yaml
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Qdrant","metadata":{"annotations":{},"name":"qdrant-sample","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","mode":"Standalone","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"1.17.0"}}
  creationTimestamp: "2026-05-14T08:43:57Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: qdrant-sample
  namespace: demo
  resourceVersion: "3259342"
  uid: 42df526b-fba5-4d21-aea4-d20ae5f36f30
spec:
  authSecret:
    activeFrom: "2026-05-14T08:43:57Z"
    apiGroup: ""
    kind: Secret
    name: qdrant-sample-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 3
    periodSeconds: 10
    timeoutSeconds: 10
  mode: Standalone
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
      - name: qdrant
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
      serviceAccountName: qdrant-sample
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: 1.17.0
status:
  conditions:
  - lastTransitionTime: "2026-05-14T08:43:57Z"
    message: 'The KubeDB operator has started the provisioning of Qdrant: demo qdrant-sample'
    observedGeneration: 1
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2026-05-14T08:44:11Z"
    message: All desired replicas are ready.
    observedGeneration: 3
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2026-05-14T08:44:22Z"
    message: database demo/qdrant-sample is accepting connection
    observedGeneration: 3
    reason: AcceptingConnection
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2026-05-14T08:44:22Z"
    message: database demo/qdrant-sample is ready
    reason: AllReplicasReady
    status: "True"
    type: Ready
  - lastTransitionTime: "2026-05-14T08:44:23Z"
    message: 'The Qdrant: demo/qdrant-sample is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  phase: Ready
```

## Connect to Qdrant

KubeDB creates a Secret containing authentication credentials for the Qdrant cluster. Let's check it:

```bash
$ kubectl get secret -n demo qdrant-sample-auth -o yaml
```
```yaml
apiVersion: v1
data:
  api-key: ZUlxZVlMcnB4dmJ4SFBsbA==
  read-only-api-key: M04yN25yakF3WWtlS0hPYg==
kind: Secret
metadata:
  annotations:
    kubedb.com/auth-active-from: "2026-05-14T08:43:57Z"
  creationTimestamp: "2026-05-14T08:43:57Z"
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: qdrant-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: qdrants.kubedb.com
  name: qdrant-sample-auth
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Qdrant
    name: qdrant-sample
    uid: 42df526b-fba5-4d21-aea4-d20ae5f36f30
  resourceVersion: "3259248"
  uid: 760da5c0-76de-48ae-838d-01a63cf90b8b
type: Opaque
```

Now, let's connect to the Qdrant cluster using port forwarding:

```bash
$ kubectl port-forward -n demo svc/qdrant-sample 6333:6333
$ export QDRANT_API_KEY=$(kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 -d)

$ curl -H "api-key: $QDRANT_API_KEY" http://localhost:6333/collections
{"result":{"collections":[{"name":"KubeDBHealthCheckCollection"}]},"status":"ok","time":0.00001235}
```

Let's create a collection with some vector data:

```bash
$ curl -X PUT http://localhost:6333/collections/demo_vectors \
  -H "Content-Type: application/json" \
  -H "api-key: $QDRANT_API_KEY" \
  -d '{
    "vectors": {
      "size": 8,
      "distance": "Cosine"
    }
  }'
{"result":true,"status":"ok","time":0.050803361}

$ curl -X PUT "http://localhost:6333/collections/demo_vectors/points?wait=true" \
  -H "Content-Type: application/json" \
  -H "api-key: $QDRANT_API_KEY" \
  -d '{
    "points": [
      {"id": 1, "vector": [0.15, 0.22, 0.31, 0.44, 0.51, 0.68, 0.73, 0.89], "payload": {"label": "apple"}},
      {"id": 2, "vector": [0.12, 0.28, 0.35, 0.42, 0.53, 0.64, 0.71, 0.85], "payload": {"label": "banana"}},
      {"id": 3, "vector": [0.18, 0.21, 0.33, 0.46, 0.50, 0.66, 0.77, 0.82], "payload": {"label": "cherry"}}
    ]
  }'
{"result":{"operation_id":1,"status":"completed"},"status":"ok","time":0.001645376}
```

Now scroll through the points to verify they were stored:

```bash
$ curl -X POST http://localhost:6333/collections/demo_vectors/points/scroll \
  -H "Content-Type: application/json" \
  -H "api-key: $QDRANT_API_KEY" \
  -d '{"limit": 5, "with_payload": true, "with_vector": false}' | jq
{
  "result": {
    "points": [
      {"id": 1, "payload": {"label": "apple"}, "version": 1},
      {"id": 2, "payload": {"label": "banana"}, "version": 1},
      {"id": 3, "payload": {"label": "cherry"}, "version": 1}
    ],
    "next_page_offset": null
  },
  "status": "ok",
  "time": 0.000086921
}
```

## AppBinding

KubeDB creates an AppBinding CR that holds the necessary information to connect with the database.

```bash
$ kubectl get appbinding -n demo -o yaml
```

```yaml
apiVersion: v1
items:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"kubedb.com/v1alpha2","kind":"Qdrant","metadata":{"annotations":{},"name":"qdrant-sample","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","mode":"Standalone","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"1.17.0"}}
      creationTimestamp: "2026-05-14T08:44:00Z"
      generation: 2
      labels:
        app.kubernetes.io/component: database
        app.kubernetes.io/instance: qdrant-sample
        app.kubernetes.io/managed-by: kubedb.com
        app.kubernetes.io/name: qdrants.kubedb.com
      name: qdrant-sample
      namespace: demo
      ownerReferences:
        - apiVersion: kubedb.com/v1alpha2
          blockOwnerDeletion: true
          controller: true
          kind: Qdrant
          name: qdrant-sample
          uid: 42df526b-fba5-4d21-aea4-d20ae5f36f30
      resourceVersion: "3259280"
      uid: 7183b027-3eda-42ed-95f7-0a366d493464
    spec:
      appRef:
        apiGroup: kubedb.com
        kind: Qdrant
        name: qdrant-sample
        namespace: demo
      clientConfig:
        service:
          name: qdrant-sample
          port: 6333
          scheme: http
      secret:
        name: qdrant-sample-auth
      type: kubedb.com/qdrant
      version: 1.17.0
kind: List
metadata:
  resourceVersion: ""
```

You can use this AppBinding to connect with the Qdrant cluster from external applications.

## Database DeletionPolicy

This field regulates the deletion process of the related resources when the `Qdrant` object is deleted. The available options are:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB prevents deletion of the database using admission webhooks. If you try to delete it, you will get an error:

```bash
$ kubectl patch -n demo qdrant/qdrant-sample -p '{"spec":{"deletionPolicy":"DoNotTerminate"}}' --type="merge"
qdrant.kubedb.com/qdrant-sample patched

$ kubectl delete qdrant -n demo qdrant-sample
The Qdrant "qdrant-sample" is invalid: spec.deletionPolicy: Invalid value: "qdrant-sample": Can not delete as deletionPolicy is set to "DoNotTerminate"
```

**Halt:**

When `deletionPolicy` is set to `Halt`, KubeDB deletes the `Qdrant` object and its pods but keeps the `PVCs`, `Secrets`, and backup snapshots intact. This allows you to recreate the database later using the same data.

At first, set the `deletionPolicy` to `Halt` and then delete the database:

```bash
$ kubectl patch -n demo qdrant/qdrant-sample -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
qdrant.kubedb.com/qdrant-sample patched

$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted
```

Now, check that the PVCs and Secrets still exist:

```bash
$ kubectl get secret,pvc -n demo
NAME                          TYPE                       DATA   AGE
secret/qdrant-sample-auth     Opaque                     2      11m
secret/qdrant-sample-f36f30   Opaque                     1      11m

NAME                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-qdrant-sample-0   Bound    pvc-0015c0ad-4ddd-404c-9d8b-b9ea1f6cc15f   1Gi        RWO            standard     <unset>                 11m

```

You can recreate your Qdrant database later using these PVCs and Secrets.

**Delete:**

When `deletionPolicy` is set to `Delete`, KubeDB deletes the `Qdrant` object, pods, and `PVCs` but keeps the `Secrets` and backup snapshots. This allows you to restore the database from a previously taken backup.

**WipeOut:**

When `deletionPolicy` is set to `WipeOut`, KubeDB deletes all resources of this database (pods, PVCs, Secrets, snapshots, etc.). There is no option to recreate the database once deleted with this policy.

```bash
$ kubectl patch -n demo qdrant/qdrant-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
qdrant.kubedb.com/qdrant-sample patched
```

> Be careful when using `WipeOut` — there is no way to recover the database after deletion.

## Next Steps

- Learn about [backup and restore](/docs/guides/qdrant/backup/overview/index.md) Qdrant using KubeStash.
- Want to set up distributed Qdrant deployment? Check [Distributed Deployment](/docs/guides/qdrant/distributed-deployment/overview.md).
- Detail concepts of [Qdrant object](/docs/guides/qdrant/concepts/qdrant.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```