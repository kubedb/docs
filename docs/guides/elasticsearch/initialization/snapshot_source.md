---
title: Initialize Elasticsearch from Snapshot
menu:
  docs_{{ .version }}:
    identifier: es-snapshot-source-initialization
    name: Using Snapshot
    parent: es-initialization-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

{{< notice type="warning" message="This doc has been deprecated and will be removed in a future release. We recommend using [Stash](/docs/guides/elasticsearch/snapshot/stash.md) to backup & restore Elasticsearch database." >}}

> Don't know how backup works?  Check [tutorial](/docs/guides/elasticsearch/snapshot/instant_backup.md) on Instant Backup.

# Initialize Elasticsearch with Snapshot

KubeDB supports Elasticsearch database initialization.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/install.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```console
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/elasticsearch](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/elasticsearch) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Snapshot

This tutorial will show you how to use KubeDB to initialize an Elasticsearch database with an existing Snapshot. So, we need a Snapshot to perform this initialization. If you don't have a Snapshot already, create one by following the tutorial [here](/docs/guides/elasticsearch/snapshot/instant_backup.md).

If you have changed either namespace or snapshot object name, please modify the YAMLs used in this tutorial accordingly.

## Initialize with Snapshot source

You have to specify the Snapshot `name` and `namespace` in the `spec.init.snapshotSource` field of your new Elasticsearch object.

Below is the YAML for Elasticsearch object that will be created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha1
kind: Elasticsearch
metadata:
  name: recovered-es
  namespace: demo
spec:
  version: 7.3.2
  databaseSecret:
    secretName: instant-elasticsearch-auth
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    snapshotSource:
      name: instant-snapshot
      namespace: demo
```

Here,

- `spec.init.snapshotSource` specifies Snapshot object information to be used in this initialization process.
  - `snapshotSource.name` refers to a Snapshot object `name`.
  - `snapshotSource.namespace` refers to a Snapshot object `namespace`.

Snapshot `instant-snapshot` in `demo` namespace belongs to Elasticsearch `instant-elasticsearch`:

```console
$ kubectl get snap -n demo instant-snapshot
NAME               DATABASENAME            STATUS      AGE
instant-snapshot   instant-elasticsearch   Succeeded   2m21s
```

> Note: Elasticsearch `recovered-es` must have same superuser credentials as Elasticsearch `instant-elasticsearch`.

[//]: # (Describe authentication part. This should match with existing one)

Now, create the Elasticsearch object.

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/elasticsearch/initialization/recovered-es.yaml
elasticsearch.kubedb.com/recovered-es created
```

When Elasticsearch database is ready, KubeDB operator launches a Kubernetes Job to initialize this database using the data from Snapshot `instant-snapshot`.

```console
$ kubectl get es -n demo recovered-es
NAME           VERSION   STATUS         AGE
recovered-es   7.3.2     Initializing   100s

$ kubectl get es -n demo recovered-es
NAME           VERSION   STATUS    AGE
recovered-es   7.3.2     Running   7m6s
```

As a final step of initialization, KubeDB Job controller adds `kubedb.com/initialized` annotation in initialized Elasticsearch object. This prevents further invocation of initialization process.

```console
$ kubedb describe es -n demo recovered-es
Name:               recovered-es
Namespace:          demo
CreationTimestamp:  Wed, 02 Oct 2019 14:54:59 +0600
Labels:             <none>
Annotations:        kubedb.com/initialized=
Status:             Running
Replicas:           1  total
Init:
  snapshotSource:
    namespace:  demo
    name:       instant-snapshot
  StorageType:  Durable
Volume:
  StorageClass:  standard
  Capacity:      1Gi
  Access Modes:  RWO

StatefulSet:          
  Name:               recovered-es
  CreationTimestamp:  Wed, 02 Oct 2019 14:55:00 +0600
  Labels:               app.kubernetes.io/component=database
                        app.kubernetes.io/instance=recovered-es
                        app.kubernetes.io/managed-by=kubedb.com
                        app.kubernetes.io/name=elasticsearch
                        app.kubernetes.io/version=7.3.2
                        kubedb.com/kind=Elasticsearch
                        kubedb.com/name=recovered-es
                        node.role.client=set
                        node.role.data=set
                        node.role.master=set
  Annotations:        <none>
  Replicas:           824635596440 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:        
  Name:         recovered-es
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=recovered-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=recovered-es
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.0.1.27
  Port:         http  9200/TCP
  TargetPort:   http/TCP
  Endpoints:    10.4.1.53:9200

Service:        
  Name:         recovered-es-master
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=recovered-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=recovered-es
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.0.2.159
  Port:         transport  9300/TCP
  TargetPort:   transport/TCP
  Endpoints:    10.4.1.53:9300

Certificate Secret:
  Name:         recovered-es-cert
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=recovered-es
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=recovered-es
  Annotations:  <none>
  
Type:  Opaque
  
Data
====
  root.jks:    863 bytes
  root.pem:    1139 bytes
  client.jks:  3035 bytes
  key_pass:    6 bytes
  node.jks:    3004 bytes

Database Secret:
  Name:         instant-elasticsearch-auth
  Labels:         app.kubernetes.io/component=database
                  app.kubernetes.io/instance=instant-elasticsearch
                  app.kubernetes.io/managed-by=kubedb.com
                  app.kubernetes.io/name=elasticsearch
                  app.kubernetes.io/version=7.3.2
                  kubedb.com/kind=Elasticsearch
                  kubedb.com/name=instant-elasticsearch
  Annotations:  <none>
  
Type:  Opaque
  
Data
====
  ADMIN_PASSWORD:  8 bytes
  ADMIN_USERNAME:  7 bytes

Topology:
  Type                Pod             StartTime                      Phase
  ----                ---             ---------                      -----
  master|client|data  recovered-es-0  2019-10-02 14:55:06 +0600 +06  Running

No Snapshots.

Events:
  Type    Reason                Age   From                    Message
  ----    ------                ----  ----                    -------
  Normal  Successful            2m    Elasticsearch operator  Successfully created Service
  Normal  Successful            2m    Elasticsearch operator  Successfully created Service
  Normal  Successful            2m    Elasticsearch operator  Successfully created StatefulSet
  Normal  Successful            1m    Elasticsearch operator  Successfully created Elasticsearch
  Normal  Successful            1m    Elasticsearch operator  Successfully created appbinding
  Normal  Initializing          1m    Elasticsearch operator  Initializing from Snapshot: "instant-snapshot"
  Normal  Successful            1m    Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful            1m    Elasticsearch operator  Successfully patched Elasticsearch
  Normal  SuccessfulInitialize  59s   Elasticsearch operator  Successfully completed initialization
  Normal  Successful            59s   Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful            29s   Elasticsearch operator  Successfully patched Elasticsearch
  Normal  Successful            29s   Elasticsearch operator  Successfully patched StatefulSet
```

## Verify initialization

Let's connect to our Elasticsearch `recovered-es` to verify that the database has been successfully initialized.

At first, forward `9200` port of `recovered-es` pod. Run following command on a separate terminal,

```console
$ kubectl port-forward -n demo recovered-es-0 9200
Forwarding from 127.0.0.1:9200 -> 9200
Forwarding from [::1]:9200 -> 9200
```

Now, we can connect to the database at `localhost:9200`. Let's find out necessary connection information first.

**Connection information:**

- Address: `localhost:9200`
- Username: Run following command to get *username*

  ```console
  $ kubectl get secrets -n demo instant-elasticsearch-auth -o jsonpath='{.data.\ADMIN_USERNAME}' | base64 -d
  elastic
  ```

- Password: Run following command to get *password*

  ```console
  $ kubectl get secrets -n demo instant-elasticsearch-auth -o jsonpath='{.data.\ADMIN_PASSWORD}' | base64 -d
  dy76ez7v
  ```

We had created an index `test` before taking snapshot of `instant-elasticsearch` database. Let's check this index is present in newly initialized database `recovered-es`.

```console
$ curl -XGET --user "elastic:dy76ez7v" "localhost:9200/test/_search?pretty"
```

```json
{
  "took" : 3,
  "timed_out" : false,
  "_shards" : {
    "total" : 1,
    "successful" : 1,
    "skipped" : 0,
    "failed" : 0
  },
  "hits" : {
    "total" : {
      "value" : 1,
      "relation" : "eq"
    },
    "max_score" : 1.0,
    "hits" : [
      {
        "_index" : "test",
        "_type" : "_doc",
        "_id" : "1",
        "_score" : 1.0,
        "_source" : {
          "title" : "Snapshot",
          "text" : "Testing instand backup",
          "date" : "2018/02/13"
        }
      }
    ]
  }
}
```

We can see from above output that `test` index is present in `recovered-es` database. That's means our database has been initialized from snapshot successfully.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl patch -n demo es/instant-elasticsearch es/recovered-es -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo es/instant-elasticsearch es/recovered-es

kubectl delete ns demo
```

## Next Steps

- Learn how to [schedule backup](/docs/guides/elasticsearch/snapshot/scheduled_backup.md)  of Elasticsearch database.
- Learn how to configure [Elasticsearch Topology](/docs/guides/elasticsearch/clustering/topology.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` CoreOS Prometheus Operator](/docs/guides/elasticsearch/monitoring/using-coreos-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
