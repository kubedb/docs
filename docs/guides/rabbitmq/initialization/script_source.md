---
title: Initialize RabbitMQ using Script Source
menu:
  docs_{{ .version }}:
    identifier: rm-script-source-initialization
    name: Using Script
    parent: rm-initialization-rabbitmq
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize RabbitMQ with Script

KubeDB supports RabbitMQ initialization using definitions files. This tutorial will show you how to use KubeDB to initialize a RabbitMQ broker from a definitions script stored in a ConfigMap.

RabbitMQ supports importing a [definitions file](https://www.rabbitmq.com/docs/definitions) (JSON format) at startup to pre-configure virtual hosts, exchanges, queues, bindings, users, and policies.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created

$ kubectl get ns demo
NAME    STATUS  AGE
demo    Active  5s
```

> Note: YAML files used in this tutorial are stored in [docs/examples/rabbitmq](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/rabbitmq) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Prepare Initialization Scripts

RabbitMQ supports initialization with a JSON definitions file. In this tutorial, we will use a `definitions.json` file to pre-create a virtual host `app-vhost`, an exchange `app-exchange`, a queue `app-queue`, and a binding between them.

We will use a ConfigMap as the script source. You can use any Kubernetes supported [volume](https://kubernetes.io/docs/concepts/storage/volumes) as a script source.

Let's create a ConfigMap with the initialization definitions file:

```bash
$ kubectl create configmap -n demo rmq-init-script \
--from-literal=definitions.json="$(curl -fsSL https://raw.githubusercontent.com/kubedb/rabbitmq-init-scripts/master/definitions.json)"
configmap/rmq-init-script created
```

## Create RabbitMQ with Script Source

Following YAML describes the RabbitMQ object with `init.script`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: RabbitMQ
metadata:
  name: script-rabbitmq
  namespace: demo
spec:
  version: "3.13.2"
  replicas: 1
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    script:
      configMap:
        name: rmq-init-script
  deletionPolicy: WipeOut
```

Here,

- `init.script` specifies the definitions file used to initialize the broker when it is being created. RabbitMQ loads the definitions from the ConfigMap volume at startup.

VolumeSource provided in `init.script` will be mounted in the Pod. RabbitMQ will automatically load the `definitions.json` file from the mounted path.

Now, let's create the RabbitMQ CRD using the YAML shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/rabbitmq/initialization/script-rabbitmq.yaml
rabbitmq.kubedb.com/script-rabbitmq created
```

Now, wait until RabbitMQ goes in `Ready` state. Verify that the broker is in `Ready` state using the following command:

```bash
$ kubectl get rabbitmq -n demo script-rabbitmq
NAME              VERSION   STATUS   AGE
script-rabbitmq   3.13.2    Ready    2m
```

## Verify Initialization

Now let's connect to our RabbitMQ instance to verify that the broker has been initialized successfully.

**Connection Information:**

- Host name/address: you can use any of these
  - Service: `script-rabbitmq.demo`
  - Pod IP: (`$ kubectl get pods script-rabbitmq-0 -n demo -o yaml | grep podIP`)
- Port: `5672` (AMQP) or `15672` (Management UI)

- Username: Run the following command to get the *username*:

  ```bash
  $ kubectl get secret -n demo script-rabbitmq-auth -o jsonpath='{.data.username}' | base64 -d
  admin
  ```

- Password: Run the following command to get the *password*:

  ```bash
  $ kubectl get secret -n demo script-rabbitmq-auth -o jsonpath='{.data.password}' | base64 -d
  S3cur3P@ssw0rd
  ```

You can verify the initialization using the RabbitMQ management CLI inside the Pod:

```bash
$ kubectl exec -it -n demo script-rabbitmq-0 -- rabbitmqctl list_vhosts
Listing vhosts ...
name
/
app-vhost
```

```bash
$ kubectl exec -it -n demo script-rabbitmq-0 -- rabbitmqctl list_exchanges --vhost app-vhost name type
Listing exchanges for vhost app-vhost ...
name    type
        direct
amq.direct      direct
amq.fanout      fanout
amq.headers     headers
amq.match       headers
amq.rabbitmq.trace      topic
amq.topic       topic
app-exchange    direct
```

```bash
$ kubectl exec -it -n demo script-rabbitmq-0 -- rabbitmqctl list_queues --vhost app-vhost name
Listing queues for vhost app-vhost ...
name
app-queue
```

We can see that `app-vhost`, `app-exchange`, and `app-queue` were created through the initialization definitions file.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo rabbitmq/script-rabbitmq
$ kubectl delete -n demo configmap/rmq-init-script
$ kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/rabbitmq/backup/overview/index.md) RabbitMQ using Stash.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
