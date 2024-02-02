---
title: Reconfigure Redis Database
menu:
  docs_{{ .version }}:
    identifier: rd-database-reconfigure
    name: Redis
    parent: rd-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Redis Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Redis database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Redis](/docs/guides/redis/concepts/redis.md)
  - [RedisOpsRequest](/docs/guides/redis/concepts/redisopsrequest.md)
  - [Reconfigure Overview](/docs/guides/redis/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/redis](/docs/examples/redis) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Redis` database using a supported version by `KubeDB` operator. Then we are going to apply `RedisOpsRequest` to reconfigure its configuration.

### Prepare Redis Database

Now, we are going to deploy a `Redis` database with version `6.2.14`.

### Deploy Redis 

At first, we will create `redis.conf` file containing required configuration settings.

```ini
$ cat redis.conf
maxclients 500
```
Here, `maxclients` is set to `500`, whereas the default value is `10000`.

Now, we will create a secret with this configuration file.

```bash
$ kubectl create secret generic -n demo rd-custom-config --from-file=./redis.conf
secret/rd-custom-config created
```

In this section, we are going to create a Redis object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `Redis` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: "6.2.14"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  configSecret:
    name: rd-custom-config
```

Let's create the `Redis` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/sample-redis-config.yaml
redis.kubedb.com/sample-redis created
```

Now, wait until `sample-redis` has status `Ready`. i.e,

```bash
$ kubectl get rd -n demo
NAME            VERSION     STATUS    AGE
sample-redis   6.2.14      Ready     23s
```

Now, we will check if the database has started with the custom configuration we have provided.

First we need to get the username and password to connect to a redis instance,
```bash
$ kubectl get secrets -n demo sample-redis-auth -o jsonpath='{.data.\username}' | base64 -d
default

$ kubectl get secrets -n demo sample-redis-auth -o jsonpath='{.data.\password}' | base64 -d
0PI1tYTyzp;YaXOh
```

Now let's connect to a redis instance and run a redis internal command to check the configuration we have provided.

```bash
$ kubectl exec -n demo  sample-redis-0  -- redis-cli config get maxclients
maxclients
500
```

As we can see from the configuration of running redis, the value of `maxclients` has been set to `500`.

### Reconfigure using new secret

Now we will reconfigure this database to set `maxclients` to `2000`. 

Now, we will edit the `redis.conf` file containing required configuration settings.

```ini
$ cat redis.conf
maxclients 2000
```

Then, we will create a new secret with this configuration file.

```bash
$ kubectl create secret generic -n demo new-custom-config --from-file=./redis.conf
secret/new-custom-config created
```

#### Create RedisOpsRequest

Now, we will use this secret to replace the previous secret using a `RedisOpsRequest` CR. The `RedisOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rdops-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-redis
  configuration:
      configSecret:
        name: new-custom-config
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-redis` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/rdops-reconfigure.yaml
redisopsrequest.ops.kubedb.com/rdops-reconfigure created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Redis` object.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                          TYPE          STATUS       AGE
rdops-reconfigure             Reconfigure   Successful   1m
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe redisopsrequest -n demo rdops-reconfigure
Name:         rdops-reconfigure
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2024-02-02T09:33:08Z
  Generation:          1
  Resource Version:    2702
  UID:                 a0ec9260-65cf-4001-b905-e0e4d0530cc9
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-custom-config
  Database Ref:
    Name:  sample-redis
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-02-02T09:33:08Z
    Message:               Redis ops request is reconfiguring the cluster
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-02-02T09:33:11Z
    Message:               reconfiguring new secret
    Observed Generation:   1
    Reason:                patchedSecret
    Status:                True
    Type:                  patchedSecret
    Last Transition Time:  2024-02-02T09:33:11Z
    Message:               reconfiguring redis
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2024-02-02T09:33:21Z
    Message:               Restarted pods after reconfiguration
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-02-02T09:33:21Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason          Age   From                         Message
  ----    ------          ----  ----                         -------
  Normal  PauseDatabase   101s  KubeDB Ops-manager Operator  Pausing Redis demo/sample-redis
  Normal  RestartPods     88s   KubeDB Ops-manager Operator  Restarted pods after reconfiguration
  Normal  ResumeDatabase  88s   KubeDB Ops-manager Operator  Resuming Redis demo/sample-redis
  Normal  ResumeDatabase  88s   KubeDB Ops-manager Operator  Successfully resumed Redis demo/sample-redis
  Normal  Successful      88s   KubeDB Ops-manager Operator  Successfully Reconfigured Database
```

Now let's connect to a redis instance and run a redis internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  sample-redis-0  -- redis-cli config get maxclients
maxclients
2000

```

As we can see from the configuration of running redis, the value of `maxclients` has been changed from `500` to `20000`. So the reconfiguration of the database is successful.


### Reconfigure using inline config

Now we will reconfigure this database again to set `maxclients` to `3000`. This time we won't use a new secret. We will use the `inlineConfig` field of the `RedisOpsRequest`. This will merge the new config in the existing secret.

#### Create RedisOpsRequest

Now, we will use the new configuration in the `data` field in the `RedisOpsRequest` CR. The `RedisOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rdops-inline-reconfig
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: sample-redis
  configuration:
    inlineConfig: maxclients 3000
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-redis` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.inlineConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/rdops-inline-reconfig.yaml
redisopsrequest.ops.kubedb.com/rdops-inline-reconfig created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                               TYPE          STATUS       AGE
rdops-inline-reconfig              Reconfigure   Successful   38s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe redisopsrequest -n demo rdops-inline-reconfig
Name:         rdops-inline-reconfig
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         RedisOpsRequest
Metadata:
  Creation Timestamp:  2024-02-02T09:42:33Z
  Generation:          1
  Resource Version:    3550
  UID:                 fceacc94-df88-42a1-8991-f77056f33a75
Spec:
  Apply:  IfReady
  Configuration:
    Inline Config:  maxclients 3000
  Database Ref:
    Name:  sample-redis
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-02-02T09:42:33Z
    Message:               Redis ops request is reconfiguring the cluster
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-02-02T09:42:36Z
    Message:               reconfiguring new secret
    Observed Generation:   1
    Reason:                patchedSecret
    Status:                True
    Type:                  patchedSecret
    Last Transition Time:  2024-02-02T09:42:36Z
    Message:               reconfiguring redis
    Observed Generation:   1
    Reason:                UpdateStatefulSets
    Status:                True
    Type:                  UpdateStatefulSets
    Last Transition Time:  2024-02-02T09:42:46Z
    Message:               Restarted pods after reconfiguration
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-02-02T09:42:46Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type    Reason          Age   From                         Message
  ----    ------          ----  ----                         -------
  Normal  PauseDatabase   27s   KubeDB Ops-manager Operator  Pausing Redis demo/sample-redis
  Normal  RestartPods     14s   KubeDB Ops-manager Operator  Restarted pods after reconfiguration
  Normal  ResumeDatabase  14s   KubeDB Ops-manager Operator  Resuming Redis demo/sample-redis
  Normal  ResumeDatabase  14s   KubeDB Ops-manager Operator  Successfully resumed Redis demo/sample-redis
  Normal  Successful      14s   KubeDB Ops-manager Operator  Successfully Reconfigured Database
```

Now let's connect to a redis instance and run a redis internal command to check the new configuration we have provided.

```bash
$ kubectl exec -n demo  sample-redis-0  -- redis-cli config get maxclients
maxclients
3000
```

As we can see from the configuration of running redis, the value of `maxclients` has been changed from `2000` to `3000`. So the reconfiguration of the database using the `inlineConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rd -n demo sample-redis
kubectl delete redisopsrequest -n demo rdops-reconfigure rdops-inline-reconfig
```