---
title: Run RabbitMQ with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: rm-using-config-file-configuration
    name: Customize Configurations
    parent: rm-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for RabbitMQ. This tutorial will show you how to use KubeDB to run a RabbitMQ with custom configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial. Run the following command to prepare your cluster for this tutorial:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: The yaml files used in this tutorial are stored in [docs/examples/rabbitmq](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

RabbitMQ allows configuring via configuration file. The default configuration file for RabbitMQ deployed by `KubeDB` can be found in `/config/rabbitmq.conf`. When `spec.configuration` is set to rabbitmq, KubeDB operator will get the secret and after that it will validate the values of the secret and then will keep the validated customizable configurations from the user. After all that this secret will be mounted to rabbitmq for use it as the configuration file.

> To learn available configuration option of Pgpool see [Configuration Options](https://www.rabbitmq.com/docs/configure).

At first, you have to create a secret with your configuration file contents as the value of this key `rabbitmq.conf`. Then, you have to specify the name of this secret in `spec.configuration.secretName` section while creating rabbitmq CRO.

## Custom Configuration

At first, create `rabbitmq.conf` file containing required configuration settings.

```bash
$ cat rabbitmq.conf
vm_memory_high_watermark.absolute = 4GB
heartbeat = 100
collect_statistics = coarse
```

Now, create the secret with this configuration file.

```bash
$ kubectl create secret generic -n demo rm-configuration --from-file=./rabbitmq.conf
secret/rm-configuration created
```

Verify the secret has the configuration file.

```bash
$  kubectl get secret -n demo rm-configuration -o yaml
apiVersion: v1
data:
  rabbitmq.conf: bnVtX2luaXRfY2hpbGRyZW4gPSA2Cm1heF9wb29sID0gNjUKY2hpbGRfbGlmZV90aW1lID0gNDAwCg==
kind: Secret
metadata:
  creationTimestamp: "2024-07-29T12:40:48Z"
  name: rm-configuration
  namespace: demo
  resourceVersion: "32076"
  uid: 80f5324a-9a65-4801-b136-21d2fa001b12
type: Opaque

$ echo bnVtX2luaXRfY2hpbGRyZW4gPSA2Cm1heF9wb29sID0gNjUKY2hpbGRfbGlmZV90aW1lID0gNDAwCg== | base64 -d
vm_memory_high_watermark.absolute = 4GB
heartbeat = 100
collect_statistics = coarse
```

Now, create rabbitmq crd specifying `spec.configuration` field.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: rm-custom-config
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  configuration:
    secretName: rm-configuration
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
    storageClassName: standard
  storageType: Durable
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/configuration/rabbitmq-config-file.yaml
rabbitmq.kubedb.com/rm-custom-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `rm-custom-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo rm-custom-config-0
NAME                 READY   STATUS    RESTARTS   AGE
rm-custom-config-0   1/1     Running   0          35s
```

Now, we will check if the pgpool has started with the custom configuration we have provided.

Now, you can exec into the pgpool pod and find if the custom configuration is there,

```bash
$ kubectl exec -it -n demo rm-custom-config-0 -- bash
rm-custom-config-0:/$ cat /config/rabbitmq.conf
log.console.level= info
stomp.default_user= $(RABBITMQ_DEFAULT_USER)
mqtt.allow_anonymous= false
log.file.level= info
loopback_users= none
log.console= true
cluster_partition_handling= pause_minority
vm_memory_high_watermark.absolute= 4GB
mqtt.default_pass= $(RABBITMQ_DEFAULT_PASS)
cluster_formation.peer_discovery_backend= rabbit_peer_discovery_k8s
listeners.tcp.default= 5672
default_user= $(RABBITMQ_DEFAULT_USER)
cluster_formation.node_cleanup.only_log_warning= true
cluster_formation.k8s.service_name= rm-custom-config-pods
heartbeat= 100
cluster_name= rm-custom-config
collect_statistics= coarse
default_pass= $(RABBITMQ_DEFAULT_PASS)
cluster_formation.k8s.host= kubernetes.default.svc.cluster.local
mqtt.default_user= $(RABBITMQ_DEFAULT_USER)
stomp.default_pass= $(RABBITMQ_DEFAULT_PASS)
queue_master_locator= min-masters
cluster_formation.k8s.address_type= hostname
rm-custom-config-0:/$ exit
exit
```

As we can see from the configuration of running rabbitmq, the value of `collect_statistics`, `heartbeat` and `vm_memory_high_watermark.absolute` has been set to our desired value successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo rm/rm-custom-config
kubectl delete -n demo secret rm-configuration
kubectl delete rm -n demo rm-custom-config
kubectl delete ns demo
```

## Next Steps

- Monitor your rabbitmq database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/rabbitmq/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/rabbitmq/monitoring/using-builtin-prometheus.md).
- Detail concepts of [RabbitMQ object](/docs/guides/rabbitmq/concepts/rabbitmq.md).
- Detail concepts of [RabbitMQVersion object](/docs/guides/rabbitmq/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
