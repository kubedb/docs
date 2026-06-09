---
title: OracleVersion CRD
menu:
  docs_{{ .version }}:
    identifier: oracle-catalog-concepts
    name: OracleVersion
    parent: oracle-concepts-oracle
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# OracleVersion

## What is OracleVersion

`OracleVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the Docker images to be used for [Oracle](https://oracle.tech/) database deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, a `OracleVersion` custom resource will be created automatically for every supported Oracle version. You have to specify the name of the `OracleVersion` CRD in `spec.version` field of the [Oracle](/docs/guides/oracle/concepts/oracle.md) CRD. Then, KubeDB will use the Docker images specified in the `OracleVersion` CRD to create your expected database.

Using a separate CRD for specifying respective Docker images allows us to modify images independent of the KubeDB operator. This also allows users to use a custom image for the database.

## OracleVersion Specification

As with all other Kubernetes objects, a `OracleVersion` needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: OracleVersion
metadata:
  name: "21.3.0"
spec:
  coordinator:
    image: "ghcr.io/kubedb/oracle-coordinator:v0.10.0"
  version: "21.3.0"
  db:
    image: "container-registry.oracle.com/database/enterprise:21.3.0.0"
  exporter:
    image: "container-registry.oracle.com/database/observability-exporter:2.2.1"
  deprecated: false
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `OracleVersion` CRD. You have to specify this name in `spec.version` field of the [Oracle](/docs/guides/oracle/concepts/oracle.md) CRD.

The naming convention for `OracleVersion` CRD follows the pattern `{Original Oracle version}`.

### spec.version

`spec.version` is a required field that specifies the original version of the Oracle database that has been used to build the Docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the Docker images specified here are supported by the current KubeDB operator.

The default value of this field is `false`. If `spec.deprecated` is set to `true`, KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a required field that specifies the Docker image which will be used to create the StatefulSet by KubeDB operator to create the expected Oracle database.

```bash
$ kubectl get oracleversions
NAME      VERSION   DB_IMAGE                    DEPRECATED   AGE
21.3.0    21.3.0    container-registry.oracle.com/database/enterprise:21.3.0.0    3d
```

## Next Steps

- Learn about the [Oracle CRD](/docs/guides/oracle/concepts/oracle.md).
- Deploy your first Oracle database with KubeDB by following the guide [here](/docs/guides/oracle/quickstart/quickstart.md).