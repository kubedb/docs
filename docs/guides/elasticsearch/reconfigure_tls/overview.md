---
title: Reconfiguring TLS/SSL Overview
menu:
    docs_{{ .version }}:
        identifier: es-reconfigure-tls-overview
        name: Overview
        parent: es-reconfigure-tls-elasticsearch
        weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of Elasticsearch

This guide will give an overview on how KubeDB Ops-manager operator reconfigures TLS configuration i.e. add TLS, remove TLS, update issuer/cluster issuer or Certificates and rotate the certificates of `Elasticsearch`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
- [Elasticsearch](/docs/guides/Elasticsearch/concepts/elasticsearch.md)
- [ElasticsearchOpsRequest](/docs/guides/Elasticsearch/concepts/elasticsearch-ops-request.md)

## How Reconfiguring Elasticsearch TLS Configuration Process Works

The following diagram shows how KubeDB Ops-manager operator reconfigures TLS of a `Elasticsearch`. Open the image in a new tab to see the enlarged version.

<figure align="center">
      <img alt="Reconfiguring TLS process of Elasticsearch" src="/docs/guides/elasticsearch/reconfigure_tls/es-tls.png">
    <figcaption align="center">Fig: Reconfiguring TLS process of Elasticsearch</figcaption>
</figure>

The Reconfiguring Elasticsearch TLS process consists of the following steps:

1. At first, a user creates a `Elasticsearch` Custom Resource Object (CRO).

2. `KubeDB` Provisioner  operator watches the `Elasticsearch` CRO.

3. When the operator finds a `Elasticsearch` CR, it creates required number of `PetSets` and related necessary stuff like secrets, services, etc.

4. Then, in order to reconfigure the TLS configuration of the `Elasticsearch` database the user creates a `ElasticsearchOpsRequest` CR with desired information.

5. `KubeDB` Ops-manager operator watches the `ElasticsearchOpsRequest` CR.

6. When it finds a `ElasticsearchOpsRequest` CR, it pauses the `Elasticsearch` object which is referred from the `ElasticsearchOpsRequest`. So, the `KubeDB` Provisioner  operator doesn't perform any operations on the `Elasticsearch` object during the reconfiguring TLS process.

7. Then the `KubeDB` Ops-manager operator will add, remove, update or rotate TLS configuration based on the Ops Request yaml.

8. Then the `KubeDB` Ops-manager operator will restart all the Pods of the database so that they restart with the new TLS configuration defined in the `ElasticsearchOpsRequest` CR.

9. After the successful reconfiguring of the `Elasticsearch` TLS, the `KubeDB` Ops-manager operator resumes the `Elasticsearch` object so that the `KubeDB` Provisioner  operator resumes its usual operations.

In the next docs, we are going to show a step by step guide on reconfiguring TLS configuration of a Elasticsearch database using `ElasticsearchOpsRequest` CRD.