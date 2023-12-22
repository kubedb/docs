---
title: Run Redis with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: rd-using-config-file-configuration
    name: Config File
    parent: rd-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Redis. This tutorial will show you how to use KubeDB to run Redis with custom configuration.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/redis](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/redis) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Redis allows configuration via a config file. When redis docker image starts, it executes `redis-server` command. If we provide a `.conf` file directory as an argument of this command, Redis server will use configuration specified in the file. To know more about configuring Redis see [here](https://redis.io/topics/config).

At first, you have to create a config file named `redis.conf` with your desired configuration. Then you have to put this file into a [secret](https://kubernetes.io/docs/concepts/configuration/secret/). You have to specify this secret in `spec.configSecret` section while creating Redis crd. KubeDB will mount this secret into `/usr/local/etc/redis` directory of the pod and the `redis.conf` file path will be sent as an argument of `redis-server` command.

In this tutorial, we will configure `databases` and `maxclients` via a custom config file. 

## Custom Configuration

At first, let's create `redis.conf` file setting `databases` and `maxclients` parameters. Default value of `databases` is 16 and `maxclients` is 10000.

```bash
$ cat <<EOF >redis.conf
databases 10
maxclients 425
EOF

$ cat redis.conf
databases 10
maxclients 425
```

> Note that config file name must be `redis.conf`

Now, create a Secret with this configuration file. 

```bash
$ kubectl create secret generic -n demo rd-configuration --from-file=./redis.conf
secret/rd-configuration created
```

Verify the Secret has the configuration file.

```bash
$ kubectl get secret -n demo rd-configuration -o yaml

apiVersion: v1
data:
  redis.conf: ZGF0YWJhc2VzIDEwCm1heGNsaWVudHMgNDI1Cgo=
kind: Secret
metadata:
  creationTimestamp: "2023-02-06T08:55:14Z"
  name: rd-configuration
  namespace: demo
  resourceVersion: "676133"
  uid: 73c4e8b5-9e9c-45e6-8b83-b6bc6f090663
type: Opaque
```

The configurations are encrypted in the secret.

Now, create Redis crd specifying `spec.configSecret` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/custom-config/redis-custom.yaml
redis.kubedb.com "custom-redis" created
```

Below is the YAML for the Redis crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: custom-redis
  namespace: demo
spec:
  version: 6.2.14
  configSecret:
    name: rd-configuration
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Now, wait a few minutes. KubeDB operator will create necessary statefulset, services etc. If everything goes well, we will see that a pod with the name `custom-redis-0` has been created.


Check if the database is ready

```bash
$ kubectl get redis -n demo
NAME           VERSION   STATUS   AGE
custom-redis   6.2.14     Ready    10m
```


Now, we will check if the database has started with the custom configuration we have provided. We will `exec` into the pod and use [CONFIG GET](https://redis.io/commands/config-get) command to check the configuration.

```bash
$ kubectl exec -it -n demo custom-redis-0 -- bash
root@custom-redis-0:/data# redis-cli
127.0.0.1:6379> ping
PONG
127.0.0.1:6379> config get databases
1) "databases"
2) "10"
127.0.0.1:6379> config get maxclients
1) "maxclients"
2) "425"
127.0.0.1:6379> exit
root@custom-redis-0:/data# 
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo rd/custom-redis -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/custom-redis patched

$ kubectl delete -n demo redis custom-redis
redis.kubedb.com "custom-redis" deleted

$ kubectl delete -n demo secret rd-configuration
secret "rd-configuration" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Learn how to use KubeDB to run a Redis server [here](/docs/guides/redis/README.md).
