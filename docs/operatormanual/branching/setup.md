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
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branching Setup and Configuration

This page enables Courier for branching within one Kubernetes cluster. For delivery between clusters, use [cross-cluster setup](/docs/operatormanual/branching/setup-ocm.md).

## Before You Begin

- Have `kubectl` and Helm configured for the cluster.
- Install KubeDB following the [Helm installation guide](/docs/setup/install/kubedb/helm.md), with a valid cluster license.
- Use compatible Courier, KubeDB Provisioner, and database images that support branching. The provisioner must recognize branch resources and adopt cloned volumes; the PostgreSQL images must support starting from branched data.
- Configure a Ceph RBD CSI StorageClass, snapshot support, and a compatible VolumeSnapshotClass as described in [snapshot configuration](/docs/operatormanual/branching/snapshots.md).

These instructions use the KubeDB installer charts. When using feature builds, set their image tags explicitly using the configuration keys below. A chart version alone does not establish that every image includes the feature.

## Enable Courier with KubeDB

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

### Using the installer checkout

To install the Courier chart from the feature checkout alongside an existing compatible KubeDB installation:

```bash
INSTALLER_DIR=/path/to/installer
helm upgrade --install kubedb-courier "$INSTALLER_DIR/charts/kubedb-courier" \
  --namespace kubedb --create-namespace \
  --wait
```

Use one Courier installation per cluster. If Courier is already managed by the KubeDB subchart, update that release instead of adding another installation.

## Configuration

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

## CRDs and upgrades

The Courier chart ships the Branch and BranchWork CRDs and installs them on a fresh install. Helm does not upgrade CRDs from a chart's `crds/` directory during `helm upgrade`. Before upgrading an existing release, apply the matching CRDs from the installer checkout:

```bash
INSTALLER_DIR=/path/to/installer
kubectl apply --server-side \
  -f "$INSTALLER_DIR/charts/kubedb-courier/crds/courier.kubedb.com_branches.yaml" \
  -f "$INSTALLER_DIR/charts/kubedb-courier/crds/courier.kubedb.com_branchworks.yaml"
```

Use CRDs from the same feature/release checkout as the operator. If server-side apply reports a field-ownership conflict, review the existing owner and schema differences before transferring ownership.

## Verify the installation

```bash
kubectl get crd branches.courier.kubedb.com branchworks.courier.kubedb.com
kubectl -n kubedb get deployment -l app.kubernetes.io/name=kubedb-courier
kubectl -n kubedb rollout status deployment \
  -l app.kubernetes.io/name=kubedb-courier --timeout=5m
kubectl get storageclass ceph-rbd
kubectl get volumesnapshotclass
```

Confirm that the Courier Deployment is available and that the snapshot class driver matches the StorageClass provisioner. Continue with [PostgreSQL branching](/docs/guides/postgres/branch/same-cluster/index.md).
