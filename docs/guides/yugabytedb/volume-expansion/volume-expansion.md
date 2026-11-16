---
title: YugabyteDB Volume Expansion Guide
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-volume-expansion-guide
    name: Volume Expansion
    parent: yugabytedb-volume-expansion
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Expand YugabyteDB Volumes

> **Not verified.** There is no example file for this page under `docs/examples` on purpose: the manifest below was applied once, but it did not complete.

## What happened in the test

On a kind cluster whose storage class does not allow expansion, the request below ran `OrphanPetSetPods`, `UpdatePetSets` and `ReadyPetSets` and then failed in `UpdateVolumes` after the 3 minute step timeout. No pod restarted and the database stayed `Ready`. The PetSet template then said `8Gi` while `YugabyteDB.spec` still said `5Gi`. Neither the `Online` nor the `Offline` mode was verified to completion, and a retry on a storage class with expansion was not done.

## Before you try

- Check that the storage class allows expansion:

```bash
$ kubectl get storageclass
```

The column `ALLOWVOLUMEEXPANSION` must be `true` for the class used by the universe.

- YugabyteDB rejects writes when the data disk has less than 5% free space, so expand before the disk is nearly full.

## The request (not verified)

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-volexp
  namespace: demo
spec:
  type: VolumeExpansion
  databaseRef: {name: yb-demo}
  volumeExpansion:
    mode: Online
    tserver: 8Gi
  timeout: 3m
  maxRetries: 1
```

Use `mode: Offline` if your CSI driver can only resize while the volume is not mounted. Because each step has a timeout (`spec.timeout`), choose a value that fits the speed of your storage.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-volexp
```
