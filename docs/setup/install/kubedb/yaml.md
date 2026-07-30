---
title: Install KubeDB using YAML
description: Install KubeDB using YAML
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-yaml
    name: YAML
    parent: install-kubedb-enterprise
    weight: 20
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Using YAML

If you prefer to not use Helm, you can generate YAMLs from KubeDB chart and deploy using `kubectl`. Here we are going to show the procedure using Helm 3.

## Get a Free License

Download a FREE license from [AppsCode License Server](https://appscode.com/issue-license?p=kubedb) before you begin. You can also automate this from your CI/CD pipeline using the [offline license server](https://github.com/appscode/offline-license-server#offline-license-server).

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

Next: [enable database engines and verify the installation](/docs/setup/install/kubedb/configuration.md).
