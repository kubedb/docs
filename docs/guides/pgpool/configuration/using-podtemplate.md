---
title: Run Pgpool with Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: using-podtemplate-configuration
    name: Customize PodTemplate
    parent: pp-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run Pgpool with Custom PodTemplate

KubeDB supports providing custom configuration for Pgpool via [PodTemplate](/docs/guides/pgpool/concepts/pgpool.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a Pgpool database with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [docs/examples/pgpool](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/pgpool) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for database pod through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for Pgpool database.

KubeDB accept following fields to set in `spec.podTemplate:`

- metadata:
  - annotations (pod's annotation)
  - labels (pod's labels)
- controller:
  - annotations (statefulset's annotation)
  - labels (statefulset's labels)
- spec:
  - volumes
  - initContainers
  - containers
  - imagePullSecrets
  - nodeSelector
  - affinity
  - serviceAccountName
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext
  - livenessProbe
  - readinessProbe
  - lifecycle

Read about the fields in details in [PodTemplate concept](/docs/guides/pgpool/concepts/pgpool.md#specpodtemplate),

## Prepare Postgres
For a Pgpool surely we will need a Postgres server so, prepare a KubeDB Postgres cluster using this [tutorial](/docs/guides/postgres/clustering/streaming_replication.md), or you can use any externally managed postgres but in that case you need to create an [appbinding](/docs/guides/pgpool/concepts/appbinding.md) yourself. In this tutorial we will use 3 node Postgres cluster named `ha-postgres`.


## CRD Configuration

Below is the YAML for the Pgpool created in this example. Here, `spec.podTemplate.spec.containers[].env` specifies additional environment variables by users.

In this tutorial, we will register additional two users at starting time of Pgpool. So, the fact is any environment variable with having `suffix: USERNAME` and `suffix: PASSWORD` will be key value pairs of username and password and will be registered in the `pool_passwd` file of Pgpool. So we can use these users after Pgpool initialize without even syncing them.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Pgpool
metadata:
  name: pp-misc-config
  namespace: demo
spec:
  version: "4.4.5"
  replicas: 1
  postgresRef:
    name: ha-postgres
    namespace: demo
  podTemplate:
    spec:
      containers:
        - name: pgpool
          env:
            - name: "ALICE_USERNAME"
              value: alice
            - name: "ALICE_PASSWORD"
              value: '123'
            - name: "BOB_USERNAME"
              value: bob
            - name: "BOB_PASSWORD"
              value: '456'
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/pgpool/configuration/pp-misc-config.yaml
pgpool.kubedb.com/pp-misc-config created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `pp-misc-config-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME               READY   STATUS    RESTARTS   AGE
pp-misc-config-0   1/1     Running   0          68s
```

Now, check if the pgpool has started with the custom configuration we have provided. We will exec in the pod and see the `pool_passwd` file if the user exists of not. We will also see if the environment variable is set or not.

```bash
$ kubectl exec -it -n demo pp-misc-config-0 -- bash
pp-misc-config-0:/$ echo $BOB_USERNAME
bob
pp-misc-config-0:/$ echo $BOB_PASSWORD
456
pp-misc-config-0:/$ echo $ALICE_USERNAME
alice
pp-misc-config-0:/$ echo $ALICE_PASSWORD
123
pp-misc-config-0:/$ cat opt/pgpool-II/etc/pool_passwd 
postgres:AESNz9O12b8N9Ngz1SpCYymv2K8wkHMWS+5TICOsbR5W1U=
bob:AESBw7fOtf4SCfFiI7vbAYpKg==
alice:AESgda2WBFwHQfKluCkXwo+MA==
pp-misc-config-0:/$ exit
exit
```
So, we can see that the additional two users Alice and Bob is successfully registered. Now we can use them. So, first let create the users through the root user postgres.

Now, you can connect to this pgpool through [psql](https://www.postgresql.org/docs/current/app-psql.html). Before that we need to port-forward to the primary service of pgpool.

```bash
$ kubectl port-forward -n demo svc/pp-misc-config 9999
Forwarding from 127.0.0.1:9999 -> 9999
```
Now, let's get the password for the root user.
```bash
$ kubectl get secrets -n demo ha-postgres-auth -o jsonpath='{.data.\password}' | base64 -d
qEeuU6cu5aH!O9CI⏎ 
```
We can use this password now,
```bash
$ psql --host=localhost --port=9999 --username=postgres postgres
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=# CREATE USER alice WITH PASSWORD '123';
CREATE ROLE
postgres=# CREATE USER bob WITH PASSWORD '456';
CREATE ROLE
postgres=# exit
```

Now, let's verify if we can to the database through pgpool with the new users,
```bash
$ export PGPASSWORD='123'
$ psql --host=localhost --port=9999 --username=alice postgres                                    ✘ 2
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=> exit
$ export PGPASSWORD='456'
$ psql --host=localhost --port=9999 --username=bob postgres
psql (16.3 (Ubuntu 16.3-1.pgdg22.04+1), server 16.1)
Type "help" for help.

postgres=> exit
```

You can see we can use these new users to connect to the database.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo pp/pp-misc-config
kubectl delete -n demo pg/ha-postgres
kubectl delete ns demo
```

If you would like to uninstall KubeDB operator, please follow the steps [here](/docs/setup/README.md).

## Next Steps

- [Quickstart Pgpool](/docs/guides/pgpool/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Pgpool database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/pgpool/monitoring/using-prometheus-operator.md).
- Monitor your Pgpool database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/pgpool/monitoring/using-builtin-prometheus.md).
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
