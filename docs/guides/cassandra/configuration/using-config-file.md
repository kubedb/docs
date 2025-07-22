---
title: Configuring cassandra Using Config File
menu:
  docs_{{ .version }}:
    identifier: cas-configuration-using-config-file
    name: Configure Using Config File
    parent: cas-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Cassandra. This tutorial will show you how to use KubeDB to run a Cassandra with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/cassandra](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/cassandra) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Cassandra allows configuring via configuration file. The default configuration file for Cassandra deployed by `KubeDB` can be found in `/etc/cassandra/cassandra.yaml`. When `spec.configSecret` is set to cassandra, KubeDB operator will get the secret and after that it will validate the values of the secret and then will keep the validated customizable configurations from the user and merge it with the remaining default config. After all that this secret will be mounted to cassandra for use it as the configuration file.

> To learn available configuration option of Cassandra see [Configuration Options](https://cassandra.apache.org/doc/4.0/cassandra/getting_started/configuring.html).

At first, you have to create a secret with your configuration file contents as the value of this key `cassandra.yaml`. Then, you have to specify the name of this secret in `spec.configSecret.name` section while creating cassandra CRO.

## Custom Configuration

At first, create `cassandra.yaml` file containing required configuration settings.

```bash
$ cat cassandra.yaml
read_request_timeout: 6000ms
write_request_timeout: 2500ms
```

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo cas-configuration --from-file=./cassandra.yaml
secret/cas-configuration created
```

Verify the secret has the configuration file.

```bash
$  kubectl get secret -n demo cas-configuration -o yaml
apiVersion: v1
data:
  cassandra.yaml: cmVhZF9yZXF1ZXN0X3RpbWVvdXQ6IDYwMDBtcwp3cml0ZV9yZXF1ZXN0X3RpbWVvdXQ6IDI1MDBtcwo=
kind: Secret
metadata:
  creationTimestamp: "2025-07-15T08:53:26Z"
  name: cas-configuration
  namespace: demo
  resourceVersion: "105786"
  uid: 135c819c-fba6-4800-9ae0-fac35312fab2
type: Opaque

$  echo  cmVhZF9yZXF1ZXN0X3RpbWVvdXQ6IDYwMDBtcwp3cml0ZV9yZXF1ZXN0X3RpbWVvdXQ6IDI1MDBtcwo= | base64 -d
read_request_timeout: 6000ms
write_request_timeout: 2500ms
```

Now, create cassandra crd specifying `spec.configSecret` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cas-custom-config
  namespace: demo
spec:
  version: 5.0.3
  configSecret:
    name: cas-configuration
  topology:
    rack:
      - name: r0
        replicas: 2
        storage:
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
        storageType: Durable
  deletionPolicy: WipeOut

```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/cassandra/configuration/cassandra-config-file.yaml
cassandra.kubedb.com/cas-custom-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `cas-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod  -n demo cas-custom-config-rack-r0-0 
NAME                          READY   STATUS    RESTARTS   AGE
cas-custom-config-rack-r0-0   1/1     Running   0          36s
```

Now, we will check if the cassandra has started with the custom configuration we have provided.

Now, you can exec into the cassandra pod and find if the custom configuration is there,

```bash
$ kubectl exec -it -n demo cas-custom-config-rack-r0-0 -- bash
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)
[cassandra@cas-custom-config-rack-r0-0 /]$ cat /etc/cassandra/cassandra.yaml | grep request_timeout
read_request_timeout: 6000ms
range_request_timeout: 10000ms
write_request_timeout: 2500ms
counter_write_request_timeout: 5000ms
truncate_request_timeout: 60000ms
request_timeout: 10000ms
[cassandra@cas-custom-config-rack-r0-0 /]$ exit
exit
```

As we can see from the configuration of running cassandra, the value of `read_request_timeout` and `write_request_timeout` has been set to our desired value successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo cas/cas-custom-config
kubectl delete -n demo secret cas-configuration
kubectl delete ns demo
```

## Next Steps

- Monitor your cassandra database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/cassandra/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/cassandra/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Cassandra object](/docs/guides/cassandra/concepts/cassandra.md).
- Detail concepts of [CassandraVersion object](/docs/guides/cassandra/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
