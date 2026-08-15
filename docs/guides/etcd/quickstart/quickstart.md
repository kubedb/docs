---
title: Etcd Quickstart
menu:
  docs_{{ .version }}:
    identifier: etcd-quickstart-quickstart
    name: Overview
    parent: etcd-quickstart-etcd
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd QuickStart

This tutorial will show you how to use KubeDB to run an [etcd](https://etcd.io/) cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install the KubeDB cli on your workstation and the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). etcd support is behind an **alpha feature gate**, so you must set `global.featureGates.Etcd=true` to install the etcd CRDs and enable the controllers:

  ```bash
  helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
    --namespace kubedb --create-namespace \
    --set-file global.license=/path/to/the/license.txt \
    --set global.featureGates.Etcd=true \
    --wait --burst-limit=10000 --debug
  ```

  If you install the operator components separately, the equivalent operator flag is `--feature-gates=Etcd=true`, and it has to be set on the provisioner, the ops-manager, the autoscaler and the crd-manager alike.

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in the cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  20h
  ```

  > etcd commits every write to disk before acknowledging it. For anything beyond this tutorial, use a StorageClass backed by SSD/NVMe.

- To keep things isolated, this tutorial uses a separate namespace called `demo`. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create namespace demo
  namespace/demo created

  $ kubectl get namespaces
  NAME          STATUS    AGE
  demo          Active    10s
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available EtcdVersion

When you have installed KubeDB, it has created an `EtcdVersion` crd for every supported etcd version. Check:

```bash
$ kubectl get etcdversions
NAME     VERSION   DB_IMAGE                              DEPRECATED   AGE
3.5.21   3.5.21    ghcr.io/appscode-images/etcd:v3.5.21                94s
3.6.4    3.6.4     ghcr.io/appscode-images/etcd:v3.6.4                 94s
```

`EtcdVersion` is a cluster-scoped resource, so it is not namespaced. You can inspect one to see the exact images and the allowed update path:

```bash
$ kubectl get etcdversion 3.6.4 -o yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: EtcdVersion
metadata:
  name: 3.6.4
spec:
  db:
    image: ghcr.io/appscode-images/etcd:v3.6.4
  exporter:
    image: ghcr.io/appscode-images/etcd:v3.6.4
  securityContext:
    runAsUser: 1000
  updateConstraints:
    allowlist:
    - '>= 3.6.4, <= 3.7.0'
  version: 3.6.4
```

## Create an Etcd Cluster

KubeDB implements an `Etcd` CRD to define the specification of an etcd cluster. Below is the `Etcd` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-quickstart
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  storageType: Durable
  storage:
    resources:
      requests:
        storage: "1Gi"
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/etcd/quickstart/etcd.yaml
etcd.kubedb.com/etcd-quickstart created
```

Here,

- `spec.version` is the name of the [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md) crd where the docker images are specified. In this tutorial, an etcd `3.6.4` cluster is created.
- `spec.replicas` is the number of etcd members. It defaults to `3`. etcd replicates with Raft, so a `2n+1` member cluster tolerates `n` member failures — use an **odd** number.
- `spec.storage` specifies the PVC spec that will be dynamically allocated to hold the etcd data directory (`/var/lib/etcd`). This storage spec will be passed to the PetSet created by the KubeDB operator to run the database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of the `Etcd` crd, or which resources KubeDB should keep or delete when you delete the `Etcd` crd. If the admission webhook is enabled, it prevents users from deleting the database as long as `spec.deletionPolicy` is set to `DoNotTerminate`.

> Note: The `spec.storage` section is used to create the PVC for the database pods. It will create a PVC with the storage size specified in the `storage.resources.requests` field. Don't specify limits here. PVCs do not get resized automatically.

The KubeDB operator watches for `Etcd` objects using the Kubernetes api. When an `Etcd` object is created, the operator creates a PetSet, a client Service, a headless governing Service, an auth Secret, and the RBAC objects, all named after the `Etcd` object.

Because etcd's bootstrap semantics require a single seed member, the PetSet is created with **one** replica. The operator then adds each further member through etcd's membership API — first as a non-voting **learner**, then promoting it to a voting member once it has caught up — so you will see the cluster grow from 1 to 3 members over the first few reconcile passes rather than all at once.

```bash
$ kubectl get etcd -n demo
NAME              VERSION   STATUS         AGE
etcd-quickstart   3.6.4     Provisioning   20s
```

Wait until `STATUS` becomes `Ready`. You can block on it:

```bash
$ kubectl wait --for=condition=Ready etcd/etcd-quickstart -n demo --timeout=10m
etcd.kubedb.com/etcd-quickstart condition met

$ kubectl get etcd -n demo
NAME              VERSION   STATUS   AGE
etcd-quickstart   3.6.4     Ready    3m2s
```

Describe the object to see the defaults the operator filled in, and the conditions it walked through:

```bash
$ kubectl describe etcd -n demo etcd-quickstart
Name:         etcd-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Etcd
Metadata:
  Creation Timestamp:  2026-01-14T08:25:26Z
  Finalizers:
    kubedb.com
  Generation:        3
  Resource Version:  12345
  UID:               dd69e514-3049-4d08-8b57-92f8246dda35
Spec:
  Auth Secret:
    Kind:  Secret
    Name:  etcd-quickstart-auth
  Deletion Policy:  WipeOut
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  etcd
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
      Init Containers:
        Name:  etcd-init
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
      Security Context:
        Fs Group:            1000
      Service Account Name:  etcd-quickstart
  Replicas:                  3
  Storage:
    Access Modes:
      ReadWriteOnce
    Resources:
      Requests:
        Storage:         1Gi
    Storage Class Name:  standard
  Storage Type:          Durable
  Version:               3.6.4
Status:
  Conditions:
    Last Transition Time:  2026-01-14T08:25:26Z
    Message:               The KubeDB operator has started the provisioning of Etcd: demo/etcd-quickstart
    Observed Generation:   1
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2026-01-14T08:26:50Z
    Message:               All replicas are ready for Etcd demo/etcd-quickstart
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2026-01-14T08:27:10Z
    Message:               The Etcd: demo/etcd-quickstart is accepting connection requests.
    Observed Generation:   3
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2026-01-14T08:27:10Z
    Message:               The Etcd: demo/etcd-quickstart is ready.
    Observed Generation:   3
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2026-01-14T08:27:13Z
    Message:               Etcd: demo/etcd-quickstart is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

The offshoot objects the operator created:

```bash
$ kubectl get petset -n demo
NAME              AGE
etcd-quickstart   3m14s

$ kubectl get pods -n demo -l app.kubernetes.io/instance=etcd-quickstart
NAME                READY   STATUS    RESTARTS   AGE
etcd-quickstart-0   1/1     Running   0          3m20s
etcd-quickstart-1   1/1     Running   0          2m48s
etcd-quickstart-2   1/1     Running   0          2m19s

$ kubectl get pvc -n demo
NAME                     STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-etcd-quickstart-0   Bound    pvc-1e1850b8-4e5c-418c-a722-89df98f28998   1Gi        RWO            standard       3m40s
data-etcd-quickstart-1   Bound    pvc-e2bb4b02-b138-4589-9e43-bcaf599b6513   1Gi        RWO            standard       3m08s
data-etcd-quickstart-2   Bound    pvc-988ab6b2-e5ed-4c75-8418-31186bd1d3db   1Gi        RWO            standard       2m39s

$ kubectl get service -n demo
NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
etcd-quickstart         ClusterIP   10.96.26.38    <none>        2379/TCP            4m15s
etcd-quickstart-pods    ClusterIP   None           <none>        2379/TCP,2380/TCP   4m15s
```

Here,

- `etcd-quickstart` is the client Service on port `2379`. Every etcd member serves the client API and forwards writes to the current Raft leader, so a single Service over all members is correct — there is no primary/standby split to route around.
- `etcd-quickstart-pods` is the headless governing Service. The stable per-member DNS names derived from it (`etcd-quickstart-0.etcd-quickstart-pods.demo.svc.cluster.local`) are what etcd uses as its peer (`2380`) and client (`2379`) URLs.

The KubeDB operator sets `status.phase` to `Ready` once the cluster is successfully created. Run the following command to see the full, modified Etcd object:

```bash
$ kubectl get etcd -n demo etcd-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Etcd","metadata":{"annotations":{},"name":"etcd-quickstart","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","version":"3.6.4"}}
  creationTimestamp: "2026-01-14T08:25:26Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: etcd-quickstart
  namespace: demo
  resourceVersion: "12345"
  uid: dd69e514-3049-4d08-8b57-92f8246dda35
spec:
  authSecret:
    kind: Secret
    name: etcd-quickstart-auth
  deletionPolicy: WipeOut
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      containers:
      - name: etcd
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
      initContainers:
      - name: etcd-init
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
      securityContext:
        fsGroup: 1000
      serviceAccountName: etcd-quickstart
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  version: 3.6.4
status:
  conditions:
  - lastTransitionTime: "2026-01-14T08:25:26Z"
    message: 'The KubeDB operator has started the provisioning of Etcd: demo/etcd-quickstart'
    observedGeneration: 1
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2026-01-14T08:26:50Z"
    message: All replicas are ready for Etcd demo/etcd-quickstart
    observedGeneration: 3
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2026-01-14T08:27:10Z"
    message: 'The Etcd: demo/etcd-quickstart is accepting connection requests.'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2026-01-14T08:27:10Z"
    message: 'The Etcd: demo/etcd-quickstart is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2026-01-14T08:27:13Z"
    message: 'Etcd: demo/etcd-quickstart is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 3
  phase: Ready
```

## Connect to the Cluster

etcd ships its own CLI, `etcdctl`, inside the database image, so the simplest way to try the cluster is from inside a member pod.

First, confirm that the cluster has three healthy members and that one of them is the Raft leader:

```bash
$ kubectl exec -it -n demo etcd-quickstart-0 -- etcdctl \
    --endpoints=http://etcd-quickstart:2379 member list -w table
+------------------+---------+-------------------+--------------------------------------------------------+--------------------------------------------------------+------------+
|        ID        | STATUS  |       NAME        |                       PEER ADDRS                       |                      CLIENT ADDRS                      | IS LEARNER |
+------------------+---------+-------------------+--------------------------------------------------------+--------------------------------------------------------+------------+
| 8e9e05c52164694d | started | etcd-quickstart-0 | http://etcd-quickstart-0.etcd-quickstart-pods.demo:2380 | http://etcd-quickstart-0.etcd-quickstart-pods.demo:2379 |      false |
| 91bc3c398fb3c146 | started | etcd-quickstart-1 | http://etcd-quickstart-1.etcd-quickstart-pods.demo:2380 | http://etcd-quickstart-1.etcd-quickstart-pods.demo:2379 |      false |
| fd422379fda50e48 | started | etcd-quickstart-2 | http://etcd-quickstart-2.etcd-quickstart-pods.demo:2380 | http://etcd-quickstart-2.etcd-quickstart-pods.demo:2379 |      false |
+------------------+---------+-------------------+--------------------------------------------------------+--------------------------------------------------------+------------+
```

> The `IS LEARNER` column is worth knowing about: while KubeDB is scaling the cluster up, a freshly added member appears here as a learner (`true`). A learner replicates the log but does not vote and does not count toward quorum, so it cannot stall the cluster while it catches up. KubeDB promotes it to a voting member automatically once it has caught up with the leader.

Now write and read a key:

```bash
$ kubectl exec -it -n demo etcd-quickstart-0 -- etcdctl \
    --endpoints=http://etcd-quickstart:2379 put /hello "hello kubedb"
OK

$ kubectl exec -it -n demo etcd-quickstart-0 -- etcdctl \
    --endpoints=http://etcd-quickstart:2379 get /hello
/hello
hello kubedb
```

Because the write went through Raft, the value is readable from any member — read it back through a different pod to see that:

```bash
$ kubectl exec -it -n demo etcd-quickstart-2 -- etcdctl \
    --endpoints=http://etcd-quickstart-2.etcd-quickstart-pods.demo:2379 get /hello
/hello
hello kubedb
```

To reach the cluster from your workstation instead, port-forward the client Service:

```bash
$ kubectl port-forward -n demo svc/etcd-quickstart 2379
Forwarding from 127.0.0.1:2379 -> 2379
```

### Credentials

KubeDB provisions a credential for etcd's built-in RBAC superuser — which etcd requires to literally be named `root` — and stores it in the Secret `<etcd-object-name>-auth`:

```bash
$ kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
root

$ kubectl get secret -n demo etcd-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
6q8u2jMOWOOZXk
```

Once etcd's RBAC is turned on for the cluster, pass those credentials to `etcdctl` with `--user=root:<password>`. Rotating the credential — and enabling RBAC in the first place — is done with a `RotateAuth` [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md), which does **not** restart the pods, because etcd applies a password change live.

You can also bring your own credential instead of letting KubeDB generate one; see [spec.authSecret](/docs/guides/etcd/concepts/etcd.md#specauthsecret).

## DoNotTerminate Property

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of the `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement the `DoNotTerminate` feature. If the admission webhook is enabled, it prevents users from deleting the database as long as `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete etcd etcd-quickstart -n demo
The Etcd "etcd-quickstart" is invalid: spec.deletionPolicy: Invalid value: "etcd-quickstart": Can not delete as deletionPolicy is set to "DoNotTerminate"
```

Now, run `kubectl edit etcd etcd-quickstart -n demo` to set `spec.deletionPolicy` to `Halt`. Then you will be able to delete/halt the database.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo etcd/etcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
etcd.kubedb.com/etcd-quickstart patched

$ kubectl delete -n demo etcd/etcd-quickstart
etcd.kubedb.com "etcd-quickstart" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Tips for Testing

If you are just testing some basic functionality, you might want to avoid the additional hassle of some safety features that are great for a production environment. You can follow these tips to avoid them.

- **Use `deletionPolicy: WipeOut`.** By default, KubeDB preserves your `PVCs` and auth `Secrets` so a database can be resumed from a previous one. If you don't want to resume the database, use `spec.deletionPolicy: WipeOut`. It deletes everything created by KubeDB for a particular Etcd crd when you delete the crd.
- **Use `spec.replicas: 1`** for a throwaway cluster. etcd has no standalone mode — a single member is simply a one-member Raft cluster — so this is a supported configuration, just one with no fault tolerance at all.
- **Use `spec.storageType: Ephemeral`** if you don't want PVCs at all. Note that this cannot be combined with `deletionPolicy: Halt`.

## Next Steps

- Detail concepts of the [Etcd object](/docs/guides/etcd/concepts/etcd.md).
- Detail concepts of the [EtcdVersion object](/docs/guides/etcd/concepts/etcdversion.md).
- Detail concepts of the [EtcdOpsRequest object](/docs/guides/etcd/concepts/etcdopsrequest.md), including etcd's own `MoveLeader`, `Defragment` and `Compact` maintenance operations.
- Detail concepts of the [EtcdAutoscaler object](/docs/guides/etcd/concepts/etcdautoscaler.md).
- Snapshot backup with the [EtcdArchiver object](/docs/guides/etcd/concepts/etcdarchiver.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
