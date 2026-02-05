---
title: Run Memcached with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: mc-using-config-file-configuration
    name: Customize Configurations
    parent: custom-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Memcached. This tutorial will show you how to use KubeDB to run Memcached with custom configuration.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/memcached](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/memcached) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Memcached does not allows to configuration via any file. However, configuration parameters can be set as arguments while starting the memcached docker image. To keep similarity with other KubeDB supported databases which support configuration through a config file, KubeDB has added an additional executable script on top of the official memcached docker image. This script parses the configuration file then set them as arguments of memcached binary.

To know more about configuring Memcached server see [here](https://github.com/memcached/memcached/wiki/ConfiguringServer).

At first, you have to create a secret with custom configuration file and provide its name in `spec.configuration.secretName`. The operator reads this Secret internally and applies the configuration automatically.

In this tutorial, we will configure [max_connections](https://github.com/memcached/memcached/blob/ee171109b3afe1f30ff053166d205768ce635342/doc/protocol.txt#L672) and [limit_maxbytes](https://github.com/memcached/memcached/blob/ee171109b3afe1f30ff053166d205768ce635342/doc/protocol.txt#L720) via secret.

Create a secret with custom configuration file:
```yaml
apiVersion: v1
stringData:
  memcached.conf: |
    --conn-limit=500
    --memory-limit=128
kind: Secret
metadata:
  name: mc-configuration
  namespace: demo
  resourceVersion: "4505"
```
Here, --con-limit means max simultaneous connections which is default value is 1024.
and --memory-limit means item memory in megabytes which default value is 64.

```bash
 $ kubectl apply -f mc-configuration.yaml
secret/mc-configuration created
```

Let's get the mc-configuration `secret` with custom configuration:

```yaml
$ kubectl get secret -n demo mc-configuration -o yaml
apiVersion: v1
data:
  memcached.conf: LS1jb25uLWxpbWl0PTUwMAotLW1lbW9yeS1saW1pdD01MTIK
kind: Secret
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Secret","metadata":{"annotations":{},"name":"mc-configuration","namespace":"demo","resourceVersion":"4505"},"stringData":{"memcached.conf":"--conn-limit=500\n--memory-limit=512\n"}}
  creationTimestamp: "2024-08-26T12:19:54Z"
  name: mc-configuration
  namespace: demo
  resourceVersion: "4580860"
  uid: 02d41fc0-590e-44d1-ae95-2ee8f9632d36
type: Opaque
```

Now, create Memcached crd specifying `spec.configuration.secretName` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/configuration/mc-custom.yaml
memcached.kubedb.com/custom-memcached created
```

Below is the YAML for the Memcached crd we just created.

```yaml
apiVersion: kubedb.com/v1
kind: Memcached
metadata:
  name: custom-memcached
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  configuration:
    secretName: mc-configuration
  podTemplate:
    spec:
      containers:
        - name: memcached
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 250m
              memory: 64Mi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services etc. If everything goes well, we will see that a pod with the name `custom-memcached-0` has been created.

Check if the database is ready

```bash
$ kubectl get mc -n demo
NAME               VERSION   STATUS   AGE
custom-memcached   1.6.22    Ready    17m
```

Now, we will check if the database has started with the custom configuration we have provided. We will use [stats](https://github.com/memcached/memcached/wiki/ConfiguringServer#inspecting-running-configuration) command to check the configuration.

We will connect to `custom-memcached-0` pod from local-machine using port-frowarding.

```bash
$ kubectl port-forward -n demo custom-memcached-0 11211
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
STAT limit_maxbytes 134217728
...
END
```

Here, `limit_maxbytes` is represented in bytes.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo mc/custom-memcached -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mc/custom-memcached

kubectl patch -n demo drmn/custom-memcached -p '{"spec":{"wipeOut":true}}' --type="merge"
kubectl delete -n demo drmn/custom-memcached

kubectl delete -n demo secret mc-configuration

kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn how to use KubeDB to run a Memcached server [here](/docs/guides/memcached/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
