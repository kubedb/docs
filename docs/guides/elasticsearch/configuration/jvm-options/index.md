---
title: Configuring Elasticsearch JVM Options
menu:
  docs_{{ .version }}:
    identifier: es-configuration-jvm-options
    name: JVM Options
    parent: es-configuration
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Configure Elasticsearch JVM Options

The Elasticsearch offers users to configure the JVM settings by using `jvm.options` file. The `jvm.options` file located at the `$ES_HOME/config` (ie. `/usr/share/elasticsearch/config`) directory.

## Deploy Elasticsearch with Custom jvm.options File

Before deploying the Elasticsearch instance, you need to create a k8s secret with the custom config files (here: `jvm.options`).

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: es-custom-config
  namespace: demo
stringData:
  jvm.options: |-
    ## G1GC Configuration

    10-:-XX:+UseG1GC
    10-13:-XX:-UseConcMarkSweepGC
    10-13:-XX:-UseCMSInitiatingOccupancyOnly
    10-:-XX:G1ReservePercent=25
    10-:-XX:InitiatingHeapOccupancyPercent=30

    ## JVM temporary directory
    -Djava.io.tmpdir=${ES_TMPDIR}

    ## heap dumps

    # generate a heap dump when an allocation from the Java heap fails
    # heap dumps are created in the working directory of the JVM
    -XX:+HeapDumpOnOutOfMemoryError

    # specify an alternative path for heap dumps; ensure the directory exists and
    # has sufficient space
    -XX:HeapDumpPath=data

    # specify an alternative path for JVM fatal error logs
    -XX:ErrorFile=logs/hs_err_pid%p.log

    # JDK 9+ GC logging
    9-:-Xlog:gc*,gc+age=trace,safepoint:file=logs/gc.log:utctime,pid,tags:filecount=32,filesize=64m
```

If you want to provide node-role specific settings, say you want to configure ingest nodes with a different setting than others in a topology cluster, add node `role` as a prefix in the file name.

```yaml
stringData:
  ingest-jvm.options: |-
    ... ...
  master-jvm.options: |-
    ... ...
  ... ... 
```

Deploy the k8s secret:

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/jvm-options/yamls/custom-config.yaml
secret/es-custom-config created
```

Now Deploy the Elasticsearch Cluster with the custom `jvm.options` file:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Elasticsearch
metadata:
  name: es-test
  namespace: demo
spec:
  # Make sure that you've mentioned the config secret name here
  configSecret:
    name: es-custom-config
  enableSSL: false 
  version: opendistro-1.12.0
  storageType: Durable
  terminationPolicy: WipeOut
  topology:
    master:
      suffix: master
      replicas: 3
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      suffix: data
      replicas: 2 
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 5Gi
    ingest:
      suffix: ingest
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```

Deploy Elasticsearch:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/configuration/jvm-options/yamls/elasticsearch.yaml
elasticsearch/es-test created
```

Wait for the Elasticsearch to become ready:

```bash
$ kubectl get elasticsearch -n demo -w
NAME          VERSION             STATUS         AGE
es-test       opendistro-1.12.0   Provisioning   12s
es-test       opendistro-1.12.0   Provisioning   2m2s
es-test       opendistro-1.12.0   Ready          2m2s
```
