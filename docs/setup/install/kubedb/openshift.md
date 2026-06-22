---
title: Install KubeDB on OpenShift
description: Install KubeDB on OpenShift
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-openshift
    name: OpenShift
    parent: install-kubedb-enterprise
    weight: 50
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Install on OpenShift

KubeDB supports OpenShift in three different ways. Pick the one that best matches your deployment workflow.

## Option A: Standard KubeDB chart with OpenShift overrides

The standard `kubedb` chart can be deployed on OpenShift by enabling the OpenShift distro flags. The `openshift` flag is also auto-detected when the cluster exposes the `project.openshift.io/v1` API, so you can leave it `false` and only switch the image flavor to UBI.

```bash
$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.distro.openshift=false \
  --set global.distro.ubi=operator \
  --wait --burst-limit=10000 --debug
```

Equivalent values file:

```yaml
global:
  distro:
    openshift: false
    ubi: "operator"
```

## Option B: Red Hat OpenShift Certified Helm Chart

AppsCode publishes a Red Hat OpenShift Certified Helm chart, [`kubedb-certified`](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-certified). This chart does **not** ship the KubeDB CRDs; you must install the companion [`kubedb-certified-crds`](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-certified-crds) chart first.

**Step 1 — Install the CRDs:**

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm upgrade -i kubedb-certified-crds appscode/kubedb-certified-crds \
  -n kubedb --create-namespace \
  --version={{< param "info.version" >}}
```

**Step 2 — Install the operator:**

```bash
$ helm upgrade -i kubedb-certified appscode/kubedb-certified \
  -n kubedb --create-namespace \
  --version={{< param "info.version" >}} \
  --set-file global.license=/path/to/the/license.txt
```

## Option C: Red Hat OpenShift OperatorHub

KubeDB is a [Red Hat Certified Operator](https://catalog.redhat.com/en/software/container-stacks/detail/6867c6a358efc229b095b8ee#overview) in the OpenShift OperatorHub catalog. You can install the KubeDB operator bundle directly from the OpenShift web console (**Operators → OperatorHub → KubeDB**) or with `oc`.

Once the operator bundle is installed, create a `Kubedb` installer resource to deploy the KubeDB operator components. Toggle the `featureGates` to match the database engines you intend to use, and set either `license` (inline content) or `licenseSecretName` (a Secret with key `key.txt`) — get a license from the [AppsCode License Server](https://appscode.com/issue-license?p=kubedb).

```yaml
apiVersion: installer.kubedb.com/v1
kind: Kubedb
metadata:
  name: kubedb
  namespace: kubedb
spec:
  global:
    featureGates:
      Cassandra: false
      ClickHouse: false
      DB2: false
      DocumentDB: false
      Druid: false
      Elasticsearch: true
      HanaDB: false
      Hazelcast: false
      Ignite: false
      Kafka: true
      MariaDB: true
      Memcached: false
      Milvus: false
      MongoDB: true
      MSSQLServer: false
      MySQL: true
      Neo4j: false
      Oracle: false
      PerconaXtraDB: false
      PgBouncer: false
      Pgpool: false
      Postgres: true
      ProxySQL: false
      Qdrant: false
      RabbitMQ: false
      Redis: true
      Singlestore: false
      Solr: false
      Weaviate: false
      ZooKeeper: false
    imagePullSecrets: []
    insecureRegistries: []
    license: ''
    licenseSecretName: ''
    networkPolicy:
      enabled: false
    registry: ''
    registryFQDN: ''
```

Apply it with:

```bash
$ oc apply -f kubedb.yaml
```

Next: [enable database engines and verify the installation](/docs/setup/install/kubedb/configuration.md).
