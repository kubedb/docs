---
title: Install | KubeDB
description: Installation guide for KubeDB
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-readme
    name: Overview
    parent: install-kubedb-enterprise
    weight: 5
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
url: /docs/{{ .version }}/setup/install/kubedb/
aliases:
  - /docs/{{ .version }}/setup/install/kubedb/README/
---

# Install KubeDB

## Get a Free License

Download a FREE license from [AppsCode License Server](https://appscode.com/issue-license?p=kubedb).

> KubeDB licensing process has been designed to work with CI/CD workflow. You can automatically obtain a license from your CI/CD pipeline by following the guide from [here](https://github.com/appscode/offline-license-server#offline-license-server).

## Choose an Installation Method

KubeDB can be installed in several ways. Pick the one that fits your workflow:

- [Helm 3](/docs/setup/install/kubedb/helm.md) — recommended for most users.
- [YAML](/docs/setup/install/kubedb/yaml.md) — render manifests and apply with `kubectl`.
- [ArgoCD](/docs/setup/install/kubedb/argocd.md) — GitOps via ArgoCD `Application` resources.
- [FluxCD](/docs/setup/install/kubedb/fluxcd.md) — GitOps via the Flux Helm Controller.
- [OpenShift](/docs/setup/install/kubedb/openshift.md) — standard chart, Red Hat certified chart, or OperatorHub.

After installing, see [Common Configuration](/docs/setup/install/kubedb/configuration.md) to enable database engines and verify the installation.

## Purchase KubeDB License

If you are interested in purchasing KubeDB license, please contact us via sales@appscode.com for further discussion. You can also set up a meeting via our [calendly link](https://calendly.com/appscode/30min).

If you are willing to purchase KubeDB but need more time to test in your dev cluster, feel free to contact sales@appscode.com. We will be happy to extend your trial period.
