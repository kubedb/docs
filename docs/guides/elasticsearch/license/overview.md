---
title: Elasticsearch License Overview
menu:
  docs_{{ .version }}:
    identifier: es-license-overview
    name: Overview
    parent: es-license-elasticsearch
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch License Activation

Elastic ships every distribution of Elasticsearch with the full X-Pack feature set built in, but most
of it (security, ILM, ML, and so on) only runs under a paid subscription or a trial license. There is
no separate "Enterprise" Docker image to switch to -- activating a subscription is a runtime call
against the cluster's `_license` API, not an image change. KubeDB lets you drive that activation
declaratively from the `Elasticsearch` CRO, or on demand via an `ElasticsearchOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)

- License activation is only supported for the `ElasticStack` distribution (the upstream Elastic
  distribution). It is not applicable to `OpenSearch`, `OpenDistro`, `SearchGuard`, or `KubeDB`
  distribution `ElasticsearchVersion`s, since none of those ship X-Pack's commercial license model.

## Two Ways to Activate a License

1. **Bring your own license (`secretRef`):** reference a Kubernetes Secret containing the signed
   license file you downloaded from Elastic, under key `license.json`.
2. **Trial (`trial: true`):** ask the operator to activate Elastic's built-in one-time 30-day trial
   license (`POST _license/start_trial`) instead. Elastic allows this exactly once per cluster; if the
   trial has already been used, the operator surfaces the failure via a condition instead of retrying
   forever.

Both are set on `spec.license` on the `Elasticsearch` object, and are mutually exclusive.

## How License Activation Works

1. A user creates an `Elasticsearch` CRO with `spec.license` set to either `secretRef` or `trial`.

2. The KubeDB webhook validates the request: `spec.license` is only accepted when the referenced
   `ElasticsearchVersion` uses the `ElasticStack` distribution, exactly one of `secretRef`/`trial` is
   set, and (for `secretRef`) the referenced Secret exists and carries a `license.json` key.

3. Once the cluster is `Ready`, the KubeDB Provisioner operator's health-check loop reads
   `spec.license` on every pass and calls the cluster's `_license` API to bring the installed license
   in line with it -- installing/replacing it for `secretRef`, or starting the trial for `trial`.

4. The operator records the outcome as `Conditions` on the `Elasticsearch` object's status:
   - `LicenseActive` -- `True` once the cluster reports an active, non-basic license. `Message`
     carries the license type and expiry date.
   - `LicenseExpiringSoon` -- `True` when the active license expires within 14 days. This is a
     warning only; the cluster stays `Ready`.
   - `LicenseSyncFailed` -- `True` with the underlying API error in `Message` when the operator could
     not read or apply the license (unreachable cluster, malformed Secret, a second trial request
     after the one-time trial was already consumed, and so on).

5. Because this reconcile is continuous, license state is self-healing: if the license is lost (a
   manual `DELETE _license`, or a restore from an unlicensed snapshot), the next health-check pass
   re-applies it from `spec.license` without any user action.

6. To rotate to a different license (a renewed BYO license, or to request the trial) without waiting
   for the next health-check pass, create an `ElasticsearchOpsRequest` with `type: RotateLicense`. The
   KubeDB Ops-manager operator applies it immediately and patches `spec.license` on the `Elasticsearch`
   object to match, so the CR stays the source of truth for future reconciles.

In the next section, we walk through activating a BYO license, requesting a trial, and rotating a
license with `RotateLicense`.
