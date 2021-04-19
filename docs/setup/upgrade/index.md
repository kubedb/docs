---
title: Upgrade | KubeDB
description: KubeDB Upgrade
menu:
  docs_{{ .version }}:
    identifier: upgrade-kubedb
    name: Upgrade
    parent: setup
    weight: 20
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Upgrading KubeDB

This guide will show you how to upgrade various KubeDB components. Here, we are going to show how to upgrade from an old KubeDB version to the new version, how to migrate between the enterprise edition and community edition, and how to update the license, etc.

## Upgrading KubeDB from `v2021.01.26`(`v0.16.x`) and older to `v2021.03.17`(`v0.17.x`)

In KubeDB `v2021.01.26`(`v0.16.x`) and prior versions, KubeDB used separate charts for KubeDB community edition, KubeDB enterprise edition, and KubeDB catalogs. In KubeDB `v2021.03.17`(`v0.17.x`), we have moved to a single combined chart for all the components for a better user experience. This enables seamless migration between the KubeDB community edition and KubeDB enterprise edition. It also removes the burden of installing individual helm charts manually. KubeDB still depends on [Stash](https://stash.run) as the backup/recovery operator and Stash must be [installed](https://stash.run/docs/latest/setup/) separately. 

In order to upgrade from KubeDB `v2021.01.26`(`v0.16.x`) to `v2021.03.17`(`v0.17.x`), please follow the following steps.

#### 1. Uninstall KubeDB Operator

Uninstall the old KubeDB operator by following the appropriate uninstallation guide of the KubeDB version that you are currently running.

>Make sure you are using the appropriate version of the uninstallation guide. The uninstallation guide for `v2021.03.17`(`v0.17.x`) will not work for `v2021.01.26`(`v0.16.x`) Use the dropdown at the sidebar of the documentation site to navigate to the appropriate version that you are currently running.

#### 2. Update KubeDB Catalog CRDs

KubeDB `v2021.03.17`(`v0.17.x`) has added some new fields in the `***Version` CRDs. Unfortunatley, Helm [does not upgrade the CRDs](https://github.com/helm/helm/issues/6581) bundled in a Helm chart if the CRDs already exist. So, to upgrde the KubeDB catalog CRD, please run the command below:

```bash
kubectl apply -f https://github.com/kubedb/installer/raw/v0.17.1/kubedb-catalog-crds.yaml
```

#### 3. Reinstall new KubeDB Operator

Now, follow the latest installation guide to install the new version of the KubeDB operator. You can find the latest installation guide [here](/docs/setup/README.md). We recommend that you do **not** follow the legacy installation guide, as the new process is much more simpler.

#### 4. Install/Upgrade Stash Operator

Now, upgrade Stash if had previously installed Stash following the instructions [here](https://stash.run/docs/v2021.03.17/setup/upgrade/). If you had not installed Stash before, please install Stash Enterprise Edition following the instructions [here](https://stash.run/docs/v2021.03.17/setup/).


## Migration Between Community Edition and Enterprise Edition

KubeDB `v2021.03.17`(`v0.17.x`) supports seamless migration between community edition and enterprise edition. You can run the following commands to migrate between them.

<ul class="nav nav-tabs" id="migrationTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="mgr-helm3-tab" data-toggle="tab" href="#mgr-helm3" role="tab" aria-controls="mgr-helm3" aria-selected="true">Helm 3</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="mgr-helm2-tab" data-toggle="tab" href="#mgr-helm2" role="tab" aria-controls="mgr-helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="mgr-yaml-tab" data-toggle="tab" href="#mgr-yaml" role="tab" aria-controls="mgr-yaml" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="migrationTabContent">
  <div class="tab-pane fade show active" id="mgr-helm3" role="tabpanel" aria-labelledby="mgr-helm3">

#### Using Helm 3

**From Community Edition to Enterprise Edition:**

In order to migrate from KubeDB community edition to KubeDB enterprise edition, please run the following command,

```bash
helm upgrade kubedb -n kube-system appscode/kubedb \
  --reuse-values \
  --set kubedb-enterprise.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/kubedb-enterprise-license.txt
```

**From Enterprise Edition to Community Edition:**

In order to migrate from KubeDB enterprise edition to KubeDB community edition, please run the following command,

```bash
helm upgrade kubedb -n kube-system appscode/kubedb \
  --reuse-values \
  --set kubedb-enterprise.enabled=false \
  --set kubedb-autoscaler.enabled=false \
  --set kubedb-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/kubedb-community-license.txt
```

</div>
<div class="tab-pane fade" id="mgr-helm2" role="tabpanel" aria-labelledby="mgr-helm2">

**Using Helm 2**

**From Community Edition to Enterprise Edition:**

To migrate from KubeDB community edition to KubeDB enterprise edition, please run the following command,

```bash
helm upgrade kubedb appscode/kubedb \
  --reuse-values \
  --set kubedb-enterprise.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/kubedb-enterprise-license.txt
```

**From Enterprise Edition to Community Edition:**

To migrate from KubeDB enterprise edition to KubeDB community edition, please run the following command,

```bash
helm upgrade kubedb appscode/kubedb \
  --reuse-values \
  --set kubedb-enterprise.enabled=false \
  --set kubedb-autoscaler.enabled=false \
  --set kubedb-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/kubedb-community-license.txt
```

</div>
<div class="tab-pane fade" id="mgr-yaml" role="tabpanel" aria-labelledby="mgr-yaml">

**Using YAML (with helm 3)**

**From Community Edition to Enterprise Edition:**

In order to migrate from KubeDB community edition to KubeDB enterprise edition, please run the following command,

```bash
# Install KubeDB enterprise edition
helm template kubedb -n kube-system appscode/kubedb \
  --version {{< param "info.version" >}} \
  --set kubedb-enterprise.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-catalog.skipDeprecated=false \
  --set global.skipCleaner=true \
  --set-file global.license=/path/to/kubedb-enterprise-license.txt | kubectl apply -f -
```

**From Enterprise Edition to Community Edition:**

In order to migrate from KubeDB enterprise edition to KubeDB community edition, please run the following command,

```bash
# Install KubeDB community edition
helm template kubedb -n kube-system appscode/kubedb \
  --version {{< param "info.version" >}} \
  --set kubedb-enterprise.enabled=false \
  --set kubedb-autoscaler.enabled=false \
  --set kubedb-catalog.skipDeprecated=false \
  --set global.skipCleaner=true \
  --set-file global.license=/path/to/kubedb-community-license.txt | kubectl apply -f -
```

</div>
</div>

## Updating License

KubeDB support updating license without requiring any re-installation. KubeDB creates a Secret named `<helm release name>-license` with the license file. You just need to update the Secret. The changes will propagate automatically to the operator and it will use the updated license going forward.

Follow the below instructions to update the license:

- Get a new license and save it into a file.
- Then, run the following upgrade command based on your installation.

<ul class="nav nav-tabs" id="luTabs" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="lu-helm3-tab" data-toggle="tab" href="#lu-helm3" role="tab" aria-controls="lu-helm3" aria-selected="true">Helm 3</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="lu-helm2-tab" data-toggle="tab" href="#lu-helm2" role="tab" aria-controls="lu-helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="lu-yaml-tab" data-toggle="tab" href="#lu-yaml" role="tab" aria-controls="lu-yaml" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="luTabContent">
  <div class="tab-pane fade show active" id="lu-helm3" role="tabpanel" aria-labelledby="lu-helm3">

#### Using Helm 3

```bash
helm upgrade kubedb -n kube-system appscode/kubedb \
  --reuse-values \
  --set-file global.license=/path/to/new/license.txt
```

</div>
<div class="tab-pane fade" id="lu-helm2" role="tabpanel" aria-labelledby="lu-helm2">

#### Using Helm 2

```bash
helm upgrade kubedb appscode/kubedb \
  --reuse-values \
  --set-file license=/path/to/new/license.txt
```

</div>
<div class="tab-pane fade" id="lu-yaml" role="tabpanel" aria-labelledby="lu-yaml">

#### Using YAML (with helm 3)

**Update License of Community Edition:**

```bash
helm template kubedb -n kube-system appscode/kubedb \
  --set kubedb-enterprise.enabled=false \
  --set kubedb-autoscaler.enabled=false \
  --set global.skipCleaner=true \
  --show-only appscode/kubedb-community/templates/license.yaml \
  --set-file global.license=/path/to/new/license.txt | kubectl apply -f -
```

**Update License of Enterprise Edition:**

```bash
helm template kubedb appscode/kubedb -n kube-system \
  --set kubedb-enterprise.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set global.skipCleaner=true \
  --show-only appscode/kubedb-enterprise/templates/license.yaml \
  --set-file global.license=/path/to/new/license.txt | kubectl apply -f -
```

</div>
</div>
