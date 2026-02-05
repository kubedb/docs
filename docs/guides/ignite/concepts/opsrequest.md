---
title: IgniteOpsRequests CRD
menu:
  docs_{{ .version }}:
    identifier: ig-opsrequest-concepts
    name: IgniteOpsRequest
    parent: ig-concepts-ignite
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# IgniteOpsRequest

## What is IgniteOpsRequest

`IgniteOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for [Ignite](https://ignite.apache.org/) administrative operations like database version updating, horizontal scaling, vertical scaling etc. in a Kubernetes native way.

## IgniteOpsRequest CRD Specifications

Like any official Kubernetes resource, a `IgniteOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, some sample `IgniteOpsRequest` CRs for different administrative operations is given below:

Sample `IgniteOpsRequest` for Horizontal Scaling of Database Cluster:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ignite-horizontal-scale-up
  namespace: demo
spec:
  type: HorizontalScaling
  databaseRef:
    name: ignite
  horizontalScaling:
    node: 3
```

Sample `IgniteOpsRequest` for Vertical Scaling of Database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: ig
  verticalScaling:
    node:
      resources:
        requests:
          memory: "2Gi"
          cpu: "1"
        limits:
          memory: "2Gi"
          cpu: "1"
  timeout: 5m
  apply: IfReady
```

Sample `IgniteOpsRequest` Objects for Reconfiguring Ignite database with config:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: reconfigure-ig-cluster
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: ig-cluster
  configuration:
    configSecret:
      name: new-custom-config
    restart: "true"
  timeout: 5m
  apply: IfReady
```

Sample `IgniteOpsRequest` Objects for Volume Expansion of Ignite:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-volume-exp-standalone
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef:
    name: ig-standalone
  volumeExpansion:
    node: 2Gi
    mode: Online
```

Sample `IgniteOpsRequest` Objects for Reconfiguring TLS of the database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-add-tls
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ig
  tls:
    issuerRef:
      name: ig-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    certificates:
      - alias: client
        subject:
          organizations:
            - ignite
          organizationalUnits:
            - client
  timeout: 5m
  apply: IfReady
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: igops-rotate
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ig
  tls:
    rotateCertificates: true
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ig-change-issuer
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ig
  tls:
    issuerRef:
      name: ig-new-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
```

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: IgniteOpsRequest
metadata:
  name: ig-ops-remove
  namespace: demo
spec:
  type: ReconfigureTLS
  databaseRef:
    name: ig
  tls:
    remove: true
```

Here, we are going to describe the various sections of a `IgniteOpsRequest` crd.

A `IgniteOpsRequest` object has the following fields in the `spec` section.

### spec.databaseRef

`spec.databaseRef` is a required field that point to the [Ignite](/docs/guides/ignite/concepts/ignite.md) object for which the administrative operations will be performed. This field consists of the following sub-field:

- **spec.databaseRef.name :** specifies the name of the [Ignite](/docs/guides/ignite/concepts/ignite.md) object.

### spec.type

`spec.type` specifies the kind of operation that will be applied to the database. Currently, the following types of operations are allowed in `IgniteOpsRequest`.

- `UpdateVersion`
- `HorizontalScaling`
- `VerticalScaling`
- `VolumeExpansion`
- `Reconfigure`
- `ReconfigureTLS`
- `Restart`