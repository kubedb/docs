---
title: Upgrade YugabyteDB Version
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-version-upgrading
    name: Version Updating
    parent: yugabytedb-update-version
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Upgrade YugabyteDB Version

This guide upgrades a universe from `2025.1.4` to `2025.2.8`.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/update-version](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/update-version) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

Create the `demo` namespace and the `spread-across-zones` PlacementPolicy as in the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md), and check that both versions are in the catalog (`kubectl get ybversion`).

## Deploy a universe on the old version

The manifest is the quickstart manifest with the name `yb-up` and `spec.version: 2025.1.4`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/update-version/yugabytedb.yaml
yugabytedb.kubedb.com/yb-up created
```

## Apply the UpdateVersion OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-up-update-version
  namespace: demo
spec:
  type: UpdateVersion
  databaseRef: {name: yb-up}
  updateVersion:
    targetVersion: 2025.2.8
    soakDuration: 2m
  rolling: {masterWait: 20s, tserverWait: 20s}
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/update-version/ops-request.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-up-update-version created

$ kubectl get ybops -n demo yb-up-update-version
```

The default soak is 10 minutes; 2 minutes were used in the test.

## What was measured

Under load (4 workers of insert plus select through the client Service):

| Measure | Result |
|---------|--------|
| Duration | 663 s (including the 2 minute soak) |
| Successful calls | 846,541 |
| Failed calls | 7, all dropped connections to restarted pods (3 x 57P01, 4 x unexpected EOF), no other failure |
| Longest period without any success | 0.39 s |

Afterwards `kubectl get yb -n demo yb-up` reported version `2025.2.8`. Conditions: `Preflight`, `RecordAutoFlagsConfig`, `UpdatePetSets`, `RollMasters`, `RollTServers`, `PromoteAutoFlags`, `Soak`, `FinalizeUpgrade`, `Successful`.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-up-update-version
$ kubectl delete yb -n demo yb-up
```
