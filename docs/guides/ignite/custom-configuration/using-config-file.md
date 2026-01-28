---
title: Run Ignite with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: ig-using-config-file-configuration
    name: Customize Configurations
    parent: ig-custom-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Using Custom Configuration File

KubeDB supports providing custom configuration for Ignite. This tutorial will show you how to use KubeDB to run Ignite with custom configuration.

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

> Note: YAML files used in this tutorial are stored in [docs/examples/ignite](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/ignite) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

Ignite does not allow to configuration via any file. However, configuration parameters can be set as arguments while starting the ignite docker image. To keep similarity with other KubeDB supported databases which support configuration through a config file, KubeDB has added an additional executable script on top of the official ignite docker image. This script parses the configuration file then set them as arguments of ignite binary.

To know more about configuring Ignite server see [here](https://ignite.apache.org/docs/ignite3/latest/administrators-guide/config/node-config).

At first, you have to create a custom configuration file and provide its name in `spec.configuration.secretName`. The operator reads this Secret internally and applies the configuration automatically.


In this tutorial, we will enable Ignite's authentication via secret.

Create a secret with custom configuration file:
```yaml
apiVersion: v1
stringData:
  node-configuration.xml: |
    <?xml version="1.0" encoding="UTF-8"?>
    <beans xmlns="http://www.springframework.org/schema/beans"
       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
       xsi:schemaLocation="http://www.springframework.org/schema/beans
                           http://www.springframework.org/schema/beans/spring-beans-3.0.xsd">
    <!-- Ignite Configuration -->
    <bean class="org.apache.ignite.configuration.IgniteConfiguration">
        <property name="authenticationEnabled" value="true"/>
    </bean>
    </beans>

kind: Secret
metadata:
  name: ignite-configuration
  namespace: demo
  resourceVersion: "4505"
```
Here, `authenticationEnabled's` default value is `false`. In this secret, we make the value `true`.

```bash
 $ kubectl apply -f ignite-configuration.yaml
secret/ignite-configuration created
```

Let's get the ignite-configuration `secret` with custom configuration:

```yaml
$ kubectl get secret -n demo ignite-configuration -o yaml
apiVersion: v1
data:
  node-configuration.xml: PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz4KPGJlYW5zIHhtbG5zPSJodHRwOi8vd3d3LnNwcmluZ2ZyYW1ld29yay5vcmcvc2NoZW1hL2JlYW5zIgogICB4bWxuczp4c2k9Imh0dHA6Ly93d3cudzMub3JnLzIwMDEvWE1MU2NoZW1hLWluc3RhbmNlIgogICB4c2k6c2NoZW1hTG9jYXRpb249Imh0dHA6Ly93d3cuc3ByaW5nZnJhbWV3b3JrLm9yZy9zY2hlbWEvYmVhbnMKICAgICAgICAgICAgICAgICAgICAgICBodHRwOi8vd3d3LnNwcmluZ2ZyYW1ld29yay5vcmcvc2NoZW1hL2JlYW5zL3NwcmluZy1iZWFucy0zLjAueHNkIj4KPCEtLSBZb3VyIElnbml0ZSBDb25maWd1cmF0aW9uIC0tPgo8YmVhbiBjbGFzcz0ib3JnLmFwYWNoZS5pZ25pdGUuY29uZmlndXJhdGlvbi5JZ25pdGVDb25maWd1cmF0aW9uIj4KCiAgICA8cHJvcGVydHkgbmFtZT0iYXV0aGVudGljYXRpb25FbmFibGVkIiB2YWx1ZT0idHJ1ZSIvPgoKPC9iZWFuPgo8L2JlYW5zPgo=
kind: Secret
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Secret","metadata":{"annotations":{},"name":"ignite-configuration","namespace":"demo","resourceVersion":"4505"},"stringData":{"node-configuration.xml":"\u003c?xml version=\"1.0\" encoding=\"UTF-8\"?\u003e\n\u003cbeans xmlns=\"http://www.springframework.org/schema/beans\"\n   xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"\n   xsi:schemaLocation=\"http://www.springframework.org/schema/beans\n                       http://www.springframework.org/schema/beans/spring-beans-3.0.xsd\"\u003e\n\u003c!-- Your Ignite Configuration --\u003e\n\u003cbean class=\"org.apache.ignite.configuration.IgniteConfiguration\"\u003e\n\n    \u003cproperty name=\"authenticationEnabled\" value=\"true\"/\u003e\n\n\u003c/bean\u003e\n\u003c/beans\u003e\n"}}
  creationTimestamp: "2025-06-02T09:37:05Z"
  name: ignite-configuration
  namespace: demo
  resourceVersion: "1391127"
  uid: 57f2a44c-d6b1-4571-bb91-fd68b3048306
type: Opaque
```

Now, create Ignite crd specifying `spec.configuration` field.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/ignite/configuration/custom-ignite.yaml
ignite.kubedb.com/custom-ignite created
```

Below is the YAML for the Ignite crd we just created.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Ignite
metadata:
  name: custom-ignite
  namespace: demo
spec:
  replicas: 3
  version: 2.17.0
  configuration:
    secretName: ignite-configuration
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services etc. If everything goes well, we will see that a pod with the name `custom-ignite-0` has been created.

Check if the database is ready

```bash
$ kubectl get ig -n demo
NAME               VERSION   STATUS   AGE
custom-ignite      2.17.0    Ready    17m
```

Now, we will check if the database has started with the custom configuration we have provided.
We will connect to `custom-ignite-0` pod:

```bash
$ kubectl exec -it -n demo ignite-quickstart-0 -c ignite -- bash
[ignite@ignite-quickstart-0 config]$ cat /ignite/config/node-configuration.xml

<?xml version="1.0" encoding="UTF-8"?>
<beans xmlns="http://www.springframework.org/schema/beans"
xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
xsi:schemaLocation="http://www.springframework.org/schema/beans
http://www.springframework.org/schema/beans/spring-beans-3.0.xsd">
    <bean id="igniteCfg" class="org.apache.ignite.configuration.IgniteConfiguration">
...
...
<property name="authenticationEnabled" value="true"></property>
    </bean>
</beans>
```

Here, we can see `authenticationEnabled's` value is `true`.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo ig/custom-ignite -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo ig/custom-ignite

kubectl delete -n demo secret ignite-configuration
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- Learn how to use KubeDB to run a Ignite server [here](/docs/guides/ignite/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
