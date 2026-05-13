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

- Now, install KubeDB operator in your cluster following the steps [here](/docs/setup/README.md)  and make sure install with helm command including `--set global.featureGates.Qdrant=true` to ensure MSSQLServer CRD installation.

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
- `spec.storage` specifies the size and StorageClass of the PVC that will be dynamically allocated to store data for each Qdrant pod. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods.
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
  ...
Spec:
  Auth Secret:
    Name:  qdrant-sample-auth
  Deletion Policy:  DoNotTerminate
  Replicas:         3
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
    Last Transition Time:  2024-10-01T10:00:00Z
    Message:               The KubeDB operator has started the provisioning of Qdrant: demo/qdrant-sample
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-10-01T10:01:30Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               The Qdrant: demo/qdrant-sample is accepting client requests.
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               DB is ready because of reason
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-10-01T10:02:00Z
    Message:               The Qdrant: demo/qdrant-sample is successfully provisioned.
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:  Ready
```

## Find Underlying Kubernetes Resources

KubeDB operator creates a StatefulSet, PVCs, PVs, and Services for the Qdrant database. Let's check them:

```bash
$ kubectl get statefulset -n demo qdrant-sample
NAME            READY   AGE
qdrant-sample   3/3     15m

$ kubectl get pvc -n demo
NAME                              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-qdrant-sample-0              Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            standard       15m
data-qdrant-sample-1              Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f13   1Gi        RWO            standard       15m
data-qdrant-sample-2              Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f14   1Gi        RWO            standard       15m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                           STORAGECLASS   REASON   AGE
pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-0       standard                15m
pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f13   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-1       standard                15m
pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f14   1Gi        RWO            Delete           Bound    demo/data-qdrant-sample-2       standard                15m

$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
qdrant-sample         ClusterIP   10.96.128.61   <none>        6333/TCP   15m
qdrant-sample-pods    ClusterIP   None           <none>        6333/TCP   15m
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
  creationTimestamp: "2025-06-01T10:00:00Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: qdrant-sample
  namespace: demo
  resourceVersion: "225923"
  uid: e5c9292b-f3a3-4dbf-95c8-1b544096e1d4
spec:
  authSecret:
    kind: Secret
    name: qdrant-sample-auth
  deletionPolicy: DoNotTerminate
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    spec:
      containers:
        - name: qdrant
          resources:
            limits:
              memory: 2Gi
            requests:
              cpu: 500m
              memory: 2Gi
      initContainers:
        - name: qdrant-init
          resources:
            limits:
              memory: 512Mi
            requests:
              cpu: 200m
              memory: 512Mi
  replicas: 3
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: "1.17.0"
status:
  conditions:
    - lastTransitionTime: "2025-06-01T10:00:00Z"
      message: 'The KubeDB operator has started the provisioning of Qdrant: demo/qdrant-sample'
      observedGeneration: 2
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2025-06-01T10:01:30Z"
      message: All replicas are ready for Qdrant demo/qdrant-sample
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2025-06-01T10:02:00Z"
      message: database demo/qdrant-sample is accepting connection
      observedGeneration: 2
      reason: AcceptingConnection
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2025-06-01T10:02:00Z"
      message: database demo/qdrant-sample is ready
      observedGeneration: 2
      reason: AllReplicasReady
      status: "True"
      type: Ready
    - lastTransitionTime: "2025-06-01T10:02:00Z"
      message: 'The Qdrant: demo/qdrant-sample is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  phase: Ready
```

## Connect to Qdrant

KubeDB creates a Secret containing authentication credentials for the Qdrant cluster. Let's check it:

```bash
$ kubectl get secret -n demo qdrant-sample-auth -o yaml
apiVersion: v1
data:
  api-key: <base64-encoded-api-key>
kind: Secret
metadata:
  name: qdrant-sample-auth
  namespace: demo
type: Opaque
```

Now, let's connect to the Qdrant cluster using port forwarding:

```bash
$ kubectl port-forward -n demo svc/qdrant-sample 6333:6333 &
$ export QDRANT_API_KEY=$(kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 -d)

$ curl -H "api-key: $QDRANT_API_KEY" http://localhost:6333/collections
{"result":{"collections":[]},"status":"ok","time":0.001}
```

## AppBinding

KubeDB creates an [AppBinding](/docs/guides/qdrant/concepts/appbinding.md) CR that holds the necessary information to connect with the database.

```bash
$ kubectl get appbinding -n demo -o yaml
```

```yaml
apiVersion: v1
items:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    kind: AppBinding
    metadata:
      creationTimestamp: "2025-06-01T10:00:30Z"
      generation: 1
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
          uid: e5c9292b-f3a3-4dbf-95c8-1b544096e1d4
      resourceVersion: "225711"
      uid: 4d111a65-cf3d-4a74-a77e-24f2dee690df
    spec:
      appRef:
        apiGroup: kubedb.com
        kind: Qdrant
        name: qdrant-sample
        namespace: demo
      clientConfig:
        service:
          name: qdrant-sample
          path: /
          port: 6333
          scheme: http
      secret:
        name: qdrant-sample-auth
      type: kubedb.com/qdrant
      version: "1.17"
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
NAME                        TYPE     DATA   AGE
secret/qdrant-sample-auth   Opaque   2      30m

NAME                                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-qdrant-sample-0   Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f12   1Gi        RWO            standard       29m
persistentvolumeclaim/data-qdrant-sample-1   Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f13   1Gi        RWO            standard       29m
persistentvolumeclaim/data-qdrant-sample-2   Bound    pvc-ccbba9d2-5556-49cd-9ce8-23c28ad56f14   1Gi        RWO            standard       29m
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