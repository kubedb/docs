---
title: Run SingleStore with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-configuration-using-podtemplate
    name: Customize PodTemplate
    parent: guides-sdb-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run SingleStore with Custom PodTemplate

KubeDB supports providing custom configuration for SingleStore via [PodTemplate](/docs/guides/singlestore/concepts/singlestore.md#spec.topology). This tutorial will show you how to use KubeDB to run a SingleStore database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/singlestore/configuration/podtemplating/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for `leaf` and `aggregator` pod through `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate`. KubeDB operator will pass the information provided in `spec.topology.aggregator.podTemplate` and `spec.topology.leaf.podTemplate` to the `aggregator` and `leaf` PetSet created for SingleStore database.

KubeDB accept following fields to set in `podTemplate:`

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

Read about the fields in details in [PodTemplate concept](/docs/guides/mysql/concepts/database/index.md#specpodtemplate),

## CRD Configuration

Below is the YAML for the SingleStore created in this example. Here, [`spec.topology.aggregator/leaf.podTemplate.spec.args`](/docs/guides/mysql/concepts/database/index.md#specpodtemplatespecargs) provides extra arguments.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: sdb-misc-config
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 1
      podTemplate:
        spec:
          containers:
          - name: singlestore
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
            args:
              - --character-set-server=utf8mb4
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: singlestore
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"     
              args:
                - --character-set-server=utf8mb4              
      storage:
        storageClassName: "standard"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/podtemplating/yamls/sdb-misc-config.yaml
singlestore.kubedb.com/sdb-misc-config created
```

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc. If everything goes well, we will see that a pod with the name `sdb-misc-config-aggregator-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                           READY   STATUS    RESTARTS   AGE
sdb-misc-config-aggregator-0   2/2     Running   0          4m51s
sdb-misc-config-leaf-0         2/2     Running   0          4m48s
sdb-misc-config-leaf-1         2/2     Running   0          4m30s
```

Now, we will check if the database has started with the custom configuration we have provided.

```bash
$ kubectl exec -it -n demo sdb-misc-config-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@sdb-misc-config-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 311
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

singlestore> SHOW VARIABLES LIKE 'char%';
+--------------------------+------------------------------------------------------+
| Variable_name            | Value                                                |
+--------------------------+------------------------------------------------------+
| character_set_client     | utf8mb4                                              |
| character_set_connection | utf8mb4                                              |
| character_set_database   | utf8mb4                                              |
| character_set_filesystem | binary                                               |
| character_set_results    | utf8mb4                                              |
| character_set_server     | utf8mb4                                              |
| character_set_system     | utf8                                                 |
| character_sets_dir       | /opt/memsql-server-8.7.10-95e2357384/share/charsets/ |
+--------------------------+------------------------------------------------------+
8 rows in set (0.00 sec)

singlestore> exit
Bye

```

Here we can see the character_set_server value is utf8mb4.  

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete singlestore -n demo sdb-misc-config

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart SingleStore](/docs/guides/singlestore/quickstart/quickstart.md) with KubeDB Operator.
- Initialize [SingleStore with Script](/docs/guides/singlestore/initialization).
- Detail concepts of [SingleStore object](/docs/guides/singlestore/concepts/singlestore.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
