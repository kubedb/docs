---
title: Install KubeDB
description: Installation guide for KubeDB
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-enterprise
    name: KubeDB
    parent: installation-guide
    weight: 20
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Install KubeDB

## Get a Free License

Download a FREE license from [AppsCode License Server](https://appscode.com/issue-license?p=kubedb).

> KubeDB licensing process has been designed to work with CI/CD workflow. You can automatically obtain a license from your CI/CD pipeline by following the guide from [here](https://github.com/appscode/offline-license-server#offline-license-server).

## Install

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3 (Recommended)</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="argocd-tab" data-toggle="tab" href="#argocd" role="tab" aria-controls="argocd" aria-selected="false">ArgoCD</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="fluxcd-tab" data-toggle="tab" href="#fluxcd" role="tab" aria-controls="fluxcd" aria-selected="false">FluxCD</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

KubeDB can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install, follow the steps below:

```bash
$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

{{< notice type="warning" message="If you are using **private Docker registries** using *self-signed certificates*, please pass the registry domains to the operator like below:" >}}

```bash
$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set global.insecureRegistries[0]=hub.example.com \
  --set global.insecureRegistries[1]=hub2.example.com \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from KubeDB chart and deploy using `kubectl`. Here we are going to show the procedure using Helm 3.

```bash
$ helm template kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt  \
  --set global.skipCleaner=true | kubectl apply -f -
```

{{< notice type="warning" message="If you are using **private Docker registries** using *self-signed certificates*, please pass the registry domains to the operator like below:" >}}

```bash
$ helm template kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt  \
  --set global.insecureRegistries[0]=hub.example.com \
  --set global.insecureRegistries[1]=hub2.example.com \
  --set global.skipCleaner=true | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

</div>
<div class="tab-pane fade" id="argocd" role="tabpanel" aria-labelledby="argocd-tab">

## Using ArgoCD

KubeDB can be deployed via [ArgoCD](https://argo-cd.readthedocs.io/) using the [Helm chart support](https://argo-cd.readthedocs.io/en/stable/user-guide/helm/) for `Application` resources. Deploy the following `Application` manifests in order to your ArgoCD cluster.

### 1. Install `ace-user-roles`

The `ace-user-roles` chart provisions the cluster roles required by KubeDB and related operators. Create the following ArgoCD `Application`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: ace-user-roles
  namespace: argocd
spec:
  destination:
    namespace: kubeops
    server: https://kubernetes.default.svc
  project: default
  source:
    chart: ace-user-roles
    helm:
      values: |
        enableClusterRoles:
          ace: false
          appcatalog: true
          catalog: false
          cert-manager: false
          kubedb: true
          kubedb-ui: false
          kubestash: true # enable if used
          kubevault: true # enable if used
          license-proxyserver: true
          metrics: true
          prometheus: false
          secrets-store: false
          stash: true # enable if used
          virtual-secrets: false
        annotations:
          "helm.sh/hook": null
          "helm.sh/hook-delete-policy": null
    repoURL: ghcr.io/appscode-charts
    targetRevision: v2026.2.16
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true
```

### 2. Install `license-proxyserver`

The `license-proxyserver` chart distributes license tokens to KubeDB and other AppsCode operators inside the cluster. Before applying the manifest below, generate an online license-proxyserver token from `https://appscode.com/billing/{org}/license-proxy-server` and replace the placeholder `token` value with it.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: license-proxyserver
  namespace: argocd
spec:
  project: default
  source:
    chart: license-proxyserver
    repoURL: ghcr.io/appscode-charts
    targetRevision: v2026.2.16
    helm:
      values: |
        platform:
          baseURL: https://appscode.com
          token: '****************************************'
  destination:
    server: "https://kubernetes.default.svc"
    namespace: kubeops
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true

  ignoreDifferences:
  - jsonPointers:
    - /data
    kind: Secret
    name: license-proxyserver-apiserver-cert
    namespace: kubeops
  - group: apiregistration.k8s.io
    kind: APIService
    name: v1alpha1.proxyserver.licenses.appscode.com
    jsonPointers:
    - /spec/caBundle
  - group: apiextensions.k8s.io
    kind: CustomResourceDefinition
    name: servicemonitors.monitoring.coreos.com
    jsonPointers:
    - /metadata/annotations
    - /spec
```

### 3. Install KubeDB

Finally, deploy the KubeDB operators themselves. The `ace-user-roles` sub-chart is disabled here because it was already installed in the first step.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: kubedb
  namespace: argocd
spec:
  project: default
  source:
    chart: kubedb
    repoURL: ghcr.io/appscode-charts
    targetRevision: {{< param "info.version" >}}
    helm:
      values: |
        ace-user-roles:
          enabled: false
  destination:
    server: "https://kubernetes.default.svc"
    namespace: kubedb
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true

  ignoreDifferences:
  - jsonPointers:
    - /data
    kind: Secret
    name: kubedb-kubedb-webhook-server-cert
    namespace: kubedb
  - jsonPointers:
    - /data
    kind: Secret
    name: kubedb-petset-cert
    namespace: kubedb
  - jsonPointers:
    - /data
    kind: Secret
    name: kubedb-sidekick-cert
    namespace: kubedb

  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    name: mutators.petset.appscode.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    name: mutators.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    name: mutators.autoscaling.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    name: mutators.elasticsearch.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    name: mutators.schema.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle

  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: validators.autoscaling.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: validators.elasticsearch.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: validators.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: validators.ops.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: validators.petset.appscode.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: validators.schema.kubedb.com
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle

  - group: apps
    kind: StatefulSet
    name: kubedb-kubedb-autoscaler
    namespace: kubedb
    jsonPointers:
    - /spec/template/metadata/annotations/reload
  - group: apps
    kind: StatefulSet
    name: kubedb-kubedb-ops-manager
    namespace: kubedb
    jsonPointers:
    - /spec/template/metadata/annotations/reload
  - group: apps
    kind: StatefulSet
    name: kubedb-kubedb-provisioner
    namespace: kubedb
    jsonPointers:
    - /spec/template/metadata/annotations/reload
  - group: apps
    kind: Deployment
    name: kubedb-kubedb-webhook-server
    namespace: kubedb
    jsonPointers:
    - /spec/template/metadata/annotations/reload
  - group: apps
    kind: Deployment
    name: kubedb-petset
    namespace: kubedb
    jsonPointers:
    - /spec/template/metadata/annotations/reload
  - group: apps
    kind: Deployment
    name: kubedb-sidekick
    namespace: kubedb
    jsonPointers:
    - /spec/template/metadata/annotations/reload
```

### 4. (Optional) Install KubeStash for Backup

To enable backup and restore for KubeDB-managed databases, deploy the `kubestash` chart via ArgoCD:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: kubestash
  namespace: argocd
spec:
  project: default
  source:
    chart: kubestash
    repoURL: ghcr.io/appscode-charts
    targetRevision: v2026.2.26
    helm:
      values: |
        ace-user-roles:
          enabled: false
  destination:
    server: "https://kubernetes.default.svc"
    namespace: kubestash
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true

  ignoreDifferences:
  - jsonPointers:
    - /data
    kind: Secret
    name: kubestash-kubestash-operator-cert
    namespace: kubestash
  - group: apps
    kind: Deployment
    name: kubestash-kubestash-operator-operator
    namespace: kubestash
    jsonPointers:
    - /spec/template/metadata/annotations/reload
  - group: apps
    kind: Deployment
    name: kubestash-kubestash-operator-webhook-server
    namespace: kubestash
    jsonPointers:
    - /spec/template/metadata/annotations/reload

  - group: admissionregistration.k8s.io
    kind: MutatingWebhookConfiguration
    name: kubestash-kubestash-operator
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
  - group: admissionregistration.k8s.io
    kind: ValidatingWebhookConfiguration
    name: kubestash-kubestash-operator
    jqPathExpressions:
    - .webhooks[].clientConfig.caBundle
```

### 5. (Optional) Install Monitoring Stack

To enable monitoring for KubeDB-managed databases, deploy the following three components in order.

#### 5a. Install `kube-prometheus-stack`

Deploy the upstream [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) chart from the `prometheus-community` repository. The `panopticon` and `monitoring-operator` charts below expect its `ServiceMonitor` selector label `release: kube-prometheus-stack`.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: kube-prometheus-stack
  namespace: argocd
spec:
  project: default
  source:
    chart: kube-prometheus-stack
    repoURL: https://prometheus-community.github.io/helm-charts
    targetRevision: 65.1.1
    helm:
      values: |
        prometheus:
          prometheusSpec:
            serviceMonitorSelectorNilUsesHelmValues: false
            serviceMonitorSelector:
              matchLabels:
                release: kube-prometheus-stack
  destination:
    server: "https://kubernetes.default.svc"
    namespace: monitoring
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true
    - ServerSideApply=true
```

#### 5b. Install `panopticon`

`panopticon` exposes KubeDB-specific Prometheus metrics for databases managed by KubeDB.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: panopticon
  namespace: argocd
spec:
  project: default
  source:
    chart: panopticon
    repoURL: ghcr.io/appscode-charts
    targetRevision: v2026.1.15
    helm:
      values: |
        ace-user-roles:
          enabled: false
        monitoring:
          agent: prometheus.io/operator
          enabled: true
          serviceMonitor:
            labels:
              release: kube-prometheus-stack
  destination:
    server: "https://kubernetes.default.svc"
    namespace: monitoring
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true

  ignoreDifferences:
  - jsonPointers:
    - /data
    kind: Secret
    name: panopticon-apiserver-cert
    namespace: monitoring
  - group: apiregistration.k8s.io
    kind: APIService
    name: v1alpha1.validators.metrics.appscode.com
    jsonPointers:
    - /spec/caBundle
  - group: apps
    kind: Deployment
    name: panopticon
    namespace: monitoring
    jsonPointers:
    - /spec/template/metadata/annotations/checksum~1apiregistration.yaml
  - group: apiextensions.k8s.io
    kind: CustomResourceDefinition
    name: servicemonitors.monitoring.coreos.com
    jsonPointers:
    - /metadata/annotations
    - /spec
```

#### 5c. Install `monitoring-operator`

`monitoring-operator` provisions Grafana dashboards and the OpenViz UI integration for KubeDB databases.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: monitoring-operator
  namespace: argocd
spec:
  project: default
  source:
    chart: monitoring-operator
    repoURL: ghcr.io/appscode-charts
    targetRevision: v2026.1.15
    helm:
      values: |
        monitoring:
          agent: "prometheus.io/operator"
          serviceMonitor:
            labels:
              release: kube-prometheus-stack
  destination:
    server: "https://kubernetes.default.svc"
    namespace: monitoring
  syncPolicy:
    automated: {}
    syncOptions:
    - CreateNamespace=true

  ignoreDifferences:
  - jsonPointers:
    - /data
    kind: Secret
    name: monitoring-operator-apiserver-cert
    namespace: monitoring
  - group: apiregistration.k8s.io
    kind: APIService
    name: v1alpha1.ui.openviz.dev
    jsonPointers:
    - /spec/caBundle
  - group: apps
    kind: Deployment
    name: monitoring-operator
    namespace: monitoring
    jsonPointers:
    - /spec/template/metadata/annotations/checksum~1apiregistration.yaml
```

To see the detailed configuration options for each chart, visit the [AppsCode Charts repository](https://github.com/appscode/charts).

</div>
<div class="tab-pane fade" id="fluxcd" role="tabpanel" aria-labelledby="fluxcd-tab">

## Using FluxCD

KubeDB can be deployed via [FluxCD](https://fluxcd.io/) using its [Helm Controller](https://fluxcd.io/flux/components/helm/) against the OCI Helm registry at `ghcr.io/appscode-charts`. Apply the manifests below in order.

### 1. Configure the OCI Helm repository

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: HelmRepository
metadata:
  name: appscode-charts
  namespace: flux-system
spec:
  type: oci
  interval: 12h
  url: oci://ghcr.io/appscode-charts
```

### 2. Create a Secret with the license

Generate a license from the [AppsCode License Server](https://appscode.com/issue-license?p=kubedb) and store it in a Secret so `HelmRelease` can reference it via `valuesFrom`.

```bash
$ kubectl create namespace kubedb
$ kubectl create secret generic kubedb-license \
  --from-file=license=/path/to/the/license.txt \
  -n kubedb
```

### 3. Install KubeDB via `HelmRelease`

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: kubedb
  namespace: kubedb
spec:
  interval: 1h
  chart:
    spec:
      chart: kubedb
      version: {{< param "info.version" >}}
      sourceRef:
        kind: HelmRepository
        name: appscode-charts
        namespace: flux-system
  install:
    createNamespace: true
    crds: CreateReplace
  upgrade:
    crds: CreateReplace
  valuesFrom:
  - kind: Secret
    name: kubedb-license
    valuesKey: license
    targetPath: global.license
  values:
    global:
      featureGates:
        Elasticsearch: true
        Kafka: true
        MariaDB: true
        MongoDB: true
        MySQL: true
        Postgres: true
        Redis: true
```

If you use a private Docker registry with self-signed certificates, add the registry hosts under `global.insecureRegistries`:

```yaml
  values:
    global:
      insecureRegistries:
      - hub.example.com
      - hub2.example.com
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

</div>
</div>

## Enable Database Engines

KubeDB ships support for many database engines, gated behind individual feature flags so the operator only installs the components you actually need. Toggle an engine on by setting its `global.featureGates.<Engine>` value to `true`. The defaults below mirror the upstream chart — `Elasticsearch`, `Kafka`, `MariaDB`, `MongoDB`, `MySQL`, `Postgres`, and `Redis` are enabled out of the box; every other engine is disabled.

```yaml
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
```

Save these values to a file (e.g. `values.yaml`) and pass it to `helm install` / `helm upgrade`:

```bash
$ helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --values values.yaml \
  --wait --burst-limit=10000 --debug
```

Or override individual engines inline with `--set`:

```bash
$ helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.featureGates.Cassandra=true \
  --set global.featureGates.ClickHouse=true \
  --wait --burst-limit=10000 --debug
```

The same `global.featureGates` map works with the ArgoCD `Application` manifests under the `spec.source.helm.values` block, with the `kubedb-certified` chart on OpenShift, and with the `Kubedb` installer CR used by the OperatorHub bundle.

## Install on OpenShift

KubeDB supports OpenShift in three different ways. Pick the one that best matches your deployment workflow.

<ul class="nav nav-tabs" id="openshiftTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="openshift-standard-tab" data-toggle="tab" href="#openshift-standard" role="tab" aria-controls="openshift-standard" aria-selected="true">Standard Chart</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="openshift-certified-tab" data-toggle="tab" href="#openshift-certified" role="tab" aria-controls="openshift-certified" aria-selected="false">Red Hat Certified Chart</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="openshift-operator-tab" data-toggle="tab" href="#openshift-operator" role="tab" aria-controls="openshift-operator" aria-selected="false">OperatorHub</a>
  </li>
</ul>
<div class="tab-content" id="openshiftTabContent">
  <div class="tab-pane fade show active" id="openshift-standard" role="tabpanel" aria-labelledby="openshift-standard-tab">

### Option A: Standard KubeDB chart with OpenShift overrides

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

</div>
<div class="tab-pane fade" id="openshift-certified" role="tabpanel" aria-labelledby="openshift-certified-tab">

### Option B: Red Hat OpenShift Certified Helm Chart

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

</div>
<div class="tab-pane fade" id="openshift-operator" role="tabpanel" aria-labelledby="openshift-operator-tab">

### Option C: Red Hat OpenShift OperatorHub

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
      MSSQLServer: false
      MariaDB: true
      Memcached: false
      Milvus: false
      MongoDB: true
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

</div>
</div>

## Verify installation

To check if KubeDB operator pods have started, run the following command:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAME                                            READY   STATUS    RESTARTS   AGE
kubedb-kubedb-autoscaler-b5dd47dc5-bxnrq        1/1     Running   0          48s
kubedb-kubedb-ops-manager-6f766b86c6-h9m66      1/1     Running   0          48s
kubedb-kubedb-provisioner-6fd44d5784-d8v9c      1/1     Running   0          48s
kubedb-kubedb-webhook-server-6cf469bdf4-72wvz   1/1     Running   0          48s
```

Once the operator pod is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
```

Now, you are ready to [create your first database](/docs/guides/README.md) using KubeDB.

## Purchase KubeDB License

If you are interested in purchasing KubeDB license, please contact us via sales@appscode.com for further discussion. You can also set up a meeting via our [calendly link](https://calendly.com/appscode/30min).

If you are willing to purchase KubeDB but need more time to test in your dev cluster, feel free to contact sales@appscode.com. We will be happy to extend your trial period.
