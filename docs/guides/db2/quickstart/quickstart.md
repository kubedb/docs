---
title: DB2 Quickstart
menu:
  docs_{{ .version }}:
    identifier: db2-quickstart-overview
    name: Overview
    parent: db2-quickstart-db2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running DB2

This tutorial shows how to run a DB2 database with KubeDB.

> Note: YAML files used in this tutorial are stored in [docs/examples/db2/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/db2/quickstart).

## Before You Begin

- You need a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install KubeDB following the setup guide: [/docs/setup/README.md](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS   AGE
demo    Active   5s
```

## Check Available StorageClass

First, check the available StorageClasses in your cluster:

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIM POLICY   STATUS   AGE
standard             kubernetes.io/host-path Delete           Bound    30s
fast-ssd             kubernetes.io/host-path Delete           Bound    30s
```

This tutorial will use the `standard` StorageClass.

## Check Available DB2Version

List all available DB2 versions that can be deployed with KubeDB:

```bash
$ kubectl get db2versions
NAME    VERSION   DB_IMAGE              COORDINATOR_IMAGE                    DEPRECATED
11.5.9  11.5.8.0   kubedb/db2:11.5.8.0    ghcr.io/kubedb/db2-coordinator:...   false
```

## Create a DB2 Database

Here's the YAML manifest for creating a standalone DB2 instance:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DB2
metadata:
  name: db2
  namespace: demo
spec:
  version: 11.5.8.0
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: Delete
```

Apply the example manifest:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/quickstart/standalone.yaml
db2.kubedb.com/db2 created
```

KubeDB operator will create the necessary PetSet, Service, and other resources. Let's wait for the DB2 instance to be ready:

```bash
$ kubectl get db2 -n demo db2 -w
NAME   VERSION   STATUS    AGE
db2    11.5.8.0    Running   2m
```

Now, you can press `Ctrl+C` to stop the watch command.

## Verify DB2 Database

Check that the DB2 instance is up and running:

```bash
$ kubectl get db2 -n demo
NAME   VERSION   STATUS    AGE
db2    11.5.8.0    Running   3m
```

To verify the database is ready, use `describe` command:

```bash
$ kubectl describe db2 -n demo db2
Name:         db2
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         DB2
...
Spec:
  Version:            11.5.8.0
  Storage Type:       Durable
  Deletion Policy:    Delete
  ...
Status:
  Conditions:
    Last Transition Time:  2025-01-09T12:05:51Z
    Status:                True
    Type:                  Ready
  Phase:                   Ready
```

When `status.phase` is `Ready`, your DB2 database is ready to use.

## Check DB2 Pod

Check that the DB2 pod is running:

```bash
$ kubectl get pod -n demo
NAME       READY   STATUS    RESTARTS   AGE
db2-0      1/1     Running   0          3m
```

## Get Database Credentials

Get the generated credentials for connecting to the DB2 database:

```bash
$ kubectl get secret -n demo db2-auth -o jsonpath='{.data.username}' | base64 -d
db2inst1

$ kubectl get secret -n demo db2-auth -o jsonpath='{.data.password}' | base64 -d
your-generated-password
```

## Access the Database

Let's access the DB2 database to verify it's working. First, port-forward the service:

```bash
$ kubectl port-forward -n demo svc/db2 50000:50000
Forwarding from 127.0.0.1:50000 -> 50000
Forwarding from [::1]:50000 -> 50000
```

Now you can connect to the database using a DB2 client. Here's an example using the DB2 command line:

```bash
$ db2 connect to testdb user db2inst1 using <password>
```

Or you can check the pod logs to verify database operations:

```bash
$ kubectl logs -f -n demo db2-0
```

## View Service & PVC

Check the Service and PersistentVolumeClaim created by KubeDB:

```bash
$ kubectl get svc -n demo
NAME       TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
db2        ClusterIP   10.96.28.88   <none>        50000/TCP   3m
```

```bash
$ kubectl get pvc -n demo
NAME      STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
db2-0     Bound    pvc-a1b2c3d4-e5f6-47a8-9b0c-1d2e3f4g5h6i   10Gi       RWO            standard       3m
```

## Delete the Database

To delete the DB2 database, use the following command:

```bash
$ kubectl delete db2 -n demo db2
db2.kubedb.com "db2" deleted
```

The PVC will also be deleted due to the `deletionPolicy: Delete` setting:

```bash
$ kubectl get pvc -n demo
No resources found in demo namespace.
```

## Cleaning up

```bash
$ kubectl delete namespace demo
namespace "demo" deleted
```

## Next Steps

- Learn about the [DB2 CRD](/docs/guides/db2/concepts/db2.md).
- Learn about the [DB2Version CRD](/docs/guides/db2/concepts/catalog.md).
- Learn about [custom RBAC](/docs/guides/db2/custom-rbac/using-custom-rbac.md) for DB2.
- Learn about using [private registry](/docs/guides/db2/private-registry/using-private-registry.md) with DB2.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
