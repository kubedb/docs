---
title: Cassandra Quickstart
menu:
  docs_{{ .version }}:
    identifier: cas-cassandra-quickstart-cassandra
    name: Cassandra
    parent: cas-quickstart-cassandra
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cassandra QuickStart

This tutorial will show you how to use KubeDB to run an [Apache Cassandra](https://cassandra.apache.org//).

<p align="center">
  <img alt="lifecycle"  src="/docs/images/cassandra/lifecycle.png">
</p>

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Cassandra=true` to ensure **Cassandra CRD**  with helm command.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [guides/cassandra/quickstart/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/cassandra/quickstart/overview/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> We have designed this tutorial to demonstrate a production setup of KubeDB managed Apache Cassandra. If you just want to try out KubeDB, you can bypass some safety features following the tips [here](/docs/guides/cassandra/quickstart/guide/index.md#tips-for-testing).

## Find Available StorageClass

We will have to provide `StorageClass` in Cassandra CRD specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  14h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Find Available CassandraVersion

When you install the KubeDB operator, it registers a CRD named [CassandraVersion](/docs/guides/cassandra/concepts/cassandraversion.md). The installation process comes with a set of tested CassandraVersion objects. Let's check available CassandraVersions by,

```bash
$ kubectl get cassandraversion
NAME    VERSION   DB_IMAGE                                             DEPRECATED   AGE
4.1.8   4.1.8     ghcr.io/appscode-images/cassandra-management:4.1.8                3m50s
5.0.3   5.0.3     ghcr.io/appscode-images/cassandra-management:5.0.3                3m50s
```

In this tutorial, we will use `5.0.3` CassandraVersion CR to create a Cassandra cluster.

## Create a Cassandra Cluster

The KubeDB operator implements a Cassandra CRD to define the specification of Cassandra.

The Cassandra instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cassandra-quickstart
  namespace: demo
spec:
  version: 5.0.3
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 600Mi
  deletionPolicy: Delete

```

Here,
- `spec.version` - is the name of the CassandraVersion CR. Here, a Cassandra of version `5.0.3` will be created.
- `spec.topology` - is the definition of the topology that will be deployed. This contains an array of racks definition.
- `spec.deletionPolicy` specifies what KubeDB should do when a user try to delete Cassandra CR. Deletion policy `Delete` will delete the database pods and PVC when the Cassandra CR is deleted.

Let's create the Cassandra CR that is shown above:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/quickstart/cassandra-quickstart.yaml
cassandra.kubedb.com/cassandra-quickstart created
```

The Cassandra's `STATUS` will go from `Provisioning` to `Ready` state within few minutes. Once the `STATUS` is `Ready`, you are ready to use the newly provisioned Cassandra cluster.

```bash
$ kubectl get cassandra -n demo -w
NAME                   TYPE                 VERSION   STATUS         AGE
cassandra-quickstart   kubedb.com/v1alpha2   5.0.3    Provisioning   17s
cassandra-quickstart   kubedb.com/v1alpha2   5.0.3    Provisioning   28s
.
.
cassandra-quickstart   kubedb.com/v1alpha2   5.0.3    Ready          82s
```

Describe the Cassandra object to observe the progress if something goes wrong or the status is not changing for a long period of time:

```bash
$ kubectl describe cassandra -n demo cassandra-quickstart
Name:         cassandra-quickstart
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  kubedb.com/v1alpha2
Kind:         Cassandra
Metadata:
  Creation Timestamp:  2025-07-14T05:57:35Z
  Finalizers:
    kubedb.com/cassandra
  Generation:        3
  Resource Version:  4981
  UID:               702f26e3-a02a-428a-9688-f0f2508ec662
Spec:
  Auth Secret:
    Name:  cassandra-quickstart-auth
  Auto Ops:
  Deletion Policy:  Delete
  Health Checker:
    Failure Threshold:  3
    Period Seconds:     20
    Timeout Seconds:    10
  Topology:
    Rack:
      Name:  r0
      Pod Template:
        Controller:
        Metadata:
        Spec:
          Containers:
            Name:  cassandra
            Resources:
              Limits:
                Memory:  1Gi
              Requests:
                Cpu:     500m
                Memory:  1Gi
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      999
              Seccomp Profile:
                Type:  RuntimeDefault
          Init Containers:
            Name:  cassandra-init
            Resources:
            Security Context:
              Allow Privilege Escalation:  false
              Capabilities:
                Drop:
                  ALL
              Run As Non Root:  true
              Run As User:      999
              Seccomp Profile:
                Type:  RuntimeDefault
          Pod Placement Policy:
            Name:  default
          Security Context:
            Fs Group:  999
      Replicas:        2
      Storage:
        Access Modes:
          ReadWriteOnce
        Resources:
          Requests:
            Storage:  600Mi
      Storage Type:   Durable
  Version:            5.0.3
Status:
  Conditions:
    Last Transition Time:  2025-07-14T06:02:12Z
    Message:               All replicas are ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  ReplicaReady
    Last Transition Time:  2025-07-14T05:57:38Z
    Message:               The KubeDB operator has started the provisioning of Cassandra: demo/cassandra-quickstart
    Observed Generation:   2
    Reason:                ProvisioningStarted
    Status:                True
    Type:                  ProvisioningStarted
    Last Transition Time:  2025-07-14T07:44:35Z
    Message:               The Cassandra: demo/cassandra-quickstart is accepting client requests
    Observed Generation:   3
    Reason:                AcceptingConnection
    Status:                True
    Type:                  AcceptingConnection
    Last Transition Time:  2025-07-14T07:44:35Z
    Message:               database demo/cassandra-quickstart is ready
    Observed Generation:   3
    Reason:                AllReplicasReady
    Status:                True
    Type:                  Ready
    Last Transition Time:  2025-07-14T06:03:39Z
    Message:               The Cassandra: demo/cassandra-quickstart is successfully provisioned.
    Observed Generation:   3
    Reason:                DatabaseSuccessfullyProvisioned
    Status:                True
    Type:                  Provisioned
  Phase:                   Ready
Events:                    <none>
```

### KubeDB Operator Generated Resources

On deployment of a Cassandra CR, the operator creates the following resources:

```bash
$ kubectl get all,secret,petset -n demo -l 'app.kubernetes.io/instance=cassandra-quickstart'
NAME                                 READY   STATUS    RESTARTS   AGE
pod/cassandra-quickstart-rack-r0-0   1/1     Running   0          108m
pod/cassandra-quickstart-rack-r0-1   1/1     Running   0          103m

NAME                                        TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)                               AGE
service/cassandra-quickstart                ClusterIP   10.43.56.40   <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   108m
service/cassandra-quickstart-rack-r0-pods   ClusterIP   None          <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   108m

NAME                                                      TYPE                   VERSION   AGE
appbinding.appcatalog.appscode.com/cassandra-quickstart   kubedb.com/cassandra   5.0.3     108m

NAME                                 TYPE                       DATA   AGE
secret/cassandra-quickstart-auth     kubernetes.io/basic-auth   2      108m
secret/cassandra-quickstart-config   Opaque                     1      108m

NAME                                                        AGE
petset.apps.k8s.appscode.com/cassandra-quickstart-rack-r0   108m
```

- `PetSet` - In topology mode, the operator creates 1 PetSet for each rack with name `{Cassandra-Name}-rack-{Rack-Name}`.
- `Services` -  For topology mode, 1 headless service for each PetSet with name `{PetSet-Name}-{pods}` is created. Other than that, 1 more service  with name `{Cassandra-Name}-{Sufix}` is created.
- `AppBinding` - an [AppBinding](/docs/guides/cassandra/concepts/appbinding.md) which hold to connect information for the Cassandra. Like other resources, it is named after the Cassandra instance.
- `Secrets` - A secret is generated for each Cassandra cluster.
    - `{Cassandra-Name}-auth` - the auth secrets which hold the `username` and `password` for the Cassandra users. Operator generates credentials for `admin` user and creates a secret for authentication.

## Connect with Cassandra database

Now, you can connect to this database using `cqlsh`. You will need `username` and `password` to connect to this database from `kubeclt exec` command. In this example, `cassandra-quickstart-auth`  secret holds username and password.

```bash
$ kubectl get secrets -n demo cassandra-quickstart-auth -o jsonpath='{.data.\username}' | base64 -d
admin

$ kubectl get secrets -n demo cassandra-quickstart-auth -o jsonpath='{.data.\password}' | base64 -d
9sPN85ctoRTnWEQV
```
We will exec into the pod `cassandra-quickstart-rack-r0-0` and connect to the database using `username` and `password`.

```bash
kubectl exec -it -n demo cassandra-quickstart-rack-r0-0  -- cqlsh -u admin -p '9sPN85ctoRTnWEQV'
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> describe keyspaces;

kubedb_keyspace  system_auth         system_schema  system_views         
system           system_distributed  system_traces  system_virtual_schema
```


## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo cassandra cassandra-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
cassandra.kubedb.com/cassandra-quickstart patched

$ kubectl delete cas cassandra-quickstart  -n demo
cassandra.kubedb.com "cassandra-quickstart" deleted

$  kubectl delete namespace demo
namespace "demo" deleted
```

## Next Steps

[//]: # (- Cassandra Clustering supported by KubeDB)

[//]: # (  - [Combined Clustering]&#40;/docs/guides/cassandra/clustering/combined-cluster/index.md&#41;)

[//]: # (  - [Topology Clustering]&#40;/docs/guides/cassandra/clustering/topology-cluster/index.md&#41;)
- Use [kubedb cli](/docs/guides/cassandra/cli/cli.md) to manage databases like kubectl for Kubernetes.

[//]: # (- Detail concepts of [Cassandra object]&#40;/docs/guides/cassandra/concepts/cassandra.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
