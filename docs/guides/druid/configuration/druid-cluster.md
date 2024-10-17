---
title: Configuring Druid Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-druid-configuration-druid-cluster
    name: Configuration
    parent: guides-druid-configuration
    weight: 10
menu_name: docs_{{ .version }}
---

> New to KubeDB? Please start [here](/docs/README.md).

> New to KubeDB? Please start [here](/docs/README.md).

# Configure Druid Cluster

In Druid cluster, there are six nodes available coordinators, overlords, brokers, routers, historicals, middleManagers. In this tutorial, we will see how to configure each node of a druid cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- - Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends on ZooKeeper for external dependency with helm command.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create namespace demo
namespace/demo created

$ kubectl get namespace
NAME                 STATUS   AGE
demo                 Active   9s
```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/druid/configuration/yamls) in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Find Available StorageClass

We will have to provide `StorageClass` in Druid CR specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
NAME                 PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
standard (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  1h
```

Here, we have `standard` StorageClass in our cluster from [Local Path Provisioner](https://github.com/rancher/local-path-provisioner).

## Use Custom Configuration

Say we want to change the default log retention time and default replication factor of creating a topic of brokers. Let's create the `middleManagers.properties` file with our desire configurations.

**middleManagers.properties:**

```properties
druid.worker.capacity=5
```

and we also want to change the metadata.log.dir of the all historicals nodes. Let's create the `historicals.properties` file with our desire configurations.

**historicals.properties:**

```properties
druid.processing.numThreads=3
```

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: configsecret
  namespace: demo
stringData:
  middleManagers.properties: |-
    druid.worker.capacity=5
  historicals.properties: |-
    druid.processing.numThreads=3
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/yamls/config-secret.yaml
secret/config-secret created
```

> To provide custom configuration for other nodes add values for the following `key` under `stringData`:
>   - Use `common.runtime.properties` for common configurations
>   - Use `coordinators.properties` for configurations of coordinators 
>   - Use `overlords.properties` for configurations of overlords
>   - Use `brokers.properties` for configurations of brokers
>   - Use `routers.properties` for configurations of routers

Now that the config secret is created, it needs to be mentioned in the [Druid](/docs/guides/druid/concepts/druid.md) object's yaml:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: druid-with-config
  namespace: demo
spec:
  version: 28.0.1
  configSecret:
    name: config-secret
  deepStorage:
    type: s3
    configSecret:
      name: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```

Now, create the Druid object by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/yamls/druid-with-config.yaml
druid.kubedb.com/druid-with-config created
```

Now, wait for the Druid to become ready:

```bash
$ kubectl get dr -n demo -w
NAME                TYPE            VERSION   STATUS         AGE
druid-with-config   kubedb.com/v1   3.6.1     Provisioning   5s
druid-with-config   kubedb.com/v1   3.6.1     Provisioning   7s
.
.
druid-with-config   kubedb.com/v1   3.6.1     Ready          2m
```

## Verify Configuration

Lets exec into one of the druid middleManagers pod that we have created and check the configurations are applied or not:

Exec into the Druid middleManagers:

```bash
$ kubectl exec -it -n demo druid-with-config-middleManagers-0 -- bash
druid@druid-with-config-middleManagers-0:~$ 
```

Now, execute the following commands to see the configurations:
```bash
druid@druid-with-config-broker-0:~$ druid-configs.sh --bootstrap-server localhost:9092 --command-config /opt/druid/config/clientauth.properties --describe --entity-type brokers --all | grep log.retention.hours
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
  log.retention.hours=100 sensitive=false synonyms={STATIC_BROKER_CONFIG:log.retention.hours=100, DEFAULT_CONFIG:log.retention.hours=168}
druid@druid-with-config-broker-0:~$ druid-configs.sh --bootstrap-server localhost:9092 --command-config /opt/druid/config/clientauth.properties --describe --entity-type brokers --all | grep default.replication.factor
  default.replication.factor=2 sensitive=false synonyms={STATIC_BROKER_CONFIG:default.replication.factor=2, DEFAULT_CONFIG:default.replication.factor=1}
  default.replication.factor=2 sensitive=false synonyms={STATIC_BROKER_CONFIG:default.replication.factor=2, DEFAULT_CONFIG:default.replication.factor=1}
```
Here, we can see that our given configuration is applied to the Druid cluster for all brokers.

Now, let's exec into one of the druid controller pod that we have created and check the configurations are applied or not:

Exec into the Druid controller:

```bash
$ kubectl exec -it -n demo druid-with-config-controller-0 -- bash
druid@druid-with-config-controller-0:~$ 
```

Now, execute the following commands to see the metadata storage directory:
```bash
druid@druid-with-config-controller-0:~$ ls /var/log/druid/
1000  cluster_id  metadata-custom
```

Here, we can see that our given configuration is applied to the controller. Metadata log directory is changed to `/var/log/druid/metadata-custom`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete dr -n demo druid-dev 

$ kubectl delete secret -n demo configsecret-combined 

$ kubectl delete namespace demo
```

## Next Steps

- Detail concepts of [Druid object](/docs/guides/druid/concepts/druid.md).
- Different Druid topology clustering modes [here](/docs/guides/druid/clustering/_index.md).

[//]: # (- Monitor your Druid database with KubeDB using [out-of-the-box Prometheus operator]&#40;/docs/guides/druid/monitoring/using-prometheus-operator.md&#41;.)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

