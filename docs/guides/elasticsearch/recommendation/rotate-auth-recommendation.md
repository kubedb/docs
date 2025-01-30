---
title: Elasticsearch Rotate Auth Recommendation
menu:
  docs_{{ .version }}:
    identifier: es-rotate-auth-recommendation
    name: Rotate Auth Recommendation
    parent: es-recommendation-elasticsearch
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Elasticsearch Version Update Recommendation

Rotating authentication secrets in database management is vital to mitigate security risks, such as credential leakage or unauthorized access, and to comply with regulatory requirements. Regular rotation limits the exposure of compromised credentials, reduces the risk of insider threats, and enforces updated security policies like stronger passwords or algorithms. It also ensures operational resilience by testing the rotation process and revoking stale or unused credentials. KubeDB provides `RotateAuth OpsRequest` which reduces manual errors, and strengthens database security with minimal effort. KubeDB Ops-manager generates Recommendation for rotating authentication secrets via this OpsRequest.

`Recommendation` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative recommendation for KubeDB managed databases like [Elasticsearch](https://www.elastic.co/products/elasticsearch) and [OpenSearch](https://opensearch.org/) in a Kubernetes native way. KubeDB generates Elasticsearch/Opensearch Rotate Auth recommendation regarding three particular cases.

1. There's been an update in the current version image
2. There's a new version available with a minor/patch fix.
3. There's a new major version available

Let's go through a demo to see version update recommendations being generated. First, get the available Elasticsearch versions provided by KubeDB.

```bash
$ kubectl get elasticsearchversions | grep xpack
xpack-6.8.23        6.8.23    ElasticStack   ghcr.io/appscode-images/elastic:6.8.23                        17h
xpack-7.13.4        7.13.4    ElasticStack   ghcr.io/appscode-images/elastic:7.13.4                        17h
xpack-7.14.2        7.14.2    ElasticStack   ghcr.io/appscode-images/elastic:7.14.2                        17h
xpack-7.16.3        7.16.3    ElasticStack   ghcr.io/appscode-images/elastic:7.16.3                        17h
xpack-7.17.15       7.17.15   ElasticStack   ghcr.io/appscode-images/elastic:7.17.15                       17h
xpack-7.17.23       7.17.23   ElasticStack   ghcr.io/appscode-images/elastic:7.17.23                       17h
xpack-7.17.25       7.17.25   ElasticStack   ghcr.io/appscode-images/elastic:7.17.25                       16h
xpack-8.11.1        8.11.1    ElasticStack   ghcr.io/appscode-images/elastic:8.11.1                        17h
xpack-8.11.4        8.11.4    ElasticStack   ghcr.io/appscode-images/elastic:8.11.4                        17h
xpack-8.13.4        8.13.4    ElasticStack   ghcr.io/appscode-images/elastic:8.13.4                        17h
xpack-8.14.1        8.14.1    ElasticStack   ghcr.io/appscode-images/elastic:8.14.1                        17h
xpack-8.14.3        8.14.3    ElasticStack   ghcr.io/appscode-images/elastic:8.14.3                        17h
xpack-8.15.0        8.15.0    ElasticStack   ghcr.io/appscode-images/elastic:8.15.0                        17h
xpack-8.15.4        8.15.4    ElasticStack   ghcr.io/appscode-images/elastic:8.15.4                        16h
xpack-8.16.0        8.16.0    ElasticStack   ghcr.io/appscode-images/elastic:8.16.0                        16h
xpack-8.2.3         8.2.3     ElasticStack   ghcr.io/appscode-images/elastic:8.2.3                         17h
xpack-8.5.3         8.5.3     ElasticStack   ghcr.io/appscode-images/elastic:8.5.3                         17h
xpack-8.6.2         8.6.2     ElasticStack   ghcr.io/appscode-images/elastic:8.6.2                         17h
xpack-8.8.2         8.8.2     ElasticStack   ghcr.io/appscode-images/elastic:8.8.2                         17h
```

Let's deploy an Elasticsearch cluster with version `xpack-8.15.0`.

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es
  namespace: demo
spec:
  version: xpack-8.15.0
  storageType: Durable
  deletionPolicy: WipeOut
  topology:
    master:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    data:
      replicas: 2
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    ingest:
      replicas: 1
      storage:
        storageClassName: "standard"
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
```



## Next Steps

- Learn about [backup & restore](/docs/guides/elasticsearch/backup/stash/overview/index.md) Elasticsearch database using Stash.
- Learn how to configure [Elasticsearch Topology Cluster](/docs/guides/elasticsearch/clustering/topology-cluster/simple-dedicated-cluster/index.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
