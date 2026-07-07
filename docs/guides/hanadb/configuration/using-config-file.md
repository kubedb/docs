---
title: HanaDB Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-configuration-using-config-file
    name: Custom Configuration
    parent: guides-hanadb-configuration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Run HanaDB with Custom Configuration

KubeDB lets you supply a custom SAP HANA `global.ini` to tune the database. This guide shows how to set
custom configuration at creation time using a `Secret`. To change configuration on a running database,
see [Reconfigure](/docs/guides/hanadb/reconfigure/reconfigure.md).

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/configuration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/configuration) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```
namespace/demo created

## Overview

SAP HANA reads its configuration from `global.ini`. KubeDB supports overriding `global.ini` settings.
The custom values are **merged** into HANA's `global.ini` — you only specify the sections and keys you
want to change. In this guide we lower the HANA memory budget by setting
`[memorymanager] global_allocation_limit` (in bytes).

## Create a Configuration Secret

Put the custom `global.ini` in a `Secret`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hanadb-configuration
  namespace: demo
stringData:
  global.ini: |
    [memorymanager]
    global_allocation_limit = 8589934592
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/configuration/hanadb-configuration.yaml
```
secret/hanadb-configuration created

The key inside the secret **must** be `global.ini`. Here `global_allocation_limit = 8589934592` caps the
HANA global allocation at 8 GiB.

## Create a HanaDB Using the Configuration Secret

Reference the secret with `spec.configuration.secretName`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-custom-config
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  configuration:
    secretName: hanadb-configuration
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 64Gi
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/configuration/standalone-cus-conf.yaml
```
hanadb.kubedb.com/hanadb-custom-config created

Wait for the database to become `Ready`:

```bash
kubectl get hanadb.kubedb.com -n demo hanadb-custom-config
```
NAME                   VERSION   STATUS   AGE
hanadb-custom-config   2.0.82    Ready    24m

## Verify the Configuration

Read the password and query `M_INIFILE_CONTENTS` to confirm HANA picked up the custom value:

```bash
HANA_PASSWORD="$(kubectl get secret hanadb-custom-config-auth -n demo -o jsonpath='{.data.password}' | base64 -d)"
```

```bash
kubectl exec -n demo hanadb-custom-config-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$HANA_PASSWORD' \
  \"SELECT LAYER_NAME, VALUE FROM M_INIFILE_CONTENTS WHERE FILE_NAME='global.ini' AND SECTION='memorymanager' AND KEY='global_allocation_limit'\""
```
LAYER_NAME,VALUE
"SYSTEM","8589934592"
"DEFAULT","0"
2 rows selected

The `SYSTEM` layer shows the custom value `8589934592` (8 GiB) merged into `global.ini` from the
configuration secret, overriding the `DEFAULT` layer.

## Cleaning Up

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-custom-config
```

```bash
kubectl delete secret -n demo hanadb-configuration
```

```bash
kubectl delete ns demo
```

## Next Steps

- Change configuration on a running database with [Reconfigure](/docs/guides/hanadb/reconfigure/reconfigure.md).
- Review the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
