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

## Offline (Air-gapped) Installation

In an air-gapped or offline cluster, the nodes cannot reach the public registries that KubeDB pulls from (`ghcr.io`, `quay.io`, and `registry.k8s.io`, plus `docker.io`, `mcr.microsoft.com`, `r.appscode.com`, `container-registry.oracle.com`, and `cr.weaviate.io` for some database engines). To install KubeDB in such an environment, mirror the Helm chart and the required images into a registry your cluster can reach, then point the operator at that registry using `global.registryFQDN` and the `proxies.*` values of the catalog sub-charts.

The examples below use `registry.example.com` as the private registry FQDN. Replace it with your own (add `:port` if your registry runs on a non-standard port).

### 1. Mirror the required images

KubeDB and its catalog reference images from the following namespaces. Mirror them into your private registry, or configure your registry as a pull-through cache for them:

- `ghcr.io/appscode`
- `ghcr.io/appscode-images`
- `ghcr.io/appscode-charts` (the Helm charts)
- `ghcr.io/kubedb` (operator components and backup or ops plugins)
- `registry.k8s.io/git-sync`

Depending on the engines you enable, database images also come from `quay.io`, `docker.io` (including `docker.io/library` for helpers such as `busybox`), `mcr.microsoft.com`, `container-registry.oracle.com`, and `cr.weaviate.io`. Mirror the registries that the engines you plan to enable actually use. The verification command in step 2 prints the exact set for your configuration, and the installer publishes ready-made image lists you can pull from directly (see step 4).

There are two common registry setups, and they affect how much you need to mirror:

- **Pull-through cache:** the registry fetches and caches each image on first request. You do not need to enumerate images in advance, and it only stores what you actually use.
- **Pre-seeded mirror:** you copy every required image into the registry ahead of time. Use the minimal image list in step 4 to keep the set small.

The installer repository ships helper scripts under [`catalog`](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/catalog) that copy every required image into your registry. They download [`crane`](https://github.com/google/go-containerregistry) automatically and copy each image to `${IMAGE_REGISTRY}/<original-repository-path>`, preserving the upstream path (the `docker.io/library` prefix is dropped). Choose the workflow that matches your network.

If the machine running the script can reach both the public registries and your private registry, copy everything directly:

```bash
wget https://github.com/kubedb/installer/raw/{{< param "info.installer" >}}/catalog/copy-images.sh
chmod +x copy-images.sh
export IMAGE_REGISTRY=registry.example.com
./copy-images.sh
```

If the cluster is fully air-gapped, run the export script on an internet-connected machine to produce a single `images.tar.gz`, carry it across (for example on removable media), then import it on the inside:

```bash
# On an internet-connected machine
wget https://github.com/kubedb/installer/raw/{{< param "info.installer" >}}/catalog/export-images.sh
chmod +x export-images.sh
./export-images.sh   # produces images.tar.gz

# Inside the air-gapped network, after copying images.tar.gz across
wget https://github.com/kubedb/installer/raw/{{< param "info.installer" >}}/catalog/import-images.sh
chmod +x import-images.sh
export IMAGE_REGISTRY=registry.example.com
./import-images.sh images.tar.gz
```

The same four scripts exist per component under `catalog/scripts/operator` (always required) and `catalog/scripts/<engine>` (one directory per database engine), so you can mirror only what you need by running the operator scripts plus the scripts for the engines you enable (see step 4).

{{< notice type="warning" message="On a k3s cluster with no private registry, run `import-into-k3s.sh images.tar.gz` instead. It loads the image tarballs straight into each node's containerd under their original names, so you skip the registry entirely and do not set `global.registryFQDN` or the `proxies.*` values." >}}

Because the scripts copy every image under your registry host with its original repository path, set `global.registryFQDN` and all `proxies.*` values to that same `IMAGE_REGISTRY` host, as shown in steps 2 and 3.

### 2. Verify the image paths are rewritten

Before installing, render the chart and confirm that every `image:` points at your private registry:

```bash
$ helm template kubedb oci://registry.example.com/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set global.registryFQDN=registry.example.com \
  --set kubedb-catalog.proxies.ghcr=registry.example.com \
  --set kubedb-catalog.proxies.quay=registry.example.com \
  --set kubedb-catalog.proxies.kubernetes=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.ghcr=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.quay=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.kubernetes=registry.example.com \
  | grep 'image:' | sort | uniq
```

Every image in the output should be prefixed with your registry FQDN. This de-duplicated list is also the precise set of images you need to make available in your registry for the chart version and engines you selected. Any line that still points at a public registry (for example `ghcr.io/...` or `docker.io/...`) means you need to set the matching proxy before installing.

The catalog sub-charts expose one proxy value per upstream registry. Set the ones your enabled engines actually use:

- `proxies.ghcr` for `ghcr.io` (the KubeDB operator, catalog, and most database images)
- `proxies.dockerHub` for `docker.io` (for example Elasticsearch, MariaDB with MaxScale, ClickHouse, Neo4j, and Percona based MongoDB)
- `proxies.dockerLibrary` for `docker.io/library` (helpers such as `busybox`)
- `proxies.quay` for `quay.io`
- `proxies.kubernetes` for `registry.k8s.io` (for example `git-sync`)
- `proxies.microsoft` for `mcr.microsoft.com` (SQL Server)
- `proxies.appscode` for `r.appscode.com`
- `proxies.oracle` for `container-registry.oracle.com` (Oracle)
- `proxies.weaviate` for `cr.weaviate.io` (Weaviate)

Set the same proxies on both `kubedb-catalog` and `kubedb-kubestash-catalog`, and point each one at wherever you mirrored that upstream. If you populated the registry with the installer scripts from step 1 (or any mirror that keeps each image at its original repository path under one host), use the bare host for every proxy, which is the same `IMAGE_REGISTRY` value (`registry.example.com`). If your registry instead separates each upstream under its own path, use the matching path, for example `--set kubedb-catalog.proxies.ghcr=registry.example.com/ghcr` and `--set kubedb-catalog.proxies.quay=registry.example.com/quay`.

### 3. Install KubeDB from the private registry

Once the chart and images are mirrored, install the operator. The `global.registryFQDN` value rewrites the operator image paths, and the `proxies.*` values rewrite the catalog (database) image paths:

```bash
$ helm upgrade -i kubedb oci://registry.example.com/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --set global.registryFQDN=registry.example.com \
  --set kubedb-catalog.proxies.ghcr=registry.example.com \
  --set kubedb-catalog.proxies.quay=registry.example.com \
  --set kubedb-catalog.proxies.kubernetes=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.ghcr=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.quay=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.kubernetes=registry.example.com \
  --wait --burst-limit=10000 --debug
```

{{< notice type="warning" message="If your private registry uses a *self-signed certificate*, also mark it insecure by adding `--set global.insecureRegistries[0]=registry.example.com`, as shown near the top of this page." >}}

### 4. Minimal image list for pre-seeded registries

If your registry cannot mirror entire namespaces (for example, because of storage limits), mirror only the images you actually need. The [installer repository](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/catalog) publishes image lists, generated from the same charts and split so you can pull only what you need:

- `catalog/imagelist.yaml`: every operator and database image for the release.
- `catalog/scripts/operator/imagelist.yaml`: the operator components (always required).
- `catalog/scripts/<engine>/imagelist.yaml`: one file per engine (for example `mariadb`, `postgres`, `mongodb`).

For a minimal mirror, take `catalog/scripts/operator/imagelist.yaml` plus the per-engine files for the engines you enable. Each of these directories also ships the matching `copy-images.sh`, `export-images.sh`, and `import-images.sh` scripts (see step 1), so you can copy exactly that subset.

The operator (always required) images are published under `ghcr.io/appscode` and `ghcr.io/kubedb`:

```
ghcr.io/appscode/kubectl-nonroot
ghcr.io/appscode/petset
ghcr.io/appscode/sidekick
ghcr.io/kubedb/kubedb-crd-manager
ghcr.io/kubedb/kubedb-provisioner
ghcr.io/kubedb/kubedb-ops-manager
ghcr.io/kubedb/kubedb-autoscaler
ghcr.io/kubedb/kubedb-webhook-server
```

The operator also pulls a small helper image from Docker Hub (`docker.io/tianon/toybox`), and backup or ops features add plugin images such as `<engine>-archiver`, `<engine>-coordinator`, `<engine>-csi-snapshotter-plugin`, and `<engine>-restic-plugin`. Each enabled engine adds its server image (for example `ghcr.io/appscode-images/<engine>`) for every version you turn on, and some engines pull upstream images from Docker Hub, Quay, Oracle, or Weaviate (for example `docker.io/mariadb/maxscale` for MariaDB with MaxScale). The per-engine image lists enumerate all of these.

To keep the set as small as possible, enable only the engines you need through `global.featureGates.<Engine>` (see [Common Configuration](/docs/setup/install/kubedb/configuration.md)), then re-run the step 2 command (or pull the matching per-engine lists) to capture the exact image tags.

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

Next: [enable database engines and verify the installation](/docs/setup/install/kubedb/configuration.md).
