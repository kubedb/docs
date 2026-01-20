---
title: Custom Configuration With Config Files
menu:
  docs_{{ .version }}:
    identifier: sl-custom-config-file
    name: Config Files
    parent: sl-custom-config
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Solr Cluster

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Solr cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [Solr](/docs/guides/solr/concepts/solr.md)
    - [Combined](/docs/guides/solr/clustering/combined_cluster.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/Solr](/docs/examples/solr) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Solr` cluster using a supported version by `KubeDB` operator. Then we are going to apply `SolrOpsRequest` to reconfigure its configuration.

### Prepare Solr Cluster

Now, we are going to deploy a `Solr` cluster with version `9.6.1`.

### Deploy Solr

At first, we will create a secret with the `solr.xml` attribute containing required configuration settings.

**server.properties:**

```properties
<int name="maxBooleanClauses">${solr.max.booleanClauses:2024}</int>
```
Here, `maxBooleanClauses` is set to `2024`, whereas the default value is `1024`.

Let's create a k8s secret containing the above configuration where the file name will be the key and the file-content as the value:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: sl-custom-config
  namespace: demo
stringData:
  "solr.xml": |
    <solr>
      <int name="maxBooleanClauses">${solr.max.booleanClauses:2024}</int>
      <backup>
        <repository name="kubedb-proxy-s3" class="org.apache.solr.s3.S3BackupRepository">
          <str name="s3.bucket.name">solrbackup</str>
          <str name="s3.region">us-east-1</str>
          <str name="s3.endpoint">http://s3proxy-s3.demo.svc:80</str>
        </repository>
      </backup>
    </solr>
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/configuration/sl-custom-config.yaml
secret/sl-custom-config created
```

In this section, we are going to create a Solr object specifying `spec.configuration` field to apply this custom configuration. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr
  namespace: demo
spec:
  configuration:
    secretName: sl-custom-config
  version: 9.6.1
  replicas: 2
  zookeeperRef:
    name: zoo
    namespace: demo
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: longhorn
```

Let's create the `Solr` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Solr/configuration/solr.yaml
solr.kubedb.com/solr created
```

Now, wait until `solr` has status `Ready`. i.e,

```bash
$ kubectl get sl -n demo
NAME     TYPE                  VERSION   STATUS   AGE
solr     kubedb.com/v1alpha2   9.6.1     Ready    10m
```

Now, we will check if the Solr has started with the custom configuration we have provided.

Exec into the Solr pod and execute the following commands to see the configurations:
```bash
$ kubectl exec -it -n demo solr-0 -- bash
Defaulted container "solr" out of: solr, init-solr (init)
solr@solr-0:/opt/solr-9.6.1$ cat /var/solr/solr.xml
<?xml version="1.0" encoding="UTF-8"?>
<solr>
  <backup>
    <repository name="kubedb-proxy-s3" class="org.apache.solr.s3.S3BackupRepository">
      <str name="s3.bucket.name">solrbackup</str>
      <str name="s3.region">us-east-1</str>
      <str name="s3.endpoint">http://s3proxy-s3.demo.svc:80</str>
    </repository>
  </backup>
  <str name="coreRootDirectory">/var/solr/data</str>
  <str name="sharedLib">${solr.sharedLib:},/opt/solr/contrib/gcs-repository/lib,/opt/solr/contrib/prometheus-exporter/lib,/opt/solr/contrib/s3-repository/lib,/opt/solr/dist</str>
  <str name="allowPaths">${solr.allowPaths:}</str>
  <int name="maxBooleanClauses">${solr.max.booleanClauses:2024}</int>
  <shardHandlerFactory name="shardHandlerFactory" class="HttpShardHandlerFactory">
    <int name="connTimeout">${connTimeout:60000}</int>
    <int name="socketTimeout">${socketTimeout:600000}</int>
  </shardHandlerFactory>
  <solrcloud>
    <int name="distribUpdateConnTimeout">${distribUpdateConnTimeout:60000}</int>
    <int name="distribUpdateSoTimeout">${distribUpdateSoTimeout:600000}</int>
    <bool name="genericCoreNodeNames">${genericCoreNodeNames:true}</bool>
    <str name="host">${host:}</str>
    <str name="hostContext">${hostContext:solr}</str>
    <int name="hostPort">${solr.port.advertise:80}</int>
    <str name="zkACLProvider">${zkACLProvider:org.apache.solr.common.cloud.DigestZkACLProvider}</str>
    <int name="zkClientTimeout">${zkClientTimeout:30000}</int>
    <str name="zkCredentialsInjector">${zkCredentialsInjector:org.apache.solr.common.cloud.VMParamsZkCredentialsInjector}</str>
    <str name="zkCredentialsProvider">${zkCredentialsProvider:org.apache.solr.common.cloud.DigestZkCredentialsProvider}</str>
  </solrcloud>
  <metrics enabled="${metricsEnabled:true}"/>
</solr>

```

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete solr -n demo solr-combined
kubectl delete secret -n demo sl-custom-config
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different Solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).