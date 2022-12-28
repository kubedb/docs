---
title: ElasticsearchVersion CRD
menu:
  docs_{{ .version }}:
    identifier: es-catalog-concepts
    name: ElasticsearchVersion
    parent: es-concepts-elasticsearch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# ElasticsearchVersion

## What is ElasticsearchVersion

`ElasticsearchVersion` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration to specify the docker images to be used for [Elasticsearch](https://www.elastic.co/products/elasticsearch), [Kibana](https://www.elastic.co/products/kibana) and [OpenSearch](https://opensearch.org/), [OpenSearch-Dashboards](https://opensearch.org/docs/latest/dashboards/index/) deployed with KubeDB in a Kubernetes native way.

When you install KubeDB, an `ElasticsearchVersion` custom resource will be created automatically for every supported Elasticsearch and OpenSearch version. You have to specify the name of `ElasticsearchVersion` CRD in `spec.version` field of [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) CRD. Then, KubeDB will use the docker images specified in the `ElasticsearchVersion` CRD to create your expected database. If you want to provision `Kibana` or `Opensearch-Dashboards`, you have to specify the name of `Elasticsearch` CRD in `spec.databaseRef.name` field of [ElasticsearchDashboard](/docs/guides/elasticsearch/concepts/elasticsearch-dashboard/index.md) CRD. Then, KubeDB will use the compatible docker image specified in the `.spec.dashboard.image` field of the `ElasticsearchVersion` CRD that Elasticsearch is using to create your expected dashboard.

Using a separate CRD for specifying respective docker images, and pod security policy names allow us to modify the images, and policies independent of the KubeDB operator. This will also allow the users to use a custom image for the database.

## ElasticsearchVersion Specification

As with all other Kubernetes objects, an ElasticsearchVersion needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: ElasticsearchVersion
metadata:
  annotations:
    meta.helm.sh/release-name: kubedb
    meta.helm.sh/release-namespace: kubedb
  creationTimestamp: "2022-12-26T04:28:09Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: kubedb
    app.kubernetes.io/managed-by: Helm
    app.kubernetes.io/name: kubedb-catalog
    app.kubernetes.io/version: v2022.12.24-rc.1
    helm.sh/chart: kubedb-catalog-v2022.12.24-rc.1
  name: xpack-8.2.0
  resourceVersion: "236918"
  uid: 55abfde5-a8cb-486b-b73c-1a2b097e96a3
spec:
  authPlugin: X-Pack
  dashboard:
    image: kibana:8.2.0
  dashboardInitContainer:
    yqImage: kubedb/elasticsearch-dashboard-init:8.2.0-xpack-v2022.05.24
  db:
    image: elasticsearch:8.2.0
  distribution: ElasticStack
  exporter:
    image: prometheuscommunity/elasticsearch-exporter:v1.3.0
  initContainer:
    image: tianon/toybox:0.8.4
    yqImage: kubedb/elasticsearch-init:8.2.0-xpack-v2022.05.24
  podSecurityPolicies:
    databasePolicyName: elasticsearch-db
  securityContext:
    runAsAnyNonRoot: true
    runAsUser: 1000
  stash:
    addon:
      backupTask:
        name: elasticsearch-backup-8.2.0
        params:
          - name: args
            value: --match=^(?![.])(?!apm-agent-configuration)(?!kubedb-system).+
      restoreTask:
        name: elasticsearch-restore-8.2.0
        params:
          - name: args
            value: --match=^(?![.])(?!apm-agent-configuration)(?!kubedb-system).+
  version: 8.2.0
```

### metadata.name

`metadata.name` is a required field that specifies the name of the `ElasticsearchVersion` CRD. You have to specify this name in `spec.version` field of [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md) CRD.

We follow this convention for naming ElasticsearchVersion CRD:

- Name format: `{Security Plugin Name}-{Application Version}-{Modification Tag}`

- Samples: `searchguard-7.9.3`, `xpack-7.9.1-v1`, `opendistro-1.12.0`, etc.

We use the original Elasticsearch docker image provided by the distributors. Then we bundle the image with the necessary sidecar and init container images which facilitate features like sysctl kernel settings, custom configuration, monitoring matrices, etc.  An image with a higher modification tag will have more features and fixes than an image with a lower modification tag. Hence, it is recommended to use ElasticsearchVersion CRD with the highest modification tag to take advantage of the latest features.

### spec.version

`spec.version` is a required field that specifies the original version of the Elasticsearch database that has been used to build the docker image specified in `spec.db.image` field.

### spec.deprecated

`spec.deprecated` is an optional field that specifies whether the docker images specified here is supported by the current KubeDB operator. For example, we have modified `kubedb/elasticsearch:7.x.x-xpack` docker images to support custom configuration and re-tagged as `kubedb/elasticsearch:7.x.x-xpack-v1`. Now, KubeDB operator `version:x.y.z` supports providing custom configuration which required `kubedb/elasticsearch:7.x.x-xpack-v1` docker images. So, we have marked `kubedb/elasticsearch:7.x.x-xpack` as deprecated in KubeDB `version:x.y.z`.

The default value of this field is `false`. If `spec.deprecated` is set `true`, the KubeDB operator will not create the database and other respective resources for this version.

### spec.db.image

`spec.db.image` is a `required` field that specifies the docker image which will be used to create StatefulSet by KubeDB provisioner operator to create the expected Elasticsearch/OpenSearch database.

### spec.dashboard.image
`spec.dashboard.image` is an `optional` field that specifies the docker image which will be used to create Deployment by KubeDB dashboard operator to create the expected Kibana/Opensearch-dashboards.

### spec.exporter.image

`spec.exporter.image` is a `required` field that specifies the image which will be used to export Prometheus metrics if monitoring is enabled.

### spec.podSecurityPolicies.databasePolicyName

`spec.podSecurityPolicies.databasePolicyName` is a required field that specifies the name of the pod security policy required to get the database server pod(s) running.

```bash
helm upgrade kubedb-operator appscode/kubedb --namespace kube-system \
  --set additionalPodSecurityPolicies[0]=custom-db-policy \
  --set additionalPodSecurityPolicies[1]=custom-snapshotter-policy
```

## Next Steps

- Learn about Elasticsearch CRD [here](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Deploy your first Elasticsearch database with KubeDB by following the guide [here](/docs/guides/elasticsearch/quickstart/overview/elasticsearch/index.md).
- Deploy your first OpenSearch database with KubeDB by following the guide [here](/docs/guides/elasticsearch/quickstart/overview/opensearch/index.md).
