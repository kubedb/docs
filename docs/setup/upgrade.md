---
title: Upgrade | KubeDB
description: KubeDB Upgrade
menu:
  docs_{{ .version }}:
    identifier: upgrade-kubedb
    name: Upgrade
    parent: setup
    weight: 30
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Upgrading KubeDB

This guide will show you how to upgrade KubeDB operator. Here, we are going to show how to update the license and how to upgrade between two KubeDB versions.

## Updating License

KubeDB support updating license without requiring any re-installation or restart. KubeDB creates a Secret named `<helm release name>-license` with the license file. You just need to update the Secret. The changes will propagate automatically to the operator and it will use the updated license going forward.

Follow the below instructions to update the license:

- Get a new license and save it into a file.
- Then, run the following upgrade command based on your installation.

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm2-tab" data-toggle="tab" href="#helm2" role="tab" aria-controls="helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

```bash
$ helm upgrade kubedb-enterprise --namespace kube-system appscode/kubedb \
    --reuse-values                                                       \
    --set-file global.license=/path/to/new/license.txt
```

</div>
<div class="tab-pane fade" id="helm2" role="tabpanel" aria-labelledby="helm2-tab">

## Using Helm 2

```bash
$ helm upgrade kubedb-enterprise appscode/kubedb  --namespace kube-system \
    --reuse-values                                        \
    --set-file global.license=/path/to/new/license.txt
```

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML (with helm 3)

```bash
$ helm template kubedb-enterprise appscode/kubedb           \
    --namespace kube-system                                 \
    --set-file global.license=/path/to/new/license.txt      \
    --show-only templates/license.yaml                      \
    --no-hooks | kubectl apply -f -
```

</div>
</div>

## Upgrading Between Community Edition and Enterprise Edition

KubeDB uses two different binaries for Community edition and Enterprise edition. So, it is not possible to upgrade between the Community edition and Enterprise edition without re-installation. However, it is possible to re-install KubeDB without losing the existing backup resources.

Follow the below instructions to re-install KubeDB:

- Uninstall the old version by following the respective uninstallation guide. Don't delete the CRDs.
- Install the new version by following the respective installation guide.

## Upgrading between patch versions

If you are upgrading KubeDB to a patch release, please reapply the [installation instructions](/docs/setup/README.md). That will upgrade the operator pod to the new version and fix any RBAC issues.
