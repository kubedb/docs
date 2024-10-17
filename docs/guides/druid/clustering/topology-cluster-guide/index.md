---
title: Druid Topology Cluster Guide
menu:
  docs_{{ .version }}:
    identifier: guides-druid-clustering-topology-cluster-guide
    name: Druid Group Replication Guide
    parent: guides-druid-clustering
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB - Druid Cluster

This tutorial will show you how to use KubeDB to provision a Druid Cluster.

## Before You Begin

Before proceeding:

- Read [druid topology cluster overview](/docs/guides/druid/clustering/topology-cluster-overview/index.md) to get a basic idea about the design and architecture of Druid.

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends on ZooKeeper for external dependency with helm command.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/druid/clustering/topology-cluster-guide/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/clustering/topology-cluster-guide/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Druid Cluster

The following is an example `Druid` object which creates a Druid cluster of six nodes (coordinators, overlords, brokers, routers, historicals and middleManager). Each with one replica.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  version: 28.0.1
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: Delete
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/clustering/topology-cluster-guide/yamls/druid-with-config.yaml
druid.kubedb.com/druid-cluster created
```

KubeDB operator watches for `Druid` objects using Kubernetes API. When a `Druid` object is created, KubeDB operator will create a new PetSet and a Service with the matching Druid object name. KubeDB operator will also create a governing service for the PetSet with the name `<druid-object-name>-pods`.

```bash
$ kubectl dba describe my -n demo druid-cluster
Name:               druid-cluster
Namespace:          demo
CreationTimestamp:  Tue, 28 Jun 2022 17:54:10 +0600
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1","kind":"Druid","metadata":{"annotations":{},"name":"druid-cluster","namespace":"demo"},"spec":{"replicas":3,"storage":{"...
Replicas:           3  total
Status:             Provisioning
StorageType:        Durable
Volume:
  StorageClass:      standard
  Capacity:          1Gi
  Access Modes:      RWO
Paused:              false
Halted:              false
Termination Policy:  WipeOut

PetSet:          
  Name:               druid-cluster
  CreationTimestamp:  Tue, 28 Jun 2022 17:54:10 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=druid-cluster
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=druids.kubedb.com
  Annotations:        <none>
  Replicas:           824640792392 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         druid-cluster
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=druid-cluster
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=druids.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.223.45
  Port:         primary  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.44:3306

Service:        
  Name:         druid-cluster-pods
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=druid-cluster
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=druids.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           None
  Port:         db  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    10.244.0.44:3306,10.244.0.46:3306,10.244.0.48:3306

Service:        
  Name:         druid-cluster-standby
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=druid-cluster
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=druids.kubedb.com
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.96.70.224
  Port:         standby  3306/TCP
  TargetPort:   db/TCP
  Endpoints:    <none>

Auth Secret:
  Name:         druid-cluster-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=druid-cluster
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=druids.kubedb.com
  Annotations:  <none>
  Type:         kubernetes.io/basic-auth
  Data:
    password:  16 bytes
    username:  4 bytes

AppBinding:
  Metadata:
    Annotations:
      kubectl.kubernetes.io/last-applied-configuration:  {"apiVersion":"kubedb.com/v1","kind":"Druid","metadata":{"annotations":{},"name":"druid-cluster","namespace":"demo"},"spec":{"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","deletionPolicy":"WipeOut","topology":{"group":{"name":"dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b"},"mode":"GroupReplication"},"version":"8.0.35"}}

    Creation Timestamp:  2022-06-28T11:54:10Z
    Labels:
      app.kubernetes.io/component:   database
      app.kubernetes.io/instance:    druid-cluster
      app.kubernetes.io/managed-by:  kubedb.com
      app.kubernetes.io/name:        druids.kubedb.com
    Name:                            druid-cluster
    Namespace:                       demo
  Spec:
    Client Config:
      Service:
        Name:    druid-cluster
        Path:    /
        Port:    3306
        Scheme:  druid
      URL:       tcp(druid-cluster.demo.svc:3306)/
    Parameters:
      API Version:  appcatalog.appscode.com/v1alpha1
      Kind:         StashAddon
      Stash:
        Addon:
          Backup Task:
            Name:  druid-backup-8.0.21
            Params:
              Name:   args
              Value:  --all-databases --set-gtid-purged=OFF
          Restore Task:
            Name:  druid-restore-8.0.21
    Secret:
      Name:   druid-cluster-auth
    Type:     kubedb.com/druid
    Version:  8.0.35

Events:
  Type     Reason      Age   From               Message
  ----     ------      ----  ----               -------
  Normal   Successful  1m    Kubedb operator  Successfully created governing service
  Normal   Successful  1m    Kubedb operator  Successfully created service for primary/standalone
  Normal   Successful  1m    Kubedb operator  Successfully created service for secondary replicas
  Normal   Successful  1m    Kubedb operator  Successfully created database auth secret
  Normal   Successful  1m    Kubedb operator  Successfully created PetSet
  Normal   Successful  1m    Kubedb operator  Successfully created Druid
  Normal   Successful  1m    Kubedb operator  Successfully created appbinding


$ kubectl get petset -n demo
NAME       READY   AGE
druid-cluster   3/3     3m47s

$ kubectl get pvc -n demo
NAME              STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-druid-cluster-0   Bound    pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            standard       4m16s
data-druid-cluster-1   Bound    pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            standard       4m11s
data-druid-cluster-2   Bound    pvc-94f1c312-50e3-41e1-94a8-a820be0abc08   1Gi        RWO            standard       4m7s
s

$ kubectl get pv -n demo
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                  STORAGECLASS   REASON   AGE
pvc-4f8538f6-a6ce-4233-b533-8566852f5b98   1Gi        RWO            Delete           Bound    demo/data-druid-cluster-0   standard                4m39s
pvc-8823d3ad-d614-4172-89ac-c2284a17f502   1Gi        RWO            Delete           Bound    demo/data-druid-cluster-1   standard                4m35s
pvc-94f1c312-50e3-41e1-94a8-a820be0abc08   1Gi        RWO            Delete           Bound    demo/data-druid-cluster-2   standard                4m31s

$ kubectl get service -n demo
NAME               TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
druid-cluster           ClusterIP      10.96.223.45    <none>        3306/TCP       5m13s
druid-cluster-pods      ClusterIP      None            <none>        3306/TCP       5m13s
druid-cluster-standby   ClusterIP      10.96.70.224    <none>        3306/TCP       5m13s

```

KubeDB operator sets the `status.phase` to `Running` once the database is successfully created. Run the following command to see the modified `Druid` object:

```yaml
$ kubectl get  my -n demo druid-cluster -o yaml | kubectl neat
apiVersion: kubedb.com/v1
kind: Druid
metadata:
  name: druid-cluster
  namespace: demo
spec:
  authSecret:
    name: druid-cluster-auth
  podTemplate:
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: druid-cluster
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: druids.kubedb.com
              namespaces:
              - demo
              topologyKey: kubernetes.io/hostname
            weight: 100
          - podAffinityTerm:
              labelSelector:
                matchLabels:
                  app.kubernetes.io/instance: druid-cluster
                  app.kubernetes.io/managed-by: kubedb.com
                  app.kubernetes.io/name: druids.kubedb.com
              namespaces:
              - demo
              topologyKey: failure-domain.beta.kubernetes.io/zone
            weight: 50
      resources:
        limits:
          cpu: 500m
          memory: 1Gi
        requests:
          cpu: 500m
          memory: 1Gi
      serviceAccountName: druid-cluster
  replicas: 3
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
  topology:
    group:
      name: dc002fc3-c412-4d18-b1d4-66c1fbfbbc9b
    mode: GroupReplication
  version: 8.0.35
status:
  observedGeneration: 2$4213139756412538772
  phase: Running
```


## Connect with Druid Database
We will use [port forwarding](https://kubernetes.io/docs/tasks/access-application-cluster/port-forward-access-application-cluster/) to connect with our routers of the Druid database. Then we will use `curl` to send `HTTP` requests to check cluster health to verify that our Druid database is working well. It is also possible to use `External-IP` to access druid nodes if you make `service` type of that node as `LoadBalancer`.

### Check the Service Health

Let's port-forward the port `8888` to local machine:

```bash
$ kubectl port-forward -n demo svc/druid-cluster-routers 8888
Forwarding from 127.0.0.1:8888 -> 8888
Forwarding from [::1]:8888 -> 8888
```

Now, the Druid cluster is accessible at `localhost:8888`. Let's check the [Service Health](https://druid.apache.org/docs/latest/api-reference/service-status-api/#get-service-health) of Routers of the Druid database.

```bash
$ curl "http://localhost:8888/status/health"
true
```
From the retrieved health information above, we can see that our Druid cluster’s status is `true`,  indicating that the service can receive API calls and is healthy. In the same way it possible to check the health of other druid nodes by port-forwarding the appropriate services.

### Access the web console

We can also access the [web console](https://druid.apache.org/docs/latest/operations/web-console) of Druid database from any browser by port-forwarding the routers in the same way shown in the aforementioned step or directly using the `External-IP` if the router service type is `LoadBalancer`.

Now hit the `http://localhost:8888` from any browser, and you will be prompted to provide the credential of the druid database. By following the steps discussed below, you can get the credential generated by the KubeDB operator for your Druid database.

**Connection information:**

- Username:

  ```bash
  $ kubectl get secret -n demo druid-cluster-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo druid-cluster-admin-cred -o jsonpath='{.data.password}' | base64 -d
  LzJtVRX5E8MorFaf
  ```

After providing the credentials correctly, you should be able to access the web console like shown below.

<p align="center">
  <img alt="lifecycle"  src="/docs/images/druid/Druid-Web-Console.png">
</p>

You can use this web console for loading data, managing datasources and tasks, and viewing server status and segment information. You can also run SQL and native Druid queries in the console.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo druid druid-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kafka.kubedb.com/druid-cluster patched

$ kubectl delete dr druid-cluster  -n demo
druid.kubedb.com "druid-cluster" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/database/index.md).
- Detail concepts of [DruidDBVersion object](/docs/guides/druid/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
