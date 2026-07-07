---
title: Configuring Recommendation Generation
menu:
  docs_{{ .version }}:
    identifier: recommendation-configuration
    name: Configuration
    parent: recommendation
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Configuring Recommendation Generation

## Prerequisites

Before using KubeDB Recommendations, ensure that:

* You have a running Kubernetes cluster with `kubectl` configured (e.g. via [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)).
* KubeDB is installed following the [setup guide](/docs/setup/install/kubedb/), with the Supervisor enabled:

  ```bash
  --set supervisor.enabled=true
  ```

* A demo namespace exists for examples:

  ```bash
  kubectl create namespace demo
  ```

  ```bash
  kubectl get namespace
  ```

### Install the Supervisor CRDs first

Enabling the Supervisor with `--set supervisor.enabled=true` requires the four Supervisor CRDs to be present in the cluster **before** the Helm install/upgrade runs. Apply them up front — every Helm command on this page assumes these CRDs exist.

> Apply the CRDs first; *then* run any of the `helm upgrade --install` commands shown below.

```bash
# ApprovalPolicy
kubectl apply -f https://raw.githubusercontent.com/kubeops/supervisor/refs/heads/master/crds/supervisor.appscode.com_approvalpolicies.yaml

# Recommendation
kubectl apply -f https://raw.githubusercontent.com/kubeops/supervisor/refs/heads/master/crds/supervisor.appscode.com_recommendations.yaml

# MaintenanceWindow
kubectl apply -f https://raw.githubusercontent.com/kubeops/supervisor/refs/heads/master/crds/supervisor.appscode.com_maintenancewindows.yaml

# ClusterMaintenanceWindow
kubectl apply -f https://raw.githubusercontent.com/kubeops/supervisor/refs/heads/master/crds/supervisor.appscode.com_clustermaintenancewindows.yaml
```

If you are upgrading an existing install, re-apply these manifests to pull in any new fields — out-of-date CRDs are the most common cause of "unknown field" errors when applying a Recommendation or MaintenanceWindow.

---

## Configuring Recommendation Generation

KubeDB lets you tune when recommendations are generated through Ops-Manager flags set at install/upgrade time. Each subsection below shows the **Helm install** form first and the **raw flag reference** below it.

> All flags live under the Helm path `kubedb-ops-manager.operator.args.<flag>` when set via the KubeDB Helm chart.

### Global

Controls how often the operator re-evaluates managed databases for new recommendations.

```bash
helm upgrade --install kubedb appscode/kubedb \
  --namespace kubedb \
  --set supervisor.enabled=true \
  --set kubedb-ops-manager.operator.args.recommendation-resync-period=1h0m0s
```

| Raw flag                          | Default   | Description                                                                                |
| --------------------------------- | --------- | ------------------------------------------------------------------------------------------ |
| `--recommendation-resync-period`  | `1h0m0s`  | How often the Ops-Manager re-scans every managed database to look for new recommendations. |

### TLS Certificate Rotation (RotateTLS)

The three `before-expiry-*` flags are **added together** to define a single lead time: "create a RotateTLS Recommendation this long before the certificate expires." By default that lead is **1 month** (`year=0`, `month=1`, `day=0`) and is fully configurable.

> **Short-duration fallback.** If the certificate's total lifespan is **less than the configured lead time** (e.g. a `1h20m` cert under the default 1-month lead), the configured lead would never fire. In that case the operator falls back to a proportional rule: the recommendation is generated after **2/3 of the lifespan has elapsed** (i.e. when ~1/3 remains). This keeps short-lived demo certificates rotating predictably without any flag changes.

```bash
helm upgrade --install kubedb appscode/kubedb \
  --namespace kubedb \
  --set supervisor.enabled=true \
  --set kubedb-ops-manager.operator.args.gen-rotate-tls-recommendation-before-expiry-year=0 \
  --set kubedb-ops-manager.operator.args.gen-rotate-tls-recommendation-before-expiry-month=1 \
  --set kubedb-ops-manager.operator.args.gen-rotate-tls-recommendation-before-expiry-day=0
```

| Raw flag                                                  | Default | Description                                                |
| --------------------------------------------------------- | ------- | ---------------------------------------------------------- |
| `--gen-rotate-tls-recommendation-before-expiry-year`      | `0`     | Years before certificate expiry to emit the recommendation. |
| `--gen-rotate-tls-recommendation-before-expiry-month`     | `1`     | Months before certificate expiry to emit the recommendation. |
| `--gen-rotate-tls-recommendation-before-expiry-day`       | `0`     | Days before certificate expiry to emit the recommendation. |

### Authentication Secret Rotation (RotateAuth)

The three `before-expiry-*` flags are added together to define a single lead time before the auth secret's rotation deadline. Defaults match the RotateTLS shape — **1 month** — and are fully configurable.

> **Short-duration fallback.** Same rule as RotateTLS: if `spec.authSecret.rotateAfter` is **shorter than the configured lead time** (e.g. `rotateAfter: 1h` under the default 1-month lead), the operator falls back to **2/3 of the lifespan** — the recommendation lands when ~1/3 of the secret's lifetime remains. With `rotateAfter: 1h` that means a recommendation about 40 minutes after the secret was created.

```bash
helm upgrade --install kubedb appscode/kubedb \
  --namespace kubedb \
  --set supervisor.enabled=true \
  --set kubedb-ops-manager.operator.args.gen-rotate-auth-recommendation-before-expiry-year=0 \
  --set kubedb-ops-manager.operator.args.gen-rotate-auth-recommendation-before-expiry-month=1 \
  --set kubedb-ops-manager.operator.args.gen-rotate-auth-recommendation-before-expiry-day=0
```

| Raw flag                                                   | Default | Description                                                  |
| ---------------------------------------------------------- | ------- | ------------------------------------------------------------ |
| `--gen-rotate-auth-recommendation-before-expiry-year`      | `0`     | Years before auth secret deadline to emit the recommendation. |
| `--gen-rotate-auth-recommendation-before-expiry-month`     | `1`     | Months before auth secret deadline to emit the recommendation. |
| `--gen-rotate-auth-recommendation-before-expiry-day`       | `0`     | Days before auth secret deadline to emit the recommendation. |

### Same-Version Update (UpdateVersion within same version)

Same-version updates (e.g. patched rebuilds of the current image) are governed by two deadline-related flags. Disable deadline enforcement to skip these recommendations entirely, or change the evaluation window to control how far ahead the operator looks.

```bash
helm upgrade --install kubedb appscode/kubedb \
  --namespace kubedb \
  --set supervisor.enabled=true \
  --set kubedb-ops-manager.operator.args.enable-deadline=false \
  --set kubedb-ops-manager.operator.args.max-evaluation-period-before-deadline=168h0m0s
```

| Raw flag                                  | Default     | Description                                                                                                       |
| ----------------------------------------- | ----------- | ----------------------------------------------------------------------------------------------------------------- |
| `--enable-deadline`                       | `false`     | When `true`, the Supervisor enforces deadlines for same-version-update recommendations.                            |
| `--max-evaluation-period-before-deadline` | `168h0m0s`  | How far ahead of the deadline the operator looks when deciding whether to emit a same-version-update recommendation. |

### Worked example

The following install sets non-default values for several knobs at once:

```bash
helm upgrade --install kubedb appscode/kubedb \
  --namespace kubedb \
  --set supervisor.enabled=true \
  --set kubedb-ops-manager.operator.args.recommendation-resync-period=30m \
  --set kubedb-ops-manager.operator.args.gen-rotate-tls-recommendation-before-expiry-month=2 \
  --set kubedb-ops-manager.operator.args.gen-rotate-auth-recommendation-before-expiry-month=1 \
  --set kubedb-ops-manager.operator.args.gen-rotate-auth-recommendation-before-expiry-day=15 \
  --set kubedb-ops-manager.operator.args.enable-deadline=true
```

What this changes:

* The operator re-scans every **30 minutes** instead of hourly.
* TLS rotation recommendations land **2 months** before expiry (more headroom for change-management approvals).
* Auth secret rotation recommendations land **~1.5 months** before the rotation deadline (`month=1`, `day=15`).
* Same-version update recommendations are **enabled** and use the default 168h evaluation window.
