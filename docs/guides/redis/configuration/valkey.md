---
title: Run Valkey with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: rd-using-config-file-configuration-valkey
    name: Valkey
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

Valkey allows configuration via a config file. When valkey docker image starts, it executes `valkey-server` command. If we provide a `.conf` file directory as an argument of this command, Valkey server will use configuration specified in the file. To know more about configuring Redis see [here](https://valkey.io/topics/valkey.conf).

At first, you have to create a config file named `valkey.conf` with your desired configuration. Then you have to put this file into a [secret](https://kubernetes.io/docs/concepts/configuration/secret/). You have to specify this secret in `spec.configuration.secretName` section while creating Redis crd. KubeDB will mount this secret into `/usr/local/etc/valkey` directory of the pod and the `valkey.conf` file path will be sent as an argument of `valkey-server` command.

In this tutorial, we will configure `databases` and `maxclients` via a custom config file.

## Custom Configuration

At first, let's create `valkey.conf` file setting `databases` and `maxclients` parameters. Default value of `databases` is 16 and `maxclients` is 10000.

```bash
$ cat <<EOF >valkey.conf
maxclients 425
EOF

$ cat valkey.conf
maxclients 425
```

> Note that config file name must be `valkey.conf`

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
  valkey.conf: bWF4Y2xpZW50cyA0MjUK
kind: Secret
metadata:
  creationTimestamp: "2025-08-01T11:49:18Z"
  name: rd-configuration
  namespace: demo
  resourceVersion: "1077435"
  uid: 402d38aa-e05b-4f2b-97e8-4771a2547872
type: Opaque
```

The configurations are encrypted in the secret.

Now, create Redis crd specifying `spec.configuration` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/custom-config/valkey-custom.yaml
redis.kubedb.com "custom-valkey" created
```

Below is the YAML for the Redis crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: custom-valkey
  namespace: demo
spec:
  version: valkey-8.1.1
  configuration:
    secretName: rd-configuration
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services etc. If everything goes well, we will see that a pod with the name `custom-valkey-0` has been created.


Check if the database is ready

```bash
$ kubectl get redis -n demo
NAME            VERSION        STATUS   AGE
custom-valkey   valkey-8.1.1   Ready    32s
```


Now, we will check if the database has started with the custom configuration we have provided. We will `exec` into the pod and use [CONFIG GET](https://redis.io/commands/config-get) command to check the configuration.

```bash
$ kubectl exec -it -n demo custom-valkey-0 -- bash
custom-valkey-0:/data$ valkey-cli
127.0.0.1:6379> ping
PONG
127.0.0.1:6379> config get maxclients
1) "maxclients"
2) "425"
127.0.0.1:6379> exit
custom-valkey-0:/data$ 
```

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl patch -n demo rd/custom-valkey -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
redis.kubedb.com/custom-valkey patched

$ kubectl delete -n demo redis custom-valkey
redis.kubedb.com "custom-redis" deleted

$ kubectl delete -n demo secret rd-configuration
secret "rd-configuration" deleted

$ kubectl delete ns demo
namespace "demo" deleted
```

## Next Steps

- Learn how to use KubeDB to run a Redis server [here](/docs/guides/redis/README.md).
