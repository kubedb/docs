---
title: Reconfigure TLS of a DurableWorkflow
menu:
  docs_{{ .version }}:
    identifier: dw-reconfigure-tls-guide
    name: Guide
    parent: dw-reconfigure-tls-durableworkflow
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure TLS of a DurableWorkflow

## Before You Begin

- Have cert-manager installed and the `dw-ca-issuer` Issuer of [TLS](/docs/guides/durableworkflow/tls/tls.md#create-an-issuer) in the `demo` namespace.
- Have a `temporal` cluster, as in the [quickstart](/docs/guides/durableworkflow/quickstart/temporal.md), to add TLS to, and the `temporal-tls` cluster of [TLS](/docs/guides/durableworkflow/tls/tls.md#temporal) for the other requests.
- You should be familiar with [how Reconfigure TLS works](/docs/guides/durableworkflow/reconfigure_tls/overview.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Add TLS

**All pods are replaced at once, and the cluster is unavailable while they start.**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: temporal
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: dw-ca-issuer
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/reconfigure-tls/add-tls.yaml
```

Applications must trust the CA and, for Temporal with the default `frontendClientAuth: Required`, present a certificate, before this is applied: plaintext clients stop working.

## Rotate the certificates

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-rotate-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: temporal-tls
  tls:
    rotateCertificates: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/reconfigure-tls/rotate.yaml
$ kubectl get secret -n demo temporal-tls-frontend-cert -o jsonpath='{.metadata.creationTimestamp}'
```

The creation time of the Secrets changes. No client call fails.

## Update a certificate

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-update-cert
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: temporal-tls
  tls:
    certificates:
    - alias: frontend
      duration: 4320h
      renewBefore: 720h
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/reconfigure-tls/update-certificate.yaml
```

## Remove TLS

**All pods are replaced at once, and the cluster is unavailable while they start.**

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DurableWorkflowOpsRequest
metadata:
  name: dw-remove-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: temporal-tls
  tls:
    remove: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/reconfigure-tls/remove-tls.yaml
```

## Cleaning up

```bash
kubectl delete dwops -n demo dw-add-tls dw-rotate-tls dw-update-cert dw-remove-tls
kubectl delete dw -n demo temporal-tls
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
