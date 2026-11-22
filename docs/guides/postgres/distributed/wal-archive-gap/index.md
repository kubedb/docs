---
title: Accept a WAL Archive Gap
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-wal-archive-gap
    name: WAL Archive Gap
    parent: guides-postgres-distributed
    weight: 75
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Accept a WAL Archive Gap

With continuous archiving (`spec.archiver`), the WAL archiver of a DC-DR Postgres pushes every WAL file of the server's timeline history to the backup storage. Point-in-time recovery (PITR) can only replay WAL that is in storage without holes.

Sometimes a WAL file goes missing from the backup storage, and no pod still has a copy. This can happen after repeated failovers in a short time. The archiver then holds its archive position at the gap, so the PITR range in the incremental Snapshot stops advancing. It keeps retrying.

## How the gap is reported

If the same gap stays open for about 10 minutes, KubeDB sets the `WALArchiveGap` condition on the Postgres:

```bash
$ kubectl get postgres -n demo pg-dr -o jsonpath='{.status.conditions[?(@.type=="WALArchiveGap")]}' | jq
{
  "type": "WALArchiveGap",
  "status": "True",
  "reason": "WALArchiveGapUnclosable",
  "message": "WAL archive gap 000000340000000000000028..00000034000000000000002A (3 file(s), timeline 52) since 2026-11-22T10:00:00Z: ...",
  ...
}
```

The message names the first and last missing files, and gives the command to accept the gap. The condition is removed as soon as the gap is closed, for example when the missing files are copied back into a pod's `pg_wal`.

## Accept the gap

Accept the gap only when the missing files cannot be recovered. After that, **PITR to any time inside the gap is impossible**.

Annotate the Postgres with the **first** missing file from the condition message:

```bash
$ kubectl annotate postgres -n demo pg-dr archiver.kubedb.com/accept-wal-gap=000000340000000000000028
```

KubeDB checks the annotation against the reported gap:

- If it names that gap's first file, the archiver releases its hold past exactly that gap. The condition reason changes to `WALArchiveGapAccepted`, and is removed once archiving moves on.
- An annotation naming any other file is ignored. It never covers a different or a larger gap.
- If another gap is found later, it is reported again and needs its own acceptance.

Then take a new full backup, so PITR has a clean starting point after the gap.
