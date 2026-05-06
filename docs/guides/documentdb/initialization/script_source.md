---
title: Initialize DocumentDB using Script Source
menu:
  docs_{{ .version }}:
    identifier: documentdb-script-source-initialization
    name: Using Script
    parent: documentdb-initialization-documentdb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize DocumentDB with Script

KubeDB supports DocumentDB database initialization. This tutorial will show you how to use KubeDB to initialize a DocumentDB database from script.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Initialization Scripts

DocumentDB supports initialization with `.sh`, `.sql` and `.sql.gz` files. In this tutorial, we will use a sample initialization script to set up the database with necessary schema and tables.

We will use a ConfigMap as script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as script source.

At first, we will create a ConfigMap from initialization script file. Then, we will provide this ConfigMap as script source in `init.script` of DocumentDB crd spec.

Let's create a ConfigMap with initialization script,

```bash
$ kubectl create configmap -n demo documentdb-init-script \
--from-literal=data.sql="CREATE SCHEMA IF NOT EXISTS data; CREATE TABLE IF NOT EXISTS data.dashboard (id SERIAL PRIMARY KEY, title VARCHAR(255), created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP);"
configmap/documentdb-init-script created
```

## Create DocumentDB with script source

Following YAML describes the DocumentDB object with `init.script`,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DocumentDB
metadata:
  name: script-documentdb
  namespace: demo
spec:
  version: "pg17-0.109.0"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 5Gi
  init:
    script:
      configMap:
        name: documentdb-init-script
```

Here,

- `init.script` specifies scripts used to initialize the database when it is being created.

VolumeSource provided in `init.script` will be mounted in Pod and will be executed while creating DocumentDB.

Now, let's create the DocumentDB crd which YAML we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/initialization/script-documentdb.yaml
documentdb.kubedb.com/script-documentdb created
```

Now, wait until DocumentDB goes in `Running` state. Verify that the database is in `Running` state using following command,

```bash
$ kubectl get documentdb -n demo script-documentdb
NAME                VERSION       STATUS    AGE
script-documentdb   pg17-0.109.0  Running   39s
```

You can use `kubectl dba describe` command to view which resources has been created by KubeDB for this DocumentDB object.

```bash
$ kubectl dba describe documentdb -n demo script-documentdb
Name:               script-documentdb
Namespace:          demo
CreationTimestamp:  Wed, 09 Jan 2025 12:05:51 +0000
Labels:             <none>
Annotations:        kubectl.kubernetes.io/last-applied-configuration={"apiVersion":"kubedb.com/v1alpha2","kind":"DocumentDB"...
Replicas:           1  total
Status:             Running
Init:
  script:
Volume:
    Type:       ConfigMap (a volume populated by a ConfigMap)
    Name:       documentdb-init-script
    Optional:   false
  StorageType:  Durable
Volume:
  StorageClass:  standard
  Capacity:      5Gi
  Access Modes:  RWO

PetSet:
  Name:               script-documentdb
  CreationTimestamp:  Wed, 09 Jan 2026 12:05:52 +0000
  Labels:             app.kubernetes.io/name=documentdbs.kubedb.com
                      app.kubernetes.io/instance=script-documentdb
  Annotations:        <none>
  Replicas:           1 desired | 1 total
  Pods Status:        1 Running / 0 Waiting / 0 Succeeded / 0 Failed

Service:
  Name:         script-documentdb
  Labels:       app.kubernetes.io/name=documentdbs.kubedb.com
                app.kubernetes.io/instance=script-documentdb
  Annotations:  <none>
  Type:         ClusterIP
  IP:           10.108.14.12
  Port:         db  27017/TCP
  TargetPort:   db/TCP
  Endpoints:    192.168.1.31:27017

Database Secret:
  Name:         script-documentdb-auth
  Labels:       app.kubernetes.io/name=documentdbs.kubedb.com
                app.kubernetes.io/instance=script-documentdb
  Annotations:  <none>

Type:  Opaque

Data
====
  password:  16 bytes
  username:  8 bytes

Topology:
  Type     Pod                  StartTime                      Phase
  ----     ---                  ---------                      -----
  primary  script-documentdb-0  2025-01-09 12:05:52 +0000     Running

No Snapshots.

Events:
  Type    Reason      Age   From                  Message
  ----    ------      ----  ----                  -------
  Normal  Successful  1m    DocumentDB operator   Successfully created Service
  Normal  Successful  57s   DocumentDB operator   Successfully created PetSet
  Normal  Successful  57s   DocumentDB operator   Successfully created DocumentDB
  Normal  Successful  57s   DocumentDB operator   Successfully patched PetSet
  Normal  Successful  57s   DocumentDB operator   Successfully patched DocumentDB
```

## Verify Initialization

Now let's connect to our DocumentDB `script-documentdb` to verify that the database has been initialized successfully.

**Connection Information:**

- Host name/address: you can use any of these
  - Service: `script-documentdb.demo`
  - Pod IP: (`$ kubectl get pods script-documentdb-0 -n demo -o yaml | grep podIP`)
- Port: `27017`
- Maintenance database: `postgres`

- Username: Run following command to get *username*,

  ```bash
  $ kubectl get secrets -n demo script-documentdb-auth -o jsonpath='{.data.username}' | base64 -d
  postgres
  ```

- Password: Run the following command to get *password*,

  ```bash
  $ kubectl get secrets -n demo script-documentdb-auth -o jsonpath='{.data.password}' | base64 -d
  <your-generated-password>
  ```

You can verify the initialization by checking the pod logs to see if the initialization scripts were executed successfully. You can also connect to DocumentDB and run commands to verify the initialized resources.

```bash
$ kubectl logs -f -n demo script-documentdb-0
```

Look for messages in the pod logs indicating that the initialization script was executed during database startup. You should see output similar to:

```
[2026-01-09T12:05:51.697Z] Starting DocumentDB server
[2026-01-09T12:05:52.753Z] initdb: initialized database
[2026-01-09T12:05:52.822Z] server started
...
[2026-01-09T12:05:53.500Z] database system is ready to accept connections
```

This confirms that the initialization scripts were executed during the database startup process, and the `data` schema with the `dashboard` table has been created as specified in the initialization script.

## Using External Script Files

You can also use external script files from various sources. Here's an example creating a ConfigMap from a file URL or local file:

```bash
# From local file
$ kubectl create configmap -n demo documentdb-init-script \
--from-file=data.sql=./init.sql

# From URL
$ kubectl create configmap -n demo documentdb-init-script \
--from-literal=data.sql="$(curl -fsSL https://example.com/init.sql)"
```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo documentdb/script-documentdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo documentdb/script-documentdb

$ kubectl delete -n demo configmap/documentdb-init-script
$ kubectl delete ns demo
```

## Next Steps
- Learn about [custom RBAC](/docs/guides/documentdb/custom-rbac/using-custom-rbac.md) for DocumentDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

