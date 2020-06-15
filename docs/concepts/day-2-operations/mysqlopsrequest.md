---
title: MySQLOpsRequest
menu:
  docs_{{ .version }}:
    identifier: mysql-ops-request
    name: MySQLOpsRequest
    parent: day-2-operations
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: day-2-operations
---

> :warning: **This doc is only for KubeDB Enterprise**: You need to be an enterprise user!

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# MySQLOpsRequest

## What is MySQLOpsRequest

`MySQLOpsRequest` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for [MySQL](https://www.mysql.com/) administrative operations like database version upgrading, resources scaling in a Kubernetes native way. You have to configure and create a `MySQLOpsRequest` object for specific operations.

## MySQLOpsRequest CRD Specifications

Like any official Kubernetes resource, a `MySQLOpsRequest` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

A sample common structure of `MySQLOpsRequest`  for `MySQL` operations is shown below,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: myops
  namespace: demo
spec:
  databaseRef:
    name: my-group
  type: Upgrade
  upgrade:
    targetVersion: 8.0.20
  type: HorizontalScaling  
  horizontalScaling:
    member: 3
  type: VerticalScaling  
  verticalScaling:
    mysql:
      requests:
        memory: "200Mi"
        cpu: "0.1"
      limits:
        memory: "300Mi"
        cpu: "0.2"
status:
  conditions:
  - lastTransitionTime: "2020-06-11T09:59:05Z"
    message: The controller has scaled/upgraded the MySQL successfully
    observedGeneration: 3
    reason: OpsRequestSuccessful
    status: "True"
    type: Successful
  observedGeneration: 3
  phase: Successful
```

Here, we are going to describe the various sections of `MySQLOpsRequest` crd.

### MySQLOpsRequest `Spec`

A `MySQLOpsRequest` object has the following fields in the `spec` section.

#### spec.databaseRef

`spec.databaseRef` is a required field specifying the reference of the [MySQL](/docs/concepts/databases/mysql.md) object where the administrative operations will be applied. This field consists of the following sub-fields:

- **spec.databaseRef.name :**  specifies the name of the [MySQL](/docs/concepts/databases/mysql.md) object.

#### spec.type

`spec.type` specifies what kind of operation will be applied to database. Currently there are three types of operation are allowed in `MySQLOpsRequest` is shown below,

- `Upgrade`, `HorizontalScaling` and `VerticalScaling`.

#### spec.upgrade

`spec.upgrade` ia a required field specifying the information of `MySQL` version upgrading. This field consists of the following sub-fields:

- `spec.upgrade.targetVersion` refers to a `MySQL` version name that is run from the current version to this targeted version.

#### spec.horizontalScaling

`spec.horizontalScaling` is a required field specifying the information of `MySQL` server node scaling. This field consists of the following sub-fields:

- `spec.horizontalScaling.member` indicates the number of server nodes of `MySQL` to be operated on.

#### spec.verticalScaling

`spec.verticalScaling` is a required field specifying the information of `MySQL` resources like `cpu`, `memory` etc that will be scaled. This field consists of the following sub-fields:

- `spec.verticalScaling.mysql` indicates the `MySQL` server resources that will be scaled. It has the below structure:
  
    ```
    requests:
      memory: "200Mi"
      cpu: "0.1"
    limits:
      memory: "300Mi"
      cpu: "0.2"
    ```
  