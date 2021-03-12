---
title: Elasticsearch Combined Cluster
menu:
  docs_{{ .version }}:
    identifier: es-combined-cluster
    name: Combined Cluster
    parent: es-clustering-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Combined Cluster

An Elasticsearch combined cluster is a group of one or more Elasticsearch nodes where each node can perform as master, data, and ingest nodes simultaneously.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install the KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/combined-cluster/yamls) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create Standalone Elasticsearch Cluster

Here, we are going to create a standalone (ie. `replicas: 1`) Elasticsearch cluster. We will use the Elasticsearch image provided by the Opendistro (`opendistro-1.12.0`) for this demo. To learn more about Elasticsearch CR, visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-standalone
  namespace: demo
spec:
  version: opendistro-1.12.0
  enableSSL: true
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate
```

Let's deploy the above example by the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/combined-cluster/yamls/es-standalone.yaml
elasticsearch.kubedb.com/es-standalone created
```

Watch the bootstrap progress:

```bash
$ kubectl get elasticsearch -n demo -w
NAME            VERSION             STATUS         AGE
es-standalone   opendistro-1.12.0   Provisioning   1m32s
es-standalone   opendistro-1.12.0   Provisioning   2m17s
es-standalone   opendistro-1.12.0   Provisioning   2m17s
es-standalone   opendistro-1.12.0   Provisioning   2m20s
es-standalone   opendistro-1.12.0   Ready          2m20s
```

Hence the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-standalone'
NAME                  READY   STATUS    RESTARTS   AGE
pod/es-standalone-0   1/1     Running   0          33m

NAME                           TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)    AGE
service/es-standalone          ClusterIP   10.96.46.11   <none>        9200/TCP   33m
service/es-standalone-master   ClusterIP   None          <none>        9300/TCP   33m
service/es-standalone-pods     ClusterIP   None          <none>        9200/TCP   33m

NAME                             READY   AGE
statefulset.apps/es-standalone   1/1     33m

NAME                                               TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-standalone   kubedb.com/elasticsearch   7.10.0    33m

NAME                                        TYPE                       DATA   AGE
secret/es-standalone-admin-cert             kubernetes.io/tls          3      33m
secret/es-standalone-admin-cred             kubernetes.io/basic-auth   2      33m
secret/es-standalone-archiver-cert          kubernetes.io/tls          3      33m
secret/es-standalone-ca-cert                kubernetes.io/tls          2      33m
secret/es-standalone-config                 Opaque                     3      33m
secret/es-standalone-http-cert              kubernetes.io/tls          3      33m
secret/es-standalone-kibanaro-cred          kubernetes.io/basic-auth   2      33m
secret/es-standalone-kibanaserver-cred      kubernetes.io/basic-auth   2      33m
secret/es-standalone-logstash-cred          kubernetes.io/basic-auth   2      33m
secret/es-standalone-readall-cred           kubernetes.io/basic-auth   2      33m
secret/es-standalone-snapshotrestore-cred   kubernetes.io/basic-auth   2      33m
secret/es-standalone-transport-cert         kubernetes.io/tls          3      33m

NAME                                         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-standalone-0   Bound    pvc-a2d3e491-1d66-4b29-bb18-d5f06905336c   1Gi        RWO            standard       33m
```

Connect to the Cluster:

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/es-standalone 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

```bash
# Get admin username & password from k8s secret
$ kubectl get secret -n demo es-standalone-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo es-standalone-admin-cred -o jsonpath='{.data.password}' | base64 -d
V,YY1.qXxoAch9)B

# Check cluster health
$ curl -XGET -k -u 'admin:V,YY1.qXxoAch9)B' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-standalone",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 1,
  "number_of_data_nodes" : 1,
  "active_primary_shards" : 1,
  "active_shards" : 1,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

## Create Multi-Node Combined Elasticsearch Cluster

Here, we are going to create a multi-node (say `replicas: 3`) Elasticsearch cluster. We will use the Elasticsearch image provided by the Opendistro (`opendistro-1.12.0`) for this demo. To learn more about Elasticsearch CR, visit [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-multinode
  namespace: demo
spec:
  version: opendistro-1.12.0
  enableSSL: true
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: DoNotTerminate
```

Let's deploy the above example by the following command:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/clustering/combined-cluster/yamls/es-multinode.yaml
elasticsearch.kubedb.com/es-multinode created
```

Watch the bootstrap progress:

```bash
$ kubectl get elasticsearch -n demo -w
NAME            VERSION             STATUS         AGE
es-multinode    opendistro-1.12.0   Provisioning   18s
es-multinode    opendistro-1.12.0   Provisioning   78s
es-multinode    opendistro-1.12.0   Provisioning   78s
es-multinode    opendistro-1.12.0   Provisioning   81s
es-multinode    opendistro-1.12.0   Ready          81s
```

Hence the cluster is ready to use.
Let's check the k8s resources created by the operator on the deployment of Elasticsearch CRO:

```bash
$ kubectl get all,secret,pvc -n demo -l 'app.kubernetes.io/instance=es-multinode'
NAME                 READY   STATUS    RESTARTS   AGE
pod/es-multinode-0   1/1     Running   0          6m12s
pod/es-multinode-1   1/1     Running   0          6m7s
pod/es-multinode-2   1/1     Running   0          6m2s

NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/es-multinode          ClusterIP   10.96.237.120   <none>        9200/TCP   6m14s
service/es-multinode-master   ClusterIP   None            <none>        9300/TCP   6m14s
service/es-multinode-pods     ClusterIP   None            <none>        9200/TCP   6m15s

NAME                            READY   AGE
statefulset.apps/es-multinode   3/3     6m12s

NAME                                              TYPE                       VERSION   AGE
appbinding.appcatalog.appscode.com/es-multinode   kubedb.com/elasticsearch   7.10.0    6m12s

NAME                                       TYPE                       DATA   AGE
secret/es-multinode-admin-cert             kubernetes.io/tls          3      6m14s
secret/es-multinode-admin-cred             kubernetes.io/basic-auth   2      6m13s
secret/es-multinode-archiver-cert          kubernetes.io/tls          3      6m13s
secret/es-multinode-ca-cert                kubernetes.io/tls          2      6m14s
secret/es-multinode-config                 Opaque                     3      6m12s
secret/es-multinode-http-cert              kubernetes.io/tls          3      6m14s
secret/es-multinode-kibanaro-cred          kubernetes.io/basic-auth   2      6m13s
secret/es-multinode-kibanaserver-cred      kubernetes.io/basic-auth   2      6m13s
secret/es-multinode-logstash-cred          kubernetes.io/basic-auth   2      6m13s
secret/es-multinode-readall-cred           kubernetes.io/basic-auth   2      6m13s
secret/es-multinode-snapshotrestore-cred   kubernetes.io/basic-auth   2      6m13s
secret/es-multinode-transport-cert         kubernetes.io/tls          3      6m14s

NAME                                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
persistentvolumeclaim/data-es-multinode-0   Bound    pvc-c031bd37-2266-4a0b-8d9f-313281379810   1Gi        RWO            standard       6m12s
persistentvolumeclaim/data-es-multinode-1   Bound    pvc-e75bc8a8-15ed-4522-b0b3-252ff6c841a8   1Gi        RWO            standard       6m7s
persistentvolumeclaim/data-es-multinode-2   Bound    pvc-6452fa80-91c6-4d71-9b93-5cff973a2625   1Gi        RWO            standard       6m2s

```

Connect to the Cluster:

```bash
# Port-forward the service to local machine
$ kubectl port-forward -n demo svc/es-multinode 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

```bash
# Get admin username & password from k8s secret
$ kubectl get secret -n demo es-multinode-admin-cred -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo es-multinode-admin-cred -o jsonpath='{.data.password}' | base64 -d
9f$A8o2pBpKL~1T8

# Check cluster health
$ curl -XGET -k -u 'admin:9f$A8o2pBpKL~1T8' "https://localhost:9200/_cluster/health?pretty"
{
  "cluster_name" : "es-multinode",
  "status" : "green",
  "timed_out" : false,
  "number_of_nodes" : 3,
  "number_of_data_nodes" : 3,
  "active_primary_shards" : 1,
  "active_shards" : 3,
  "relocating_shards" : 0,
  "initializing_shards" : 0,
  "unassigned_shards" : 0,
  "delayed_unassigned_shards" : 0,
  "number_of_pending_tasks" : 0,
  "number_of_in_flight_fetch" : 0,
  "task_max_waiting_in_queue_millis" : 0,
  "active_shards_percent_as_number" : 100.0
}
```

## Cleaning Up

TO cleanup the k8s resources created by this tutorial, run:

```bash
# standalone cluster
$ kubectl patch -n demo elasticsearch es-standalone -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete elasticsearch -n demo es-standalone

# multinode cluster
$ kubectl patch -n demo elasticsearch es-multinode -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete elasticsearch -n demo es-multinode

# delete namespace
$ kubectl delete namespace demo
```

## Next Steps

- Deploy [topology cluster](/docs/guides/elasticsearch/clustering/topology-cluster/index.md)
- Learn about [taking backup](/docs/guides/elasticsearch/backup/overview/index.md) of Elasticsearch database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).