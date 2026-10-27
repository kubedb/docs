---
title: YugabyteDBVersion CRD
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-catalog-concepts
    name: YugabyteDBVersion
    parent: yugabytedb-concepts
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# YugabyteDBVersion

## What is YugabyteDBVersion

`YugabyteDBVersion` (`catalog.kubedb.com/v1alpha1`, short name `ybversion`, cluster scoped) is the catalog CRD that holds the image and version metadata of KubeDB-managed YugabyteDB. The value of `YugabyteDB.spec.version` must be the name of one of these objects. The KubeDB installer creates them; you do not apply them yourself.

## YugabyteDBVersion Specification

> Note: YAML files used in this guide are stored in [docs/examples/yugabytedb/catalog](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/catalog) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: YugabyteDBVersion
metadata:
  name: 2025.2.8
spec:
  version: 2025.2.8
  build: 2025.2.8.0-b60
  db:
    image: docker.io/yugabytedb/yugabyte:2025.2.8.0-b60
  securityContext:
    runAsUser: 10001
  ysqlMajorVersion: 15
  memory:
    masterLimitPercent: 85
    tserverLimitPercent: 70
  defaultFlags:
    master:
      use_memory_defaults_optimized_for_ysql: "true"
      graceful_shutdown: "true"
      enforce_tablet_replica_limits: "true"
      split_respects_tablet_replica_limits: "true"
    tserver:
      use_memory_defaults_optimized_for_ysql: "true"
      graceful_shutdown: "true"
  features:
    clientOnlyCertReload: true
  updateConstraints:
    allowlist:
    - '>= 2025.1.4, < 2025.2.8'
  deprecated: false
```

### Key fields

- `metadata.name` is the value to use in `YugabyteDB.spec.version`.
- `spec.version` is the YugabyteDB version and `spec.build` the upstream build of the image.
- `spec.db.image` is the image used for both the `yb-master` and `yb-tserver` containers.
- `spec.securityContext.runAsUser` is the user the containers run as (`10001`).
- `spec.ysqlMajorVersion` is the PostgreSQL major version that YSQL is built on (`15`). A version update between different YSQL major versions needs a different flow, which is not part of this release.
- `spec.memory` sets what share of the container memory limit is given to `--memory_limit_hard_bytes`: 85% for masters and 70% for tservers (the tserver keeps the rest for the PostgreSQL backends, which run outside its memory tracker). For example, a tserver limit of `2560Mi` gives `1879048192` bytes. Because of this, a [VerticalScaling](/docs/guides/yugabytedb/scaling/vertical-scaling/scale-vertically/) recomputes the flag and restarts the pods.
- `spec.defaultFlags` are written to the flag files before your own configuration.
- `spec.features.clientOnlyCertReload` is true when certificates can be hot reloaded even if only client-to-node TLS is on (2025.2.1 and later).
- `spec.updateConstraints` restricts which versions can be updated from (`allowlist`) or to (`denylist`). With the catalog above you can update from `2025.1.4` to `2025.2.8`.
- `spec.deprecated` and `spec.endOfLife` mark versions that should be avoided.

### Available versions

```bash
$ kubectl get ybversion
NAME       VERSION    DB_IMAGE                                        DEPRECATED   AGE
2025.1.4   2025.1.4   docker.io/yugabytedb/yugabyte:2025.1.4.0-b103   false        5h39m
2025.2.8   2025.2.8   docker.io/yugabytedb/yugabyte:2025.2.8.0-b60    false        5h39m
```

## Next Steps

- Read the [YugabyteDB CRD concept](/docs/guides/yugabytedb/concepts/yugabytedb.md).
- Run the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md).
- Upgrade with [UpdateVersion](/docs/guides/yugabytedb/update-version/versionupgrading/).
