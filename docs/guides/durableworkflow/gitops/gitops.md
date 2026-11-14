---
title: DurableWorkflow GitOps
description: DurableWorkflow GitOps
menu:
  docs_{{ .version }}:
    identifier: dw-gitops-guide
    name: Guide
    parent: dw-gitops-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# GitOps DurableWorkflow using KubeDB GitOps Operator

This guide shows how to manage a `DurableWorkflow` with the `KubeDB` GitOps operator: you keep the desired state in Git, and the operator turns each change into the right ops request.

{{< notice type="warning" message="The GitOps support for DurableWorkflow is implemented but has not been verified on a cluster. It needs the GitOps controller change kubedb/gitops#97, which is not released yet. The mapping below is taken from that code." >}}

## Before You Begin

- A Kubernetes cluster and `kubectl`. Install KubeDB following the steps [here](/docs/setup/README.md), and pass `--set kubedb-crd-manager.installGitOpsCRDs=true` to enable the GitOps operator, as in the other [GitOps guides](/docs/guides/mongodb/gitops/gitops.md).
- A GitOps tool such as `ArgoCD` or `FluxCD`, synchronizing a Git repository with the cluster. The installation and the creation of the application are the same as in the [MongoDB GitOps guide](/docs/guides/mongodb/gitops/gitops.md).
- A `Ready` Postgres named `pg` in the `demo` namespace, as in [Prepare Dependencies](/docs/guides/durableworkflow/quickstart/prerequisites.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Create a DurableWorkflow with GitOps

Commit this object to the repository that your GitOps tool watches. It is a `DurableWorkflow` of the group `gitops.kubedb.com`, with the same spec as the `kubedb.com` kind:

```yaml
apiVersion: gitops.kubedb.com/v1alpha1
kind: DurableWorkflow
metadata:
  name: temporal-gitops
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: durableworkflow
              resources:
                requests:
                  cpu: 500m
                  memory: 512Mi
                limits:
                  cpu: "1"
                  memory: 1Gi
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
  configuration:
    inline:
      dynamic_config.yaml: |
        limit.maxIDLength:
          - value: 255
            constraints: {}
  deletionPolicy: WipeOut
```

```bash
$ kubectl get durableworkflows.gitops.kubedb.com -n demo
$ kubectl get dw -n demo
```

The GitOps operator validates the object with the webhook of the `DurableWorkflow`, and creates the `kubedb.com` `DurableWorkflow` of the same name, owned by the GitOps object. The KubeDB operator then provisions the cluster as usual.

## Change the cluster in Git

When you change the object in Git, the GitOps operator compares it with the live `DurableWorkflow` and creates the matching `DurableWorkflowOpsRequest`:

| Change in the GitOps object                                  | Ops request created                                                                        |
|--------------------------------------------------------------|--------------------------------------------------------------------------------------------|
| `spec.version`                                               | [UpdateVersion](/docs/guides/durableworkflow/update-version/overview.md) to the new version |
| `spec.topology.<role>.replicas`                              | [HorizontalScaling](/docs/guides/durableworkflow/scaling/horizontal-scaling.md), one request for all roles that changed |
| resources of the `durableworkflow` container in `spec.topology.<role>.podTemplate` | [VerticalScaling](/docs/guides/durableworkflow/scaling/vertical-scaling.md), one request for all roles that changed |
| `spec.configuration` (`secretName`, `inline`)                | [Reconfigure](/docs/guides/durableworkflow/reconfigure/overview.md); removing the configuration sets `removeCustomConfig` |
| `spec.tls`                                                   | [ReconfigureTLS](/docs/guides/durableworkflow/reconfigure_tls/overview.md); removing `tls` removes TLS |

The roles are `frontend`, `history`, `matching`, `worker` and `internalFrontend` (Temporal), and a role needs to be in both objects to be compared.

Fields that need no ops request are applied to the `DurableWorkflow` directly: `serviceTemplates`, `halted`, `healthChecker`, `deletionPolicy`, `archival`, `ui`, `adminTools`, `authorization`, `frontendClientAuth` and `monitor`. A restart is never requested; a change in these fields is applied by the operator the way an edit would be.

Not driven by the spec:

- **Authentication.** [RotateAuth](/docs/guides/durableworkflow/rotateauth/overview.md) is not a spec change, because the secrets live on the persistence stores. Create the ops request yourself.
- **Storage.** There are no volumes, so no VolumeExpansion and no storage migration. A managed database is scaled with the ops requests of its own kind.
- **Persistence and `numHistoryShards`.** They are read when the object is created. Later edits in Git are not propagated. `spec.persistence.managed` is immutable anyway.
- **Autoscaling.** The autoscaler is not wired to the GitOps operator for DurableWorkflow yet.

For example, to scale the `worker` role to 3 pods, change `replicas` in Git:

```yaml
    worker:
      replicas: 3
```

After the GitOps tool synchronizes, a `DurableWorkflowOpsRequest` named `temporal-gitops-horizontalscaling-<suffix>` appears:

```bash
$ kubectl get durableworkflowopsrequest -n demo
```

The last five operations and their status are in `status.gitops.gitOpsInfo` of the GitOps object.

## Cleaning up

Remove the object from Git, or delete it:

```bash
kubectl delete durableworkflows.gitops.kubedb.com -n demo temporal-gitops
```

## Next Steps

- Learn about [DurableWorkflowOpsRequest](/docs/guides/durableworkflow/concepts/durableworkflow-ops-request/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
