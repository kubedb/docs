---
title: CLI | KubeDB
menu:
  docs_{{ .version }}:
    identifier: pb-cli-cli
    name: Quickstart
    parent: pb-cli-pgbouncer
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Manage KubeDB objects using CLIs

## KubeDB CLI

KubeDB comes with its own cli. It is called `kubedb` cli. `kubedb` can be used to manage any KubeDB object. `kubedb` cli also performs various validations to improve ux. To install KubeDB cli on your workstation, follow the steps [here](/docs/setup/README.md).

### How to Create objects

`kubectl create` creates a pgbouncer CRD object in `default` namespace by default. Following command will create a PgBouncer object as specified in `pgbouncer.yaml`.

```console
$ kubectl create -f pgbouncer-demo.yaml
pgbouncer "pgbouncer-demo" created
```

You can provide namespace as a flag `--namespace`. Provided namespace should match with namespace specified in input file.

```console
$ kubectl create -f pgbouncer-demo.yaml --namespace=kube-system
pgbouncer "pgbouncer-demo" created
```

`kubectl create` command also considers `stdin` as input.

```console
cat pgbouncer-demo.yaml | kubectl create -f -
```

### How to List Objects

`kubectl get` command allows users to list or find any KubeDB object. To list all PgBouncer objects in `default` namespace, run the following command:

```console
$ kubectl get pgbouncer
NAME            VERSION   STATUS    AGE
pgbouncer-demo   1.11.0    Running   13m
pgbouncer-dev    1.11.0    Running   11m
pgbouncer-prod   1.11.0    Running   11m
pgbouncer-qa     1.11.0    Running   10m
```

To get YAML of an object, use `--output=yaml` flag.

```yaml
$ kubectl get pgbouncer pgbouncer-demo --output=yaml
apiVersion: kubedb.com/v1alpha1
kind: PgBouncer
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha1","kind":"PgBouncer","metadata":{"annotations":{},"name":"pgbouncer-demo","namespace":"demo"},"spec":{"connectionPool":{"adminUsers":["admin","admin1"],"maxClientConnections":20,"reservePoolSize":5},"databases":[{"alias":"postgres","databaseName":"postgres","databaseRef":{"name":"quick-postgres"}},{"alias":"tmpdb","databaseName":"mydb","databaseRef":{"name":"quick-postgres"}}],"monitor":{"agent":"prometheus.io/builtin"},"replicas":1,"userListSecretRef":{"name":"db-user-pass"},"version":"1.12.0"}}
  creationTimestamp: "2019-10-31T10:34:04Z"
  finalizers:
  - kubedb.com
  generation: 1
  name: pgbouncer-demo
  namespace: demo
  resourceVersion: "4733"
  selfLink: /apis/kubedb.com/v1alpha1/namespaces/demo/pgbouncers/pgbouncer-demo
  uid: 158b7c58-ecb2-4a77-bceb-081489b4921a
spec:
  connectionPool:
    adminUsers:
    - admin
    - admin1
    poolMode: session
    port: 5432
    reservePoolSize: 5
  databases:
  - alias: postgres
    databaseName: postgres
    databaseRef:
      name: quick-postgres
      namespace: ""
  - alias: tmpdb
    databaseName: mydb
    databaseRef:
      name: quick-postgres
      namespace: ""
  monitor:
    agent: prometheus.io/builtin
    prometheus:
      port: 56790
    resources: {}
  podTemplate:
    controller: {}
    metadata: {}
    spec:
      resources: {}
  replicas: 1
  serviceTemplate:
    metadata: {}
    spec: {}
  userListSecretRef:
    name: db-user-pass
  version: 1.12.0
status:
  observedGeneration: 1$6208915667192219204
  phase: Running
```

To get JSON of an object, use `--output=json` flag.

```console
kubectl get pgbouncer pgbouncer-demo --output=json
```

To list all KubeDB objects, use following command:

```console
$ kubectl get all -n demo -o wide
NAME                   READY   STATUS    RESTARTS   AGE     IP           NODE          NOMINATED NODE   READINESS GATES
pod/pgbouncer-demo-0   2/2     Running   0          5m53s   10.244.1.3   kind-worker   <none>           <none>

NAME                           TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE          SELECTOR
service/kubedb                 ClusterIP   None            <none>        <none>      5m54s        <none>
service/pgbouncer-demo         ClusterIP   10.98.95.4      <none>        5432/TCP    5m54s        kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-demo
service/pgbouncer-demo-stats   ClusterIP   10.107.214.97   <none>        56790/TCP   5m38s        kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-demo

NAME                              READY       AGE             CONTAINERS           IMAGES
statefulset.apps/pgbouncer-demo   1/1         5m53s           pgbouncer,exporter   kubedb/pgbouncer:1.12.0,kubedb/pgbouncer_exporter:v0.1.1

NAME                                  VERSION     STATUS          AGE
pgbouncer.kubedb.com/pgbouncer-demo   1.12.0      Running         5m54s

```

Flag `--output=wide` is used to print additional information.

List command supports short names for each object types. You can use it like `kubectl get <short-name>`. Below are the short name for KubeDB objects:

- Postgres: `pg`
- PgBouncer: `pb`
- Snapshot: `snap`
- DormantDatabase: `drmn`

You can print labels with objects. The following command will list all Snapshots with their corresponding labels.

```console
$ kubectl get pb -n demo --show-labels
NAME                            DATABASE                STATUS      AGE       LABELS
pgbouncer-demo                  pb/pgbouncer-demo       Succeeded   11m       kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-demo
pgbouncer-tmp                   pb/postgres-demo        Succeeded   1h        kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-tmp
```

You can also filter list using `--selector` flag.

```console
$ kubectl get pb --selector='kubedb.com/kind=PgBouncer' --show-labels
NAME                            DATABASE           STATUS      AGE       LABELS
pgbouncer-demo                  pb/pgbouncer-demo  Succeeded   11m       kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-demo
pgbouncer-dev                   pb/postgres-demo   Succeeded   1h        kubedb.com/kind=PgBouncer,kubedb.com/name=pgbouncer-dev
```

To print only object name, run the following command:

```console
$ kubectl get all -n demo -o name
pod/pgbouncer-demo-0
service/kubedb
service/pgbouncer-demo
service/pgbouncer-demo-stats
statefulset.apps/pgbouncer-demo
pgbouncer.kubedb.com/pgbouncer-demo
```

### How to Describe Objects

`kubectl dba describe` command allows users to describe any KubeDB object. The following command will describe PgBouncer `pgbouncer-demo` with relevant information.

```console
Name:         pgbouncer-demo
Namespace:    default
API Version:  kubedb.com/v1alpha1
Kind:         PgBouncer
Metadata:
  Creation Timestamp:  2019-09-09T09:27:48Z
  Finalizers:
    kubedb.com
  Generation:        1
  Resource Version:  303596
  Self Link:         /apis/kubedb.com/v1alpha1/namespaces/demo/pgbouncers/pgbouncer-demo
  UID:               f59c58da-ae21-403d-a4ce-affc8e10345c
Spec:
  Connection Pool:
    Admin Users:
      admin
    Listen Address:     *
    Listen Port:        5432
    Max Client Conn:    20
    Pool Mode:          session
    Reserve Pool Size:  5
  Databases:
    Alias:                  postgres
    App Binding Name:       postgres-demo
    App Binding Namespace:  demo
    Database Name:          postgres
  Replicas:                 1
  Service Template:
    Metadata:
    Spec:
  User List:
    Secret Name:       db-userlist
    Secret Namespace:  demo
  Version:             1.11.0
Status:
  Observed Generation:  1$6208915667192219204
  Phase:                Running
Events:
  Type    Reason      Age   From                Message
  ----    ------      ----  ----                -------
  Normal  Successful  13m   PgBouncer operator  Successfully created Service
  Normal  Successful  13m   PgBouncer operator  Successfully created PgBouncer configMap
  Normal  Successful  13m   PgBouncer operator  Successfully created StatefulSet
  Normal  Successful  13m   PgBouncer operator  Successfully created PgBouncer statefulset
  Normal  Successful  13m   PgBouncer operator  Successfully patched StatefulSet
  Normal  Successful  13m   PgBouncer operator  Successfully patched PgBouncer statefulset

```

`kubectl dba describe` command provides following basic information about a database.

- StatefulSet
- Storage (Persistent Volume)
- Service
- Secret (If available)
- Topology (If available)
- Snapshots (If any)
- Monitoring system (If available)

To hide events on KubeDB object, use flag `--show-events=false`

To describe all PgBouncer objects in `default` namespace, use following command

```console
kubectl dba describe pb
```

To describe all PgBouncer objects from every namespace, provide `--all-namespaces` flag.

```console
kubectl dba describe pb --all-namespaces
```

To describe all KubeDB objects from every namespace, use the following command:

```console
kubectl dba describe all --all-namespaces
```

You can also describe KubeDb objects with matching labels. The following command will describe all Elasticsearch & PgBouncer objects with specified labels from every namespace.

```console
kubectl dba describe pg,es --all-namespaces --selector='group=dev'
```

To learn about various options of `describe` command, please visit [here](/docs/reference/kubectl-dba_describe.md).

### How to Edit Objects

`kubectl edit` command allows users to directly edit any KubeDB object. It will open the editor defined by _KUBEDB_EDITOR_, or _EDITOR_ environment variables, or fall back to `nano`.

Let's edit an existing running PgBouncer object to setup [Monitoring](/docs/guides/pgbouncer/monitoring/using-coreos-prometheus-operator.md). The following command will open PgBouncer `pgbouncer-demo` in editor.

```console
$ kubectl edit pb pgbouncer-demo

# Add following to Spec to configure monitoring:
  monitor:
    agent: prometheus.io/coreos-operator
    prometheus:
      namespace: demo
      labels:
        k8s-app: prometheus
      interval: 10s
pgbouncer "pgbouncer-demo" edited
```

#### Edit restrictions

Various fields of a KubeDb object can't be edited using `edit` command. The following fields are restricted from updates for all KubeDB objects:

- _apiVersion_
- _kind_
- _metadata.name_
- _metadata.namespace_

### How to Delete Objects

`kubectl delete` command will delete an object in `default` namespace by default unless namespace is provided. The following command will delete a PgBouncer `pgbouncer-dev` in default namespace

```console
$ kubectl delete pgbouncer pgbouncer-dev
pgbouncer "pgbouncer-dev" deleted
```

You can also use YAML files to delete objects. The following command will delete a PgBouncer using the type and name specified in `pgbouncer.yaml`.

```console
$ kubectl delete -f pgbouncer.yaml
PgBouncer "pgbouncer-dev" deleted
```

`kubectl delete` command also takes input from `stdin`.

```console
cat pgbouncer.yaml | kubectl delete -f -
```

To delete objects with matching labels, use `--selector` flag. The following command will delete PgBouncers with label `pgbouncer.kubedb.com/name=pgbouncer-demo`.

```console
kubectl delete pgbouncer -l pgbouncer.kubedb.com/name=pgbouncer-demo
```

## Using Kubectl

You can use Kubectl with KubeDB objects like any other CRDs. Below are some common examples of using Kubectl with KubeDB objects.

```console
# Create objects
$ kubectl create -f

# List objects
$ kubectl get pgbouncer
$ kubectl get pgbouncer.kubedb.com

# Delete objects
$ kubectl delete pgbouncer <name>
```

## Next Steps

- Learn how to use KubeDB to run a PgBouncer [here](/docs/guides/pgbouncer/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
