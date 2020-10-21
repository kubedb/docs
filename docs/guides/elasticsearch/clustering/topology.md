---
title: Elasticsearch Cluster Topology
menu:
  docs_{{ .version }}:
    identifier: es-topology-clustering
    name: Topology
    parent: es-clustering-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Topology

KubeDB Elasticsearch supports multi-node database cluster.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create multi-node Elasticsearch

Elasticsearch can be created with multiple nodes. If you want to create an Elasticsearch cluster with three nodes, you need to set `spec.replicas` to `3`. In this case, all of these three nodes will act as *master*, *data* and *client*.

Check following Elasticsearch object

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: multi-node-es
  namespace: demo
spec:
  version: 7.3.2
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Here,

- `spec.replicas` is the number of nodes in the Elasticsearch cluster. Here, we are creating a three node Elasticsearch cluster.

> Note: If `spec.topology` is set, you won't able to `spec.replicas`. KubeDB will reject the create request for Elasticsearch crd from validating webhook.

Create example above with following command

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/clustering/multi-node-es.yaml
elasticsearch.kubedb.com/multi-node-es created
```

```bash
$ kubectl get es -n demo
NAME            VERSION   STATUS    AGE
multi-node-es   7.3.2     Running   7m38s
```

Let's describe Elasticsearch object `multi-node-es` while Running

```yaml
$ kubectl dba describe es -n demo multi-node-es
Name:               multi-node-es
Namespace:          demo
CreationTimestamp:  Wed, 02 Oct 2019 10:37:14 +0600
Labels:             <none>
Annotations:        <none>
Status:             Running
Replicas:           3  total
  StorageType:      Durable
Volume:
  StorageClass:  standard
  Capacity:      1Gi
  Access Modes:  RWO

StatefulSet:          
  Name:               multi-node-es
  CreationTimestamp:  Wed, 02 Oct 2019 10:37:15 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=multi-node-es
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=elasticsearch
                        app.kubernetes.io/version=7.3.2
                        kubedb.com/kind=Elasticsearch
                        kubedb.com/name=multi-node-es
                        node.role.client=set
                        node.role.data=set
                        node.role.master=set
  Annotations:        <none>
  Replicas:           824641370200 desired | 3 total
  Pods Status:        3 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         multi-node-es
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=multi-node-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=multi-node-es
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.0.13.4
  Port:         http  9200/TCP
  TargetPort:   http/TCP
  Endpoints:    10.4.0.10:9200,10.4.1.7:9200,10.4.1.8:9200

Service:        
  Name:         multi-node-es-master
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=multi-node-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=multi-node-es
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.0.11.165
  Port:         transport  9300/TCP
  TargetPort:   transport/TCP
  Endpoints:    10.4.0.10:9300,10.4.1.7:9300,10.4.1.8:9300

Database Secret:
  Name:         multi-node-es-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=multi-node-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=multi-node-es
  Annotations:  <none>
  
Type:  Opaque
  
Data
====
  ADMIN_PASSWORD:  8 bytes
  ADMIN_USERNAME:  7 bytes

Certificate Secret:
  Name:         multi-node-es-cert
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=multi-node-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=multi-node-es
  Annotations:  <none>
  
Type:  Opaque
  
Data
====
  node.jks:    3007 bytes
  root.jks:    863 bytes
  root.pem:    1139 bytes
  client.jks:  3038 bytes
  key_pass:    6 bytes

Topology:
  Type                Pod              StartTime                      Phase
  ----                ---              ---------                      -----
  master|client|data  multi-node-es-0  2019-10-02 10:41:16 +0600 +06  Running
  client|data|master  multi-node-es-1  2019-10-02 10:41:36 +0600 +06  Running
  master|client|data  multi-node-es-2  2019-10-02 10:42:04 +0600 +06  Running

No Snapshots.

Events:
  Type    Reason      Age   From                    Message
  ----    ------      ----  ----                    -------
  Normal  Successful  7m    Elasticsearch operator  Successfully created Service
  Normal  Successful  7m    Elasticsearch operator  Successfully created Service
  Normal  Successful  2m    Elasticsearch operator  Successfully created StatefulSet
  Normal  Successful  1m    Elasticsearch operator  Successfully created Elasticsearch
  Normal  Successful  1m    Elasticsearch operator  Successfully created appbinding
  Normal  Successful  1m    Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful  1m    Elasticsearch operator  Successfully patched Elasticsearch
```

Here, we can see in `Topology` section that all three Pods are acting as *master*, *data* and *client*.

## Create Elasticsearch with dedicated node

If you want to use separate node for *master*, *data* and *client* role, you need to configure `spec.topology` field of Elasticsearch crd.

In this tutorial, we will create following Elasticsearch with topology

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: topology-es
  namespace: demo
spec:
  version: 7.3.2
  storageType: Durable
  topology:
    master:
      prefix: master
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      prefix: data
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    client:
      prefix: client
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Here,

- `spec.topology` point to the number of pods we want as dedicated `master`, `client` and `data` nodes and also specify prefix, storage, resources for the pods.

Let's create this Elasticsearch object

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/clustering/topology-es.yaml
elasticsearch.kubedb.com/topology-es created
```

When this object is created, Elasticsearch database has started with 5 pods under 3 different StatefulSets.

```bash
$ kubectl get statefulset -n demo --show-labels --selector="kubedb.com/kind=Elasticsearch,kubedb.com/name=topology-es"
NAME                 READY   AGE     LABELS
client-topology-es   2/2     2m44s   app.kubernetes.io/component=database,app.kubernetes.io/instance=topology-es,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=elasticsearch,app.kubernetes.io/version=7.3.2,kubedb.com/kind=Elasticsearch,kubedb.com/name=topology-es,node.role.client=set
data-topology-es     2/2     81s     app.kubernetes.io/component=database,app.kubernetes.io/instance=topology-es,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=elasticsearch,app.kubernetes.io/version=7.3.2,kubedb.com/kind=Elasticsearch,kubedb.com/name=topology-es,node.role.data=set
master-topology-es   1/1     109s    app.kubernetes.io/component=database,app.kubernetes.io/instance=topology-es,app.kubernetes.io/managed-by=kubedb.com,app.kubernetes.io/name=elasticsearch,app.kubernetes.io/version=7.3.2,kubedb.com/kind=Elasticsearch,kubedb.com/name=topology-es,node.role.master=set
```

Three StatefulSets are created for *client*, *data* and *master* node respectively.

- client-topology-es

    ```yaml
    spec:
      topology:
        client:
          prefix: client
          replicas: 2
          storage:
            storageClassName: "standard"
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 1Gi
    ```

    This configuration creates a StatefulSet named `client-topology-es` for client node

  - `spec.replicas` is set to `2`. Two dedicated nodes is created as client.
  - Label `node.role.client: set` is added in Pods
  - Each Pod will receive a single PersistentVolume with a StorageClass of **standard** and **1Gi** of provisioned storage.

- data-topology-es

    ```yaml
    spec:
      topology:
        data:
          prefix: data
          replicas: 2
          storage:
            storageClassName: "standard"
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 1Gi
    ```

  This configuration creates a StatefulSet named `data-topology-es` for data node.

  - `spec.replicas` is set to `2`. Two dedicated nodes is created for data.
  - Label `node.role.data: set` is added in Pods
  - Each Pod will receive a single PersistentVolume with a StorageClass of **standard** and **1 Gib** of provisioned storage. 

- master-topology-es

    ```yaml
    spec:
      topology:
        master:
          prefix: master
          replicas: 1
          storage:
            storageClassName: "standard"
            accessModes:
            - ReadWriteOnce
            resources:
              requests:
                storage: 1Gi
    ```

    This configuration creates a StatefulSet named `data-topology-es` for master node

  - `spec.replicas` is set to `1`. One dedicated node is created as master.
  - Label `node.role.master: set` is added in Pods
  - Each Pod will receive a single PersistentVolume with a StorageClass of **standard** and **1Gi** of provisioned storage.

> Note: StatefulSet name format: `{topology-prefix}-{elasticsearch-name}`

Let's describe this Elasticsearch

```bash
$ kubectl dba describe es -n demo topology-es
Name:               topology-es
Namespace:          demo
CreationTimestamp:  Wed, 02 Oct 2019 10:46:12 +0600
Labels:             <none>
Annotations:        <none>
Status:             Running
  StorageType:      Durable
No volumes.

StatefulSet:          
  Name:               client-topology-es
  CreationTimestamp:  Wed, 02 Oct 2019 10:46:13 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=topology-es
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=elasticsearch
                        app.kubernetes.io/version=7.3.2
                        kubedb.com/kind=Elasticsearch
                        kubedb.com/name=topology-es
                        node.role.client=set
  Annotations:        <none>
  Replicas:           824638512252 desired | 2 total
  Pods Status:        2 Running / 0 Waiting / 0 Succeeded / 0 Failed

StatefulSet:          
  Name:               data-topology-es
  CreationTimestamp:  Wed, 02 Oct 2019 10:47:36 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=topology-es
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=elasticsearch
                        app.kubernetes.io/version=7.3.2
                        kubedb.com/kind=Elasticsearch
                        kubedb.com/name=topology-es
                        node.role.data=set
  Annotations:        <none>
  Replicas:           824639227260 desired | 2 total
  Pods Status:        2 Running / 0 Waiting / 0 Succeeded / 0 Failed

StatefulSet:          
  Name:               master-topology-es
  CreationTimestamp:  Wed, 02 Oct 2019 10:47:08 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=topology-es
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=elasticsearch
                        app.kubernetes.io/version=7.3.2
                        kubedb.com/kind=Elasticsearch
                        kubedb.com/name=topology-es
                        node.role.master=set
  Annotations:        <none>
  Replicas:           824639229484 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         topology-es
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=topology-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=topology-es
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.0.0.102
  Port:         http  9200/TCP
  TargetPort:   http/TCP
  Endpoints:    10.4.0.11:9200,10.4.1.9:9200

Service:        
  Name:         topology-es-master
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=topology-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=topology-es
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.0.12.178
  Port:         transport  9300/TCP
  TargetPort:   transport/TCP
  Endpoints:    10.4.1.10:9300

Database Secret:
  Name:         topology-es-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=topology-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=topology-es
  Annotations:  <none>
  
Type:  Opaque
  
Data
====
  ADMIN_PASSWORD:  8 bytes
  ADMIN_USERNAME:  7 bytes

Certificate Secret:
  Name:         topology-es-cert
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=topology-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=topology-es
  Annotations:  <none>
  
Type:  Opaque
  
Data
====
  client.jks:  3032 bytes
  key_pass:    6 bytes
  node.jks:    3005 bytes
  root.jks:    863 bytes
  root.pem:    1139 bytes

Topology:
  Type    Pod                   StartTime                      Phase
  ----    ---                   ---------                      -----
  client  client-topology-es-0  2019-10-02 10:46:19 +0600 +06  Running
  client  client-topology-es-1  2019-10-02 10:46:47 +0600 +06  Running
  data    data-topology-es-0    2019-10-02 10:47:42 +0600 +06  Running
  data    data-topology-es-1    2019-10-02 10:48:02 +0600 +06  Running
  master  master-topology-es-0  2019-10-02 10:47:14 +0600 +06  Running

No Snapshots.

Events:
  Type    Reason      Age   From                    Message
  ----    ------      ----  ----                    -------
  Normal  Successful  3m    Elasticsearch operator  Successfully created Service
  Normal  Successful  3m    Elasticsearch operator  Successfully created Service
  Normal  Successful  2m    Elasticsearch operator  Successfully created StatefulSet
  Normal  Successful  1m    Elasticsearch operator  Successfully created StatefulSet
  Normal  Successful  1m    Elasticsearch operator  Successfully created StatefulSet
  Normal  Successful  48s   Elasticsearch operator  Successfully created Elasticsearch
  Normal  Successful  48s   Elasticsearch operator  Successfully created appbinding
  Normal  Successful  47s   Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful  47s   Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful  47s   Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful  17s   Elasticsearch operator  Successfully patched Elasticsearch
```

Here, we can see from `Topology` section that 2 pods working as *client*, 2 pods working as *data* and 1 pod working as *master*.

Two services are also created for this Elasticsearch object.

- Service *`quick-elasticsearch`* targets all Pods which are acting as *client* node
- Service *`quick-elasticsearch-master`* targets all Pods which are acting as *master* node

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo es/multi-node-es es/topology-es -p '{"spec":{"terminationPolicy": "WipeOut"}}' --type="merge"
kubectl delete -n demo es/multi-node-es es/topology-es

kubectl delete ns demo
```

## Next Steps

- Learn about [taking backup](/docs/guides/elasticsearch/backup/stash.md) of Elasticsearch database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
