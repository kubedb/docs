---
title: Cadence Quickstart
menu:
  docs_{{ .version }}:
    identifier: dw-cadence-quickstart
    name: Cadence
    parent: dw-quickstart-durableworkflow
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cadence Quickstart

This tutorial shows how to run a Cadence cluster with KubeDB and create a domain in it.

## Before You Begin

- Complete [Prepare Dependencies](/docs/guides/durableworkflow/quickstart/prerequisites.md). You need the `demo` namespace and a `Ready` Postgres named `pg` with its `AppBinding`.

- You should be familiar with the following `KubeDB` concepts:
  - [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md)
  - [DurableWorkflowVersion](/docs/guides/durableworkflow/concepts/catalog/index.md)
  - [AppBinding](/docs/guides/durableworkflow/concepts/appbinding/index.md)

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Cadence Cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: cadence
  namespace: demo
spec:
  version: cadence-1.4.2
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  deletionPolicy: WipeOut
```

It is the Temporal manifest with another `spec.version`. The differences are in what the operator does with it:

- Cadence has no internal frontend. `spec.topology.internalFrontend` is rejected by the webhook.
- The frontend Service `cadence` exposes the gRPC port `7833`. The TChannel ports (`7933` to `7939`) exist only for the traffic between the roles, which also carries membership.
- The schema is applied with `cadence-server update-schema`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/quickstart/cadence.yaml
durableworkflow.kubedb.com/cadence created

$ kubectl get dw -n demo cadence
NAME      VERSION         STATUS   AGE
cadence   cadence-1.4.2   Ready    3m
```

The operator created the databases `cadence` and `cadence_visibility` and the user `dw_cadence` in Postgres, the same way as for Temporal.

## Use the Cluster

The Cadence CLI is in the server image, which the admin tools pod uses:

```bash
$ kubectl patch dw -n demo cadence --type merge -p '{"spec":{"adminTools":{"enabled":true}}}'
durableworkflow.kubedb.com/cadence patched
```

Create a Cadence domain with a [WorkflowNamespace](/docs/guides/durableworkflow/concepts/workflownamespace/index.md) object. Cadence needs an owner and keeps whole days, so a retention of `72h` is 3 days:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: default
  namespace: demo
spec:
  databaseRef:
    name: cadence
  ownerEmail: admin@example.com
  retention: 72h
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/quickstart/cadence-domain.yaml
workflownamespace.kubedb.com/default created
```

See [Manage Namespaces and Domains](/docs/guides/durableworkflow/namespaces/workflownamespace.md). Cadence 1.4.2 cannot delete a domain: deleting the object leaves the domain deprecated.

Alternatively, register the domain by hand with the CLI instead of the object (do not do both for the same name):

```bash
$ kubectl exec -n demo cadence-admin-tools-0 -- cadence --address cadence:7833 --transport grpc --domain default domain register --retention 3
```

Then check the cluster:

```bash
$ kubectl exec -n demo cadence-admin-tools-0 -- cadence --address cadence:7833 --transport grpc admin cluster describe
```

Applications connect to `cadence.demo.svc:7833` with a Cadence client that uses gRPC.

## Cleaning up

```bash
kubectl delete dw -n demo cadence
kubectl delete postgres -n demo pg
kubectl delete ns demo
```

## Troubleshooting

- A pod crash-loops with a database connection error, or Postgres logs `too many clients`: the [connection budget](/docs/guides/durableworkflow/quickstart/prerequisites.md#the-connection-budget) is exceeded.
- The `DurableWorkflow` stays in `Provisioning`: check that the `AppBinding` in `spec.persistence.default` exists and that its database is `Ready`, then look at the schema Job (`kubectl get jobs -n demo`).
- The webhook rejects the object: the message names the field. Common causes are an unknown `spec.version`, `internalFrontend` on Cadence, or a missing `kafkaRef` with an Elasticsearch or OpenSearch visibility store.

## Next Steps

- Understand the [roles and high availability](/docs/guides/durableworkflow/clustering/roles-and-ha.md).
- Secure it with [TLS](/docs/guides/durableworkflow/tls/tls.md). Cadence TLS has a prerequisite.
- Manage [domains](/docs/guides/durableworkflow/namespaces/workflownamespace.md) and use [advanced visibility](/docs/guides/durableworkflow/configuration/advanced-visibility.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
