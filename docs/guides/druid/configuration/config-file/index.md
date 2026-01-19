---
title: Configuring Druid Cluster
menu:
  docs_{{ .version }}:
    identifier: guides-druid-configuration-config-file
    name: Configuration File
    parent: guides-druid-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

> New to KubeDB? Please start [here](/docs/README.md).

# Configure Druid Cluster

In Druid cluster, there are six nodes available coordinators, overlords, brokers, routers, historicals, middleManagers. In this tutorial, we will see how to configure each node of a druid cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends on ZooKeeper for external dependency with helm command.

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

Before deploying `Druid` cluster, we need to prepare the external dependencies.

## Create External Dependency (Deep Storage)

Before proceeding further, we need to prepare deep storage, which is one of the external dependency of Druid and used for storing the segments. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/application-level/examples/deep-storage-config.yaml
secret/deep-storage-config created
```

## Use Custom Configuration

Say we want to change the default maximum number of tasks the MiddleManager can accept. Let's create the `middleManagers.properties` file with our desire configurations.

**middleManagers.properties:**

```properties
druid.worker.capacity=5
```

and we also want to change the number of processing threads to have available for parallel processing of segments of the historicals nodes. Let's create the `historicals.properties` file with our desire configurations.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/config-file/yamls/config-secret.yaml
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
  configuration:
    secretName: config-secret
  deepStorage:
    type: s3
    configuration:
      secretName: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```

Now, create the Druid object by the following command:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/configuration/config-file/yamls/druid-with-monitoring.yaml
druid.kubedb.com/druid-with-config created
```

Now, wait for the Druid to become ready:

```bash
$ kubectl get dr -n demo -w
NAME                TYPE                  VERSION   STATUS         AGE
druid-with-config   kubedb.com/v1alpha2   28.0.1     Provisioning   5s
druid-with-config   kubedb.com/v1alpha2   28.0.1    Provisioning   7s
.
.
druid-with-config   kubedb.com/v1alpha2   28.0.1     Ready          2m
```

## Verify Configuration

Lets exec into one of the druid middleManagers pod that we have created and check the configurations are applied or not:

Exec into the Druid middleManagers:

```bash
$ kubectl exec -it -n demo druid-with-config-middleManagers-0 -- bash
Defaulted container "druid" out of: druid, init-druid (init)
bash-5.1$   
```

Now, execute the following commands to see the configurations:
```bash
bash-5.1$ cat conf/druid/cluster/data/middleManager/runtime.properties | grep druid.worker.capacity
druid.worker.capacity=5
```
Here, we can see that our given configuration is applied to the Druid cluster for all brokers.

Now, lets exec into one of the druid historicals pod that we have created and check the configurations are applied or not:

Exec into the Druid historicals:

```bash
$ kubectl exec -it -n demo druid-with-config-historicals-0 -- bash
Defaulted container "druid" out of: druid, init-druid (init)
bash-5.1$   
```

Now, execute the following commands to see the metadata storage directory:
```bash
bash-5.1$ cat conf/druid/cluster/data/historical/runtime.properties | grep druid.processing.numThreads 
druid.processing.numThreads=3
```

Here, we can see that our given configuration is applied to the historicals.

### Verify Configuration Change from Druid UI
You can also see the configuration changes from the druid ui. For that, follow the following steps:

First port-forward the port `8888` to local machine:

```bash
$ kubectl port-forward -n demo svc/druid-with-config-routers 8888
Forwarding from 127.0.0.1:8888 -> 8888
Forwarding from [::1]:8888 -> 8888
```


Now hit the `http://localhost:8888` from any browser, and you will be prompted to provide the credential of the druid database. By following the steps discussed below, you can get the credential generated by the KubeDB operator for your Druid database.

**Connection information:**

- Username:

  ```bash
  $ kubectl get secret -n demo druid-with-config-admin-cred -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password:

  ```bash
  $ kubectl get secret -n demo druid-with-config-admin-cred -o jsonpath='{.data.password}' | base64 -d
  LzJtVRX5E8MorFaf
  ```

After providing the credentials correctly, you should be able to access the web console like shown below.

<p align="center">
  <img alt="druid-ui"  src="/docs/guides/druid/configuration/config-file/images/druid-updated-ui.png">
</p>


You can see that there are 5 task slots reflecting with our provided custom configuration of `druid.worker.capacity=5`.

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
