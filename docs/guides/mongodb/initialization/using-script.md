---
title: Initialize MongoDB using Script
menu:
  docs_{{ .version }}:
    identifier: mg-using-script-initialization
    name: Using Script
    parent: mg-initialization-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize MongoDB using Script

This tutorial will show you how to use KubeDB to initialize a MongoDB database with .js and/or .sh script.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

  In this tutorial we will use .js script stored in GitHub repository [kubedb/mongodb-init-scripts](https://github.com/kubedb/mongodb-init-scripts).

> Note: The yaml files used in this tutorial are stored in [docs/examples/mongodb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/mongodb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Initialization Scripts

MongoDB supports initialization with `.sh` and `.js` files. In this tutorial, we will use `init.js` script from [mongodb-init-scripts](https://github.com/kubedb/mongodb-init-scripts) git repository to insert data inside `kubedb` DB.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from `init.js` file. Then, we will provide this ConfigMap as script source in `init.script` of MongoDB crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo mg-init-script \
--from-literal=init.js="$(curl -fsSL https://github.com/kubedb/mongodb-init-scripts/raw/master/init.js)"
configmap/mg-init-script created
```

## Create a MongoDB database with Init-Script

Below is the `MongoDB` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mgo-init-script
  namespace: demo
spec:
  version: "4.4.26"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: mg-init-script
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/Initialization/replicaset.yaml
mongodb.kubedb.com/mgo-init-script created
```

Here,

- `spec.init.script` specifies a script source used to initialize the database before database server starts. The scripts will be executed alphabatically. In this tutorial, a sample .js script from the git repository `https://github.com/kubedb/mongodb-init-scripts.git` is used to create a test database. You can use other [volume sources](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes).  The \*.js and/or \*.sh sripts that are stored inside the root folder will be executed alphabatically. The scripts inside child folders will be skipped.

KubeDB operator watches for `MongoDB` objects using Kubernetes api. When a `MongoDB` object is created, KubeDB operator will create a new StatefulSet and a Service with the matching MongoDB object name. KubeDB operator will also create a governing service for StatefulSets with the name `<mongodb-crd-name>-gvr`, if one is not already present. No MongoDB specific RBAC roles are required for [RBAC enabled clusters](/docs/setup/README.md#using-yaml).

```bash
$ kubectl dba describe mg -n demo mgo-init-script
Name:               mgo-init-script
Namespace:          demo
CreationTimestamp:  Thu, 11 Feb 2021 10:58:22 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-init-script","namespace":"demo"},"spec":{"init":{"script"...
Replicas:           1  total
Status:             Ready
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  Delete

StatefulSet:          
  Name:               mgo-init-script
  CreationTimestamp:  Thu, 11 Feb 2021 10:58:23 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=mgo-init-script
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:        <none>
  Replicas:           824638316568 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         mgo-init-script
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-init-script
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.107.34.91
  Port:         primary  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    [10.107.34.91]:27017

Service:        
  Name:         mgo-init-script-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-init-script
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    [10.107.34.91]:27017

Auth Secret:
  Name:         mgo-init-script-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=mgo-init-script
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=mongodbs.kubedb.com
  Annotations:  <none>
  Type:         Opaque
  Data:
    password:  16 bytes
    username:  4 bytes

Init:
  Script Source:
    Volume:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      mg-init-script
    Optional:  false

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"mg-init-script"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"4.4.26"}}

    Creation Timestamp:  2021-02-11T04:58:42Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    mgo-init-script
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        mongodbs.kubedb.com
    Name:                            mgo-init-script
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    mgo-init-script
        Port:    27017
        Scheme:  mongodb
    Secret:
      Name:   mgo-init-script-auth
    Type:     kubedb.com/mongodb
    Version:  4.4.26

Events:
  Type    Reason      Age   From              Message
  ----    ------      ----  ----              -------
  Normal  Successful  47s   MongoDB operator  Successfully created stats service
  Normal  Successful  47s   MongoDB operator  Successfully created Service
  Normal  Successful  46s   MongoDB operator  Successfully  stats service
  Normal  Successful  46s   MongoDB operator  Successfully  stats service
  Normal  Successful  27s   MongoDB operator  Successfully created appbinding
  Normal  Successful  27s   MongoDB operator  Successfully patched StatefulSet demo/mgo-init-script
  Normal  Successful  27s   MongoDB operator  Successfully patched MongoDB

$ kubectl get statefulset -n demo
NAME              READY   AGE
mgo-init-script   1/1     30s

$ kubectl get pvc -n demo
NAME                        STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
datadir-mgo-init-script-0   Bound     pvc-a10d636b-c08c-11e8-b4a9-0800272618ed   1Gi        RWO            standard       11m

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS    CLAIM                            STORAGECLASS   REASON    AGE
pvc-a10d636b-c08c-11e8-b4a9-0800272618ed   1Gi        RWO            Delete           Bound     demo/datadir-mgo-init-script-0   standard                 12m

$ kubectl get service -n demo
NAME                  TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
mgo-init-script       ClusterIP   10.107.34.91   <none>        27017/TCP   52s
mgo-init-script-pods  ClusterIP   None           <none>        27017/TCP   52s
```

KubeDB operator sets the `status.phase` to `Ready` once the database is successfully created. Run the following command to see the modified MongoDB object:

```yaml
$ kubectl get mg -n demo mgo-init-script -o yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"mgo-init-script","namespace":"demo"},"spec":{"init":{"script":{"configMap":{"name":"mg-init-script"}}},"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"version":"4.4.26"}}
  creationTimestamp: "2021-02-10T04:38:52Z"
  finalizers:
    - kubedb.com
  generation: 3
  managedFields:
    - apiVersion: kubedb.com/v1alpha2
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:kubectl.kubernetes.io/last-applied-configuration: {}
        f:spec:
          .: {}
          f:init:
            .: {}
            f:script:
              .: {}
              f:configMap:
                .: {}
                f:name: {}
          f:storage:
            .: {}
            f:accessModes: {}
            f:resources:
              .: {}
              f:requests:
                .: {}
                f:storage: {}
            f:storageClassName: {}
          f:version: {}
      manager: kubectl-client-side-apply
      operation: Update
      time: "2021-02-10T04:38:52Z"
    - apiVersion: kubedb.com/v1alpha2
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers: {}
        f:spec:
          f:authSecret:
            .: {}
            f:name: {}
          f:init:
            f:initialized: {}
        f:status:
          .: {}
          f:conditions: {}
          f:observedGeneration: {}
          f:phase: {}
      manager: mg-operator
      operation: Update
      time: "2021-02-10T04:39:16Z"
  name: mgo-init-script
  namespace: demo
  resourceVersion: "98944"
  uid: 5f13be2a-9a47-4b7e-9b83-a00b9bc89438
spec:
  authSecret:
    name: mgo-init-script-auth
  init:
    initialized: true
    script:
      configMap:
        name: mg-init-script
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
                    app.kubernetes.io/instance: mgo-init-script
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
                namespaces:
                  - demo
                topologyKey: kubernetes.io/hostname
              weight: 100
            - podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app.kubernetes.io/instance: mgo-init-script
                    app.kubernetes.io/managed-by: kubedb.com
                    app.kubernetes.io/name: mongodbs.kubedb.com
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
            --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then \n
            \         exit 0\n        fi\n        exit 1"
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
            --quiet --eval \"db.adminCommand('ping').ok\" ) -eq \"1\" ]]; then \n
            \         exit 0\n        fi\n        exit 1"
        failureThreshold: 3
        periodSeconds: 10
        successThreshold: 1
        timeoutSeconds: 1
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: mgo-init-script
  replicas: 1
  sslMode: disabled
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageEngine: wiredTiger
  storageType: Durable
  deletionPolicy: Delete
  version: 4.4.26
status:
  conditions:
    - lastTransitionTime: "2021-02-10T04:38:53Z"
      message: 'The KubeDB operator has started the provisioning of MongoDB: demo/mgo-init-script'
      reason: DatabaseProvisioningStartedSuccessfully
      status: "True"
      type: ProvisioningStarted
    - lastTransitionTime: "2021-02-10T04:39:16Z"
      message: All desired replicas are ready.
      reason: AllReplicasReady
      status: "True"
      type: ReplicaReady
    - lastTransitionTime: "2021-02-10T04:39:33Z"
      message: 'The MongoDB: demo/mgo-init-script is accepting client requests.'
      observedGeneration: 3
      reason: DatabaseAcceptingConnectionRequest
      status: "True"
      type: AcceptingConnection
    - lastTransitionTime: "2021-02-10T04:39:33Z"
      message: 'The MongoDB: demo/mgo-init-script is ready.'
      observedGeneration: 3
      reason: ReadinessCheckSucceeded
      status: "True"
      type: Ready
    - lastTransitionTime: "2021-02-10T04:39:16Z"
      message: 'The MongoDB: demo/mgo-init-script is successfully provisioned.'
      observedGeneration: 2
      reason: DatabaseSuccessfullyProvisioned
      status: "True"
      type: Provisioned
  observedGeneration: 3
  phase: Ready
```

Please note that KubeDB operator has created a new Secret called `mgo-init-script-auth` *(format: {mongodb-object-name}-auth)* for storing the password for MongoDB superuser. This secret contains a `username` key which contains the *username* for MongoDB superuser and a `password` key which contains the *password* for MongoDB superuser.
If you want to use an existing secret please specify that when creating the MongoDB object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password`.

```bash
$ kubectl get secrets -n demo mgo-init-script-auth -o yaml
apiVersion: v1
data:
  password: eGtBaTRmRVpmSVFrNmczVw==
  user: cm9vdA==
kind: Secret
metadata:
  creationTimestamp: "2019-02-06T09:43:54Z"
  labels:
    app.kubernetes.io/name: mongodbs.kubedb.com
    app.kubernetes.io/instance: mgo-init-script
  name: mgo-init-script-auth
  namespace: demo
  resourceVersion: "89594"
  selfLink: /api/v1/namespaces/demo/secrets/mgo-init-script-auth
  uid: b7cf2369-29f3-11e9-aebf-080027875192
type: Opaque
```

Now, you can connect to this database through [mongo-shell](https://docs.mongodb.com/v3.4/mongo/). In this tutorial, we are connecting to the MongoDB server from inside the pod.

```bash
$ kubectl get secrets -n demo mgo-init-script-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo mgo-init-script-auth -o jsonpath='{.data.\password}' | base64 -d
oEwk7IGxCPM5OWo5

$ kubectl exec -it mgo-init-script-0 -n demo sh

> mongo admin
MongoDB shell version v3.4.10
connecting to: mongodb://127.0.0.1:27017/admin
MongoDB server version: 3.4.10
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	http://docs.mongodb.org/
Questions? Try the support group
	http://groups.google.com/group/mongodb-user

> db.auth("root","oEwk7IGxCPM5OWo5")
1

> show dbs
admin   0.000GB
config  0.000GB
kubedb  0.000GB
local   0.000GB

> use kubedb
switched to db kubedb

> db.people.find()
{ "_id" : ObjectId("5ba9d667981f02e927b6788e"), "firstname" : "kubernetes", "lastname" : "database" }

> exit
bye
```

As you can see here, the initial script has successfully created a database named `kubedb` and inserted data into that database successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mg/mgo-init-script -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mg/mgo-init-script

kubectl delete ns demo
```

## Next Steps

- [Backup and Restore](/docs/guides/mongodb/backup/overview/index.md) MongoDB databases using Stash.
- Initialize [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Monitor your MongoDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Detail concepts of [MongoDB object](/docs/guides/mongodb/concepts/mongodb.md).
- Detail concepts of [MongoDBVersion object](/docs/guides/mongodb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
