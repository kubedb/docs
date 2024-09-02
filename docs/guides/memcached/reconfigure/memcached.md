---
title: Reconfigure Memcached Database
menu:
  docs_{{ .version }}:
    identifier: mc-database-reconfigure
    name: Memcached
    parent: mc-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Memcached Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure a Memcached database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Memcached](/docs/guides/memcached/concepts/memcached.md)
  - [MemcachedOpsRequest](/docs/guides/memcached/concepts/memcached-opsrequest.md)
  - [Reconfigure Overview](/docs/guides/memcached/reconfigure/overview.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/memcached](/docs/examples/memcached) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

Now, we are going to deploy a  `Memcached` database using a supported version by `KubeDB` operator. Then we are going to apply `MemcachedOpsRequest` to reconfigure its configuration.

### Prepare Memcached Database

Now, we are going to deploy a `Memcached` database with version `1.6.22`.

### Deploy Memcached 

At first, we will create `secret` named mc-configuration containing required configuration settings.

```yaml
apiVersion: v1
stringData:
  memcached.conf: |
    --conn-limit=500
kind: Secret
metadata:
  name: mc-configuration
  namespace: demo
  resourceVersion: "4505"
```
Here, `maxclients` is set to `500`, whereas the default value is `1024`.

Now, we will apply the secret with custom configuration.
```bash
$ kubectl create -f mc-configuration
secret/mc-configuration created
```

In this section, we are going to create a Memcached object specifying `spec.configSecret` field to apply this custom configuration. Below is the YAML of the `Memcahced` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: memcd-quickstart
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  configSecret:
    name: mc-configuration
  deletionPolicy: WipeOut
```

Let's create the `Memcached` CR we have shown above,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure/sample-redis-config.yaml
memcached.kubedb.com/memcd-quickstart created
```

Now, wait until `memcd-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get mc -n demo
NAME               VERSION     STATUS    AGE
memcd-quickstart   1.6.22      Ready     23s
```

Now, we will check if the database has started with the custom configuration we have provided.

We will connect to `memcd-quickstart-0` pod from local-machine using port-frowarding.

```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
```

Now, connect to the memcached server from a different terminal through `telnet`.

```bash
$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
stats
...
STAT max_connections 500
...
END
```

As we can see from the configuration of running memcached, the value of `maxclients` has been set to `500`.

### Reconfigure using new secret

Now we will reconfigure this database to set `maxclients` to `2000`. 

At first, we will create `secret` named new-configuration containing required configuration settings.

```yaml
apiVersion: v1
stringData:
  memcached.conf: |
    --conn-limit=2000
kind: Secret
metadata:
  name: new-configuration
  namespace: demo
  resourceVersion: "4505"
```
Here, `maxclients` is set to `2000`.

Now, we will apply the secret with custom configuration.
```bash
$ kubectl create -f new-configuration
secret/new-configuration created

#### Create MemcachedOpsRequest

Now, we will use this secret to replace the previous secret using a `MemcachedOpsRequest` CR. The `MemcachedOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-reconfig
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: memcd-quickstart
  configuration:
    configSecret:
      name: new-configuration
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `memcd-quickstart` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.configSecret.name` specifies the name of the new secret.

Let's create the `MemcachedOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/reconfigure/ops-request-reconfigure.yaml
memcachedopsrequest.ops.kubedb.com/memcd-reconfig created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will update the `configSecret` of `Memcached` object.

Let's wait for `MemcachedOpsRequest` to be `Successful`.  Run the following command to watch `MemcahcedOpsRequest` CR,

```bash
$ watch kubectl get memcahcedopsrequest -n demo
Every 2.0s: kubectl get memcachedopsrequest -n demo
NAME                          TYPE          STATUS       AGE
memcd-reconfig                Reconfigure   Successful   1m
```

We can see from the above output that the `MemcachedOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe memcachedopsrequest -n demo memcd-reconfig
Name:         memcd-reconfig
Namespace:    demo
Labels:       <none>
Annotations:  <none>
API Version:  ops.kubedb.com/v1alpha1
Kind:         MemcachedOpsRequest
Metadata:
  Creation Timestamp:  2024-09-02T11:59:59Z
  Generation:          1
  Resource Version:    166566
  UID:                 bb4a1057-ccfa-49c9-8d07-e03cb631a0c9
Spec:
  Apply:  IfReady
  Configuration:
    Config Secret:
      Name:  new-configuration
  Database Ref:
    Name:  memcd-quickstart
  Type:    Reconfigure
Status:
  Conditions:
    Last Transition Time:  2024-09-02T11:59:59Z
    Message:               Memcached ops request is reconfiguring the cluster
    Observed Generation:   1
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2024-09-02T12:00:02Z
    Message:               reconfiguring memcached
    Observed Generation:   1
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2024-09-02T12:00:07Z
    Message:               evict pod; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  EvictPod--memcd-quickstart-0
    Last Transition Time:  2024-09-02T12:00:07Z
    Message:               is pod ready; ConditionStatus:False
    Observed Generation:   1
    Status:                False
    Type:                  IsPodReady
    Last Transition Time:  2024-09-02T12:00:12Z
    Message:               is pod ready; ConditionStatus:True; PodName:memcd-quickstart-0
    Observed Generation:   1
    Status:                True
    Type:                  IsPodReady--memcd-quickstart-0
    Last Transition Time:  2024-09-02T12:00:12Z
    Message:               Restarted pods after reconfiguration
    Observed Generation:   1
    Reason:                RestartPods
    Status:                True
    Type:                  RestartPods
    Last Transition Time:  2024-09-02T12:00:13Z
    Message:               Successfully completed the modification process.
    Observed Generation:   1
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Observed Generation:     1
  Phase:                   Successful
Events:
  Type     Reason                                                          Age   From                         Message
  ----     ------                                                          ----  ----                         -------
  Normal   PauseDatabase                                                   51s   KubeDB Ops-manager Operator  Pausing Memcached demo/memcd-quickstart
  Normal   RestartPods                                                     38s   KubeDB Ops-manager Operator  Restarted pods after reconfiguration
  Normal   ResumeDatabase                                                  38s   KubeDB Ops-manager Operator  Resuming Memcached demo/memcd-quickstart
  Normal   ResumeDatabase                                                  38s   KubeDB Ops-manager Operator  Successfully resumed Memcached demo/memcd-quickstart
  Normal   Successful                                                      38s   KubeDB Ops-manager Operator  Successfully Reconfigured Database

```

Now need to check the new configuration we have provided.

Now, wait until `memcd-quickstart` has status `Ready`. i.e,

```bash
$ kubectl get mc -n demo
NAME               VERSION     STATUS    AGE
memcd-quickstart   1.6.22      Ready     20s
```

Now, we will check if the database has started with the custom configuration we have provided.

We will connect to `memcd-quickstart-0` pod from local-machine using port-frowarding.

```bash
$ kubectl port-forward -n demo memcd-quickstart-0 11211
Forwarding from 127.0.0.1:11211 -> 11211
Forwarding from [::1]:11211 -> 11211
```

Now, connect to the memcached server from a different terminal through `telnet`.

```bash
$ telnet 127.0.0.1 11211
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
stats
...
STAT max_connections 2000
...
END
```

As we can see from the configuration of running memcached, the value of `maxclients` has been updated to `2000`.

As we can see from the configuration of running memcached, the value of `maxclients` has been changed from `500` to `2000`. So the reconfiguration of the database is successful.


### Reconfigure using apply config

Now we will reconfigure this database again to set `maxclients` to `3000`. This time we won't use a new secret. We will use the `applyConfig` field of the `MemcachedOpsRequest`. This will merge the new config in the existing secret.

#### Create RedisOpsRequest

Now, we will use the new configuration in the `data` field in the `MemcahcedOpsRequest` CR. The `MemcachedOpsRequest` yaml is given below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MemcachedOpsRequest
metadata:
  name: memcd-reconfig
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: memcd-quickstart
  configuration:
    applyConfig:
      memcached.conf: |
        --conn-limit=3000
```

Here,

- `spec.databaseRef.name` specifies that we are reconfiguring `sample-redis` database.
- `spec.type` specifies that we are performing `Reconfigure` on our database.
- `spec.configuration.applyConfig` specifies the new configuration that will be merged in the existing secret.

Let's create the `RedisOpsRequest` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/rdops-apply-reconfig.yaml
redisopsrequest.ops.kubedb.com/rdops-apply-reconfig created
```

#### Verify the new configuration is working 

If everything goes well, `KubeDB` Ops-manager operator will merge this new config with the existing configuration.

Let's wait for `RedisOpsRequest` to be `Successful`.  Run the following command to watch `RedisOpsRequest` CR,

```bash
$ watch kubectl get redisopsrequest -n demo
Every 2.0s: kubectl get redisopsrequest -n demo
NAME                               TYPE          STATUS       AGE
rdops-apply-reconfig              Reconfigure   Successful   38s
```

We can see from the above output that the `RedisOpsRequest` has succeeded. If we describe the `RedisOpsRequest` we will get an overview of the steps that were followed to reconfigure the database.

```bash
$ kubectl describe redisopsrequest -n demo rdops-apply-reconfig
Name:         rdops-apply-reconfig
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
    Apply Config:  maxclients 3000
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
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
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

As we can see from the configuration of running redis, the value of `maxclients` has been changed from `2000` to `3000`. So the reconfiguration of the database using the `applyConfig` field is successful.


## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete rd -n demo sample-redis
kubectl delete redisopsrequest -n demo rdops-reconfigure rdops-apply-reconfig
```