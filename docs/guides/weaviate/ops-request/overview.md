---
title: Weaviate Ops Request Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-ops-request-overview
    name: Overview
    parent: weaviate-ops-request
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Weaviate Ops Request

This guide provides an overview of day-2 operations for Weaviate databases managed by KubeDB.

A `WeaviateOpsRequest` allows you to run operational workflows in a declarative way. You submit a request, then monitor status until the operation completes.

## Before You Begin

- Deploy Weaviate first using the [quickstart guide](/docs/guides/weaviate/quickstart/quickstart.md).
- Install KubeDB and Ops Manager following [setup docs](/docs/setup/README.md).
- Use a dedicated namespace for testing.

```bash
kubectl create ns demo
```

## Documented Operation Categories

- [Reconfigure](/docs/guides/weaviate/reconfigure/overview.md)
- [VerticalScaling](/docs/guides/weaviate/scaling/vertical-scaling/overview.md)
- [VolumeExpansion](/docs/guides/weaviate/volume-expansion/overview.md)
- [UpdateVersion](/docs/guides/weaviate/update-version/overview.md)
- [RotateAuth](/docs/guides/weaviate/rotate-auth/overview.md)
- [Restart](/docs/guides/weaviate/restart/restart.md)

## How Ops Requests Work

Every operation follows the same high-level flow:

1. Deploy or identify a healthy `Weaviate` database.
2. Apply one `WeaviateOpsRequest` manifest for a single operation type.
3. Monitor request status and database readiness.
4. Validate the expected outcome.

Use these commands for monitoring:

```bash
kubectl get weaviateopsrequest -n demo
kubectl describe weaviateopsrequest -n demo <opsrequest-name>
kubectl get weaviate -n demo -w
kubectl get pods -n demo -l app.kubernetes.io/instance=weaviate-sample
```

For best results, avoid applying multiple ops requests for the same database at the same time.

## Next Steps

- Choose the specific operation page that matches your intended change.
- Apply one operation at a time and verify the request reaches `Successful` state.
