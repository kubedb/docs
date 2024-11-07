---
title: Reconfigure MSSQLServer Replicaset
menu:
  docs_{{ .version }}:
    identifier: mg-reconfigure-replicaset
    name: Replicaset
    parent: mg-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure MSSQLServer Replicaset Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a MSSQLServer Replicaset.

## Before You Begin

- You need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md). Make sure install with helm command including `--set global.featureGates.MSSQLServer=true` to ensure MSSQLServer CRD installation.

- To configure TLS/SSL in `MSSQLServer`, `KubeDB` uses `cert-manager` to issue certificates. So first you have to make sure that the cluster has `cert-manager` installed. To install `cert-manager` in your cluster following steps [here](https://cert-manager.io/docs/installation/kubernetes/).

- You should be familiar with the following `KubeDB` concepts:
  - [MSSQLServer](/docs/guides/mssqlserver/concepts/mssqlserver.md)
  - [ReplicaSet](/docs/guides/mssqlserver/clustering/replicaset.md)
  - [MSSQLServerOpsRequest](/docs/guides/mssqlserver/concepts/opsrequest.md)
  - [Reconfigure Overview](/docs/guides/mssqlserver/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/mssqlserver](/docs/examples/mssqlserver) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `MSSQLServer` Replicaset using a supported version by `KubeDB` operator. Then we are going to apply `MSSQLServerOpsRequest` to reconfigure its configuration.

### Prepare MSSQLServer Replicaset

Now, we are going to deploy a `MSSQLServer` Replicaset database with version `4.4.26`.

### Deploy MSSQLServer 

At first, we will create `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 10000
```
Here, `maxIncomingConnections` is set to `10000`, whereas the default value is `65536`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo mg-custom-config --from-file=./mongod.conf
secret/mg-custom-config created
```

In this section, we are going to create a MSSQLServer object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `MSSQLServer` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MSSQLServer
metadata:
  name: mg-replicaset
  namespace: demo
spec:
  version: "4.4.26"
  replicas: 3
  replicaSet:
    name: rs0
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configSecret:
    name: mg-custom-config
```

Let's create the `MSSQLServer` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure/mg-replicaset-config.yaml
MSSQLServer.kubedb.com/mg-replicaset created
```

Now, wait until `mg-replicaset` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo                                                                                                                                            
NAME            VERSION   STATUS   AGE
mg-replicaset   4.4.26     Ready    19m
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a MSSQLServer instance,
```bash
$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\username}' | base64 -d                                                                       
root

$ kubectl get secrets -n demo mg-replicaset-auth -o jsonpath='{.data.\password}' | base64 -d                                                                         
nrKuxni0wDSMrgwy
```

Now let's connect to a MSSQLServer instance and run a MSSQLServer internal command to check the configuration we have provided.

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet                        
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--replSet=rs0",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "*",
			"ipv6" : true,
			"maxIncomingConnections" : 10000,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "rs0"
		},
		"security" : {
			"authorization" : "enabled",
			"clusterAuthMode" : "keyFile",
			"keyFile" : "/data/configdb/key.txt"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1614668500, 1),
		"signature" : {
			"hash" : BinData(0,"7sh886HhsNYajGxYGp5Jxi52IzA="),
			"keyId" : NumberLong("6934943333319966722")
		}
	},
	"operationTime" : Timestamp(1614668500, 1)
}
```

As we can see from the configuration of ready MSSQLServer, the value of `maxIncomingConnections` has been set to `10000`.

### Reconfigure using new config secret

Now we will reconfigure this database to set `maxIncomingConnections` to `20000`. 

Now, we will edit the `mongod.conf` file containing required configuration settings.

```ini
$ cat mongod.conf
net:
   maxIncomingConnections: 20000
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./mongod.conf
secret/new-custom-config created
```

#### Create MSSQLServerOpsRequest

Now, we will use this secret to replace the previous secret using a `MSSQLServerOpsRequest` CR. The `MSSQLServerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mops-reconfigure-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  configuration:
    replicaSet:
      configSecret:
        name: new-custom-config
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-replicaset` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.customConfig.replicaSet.configSecret.name` specifies the name of the new secret.
- `spec.customConfig.arbiter.configSecret.name` could also be specified with a config-secret.
- Have a look [here](/docs/guides/mssqlserver/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure/mops-reconfigure-replicaset.yaml
MSSQLServeropsrequest.ops.kubedb.com/mops-reconfigure-replicaset created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `MSSQLServer` object.

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ watch kubectl get MSSQLServeropsrequest -n demo
Every 2.0s: kubectl get MSSQLServeropsrequest -n demo
NAME                          TYPE          STATUS       AGE
mops-reconfigure-replicaset   Reconfigure   Successful   113s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe MSSQLServeropsrequest -n demo mops-reconfigure-replicaset 
Name:         mops-reconfigure-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T07:04:31Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:configuration:
          .:
          f:replicaSet:
            .:
            f:configSecret:
              .:
              f:name:
        f:databaseRef:
          .:
          f:name:
        f:readinessCriteria:
          .:
          f:objectsCountDiffPercentage:
          f:oplogMaxLagSeconds:
        f:timeout:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T07:04:31Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:configuration:
          f:replicaSet:
            f:podTemplate:
              .:
              f:controller:
              f:metadata:
              f:spec:
                .:
                f:resources:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-02T07:04:31Z
  Resource Version:  29869
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mssqlserveropsrequests/mops-reconfigure-replicaset
  UID:               064733d6-19db-4153-82f7-bc0580116ee6
Spec:
  Apply: IfReady
  Configuration:
    Replica Set:
      Config Secret:
        Name:  new-custom-config
  Database Ref:
    Name:  mg-replicaset
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2021-03-02T07:04:31Z
    Message:               MSSQLServer ops request is reconfiguring database
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2021-03-02T07:06:21Z
    Message:               Successfully Reconfigured MSSQLServer
    Observed Generation:   1
    Reason:                ReconfigureReplicaset
    Status:                True
    Type:                  ReconfigureReplicaset
    Last Transition Time:  2021-03-02T07:06:21Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                        Message
  ----    ------                 ----   ----                        -------
  Normal  PauseDatabase          2m55s  KubeDB Ops-manager operator  Pausing MSSQLServer demo/mg-replicaset
  Normal  PauseDatabase          2m55s  KubeDB Ops-manager operator  Successfully paused MSSQLServer demo/mg-replicaset
  Normal  ReconfigureReplicaset  65s    KubeDB Ops-manager operator  Successfully Reconfigured MSSQLServer
  Normal  ResumeDatabase         65s    KubeDB Ops-manager operator  Resuming MSSQLServer demo/mg-replicaset
  Normal  ResumeDatabase         65s    KubeDB Ops-manager operator  Successfully resumed MSSQLServer demo/mg-replicaset
  Normal  Successful             65s    KubeDB Ops-manager operator  Successfully Reconfigured Database
```

Now let's connect to a MSSQLServer instance and run a MSSQLServer internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--replSet=rs0",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "*",
			"ipv6" : true,
			"maxIncomingConnections" : 20000,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "rs0"
		},
		"security" : {
			"authorization" : "enabled",
			"clusterAuthMode" : "keyFile",
			"keyFile" : "/data/configdb/key.txt"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1614668887, 1),
		"signature" : {
			"hash" : BinData(0,"5q35Y51+YpbVHFKoaU7lUWi38oY="),
			"keyId" : NumberLong("6934943333319966722")
		}
	},
	"operationTime" : Timestamp(1614668887, 1)
}
```

As we can see from the configuration of ready MSSQLServer, the value of `maxIncomingConnections` has been changed from `10000` to `20000`. So the reconfiguration of the database is successful.


### Reconfigure using apply config

Now we will reconfigure this database again to set `maxIncomingConnections` to `30000`. This time we won't use a new secret. We will use the `applyConfig` field of the `MSSQLServerOpsRequest`. This will merge the new config in the existing secret.

#### Create MSSQLServerOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `MSSQLServerOpsRequest` CR. The `MSSQLServerOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MSSQLServerOpsRequest
metadata:
  name: mops-reconfigure-apply-replicaset
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: mg-replicaset
  configuration:
    replicaSet:
      applyConfig:
        mongod.conf: |-
          net:
            maxIncomingConnections: 30000
  readinessCriteria:
    oplogMaxLagSeconds: 20
    objectsCountDiffPercentage: 10
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `mops-reconfigure-apply-replicaset` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.replicaSet.applyConfig` specifies the new configuration that will be merged in the existing secret.
- `spec.customConfig.arbiter.configSecret.name` could also be specified with a config-secret.
- Have a look [here](/docs/guides/mssqlserver/concepts/opsrequest.md#specreadinesscriteria) on the respective sections to understand the `readinessCriteria`, `timeout` & `apply` fields.

Let's create the `MSSQLServerOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mssqlserver/reconfigure/mops-reconfigure-apply-replicaset.yaml
MSSQLServeropsrequest.ops.kubedb.com/mops-reconfigure-apply-replicaset created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `MSSQLServerOpsRequest` to be `Successful`.  Run the following command to watch `MSSQLServerOpsRequest` CR,

```bash
$ watch kubectl get MSSQLServeropsrequest -n demo
Every 2.0s: kubectl get MSSQLServeropsrequest -n demo
NAME                               TYPE          STATUS       AGE
mops-reconfigure-apply-replicaset   Reconfigure   Successful   109s
```

We can see from the above output that the `MSSQLServerOpsRequest` has succeeded. If we describe the `MSSQLServerOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe MSSQLServeropsrequest -n demo mops-reconfigure-apply-replicaset
Name:         mops-reconfigure-apply-replicaset
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MSSQLServerOpsRequest
Metadata:
  Creation Timestamp:  2021-03-02T07:09:39Z
  Generation:          1
  Managed Fields:
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .:
          f:kubectl.kubernetes.io/last-applied-configuration:
      f:spec:
        .:
        f:apply:
        f:configuration:
          .:
          f:replicaSet:
            .:
            f:applyConfig:
        f:databaseRef:
          .:
          f:name:
        f:readinessCriteria:
          .:
          f:objectsCountDiffPercentage:
          f:oplogMaxLagSeconds:
        f:timeout:
        f:type:
    Manager:      kubectl-client-side-apply
    Operation:    Update
    Time:         2021-03-02T07:09:39Z
    API Version:  ops.kubedb.com/v1alpha1
    Fields Type:  FieldsV1
    fieldsV1:
      f:spec:
        f:configuration:
          f:replicaSet:
            f:podTemplate:
              .:
              f:controller:
              f:metadata:
              f:spec:
                .:
                f:resources:
      f:status:
        .:
        f:conditions:
        f:observedGeneration:
        f:phase:
    Manager:         kubedb-enterprise
    Operation:       Update
    Time:            2021-03-02T07:09:39Z
  Resource Version:  31005
  Self Link:         /apis/ops.kubedb.com/v1alpha1/namespaces/demo/mssqlserveropsrequests/mops-reconfigure-apply-replicaset
  UID:               0137442b-1b04-43ed-8de7-ecd913b44065
Spec:
  Apply: IfReady
  Configuration:
    Replica Set:
      Apply Config:  net:
  maxIncomingConnections: 30000

  Database Ref:
    Name:  mg-replicaset
  Readiness Criteria:
    Objects Count Diff Percentage:  10
    Oplog Max Lag Seconds:          20
  Timeout:                          5m
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2021-03-02T07:09:39Z
    Message:               MSSQLServer ops request is reconfiguring database
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2021-03-02T07:11:14Z
    Message:               Successfully Reconfigured MSSQLServer
    Observed Generation:   1
    Reason:                ReconfigureReplicaset
    Status:                True
    Type:                  ReconfigureReplicaset
    Last Transition Time:  2021-03-02T07:11:14Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason                 Age    From                        Message
  ----    ------                 ----   ----                        -------
  Normal  PauseDatabase          9m20s  KubeDB Ops-manager operator  Pausing MSSQLServer demo/mg-replicaset
  Normal  PauseDatabase          9m20s  KubeDB Ops-manager operator  Successfully paused MSSQLServer demo/mg-replicaset
  Normal  ReconfigureReplicaset  7m45s  KubeDB Ops-manager operator  Successfully Reconfigured MSSQLServer
  Normal  ResumeDatabase         7m45s  KubeDB Ops-manager operator  Resuming MSSQLServer demo/mg-replicaset
  Normal  ResumeDatabase         7m45s  KubeDB Ops-manager operator  Successfully resumed MSSQLServer demo/mg-replicaset
  Normal  Successful             7m45s  KubeDB Ops-manager operator  Successfully Reconfigured Database
```

Now let's connect to a MSSQLServer instance and run a MSSQLServer internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  mg-replicaset-0  -- mongo admin -u root -p nrKuxni0wDSMrgwy --eval "db._adminCommand( {getCmdLineOpts: 1})" --quiet
{
	"argv" : [
		"mongod",
		"--dbpath=/data/db",
		"--auth",
		"--ipv6",
		"--bind_ip_all",
		"--port=27017",
		"--tlsMode=disabled",
		"--replSet=rs0",
		"--keyFile=/data/configdb/key.txt",
		"--clusterAuthMode=keyFile",
		"--config=/data/configdb/mongod.conf"
	],
	"parsed" : {
		"config" : "/data/configdb/mongod.conf",
		"net" : {
			"bindIp" : "*",
			"ipv6" : true,
			"maxIncomingConnections" : 30000,
			"port" : 27017,
			"tls" : {
				"mode" : "disabled"
			}
		},
		"replication" : {
			"replSet" : "rs0"
		},
		"security" : {
			"authorization" : "enabled",
			"clusterAuthMode" : "keyFile",
			"keyFile" : "/data/configdb/key.txt"
		},
		"storage" : {
			"dbPath" : "/data/db"
		}
	},
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1614669580, 1),
		"signature" : {
			"hash" : BinData(0,"u/xTAa4aW/8bsRvBYPffwQCeTF0="),
			"keyId" : NumberLong("6934943333319966722")
		}
	},
	"operationTime" : Timestamp(1614669580, 1)
}
```

As we can see from the configuration of ready MSSQLServer, the value of `maxIncomingConnections` has been changed from `20000` to `30000`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete mg -n demo mg-replicaset
kubectl delete MSSQLServeropsrequest -n demo mops-reconfigure-replicaset mops-reconfigure-apply-replicaset
```