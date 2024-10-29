---
title: Run SingleStore with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-sdb-configuration-using-config-file
    name: Config File
    parent: guides-sdb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for SingleStore. This tutorial will show you how to use KubeDB to run a SingleStore database with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  
  $ kubectl get ns demo
  NAME    STATUS  AGE
  demo    Active  5s
  ```

> Note: YAML files used in this tutorial are stored in [docs/guides/singlestore/configuration/config-file/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/singlestore/configuration/config-file/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

SingleStore allows to configure database via configuration file. The default configuration for SingleStore can be found in `/var/lib/memsql/instance/memsql.cnf` file. When SingleStore starts, it will look for custom configuration file in `/etc/memsql/conf.d` directory. If configuration file exist, SingleStore instance will use combined startup setting from both `/var/lib/memsql/instance/memsql.cnf` and `*.cnf` files in `/etc/memsql/conf.d` directory. This custom configuration will overwrite the existing default one. To know more about configuring SingleStore see [here](https://docs.singlestore.com/db/v8.7/reference/configuration-reference/cluster-config-files/singlestore-server-config-files/).

At first, you have to create a config file with `.cnf` extension with your desired configuration. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You need to specify this volume in the `spec.configSecret` section when creating the SingleStore CRD for `Standalone` mode. Additionally, you can modify your `aggregator` and `leaf` nodes separately by providing separate configSecrets, or use the same one in the `spec.topology.aggregator.configSecret` and `spec.topology.leaf.configSecret` sections when creating the SingleStore CRD for `Cluster` mode. KubeDB will mount this volume into `/etc/memsql/conf.d` directory of the database pod.

In this tutorial, we will configure [max_connections](https://docs.singlestore.com/db/v8.7/reference/configuration-reference/engine-variables/list-of-engine-variables/#in-depth-variable-definitions) and [read_buffer_size](https://dev.mysql.com/doc/refman/8.0/en/server-system-variables.html#sysvar_read_buffer_size) via a custom config file. We will use configMap as volume source.

## Create SingleStore License Secret

We need SingleStore License to create SingleStore Database. So, Ensure that you have acquired a license and then simply pass the license by secret.

```bash
$ kubectl create secret generic -n demo license-secret \
                --from-literal=username=license \
                --from-literal=password='your-license-set-here'
secret/license-secret created
```

## Custom Configuration

At first, let's create `sdb-config.cnf` file setting `max_connections` and `read_buffer_size` parameters.

```bash
cat <<EOF > sdb-config.cnf
[server]
max_connections = 250
read_buffer_size = 1048576
EOF

$ cat sdb-config.cnf
[server]
max_connections = 250
read_buffer_size = 122880
```

Now, create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo sdb-configuration --from-file=./sdb-config.cnf
configmap/sdb-configuration created
```

Verify the secret has the configuration file.

```yaml
$ kubectl get secret -n demo sdb-configuration -o yaml
apiVersion: v1
data:
  sdb-config.cnf: W3NlcnZlcl0KbWF4X2Nvbm5lY3Rpb25zID0gMjUwCnJlYWRfYnVmZmVyX3NpemUgPSAxMjI4ODAK
kind: Secret
metadata:
  creationTimestamp: "2024-10-02T12:54:35Z"
  name: sdb-configuration
  namespace: demo
  resourceVersion: "99627"
  uid: c2621d8e-ebca-4300-af05-0180512ce031
type: Opaque


```

Now, create SingleStore crd specifying `spec.topology.aggregator.configSecret` and `spec.topology.leaf.configSecret` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/singlestore/configuration/config-file/yamls/sdb-custom.yaml
singlestore.kubedb.com/custom-sdb created
```

Below is the YAML for the SingleStore crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Singlestore
metadata:
  name: custom-sdb
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      configSecret:
        name: sdb-configuration
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
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 2
      configSecret:
        name: sdb-configuration
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

Now, wait a few minutes. KubeDB operator will create necessary PVC, petset, services, secret etc.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                      READY   STATUS    RESTARTS   AGE
custom-sdb-aggregator-0   2/2     Running   0          94s
custom-sdb-aggregator-1   2/2     Running   0          88s
custom-sdb-leaf-0         2/2     Running   0          91s
custom-sdb-leaf-1         2/2     Running   0          86s

$ kubectl get sdb -n demo
NAME         TYPE                  VERSION   STATUS   AGE
custom-sdb   kubedb.com/v1alpha2   8.7.10    Ready    4m29s

```

We can see the database is in ready phase so it can accept conncetion.

Now, we will check if the database has started with the custom configuration we have provided.

> Read the comment written for the following commands. They contain the instructions and explanations of the commands.

```bash
# Connceting to the database
$ kubectl exec -it -n demo custom-sdb-aggregator-0 -- bash
Defaulted container "singlestore" out of: singlestore, singlestore-coordinator, singlestore-init (init)
[memsql@custom-sdb-aggregator-0 /]$ memsql -uroot -p$ROOT_PASSWORD
singlestore-client: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 208
Server version: 5.7.32 SingleStoreDB source distribution (compatible; MySQL Enterprise & MySQL Commercial)

Copyright (c) 2000, 2022, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

# value of `max_conncetions` is same as provided 
singlestore> show variables like 'max_connections';
+-----------------+-------+
| Variable_name   | Value |
+-----------------+-------+
| max_connections | 250   |
+-----------------+-------+
1 row in set (0.00 sec)

# value of `read_buffer_size` is same as provided
singlestore> show variables like 'read_buffer_size';
+------------------+--------+
| Variable_name    | Value  |
+------------------+--------+
| read_buffer_size | 122880 |
+------------------+--------+
1 row in set (0.00 sec)

singlestore> exit
Bye

```
## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo my/custom-sdb -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo my/custom-sdb
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps
- [Quickstart SingleStore](/docs/guides/singlestore/quickstart/quickstart.md) with KubeDB Operator.
- Detail concepts of [singlestore object](/docs/guides/singlestore/concepts/singlestore.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
