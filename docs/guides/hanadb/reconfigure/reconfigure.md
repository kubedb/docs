---
title: Reconfigure HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-reconfigure-reconfigure
    name: Reconfigure
    parent: guides-hanadb-reconfigure
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure HanaDB

This guide shows how to change the custom `global.ini` configuration of a running HanaDB using a
`HanaDBOpsRequest` of type `Reconfigure`.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/reconfigure](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```
namespace/demo created

## Deploy a HanaDB

Deploy a standalone HanaDB to reconfigure:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: HanaDB
metadata:
  name: hanadb-standalone
  namespace: demo
spec:
  version: "2.0.82"
  replicas: 1
  storageType: Durable
  podTemplate:
    spec:
      containers:
      - name: hanadb
        resources:
          requests:
            cpu: "1500m"
            memory: "8Gi"
          limits:
            cpu: "4"
            memory: "14Gi"
  storage:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 64Gi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure/standalone-ops.yaml
```
hanadb.kubedb.com/hanadb-standalone created

Wait until `hanadb-standalone` is `Ready`.

## Check Current Configuration

Read the current value of `[memorymanager] global_allocation_limit`:

```bash
HANA_PASSWORD="$(kubectl get secret hanadb-standalone-auth -n demo -o jsonpath='{.data.password}' | base64 -d)"
```

```bash
kubectl exec -n demo hanadb-standalone-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$HANA_PASSWORD' \
  \"SELECT LAYER_NAME, VALUE FROM M_INIFILE_CONTENTS WHERE FILE_NAME='global.ini' AND SECTION='memorymanager' AND KEY='global_allocation_limit'\""
```
LAYER_NAME,VALUE
"DEFAULT","0"
1 row selected

The database starts with no custom override (only the `DEFAULT` layer, value `0` = HANA decides).

## Create a Reconfigure HanaDBOpsRequest

Change `global_allocation_limit` to 9 GiB (`9663676416` bytes) and force a restart so the value takes
effect:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: hanadb-standalone
  configuration:
    applyConfig:
      global.ini: |
        [memorymanager]
        global_allocation_limit = 9663676416
    restart: "true"
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/reconfigure/reconfigure.yaml
```
hanadbopsrequest.ops.kubedb.com/hdbops-reconfigure created

Here,

- `spec.configuration.applyConfig` carries the `global.ini` fragment. The fragment is **merged** with the
  database's existing inline configuration (only the keys you list are changed).
- `spec.configuration.restart: "true"` forces a pod restart so the new `global.ini` is applied. Use
  `"auto"` (the default) to let KubeDB decide, or `"false"` to skip the restart.

## Verify the Reconfiguration

Wait for the ops request to reach `Successful`:

```bash
kubectl get hdbops -n demo hdbops-reconfigure
```
NAME                 TYPE          STATUS       AGE
hdbops-reconfigure   Reconfigure   Successful   111s

```bash
kubectl describe hdbops -n demo hdbops-reconfigure
```
...
Status:
  Conditions:
    Reason:  Reconfigure
    Status:  True
    Type:    Reconfigure
    Reason:  DatabasePauseSucceeded
    Status:  True
    Type:    DatabasePauseSucceeded
    Reason:  UpdatePetSets
    Status:  True
    Type:    UpdatePetSets
    Reason:  RestartPods
    Status:  True
    Type:    RestartPods
    Reason:  Successful
    Status:  True
    Type:    Successful
  Phase:     Successful

Confirm the new value is live (it shows up under the `SYSTEM` layer once the pod has restarted with the
updated configuration):

```bash
kubectl exec -n demo hanadb-standalone-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$HANA_PASSWORD' \
  \"SELECT LAYER_NAME, VALUE FROM M_INIFILE_CONTENTS WHERE FILE_NAME='global.ini' AND SECTION='memorymanager' AND KEY='global_allocation_limit'\""
```
LAYER_NAME,VALUE
"DEFAULT","0"
"SYSTEM","9663676416"
2 rows selected

The `SYSTEM` layer now carries the new value `9663676416` (9 GiB).

To **remove** all custom configuration, set `spec.configuration.removeCustomConfig: true` in the
`HanaDBOpsRequest` instead of `applyConfig`.

## Cleaning Up

```bash
kubectl delete hdbops -n demo hdbops-reconfigure
```

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-standalone
```

```bash
kubectl delete ns demo
```

## Next Steps

- Set initial configuration at creation time with [Custom Configuration](/docs/guides/hanadb/configuration/using-config-file.md).
- [Vertically scale](/docs/guides/hanadb/scaling/vertical-scaling/vertical-scaling.md) a HanaDB.
