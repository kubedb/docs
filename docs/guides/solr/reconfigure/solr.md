---
title: Reconfigure Solr
menu:
  docs_{{ .version }}:
    identifier: sl-reconfigure-solr
    name: Reconfigure Solr
    parent: sl-reconfigure
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
    - [SolrOpsRequest](/docs/guides/solr/concepts/solropsrequests.md)
    - [Reconfigure Overview](/docs/guides/solr/reconfigure/overview.md)

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure/sl-custom-config.yaml
secret/sl-custom-config created
```

In this section, we are going to create a Solr object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `Solr` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Solr
metadata:
  name: solr
  namespace: demo
spec:
  configSecret:
    name: sl-custom-config
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/Solr/reconfigure/solr.yaml
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
Here, we can see that our given configuration is applied to the Solr cluster. `maxBooleanClauses` is set to `2024`.

### Reconfigure using new config secret

Now we will reconfigure this cluster to set `maxBooleanClauses` to `2030`.

Now, update our `solr.xml` file with the new configuration.

**server.properties:**

```properties
<int name="maxBooleanClauses">${solr.max.booleanClauses:2030}</int>
```

Then, we will create a new secret with this configuration file.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: new-sl-custom-config
  namespace: demo
stringData:
  "solr.xml": |
    <solr>
      <int name="maxBooleanClauses">${solr.max.booleanClauses:2030}</int>
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure/new-sl-custom-config.yaml
secret/new-sl-custom-config created
```

#### Create SolrOpsRequest

Now, we will use this secret to replace the previous secret using a `SolrOpsRequest` CR. The `SolrOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: sl-reconfigure-custom-config
  namespace: demo
spec:
  apply: IfReady
  configuration:
    configSecret:
      name: new-sl-custom-config
  databaseRef:
    name: solr
  type: Reconfigure
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `Solr-dev` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configSecret.name` specifies the name of the new secret.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure/sl-reconfigure-custom-config.yaml
solropsrequest.ops.kubedb.com/sl-reconfigure-custom-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Solr` object.

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CR,

```bash
$ kubectl get slops -n demo 
NAME                           TYPE          STATUS       AGE
sl-reconfigure-custom-config   Reconfigure   Successful   5m24s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe slops -n demo sl-reconfigure-custom-config 
Name:         sl-reconfigure-custom-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-05T12:59:25Z
  Generation:          1
  Resource Version:    1665913
  UID:                 7bb29ead-8322-4ac3-9375-6dd8594882d1
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-sl-custom-config
  Database Ref:
    Name:  solr
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-11-05T12:59:25Z
    Message:               Solr ops-request has started to reconfigure Solr nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-11-05T12:59:33Z
    Message:               successfully reconciled the Solr with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-05T13:01:39Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-05T12:59:38Z
    Message:               get pod; ConditionStatus:True; PodName:solr-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-0
    Last Transition Time:  2024-11-05T12:59:38Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-0
    Last Transition Time:  2024-11-05T12:59:43Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-05T13:00:38Z
    Message:               get pod; ConditionStatus:True; PodName:solr-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-1
    Last Transition Time:  2024-11-05T13:00:38Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-1
    Last Transition Time:  2024-11-05T13:01:39Z
    Message:               Successfully completed reconfigure Solr
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  5m45s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/sl-reconfigure-custom-config
  Normal   Starting                                                  5m45s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr
  Normal   Successful                                                5m45s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr for SolrOpsRequest: sl-reconfigure-custom-config
  Normal   UpdatePetSets                                             5m37s  KubeDB Ops-manager Operator  successfully reconciled the Solr with new configure
  Warning  get pod; ConditionStatus:True; PodName:solr-0    5m32s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-0  5m32s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-0
  Warning  running pod; ConditionStatus:False                        5m27s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-1    4m32s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-1
  Warning  evict pod; ConditionStatus:True; PodName:solr-1  4m32s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-1
  Normal   RestartNodes                                              3m31s  KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  3m31s  KubeDB Ops-manager Operator  Resuming Solr database: demo/solr
  Normal   Successful                                                3m31s  KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr for SolrOpsRequest: sl-reconfigure-custom-config
  Normal   RestartNodes                                              3m31s  KubeDB Ops-manager Operator  Successfully restarted all nodes
```

Now let's exec one of the instance and cat solr.xml file to check the new configuration we have provided.

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
  <int name="maxBooleanClauses">${solr.max.booleanClauses:2030}</int>
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

As we can see from the configuration of ready Solr, the value of `log.retention.hours` has been changed from `2024` to `2030`. So the reconfiguration of the cluster is successful.


### Reconfigure using apply config

Now we will reconfigure this cluster again to set `maxBooleanClauses` to `2024`. This time we won't use a new secret. We will use the `applyConfig` field of the `SolrOpsRequest`. This will merge the new config in the existing secret.

#### Create SolrOpsRequest

Now, we will use the new configuration in the `applyConfig` field in the `SolrOpsRequest` CR. The `SolrOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: SolrOpsRequest
metadata:
  name: sl-reconfigure-apply-config
  namespace: demo
spec:
  apply: IfReady
  configuration:
     applyConfig:
       solr.xml: |
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
  databaseRef:
    name: solr
  type: Reconfigure
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `solr` cluster.
- `spec.type` specifies that we are performing `Reconfigure` on Solr.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `SolrOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/solr/reconfigure/sl-reconfigure-apply-config.yaml
Solropsrequest.ops.kubedb.com/sl-reconfigure-apply-config created
```

#### Verify the new configuration is working

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `SolrOpsRequest` to be `Successful`.  Run the following command to watch `SolrOpsRequest` CR,

```bash
$ kubectl get slops -n demo
NAME                           TYPE          STATUS       AGE
sl-reconfigure-custom-config   Reconfigure   Successful   2m22s
```

We can see from the above output that the `SolrOpsRequest` has succeeded. If we describe the `SolrOpsRequest` we will get an overview of the steps that were followed to reconfigure the cluster.

```bash
$ kubectl describe slops -n demo sl-reconfigure-custom-config 
Name:         sl-reconfigure-custom-config
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         SolrOpsRequest
Metadata:
  Creation Timestamp:  2024-11-05T13:09:55Z
  Generation:          1
  Resource Version:    1666897
  UID:                 3fd6a300-5ed2-4c0d-b1fd-9102a44b37ce
Spec:
  Apply:  IfReady
  Configuration:
    Apply Config:
      solr.xml:  <solr>
  <int name="maxBooleanClauses">${solr.max.booleanClauses:2024}</int>
  <backup>
    <repository name="kubedb-proxy-s3" class="org.apache.solr.s3.S3BackupRepository">
      <str name="s3.bucket.name">solrbackup</str>
      <str name="s3.region">us-east-1</str>
      <str name="s3.endpoint">http://s3proxy-s3.demo.svc:80</str>
    </repository>
  </backup>
</solr>

  Database Ref:
    Name:  solr
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-11-05T13:09:55Z
    Message:               Solr ops-request has started to reconfigure Solr nodes
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-11-05T13:09:59Z
    Message:               Successfully prepared user provided custom config secret
    Observed Generation:   1
    Reason:                PrepareCustomConfig
    Status:                True
    Type:                  PrepareCustomConfig
    Last Transition Time:  2024-11-05T13:10:04Z
    Message:               successfully reconciled the Solr with new configure
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-11-05T13:11:55Z
    Message:               Successfully restarted all nodes
    Observed Generation:   1
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2024-11-05T13:10:09Z
    Message:               get pod; ConditionStatus:True; PodName:solr-0
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-0
    Last Transition Time:  2024-11-05T13:10:09Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-0
    Last Transition Time:  2024-11-05T13:10:14Z
    Message:               running pod; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  RunningPod
    Last Transition Time:  2024-11-05T13:10:54Z
    Message:               get pod; ConditionStatus:True; PodName:solr-1
    Observed Generation:   1
    Status:                True
    Type:                  GetPod--solr-1
    Last Transition Time:  2024-11-05T13:10:54Z
    Message:               evict pod; ConditionStatus:True; PodName:solr-1
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--solr-1
    Last Transition Time:  2024-11-05T13:11:55Z
    Message:               Successfully completed reconfigure Solr
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                    Age    From                         Message
  ----     ------                                                    ----   ----                         -------
  Normal   Starting                                                  2m52s  KubeDB Ops-manager Operator  Start processing for SolrOpsRequest: demo/sl-reconfigure-custom-config
  Normal   Starting                                                  2m51s  KubeDB Ops-manager Operator  Pausing Solr databse: demo/solr
  Normal   Successful                                                2m51s  KubeDB Ops-manager Operator  Successfully paused Solr database: demo/solr for SolrOpsRequest: sl-reconfigure-custom-config
  Normal   UpdatePetSets                                             2m43s  KubeDB Ops-manager Operator  successfully reconciled the Solr with new configure
  Warning  get pod; ConditionStatus:True; PodName:solr-0    2m38s  KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-0
  Warning  evict pod; ConditionStatus:True; PodName:solr-0  2m38s  KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-0
  Warning  running pod; ConditionStatus:False                        2m33s  KubeDB Ops-manager Operator  running pod; ConditionStatus:False
  Warning  get pod; ConditionStatus:True; PodName:solr-1    113s   KubeDB Ops-manager Operator  get pod; ConditionStatus:True; PodName:solr-1
  Warning  evict pod; ConditionStatus:True; PodName:solr-1  113s   KubeDB Ops-manager Operator  evict pod; ConditionStatus:True; PodName:solr-1
  Normal   RestartNodes                                              52s    KubeDB Ops-manager Operator  Successfully restarted all nodes
  Normal   Starting                                                  52s    KubeDB Ops-manager Operator  Resuming Solr database: demo/solr
  Normal   Successful                                                52s    KubeDB Ops-manager Operator  Successfully resumed Solr database: demo/solr for SolrOpsRequest: sl-reconfigure-custom-config
```

Now let's exec into one of the instance and cat `solr.xml` file to check the new configuration we have provided.

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

As we can see from the configuration of ready Solr, the value of `maxBooleanClauses` has been changed from `2030` to `2024`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete sl -n demo solr
kubectl delete solropsrequest -n demo sl-reconfigure-custom-config sl-reconfigure-apply-config
kubectl delete secret -n demo sl-custom-config new-sl-custom-config
kubectl delete namespace demo
```

## Next Steps

- Detail concepts of [Solr object](/docs/guides/solr/concepts/solr.md).
- Different solr topology clustering modes [here](/docs/guides/solr/clustering/topology_cluster.md).
- Monitor your Solr database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/solr/monitoring/prometheus-operator.md).

- Monitor your Solr database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/solr/monitoring/prometheus-builtin.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
