---
title: Run Milvus on OpenShift and in an Air-gapped Cluster
description: Installing KubeDB for Milvus on OpenShift 4.19, mirroring images for a disconnected cluster, and running Milvus pods under the restricted-v2 SCC
menu:
  docs_{{ .version }}:
    identifier: milvus-openshift-guide
    name: Guide
    parent: milvus-openshift
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code, charts and existing install docs; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Run Milvus on OpenShift and in an Air-gapped Cluster

This guide covers what is specific to Milvus when KubeDB runs on OpenShift 4.19 with a private image registry and no outbound internet access. It builds on two existing pages and does not repeat them:

- [Install KubeDB on OpenShift](/docs/setup/install/kubedb/openshift.md): the three install options and the OpenShift distro flags.
- [Offline (Air-gapped) Installation](/docs/setup/install/kubedb/helm.md#offline-air-gapped-installation): mirroring images and rewriting registry paths.

> **Maturity.** Everything here is derived from the installer charts, the operator source and the existing install docs. It has **not** been verified on a live OpenShift cluster. The one point that needs a decision before you start is the Milvus pod user ID, in [Pod user IDs and the `restricted-v2` SCC](#pod-user-ids-and-the-restricted-v2-scc).

## Install KubeDB with only Milvus enabled

Keep the operator footprint and the set of images you must mirror small by enabling only the engines you use. Milvus needs the `Etcd` engine too, because its metadata store is a KubeDB `Etcd`:

```yaml
global:
  distro:
    openshift: true
    ubi: "operator"
  featureGates:
    Milvus: true
    Etcd: true
    # switch the other engines off, for example:
    Elasticsearch: false
    Kafka: false
    MariaDB: false
    MongoDB: false
    MySQL: false
    Redis: false
    Postgres: false
```

The `openshift` flag is also detected automatically when the cluster serves the `project.openshift.io/v1` API. On OpenShift, the operator's own pods are rendered **without** a fixed `runAsUser`, `runAsGroup`, `fsGroup` or `supplementalGroups` so the SCC can assign them. The database pods that KubeDB creates (Milvus, Etcd) are a different matter; see below.

The operator is installed cluster-wide: the provisioner chart ships a `ClusterRole` and `ClusterRoleBinding`, and this guide knows of no chart value that restricts the operator to a set of namespaces. Plan the install on that basis. Only the Milvus databases themselves live in the namespaces you choose.

## Mirror the images for a disconnected cluster

Follow the [Offline (Air-gapped) Installation](/docs/setup/install/kubedb/helm.md#offline-air-gapped-installation) steps, then add the Milvus-specific images below.

### Which images Milvus needs

From the installer's `catalog/imagelist.yaml`, the Milvus-related images at this release are:

| Image | Used for |
|---|---|
| `ghcr.io/appscode-images/milvus:<version>` | The Milvus server, one per catalog version you enable (`2.6.7`, `2.6.9`, `2.6.11`, `3.0.0`) |
| `ghcr.io/appscode-images/etcd:v3.5.21`, `:v3.6.4` | The `Etcd` that holds Milvus metadata |
| `ghcr.io/kubedb/milvus-archiver:v0.2.0` | The continuous change-log archiver ([backup guide](/docs/guides/milvus/backup/guide.md)) |
| `ghcr.io/kubedb/milvus-backup-plugin:v0.1.0-rc.0` | KubeStash logical and physical backup and restore |
| `ghcr.io/kubedb/milvus-csi-snapshotter-plugin:v0.2.0` | CSI snapshots of a Standalone data volume |
| `ghcr.io/kubedb/etcd-restic-plugin:v0.2.0_3.5.21`, `_3.6.4` | Etcd snapshot backup and restore |
| `ghcr.io/kubedb/kubedb-manifest-plugin:<tag>` | `manifest-backup` and `manifest-restore` |

On top of these come the operator images from `catalog/scripts/operator/imagelist.yaml`, which are always required. The Milvus-specific `catalog/scripts/milvus/imagelist.yaml` lists only the Milvus server images and the archiver. It does **not** list the backup and snapshotter plugins, which appear only in the full `catalog/imagelist.yaml`. If you use KubeStash backups, mirror them too.

At this release, `catalog/scripts/milvus` and `catalog/scripts/etcd` contain only an `imagelist.yaml` and no `copy-images.sh`, `export-images.sh` or `import-images.sh`; `catalog/scripts/operator` has all four. For the Milvus and etcd images, copy from the full list. The installer's scripts copy each image to `${IMAGE_REGISTRY}/<original repository path>` without the source host, and this loop, run from a checkout of `kubedb/installer` at your release tag, does the same with [`crane`](https://github.com/google/go-containerregistry):

```bash
export IMAGE_REGISTRY=registry.example.com
for img in $(grep -E 'milvus|etcd|kubedb-manifest-plugin' catalog/imagelist.yaml | sed 's/^- //'); do
  crane copy "$img" "$IMAGE_REGISTRY/${img#*/}"
done
```

For a registry the connected machine cannot reach directly, use `crane pull` and `crane push` with tarballs the way `export-images.sh` and `import-images.sh` do.

### Point the chart and the catalog at your registry

Set `global.registryFQDN` and the `proxies.*` values to your registry host, as in step 3 of the offline guide. For Milvus the server and plugin images come from `ghcr.io`, so the `ghcr` proxy is the one that matters, on both catalogs. The operator's helper image `docker.io/tianon/toybox` comes from Docker Hub, so set `dockerHub` too:

```bash
$ helm upgrade -i kubedb oci://registry.example.com/appscode-charts/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt \
  --values values.yaml \
  --set global.registryFQDN=registry.example.com \
  --set kubedb-catalog.proxies.ghcr=registry.example.com \
  --set kubedb-catalog.proxies.dockerHub=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.ghcr=registry.example.com \
  --set kubedb-kubestash-catalog.proxies.dockerHub=registry.example.com \
  --wait --burst-limit=10000 --debug
```

Run the `helm template ... | grep 'image:'` check from step 2 of the offline guide first. Every image must resolve to your registry, and that output is also the exact list to mirror for your chart version.

If your registry needs credentials, set `global.imagePullSecrets`. If it uses a self-signed certificate, add it to `global.insecureRegistries`. Both are in the `Kubedb` resource shown on the [OpenShift install page](/docs/setup/install/kubedb/openshift.md).

### Pull secrets for the Milvus pods

The operator's pull secret is separate from the one the database pods use. Give the Milvus pods theirs through the pod template:

```yaml
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: registry-credentials
```

In Distributed mode, set it on each role's pod template under `spec.topology.distributed.<role>.podTemplate`.

### The operator reads the registry at reconcile time

When the operator builds a Milvus PetSet it resolves the image tag to a digest by contacting the registry named in the `MilvusVersion` image, and it builds the registry credentials from the pod template's `imagePullSecrets`. If that registry is not reachable from the operator pod, reconciliation fails with `Failed to get image with digest`. In a disconnected cluster that means the **operator** must be able to reach your in-cluster registry, not only the nodes. Make sure the network policy for the operator namespace allows it.

## Pod user IDs and the `restricted-v2` SCC

This is the main thing to settle before you deploy a Milvus on OpenShift.

**Current behavior.** Each `MilvusVersion` and `EtcdVersion` in the catalog sets `spec.securityContext.runAsUser: 1000`. When you create a `Milvus`, the defaulting webhook fills any unset fields with:

- pod `fsGroup` and container `runAsUser` and `runAsGroup` of `1000`,
- `runAsNonRoot: true`, `allowPrivilegeEscalation: false`, all capabilities dropped, and the default seccomp profile.

The `EtcdVersion` entries that back the Etcd KubeDB creates for a Milvus also set `runAsUser: 1000`.

**What that means on OpenShift.** The `restricted-v2` SCC assigns user and group IDs from the range on the namespace and rejects a pod that asks for a fixed ID outside it. A Milvus pod that asks for `runAsUser: 1000` and `fsGroup: 1000` is therefore rejected unless the namespace range happens to include `1000` or you grant an SCC that allows it. The operator's own pods avoid this by dropping those fields on OpenShift; the Milvus and Etcd pods do not.

**Two ways to run, pick one:**

1. **Use IDs inside the namespace range.** Read the range, then set the IDs in the pod template. The defaults only fill fields that are unset, so your values win:

   ```bash
   $ oc get ns demo -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.uid-range}'
   1000680000/10000
   ```

   ```yaml
   spec:
     podTemplate:
       spec:
         securityContext:
           fsGroup: 1000680000
         containers:
           - name: milvus
             securityContext:
               runAsUser: 1000680000
               runAsGroup: 1000680000
   ```

   For a Distributed Milvus, set the same under each role's `podTemplate`. The `Etcd` that KubeDB creates for the Milvus takes its security context from the `EtcdVersion`, so check that its pods are admitted too, and check the `supplementalGroups` range for the group IDs.
2. **Grant an SCC that allows the fixed ID,** for example by binding a custom SCC to the service account the Milvus pods run as. The SCC and RBAC list for each role is something a cluster administrator has to approve; this guide does not ship one.

**Not verified.** That the Milvus image runs correctly as an arbitrary user ID (file ownership and permissions inside the image), and that `restricted-v2` admits the pods with the overrides above. A change to remove the fixed user ID on OpenShift is being worked on separately; it is not available in this release.

## GPU and SR-IOV pods

[GPU scheduling and SR-IOV networking](/docs/guides/milvus/gpu-sriov/guide.md) add requirements on top of the Milvus pod:

- The Milvus container's default security context **drops all capabilities and disallows privilege escalation**. If your NIC or RDMA setup needs a capability (for example to lock memory), you must add it by overriding the container's `securityContext` in the pod template, and the SCC that admits the pod must allow it. This guide does not list the capabilities your hardware needs.
- GPU nodes and SR-IOV virtual functions are requested as extended resources and node selectors in the pod template. They do not change the SCC by themselves.
- The SR-IOV advertise-IP init container uses the Milvus image and needs the `ip` command. It also runs under the SCC, so it needs the same user-ID decision as the main container.
- The `NetworkAttachmentDefinition` and the SR-IOV and GPU operators are cluster-administrator prerequisites. KubeDB does not create them.

Because capability and SCC needs depend on the hardware and drivers, treat the privilege list for GPU and SR-IOV roles as something to determine on a test cluster and approve explicitly. This guide has not determined it.

## Operating with outbound traffic denied

With egress blocked from the Milvus namespace, check these paths:

- **Object storage.** Milvus and the backup plugins talk to your S3-compatible store. It must be reachable from the Milvus pods and from the backup job pods. Keep backups in a separate bucket with separate credentials.
- **The registry**, for the operator's digest lookup, as above.
- **KubeStash and the backup plugins** run as jobs and need the plugin images from your registry and your object store, nothing else.
- **The license.** KubeDB takes its license with `--set-file global.license=`. See the [offline license server](https://github.com/appscode/offline-license-server) mentioned on the install pages for obtaining one without internet access. Whether the operator needs any outbound call at run time beyond that is not covered here.
- **Network policy.** If you enable `global.networkPolicy`, the generated policies allow the in-cluster paths listed under [Required network communication](/docs/setup/install/kubedb/configuration.md#required-network-communication). If you write your own, permit those same paths.

## What is not covered

- A tested SCC and RBAC list for the operator and each Milvus role.
- The Milvus image under an arbitrary user ID.
- Hosted control plane specifics. Nothing in the charts distinguishes hosted control planes from standard OpenShift.
- Verification on a live OpenShift 4.19 cluster.

## Related

- [Install KubeDB on OpenShift](/docs/setup/install/kubedb/openshift.md)
- [Offline (Air-gapped) Installation](/docs/setup/install/kubedb/helm.md#offline-air-gapped-installation)
- [GPU and SR-IOV Networking for Milvus](/docs/guides/milvus/gpu-sriov/guide.md)
- [Backup & Restore Milvus](/docs/guides/milvus/backup/guide.md)
