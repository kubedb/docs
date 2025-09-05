---
title: Druid Failover and DR Scenarios Overview
menu:
    docs_{{ .version }}:
        identifier: druid-failover
        name: Overview
        parent: guides-druid-fdr
        weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Exploring Fault Tolerance in Druid with KubeDB

## Understanding High Availability and Failover in Druid on KubeDB

`Failover` in Druid refers to the process of automatically switching to a standby or replica node
when a critical service (like Coordinator or Overlord) fails. In distributed analytics systems, 
this ensures that ingestion, query, and management operations remain available even if one or more 
pods go down. KubeDB makes this seamless by managing Druid's lifecycle and health on Kubernetes.

Druid's architecture consists of several node types:
- **Coordinator**: Manages data segment availability and balancing.
- **Overlord**: Handles task management and ingestion.
- **Broker**: Routes queries to Historical and Real-time nodes.
- **Historical**: Stores immutable, queryable data segments.
- **MiddleManager**: Executes ingestion tasks.

KubeDB supports running multiple replicas for each role, providing high availability and automated failover. If a pod fails, KubeDB ensures a replacement is started and the cluster remains operational.

In this guide, you'll:
- Deploy a highly available Druid cluster
- Verify the roles and health of Druid pods
- Simulate failures and observe automated failover
- Validate data/query continuity

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends on ZooKeeper for external dependency  with helm command.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [guides/druid/quickstart/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/quickstart/overview/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Apache Druid. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/druid/quickstart/guide/index.md#tips-for-testing).

## Find Available StorageClass

We will have to provide `StorageClass` in Druid CRD specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  14h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Find Available DruidVersion

When you install the KubeDB operator, it registers a CRD named [DruidVersion](/docs/guides/druid/concepts/druidversion.md). The installation process comes with a set of tested DruidVersion objects. Let's check available DruidVersions by,

```bash
$ kubectl get druidversion
NAME     VERSION   DB_IMAGE                               DEPRECATED   AGE
28.0.1   28.0.1    ghcr.io/appscode-images/druid:28.0.1                24h
30.0.1   30.0.1    ghcr.io/appscode-images/druid:30.0.1                24h
31.0.0   31.0.0    ghcr.io/appscode-images/druid:31.0.0                24h

```

Notice the `DEPRECATED` column. Here, `true` means that this DruidVersion is deprecated for the current KubeDB version. KubeDB will not work for deprecated DruidVersion. You can also use the short from `drversion` to check available DruidVersions.

In this tutorial, we will use `28.0.1` DruidVersion CR to create a Druid cluster.

## Get External Dependencies Ready

### Deep Storage

One of the external dependency of Druid is deep storage where the segments are stored. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/quickstart/deep-storage-config.yaml
secret/deep-storage-config created
```
## Deploy a Highly Available Druid Cluster

## Create a Druid Cluster

The KubeDB operator implements a Druid CRD to define the specification of Druid.

The Druid instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 31.0.0
  deletionPolicy: Delete
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    coordinators:
      replicas: 2

    overlords:
      replicas: 2

    brokers:
      replicas: 2

    historicals:
      replicas: 2

    middleManagers:
      replicas: 2

    routers:
      replicas: 2

```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/druid/quickstart/druid-with-monitoring.yaml
druid.kubedb.com/druid-quickstart created
```

The Druid's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the newly provisioned Druid cluster.

```bash
$ kubectl get druid -n demo -w
NAME            TYPE                  VERSION   STATUS   AGE
druid-cluster   kubedb.com/v1alpha2   31.0.0    Ready    4m12s

```
## Inspect Druid Pod Roles and Health


You can monitor on another terminal the status until all pods are ready:
```shell
$ watch kubectl get my,petset,pods -n demo
```
See the database is ready.

```shell
$ kubectl get druid,petset,pods -n demo
NAME                             TYPE                  VERSION   STATUS   AGE
druid.kubedb.com/druid-cluster   kubedb.com/v1alpha2   31.0.0    Ready    58m

NAME                                                        AGE
petset.apps.k8s.appscode.com/druid-cluster-brokers          57m
petset.apps.k8s.appscode.com/druid-cluster-coordinators     57m
petset.apps.k8s.appscode.com/druid-cluster-historicals      57m
petset.apps.k8s.appscode.com/druid-cluster-middlemanagers   57m
petset.apps.k8s.appscode.com/druid-cluster-mysql-metadata   58m
petset.apps.k8s.appscode.com/druid-cluster-overlords        57m
petset.apps.k8s.appscode.com/druid-cluster-routers          57m
petset.apps.k8s.appscode.com/druid-cluster-zk               58m

NAME                                 READY   STATUS    RESTARTS   AGE
pod/druid-cluster-brokers-0          1/1     Running   0          57m
pod/druid-cluster-brokers-1          1/1     Running   0          57m
pod/druid-cluster-coordinators-0     1/1     Running   0          57m
pod/druid-cluster-historicals-0      1/1     Running   0          57m
pod/druid-cluster-historicals-1      1/1     Running   0          57m
pod/druid-cluster-middlemanagers-0   1/1     Running   0          57m
pod/druid-cluster-middlemanagers-1   1/1     Running   0          57m
pod/druid-cluster-mysql-metadata-0   2/2     Running   0          58m
pod/druid-cluster-mysql-metadata-1   2/2     Running   0          58m
pod/druid-cluster-mysql-metadata-2   2/2     Running   0          58m
pod/druid-cluster-overlords-0        1/1     Running   0          57m
pod/druid-cluster-overlords-1        1/1     Running   0          57m
pod/druid-cluster-routers-0          1/1     Running   0          57m
pod/druid-cluster-zk-0               1/1     Running   0          58m
pod/druid-cluster-zk-1               1/1     Running   0          58m
pod/druid-cluster-zk-2               1/1     Running   0          58m
pod/myminio-default-0                2/2     Running   0          59m

```


You can check the roles and status of Druid pods using labels:

```bash
$ kubectl get pods -n demo --show-labels | grep role
druid-cluster-brokers-0          1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-brokers-64667d6fbb,kubedb.com/role=brokers,statefulset.kubernetes.io/pod-name=druid-cluster-brokers-0
druid-cluster-brokers-1          1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-cluster-brokers-64667d6fbb,kubedb.com/role=brokers,statefulset.kubernetes.io/pod-name=druid-cluster-brokers-1
druid-cluster-coordinators-0     1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-coordinators-955d5f7c4,kubedb.com/role=coordinators,statefulset.kubernetes.io/pod-name=druid-cluster-coordinators-0
druid-cluster-historicals-0      1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-historicals-54894c9748,kubedb.com/role=historicals,statefulset.kubernetes.io/pod-name=druid-cluster-historicals-0
druid-cluster-historicals-1      1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-cluster-historicals-54894c9748,kubedb.com/role=historicals,statefulset.kubernetes.io/pod-name=druid-cluster-historicals-1
druid-cluster-middlemanagers-0   1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-middlemanagers-5556d8775c,kubedb.com/role=middleManagers,statefulset.kubernetes.io/pod-name=druid-cluster-middlemanagers-0
druid-cluster-middlemanagers-1   1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-cluster-middlemanagers-5556d8775c,kubedb.com/role=middleManagers,statefulset.kubernetes.io/pod-name=druid-cluster-middlemanagers-1
druid-cluster-mysql-metadata-0   2/2     Running   0          59m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster-mysql-metadata,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-mysql-metadata-55c56b8549,kubedb.com/role=primary,statefulset.kubernetes.io/pod-name=druid-cluster-mysql-metadata-0
druid-cluster-mysql-metadata-1   2/2     Running   0          59m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster-mysql-metadata,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-cluster-mysql-metadata-55c56b8549,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=druid-cluster-mysql-metadata-1
druid-cluster-mysql-metadata-2   2/2     Running   0          59m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster-mysql-metadata,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=mysqls.kubedb.com,apps.kubernetes.io/pod-index=2,controller-revision-hash=druid-cluster-mysql-metadata-55c56b8549,kubedb.com/role=standby,statefulset.kubernetes.io/pod-name=druid-cluster-mysql-metadata-2
druid-cluster-overlords-0        1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-overlords-d8fd8d477,kubedb.com/role=overlords,statefulset.kubernetes.io/pod-name=druid-cluster-overlords-0
druid-cluster-overlords-1        1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=1,controller-revision-hash=druid-cluster-overlords-d8fd8d477,kubedb.com/role=overlords,statefulset.kubernetes.io/pod-name=druid-cluster-overlords-1
druid-cluster-routers-0          1/1     Running   0          58m   app.kubernetes.io/component=database,app.kubernetes.io/instance=druid-cluster,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=druids.kubedb.com,apps.kubernetes.io/pod-index=0,controller-revision-hash=druid-cluster-routers-86f759b75b,kubedb.com/role=routers,statefulset.kubernetes.io/pod-name=druid-cluster-routers-0

```

## Insert and Query Data (Optional)

If you have Druid's web console or REST API exposed, you can submit a simple ingestion task and run a query to verify data flow. For example, use the Druid console or API to submit a batch ingestion and then query the data.

## How Failover Works in Druid with KubeDB

KubeDB continuously monitors the health of Druid pods. If a Coordinator, Overlord, or any other critical pod fails (due to crash, node failure, or manual deletion), KubeDB:
- Detects the failure
- Automatically creates a replacement pod
- Ensures the new pod joins the cluster and resumes its role

This process is automatic and typically completes within seconds, ensuring minimal disruption. YOu can learn more from [here](https://druid.apache.org/docs/latest/operations/high-availability/).

## Hands-on Failover Testing
### case 1: Delete a Zookeeper Cluster
For highly-available ZooKeeper, `KubeDB` provides a cluster of 3 `Zookeeper` nodes. You can delete one of the `Zookeeper` pods to see how `KubeDB` handles failover.
**Delete a `Zookeeper` pod**
```bash
$ kubectl delete pod -n demo druid-cluster-zk-0
pod "druid-cluster-zk-0" deleted
```
in another terminal you can watch their status 
```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```

```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
You can not see that a new `druid-cluster-zk-0` pod is created automatically and joins the `Zookeeper` ensemble. Because `Zookeeper` is highly available, the `Druid` cluster continues to function without interruption.

**Delete all of the `Zookeeper` pods**
```shell
kubectl delete pod -n demo druid-cluster-zk-0 druid-cluster-zk-1 druid-cluster-zk-2
pod "druid-cluster-zk-0" deleted
pod "druid-cluster-zk-1" deleted
pod "druid-cluster-zk-2" deleted
```
```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
For a moment, all the nodes may disappear, but shortly after, you’ll see that all the `Zookeeper` pods remain unchange
```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0
```

### case 2: Delete a MySQL Pod
Druid uses MySQL for metadata storage. Each of the nods has their role also, you can see the role of each pod.

```shell
watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```

```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
**Delete the `primary` MySQL pod**

You can delete `druid-cluster-mysql-metadata-0` pods which has `primary` role to see how KubeDB handles failover.
```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 primary
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
You can see how quickly another pod takes the role of `primary` and the deleted pod is recreated automatically after a few seconds with the role of `standby`. 
```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 standby
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 primary
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```

**Delete a `standby` MySQL pod**
You can also delete `druid-cluster-mysql-metadata-1` pods which has `standby` role to see how KubeDB handles failover.
```shell
kubectl delete pod -n demo druid-cluster-mysql-metadata-0 druid-cluster-mysql-metadata-1
pod "druid-cluster-mysql-metadata-0" deleted
pod "druid-cluster-mysql-metadata-1" deleted
```

For few seconds, you will see that  `standby` roles are missing.
```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 
druid-cluster-mysql-metadata-1
druid-cluster-mysql-metadata-2 primary
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
After 10-30 seconds you can see that the deleted pods are recreated automatically and join the MySQL cluster with their respective roles.
```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 standby
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 primary
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
**Delete all of the `MySQL` Pods**
You can also delete all of the `MySQL` pods to see how KubeDB handles failover.
```shell
kubectl delete pod -n demo druid-cluster-mysql-metadata-0 druid-cluster-mysql-metadata-1 druid-cluster-mysql-metadata-2
pod "druid-cluster-mysql-metadata-0" deleted
pod "druid-cluster-mysql-metadata-1" deleted
pod "druid-cluster-mysql-metadata-2" deleted
```
You may notice that the `MySQL` pods briefly lose their assigned roles for a few seconds. To maintain stability in a production environment, it's best to avoid deleting all `MySQL` pods at once.
```shell
druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 
druid-cluster-mysql-metadata-1
druid-cluster-mysql-metadata-2
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0
```
It will show an error when you will try to access the Druid web console or REST API.
<figure align="center">
  <img alt="Rotate Authentication process of Kafka" src="/docs/guides/druid/failover/error1.png">
<figcaption align="center">Fig:Error in Druid API</figcaption>
</figure>

### Case 3: Delete a Broker Pod

Druid Brokers can be scaled out and all running servers will be active and queryable. We 
recommend placing them behind a load balancer.

Delete a `Broker` pod and observe failover:

```bash
$ kubectl delete pod -n demo druid-cluster-brokers-0
pod "druid-cluster-brokers-0" deleted

```

Monitor the pods:

```shell
$ watch -n 2 "kubectl get pods -n demo -o jsonpath='{range .items[*]}{.metadata.name} {.metadata.labels.kubedb\\.com/role}{\"\\n\"}{end}'"
```
```bash

druid-cluster-brokers-0 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-routers-0 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0
```

Since `Broker` pods are stateless, they can be seamlessly recreated without affecting the cluster's operations, ensuring no disruptions occur.
### Case 4: Delete an Overlord Or Coordinator Pod
For highly-available Apache Druid Coordinators and Overlords, we recommend to run multiple 
servers. If they are all configured to use the same ZooKeeper cluster and metadata storage,
then they will automatically failover between each other as necessary. Only one will be active 
at a time, but inactive servers will redirect to the currently active server.

**Delete a Coordinator Pod**
```bash
$ kubectl delete pod -n demo druid-cluster-coordinators-0
pod "druid-cluster-coordinators-0" deleted
``` 
```shell

druid-cluster-brokers-0 brokers
druid-cluster-brokers-1 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-coordinators-1 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-historicals-1 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-middlemanagers-1 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-overlords-0 overlords
druid-cluster-overlords-1 overlords
druid-cluster-routers-0 routers
druid-cluster-routers-1 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
**Delete All Coordinator Pods**
```bash
**Delete a Overlord Pod**

```bash
kubectl delete pod -n demo druid-cluster-overlords-0
pod "druid-cluster-overlords-0" deleted
```
```shell

druid-cluster-brokers-0 brokers
druid-cluster-brokers-1 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-coordinators-1 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-historicals-1 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-middlemanagers-1 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-overlords-0 overlords
druid-cluster-overlords-1 overlords
druid-cluster-routers-0 routers
druid-cluster-routers-1 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
**Delete All Overlord Pods**
```bash
$ kubectl delete pod -n demo druid-cluster-overlords-0 druid-cluster-overlords-1
pod "druid-cluster-overlords-0" deleted
pod "druid-cluster-overlords-1" deleted
```

Again, KubeDB will recreate the pod and maintain ingestion availability.
```shell
druid-cluster-brokers-0 brokers
druid-cluster-brokers-1 brokers
druid-cluster-coordinators-0 coordinators
druid-cluster-coordinators-1 coordinators
druid-cluster-historicals-0 historicals
druid-cluster-historicals-1 historicals
druid-cluster-middlemanagers-0 middleManagers
druid-cluster-middlemanagers-1 middleManagers
druid-cluster-mysql-metadata-0 primary
druid-cluster-mysql-metadata-1 standby
druid-cluster-mysql-metadata-2 standby
druid-cluster-overlords-0 overlords
druid-cluster-overlords-1 overlords
druid-cluster-routers-0 routers
druid-cluster-routers-1 routers
druid-cluster-zk-0
druid-cluster-zk-1
druid-cluster-zk-2
myminio-default-0

```
Even if you delete any pod, KubeDB will automatically recreate it, ensuring the cluster stays healthy and fully functional without interruption

## Cleanup

To clean up run:

```bash
kubectl delete druid -n demo druid-cluster
kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/druid/backup/overview/index.md) for Druid using Stash.
- Monitor your Druid cluster with [Prometheus integration](/docs/guides/druid/monitoring/using-builtin-prometheus.md).
- Explore Druid [configuration options](/docs/guides/druid/configuration/_index.md).
- Contribute to KubeDB: [contribution guidelines](/docs/CONTRIBUTING.md).

