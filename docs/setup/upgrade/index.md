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

This guide will show you how to upgrade various KubeDB components. Here, we are going to show how to upgrade from an old KubeDB version to the new version, and how to update the license, etc.

## Upgrading KubeDB from `v2021.xx.xx` to `{{< param "info.version" >}}`

In order to upgrade from KubeDB `v2021.xx.xx` to `{{< param "info.version" >}}`, please follow the following steps.

{{< notice type="warning" message="Please note that since v2021.08.23, we recommend installing KubeDB operator in the `kubedb` namespace. The upgrade instructions on this page assumes so. If you have currently installed the operator in a different namespace like `kube-system`, either follow the instructions with appropriate updates, or first uninstall the existing operator and then reinstall in the `kubedb` namespace." >}}

#### 1. Update KubeDB Catalog CRDs

Helm [does not upgrade the CRDs](https://github.com/helm/helm/issues/6581) bundled in a Helm chart if the CRDs already exist. So, to upgrade the KubeDB catalog CRD, please run the command below:

```bash
kubectl apply -f https://github.com/kubedb/installer/raw/{{< param "info.version" >}}/crds/kubedb-catalog-crds.yaml

kubectl apply -f https://github.com/kubeops/petset/raw/v0.0.6/crds/apps.k8s.appscode.com_petsets.yaml
kubectl apply -f https://github.com/kubeops/petset/raw/v0.0.6/crds/apps.k8s.appscode.com_placementpolicies.yaml
```

#### 2. Upgrade KubeDB Operator

Now, upgrade the KubeDB helm chart using the following command. You can find the latest installation guide [here](/docs/setup/README.md). We recommend that you do **not** follow the legacy installation guide, as the new process is much more simpler.

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set kubedb-catalog.skipDeprecated=false \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

{{< notice type="warning" message="If you are using **private Docker registries** using *self-signed certificates*, please pass the registry domains to the operator like below:" >}}

```bash
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set kubedb-catalog.skipDeprecated=false \
  --set global.insecureRegistries[0]=hub.example.com \
  --set global.insecureRegistries[1]=hub2.example.com \
  --set-file global.license=/path/to/the/license.txt \
  --wait --burst-limit=10000 --debug
```

#### 3. Install/Upgrade Stash Operator

Now, upgrade Stash if had previously installed Stash following the instructions [here](https://stash.run/docs/v2021.06.23/setup/upgrade/). If you had not installed Stash before, please install Stash following the instructions [here](https://stash.run/docs/v2021.06.23/setup/).


## Upgrading KubeDB from `v2021.01.26`(`v0.16.x`) and older to `{{< param "info.version" >}}`

In KubeDB `v2021.01.26`(`v0.16.x`) and prior versions, KubeDB used separate charts for KubeDB community edition, KubeDB enterprise edition, and KubeDB catalogs. In KubeDB `{{< param "info.version" >}}`, we have moved to a single combined chart for all the components for a better user experience. KubeDB still depends on [Stash](https://stash.run) as the backup/recovery operator and Stash must be [installed](https://stash.run/docs/latest/setup/) separately. 

In order to upgrade from KubeDB `v2021.01.26`(`v0.16.x`) to `{{< param "info.version" >}}`, please follow the following steps.

#### 1. Uninstall KubeDB Operator

Uninstall the old KubeDB operator by following the appropriate uninstallation guide of the KubeDB version that you are currently running.

> Make sure you are using the appropriate version of the uninstallation guide. Use the dropdown at the sidebar of the documentation site to navigate to the appropriate version that you are currently running.

#### 2. Update KubeDB Catalog CRDs

Helm [does not upgrade the CRDs](https://github.com/helm/helm/issues/6581) bundled in a Helm chart if the CRDs already exist. So, to upgrde the KubeDB catalog CRD, please run the command below:

```bash
kubectl apply -f https://github.com/kubedb/installer/raw/{{< param "info.version" >}}/crds/kubedb-catalog-crds.yaml
```

#### 3. Reinstall new KubeDB Operator

Now, follow the latest installation guide to install the new version of the KubeDB operator. You can find the latest installation guide [here](/docs/setup/README.md). We recommend that you do **not** follow the legacy installation guide, as the new process is much more simpler.

#### 4. Install/Upgrade Stash Operator

Now, upgrade Stash if had previously installed Stash following the instructions [here](https://stash.run/docs/latest/setup/upgrade/). If you had not installed Stash before, please install Stash following the instructions [here](https://stash.run/docs/latest/setup/).


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
    <a class="nav-link" id="lu-yaml-tab" data-toggle="tab" href="#lu-yaml" role="tab" aria-controls="lu-yaml" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="luTabContent">
  <div class="tab-pane fade show active" id="lu-helm3" role="tabpanel" aria-labelledby="lu-helm3">

#### Using Helm 3

```bash
# detect current version
helm ls -A | grep kubedb

# update license key keeping the current version
helm upgrade -i kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version=<cur_version> \
  --reuse-values \
  --set-file global.license=/path/to/new/license.txt \
  --wait --burst-limit=10000 --debug
```

</div>
<div class="tab-pane fade" id="lu-yaml" role="tabpanel" aria-labelledby="lu-yaml">

#### Using YAML (with helm 3)

```bash
# detect current version
helm ls -A | grep kubedb

# update license key keeping the current version
helm template kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version=<cur_version> \
  --namespace kubedb --create-namespace \
  --set global.skipCleaner=true \
  --show-only appscode/kubedb/templates/license.yaml \
  --set-file global.license=/path/to/new/license.txt | kubectl apply -f -
```

</div>
</div>
