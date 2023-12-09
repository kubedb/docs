---
title: CLI | KubeDB
menu:
  docs_{{ .version }}:
    identifier: es-cli-cli
    name: Quickstart
    parent: es-cli-elasticsearch
    weight: 100
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Manage KubeDB objects using CLIs

## KubeDB CLI

KubeDB comes with its own cli. It is called `kubedb` cli. `kubedb` can be used to manage any KubeDB object. `kubedb` cli also performs various validations to improve ux. To install KubeDB cli on your workstation, follow the steps [here](/docs/setup/README.md).

### How to Create objects

`kubectl create` creates a database CRD object in `default` namespace by default. Following command will create an Elasticsearch object as specified in `elasticsearch.yaml`.

```bash
$ kubectl create -f elasticsearch-demo.yaml
elasticsearch.kubedb.com/elasticsearch-demo created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```bash
$ kubectl create -f elasticsearch-demo.yaml --namespace=kube-system
elasticsearch.kubedb.com/elasticsearch-demo created
```

`kubectl create` command also considers `stdin` as input.

```bash
cat elasticsearch-demo.yaml | kubectl create -f -
```

### How to List Objects

`kubectl get` command allows users to list or find any KubeDB object. To list all Elasticsearch objects in `default` namespace, run the following command:

```bash
$ kubectl get elasticsearch
NAME                 VERSION   STATUS    AGE
elasticsearch-demo   7.3.2     Running   1m
```

To get YAML of an object, use `--output=yaml` flag.

```yaml
$ kubectl get elasticsearch elasticsearch-demo --output=yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  creationTimestamp: 2018-10-08T14:22:19Z
  finalizers:
  - kubedb.com
  generation: 3
  name: elasticsearch-demo
  namespace: demo
  resourceVersion: "51660"
  selfLink: /apis/kubedb.com/v1alpha2/namespaces/default/elasticsearches/elasticsearch-demo
  uid: 90a54c9e-cb05-11e8-8d51-9eed48c5e947
spec:
  authSecret:
    name: elasticsearch-demo-auth
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
  replicas: 1
  storage:
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  terminationPolicy: Halt
  version: xpack-8.11.1
status:
  observedGeneration: 3$4212299729528774793
  phase: Running
```

To get JSON of an object, use `--output=json` flag.

```bash
$ kubectl get elasticsearch elasticsearch-demo --output=json
```

To list all KubeDB objects, use following command:

```bash
$ kubectl get all -o wide
NAME                       READY     STATUS    RESTARTS   AGE       IP              NODE              NOMINATED NODE
pod/elasticsearch-demo-0   1/1       Running   0          2m        192.168.1.105   4gb-pool-crtbqq   <none>

NAME                                TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE       SELECTOR
service/elasticsearch-demo          ClusterIP   10.98.224.23    <none>        9200/TCP   2m        app.kubernetes.io/name=elasticsearches.kubedb.com,app.kubernetes.io/instance=elasticsearch-demo,node.role.client=set
service/elasticsearch-demo-master   ClusterIP   10.100.87.240   <none>        9300/TCP   2m        app.kubernetes.io/name=elasticsearches.kubedb.com,app.kubernetes.io/instance=elasticsearch-demo,node.role.master=set
service/kubedb                      ClusterIP   None            <none>        <none>     2m        <none>
service/kubernetes                  ClusterIP   10.96.0.1       <none>        443/TCP    9h        <none>

NAME                                  DESIRED   CURRENT   AGE       CONTAINERS      IMAGES
statefulset.apps/elasticsearch-demo   1         1         2m        elasticsearch   kubedbci/elasticsearch:5.6-v1

NAME                                               VERSION   DB_IMAGE                          DEPRECATED   AGE
elasticsearchversion.catalog.kubedb.com/5.6        5.6       kubedbci/elasticsearch:5.6        true         5h
elasticsearchversion.catalog.kubedb.com/5.6-v1     5.6       kubedbci/elasticsearch:5.6-v1                  5h
elasticsearchversion.catalog.kubedb.com/5.6.4      5.6.4     kubedbci/elasticsearch:5.6.4      true         5h
elasticsearchversion.catalog.kubedb.com/5.6.4-v1   5.6.4     kubedbci/elasticsearch:5.6.4-v1                5h
elasticsearchversion.catalog.kubedb.com/6.2        6.2       kubedbci/elasticsearch:6.2        true         5h
elasticsearchversion.catalog.kubedb.com/6.2-v1     6.2       kubedbci/elasticsearch:6.2-v1                  5h
elasticsearchversion.catalog.kubedb.com/6.2.4      6.2.4     kubedbci/elasticsearch:6.2.4      true         5h
elasticsearchversion.catalog.kubedb.com/6.2.4-v1   6.2.4     kubedbci/elasticsearch:6.2.4-v1                5h
elasticsearchversion.catalog.kubedb.com/6.3        6.3       kubedbci/elasticsearch:6.3        true         5h
elasticsearchversion.catalog.kubedb.com/6.3-v1     6.3       kubedbci/elasticsearch:6.3-v1                  5h
elasticsearchversion.catalog.kubedb.com/6.3.0      6.3.0     kubedbci/elasticsearch:6.3.0      true         5h
elasticsearchversion.catalog.kubedb.com/6.3.0-v1   6.3.0     kubedbci/elasticsearch:6.3.0-v1                5h
elasticsearchversion.catalog.kubedb.com/6.4        6.4       kubedbci/elasticsearch:6.4                     5h
elasticsearchversion.catalog.kubedb.com/6.4.0      6.4.0     kubedbci/elasticsearch:6.4.0                   5h

NAME                                          VERSION   STATUS    AGE
elasticsearch.kubedb.com/elasticsearch-demo   5.6-v1    Running   2m
NAME                                     DATABASE                     BUCKET              STATUS      AGE
snap/elasticsearch-demo-20170605-073557  es/elasticsearch-demo        gs:bucket-name      Succeeded   9m
snap/snapshot-20171212-114700            es/elasticsearch-demo        gs:bucket-name      Succeeded   1h
```

Flag `--output=wide` is used to print additional information.

List command supports short names for each object types. You can use it like `kubectl get <short-name>`. Below are the short name for KubeDB objects:

- Elasticsearch: `es`
- Snapshot: `snap`
- DormantDatabase: `drmn`

You can print labels with objects. The following command will list all Snapshots with their corresponding labels.

```bash
$ kubectl get snap --show-labels
NAME                                 DATABASE                     STATUS      AGE       LABELS
elasticsearch-demo-20170605-073557   es/elasticsearch-demo        Succeeded   11m       app.kubernetes.io/name=elasticsearches.kubedb.com,app.kubernetes.io/instance=elasticsearch-demo
snapshot-20171212-114700             es/elasticsearch-demo        Succeeded   1h        app.kubernetes.io/name=elasticsearches.kubedb.com,app.kubernetes.io/instance=elasticsearch-demo
```

You can also filter list using `--selector` flag.

```bash
$ kubectl get snap --selector='app.kubernetes.io/name=elasticsearches.kubedb.com' --show-labels
NAME                                 DATABASE                STATUS      AGE       LABELS
elasticsearch-demo-20171212-073557   es/elasticsearch-demo   Succeeded   14m       app.kubernetes.io/name=elasticsearches.kubedb.com,app.kubernetes.io/instance=elasticsearch-demo
snapshot-20171212-114700             es/elasticsearch-demo   Succeeded   2h        app.kubernetes.io/name=elasticsearches.kubedb.com,app.kubernetes.io/instance=elasticsearch-demo
```

To print only object name, run the following command:

```bash
$ kubectl get all -o name
pod/elasticsearch-demo-0
service/elasticsearch-demo
service/elasticsearch-demo-master
service/kubedb
service/kubernetes
statefulset.apps/elasticsearch-demo
elasticsearchversion.catalog.kubedb.com/5.6
elasticsearchversion.catalog.kubedb.com/5.6-v1
elasticsearchversion.catalog.kubedb.com/5.6.4
elasticsearchversion.catalog.kubedb.com/5.6.4-v1
elasticsearchversion.catalog.kubedb.com/6.2
elasticsearchversion.catalog.kubedb.com/6.2-v1
elasticsearchversion.catalog.kubedb.com/6.2.4
elasticsearchversion.catalog.kubedb.com/6.2.4-v1
elasticsearchversion.catalog.kubedb.com/6.3
elasticsearchversion.catalog.kubedb.com/6.3-v1
elasticsearchversion.catalog.kubedb.com/6.3.0
elasticsearchversion.catalog.kubedb.com/6.3.0-v1
elasticsearchversion.catalog.kubedb.com/6.4
elasticsearchversion.catalog.kubedb.com/6.4.0
elasticsearch.kubedb.com/elasticsearch-demo
```

### How to Describe Objects

`kubectl dba describe` command allows users to describe any KubeDB object. The following command will describe Elasticsearch database `elasticsearch-demo` with relevant information.

```bash
$ kubectl dba describe es elasticsearch-demo
Name:               elasticsearch-demo
Namespace:          default
CreationTimestamp:  Mon, 08 Oct 2018 20:22:19 +0600
Labels:             <none>
Annotations:        <none>
Status:             Running
Replicas:           1  total
  StorageType:      Durable
Volume:
  StorageClass:  standard
  Capacity:      1Gi
  Access Modes:  RWO

StatefulSet:
  Name:               elasticsearch-demo
  CreationTimestamp:  Mon, 08 Oct 2018 20:22:22 +0600
  Labels:               app.kubernetes.io/name=elasticsearches.kubedb.com
                        app.kubernetes.io/instance=elasticsearch-demo
                        node.role.client=set
                        node.role.data=set
                        node.role.master=set
  Annotations:        <none>
  Replicas:           824642046536 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:         elasticsearch-demo
  Labels:         app.kubernetes.io/name=elasticsearches.kubedb.com
                  app.kubernetes.io/instance=elasticsearch-demo
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.98.224.23
  Port:         http  9200/TCP
  TargetPort:   http/TCP
  Endpoints:    192.168.1.105:9200

Service:
  Name:         elasticsearch-demo-master
  Labels:         app.kubernetes.io/name=elasticsearches.kubedb.com
                  app.kubernetes.io/instance=elasticsearch-demo
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.100.87.240
  Port:         transport  9300/TCP
  TargetPort:   transport/TCP
  Endpoints:    192.168.1.105:9300

Certificate Secret:
  Name:         elasticsearch-demo-cert
  Labels:         app.kubernetes.io/name=elasticsearches.kubedb.com
                  app.kubernetes.io/instance=elasticsearch-demo
  Annotations:  <none>

Type:  Opaque

Data
====
  key_pass:     6 bytes
  node.jks:     3015 bytes
  root.jks:     864 bytes
  sgadmin.jks:  3011 bytes

Database Secret:
  Name:         elasticsearch-demo-auth
  Labels:         app.kubernetes.io/name=elasticsearches.kubedb.com
                  app.kubernetes.io/instance=elasticsearch-demo
  Annotations:  <none>

Type:  Opaque

Data
====
  sg_roles.yml:           312 bytes
  sg_roles_mapping.yml:   73 bytes
  ADMIN_PASSWORD:         8 bytes
  READALL_USERNAME:       7 bytes
  sg_action_groups.yml:   430 bytes
  sg_internal_users.yml:  156 bytes
  ADMIN_USERNAME:         5 bytes
  READALL_PASSWORD:       8 bytes
  sg_config.yml:          242 bytes

Topology:
  Type                Pod                   StartTime                      Phase
  ----                ---                   ---------                      -----
  data|master|client  elasticsearch-demo-0  2018-10-08 20:22:23 +0600 +06  Running

No Snapshots.

Events:
  Type    Reason      Age   From                    Message
  ----    ------      ----  ----                    -------
  Normal  Successful  6m    Elasticsearch operator  Successfully created Service
  Normal  Successful  6m    Elasticsearch operator  Successfully created Service
  Normal  Successful  6m    Elasticsearch operator  Successfully created StatefulSet
  Normal  Successful  5m    Elasticsearch operator  Successfully created Elasticsearch
  Normal  Successful  5m    Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful  5m    Elasticsearch operator  Successfully patched Elasticsearch
  Normal  Successful  5m    Elasticsearch operator  Successfully patched StatefulSet
  Normal  Successful  4m    Elasticsearch operator  Successfully patched Elasticsearch
```

`kubectl dba describe` command provides following basic information about a database.

- StatefulSet
- Storage (Persistent Volume)
- Service
- Secret (If available)
- Topology (If available)
- Snapshots (If any)
- Monitoring system (If available)

To hide details about StatefulSet & Service, use flag `--show-workload=false`
To hide details about Secret, use flag `--show-secret=false`
To hide events on KubeDB object, use flag `--show-events=false`

To describe all Elasticsearch objects in `default` namespace, use following command

```bash
$ kubectl dba describe es
```

To describe all Elasticsearch objects from every namespace, provide `--all-namespaces` flag.

```bash
$ kubectl dba describe es --all-namespaces
```

To describe all KubeDB objects from every namespace, use the following command:

```bash
$ kubectl dba describe all --all-namespaces
```

You can also describe KubeDb objects with matching labels. The following command will describe all Elasticsearch objects with specified labels from every namespace.

```bash
$ kubectl dba describe es --all-namespaces --selector='group=dev'
```

To learn about various options of `describe` command, please visit [here](/docs/reference/cli/kubectl-dba_describe.md).


#### Edit restrictions

Various fields of a KubeDb object can't be edited using `edit` command. The following fields are restricted from updates for all KubeDB objects:

- apiVersion
- kind
- metadata.name
- metadata.namespace
- status

If StatefulSets or Deployments exists for a database, following fields can't be modified as well.

Elasticsearch:

- spec.init
- spec.storageType
- spec.storage
- spec.podTemplate.spec.nodeSelector
- spec.podTemplate.spec.env

For DormantDatabase, `spec.origin` can't be edited using `kubectl edit`

### How to Delete Objects

`kubectl delete` command will delete an object in `default` namespace by default unless namespace is provided. The following command will delete an Elasticsearch `elasticsearch-dev` in default namespace

```bash
$ kubectl delete elasticsearch elasticsearch-demo
elasticsearch.kubedb.com "elasticsearch-demo" deleted
```

You can also use YAML files to delete objects. The following command will delete an Elasticsearch using the type and name specified in `elasticsearch.yaml`.

```bash
$ kubectl delete -f elasticsearch-demo.yaml
elasticsearch.kubedb.com "elasticsearch-demo" deleted
```

`kubectl delete` command also takes input from `stdin`.

```bash
cat elasticsearch.yaml | kubectl delete -f -
```

To delete database with matching labels, use `--selector` flag. The following command will delete elasticsearch with label `elasticsearch.app.kubernetes.io/instance=elasticsearch-demo`.

```bash
$ kubectl delete elasticsearch -l elasticsearch.app.kubernetes.io/instance=elasticsearch-demo
```

## Using Kubectl

You can use Kubectl with KubeDB objects like any other CRDs. Below are some common examples of using Kubectl with KubeDB objects.

```bash
# List objects
$ kubectl get elasticsearch
$ kubectl get elasticsearch.kubedb.com

# Delete objects
$ kubectl delete elasticsearch <name>
```

## Next Steps

- Learn how to use KubeDB to run an Elasticsearch database [here](/docs/guides/elasticsearch/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
