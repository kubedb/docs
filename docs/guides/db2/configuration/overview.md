---
title: DB2 Configuration Overview
menu:
  docs_{{ .version }}:
    identifier: db2-configuration-overview
    name: Overview
    parent: db2-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# DB2 Configuration

This guide shows the main DB2 configuration knobs you can adjust before creating or updating a database.

## Before You Begin

- Review the base DB2 manifest from [Quickstart](/docs/guides/db2/quickstart/quickstart.md).
- Decide whether you need to change credentials, pod resources, service exposure, or probe behavior.

## Common Configuration Fields

- spec.authSecret for custom credentials.
- spec.podTemplate for resources, scheduling, and environment.
- spec.serviceTemplates for service exposure.
- spec.healthChecker for readiness and liveness thresholds.

## How to Apply Configuration

Start from the quickstart manifest, add the fields you need under `spec`, and then re-apply the DB2 resource.

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/db2/quickstart/standalone.yaml
kubectl edit db2 -n demo db2
```

## Verify

```bash
kubectl get db2 -n demo db2 -o yaml
kubectl describe db2 -n demo db2
```
