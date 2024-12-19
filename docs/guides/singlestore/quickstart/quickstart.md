---
title: SingleStore Quickstart
menu:
  docs_{{ .version }}:
    identifier: sdb-quickstart-quickstart
    name: Overview
    parent: sdb-quickstart-singlestore
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# SingleStore QuickStart

This tutorial will show you how to use KubeDB to run a SingleStore database.

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/singlestore/quickstart/images/singlestore-lifecycle.png">
</p>

> Note: The yaml files used in this tutorial are stored in [docs/guides/singlestore/quickstart/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/quickstart/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure install with helm command including `--set global.featureGates.Singlestore=true` to ensure SingleStore crd.
- [StorageClass](https://kubernetes.io/docs/concepts/storage/storage-classes/) is required to run KubeDB. Check the available StorageClass in cluster.

  ```bash
  $ kubectl get storageclasses
  NAME                 PROVISIONER             RECLAIMPOLICY     VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  standard (default)   rancher.io/local-path   Delete            WaitForFirstConsumer   false                  6h22m
  ```

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available SingleStoreVersion

When you have installed KubeDB, it has created `SinglestoreVersion` crd for all supported SingleStore versions. Check it by using the `kubectl get singlestoreversions` command. You can also use `sdbv` shorthand instead of `singlestoreversions`.

```bash
 $ kubectl get singlestoreversions.catalog.kubedb.com
NAME     VERSION   DB_IMAGE                                                          DEPRECATED   AGE
8.1.32   8.1.32    ghcr.io/appscode-images/singlestore-node:alma-8.1.32-e3d3cde6da                2d1h
8.5.30   8.5.30    ghcr.io/appscode-images/singlestore-node:alma-8.5.30-4f46ab16a5                2d1h
8.5.7    8.5.7     ghcr.io/appscode-images/singlestore-node:alma-8.5.7-bf633c1a54                 2d1h
8.7.10   8.7.10    ghcr.io/appscode-images/singlestore-node:alma-8.7.10-95e2357384                2d1h
8.7.21   8.7.21    ghcr.io/appscode-images/singlestore-node:alma-8.7.21-f0b8de04d5                2d1h
8.9.3    8.9.3     ghcr.io/appscode-images/singlestore-node:alma-8.9.3-bfa36a984a                 2d1h
```
## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Create a SingleStore database

KubeDB implements a `Singlestore` CRD to define the specification of a SingleStore database. Below is the `Singlestore` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-quickstart
  namespace: demo
spec:
  version: "8.5.7"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "0.5"
              requests:
                memory: "2Gi"
                cpu: "0.5"
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "0.5"
              requests:
                memory: "2Gi"
                cpu: "0.5"
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
  serviceTemplates:
  - alias: primary
    spec:
      type: LoadBalancer
      ports:
        - name: http
          port: 9999
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/quickstart/yamls/quickstart.yaml
singlestore.kubedb.com/sdb-quickstart created
```
Here,

- `spec.version` is the name of the SinglestoreVersion CRD where the docker images are specified. In this tutorial, a SingleStore `8.5.37` database is going to be created.
- `spec.topology` specifies that it will be used as cluster mode. If this field is nil it will be work as standalone mode.
- `spec.topology.aggregator.replicas` or `spec.topology.leaf.replicas` specifies that the number replicas that will be used for aggregator or leaf.
- `spec.storageType` specifies the type of storage that will be used for SingleStore database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create SingleStore database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology.aggregator.storage` or `spec.topology.leaf.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Singlestore` crd or which resources KubeDB should keep or delete when you delete `Singlestore` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in `storage.resources.requests` field. Don't specify limits here. PVC does not get resized automatically.

KubeDB operator watches for `Singlestore` objects using Kubernetes api. When a `Singlestore` object is created, KubeDB operator will create new PetSet and Service with the matching SingleStore object name. KubeDB operator will also create a governing service for PetSets, if one is not already present.

```bash
$ kubectl get petset -n demo
NAME                               READY   AGE
sdb-quickstart-leaf                2/2     33s
sdb-quickstart-aggregator          1/1     37s
$ kubectl get pvc -n demo
NAME                                      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
data-sdb-quickstart-leaf-0                Bound    pvc-4f45c51b-47d4-4254-8275-782bf3588667   10Gi       RWO            standard       <unset>                 42s
data-sdb-quickstart-leaf-1                Bound    pvc-769e68f4-80a9-4e3e-b2bc-e974534b9dee   10Gi       RWO            standard       <unset>                 35s
data-sdb-quickstart-aggregator-0          Bound    pvc-75057e3d-e1d7-4770-905b-6049f2edbcde   1Gi        RWO            standard       <unset>                 46s
$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                                          STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-4f45c51b-47d4-4254-8275-782bf3588667   10Gi       RWO            Delete           Bound    demo/data-sdb-quickstart-leaf-0                standard       <unset>                          87s
pvc-75057e3d-e1d7-4770-905b-6049f2edbcde   1Gi        RWO            Delete           Bound    demo/data-sdb-quickstart-master-aggregator-0   standard       <unset>                          91s
pvc-769e68f4-80a9-4e3e-b2bc-e974534b9dee   10Gi       RWO            Delete           Bound    demo/data-sdb-quickstart-leaf-1                standard       <unset>                          80s
$ kubectl get service -n demo
NAME                  TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)                         AGE
sdb-quickstart        LoadBalancer   10.96.27.144   192.10.25.36  3306:32076/TCP,8081:30910/TCP   2m1s
sdb-quickstart-pods   ClusterIP      None           <none>        3306/TCP                        2m1s

```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified Singlestore object:

```yaml
➤ kubectl get sdb -n demo sdb-quickstart -oyaml

 apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Singlestore","metadata":{"annotations":{},"name":"sdb-quickstart","namespace":"demo"},"spec":{"licenseSecret":{"name":"license-secret"},"serviceTemplates":[{"alias":"primary","spec":{"ports":[{"name":"http","port":9999}],"type":"LoadBalancer"}}],"storageType":"Durable","deletionPolicy":"WipeOut","topology":{"aggregator":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"0.5","memory":"2Gi"},"requests":{"cpu":"0.5","memory":"2Gi"}}}]}},"replicas":1,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"}},"leaf":{"podTemplate":{"spec":{"containers":[{"name":"singlestore","resources":{"limits":{"cpu":"0.5","memory":"2Gi"},"requests":{"cpu":"0.5","memory":"2Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"10Gi"}},"storageClassName":"standard"}}},"version":"8.5.7"}}
  creationTimestamp: "2024-05-06T06:52:58Z"
  finalizers:
  - kubedb.com
  generation: 2
  name: sdb-quickstart
  namespace: demo
  resourceVersion: "448498"
  uid: 29d6a814-e801-45b5-8217-b59fc77d84e5
spec:
  authSecret:
    name: sdb-quickstart-root-cred
  healthChecker:
    failureThreshold: 1
    periodSeconds: 10
    timeoutSeconds: 10
  licenseSecret:
    name: license-secret
  podPlacementPolicy:
    name: default
  serviceTemplates:
  - alias: primary
    metadata: {}
    spec:
      ports:
      - name: http
        port: 9999
      type: LoadBalancer
  storageType: Durable
  deletionPolicy: WipeOut
  topology:
    aggregator:
      podPlacementPolicy:
        name: default
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                cpu: 500m
                memory: 2Gi
              requests:
                cpu: 500m
                memory: 2Gi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          - name: singlestore-coordinator
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
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          initContainers:
          - name: singlestore-init
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
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          securityContext:
            fsGroup: 999
      replicas: 1
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    leaf:
      podPlacementPolicy:
        name: default
      podTemplate:
        controller: {}
        metadata: {}
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                cpu: 500m
                memory: 2Gi
              requests:
                cpu: 500m
                memory: 2Gi
            securityContext:
              allowPrivilegeEscalation: false
              capabilities:
                drop:
                - ALL
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          - name: singlestore-coordinator
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
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          initContainers:
          - name: singlestore-init
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
              runAsGroup: 998
              runAsNonRoot: true
              runAsUser: 999
              seccompProfile:
                type: RuntimeDefault
          securityContext:
            fsGroup: 999
      replicas: 2
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
        storageClassName: standard
  version: 8.5.7
status:
  conditions:
  - lastTransitionTime: "2024-05-06T06:53:06Z"
    message: 'The KubeDB operator has started the provisioning of Singlestore: demo/sdb-quickstart'
    observedGeneration: 2
    reason: DatabaseProvisioningStartedSuccessfully
    status: "True"
    type: ProvisioningStarted
  - lastTransitionTime: "2024-05-06T06:56:05Z"
    message: All Aggregator replicas are ready for Singlestore demo/sdb-quickstart
    observedGeneration: 2
    reason: AllReplicasReady
    status: "True"
    type: ReplicaReady
  - lastTransitionTime: "2024-05-06T06:54:17Z"
    message: database demo/sdb-quickstart is accepting connection
    observedGeneration: 2
    reason: AcceptingConnection
    status: "True"
    type: AcceptingConnection
  - lastTransitionTime: "2024-05-06T06:54:17Z"
    message: database demo/sdb-quickstart is ready
    observedGeneration: 2
    reason: AllReplicasReady
    status: "True"
    type: Ready
  - lastTransitionTime: "2024-05-06T06:54:18Z"
    message: 'The Singlestore: demo/sdb-quickstart is successfully provisioned.'
    observedGeneration: 2
    reason: DatabaseSuccessfullyProvisioned
    status: "True"
    type: Provisioned
  phase: Ready
 

```

## Connect with SingleStore database

KubeDB operator has created a new Secret called `sdb-quickstart-root-cred` *(format: {singlestore-object-name}-root-cred)* for storing the password for `singlestore` superuser. This secret contains a `username` key which contains the *username* for SingleStore superuser and a `password` key which contains the *password* for SingleStore superuser.

If you want to use an existing secret please specify that when creating the SingleStore object using `spec.authSecret.name`. While creating this secret manually, make sure the secret contains these two keys containing data `username` and `password` and also make sure of using `root` as value of `username`. For more details see [here](/docs/guides/mysql/concepts/database/index.md#specdatabasesecret).

Now, we need `username` and `password` to connect to this database from `kubectl exec` command. In this example  `sdb-quickstart-root-cred` secret holds username and password

```bash
$ kubectl get pod -n demo sdb-quickstart-master-aggregator-0 -oyaml | grep podIP
  podIP: 10.244.0.14
$ kubectl get secrets -n demo sdb-quickstart-root-cred -o jsonpath='{.data.\username}' | base64 -d
  root
$ kubectl get secrets -n demo sdb-quickstart-root-cred -o jsonpath='{.data.\password}' | base64 -d
  J0h_BUdJB8mDO31u
```
we will exec into the pod `sdb-quickstart-master-aggregator-0` and connect to the database using username and password

```bash
$ kubectl exec -it -n demo sdb-quickstart-aggregator-0 -- bash
  Defaulting container name to singlestore.
  Use 'kubectl describe pod/sdb-quickstart-aggregator-0 -n demo' to see all of the containers in this pod.
  
  [memsql@sdb-quickstart-master-aggregator-0 /]$ memsql -uroot -p"J0h_BUdJB8mDO31u"
  singlestore-client: [Warning] Using a password on the command line interface can be insecure.
  Welcome to the MySQL monitor.  Commands end with ; or \g.
  Your MySQL connection id is 1114
  Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)
  
  Copyright (c) 2000, 2016, Oracle and/or its affiliates. All rights reserved.
  
  Oracle is a registered trademark of Oracle Corporation and/or its
  affiliates. Other names may be trademarks of their respective
  owners.
  
  Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.
  
  singlestore> show databases;
  +--------------------+
  | Database           |
  +--------------------+
  | cluster            |
  | information_schema |
  | memsql             |
  | singlestore_health |
  +--------------------+
  4 rows in set (0.00 sec)

```
You can also connect with database management tools like [singlestore-studio](https://docs.singlestore.com/db/v8.5/reference/singlestore-tools-reference/singlestore-studio/)

You can simply access to SingleStore studio by forwarding the Primary service port to any of your localhost port. Or, Accessing through ExternalP's 8081 port is also an option.

```bash
$ kubectl port-forward -n demo service/sdb-quickstart 8081
Forwarding from 127.0.0.1:8081 -> 8081
Forwarding from [::1]:8081 -> 8081
```
Lets, open your browser and go to the http://localhost:8081 or with TLS https://localhost:8081 then click on `Add or Create Cluster` option.
Then choose `Add Existing Cluster` and click on `next` and you will get an interface like that below:

<p align="center">
  <img alt="studio-1"  src="/docs/guides/singlestore/quickstart/images/studio-1.png">
</p>

After giving the all information you can see like this below UI image.

<p align="center">
  <img alt="studio-1"  src="/docs/guides/singlestore/quickstart/images/studio-2.png">
</p>

## Database DeletionPolicy

This field is used to regulate the deletion process of the related resources when `Singlestore` object is deleted. User can set the value of this field according to their needs. The available options and their use case scenario is described below:

**DoNotTerminate:**

When `deletionPolicy` is set to `DoNotTerminate`, KubeDB takes advantage of `ValidationWebhook` feature in Kubernetes 1.9.0 or later clusters to implement `DoNotTerminate` feature. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. You can see this below:

```bash
$ kubectl delete sdb sdb-quickstart -n demo
The Singlestore "sdb-quickstart" is invalid: spec.deletionPolicy: Invalid value: "sdb-quickstart": Can not delete as deletionPolicy is set to "DoNotTerminate"
```

Now, run `kubectl patch -n demo sdb sdb-quickstart -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"` to set `spec.deletionPolicy` to `Halt` (which deletes the singlestore object and keeps PVC, snapshots, Secrets intact) or remove this field (which default to `Delete`). Then you will be able to delete/halt the database.

Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy).

**Halt:**

Suppose you want to reuse your database volume and credential to deploy your database in future using the same configurations. But, right now you just want to delete the database except the database volumes and credentials. In this scenario, you must set the `Singlestore` object `deletionPolicy` to `Halt`.

When the [DeletionPolicy](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy) is set to `halt` and the Singlestore object is deleted, the KubeDB operator will delete the PetSet and its pods but leaves the `PVCs`, `secrets` and database backup data(`snapshots`) intact. You can set the `deletionPolicy` to `halt` in existing database using `patch` command for testing.

At first, run `kubectl patch -n demo sdb sdb-quickstart -p '{"spec":{"deletionPolicy":"Halt"}}' --type="merge"`. Then delete the singlestore object,

```bash
$ kubectl delete sdb sdb-quickstart -n demo
singlestore.kubedb.com "sdb-quickstart" deleted
```

Now, run the following command to get all singlestore resources in `demo` namespaces,

```bash
$ kubectl get petset,svc,secret,pvc -n demo
NAME                              TYPE                       DATA   AGE
secret/sdb-quickstart-root-cred   kubernetes.io/basic-auth   2      3m35s

NAME                                                            STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-sdb-quickstart-leaf-0                Bound    pvc-389f40a8-09bc-4724-aa52-94705d56ff77   1Gi        RWO            standard       <unset>                 3m18s
persistentvolumeclaim/data-sdb-quickstart-leaf-1                Bound    pvc-8dfbf04e-41a8-4cdd-ba14-7ad42d8701bb   1Gi        RWO            standard       <unset>                 3m11s
persistentvolumeclaim/data-sdb-quickstart-master-aggregator-0   Bound    pvc-c4f7d255-7307-4455-b195-70c71b81706f   1Gi        RWO            standard       <unset>                 3m29s

```

From the above output, you can see that all singlestore resources(`PetSet`, `Service`, etc.) are deleted except `PVC` and `Secret`. You can recreate your singlestore again using this resources.

>You can also set the `deletionPolicy` to `Halt`(deprecated). It's behavior same as `halt` and right now `Halt` is replaced by `Halt`.

**Delete:**

If you want to delete the existing database along with the volumes used, but want to restore the database from previously taken `snapshots` and `secrets` then you might want to set the `Singlestore` object `deletionPolicy` to `Delete`. In this setting, `PetSet` and the volumes will be deleted. If you decide to restore the database, you can do so using the snapshots and the credentials.

When the [DeletionPolicy](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy) is set to `Delete` and the Singlestore object is deleted, the KubeDB operator will delete the PetSet and its pods along with PVCs but leaves the `secret` and database backup data(`snapshots`) intact.

Suppose, we have a database with `deletionPolicy` set to `Delete`. Now, are going to delete the database using the following command:

```bash
$ kubectl delete sdb sdb-quickstart -n demo
singlestore.kubedb.com "sdb-quickstart" deleted
```

Now, run the following command to get all singlestore resources in `demo` namespaces,

```bash
$ kubectl get petset,svc,secret,pvc -n demo
NAME                              TYPE                       DATA   AGE
secret/sdb-quickstart-root-cred   kubernetes.io/basic-auth   2      17m

```

From the above output, you can see that all singlestore resources(`PetSet`, `Service`, `PVCs` etc.) are deleted except `Secret`.

>If you don't set the deletionPolicy then the kubeDB set the DeletionPolicy to Delete by-default.

**WipeOut:**

You can totally delete the `Singlestore` database and relevant resources without any tracking by setting `deletionPolicy` to `WipeOut`. KubeDB operator will delete all relevant resources of this `Singlestore` database (i.e, `PVCs`, `Secrets`, `Snapshots`) when the `deletionPolicy` is set to `WipeOut`.

Suppose, we have a database with `deletionPolicy` set to `WipeOut`. Now, are going to delete the database using the following command:

```yaml
$ kubectl delete sdb sdb-quickstart -n demo
singlestore.kubedb.com "singlestore-quickstart" deleted
```

Now, run the following command to get all singlestore resources in `demo` namespaces,

```bash
$ kubectl get petset,svc,secret,pvc -n demo
No resources found in demo namespace.
```

From the above output, you can see that all singlestore resources are deleted. There is no option to recreate/reinitialize your database if `deletionPolicy` is set to `Delete`.

>Be careful when you set the `deletionPolicy` to `Delete`. Because there is no option to trace the database resources if once deleted the database.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo singlestore/sdb-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo singlestore/sdb-quickstart
kubectl delete ns demo
```

## Tips for Testing

If you are just testing some basic functionalities, you might want to avoid additional hassles due to some safety features that are great for production environment. You can follow these tips to avoid them.

1. **Use `storageType: Ephemeral`**. Databases are precious. You might not want to lose your data in your production environment if database pod fail. So, we recommend to use `spec.storageType: Durable` and provide storage spec in `spec.storage` section. For testing purpose, you can just use `spec.storageType: Ephemeral`. KubeDB will use [emptyDir](https://kubernetes.io/docs/concepts/storage/volumes/#emptydir) for storage. You will not require to provide `spec.storage` section.
2. **Use `deletionPolicy: WipeOut`**. It is nice to be able to delete everything created by KubeDB for a particular Singlestore crd when you delete the crd. For more details about deletion policy, please visit [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy).

## Next Steps