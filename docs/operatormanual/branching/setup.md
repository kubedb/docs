---
title: Branching Setup and Configuration
menu:
  docs_{{ .version }}:
    identifier: branching-setup
    name: Setup and Configuration
    parent: branching
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
aliases:
  - /docs/{{ .version }}/operatormanual/branching/setup-ocm/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branching Setup and Configuration

Choose the setup for your environment:

- [Single-cluster installation](/docs/operatormanual/branching/setup.md#single-cluster-installation) enables branching between namespaces in one Kubernetes cluster.
- [Cross-cluster installation with OCM](/docs/operatormanual/branching/setup.md#cross-cluster-installation-with-ocm) enables branching between managed clusters in an existing Open Cluster Management environment.

## Single-Cluster Installation

### Before You Begin

- Have `kubectl` and Helm configured for the cluster.
- Install KubeDB following the [Helm installation guide](/docs/setup/install/kubedb/helm.md), with a valid cluster license.
- Use compatible Courier, KubeDB Provisioner, and database images that support branching. The provisioner must recognize branch resources and adopt cloned volumes; the PostgreSQL images must support starting from branched data.
- Configure a Ceph RBD CSI StorageClass, snapshot support, and a compatible VolumeSnapshotClass as described in [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

These instructions use the KubeDB installer charts. When using feature builds, set their image tags explicitly using the configuration keys below. A chart version alone does not establish that every image includes the feature.

### Enable Courier with KubeDB

For a new installation, enable the Courier subchart:

```bash
helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/license.txt \
  --set kubedb-courier.enabled=true \
  --wait --burst-limit=10000
```

For an existing Helm release, first update the Branch CRDs as described below, then enable Courier while preserving existing values:

```bash
helm upgrade kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb \
  --reuse-values \
  --set kubedb-courier.enabled=true \
  --wait --burst-limit=10000
```

Use the release name and namespace of your existing installation.

#### Using the installer checkout

To install the Courier chart from the feature checkout alongside an existing compatible KubeDB installation:

```bash
INSTALLER_DIR=/path/to/installer
helm upgrade --install kubedb-courier "$INSTALLER_DIR/charts/kubedb-courier" \
  --namespace kubedb --create-namespace \
  --wait
```

Use one Courier installation per cluster. If Courier is already managed by the KubeDB subchart, update that release instead of adding another installation.

### Single-cluster configuration

The following keys apply when Courier is installed through the KubeDB chart. Drop the `kubedb-courier.` prefix for a separate Courier release.

| Helm value | Purpose |
|---|---|
| `kubedb-courier.enabled` | Enable the Courier subchart; the default is `false`. |
| `kubedb-courier.image.registry` | Image namespace; the default is `kubedb`. |
| `kubedb-courier.image.repository` | Image repository; the default is `kubedb-courier`. |
| `kubedb-courier.image.tag` | Courier image tag; an empty value uses the chart's app version. |
| `kubedb-courier.registryFQDN` | Registry host; the default is `ghcr.io`. |
| `kubedb-courier.resources` | Courier container resource requests and limits. |
| `kubedb-courier.nodeSelector`, `kubedb-courier.tolerations`, `kubedb-courier.affinity` | Courier pod placement. |
| `kubedb-provisioner.operator.tag` | Provisioner image tag in the KubeDB release. |

For example, pass `--set kubedb-courier.image.tag=<courier-tag>` and `--set kubedb-provisioner.operator.tag=<provisioner-tag>` to the KubeDB Helm command when selecting feature builds. Use the [installation guide](/docs/setup/install/kubedb/helm.md) for database image registry configuration.

### CRDs and upgrades

The Courier chart ships the Branch and BranchWork CRDs and installs them on a fresh install. Helm does not upgrade CRDs from a chart's `crds/` directory during `helm upgrade`. Before upgrading an existing release, apply the matching CRDs from the installer checkout:

```bash
INSTALLER_DIR=/path/to/installer
kubectl apply --server-side \
  -f "$INSTALLER_DIR/charts/kubedb-courier/crds/courier.kubedb.com_branches.yaml" \
  -f "$INSTALLER_DIR/charts/kubedb-courier/crds/courier.kubedb.com_branchworks.yaml"
```

Use CRDs from the same feature/release checkout as the operator. If server-side apply reports a field-ownership conflict, review the existing owner and schema differences before transferring ownership.

### Verify the single-cluster installation

```bash
kubectl get crd branches.courier.kubedb.com branchworks.courier.kubedb.com
kubectl -n kubedb get deployment -l app.kubernetes.io/name=kubedb-courier
kubectl -n kubedb rollout status deployment \
  -l app.kubernetes.io/name=kubedb-courier --timeout=5m
kubectl get storageclass ceph-rbd
kubectl get volumesnapshotclass
```

Confirm that the Courier Deployment is available and that the snapshot class driver matches the StorageClass provisioner. Continue with [PostgreSQL branching](/docs/guides/postgres/branch/same-cluster/index.md).

## Cross-Cluster Installation with OCM

This section installs Courier into an **existing OCM environment** to enable cross-cluster branching. It assumes the hub and managed clusters are already configured.

### OCM prerequisites

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

### Install the Courier addon manager on the hub

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

### Enable the addon on both managed clusters

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

### Cross-cluster configuration

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

### Verify cross-cluster Courier

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

Continue with the [PostgreSQL cross-cluster tutorial](/docs/guides/postgres/branch/cross-cluster/index.md).
