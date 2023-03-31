---
title: Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-readme-elasticsearch
    name: Elasticsearch
    parent: es-elasticsearch-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/elasticsearch/
aliases:
  - /docs/{{ .version }}/guides/elasticsearch/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Elasticsearch Features

| Features                                                                                                                                                                                                                                                                        | Community     | Enterprise |
|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------| :----------:  |:----------:|
| Combined Cluster (n nodes with master,data,ingest: ture; n >= 1 )                                                                                                                                                                                                               |   &#10003;    |  &#10003;  |
| Topology Cluster (n master, m data, x ingest nodes; n,m,x >= 1 )                                                                                                                                                                                                                |   &#10003;    |  &#10003;  |
| Hot-Warm-Cold Topology Cluster (a hot, b warm, c cold nodes; a,b,c >= 1 )                                                                                                                                                                                                       |   &#10003;    |  &#10003;  |
| TLS: Add, Remove, Update, Rotate ( [Cert Manager](https://cert-manager.io/docs/) )                                                                                                                                                                                              |   &#10007;    |  &#10003;  |
| Automated Version Update                                                                                                                                                                                                                                                        |   &#10007;    |  &#10003;  |
| Automatic Vertical Scaling                                                                                                                                                                                                                                                      |   &#10007;    |  &#10003;  |
| Automated Horizontal Scaling                                                                                                                                                                                                                                                    |   &#10007;    |  &#10003;  |
| Automated Volume Expansion                                                                                                                                                                                                                                                      |   &#10007;    |  &#10003;  |
| Backup/Recovery: Instant, Scheduled ( [Stash](https://stash.run/) )                                                                                                                                                                                                             |   &#10003;    |  &#10003;  |
| Dashboard ( Kibana , Opensearch-Dashboards )                                                                                                                                                                                                                                    |   &#10003;    |  &#10003;  |
| Grafana Dashboards                                                                                                                                                                                                                                                              |    &#10007;    |  &#10003;  |
| Initialization from Snapshot ( [Stash](https://stash.run/) )                                                                                                                                                                                                                    |   &#10003;    |  &#10003;  |
| Authentication ( [OpensSearch](https://opensearch.org/) / [X-Pack](https://www.elastic.co/guide/en/elasticsearch/reference/7.9/setup-xpack.html) / [OpenDistro](https://opendistro.github.io/for-elasticsearch-docs/) / [Search Guard](https://docs.search-guard.com/latest/) ) |   &#10003;    |  &#10003;  |
| Authorization ( [OpensSearch](https://opensearch.org/) / [X-Pack](https://www.elastic.co/guide/en/elasticsearch/reference/7.9/setup-xpack.html) / [OpenDistro](https://opendistro.github.io/for-elasticsearch-docs/) / [Search Guard](https://docs.search-guard.com/latest/) )  |   &#10003;    |  &#10003;  |
| Persistent Volume                                                                                                                                                                                                                                                               |   &#10003;    |  &#10003;  |
| Exports Prometheus Matrices                                                                                                                                                                                                                                                     |   &#10003;    |  &#10003;  |
| Custom Configuration                                                                                                                                                                                                                                                            |   &#10003;    |  &#10003;  |
| Using Custom Docker Image                                                                                                                                                                                                                                                       |   &#10003;    |  &#10003;  |
| Initialization From Script                                                                                                                                                                                                                                                      |   &#10007;    |  &#10007;  |
| Autoscaling (vertically)	                                                                                                                                                                                                                                                       |   &#10007;    |  &#10003;  |

## Lifecycle of Elasticsearch Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->
<p align="center">
  <img alt="lifecycle"  src="/docs/guides/elasticsearch/images/Lifecycle-of-an-Elasticsearch-CRD-complete.png">
</p>



## Available Elasticsearch Versions

KubeDB supports `Elasticsearch` provided by Elastic with `xpack` auth plugin. `Opensearch` and `Opendistro` are supported too. KubeDB also supports some versions of Elasticsearch with `searchguard` auth plugin. Compatible `Kibana` and `Opensearch-Dashboards` are supported by Most of the Elasticsearch versions with `xpack` auth plugin and OpenSearch. `Kibana` and `Opensearch-Dashboards` can be provisioned externally or by using KubeDB with `ElasticsearchDashboard` CRD. 

<table>
<tr><th>X-Pack</th><th>OpenSearch</th></tr>
<tr>
<td>

| Version | ElasticSearch | Dashboard(Kibana) |
|:-------:|:-------------:|:-----------------:|
|  8.5.x  |   &#10003;    |     &#10003;      |
|  8.2.x  |   &#10003;    |     &#10003;      |
| 7.17.x  |   &#10003;    |     &#10003;      |
| 7.16.x  |   &#10003;    |     &#10003;      |
| 7.14.x  |   &#10003;    |     &#10003;      |
| 7.13.x  |   &#10003;    |     &#10003;      |
| 7.12.x  |   &#10003;    |     &#10003;      |
|  7.9.x  |   &#10003;    |     &#10003;      |
|  7.8.x  |   &#10003;    |     &#10007;      |
|  7.7.x  |   &#10003;    |     &#10007;      |
|  7.6.x  |   &#10003;    |     &#10007;      |
|  7.5.x  |   &#10003;    |     &#10007;      |
|  7.4.x  |   &#10003;    |     &#10007;      |
|  7.3.x  |   &#10003;    |     &#10007;      |
|  7.2.x  |   &#10003;    |     &#10007;      |
|  7.1.x  |   &#10003;    |     &#10007;      |
|  7.0.x  |   &#10003;    |     &#10007;      |
|  6.8.x  |   &#10003;    |     &#10003;      |
</td>
<td style="vertical-align:top">

| Version | OpenSearch | Dashboard<br/>(OpenSearch-Dashboards) |
|:-------:|:----------:|:-------------------------------------:|
|  1.3.x  |  &#10003;  |               &#10003;                |
|  1.2.x  |  &#10003;  |               &#10003;                |
|  1.1.x  |  &#10003;  |               &#10003;                |
</td>
</tr>
<tr><th>OpenDistro</th><th>SearchGuard</th></tr>
<tr>
<td>

| Version | ElasticSearch | Dashboard(Kibana) |
|:-------:|:-------------:|:-----------------:|
| 7.10.x  |   &#10003;    |     &#10007;      |
|  7.9.x  |   &#10003;    |     &#10007;      |
|  7.8.x  |   &#10003;    |     &#10007;      |
|  7.7.x  |   &#10003;    |     &#10007;      |
|  7.6.x  |   &#10003;    |     &#10007;      |
|  7.4.x  |   &#10003;    |     &#10007;      |
|  7.3.x  |   &#10003;    |     &#10007;      |
|  7.2.x  |   &#10003;    |     &#10007;      |
|  7.1.x  |   &#10003;    |     &#10007;      |
|  7.0.x  |   &#10003;    |     &#10007;      |
</td>
<td style="vertical-align:top">

|  Version   | ElasticSearch | Dashboard(Kibana) |
|:----------:|:-------------:|:-----------------:|
|   7.14.x   |   &#10003;    |     &#10007;      |
|   7.10.x   |   &#10003;    |     &#10007;      |
|   7.9.x    |   &#10003;    |     &#10007;      |
|   7.8.x    |   &#10003;    |     &#10007;      |
|   7.5.x    |   &#10003;    |     &#10007;      |
|   7.3.x    |   &#10003;    |     &#10007;      |
|   7.1.x    |   &#10003;    |     &#10007;      |
|   7.0.x    |   &#10003;    |     &#10007;      |
|   6.8.x    |   &#10003;    |     &#10007;      |
</td>

</tr>

</table>



> The listed ElasticsearchVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog/index.md) object with your custom Elasticsearch image.

## User Guide

- [Quickstart Elasticsearch](/docs/guides/elasticsearch/quickstart/overview/elasticsearch/index.md) with KubeDB Operator.
- [Quickstart OpenSearch](/docs/guides/elasticsearch/quickstart/overview/opensearch/index.md) with KubeDB Operator.
- [Quickstart Kibana](/docs/guides/elasticsearch/elasticsearch-dashboard/kibana/index.md) with KubeDB Operator.
- [Quickstart OpenSearch-Dashboards](/docs/guides/elasticsearch/elasticsearch-dashboard/opensearch-dashboards/index.md) with KubeDB Operator.
- [Elasticsearch Clustering](/docs/guides/elasticsearch/clustering/combined-cluster/index.md) supported by KubeDB
- [Backup & Restore Elasticsearch](/docs/guides/elasticsearch/backup/overview/index.md) database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Use [kubedb cli](/docs/guides/elasticsearch/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
