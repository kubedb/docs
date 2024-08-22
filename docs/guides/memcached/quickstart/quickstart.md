---
title: Memcached Quickstart
menu:
  docs_{{ .version }}:
    identifier: mc-quickstart-quickstart
    name: Overview
    parent: mc-quickstart-memcached
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Memcached QuickStart

This tutorial will show you how to use KubeDB to run a Memcached server.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/memcached/memcached-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/examples/memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME      STATUS    AGE
demo      Active    1s
```

## Find Available MemcachedVersion

When you have installed KubeDB, it has created `MemcachedVersion` crd for all supported Memcached versions. Check 0

```bash
$ kubectl get memcachedversions
NAME        VERSION    DB_IMAGE                                            DEPRECATED   AGE
1.5.22      1.5.22     ghcr.io/appscode-images/memcached:1.5.22-alpine                  2h
1.6.22      1.6.22     ghcr.io/appscode-images/memcached:1.6.22-alpine                  2h
1.6.29      1.6.29     ghcr.io/appscode-images/memcached:1.6.29-alpine                  2h
```

## Create a Memcached server

KubeDB implements a `Memcached` CRD to define the specification of a Memcached server. Below is the `Memcached` object created in this tutorial.

`Note`: If your `KubeDB version` is less or equal to `v2024.6.4`, You have to use `v1alpha2` apiVersion.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 3
  version: "1.6.22"
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: DoNotTerminate

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/quickstart/demo-v1.yaml
memcached.kubedb.com/memcd-quickstart created
```

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 3
  version: "1.6.22"
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
  terminationPolicy: DoNotTerminate
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/quickstart/demo-v1alpha2.yaml
memcached.kubedb.com/memcd-quickstart created
```

Here,

- `spec.replicas` is an optional field that specifies the number of desired Instances/Replicas of Memcached server. It defaults to 1.
- `spec.version` is the version of Memcached server. In this tutorial, a Memcached 1.5.4 database is going to be created.
- `spec.resource` is an optional field that specifies how much CPU and memory (RAM) each Container needs. To learn details about Managing Compute Resources for Containers, please visit [here](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/).
- `spec.deletionPolicy` or `spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Memcached` crd or which resources KubeDB should keep or delete when you delete `Memcached` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/memcached/concepts/memcached.md#specdeletionpolicy)

KubeDB operator watches for `Memcached` objects using Kubernetes api. When a `Memcached` object is created, KubeDB operator will create a new PetSet and a Service with the matching Memcached object name.
```bash
$ kubectl get mc -n demo
NAME               VERSION    STATUS    AGE
memcd-quickstart   1.6.22     Running   2m

$ kubectl describe mc -n demo memcd-quickstart
Name:         memcd-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1
Kind:         Memcached
Metadata:
  Creation Timestamp:  2024-08-22T13:54:45Z
  Finalizers:
    kubedb.com
  Generation:        1
  Resource Version:  3428
  UID:               4c8bea2e-c4a3-4310-9a7d-d8b60ac47d5b
Spec:
  Deletion Policy:  DoNotTerminate
  Health Checker:
    Failure Threshold:  1
    Period Seconds:     10
    Timeout Seconds:    10
  Pod Template:
    Controller:
    Metadata:
    Spec:
      Containers:
        Name:  memcached
        Resources:
          Limits:
            Cpu:     500m
            Memory:  128Mi
          Requests:
            Cpu:     250m
            Memory:  64Mi
        Security Context:
          Allow Privilege Escalation:  false
          Capabilities:
            Drop:
              ALL
          Run As Group:     999
          Run As Non Root:  true
          Run As User:      999
          Seccomp Profile:
            Type:  RuntimeDefault
      Pod Placement Policy:
        Name:  default
      Security Context:
        Fs Group:            999
      Service Account Name:  memcd-quickstart
  Replicas:                  3
  Version:                   1.6.22
Status:
  Conditions:
    Last Transition Time:  2024-08-22T13:54:45Z
    Message:               The KubeDB operator has started the provisioning of Memcached: demo/memcd-quickstart
    Reason:                DatabaseProvisioningStartedSuccessfully
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2024-08-22T13:54:55Z
    Message:               All desired replicas are ready.
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2024-08-22T13:55:05Z
    Message:               The Memcached: demo/memcd-quickstart is accepting mcClient requests.
    Observed Generation:   1
    Reason:                DatabaseAcceptingConnectionRequest
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2024-08-22T13:55:05Z
    Message:               The Memcached: demo/memcd-quickstart is ready.
    Observed Generation:   1
    Reason:                ReadinessCheckSucceeded
    Status:                True
    Type:                  Ready
    Last Transition Time:  2024-08-22T13:55:05Z
    Message:               The Memcached: demo/memcd-quickstart is successfully provisioned.
    Observed Generation:   1
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Observed Generation:     1
  Phase:                   Ready
Events:
  Type    Reason      Age    From             Message
  ----    ------      ----   ----             -------
  Normal  Successful  5m37s  KubeDB Operator  Successfully created governing service
  Normal  Successful  5m37s  KubeDB Operator  Successfully created Service
  Normal  Successful  5m37s  KubeDB Operator  Successfully created appbinding
  Normal  Successful  5m28s  KubeDB Operator  Successfully patched PetSet
  Normal  Successful  5m28s  KubeDB Operator  Successfully patched Memcached
  Normal  Successful  5m28s  KubeDB Operator  Successfully patched PetSet
  Normal  Successful  5m28s  KubeDB Operator  Successfully patched Memcached
  Normal  Successful  5m18s  KubeDB Operator  Successfully patched PetSet
  Normal  Successful  5m18s  KubeDB Operator  Successfully patched Memcached
  Normal  Successful  5m18s  KubeDB Operator  Successfully patched PetSet
  Normal  Successful  5m18s  KubeDB Operator  Successfully patched Memcached
  Normal  Successful  5m18s  KubeDB Operator  Successfully patched PetSet
  Normal  Successful  5m18s  KubeDB Operator  Successfully patched Memcached

$ kubectl get petset -n demo
NAME               AGE
memcd-quickstart   8m15s

$ kubectl get service -n demo
NAME                    TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
memcd-quickstart        ClusterIP   10.96.115.90   <none>        11211/TCP   9m7s
memcd-quickstart-pods   ClusterIP   None           <none>        11211/TCP   9m7s
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified Memcached object:

```yaml
$ kubectl get mc -n demo memcd-quickstart -o yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  creationTimestamp: 2018-10-03T09:40:38Z
  finalizers:
  - kubedb.com
  generation: 1
  name: memcd-quickstart
  namespace: demo
  resourceVersion: "23592"
  selfLink: /apis/kubedb.com/v1/namespaces/demo/memcacheds/memcd-quickstart
  uid: 62b08ec3-c6f0-11e8-8ebc-0800275bbbee
spec:
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
  replicas: 3
  deletionPolicy: Delete
  version: 1.6.22
status:
  observedGeneration: 1$4210395375389091791
  phase: Running
```

Now, you can connect to this Memcached cluster using `telnet`.
Here, we will connect to Memcached server from local-machine through port-forwarding.

```bash
$ kubectl get pods -n demo
NAME                                READY     STATUS    RESTARTS   AGE
memcd-quickstart-57d88d6595-gfptm   1/1       Running   0          3m
memcd-quickstart-57d88d6595-wmp5p   1/1       Running   0          3m
memcd-quickstart-57d88d6595-xf4z2   1/1       Running   0          3m

// We will connect to `memcd-quickstart-667cd68854-gs69q` pod from local-machine using port-frowarding.
$ kubectl port-forward -n demo memcd-quickstart-57d88d6595-gfptm 11211
Forwarding from 127.0.0.1:11211 -> 11211

# Connect Memcached cluster from localmachine through telnet.
~ $ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.

# Save data Command:
set my_key 0 2592000 1
2
# Output:
STORED

# Meaning:
# 0       => no flags
# 2592000 => TTL (Time-To-Live) in [s]
# 1       => size in byte
# 2       => value

# View data command
get my_key
# Output
VALUE my_key 0 1
2
END

# Exit
quit
```

## DoNotTerminate Property

When `deletionPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete mc memcd-quickstart -n demo
Error from server (BadRequest): admission webhook "memcached.validators.kubedb.com" denied the request: memcached "memcd-quickstart" can't be halted. To delete, change spec.deletionPolicy
```

Now, run `kubectl edit mc memcd-quickstart -n demo` to set `spec.deletionPolicy` to `Halt` (which creates `dormantdatabase` when memcached is deleted and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Halt`). Then you will be able to delete/halt the database. 

Learn details of all `DeletionPolicy` [here](/docs/guides/memcached/concepts/memcached.md#specdeletionpolicy)

## Halt Database

When [DeletionPolicy](/docs/guides/memcached/concepts/memcached.md#specdeletionpolicy) is set to `Halt`, it will halt the Memcached server instead of deleting it. Here, you delete the Memcached object, KubeDB operator will delete the Deployment and its pods. In KubeDB parlance, we say that `memcd-quickstart` Memcached server has entered into dormant state. This is represented by KubeDB operator by creating a matching DormantDatabase object.

```bash
$ kubectl delete mc memcd-quickstart -n demo
memcached.kubedb.com "memcd-quickstart" deleted

$ kubectl get drmn -n demo memcd-quickstart
NAME               STATUS    AGE
memcd-quickstart   Pausing   21s

$ kubectl get drmn -n demo memcd-quickstart
NAME               STATUS    AGE
memcd-quickstart   Halted    2m
```

```yaml
$ kubectl get drmn -n demo memcd-quickstart -o yaml
apiVersion: kubedb.com/v1
kind: DormantDatabase
metadata:
  creationTimestamp: 2018-10-03T09:49:16Z
  finalizers:
  - kubedb.com
  generation: 1
  labels:
    app.kubernetes.io/name: memcacheds.kubedb.com
  name: memcd-quickstart
  namespace: demo
  resourceVersion: "24242"
  selfLink: /apis/kubedb.com/v1/namespaces/demo/dormantdatabases/memcd-quickstart
  uid: 97ad28ef-c6f1-11e8-8ebc-0800275bbbee
spec:
  origin:
    metadata:
      creationTimestamp: 2018-10-03T09:40:38Z
      name: memcd-quickstart
      namespace: demo
    spec:
      memcached:
        podTemplate:
          controller: {}
          metadata: {}
          spec:
            resources:
              limits:
                cpu: 500m
                memory: 128Mi
              requests:
                cpu: 250m
                memory: 64Mi
        replicas: 3
        deletionPolicy: Halt
        version: 1.6.22
status:
  observedGeneration: 1$7678503742307285743
  pausingTime: 2018-10-03T09:50:10Z
  phase: Halted
```

Here,

- `spec.origin` is the spec of the original spec of the original Memcached object.
- `status.phase` points to the current database state `Halted`.

## Resume Dormant Database

To resume the database from the dormant state, create same `Memcached` object with same Spec.

In this tutorial, the dormant database can be resumed by creating `Memcached` database using demo-1.yaml file.

The below command resumes the dormant database `memcd-quickstart`.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/quickstart/demo-1.yaml
memcached.kubedb.com/memcd-quickstart created
```

## Wipeout Dormant Database

You can wipe out a DormantDatabase while deleting the objet by setting `spec.wipeOut` to true. KubeDB operator will delete any relevant resources of this `Memcached` database.

```yaml
$ kubectl delete mc memcd-quickstart -n demo
memcached "memcd-quickstart" deleted

$ kubectl edit drmn -n demo memcd-quickstart
apiVersion: kubedb.com/v1
kind: DormantDatabase
metadata:
  name: memcd-quickstart
  namespace: demo
  ...
spec:
  wipeOut: true
  ...
status:
  phase: Halted
  ...
```

If `spec.wipeOut` is not set to true while deleting the `dormantdatabase` object, then only this object will be deleted and `kubedb-operator` won't delete related Secrets.

## Delete DormantDatabase

As it is already discussed above, `DormantDatabase` can be deleted with or without wiping out the resources. To delete the `dormantdatabase`,

```bash
$ kubectl delete drmn memcd-quickstart -n demo
dormantdatabase "memcd-quickstart" deleted
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mc/memcd-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mc/memcd-quickstart

kubectl patch -n demo drmn/memcd-quickstart -p '{"spec":{"wipeOut":true}}' --type="merge"
kubectl delete -n demo drmn/memcd-quickstart

kubectl delete ns demo
```

## Next Steps

- Monitor your Memcached server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md).
- Monitor your Memcached server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/memcached/private-registry/using-private-registry.md) to deploy Memcached with KubeDB.
- Detail concepts of [Memcached object](/docs/guides/memcached/concepts/memcached.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
