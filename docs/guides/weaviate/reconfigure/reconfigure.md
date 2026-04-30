---
title: Reconfigure Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-reconfigure-cluster
    name: Cluster
    parent: weaviate-reconfigure
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Weaviate

This repository does not currently contain a `WeaviateOpsRequest` Go type or CRD.

Because of that, there is no repository-backed manifest for a Weaviate reconfiguration request yet. The example files under `docs/examples/weaviate` are placeholders and should not be treated as CRD-validated manifests until the API is added.

## What You Can Safely Do Today

- Use `spec.configuration.secretName` in `Weaviate` CRD for configuration changes.
- Apply those changes through a normal `Weaviate` spec update workflow.

## How to Confirm API Availability in Your Cluster

Run the following commands to check whether your installed release has introduced the missing API:

```bash
kubectl get crd | grep -i weaviate
kubectl get crd | grep -i opsrequest
```

If your cluster includes `weaviateopsrequests.ops.kubedb.com`, follow your release-specific docs. Otherwise, keep using direct `Weaviate` spec updates.