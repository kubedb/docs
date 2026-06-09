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

---

## MongoDB Recommendation

### Overview

`Recommendation` is a Kubernetes native CRD created by the **KubeDB Ops-Manager** controller and managed by **KubeDB Supervisor**. Once a KubeDB-managed MongoDB cluster is `Ready`, the Ops-Manager watches its state and generates `Recommendation` objects when defined conditions are met.

Each recommendation is reviewed and acted on in one of two ways:

- **Manually** — by patching `approvalStatus: Approved` on the recommendation
- **Automatically** — by attaching an `ApprovalPolicy` that references a `MaintenanceWindow`

No operation is triggered until a recommendation is explicitly approved. Once approved, the KubeDB Supervisor creates the corresponding `ElasticsearchOpsRequest` and tracks it to completion.
These recommendations help operators proactively manage their database systems by identifying when to perform tasks such as:

* Version upgrades
* TLS certificate rotation
* Authentication credential rotation

Each recommendation can be reviewed and executed manually or integrated into automated operational workflows, improving overall system reliability, security, and maintainability.

<p align="center">
  <img alt="Recommendation Lifecycle" src="/docs/operatormanual/recommendation/images/recommendation-generation.png">
</p>

---

## Prerequisites

Before proceeding, ensure that the following requirements are met:

* A running Kubernetes cluster

* `kubectl` configured to communicate with the cluster

* A cluster provisioned using tools like [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) (if not already available)

* KubeDB operator installed following the guide [here](/docs/setup/install/_index.md)

* Supervisor component enabled during installation:

```bash
--set supervisor.enabled=true
```

* A dedicated namespace for running examples:

```bash
$ kubectl create namespace demo
$ kubectl get namespace
```
* You should be familiar with the following `KubeDB` concepts:
  - [MonggoDBOpsRequest](/docs/guides/mongodb/concepts/opsrequest.md)
  - [MongoDBRotateAuth](/docs/guides/mongodb/rotate-auth/overview.md)
  - [MongoDBTLS](/docs/guides/mongodb/tls/overview.md)
  - [MongoDBUpdateVersion](/docs/guides/mongodb/update-version/overview.md)
---

## Find Available StorageClass

We will have to provide `StorageClass` in MongoDB CRD specification. Check available `StorageClass` in your cluster using the following command,

```bash
$ kubectl get storageclass
```

---

> This document provides a high-level overview with illustrative examples. To fully understand and apply these recommendations in your database, follow the linked guides and the [Recommendation Overview](/docs/operatormanual/recommendation/overview.md)

---

## Recommendation Types

KubeDB currently supports the following recommendation categories for MongoDB:

1. [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)
2. [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md)
3. [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md)


These recommendations are generated based on cluster configuration, resource lifecycle, and predefined thresholds.

---

## How Recommendations Are Generated

KubeDB’s recommendation engine continuously monitors your MongoDB clusters and evaluates key configuration fields such as version, authentication secret lifecycle, and TLS certificate duration. Based on these inputs, it automatically generates actionable recommendations when predefined thresholds are reached.

In practice, this means you do not need to manually track credential expiry, certificate validity, or version updates. Instead, KubeDB proactively suggests the right operation at the right time.

---

## Authentication Secret Rotation

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-alone
  namespace: demo
spec:
  version: "8.0.10"
  authSecret:
    name: mg-alone-auth
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
* If the secret lifespan is less than one month, a recommendation is generated when approximately one-third of its validity remains.

Once approved, KubeDB creates an opsrequest to rotate the credentials automatically, ensuring:

* No expired credentials
* Improved security posture
* Reduced manual intervention

---

## TLS Certificate Rotation

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-tls
  namespace: demo
spec:
  version: "8.0.10"
  tls:
    certificates:
      - alias: client
        duration: 1h20m
      - alias: server
        duration: 2h10m
```

In this configuration:

* The `spec.tls.certificates.duration` field defines how long each certificate remains valid

KubeDB tracks each certificate individually and generates a **RotateTLS Recommendation** before expiry:

* If duration is short → recommendation at ~1/3 remaining
* If certificate duration is greater than one month, the recommendation is generated when less than one month of validity remains.

For example, a `1h20m` certificate triggers a recommendation after about 50–55 minutes.

Once approved, KubeDB executes a TLS reconfigure opsrequest, ensuring:

* Continuous secure communication
* No unexpected certificate expiry
* Seamless certificate renewal

---

## Version Update Recommendation

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-alone
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

* KubeDB continuously monitors the running version of the database

A **Version Update Recommendation** is generated when:

* A newer container image is available for the current version
* A patch version is released
* A newer minor or major version becomes available
* Changes are introduced in the existing version image (e.g., security fixes or image updates without a version bump)

For example:

> Recommending version update from `xpack-9.1.9` to `xpack-9.2.3`

Once approved, KubeDB automatically creates and executes the corresponding opsrequest to perform the version upgrade with minimal disruption.

This ensures:

* Timely adoption of security patches and fixes
* Access to new features and improvements
* Consistent performance and stability across deployments


This significantly reduces operational overhead while improving the reliability, security, and maintainability of your MongoDB clusters.
