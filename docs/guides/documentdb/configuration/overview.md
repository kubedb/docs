---
title: DocumentDB Configuration Overview
menu:
  docs_{{ .version }}:
    identifier: documentdb-configuration-overview
    name: Overview
    parent: documentdb-configuration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# DocumentDB Configuration

This guide shows the main DocumentDB configuration fields you can tune before or after initial deployment.

## Before You Begin

- Review the base manifest used in [Quickstart](/docs/guides/documentdb/quickstart/quickstart.md).
- Identify whether you need to change topology, authentication, services, or probes.

## Common Configuration Fields

- spec.replicas for number of database instances.
- spec.authSecret for custom credentials.
- spec.podTemplate for pod-level customization.
- spec.serviceTemplates for service exposure.
- spec.healthChecker for probe timing and thresholds.

## How to Apply Configuration

Update the DocumentDB resource manifest with the desired fields and apply it again.

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/documentdb/quickstart/standalone.yaml
kubectl edit documentdb -n demo documentdb
```

## Verify

```bash
kubectl get documentdb -n demo documentdb -o yaml
kubectl describe documentdb -n demo documentdb
```
