---
title: Cross-Cluster Branching Setup and Configuration
menu:
  docs_{{ .version }}:
    identifier: branching-setup-ocm
    name: Cross-Cluster Setup
    parent: branching
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cross-Cluster Branching Setup and Configuration

This guide installs Courier into an **existing OCM environment** to enable cross-cluster branching. It assumes the hub and managed clusters are already configured.

## Before You Begin

Ensure that:

- The OCM hub has working registration, work delivery, and addon management, including the `AddOnTemplate` API.
- Source and target are registered, accepted, and available managed clusters. Their `ManagedCluster` names are used as Courier cluster identities.
- The hub can deliver `ManifestWork` to both clusters, and addon registration can issue credentials for the managed-cluster agents.
- KubeDB, valid per-cluster licenses, and compatible branching-capable operator/database images are installed on the participating clusters. The hub only needs OCM and the Courier addon manager.
- Both database clusters use **the same network-reachable CoW storage backend**, such as Ceph RBD, with compatible CSI settings. Source data must already reside on that backend. See [snapshot configuration](/docs/operatormanual/branching/snapshots.md).
- `kubectl` contexts named `hub`, `source`, and `target` select the intended clusters. In these examples, the managed-cluster names are also `source` and `target`.

Verify the existing OCM registration:

```bash
kubectl --context hub get managedclusters source target
```

Both clusters should report `JOINED=True` and `AVAILABLE=True` with hub acceptance enabled.

## Install the Courier addon manager on the hub

The installer includes a dedicated chart for the Courier manager and its OCM addon definitions:

```bash
INSTALLER_DIR=/path/to/installer
helm --kube-context hub upgrade --install kubedb-courier-addon-manager \
  "$INSTALLER_DIR/charts/kubedb-courier-addon-manager" \
  --namespace kubedb --create-namespace \
  --wait
```

The chart installs the hub Courier CRDs, manager Deployment, RBAC, `ClusterManagementAddOn`, and `AddOnTemplate`. It also delegates the permissions needed by OCM's addon manager and work agents. Use the same Courier build for the manager and agents.

Prepare Branch and BranchWork CRDs on both managed clusters before enabling the addon:

```bash
for context in source target; do
  kubectl --context "$context" apply --server-side \
    -f "$INSTALLER_DIR/charts/kubedb-courier/crds/courier.kubedb.com_branches.yaml" \
    -f "$INSTALLER_DIR/charts/kubedb-courier/crds/courier.kubedb.com_branchworks.yaml"
done
```

For upgrades, also apply matching CRDs on the hub before updating the manager, using the [CRD upgrade procedure](/docs/operatormanual/branching/setup.md#crds-and-upgrades) with `--context hub`. Use schemas matching the operator build on all three clusters.

The managed-cluster addon supplies the Courier agents. Disable any existing standalone Courier Deployment on those clusters as part of the transition, so each cluster runs a single Courier agent installation. For a Courier subchart, set `kubedb-courier.enabled=false` in the existing KubeDB release; retain its CRDs.

## Enable the addon on both managed clusters

By default, the chart uses manual addon placement. Create one `ManagedClusterAddOn` in each cluster's hub namespace:

```yaml
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: courier
  namespace: source
spec:
  installNamespace: kubedb
---
apiVersion: addon.open-cluster-management.io/v1alpha1
kind: ManagedClusterAddOn
metadata:
  name: courier
  namespace: target
spec:
  installNamespace: kubedb
```

Save this manifest as `courier-addons.yaml`, then apply it to the hub:

```bash
kubectl --context hub apply -f courier-addons.yaml
```

OCM deploys `courier-agent` on each managed cluster. The addon supplies each agent's cluster name and scoped hub credentials. Keep each `installNamespace` consistent with `agent.installNamespace` in the Helm values.

## Configuration

| Helm value | Default | Purpose |
|---|---|---|
| `image.registry` | `ghcr.io/kubedb` | Registry and image namespace for the shared Courier image. |
| `image.repository` | `kubedb-courier` | Shared image repository. |
| `image.tag` | Chart app version | Manager and agent image tag. |
| `manager.replicaCount` | `1` | Hub manager replicas. |
| `manager.leaderElection` | `true` | Leader election for the manager. |
| `manager.maxFeedbackWatch` | `100` | Concurrent watch-based feedback limit; additional feedback uses polling. |
| `agent.installNamespace` | `kubedb` | Agent namespace on managed clusters. |
| `agent.image.registry`, `agent.image.repository`, `agent.image.tag` | Shared image values | Optional agent image overrides. |
| `agent.placement.name` | Empty | Existing OCM Placement used for automatic addon placement. |
| `agent.placement.namespace` | `open-cluster-management` | Namespace of the existing Placement. |
| `ocm.addonManager.namespace` | `open-cluster-management-hub` | Existing OCM addon-manager namespace. |
| `ocm.addonManager.serviceAccountName` | `addon-manager-controller-sa` | Existing addon-manager ServiceAccount. |
| `ocm.workAgent.namespace` | `open-cluster-management-agent` | Existing work-agent namespace on managed clusters. |
| `ocm.workAgent.serviceAccountName` | `klusterlet-work-sa` | Existing work-agent ServiceAccount. |

Set the OCM identity values to match the existing installation so the chart grants permissions to the correct ServiceAccounts. To use an existing Placement instead of manual enablement, set `agent.placement.name` and `agent.placement.namespace` when installing the chart.

## Verify Courier

```bash
kubectl --context hub -n kubedb rollout status deployment/kubedb-courier-addon-manager --timeout=5m
kubectl --context hub get clustermanagementaddon courier
kubectl --context hub get addontemplate
kubectl --context hub -n source get managedclusteraddon courier -o yaml
kubectl --context hub -n target get managedclusteraddon courier -o yaml
kubectl --context source -n kubedb rollout status deployment/courier-agent --timeout=5m
kubectl --context target -n kubedb rollout status deployment/courier-agent --timeout=5m
```

Check that addon registration is applied and that availability conditions are `True`. Confirm the agents start controller workers in their logs:

```bash
kubectl --context source -n kubedb logs deployment/courier-agent --tail=100
kubectl --context target -n kubedb logs deployment/courier-agent --tail=100
```

Check Branch API availability on each cluster:

```bash
for context in hub source target; do
  kubectl --context "$context" get crd \
  branches.courier.kubedb.com branchworks.courier.kubedb.com
done
```

Continue with the [PostgreSQL cross-cluster tutorial](/docs/guides/postgres/branch/cross-cluster.md).
