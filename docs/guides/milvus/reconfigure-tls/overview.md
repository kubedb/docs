---
title: Reconfigure Milvus TLS/SSL Overview
menu:
  docs_{{ .version }}:
    identifier: milvus-reconfigure-tls-overview
    name: Overview
    parent: milvus-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of Milvus

This guide will give an overview on how KubeDB Ops-manager operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer and rotate the certificates of a `Milvus` database.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

## How Reconfiguring Milvus TLS Configuration Process Works

A `MilvusOpsRequest` of type `ReconfigureTLS` drives every TLS change. Depending on which field you set in `spec.tls`, the operator performs one of four operations:

| Field | Operation |
| --- | --- |
| `spec.tls.issuerRef` + `spec.tls.external`/`spec.tls.internal` | **Add** TLS to a non-TLS database (or update the protocol mode). |
| `spec.tls.rotateCertificates: true` | **Rotate** the existing certificates (re-issue from the same issuer). |
| `spec.tls.issuerRef` pointing at a new issuer | **Change the issuer** so future certificates chain to a different CA. |
| `spec.tls.remove: true` | **Remove** TLS from the database. |

The high-level flow is:

1. A user creates a `MilvusOpsRequest` of type `ReconfigureTLS`.
2. The operator validates the request and pauses the referenced `Milvus` database.
3. cert-manager issues (or re-issues) the `server` and `client` certificates into the `<db>-server-cert` and `<db>-client-cert` secrets, or those secrets are removed for a TLS-removal request.
4. The operator updates the rendered `milvus.yaml` (`internaltls`/`tlsMode`/`security`) and the PetSets, mounting the certificates at `/milvus/tls`.
5. Pods are restarted to pick up the new TLS material.
6. Once all pods are healthy, the operator resumes the database and marks the `MilvusOpsRequest` as `Successful`.

## Milvus TLS Layers

Milvus exposes two independent TLS surfaces, both configured under `spec.tls`:

- **`external`** — controls client-facing traffic (gRPC/REST on port `19530`). Modes: `Disabled`, `TLS` (server-only) and `mTLS` (mutual — clients must present the `client` certificate).
- **`internal`** — controls inter-component traffic between Milvus roles. Modes: `Disabled` and `TLS`.

Certificates are described by aliases:

- `server` — server certificate used by the database endpoints.
- `client` — client certificate (used by the database for mutual auth and mounted for client tooling).

In the next docs, we will see a step-by-step guide on reconfiguring TLS of a Milvus database using `MilvusOpsRequest`.
