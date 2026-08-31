---
title: Run PgBouncer with Custom PodTemplate
menu:
  docs_{{ .version }}:
    identifier: pb-configuration-usingpodtemplate
    name: Customize PodTemplate
    parent: pb-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run PgBouncer with Custom PodTemplate

KubeDB supports providing custom configuration for PgBouncer via [PodTemplate](/docs/guides/pgbouncer/concepts/pgbouncer.md#specpodtemplate). This tutorial will show you how to use KubeDB to run a PgBouncer server with custom configuration using PodTemplate.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- You will need a PostgreSQL server for PgBouncer to connect to. You can prepare one by following the [PgBouncer quickstart](/docs/guides/pgbouncer/quickstart/quickstart.md) tutorial. In this tutorial, we will use a Postgres named `quick-postgres` in the `demo` namespace.

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

> Note: YAML files used in this tutorial are stored in [here](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/pgbouncer/configuration/using-pod-template/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Overview

KubeDB allows providing a template for PgBouncer pods through `spec.podTemplate`. KubeDB operator will pass the information provided in `spec.podTemplate` to the PetSet created for the PgBouncer server.

KubeDB accepts the following fields to set in `spec.podTemplate`:

- metadata:
  - annotations (pod's annotation)
- controller:
  - annotations (petset's annotation)
- spec:
  - containers
  - env
  - resources
  - initContainers
  - imagePullSecrets
  - nodeSelector
  - schedulerName
  - tolerations
  - priorityClassName
  - priority
  - securityContext

Read about the fields in detail in [PodTemplate concept](/docs/guides/pgbouncer/concepts/pgbouncer.md#specpodtemplate).

## CRD Configuration

Below is the YAML for the PgBouncer used in this example. Here, `spec.podTemplate.metadata.annotations` adds a custom annotation to the pod and `spec.podTemplate.spec.containers[].resources` requests compute resources for the `pgbouncer` container.

```yaml
apiVersion: kubedb.com/v1
kind: PgBouncer
metadata:
  name: sample-pgbouncer
  namespace: demo
spec:
  version: "1.18.0"
  replicas: 1
  database:
    syncUsers: true
    databaseName: "postgres"
    databaseRef:
      name: "quick-postgres"
      namespace: demo
  connectionPool:
    port: 5432
  podTemplate:
    metadata:
      annotations:
        passMe: ToPbPod
    spec:
      containers:
      - name: pgbouncer
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/pgbouncer/configuration/using-pod-template/examples/pb-misc-config.yaml
pgbouncer.kubedb.com/sample-pgbouncer created
```

Now, wait a few minutes. KubeDB operator will create necessary petset, services, secret etc. If everything goes well, we will see that a pod with the name `sample-pgbouncer-0` has been created.

Check that the petset's pod is running

```bash
$ kubectl get pod -n demo
NAME                 READY   STATUS    RESTARTS   AGE
sample-pgbouncer-0   1/1     Running   0          45s
```

Now, we will check if the pod has started with the custom configuration we have provided.

Check that the annotation has been set on the pod.

```bash
$ kubectl get pod -n demo sample-pgbouncer-0 -o jsonpath='{.metadata.annotations}'
map[passMe:ToPbPod]
```

Check that the resource requests have been set on the `pgbouncer` container.

```bash
$ kubectl get pod -n demo sample-pgbouncer-0 -o jsonpath='{.spec.containers[?(@.name=="pgbouncer")].resources}'
{"requests":{"cpu":"250m","memory":"256Mi"}}
```

We can see that both the annotation and the resource requests we provided through `spec.podTemplate` have been applied to the pod.

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete pgbouncer -n demo sample-pgbouncer
pgbouncer.kubedb.com "sample-pgbouncer" deleted
$ kubectl delete ns demo
namespace "demo" deleted
```
