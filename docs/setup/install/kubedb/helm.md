---
title: Install KubeDB using Helm 3
description: Install KubeDB using Helm 3
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-helm
    name: Helm 3
    parent: install-kubedb-enterprise
    weight: 10
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Using Helm 3

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

## (Alternative) Use License Proxyserver instead of a license file

Instead of passing a license file to every operator, you can install the `license-proxyserver` chart once. It distributes license tokens to KubeDB and other AppsCode operators inside the cluster, so the `helm install kubedb` command no longer needs `--set-file global.license`.

Generate an online license-proxyserver token by following the [License Proxyserver guide](https://kubedb.com/docs/platform/v2026.5.22/guides/license-management/license-proxyserver/), then install the chart with that token:

```bash
$ helm install license-proxyserver oci://ghcr.io/appscode-charts/license-proxyserver \
  --version v2026.2.16 \
  --namespace kubeops --create-namespace \
  --set platform.baseURL=https://appscode.com \
  --set platform.token=<your-token> \
  --wait --burst-limit=10000 --debug
```

With the proxyserver running, install KubeDB without the license flag:

```bash
$ helm install kubedb oci://ghcr.io/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --wait --burst-limit=10000 --debug
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

Next: [enable database engines and verify the installation](/docs/setup/install/kubedb/configuration.md).
