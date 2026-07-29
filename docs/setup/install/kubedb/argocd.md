---
title: Install KubeDB using ArgoCD
description: Install KubeDB using ArgoCD
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-argocd
    name: ArgoCD
    parent: install-kubedb-enterprise
    weight: 30
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Using ArgoCD

KubeDB can be deployed via [ArgoCD](https://argo-cd.readthedocs.io/) using the [Helm chart support](https://argo-cd.readthedocs.io/en/stable/user-guide/helm/) for `Application` resources. Deploy the following `Application` manifests in order to your ArgoCD cluster.

Ready-to-use `Application` manifests for KubeDB and the rest of the AppsCode stack (e.g. `kubestash`, `kubevault`, `stash`, `panopticon`, `monitoring-operator`) are maintained in the [appscode/gitops](https://github.com/appscode/gitops/tree/2025-06/argocd/helm) repository. Install `ace-user-roles` and `license-proxyserver` first, then pick whichever component manifests you need from there.

## 1. Install `ace-user-roles`

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

## 2. Install `license-proxyserver`

The `license-proxyserver` chart distributes license tokens to KubeDB and other AppsCode operators inside the cluster. Before applying the manifest below, generate an online license-proxyserver token by following the [License Proxyserver guide](https://kubedb.com/docs/platform/v2026.5.22/guides/license-management/license-proxyserver/) and replace the placeholder `token` value with it.

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
```

## 3. Install KubeDB

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

To see the detailed configuration options for each chart, visit the [AppsCode Charts repository](https://github.com/appscode/charts).

Next: [enable database engines and verify the installation](/docs/setup/install/kubedb/configuration.md).
