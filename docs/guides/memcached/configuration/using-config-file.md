---
title: Run Memcached with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: mc-using-config-file-configuration
    name: Config File
    parent: mc-configuration
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

At first, you have to create a config file named `memcached.conf` with your desired configuration. Then you have to put this file into a [volume](https://kubernetes.io/docs/concepts/storage/volumes/). You have to specify this volume in `spec.configSecret` section while creating Memcached crd. KubeDB will mount this volume into `/usr/config` directory of the database pod.

In this tutorial, we will configure [max_connections](https://github.com/memcached/memcached/blob/ee171109b3afe1f30ff053166d205768ce635342/doc/protocol.txt#L672) and [limit_maxbytes](https://github.com/memcached/memcached/blob/ee171109b3afe1f30ff053166d205768ce635342/doc/protocol.txt#L720) via a custom config file. We will use a Secret as volume source.

**Configuration File Format:**
KubeDB support providing `memcached.conf` file in the following formats,

```ini
# maximum simultaneous connection
-c 500
# maximum allowed memory for the database in MB.
-m 128
```

or

```ini
# This is a comment line. It will be ignored.
--conn-limit=500
--memory-limit=128
```

or

```ini
# This is a comment line. It will be ignored.
conn-limit = 500
memory-limit = 128
```

## Custom Configuration

At first, let's create `memcached.conf` file setting `max_connections` and `limit_maxbytes` parameters. Default value of `max_connections` is 1024 and `limit_maxbytes` is 64MB (68157440 bytes).

```ini
$ cat <<EOF >memcached.conf
-c 500
# maximum allowed memory in MB
-m 128
EOF

$ cat memcached.conf
-c 500
# maximum allowed memory in MB
-m 128
```

> Note that config file name must be `memcached.conf`

Now, create a Secret with this configuration file.

```bash
 $ kubectl create secret generic -n demo mc-configuration --from-file=./memcached.conf
secret/mc-configuration created
```

Verify the Secret has the configuration file.

```yaml
$ kubectl get secrets -n demo mc-configuration -o yaml
apiVersion: v1
stringData:
  memcached.conf: |
    -c 500
    # maximum allowed memory in MB
    -m 128
kind: Secret
metadata:
  creationTimestamp: 2018-10-04T05:29:37Z
  name: mc-configuration
  namespace: demo
  resourceVersion: "4505"
  selfLink: /api/v1/namespaces/demo/secrets/mc-configuration
  uid: 7c38b5fd-c796-11e8-bb11-0800272ad446
```

Now, create Memcached crd specifying `spec.configSecret` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/memcached/configuration/mc-custom.yaml
memcached.kubedb.com/custom-memcached created
```

Below is the YAML for the Memcached crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Memcached
metadata:
  name: custom-memcached
  namespace: demo
spec:
  replicas: 1
  version: "1.6.22"
  configSecret:
    name: mc-configuration
  podTemplate:
    spec:
      resources:
        limits:
          cpu: 500m
          memory: 128Mi
        requests:
          cpu: 250m
          memory: 64Mi
```

Now, wait a few minutes. KubeDB operator will create the necessary deployment, services etc. If everything goes well, we will see that a deployment with the name `custom-memcached` has been created.

Check that the pods for the deployment is running:

```bash
$ kubectl get pods -n demo
NAME                                READY     STATUS    RESTARTS   AGE
custom-memcached-747b866f4b-j6clt   1/1       Running   0          5m
```

Now, we will check if the database has started with the custom configuration we have provided. We will use [stats](https://github.com/memcached/memcached/wiki/ConfiguringServer#inspecting-running-configuration) command to check the configuration.

We will connect to `custom-memcached-5b5866f5b8-cbc2d` pod from local-machine using port-frowarding.

```bash
$ kubectl port-forward -n demo custom-memcached-5b5866f5b8-cbc2d  11211
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
kubectl patch -n demo mc/custom-memcached -p '{"spec":{"terminationPolicy":"WipeOut"}}' --type="merge"
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
