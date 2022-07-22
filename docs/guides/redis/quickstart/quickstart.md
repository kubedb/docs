---
title: Redis Quickstart
menu:
  docs_{{ .version }}:
    identifier: rd-quickstart-quickstart
    name: Overview
    parent: rd-quickstart-redis
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Redis QuickStart

This tutorial will show you how to use KubeDB to run a Redis server.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/redis/redis-lifecycle.png">
</p>

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY       VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION              AGE
  standard (default)   rancher.io/local-path      Delete          WaitForFirstConsumer           false                      4h
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create namespace demo
  namespace/demo created

  $ kubectl get namespaces
  NAME          STATUS    AGE
  demo          Active    10s
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available RedisVersion

When you have installed KubeDB, it has created `RedisVersion` crd for all supported Redis versions. Check:

```bash
$ kubectl get redisversions
  NAME       VERSION   DB_IMAGE                DEPRECATED   AGE

  4.0.11     4.0.11    kubedb/redis:4.0.11                  31s
  4.0.6-v2   4.0.6     kubedb/redis:4.0.6-v2                31s
  5.0.3-v1   5.0.3     kubedb/redis:5.0.3-v1                31s
  6.0.6      6.0.6     kubedb/redis:6.0.6                   31s
  6.2.5      6.2.5     redis:6.2.5                          31s
```

## Create a Redis server

KubeDB implements a `Redis` CRD to define the specification of a Redis server. Below is the `Redis` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: redis-quickstart
  namespace: demo
spec:
  version: 6.0.6
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/quickstart/demo-1.yaml
redis.kubedb.com/redis-quickstart created
```

Here,

- `spec.version` is name of the RedisVersion crd where the docker images are specified. In this tutorial, a Redis 6.0.6 database is created.
- `spec.storageType` specifies the type of storage that will be used for Redis server. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Redis server using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.storage` specifies PVC spec that will be dynamically allocated to store data for this database. This storage spec will be passed to the StatefulSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.terminationPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Redis` crd or which resources KubeDB should keep or delete when you delete `Redis` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. Learn details of all `TerminationPolicy` [here](/docs/guides/redis/concepts/redis.md#specterminationpolicy)

> Note: spec.storage section is used to create PVC for database pod. It will create PVC with storage size specified instorage.resources.requests field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `Redis` objects using Kubernetes api. When a `Redis` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching Redis object name. KubeDB operator will also create a governing service for StatefulSets with the name `kubedb`, if one is not already present.

```bash
$ kubectl get rd -n demo
NAME               VERSION   STATUS    AGE
redis-quickstart   6.0.6     Running   1m

$ kubectl dba describe rd -n demo redis-quickstart
Name:               redis-quickstart
Namespace:          demo
CreationTimestamp:  Tue, 31 May 2022 10:31:38 +0600
Labels:             <none>
Annotations:        <none>
Replicas:           1  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  DoNotTerminate

StatefulSet:          
  Name:               redis-quickstart
  CreationTimestamp:  Tue, 31 May 2022 10:31:38 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=redis-quickstart
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=redises.kubedb.com
  Annotations:        <none>
  Replicas:           824644335612 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         redis-quickstart
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=redis-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=redises.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.216.57
  Port:         primary  6379/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.58:6379

Service:        
  Name:         redis-quickstart-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=redis-quickstart
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=redises.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  6379/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.58:6379

AppBinding:
  Metadata:
    Creation Timestamp:  2022-05-31T04:31:38Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    redis-quickstart
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        redises.kubedb.com
    Name:                            redis-quickstart
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    redis-quickstart
        Port:    6379
        Scheme:  redis
    Parameters:
      API Version:  config.kubedb.com/v1alpha1
      Kind:         RedisConfiguration
      Stash:
        Addon:
          Backup Task:
            Name:  redis-backup-6.2.5
          Restore Task:
            Name:  redis-restore-6.2.5
    Secret:
      Name:   redis-quickstart-auth
    Type:     kubedb.com/redis
    Version:  6.0.6

Events:
  Type    Reason      Age   From            Message
  ----    ------      ----  ----            -------
  Normal  Successful  2m    Redis Operator  Successfully created governing service
  Normal  Successful  2m    Redis Operator  Successfully created Service
  Normal  Successful  2m    Redis Operator  Successfully created appbinding


$ kubectl get statefulset -n demo
NAME               READY   AGE
redis-quickstart    1/1    1m

$ kubectl get pvc -n demo
NAME                      STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-redis-quickstart-0   Bound     pvc-6e457226-c53f-11e8-9ba7-0800274bef12   1Gi        RWO            standard       2m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                          STORAGECLASS   REASON    AGE
pvc-6e457226-c53f-11e8-9ba7-0800274bef12   1Gi        RWO            Delete           Bound     demo/data-redis-quickstart-0   standard                 2m

$ kubectl get service -n demo
NAME                      TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)    AGE
redis-quickstart-pods   ClusterIP       None             <none>        <none>     2m
redis-quickstart        ClusterIP   10.108.149.205       <none>        6379/TCP   2m
```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified Redis object:

```yaml
$ kubectl get rd -n demo redis-quickstart -o yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  creationTimestamp: "2022-05-31T04:31:38Z"
  finalizers:
    - kubedb.com
  generation: 2
  name: redis-quickstart
  namespace: demo
  resourceVersion: "63624"
  uid: 7ffc9d73-94df-4475-9656-a382f380c293
spec:
  allowedSchemas:
    namespaces:
      from: Same
  authSecret:
    name: redis-quickstart-auth
  coordinator:
    resources: {}
  mode: Standalone
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: redis-quickstart
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: redises.kubedb.com
                namespaces:
                  - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: redis-quickstart
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: redises.kubedb.com
                namespaces:
                  - demo
                topologyKey: failure-domain.beta.kubernetes.io/zone
              weight: 50
      resources:
        limits:
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: redis-quickstart
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
  version: 6.0.6
status:
  conditions:
    - lastTransitionTime: "2022-05-31T04:31:38Z"
      message: 'The KubeDB operator has started the provisioning of Redis: demo/redis-quickstart'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2022-05-31T04:31:43Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2022-05-31T04:31:48Z"
      message: 'The Redis: demo/redis-quickstart is accepting rdClient requests.'
      observedGeneration: 2
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2022-05-31T04:31:48Z"
      message: 'The Redis: demo/redis-quickstart is ready.'
      observedGeneration: 2
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2022-05-31T04:31:48Z"
      message: 'The Redis: demo/redis-quickstart is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 2
  phase: Ready

```

Now, you can connect to this database through [redis-cli](https://redis.io/topics/rediscli). In this tutorial, we are connecting to the Redis server from inside of pod.

```bash
$ kubectl exec -it -n demo redis-quickstart-0 -- sh

/data > redis-cli

127.0.0.1:6379> ping
PONG

#save data
127.0.0.1:6379> SET mykey "Hello"
OK

# view data
127.0.0.1:6379> GET mykey
"Hello"

127.0.0.1:6379> exit

/data > exit
```

## DoNotTerminate Property

When `terminationPolicy` is `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.terminationPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete rd redis-quickstart -n demo
Error from server (BadRequest): admission webhook "redis.validators.kubedb.com" denied the request: redis "redis-quickstart" can't be halted. To delete, change spec.terminationPolicy
```

Now, run `kubectl edit rd redis-quickstart -n demo` to set `spec.terminationPolicy` to `Halt` . Then you will be able to delete/halt the database. 

Learn details of all `TerminationPolicy` [here](/docs/guides/redis/concepts/redis.md#specterminationpolicy)

## Halt Database

When [TerminationPolicy](/docs/guides/redis/concepts/redis.md#specterminationpolicy) is set to `Halt`, it will halt the Redis server instead of deleting it. Here, If you delete the Redis object, KubeDB operator will delete the StatefulSet and its pods but leaves the PVCs unchanged. In KubeDB parlance, we say that `redis-quickstart` Redis server has entered into the dormant state. This is represented by KubeDB operator by creating a matching DormantDatabase object.

```bash
$ kubectl delete rd redis-quickstart -n demo
redis.kubedb.com "redis-quickstart" deleted
```
Check resources:
```bash
NAME                           TYPE                       DATA   AGE
secret/redis-quickstart-auth   kubernetes.io/basic-auth   2      21m

NAME                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-redis-quickstart-0   Bound    pvc-13dce65a-6cc8-4089-9593-139a32ca5134   1Gi        RWO            standard       21m
```

## Resume Redis
Say, the Redis CR was deleted with spec.terminationPolicy to Halt and you want to re-create the Elasticsearch cluster using the existing auth secrets and the PVCs.

You can do it by simply re-deploying the original Redis object:
```bash
kubectl create -f https://github.com/kubedb/docs/raw/v2022.05.24/docs/examples/redis/quickstart/demo-1.yaml
redis.kubedb.com/redis-quickstart created
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo rd/redis-quickstart -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo rd/redis-quickstart

kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `terminationPolicy: WipeOut`**. It is nice to be able to resume database from previous one.So, we preserve all your `PVCs`, auth `Secrets`. If you don't want to resume database, you can just use `spec.terminationPolicy: WipeOut`. It will delete everything created by KubeDB for a particular Redis crd when you delete the crd. For more details about termination policy, please visit [here](/docs/guides/redis/concepts/redis.md#specterminationpolicy).

## Next Steps

- Monitor your Redis server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
