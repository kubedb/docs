---
title: MongoDB Sharding Guide with Arbiter
menu:
  docs_{{ .version }}:
    identifier: mg-arbiter-sharding
    name: Sharding with Arbiter
    parent: mg-arbiter
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Sharding

This tutorial will show you how to use KubeDB to run a sharded MongoDB cluster with arbiter.

## Before You Begin

Before proceeding:

- Read [mongodb arbiter concept](/docs/guides/mongodb/arbiter/concept.md) to get the concept about MongoDB Replica Set Arbiter.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Sharded MongoDB Cluster

To deploy a MongoDB Sharding, user have to specify `spec.shardTopology` option in `Mongodb` CRD.

The following is an example of a `Mongodb` object which creates MongoDB Sharding of three type of members.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mongo-sh-arb
  namespace: demo
spec:
  version: "4.4.26"
  shardTopology:
    configServer:
      replicas: 3
      storage:
        resources:
          requests:
            storage: 500Mi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: mongo
            resources:
              requests:
                cpu: "400m"
                memory: "300Mi"
      shards: 2
      storage:
        resources:
          requests:
            storage: 500Mi
        storageClassName: standard
  arbiter:
    podTemplate:
      spec:
        resources:
          requests:
            cpu: "200m"
            memory: "200Mi"
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/arbiter/sharding.yaml
mongodb.kubedb.com/mongo-sh-arb created
```

Here,

- `spec.shardTopology` represents the topology configuration for sharding.
  - `shard` represents configuration for Shard component of mongodb.
    - `shards` represents number of shards for a mongodb deployment. Each shard is deployed as a [replicaset](/docs/guides/mongodb/clustering/replication_concept.md).
    - `replicas` represents number of replicas of each shard replicaset.
    - `prefix` represents the prefix of each shard node.
    - `configSecret` is an optional field to provide custom configuration file for shards (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used.
    - `podTemplate` is an optional configuration for pods.
    - `storage` to specify pvc spec for each node of sharding. You can specify any StorageClass available in your cluster with appropriate resource requests.
  - `configServer` represents configuration for ConfigServer component of mongodb.
    - `replicas` represents number of replicas for configServer replicaset. Here, configServer is deployed as a replicaset of mongodb.
    - `prefix` represents the prefix of configServer nodes.
    - `configSecret` is an optional field to provide custom configuration file for configSource (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used.
    - `podTemplate` is an optional configuration for pods.
    - `storage` to specify pvc spec for each node of configServer. You can specify any StorageClass available in your cluster with appropriate resource requests.
  - `mongos` represents configuration for Mongos component of mongodb. `Mongos` instances run as stateless components (deployment).
    - `replicas` represents number of replicas of `Mongos` instance. Here, Mongos is not deployed as replicaset.
    - `prefix` represents the prefix of mongos nodes.
    - `configSecret` is an optional field to provide custom configuration file for mongos (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise a default configuration file will be used.
    - `podTemplate` is an optional configuration for pods.
- `spec.keyFileSecret` (optional) is a secret name that contains keyfile (a random string)against `key.txt` key. Each mongod instances in the replica set and `shardTopology` uses the contents of the keyfile as the shared password for authenticating other members in the replicaset. Only mongod instances with the correct keyfile can join the replica set. _User can provide the `keyFileSecret` by creating a secret with key `key.txt`. See [here](https://docs.mongodb.com/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/#create-a-keyfile) to create the string for `keyFileSecret`._ If `keyFileSecret` is not given, KubeDB operator will generate a `keyFileSecret` itself.
- `spec.arbiter` denotes arbiter spec of the deployed MongoDB CRD. There are two fields under it : configSecret & podTemplate. `spec.arbiter.configSecret` is an optional field to provide custom configuration file for database (i.e mongod.cnf). If specified, this file will be used as configuration file otherwise default configuration file will be used. `spec.arbiter.podTemplate` holds the arbiter-podSpec. `null` value of it, instructs kubedb operator to use  the default arbiter podTemplate.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create some new PetSets : 1 for mongos, 1 for configServer, and 1 for each of the shard & arbiter. It creates a primary Service with the matching MongoDB object name. KubeDB operator will also create governing services for PetSets with the name `<mongodb-name>-<node-type>-pods`.

MongoDB `mongo-sh-arb` state,

```bash
$ kubectl get mg -n demo
NAME                              VERSION   STATUS   AGE
mongodb.kubedb.com/mongo-sh-arb   4.4.26     Ready    97s
```

All the types of nodes `Shard`, `ConfigServer` & `Mongos` are deployed as petset.

```bash
$ kubectl get petset -n demo
NAME                                           READY   AGE
petset.apps/mongo-sh-arb-configsvr        3/3     97s
petset.apps/mongo-sh-arb-mongos           2/2     29s
petset.apps/mongo-sh-arb-shard0           2/2     97s
petset.apps/mongo-sh-arb-shard0-arbiter   1/1     53s
petset.apps/mongo-sh-arb-shard1           2/2     97s
petset.apps/mongo-sh-arb-shard1-arbiter   1/1     52s
```

All PVCs and PVs for MongoDB `mongo-sh-arb`,

```bash
$ kubectl get pvc -n demo
NAME                                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mongo-sh-arb-configsvr-0        Bound    pvc-a9589ccb-24c2-4d17-8174-1e552d63d943   500Mi      RWO            standard       97s
persistentvolumeclaim/datadir-mongo-sh-arb-configsvr-1        Bound    pvc-697aa035-6ff2-45c4-8e00-0787b520159b   500Mi      RWO            standard       75s
persistentvolumeclaim/datadir-mongo-sh-arb-configsvr-2        Bound    pvc-2548ee7e-5416-4ddc-960b-33d17bd53b43   500Mi      RWO            standard       52s
persistentvolumeclaim/datadir-mongo-sh-arb-shard0-0           Bound    pvc-a5cdb597-ad01-4362-b56e-c5d6226a38bb   500Mi      RWO            standard       97s
persistentvolumeclaim/datadir-mongo-sh-arb-shard0-1           Bound    pvc-ae9e594a-7370-4339-9f51-6ec07588c8e0   500Mi      RWO            standard       75s
persistentvolumeclaim/datadir-mongo-sh-arb-shard0-arbiter-0   Bound    pvc-8296c2bc-dfc0-47f4-b651-01fb802bf751   500Mi      RWO            standard       53s
persistentvolumeclaim/datadir-mongo-sh-arb-shard1-0           Bound    pvc-33cde211-4ed5-49a9-b7a8-48e94690e12d   500Mi      RWO            standard       97s
persistentvolumeclaim/datadir-mongo-sh-arb-shard1-1           Bound    pvc-569cedf8-b16e-4616-ae1d-74168aacc227   500Mi      RWO            standard       74s
persistentvolumeclaim/datadir-mongo-sh-arb-shard1-arbiter-0   Bound    pvc-c65c7054-a9de-40c4-9797-4d0a730e9c5b   500Mi      RWO            standard       52s

$ kubectl get pv -n demo
NAME                                                        CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                        STORAGECLASS   REASON   AGE
persistentvolume/pvc-2548ee7e-5416-4ddc-960b-33d17bd53b43   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-configsvr-2        standard                50s
persistentvolume/pvc-33cde211-4ed5-49a9-b7a8-48e94690e12d   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-shard1-0           standard                93s
persistentvolume/pvc-569cedf8-b16e-4616-ae1d-74168aacc227   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-shard1-1           standard                71s
persistentvolume/pvc-697aa035-6ff2-45c4-8e00-0787b520159b   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-configsvr-1        standard                73s
persistentvolume/pvc-8296c2bc-dfc0-47f4-b651-01fb802bf751   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-shard0-arbiter-0   standard                52s
persistentvolume/pvc-a5cdb597-ad01-4362-b56e-c5d6226a38bb   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-shard0-0           standard                94s
persistentvolume/pvc-a9589ccb-24c2-4d17-8174-1e552d63d943   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-configsvr-0        standard                94s
persistentvolume/pvc-ae9e594a-7370-4339-9f51-6ec07588c8e0   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-shard0-1           standard                73s
persistentvolume/pvc-c65c7054-a9de-40c4-9797-4d0a730e9c5b   500Mi      RWO            Delete           Bound    demo/datadir-mongo-sh-arb-shard1-arbiter-0   standard                49s
```

Services created for MongoDB `mongo-sh-arb`

```bash
$ kubectl get svc -n demo
NAME                                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
service/mongo-sh-arb                  ClusterIP   10.96.34.129   <none>        27017/TCP   97s
service/mongo-sh-arb-configsvr-pods   ClusterIP   None           <none>        27017/TCP   97s
service/mongo-sh-arb-mongos-pods      ClusterIP   None           <none>        27017/TCP   97s
service/mongo-sh-arb-shard0-pods      ClusterIP   None           <none>        27017/TCP   97s
service/mongo-sh-arb-shard1-pods      ClusterIP   None           <none>        27017/TCP   97s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. It has also defaulted some field of crd object. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mongo-sh-arb -o yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MongoDB","metadata":{"annotations":{},"name":"mongo-sh-arb","namespace":"demo"},"spec":{"arbiter":{"podTemplate":{"spec":{"requests":{"cpu":"200m","memory":"200Mi"},"resources":null}}},"shardTopology":{"configServer":{"replicas":3,"storage":{"resources":{"requests":{"storage":"500Mi"}},"storageClassName":"standard"}},"mongos":{"replicas":2},"shard":{"podTemplate":{"spec":{"resources":{"requests":{"cpu":"400m","memory":"300Mi"}}}},"replicas":2,"shards":2,"storage":{"resources":{"requests":{"storage":"500Mi"}},"storageClassName":"standard"}}},"deletionPolicy":"WipeOut","version":"4.4.26"}}
  creationTimestamp: "2022-04-21T09:29:07Z"
  finalizers:
  - kubedb.com
  generation: 3
  name: mongo-sh-arb
  namespace: demo
  resourceVersion: "31916"
  uid: 0a31ab30-0002-400e-a312-f7e343ec6894
spec:
  allowedSchemas:
    namespaces:
      from: Same
  arbiter:
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
                    app.kubernetes.io/instance: mongo-sh-arb
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                    mongodb.kubedb.com/node.shard: mongo-sh-arb-shard${SHARD_INDEX}
                namespaces:
                - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mongo-sh-arb
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                    mongodb.kubedb.com/node.shard: mongo-sh-arb-shard${SHARD_INDEX}
                namespaces:
                - demo
                topologyKey: failure-domain.beta.kubernetes.io/zone
              weight: 50
        livenessProbe:
          exec:
            command:
            - bash
            - -c
            - "set -x; if [[ $(mongo admin --host=localhost   --quiet --eval \"db.adminCommand('ping').ok\"
              ) -eq \"1\" ]]; then \n          exit 0\n        fi\n        exit 1"
          failureThreshold: 3
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 5
        readinessProbe:
          exec:
            command:
            - bash
            - -c
            - "set -x; if [[ $(mongo admin --host=localhost   --quiet --eval \"db.adminCommand('ping').ok\"
              ) -eq \"1\" ]]; then \n          exit 0\n        fi\n        exit 1"
          failureThreshold: 3
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 500m
            memory: 1Gi
  authSecret:
    name: mongo-sh-arb-auth
  clusterAuthMode: keyFile
  coordinator:
    resources: {}
  keyFileSecret:
    name: mongo-sh-arb-key
  shardTopology:
    configServer:
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
                      app.kubernetes.io/instance: mongo-sh-arb
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                      mongodb.kubedb.com/node.config: mongo-sh-arb-configsvr
                  namespaces:
                  - demo
                  topologyKey: kubernetes.io/hostname
                weight: 100
              - podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      app.kubernetes.io/instance: mongo-sh-arb
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                      mongodb.kubedb.com/node.config: mongo-sh-arb-configsvr
                  namespaces:
                  - demo
                  topologyKey: failure-domain.beta.kubernetes.io/zone
                weight: 50
          livenessProbe:
            exec:
              command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
                --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
                --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then
                \n          exit 0\n        fi\n        exit 1"
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 5
          readinessProbe:
            exec:
              command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
                --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
                --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then
                \n          exit 0\n        fi\n        exit 1"
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
          serviceAccountName: mongo-sh-arb
      replicas: 3
      storage:
        resources:
          requests:
            storage: 500Mi
        storageClassName: standard
    mongos:
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
                      app.kubernetes.io/instance: mongo-sh-arb
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                      mongodb.kubedb.com/node.mongos: mongo-sh-arb-mongos
                  namespaces:
                  - demo
                  topologyKey: kubernetes.io/hostname
                weight: 100
              - podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      app.kubernetes.io/instance: mongo-sh-arb
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                      mongodb.kubedb.com/node.mongos: mongo-sh-arb-mongos
                  namespaces:
                  - demo
                  topologyKey: failure-domain.beta.kubernetes.io/zone
                weight: 50
          lifecycle:
            preStop:
              exec:
                command:
                - bash
                - -c
                - 'mongo admin --username=$MONGO_INITDB_ROOT_USERNAME --password=$MONGO_INITDB_ROOT_PASSWORD
                  --quiet --eval "db.adminCommand({ shutdown: 1 })" || true'
          livenessProbe:
            exec:
              command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
                --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
                --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then
                \n          exit 0\n        fi\n        exit 1"
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 5
          readinessProbe:
            exec:
              command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
                --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
                --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then
                \n          exit 0\n        fi\n        exit 1"
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              memory: 1Gi
            requests:
              cpu: 500m
              memory: 1Gi
          serviceAccountName: mongo-sh-arb
      replicas: 2
    shard:
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
                      app.kubernetes.io/instance: mongo-sh-arb
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                      mongodb.kubedb.com/node.shard: mongo-sh-arb-shard${SHARD_INDEX}
                  namespaces:
                  - demo
                  topologyKey: kubernetes.io/hostname
                weight: 100
              - podAffinityTerm:
                  labelSelector:
                    matchLabels:
                      app.kubernetes.io/instance: mongo-sh-arb
                      app.kubernetes.io/managed-by: kubedb.com
                      app.kubernetes.io/name: mongodbs.kubedb.com
                      mongodb.kubedb.com/node.shard: mongo-sh-arb-shard${SHARD_INDEX}
                  namespaces:
                  - demo
                  topologyKey: failure-domain.beta.kubernetes.io/zone
                weight: 50
          livenessProbe:
            exec:
              command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
                --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
                --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then
                \n          exit 0\n        fi\n        exit 1"
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 5
          readinessProbe:
            exec:
              command:
              - bash
              - -c
              - "set -x; if [[ $(mongo admin --host=localhost  --username=$MONGO_INITDB_ROOT_USERNAME
                --password=$MONGO_INITDB_ROOT_PASSWORD --authenticationDatabase=admin
                --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then
                \n          exit 0\n        fi\n        exit 1"
            failureThreshold: 3
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              memory: 300Mi
            requests:
              cpu: 400m
              memory: 300Mi
          serviceAccountName: mongo-sh-arb
      replicas: 2
      shards: 2
      storage:
        resources:
          requests:
            storage: 500Mi
        storageClassName: standard
  sslMode: disabled
  storageEngine: wiredTiger
  storageType: Durable
  deletionPolicy: WipeOut
  version: 4.4.26
status:
  conditions:
  - lastTransitionTime: "2022-04-21T09:29:07Z"
    message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mongo-sh-arb'
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2022-04-21T09:30:39Z"
    message: All desired replicas are ready.
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2022-04-21T09:30:37Z"
    message: 'The MongoDB: demo/mongo-sh-arb is accepting client requests.'
    observedGeneration: 3
    reason: DatabaseAcceptingConnectionRequest
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2022-04-21T09:30:37Z"
    message: 'The MongoDB: demo/mongo-sh-arb is ready.'
    observedGeneration: 3
    reason: ReadinessCheckSucceeded
    status: "True"
    type: Ready
  - lastTransitionTime: "2022-04-21T09:30:39Z"
    message: 'The MongoDB: demo/mongo-sh-arb is successfully provisioned.'
    observedGeneration: 3
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  observedGeneration: 3
  phase: Ready

```

Please note that KubeDB operator has created a new Secret called `mongo-sh-arb-auth` _(format: {mongodb-object-name}-auth)_ for storing the password for `mongodb` superuser. This secret contains a `username` key which contains the _username_ for MongoDB superuser and a `password` key which contains the _password_ for MongoDB superuser.

If you want to use custom or existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`. For more details, please see [here](/docs/guides/mongodb/concepts/mongodb.md#specauthsecret).

## Connection Information

- Hostname/address: you can use any of these
  - Service: `mongo-sh-arb.demo`
  - Pod IP: (`$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-arb-mongos -o yaml | grep podIP`)
- Port: `27017`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo mongo-sh-arb-auth -o jsonpath='{.data.\username}' | base64 -d
  root
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo mongo-sh-arb-auth -o jsonpath='{.data.\password}' | base64 -d
  6&UiN5;qq)Tnai=7
  ```

Now, you can connect to this database through [mongo-shell](https://docs.mongodb.com/v4.2/mongo/).

## Sharded Data

In this tutorial, we will insert sharded and unsharded document, and we will see if the data actually sharded across cluster or not.

```bash
$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-arb-mongos
NAME                    READY   STATUS    RESTARTS   AGE
mongo-sh-arb-mongos-0   1/1     Running   0          6m34s
mongo-sh-arb-mongos-1   1/1     Running   0          6m20s

$ kubectl exec -it mongo-sh-arb-mongos-0 -n demo bash

mongodb@mongo-sh-mongos-0:/$ mongo admin -u root -p '6&UiN5;qq)Tnai=7'
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("bf87addd-4245-45b1-a470-fabb3dcc19ab") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
---
The server generated these startup warnings when booting: 
        2022-04-21T09:30:28.259+00:00: You are running this process as the root user, which is not recommended
---
mongos>
```

To detect if the MongoDB instance that your client is connected to is mongos, use the isMaster command. When a client connects to a mongos, isMaster returns a document with a `msg` field that holds the string `isdbgrid`.

```bash
mongos> rs.isMaster()
{
	"ismaster" : true,
	"msg" : "isdbgrid",
	"maxBsonObjectSize" : 16777216,
	"maxMessageSizeBytes" : 48000000,
	"maxWriteBatchSize" : 100000,
	"localTime" : ISODate("2022-04-21T09:38:52.370Z"),
	"logicalSessionTimeoutMinutes" : 30,
	"connectionId" : 253,
	"maxWireVersion" : 9,
	"minWireVersion" : 0,
	"topologyVersion" : {
		"processId" : ObjectId("62612434ea3cf5a7339dd36d"),
		"counter" : NumberLong(0)
	},
	"ok" : 1,
	"operationTime" : Timestamp(1650533931, 30),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1650533931, 30),
		"signature" : {
			"hash" : BinData(0,"QhqwrAXFhPjlpvfTOPwNAESUR8c="),
			"keyId" : NumberLong("7088986810746929174")
		}
	}
}
```

`mongo-sh-arb` Shard status,

```bash
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("626123f2f1e4f6821ec73945")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-arb-shard0-0.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard0-1.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-arb-shard1-0.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard1-1.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  yes
        Collections with active migrations: 
                config.system.sessions started at Thu Apr 21 2022 09:39:13 GMT+0000 (UTC)
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                279 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	745
                                shard1	279
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("79db6e4a-dcb1-4f1a-86c2-dcd86a944893"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
```






As `sh.status()` command only shows the data bearing members, if we want to assure that arbiter has been added correctly we need to exec into any shard-pod & run `rs.status()` command against the admin database. Open another terminal : 


```bash
kubectl exec -it pod/mongo-sh-arb-shard0-1 -n demo bash

root@mongo-sh-arb-shard0-1:/ mongo admin -u root -p '6&UiN5;qq)Tnai=7'
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.

shard0:PRIMARY> rs.status().members
[
	{
		"_id" : 0,
		"name" : "mongo-sh-arb-shard0-0.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 2,
		"stateStr" : "SECONDARY",
		"uptime" : 350,
		"optime" : {
			"ts" : Timestamp(1650535338, 18),
			"t" : NumberLong(3)
		},
		"optimeDurable" : {
			"ts" : Timestamp(1650535338, 18),
			"t" : NumberLong(3)
		},
		"optimeDate" : ISODate("2022-04-21T10:02:18Z"),
		"optimeDurableDate" : ISODate("2022-04-21T10:02:18Z"),
		"lastHeartbeat" : ISODate("2022-04-21T10:02:35.951Z"),
		"lastHeartbeatRecv" : ISODate("2022-04-21T10:02:34.999Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncSourceHost" : "mongo-sh-arb-shard0-1.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",
		"syncSourceId" : 1,
		"infoMessage" : "",
		"configVersion" : 4,
		"configTerm" : 3
	},
	{
		"_id" : 1,
		"name" : "mongo-sh-arb-shard0-1.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 1,
		"stateStr" : "PRIMARY",
		"uptime" : 352,
		"optime" : {
			"ts" : Timestamp(1650535338, 18),
			"t" : NumberLong(3)
		},
		"optimeDate" : ISODate("2022-04-21T10:02:18Z"),
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"electionTime" : Timestamp(1650535017, 1),
		"electionDate" : ISODate("2022-04-21T09:56:57Z"),
		"configVersion" : 4,
		"configTerm" : 3,
		"self" : true,
		"lastHeartbeatMessage" : ""
	},
	{
		"_id" : 2,
		"name" : "mongo-sh-arb-shard0-arbiter-0.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",
		"health" : 1,
		"state" : 7,
		"stateStr" : "ARBITER",
		"uptime" : 328,
		"lastHeartbeat" : ISODate("2022-04-21T10:02:35.950Z"),
		"lastHeartbeatRecv" : ISODate("2022-04-21T10:02:35.585Z"),
		"pingMs" : NumberLong(0),
		"lastHeartbeatMessage" : "",
		"syncSourceHost" : "",
		"syncSourceId" : -1,
		"infoMessage" : "",
		"configVersion" : 4,
		"configTerm" : 3
	}
]
```

Enable sharding to collection `songs.list` and insert document. See [`sh.shardCollection(namespace, key, unique, options)`](https://docs.mongodb.com/manual/reference/method/sh.shardCollection/#sh.shardCollection) for details about `shardCollection` command.

```bash
mongos> sh.enableSharding("songs");
{
	"ok" : 1,
	"operationTime" : Timestamp(1650534119, 40),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1650534119, 40),
		"signature" : {
			"hash" : BinData(0,"vtfzghRf+pGMDwsY/W3y/irgF1s="),
			"keyId" : NumberLong("7088986810746929174")
		}
	}
}

mongos> sh.shardCollection("songs.list", {"myfield": 1});
{
	"collectionsharded" : "songs.list",
	"collectionUUID" : UUID("320eccb3-1987-4ac9-affb-61fe2b9284a7"),
	"ok" : 1,
	"operationTime" : Timestamp(1650534144, 45),
	"$clusterTime" : {
		"clusterTime" : Timestamp(1650534144, 45),
		"signature" : {
			"hash" : BinData(0,"F6KJ8uibwEmuoAi4YPvLYFR71eg="),
			"keyId" : NumberLong("7088986810746929174")
		}
	}
}

mongos> use songs
switched to db songs

mongos> db.list.insert({"led zeppelin": "stairway to heaven", "slipknot": "psychosocial"});
WriteResult({ "nInserted" : 1 })

mongos> db.list.insert({"pink floyd": "us and them", "nirvana": "smells like teen spirit", "john lennon" : "imagine" });
WriteResult({ "nInserted" : 1 })

mongos> db.list.find()
{ "_id" : ObjectId("6261275c18807d1843328e08"), "led zeppelin" : "stairway to heaven", "slipknot" : "psychosocial" }
{ "_id" : ObjectId("6261281c18807d1843328e09"), "pink floyd" : "us and them", "nirvana" : "smells like teen spirit", "john lennon" : "imagine" }
```

Run [`sh.status()`](https://docs.mongodb.com/manual/reference/method/sh.status/) to see whether the `songs` database has sharding enabled, and the primary shard for the `songs` database.

The Sharded Collection section `sh.status.databases.<collection>` provides information on the sharding details for sharded collection(s) (E.g. `songs.list`). For each sharded collection, the section displays the shard key, the number of chunks per shard(s), the distribution of documents across chunks, and the tag information, if any, for shard key range(s).

```bash
mongos> sh.status();
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("626123f2f1e4f6821ec73945")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-arb-shard0-0.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard0-1.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-arb-shard1-0.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard1-1.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                512 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	512
                                shard1	512
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("79db6e4a-dcb1-4f1a-86c2-dcd86a944893"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
        {  "_id" : "songs",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("5a61681f-e427-463f-85ca-c1f0d8854a3b"),  "lastMod" : 1 } }
                songs.list
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0) 
```

Now create another database where partiotioned is not applied and see how the data is stored.

```bash
mongos> use demo
switched to db demo

mongos> db.anothercollection.insert({"myfield": "ccc", "otherfield": "this is non sharded", "kube" : "db" });
WriteResult({ "nInserted" : 1 })

mongos> db.anothercollection.insert({"myfield": "aaa", "more": "field" });
WriteResult({ "nInserted" : 1 })


mongos> db.anothercollection.find()
{ "_id" : ObjectId("626128f618807d1843328e0a"), "myfield" : "ccc", "otherfield" : "this is non sharded", "kube" : "db" }
{ "_id" : ObjectId("6261293c18807d1843328e0b"), "myfield" : "aaa", "more" : "field" }
```

Now, eventually `sh.status()`

```
mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("626123f2f1e4f6821ec73945")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-arb-shard0-0.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard0-1.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-arb-shard1-0.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard1-1.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  0
        Migration Results for the last 24 hours: 
                512 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	512
                                shard1	512
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "demo",  "primary" : "shard1",  "partitioned" : false,  "version" : {  "uuid" : UUID("8af87f8c-b4ae-4d04-854f-d2ede7465acd"),  "lastMod" : 1 } }
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("79db6e4a-dcb1-4f1a-86c2-dcd86a944893"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
        {  "_id" : "songs",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("5a61681f-e427-463f-85ca-c1f0d8854a3b"),  "lastMod" : 1 } }
                songs.list
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0) 
```

Here, `demo` database is not partitioned and all collections under `demo` database are stored in it's primary shard, which is `shard0`.

## Halt Database

When [DeletionPolicy](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy) is set to halt, and you delete the mongodb object, the KubeDB operator will delete the PetSet and its pods but leaves the PVCs, secrets and database backup (snapshots) intact. Learn details of all `DeletionPolicy` [here](/docs/guides/mongodb/concepts/mongodb.md#specdeletionpolicy).

You can also keep the mongodb object and halt the database to resume it again later. If you halt the database, the kubedb will delete the petsets and services but will keep the mongodb object, pvcs, secrets and backup (snapshots).

To halt the database, first you have to set the deletionPolicy to `Halt` in existing database. You can use the below command to set the deletionPolicy to `Halt`, if it is not already set.

```bash
$ kubectl patch -n demo mg/mongo-sh-arb -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"
mongodb.kubedb.com/mongo-sh-arb patched
```

Then, you have to set the `spec.halted` as true to set the database in a `Halted` state. You can use the below command.

```bash
$ kubectl patch -n demo mg/mongo-sh-arb -p '{"spec":{"halted":true}}' --type="merge"
mongodb.kubedb.com/mongo-sh-arb patched
```

After that, kubedb will delete the petsets and services and you can see the database Phase as `Halted`.

Now, you can run the following command to get all mongodb resources in demo namespaces,

```bash
$ kubectl get mg,sts,svc,secret,pvc -n demo
NAME                              VERSION   STATUS   AGE
mongodb.kubedb.com/mongo-sh-arb   4.4.26     Halted   26m

NAME                         TYPE                                  DATA   AGE
secret/default-token-bg2wb   kubernetes.io/service-account-token   3      26m
secret/mongo-sh-arb-auth     Opaque                                2      26m
secret/mongo-sh-arb-key      Opaque                                1      26m

NAME                                                          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/datadir-mongo-sh-arb-configsvr-0        Bound    pvc-a9589ccb-24c2-4d17-8174-1e552d63d943   500Mi      RWO            standard       26m
persistentvolumeclaim/datadir-mongo-sh-arb-configsvr-1        Bound    pvc-697aa035-6ff2-45c4-8e00-0787b520159b   500Mi      RWO            standard       26m
persistentvolumeclaim/datadir-mongo-sh-arb-configsvr-2        Bound    pvc-2548ee7e-5416-4ddc-960b-33d17bd53b43   500Mi      RWO            standard       25m
persistentvolumeclaim/datadir-mongo-sh-arb-shard0-0           Bound    pvc-a5cdb597-ad01-4362-b56e-c5d6226a38bb   500Mi      RWO            standard       26m
persistentvolumeclaim/datadir-mongo-sh-arb-shard0-1           Bound    pvc-ae9e594a-7370-4339-9f51-6ec07588c8e0   500Mi      RWO            standard       26m
persistentvolumeclaim/datadir-mongo-sh-arb-shard0-arbiter-0   Bound    pvc-8296c2bc-dfc0-47f4-b651-01fb802bf751   500Mi      RWO            standard       25m
persistentvolumeclaim/datadir-mongo-sh-arb-shard1-0           Bound    pvc-33cde211-4ed5-49a9-b7a8-48e94690e12d   500Mi      RWO            standard       26m
persistentvolumeclaim/datadir-mongo-sh-arb-shard1-1           Bound    pvc-569cedf8-b16e-4616-ae1d-74168aacc227   500Mi      RWO            standard       26m
persistentvolumeclaim/datadir-mongo-sh-arb-shard1-arbiter-0   Bound    pvc-c65c7054-a9de-40c4-9797-4d0a730e9c5b   500Mi      RWO            standard       25m
```

From the above output, you can see that MongoDB object, PVCs, Secret are still there.

## Resume Halted Database

Now, to resume the database, i.e. to get the same database setup back again, you have to set the the `spec.halted` as false. You can use the below command.

```bash
$ kubectl patch -n demo mg/mongo-sh-arb -p '{"spec":{"halted":false}}' --type="merge"
mongodb.kubedb.com/mongo-sh-arb patched
```

When the database is resumed successfully, you can see the database Status is set to `Ready`.

```bash
$ kubectl get mg -n demo
NAME                              VERSION   STATUS   AGE
mongodb.kubedb.com/mongo-sh-arb   4.4.26     Ready    28m
```

Now, If you again exec into `pod` and look for previous data, you will see that, all the data persists.

```bash
$ kubectl get po -n demo -l mongodb.kubedb.com/node.mongos=mongo-sh-arb-mongos
NAME                    READY   STATUS    RESTARTS   AGE
mongo-sh-arb-mongos-0   1/1     Running   0          89s
mongo-sh-arb-mongos-1   1/1     Running   0          29s


$ kubectl exec -it mongo-sh-arb-mongos-0 -n demo bash

mongodb@mongo-sh-mongos-0:/$ mongo admin -u root -p '6&UiN5;qq)Tnai=7'

mongos> use songs
switched to db songs

mongos> db.list.find()
{ "_id" : ObjectId("6261275c18807d1843328e08"), "led zeppelin" : "stairway to heaven", "slipknot" : "psychosocial" }
{ "_id" : ObjectId("6261281c18807d1843328e09"), "pink floyd" : "us and them", "nirvana" : "smells like teen spirit", "john lennon" : "imagine" }

mongos> sh.status()
--- Sharding Status --- 
  sharding version: {
  	"_id" : 1,
  	"minCompatibleVersion" : 5,
  	"currentVersion" : 6,
  	"clusterId" : ObjectId("626123f2f1e4f6821ec73945")
  }
  shards:
        {  "_id" : "shard0",  "host" : "shard0/mongo-sh-arb-shard0-0.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard0-1.mongo-sh-arb-shard0-pods.demo.svc.cluster.local:27017",  "state" : 1 }
        {  "_id" : "shard1",  "host" : "shard1/mongo-sh-arb-shard1-0.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017,mongo-sh-arb-shard1-1.mongo-sh-arb-shard1-pods.demo.svc.cluster.local:27017",  "state" : 1 }
  active mongoses:
        "4.4.26" : 2
  autosplit:
        Currently enabled: yes
  balancer:
        Currently enabled:  yes
        Currently running:  no
        Failed balancer rounds in last 5 attempts:  1
        Last reported error:  Could not find host matching read preference { mode: "primary" } for set shard0
        Time of Reported error:  Thu Apr 21 2022 09:57:04 GMT+0000 (UTC)
        Migration Results for the last 24 hours: 
                512 : Success
  databases:
        {  "_id" : "config",  "primary" : "config",  "partitioned" : true }
                config.system.sessions
                        shard key: { "_id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	512
                                shard1	512
                        too many chunks to print, use verbose if you want to force print
        {  "_id" : "demo",  "primary" : "shard1",  "partitioned" : false,  "version" : {  "uuid" : UUID("8af87f8c-b4ae-4d04-854f-d2ede7465acd"),  "lastMod" : 1 } }
        {  "_id" : "kubedb-system",  "primary" : "shard0",  "partitioned" : true,  "version" : {  "uuid" : UUID("79db6e4a-dcb1-4f1a-86c2-dcd86a944893"),  "lastMod" : 1 } }
                kubedb-system.health-check
                        shard key: { "id" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard0	1
                        { "id" : { "$minKey" : 1 } } -->> { "id" : { "$maxKey" : 1 } } on : shard0 Timestamp(1, 0) 
        {  "_id" : "songs",  "primary" : "shard1",  "partitioned" : true,  "version" : {  "uuid" : UUID("5a61681f-e427-463f-85ca-c1f0d8854a3b"),  "lastMod" : 1 } }
                songs.list
                        shard key: { "myfield" : 1 }
                        unique: false
                        balancing: true
                        chunks:
                                shard1	1
                        { "myfield" : { "$minKey" : 1 } } -->> { "myfield" : { "$maxKey" : 1 } } on : shard1 Timestamp(1, 0) 
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mongo-sh-arb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mongo-sh-arb

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) process of MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
