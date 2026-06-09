---
title: MongoDB Recommendation Overview
menu:
  docs_{{ .version }}:
    identifier: mg-recommendation-overview
    name: Recommendation Overview
    parent: mg-recommendation-mongodb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Recommendation

## Overview

A `Recommendation` is a Kubernetes-native CRD created by the **KubeDB Ops-Manager** and reconciled by the **KubeDB Supervisor**. For an MongoDB cluster managed by KubeDB, the Ops-Manager watches the database's state and emits a Recommendation whenever it detects an action you should take a newer version, an expiring TLS certificate, or an authentication secret nearing its rotation deadline.

Nothing runs until the Recommendation is approved either by you (`status.approvalStatus: Approved`) or automatically through an `ApprovalPolicy` bound to a `MaintenanceWindow`. Once approved, the Supervisor creates the corresponding `MongoDBOpsRequest` and tracks it to completion.

This page is the **MongoDB-specific** intro: which recommendations apply to MongoDB, which spec fields trigger them, and how to tune the Ops-Manager flags that control generation timing. For the architecture diagram, the full Recommendation lifecycle, and end-to-end walkthroughs, see the [operator manual Recommendation Overview](/docs/operatormanual/recommendation/README.md).

<p align="center">
  <img alt="Recommendation Lifecycle" src="/docs/operatormanual/recommendation/images/recommendation-generation.png">
</p>

---

## Prerequisites

Before proceeding, ensure that:

* You have a running Kubernetes cluster with `kubectl` configured (e.g. via [kind](https://kind.sigs.k8s.io/docs/user/quick-start/)).
* KubeDB is installed following the [setup guide](/docs/setup/install/kubedb.md), with the Supervisor enabled:

  ```bash
  --set supervisor.enabled=true
  ```

* A demo namespace exists for examples:

  ```bash
  $ kubectl create namespace demo
  $ kubectl get namespace
  ```

* You are familiar with the relevant KubeDB concepts:
  - [MongoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [MongoDBRotateAuth](/docs/guides/mongodb/rotate-auth/rotateauth.md)
  - [MongoDBReconfigureTLS](/docs/guides/mongodb/tls/overview.md)
  - [MongoDBUpdateVersion](/docs/guides/mongodb/update-version/overview.md)

### Install the Supervisor CRDs first

Enabling the Supervisor with `--set supervisor.enabled=true` requires the four Supervisor CRDs to be present in the cluster **before** the Helm install/upgrade runs. Apply them up front — every Helm command later on this page assumes these CRDs exist.

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

## Find Available StorageClass

You will need to provide a `StorageClass` in the MongoDB spec. List what's available:

```bash
$ kubectl get storageclass
```

---

## Recommendation types for MongoDB

| Type                              | Triggered when                                                                    | Walkthrough                                                                                                  |
| --------------------------------- | --------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| **Version Update**                | A newer major, minor, or patch version becomes available                          | [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)        |
| **Same-Version Update**           | The container image for your *current* version is refreshed (e.g. security patch) | [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)        |
| **TLS Certificate Rotation**      | An issued certificate is approaching its expiry threshold                         | [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md)  |
| **Authentication Secret Rotation** | The auth secret is approaching its `rotateAfter` deadline                         | [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md) |

---

## Triggers specific to MongoDB

This section shows the minimal MongoDB CR fields that cause each recommendation to be generated. For deeper, end-to-end walkthroughs use the links in the table above.

### Authentication Secret Rotation

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-recommendation
  namespace: demo
spec:
  version: "8.0.10"
  authSecret:
    kind: Secret
    name: mg-auth
    rotateAfter: 1h
  storage:
    resources:
      requests:
        storage: 500Mi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

 In this configuration:

 * The `rotateAfter` field defines how long the authentication secret remains valid

 KubeDB monitors the configured lifecycle and generates a RotateAuth Recommendation based on the following conditions:

 * If the secret lifespan is greater than one month, a recommendation is generated when less than one month of validity remains

 * If the secret lifespan is less than one month, a recommendation is generated when approximately one-third of its validity remains

This behavior is configurable, and users can customize the recommendation timing using the RotateAuth flags mentioned in the corresponding section.
Once approved, KubeDB creates an opsrequest to rotate the credentials automatically, ensuring:

 * No expired credentials

 * Improved security posture

 * Reduced manual intervention

### TLS Certificate Rotation

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-tls
  namespace: demo
spec:
  version: "8.0.10"
  tls:
    issuerRef:
      apiGroup: "cert-manager.io"
      kind: Issuer
      name: mongo-ca-issuer
    certificates:
      - alias: client
        duration: 1h20m
      - alias: server
        duration: 2h10m
```
In this configuration:

* The `spec.tls.certificates.duration` field defines how long each certificate remains valid

KubeDB monitors the configured lifecycle and generates a RotateTLS Recommendation based on the following conditions:

* If the certificate duration is greater than one month, a recommendation is generated when less than one month of validity remains

* If the certificate duration is less than one month, a recommendation is generated when approximately one-third of its validity remains

This behavior is configurable, and users can customize the recommendation timing using the RotateTLS flags mentioned in the corresponding section.
Once approved, KubeDB creates an opsrequest to reconfigure TLS automatically, ensuring:

* Continuous secure communication

* No unexpected certificate expiry

* Seamless certificate renewal


### Version Update

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-recommendation
  namespace: demo
spec:
  version: "8.0.10"
  storage:
    resources:
      requests:
        storage: 500Mi
    storageClassName: local-path
  deletionPolicy: WipeOut
```
In this configuration:

* KubeDB monitors the running version of the database

KubeDB monitors the configured lifecycle and generates a VersionUpdate Recommendation based on the following conditions:

* If a newer container image is available for the current version, a recommendation is generated

* If a patch version is released, a recommendation is generated

* If a newer minor or major version becomes available, a recommendation is generated

* If changes are introduced in the existing version image (e.g., security fixes or image updates without a version bump), a recommendation is generated

For example: Recommending version update from `8.0.10` to `8.0.12`

Once approved, KubeDB creates an opsrequest to perform the version upgrade automatically, ensuring:

* Timely adoption of security patches and fixes

* Access to new features and improvements

* Consistent performance and stability across deployments

This significantly reduces operational overhead while improving the reliability, security, and maintainability of your MongoDB clusters.

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

---

For the complete cross-database Recommendation lifecycle, scheduling model, and field reference, see the [Recommendation Overview](/docs/operatormanual/recommendation/overview.md) and [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md) in the operator manual.
