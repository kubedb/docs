---
title: DurableWorkflow Recommendation Overview
menu:
  docs_{{ .version }}:
    identifier: dw-recommendation-overview
    name: DurableWorkflow Recommendation
    parent: dw-recommendation-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# DurableWorkflow Recommendation

A `Recommendation` is a Kubernetes-native CRD created by the **KubeDB Ops-Manager** and reconciled by the **KubeDB Supervisor**. For a DurableWorkflow, the Ops-Manager watches the cluster and emits a Recommendation whenever there is something you should do. Nothing runs until the Recommendation is approved, by you (`status.approvalStatus: Approved`) or by an `ApprovalPolicy` bound to a `MaintenanceWindow`. After that, the Supervisor creates the embedded `DurableWorkflowOpsRequest` and follows it to completion.

{{< notice type="warning" message="The recommendations for DurableWorkflow are implemented but have not been verified on a cluster yet." >}}

This page is the DurableWorkflow-specific intro. For the prerequisites, the Helm flags that control the timing and the full lifecycle, see:

- [Recommendation Configuration](/docs/operatormanual/recommendation/configuration.md)
- [Recommendation Overview](/docs/operatormanual/recommendation)

## Recommendation types

| Type                               | Triggered when                                                         | Embedded ops request |
|------------------------------------|------------------------------------------------------------------------|----------------------|
| **Version Update**                 | A newer, non-deprecated `DurableWorkflowVersion` is available           | `UpdateVersion`      |
| **TLS Certificate Rotation**       | A certificate of the cluster is approaching its expiry                  | `ReconfigureTLS` (`rotateCertificates: true`) |
| **Authentication Secret Rotation** | The default-store credentials are older than `authSecret.rotateAfter`  | `RotateAuth`         |

Rotation of the credentials is only proposed when the operator manages them (`createDatabase` is not `false`). One request rotates the users of all the stores that the operator manages, so the age of the Secret of the default store stands for them all.

The related guides are [Update Version](/docs/guides/durableworkflow/update-version/overview.md), [Reconfigure TLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md) and [Rotate Authentication](/docs/guides/durableworkflow/rotateauth/overview.md).

## The fields that trigger them

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-reco
  namespace: demo
spec:
  version: temporal-1.32.1
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
      authSecret:
        kind: Secret
        name: temporal-reco-default-auth
        rotateAfter: 15m
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: dw-ca-issuer
    certificates:
    - alias: internode
      duration: 61m
    - alias: frontend
      duration: 61m
    - alias: client
      duration: 61m
  deletionPolicy: WipeOut
```

- `persistence.default.authSecret.rotateAfter` makes the credential due 15 minutes after it became active. Use a realistic value in production, such as `2160h`.
- `tls.certificates[].duration` sets short-lived certificates, so a rotation shows up quickly in a test. Use the default in production.
- A version recommendation appears when a newer `DurableWorkflowVersion` that the current one may be updated to exists. The catalog has no such version for `temporal-1.32.1` and `cadence-1.4.2` today.

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/recommendation/temporal.yaml
$ kubectl get recommendation -n demo
```

## Applying a Recommendation

Each `Recommendation` embeds the exact ops request in `spec.operation`. The Supervisor applies it in a maintenance window, or you can apply it yourself:

```bash
$ kubectl get recommendation -n demo <name> -o jsonpath='{.spec.operation}' | kubectl apply -f -
```

## Cleaning up

```bash
kubectl delete recommendation -n demo --all
kubectl delete dw -n demo temporal-reco
```

## Next Steps

- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
