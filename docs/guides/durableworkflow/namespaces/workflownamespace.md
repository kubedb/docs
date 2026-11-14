---
title: Manage Namespaces and Domains
menu:
  docs_{{ .version }}:
    identifier: dw-workflownamespace-guide
    name: WorkflowNamespace
    parent: dw-namespaces-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Manage Temporal Namespaces and Cadence Domains

This guide shows how to create, update and delete a Temporal namespace or a Cadence domain with a `WorkflowNamespace` object, so that it is declared in Kubernetes like the rest of the cluster.

{{< notice type="info" message="Verified on a kind cluster: create, a retention update (Temporal 168h to 240h, applied), delete (the Temporal namespace was removed), archival enabled on a namespace, and a Cadence domain created with the retention rounded up to whole days (49h became 3 days)." >}}

## Before You Begin

- A `Ready` Temporal cluster named `temporal` and a Cadence cluster named `cadence` in the `demo` namespace, as in the [Temporal](/docs/guides/durableworkflow/quickstart/temporal.md) and [Cadence](/docs/guides/durableworkflow/quickstart/cadence.md) quickstarts.
- You should be familiar with [WorkflowNamespace](/docs/guides/durableworkflow/concepts/workflownamespace/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a Temporal namespace

```yaml
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: orders
  namespace: demo
spec:
  databaseRef:
    name: temporal
  description: Order processing
  retention: 168h
  deletionPolicy: Delete
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/namespaces/temporal-namespace.yaml
workflownamespace.kubedb.com/orders created

$ kubectl get wfns -n demo orders
NAME     DATABASE   NAME     RETENTION   STATUS   AGE
orders   temporal   orders   168h        Ready    20s
```

The operator waits for `temporal` to be `Ready`, registers the namespace `orders` if the frontend does not have it, and sets the phase to `Ready`. Check it with the admin tools pod (enable it as in the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md#use-the-cluster)):

```bash
$ kubectl exec -n demo temporal-admin-tools-0 -- temporal operator namespace describe --namespace orders
```

## Update the retention

Edit the object. The operator applies the change to the engine:

```bash
$ kubectl patch wfns -n demo orders --type merge -p '{"spec":{"retention":"240h"}}'
workflownamespace.kubedb.com/orders patched

$ kubectl get wfns -n demo orders
NAME     DATABASE   NAME     RETENTION   STATUS   AGE
orders   temporal   orders   240h        Ready    3m
```

`spec.name` and `spec.databaseRef` are immutable. The description and the owner can be changed but not cleared. A change made by hand in the engine is reverted within five minutes.

## Create a Cadence domain

A Cadence domain needs an owner, and keeps its closed workflows for whole days. A retention of `49h` is rounded up to 3 days:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: orders
  namespace: demo
spec:
  databaseRef:
    name: cadence
  description: Order processing
  ownerEmail: orders@example.com
  retention: 49h
  deletionPolicy: Delete
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/namespaces/cadence-domain.yaml
workflownamespace.kubedb.com/orders created
```

`spec.retention` keeps what you wrote; `status.retention` is what the engine holds:

```bash
$ kubectl get wfns -n demo orders -o jsonpath='{.spec.retention} {.status.retention}{"\n"}'
49h 72h0m0s
```

Retention below `24h` is rejected by the webhook for Cadence.

## Enable archival on a namespace

When the `DurableWorkflow` has [archival](/docs/guides/durableworkflow/configuration/archival.md) configured, a namespace archives its closed workflows to those URIs if you set `archival.enabled`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: audit
  namespace: demo
spec:
  databaseRef:
    name: temporal-archival
  retention: 72h
  archival:
    enabled: true
```

The webhook rejects `archival.enabled` for a `DurableWorkflow` without `spec.archival`.

## A cluster with authorization

If the `DurableWorkflow` has [JWT authorization](/docs/guides/durableworkflow/authorization/authorization.md), the operator needs a token to call the frontend. Put an admin JWT in the key `token` of a Secret and refer to it with `tokenSecretRef`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: temporal-admin-token
  namespace: demo
type: Opaque
stringData:
  token: <an admin JWT, for example with the permission default:admin>
---
apiVersion: kubedb.com/v1alpha2
kind: WorkflowNamespace
metadata:
  name: orders
  namespace: demo
spec:
  databaseRef:
    name: temporal-authz
  retention: 168h
  tokenSecretRef:
    name: temporal-admin-token
```

For Temporal the token needs the `admin` permission (for example `default:admin`). For Cadence it needs the admin claim. Without `tokenSecretRef`, a cluster with authorization leaves the object in an error state. A JWT expires, so replace the Secret before the token does; the operator reads it on every sync.

## Protect a namespace

`deletionPolicy: DoNotTerminate` makes the webhook refuse the deletion of the object:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/namespaces/do-not-terminate.yaml
```

## Delete a namespace

With `deletionPolicy: Delete`, deleting the object deletes the namespace in the engine, and **with it all its workflows and histories**.

```bash
$ kubectl delete wfns -n demo orders
workflownamespace.kubedb.com "orders" deleted
```

- **Temporal**: the namespace is removed, and the engine deletes its data in the background.
- **Cadence**: Cadence 1.4.2 has no `DeleteDomain`. The operator deprecates the domain, and the object is released, but the domain stays **DEPRECATED** in the engine. A Warning event `DeprecatedOnly` says so:

```bash
$ kubectl get events -n demo --field-selector reason=DeprecatedOnly
```

There is no policy that leaves the namespace in the engine: the only other value, `DoNotTerminate`, refuses the deletion of the object. If the `DurableWorkflow` is gone or is being deleted, the object is released without any call, because the namespace goes with the cluster.

## The CLI as an alternative

A namespace you create with the engine CLI is not touched by the operator. The CLI still works, and is useful to inspect what the operator did:

```bash
$ kubectl exec -n demo temporal-admin-tools-0 -- temporal operator namespace create --namespace default --retention 72h
$ kubectl exec -n demo cadence-admin-tools-0 -- cadence --address cadence:7833 --transport grpc --domain default domain register --retention 3
```

If you also declare the same name with a `WorkflowNamespace`, the operator adopts it and brings the retention, description, owner and archival in line with the object.

## Cleaning up

```bash
kubectl delete wfns -n demo --all
```

## Next Steps

- [Archival](/docs/guides/durableworkflow/configuration/archival.md) of closed workflows.
- [Authorization](/docs/guides/durableworkflow/authorization/authorization.md) of the API calls.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
