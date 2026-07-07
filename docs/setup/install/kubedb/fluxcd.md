---
title: Install KubeDB using FluxCD
description: Install KubeDB using FluxCD
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-fluxcd
    name: FluxCD
    parent: install-kubedb-enterprise
    weight: 40
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Using FluxCD

KubeDB can be deployed via [FluxCD](https://fluxcd.io/) using its [Helm Controller](https://fluxcd.io/flux/components/helm/) against the OCI Helm registry at `ghcr.io/appscode-charts`. Apply the manifests below in order.

## 1. Configure the OCI Helm repository

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

## 2. Create a Secret with the license

Generate a license from the [AppsCode License Server](https://appscode.com/issue-license?p=kubedb) and store it in a Secret so `HelmRelease` can reference it via `valuesFrom`.

```bash
kubectl create namespace kubedb
```

```bash
kubectl create secret generic kubedb-license \
  --from-file=license=/path/to/the/license.txt \
  -n kubedb
```

## 3. Install KubeDB via `HelmRelease`

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

## (Alternative) Use License Proxyserver instead of a license Secret

Instead of creating a per-cluster license Secret (steps 2–3 above), you can deploy the `license-proxyserver` chart. It distributes license tokens to KubeDB and other AppsCode operators inside the cluster, so the `kubedb` `HelmRelease` no longer needs to mount a license. Use this approach in place of the `kubedb-license` Secret.

### a. Install `license-proxyserver`

Generate an online license-proxyserver token by following the [License Proxyserver guide](https://kubedb.com/docs/platform/v2026.5.22/guides/license-management/license-proxyserver/), then store it in a Secret that the `HelmRelease` references via `valuesFrom`:

```bash
cat > license-proxyserver.yaml <<'EOF'
platform:
  baseURL: https://appscode.com
  token: '****************************************'
EOF
```

```bash
kubectl create secret generic ace-licenseserver-cred \
  --from-file=license-proxyserver.yaml \
  -n kubeops
```

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2
kind: HelmRelease
metadata:
  name: license-proxyserver
  namespace: kubeops
spec:
  interval: 5m
  chart:
    spec:
      chart: license-proxyserver
      version: v2026.2.16
      sourceRef:
        kind: HelmRepository
        name: appscode-charts
        namespace: flux-system
  install:
    createNamespace: true
    crds: CreateReplace
  upgrade:
    crds: CreateReplace
  values:
    registryFQDN: ghcr.io
  valuesFrom:
  - kind: Secret
    name: ace-licenseserver-cred
    valuesKey: license-proxyserver.yaml
    optional: true
```

### b. Install KubeDB without a license Secret

With the proxyserver running, deploy the `kubedb` `HelmRelease` without the `valuesFrom` license Secret — drop step 2 and the `valuesFrom` block from step 3:

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

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

Next: [enable database engines and verify the installation](/docs/setup/install/kubedb/configuration.md).
