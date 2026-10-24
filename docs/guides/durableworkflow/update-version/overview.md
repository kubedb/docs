---
title: Update Version Overview
menu:
  docs_{{ .version }}:
    identifier: dw-updateversion-overview
    name: Overview
    parent: dw-updateversion-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Update Version Overview

This page explains how the KubeDB ops-manager updates a DurableWorkflow without downtime, and how it rolls back. The same pod replacement is used by Restart, Reconfigure of the static configuration, VerticalScaling, RotateAuth and the rotation of Cadence certificates.

## How the pods are rolled

The PetSets of the roles use the `OnDelete` update strategy, so Kubernetes never replaces a pod by itself. A pod needs replacing when its `controller-revision-hash` differs from the `updateRevision` of its PetSet. A restart, a new static configuration, a new resource request or a new version all work by creating a new revision, and the same code then replaces the pods:

1. The roles are rolled in the order of `rollout.roleOrder` of the `DurableWorkflowVersion`: for Temporal `history`, `matching`, `internalFrontend`, `frontend`, `worker`, and for Cadence `history`, `matching`, `frontend`, `worker`. Within a role it is **one pod at a time**, and the next pod waits until the previous one is ready. A PodDisruptionBudget with `maxUnavailable: 1` keeps anything else from taking a second pod.
2. **Surge.** For every role except `history`, the ops-manager first scales the role up by one pod, rolls, and scales it back down, so a role never has less capacity than configured. `history` is rolled without a surge, because adding and removing a pod moves shards twice. Set `updateVersion.surge` to override it per role.
3. **Ready means able to serve.** For Temporal the readiness uses the gRPC health protocol, so a replaced `history` pod becomes ready only after it owns its shards, and a `frontend` pod only when its membership is set up. For Cadence the frontend runs a CLI check, and the other roles check their port.
4. **Draining.** Before a pod stops, it sleeps for `rollout.preStopSleep` (10 seconds by default) so that the endpoints are removed everywhere. The engines then drain on `SIGTERM`: the drain durations in the catalog (frontend 10 to 60 seconds, matching 10, history 30) are set in the dynamic configuration, because the engines default them to 0 and would drop requests on every replacement. A `history` pod leaves the membership ring first and keeps serving during its drain, so another pod takes its shards before it stops.
5. **History pacing.** At least `rollout.historyMinInterval` (60 seconds by default) passes between two `history` pods, so the database can absorb the cold caches of the new pod.
6. **Timeout.** A replacement pod may take a third of the request `timeout` to become ready (5 minutes when no `timeout` is set). If it does not, the request fails.

This was measured with a load generator that started and signaled a workflow every 0.3 seconds, with 2 replicas per role on Postgres. Restart, UpdateVersion (Temporal and Cadence), scaling, Reconfigure and certificate rotation produced **0 failed calls**. Using a plain TCP readiness probe instead of the checks above failed 6 to 7 calls per run, because an open port says nothing about whether a Temporal frontend can serve.

## How UpdateVersion works

```text
validate -> update the schema -> render the new version -> roll the pods -> verify -> switch spec.version
                                                              |
                                                     a pod does not become ready
                                                              v
                                                    roll back the replaced pods -> Failed
```

1. **Webhook.** The request is rejected if the target is of the other distribution (changing between Temporal and Cadence is a data migration), if the target is not in the `updateConstraints.allowlist` of the current version, or if the schema update of the target cannot run online (`rollout.onlineSchemaUpdate: false`) and `updateVersion.allowDowntime` is not `true`. It warns if the cluster is not highly available.
2. **Schema first.** A schema Job applies the schema of the target version to the default and the visibility store **while every pod still runs the current version**. Both engines document that the previous version runs on the newer schema.
3. **Render.** The configuration and the PetSets are updated to the target version. The running pods are not touched.
4. **Roll.** The pods are replaced as above.
5. **Verify** the servers, then `spec.version` of the DurableWorkflow is switched to the target.

### Rollback

If a replacement pod does not become ready, for example because the new image crashes, the ops-manager stops, points the PetSets back at the current version and replaces the already-updated pods again, one at a time. This is safe because the previous version runs on the newer schema, so **the schema stays at the newer version**. The request ends `Failed`, and the cluster kept serving during the whole time. A crash-looping target was detected after about 3 minutes and rolled back with 0 failed calls. A target image that does not exist fails earlier and changes nothing, since the operator resolves the images to digests before it touches a pod.

### Which versions can I update to

The `updateConstraints.allowlist` of the **current** `DurableWorkflowVersion` says what it can be updated to, for example `> 1.32.1, < 1.33.0` for `temporal-1.32.1`. Update through the next versions one after another: neither engine guarantees that two versions that are not adjacent can run together. Cadence's guarantee is between adjacent releases, patches included.

Today the KubeDB catalog has one version of each engine, `temporal-1.32.1` and `cadence-1.4.2`, so there is nothing newer to update to yet. The update paths that were tested are Temporal 1.32.0 to 1.32.1 and Cadence 1.4.1 to 1.4.2, using catalog entries of the previous versions.

## Next Steps

- Run an [Update Version](/docs/guides/durableworkflow/update-version/durableworkflow.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
