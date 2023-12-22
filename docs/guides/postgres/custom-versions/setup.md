---
title: Setup Custom PostgresVersions
menu:
  docs_{{ .version }}:
    identifier: pg-custom-versions-setup-postgres
    name: Overview
    parent: pg-custom-versions-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

## Setting up Custom PostgresVersions

PostgresVersions are KubeDB crds that define the docker images KubeDB will use when deploying a postgres database. For more details about PostgresVersion crd, please visit [here](/docs/guides/postgres/concepts/catalog.md).

## Creating a Custom Postgres Database Image for KubeDB

The best way to create a custom image is to build on top of the existing kubedb image.

```docker
FROM kubedb/postgres:10.2-v3

ENV TIMESCALEDB_VERSION 0.9.1

RUN set -ex \
    && apk add --no-cache --virtual .fetch-deps \
    ca-certificates \
    openssl \
    tar \
    && mkdir -p /build/timescaledb \
    && wget -O /timescaledb.tar.gz https://github.com/timescale/timescaledb/archive/$TIMESCALEDB_VERSION.tar.gz \
    && tar -C /build/timescaledb --strip-components 1 -zxf /timescaledb.tar.gz \
    && rm -f /timescaledb.tar.gz \
    \
    && apk add --no-cache --virtual .build-deps \
    coreutils \
    dpkg-dev dpkg \
    gcc \
    libc-dev \
    make \
    cmake \
    util-linux-dev \
    \
    && cd /build/timescaledb \
    && ./bootstrap \
    && cd build && make install \
    && cd ~ \
    \
    && apk del .fetch-deps .build-deps \
    && rm -rf /build

RUN sed -r -i "s/[#]*\s*(shared_preload_libraries)\s*=\s*'(.*)'/\1 = 'timescaledb,\2'/;s/,'/'/" /scripts/primary/postgresql.conf
```

From there, we would define a PostgresVersion that contains this new image. Let's say we tagged it as `myco/postgres:timescale-0.9.1`

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: PostgresVersion
metadata:
  name: timescaledb-2.1.0-pg13
spec:
  coordinator:
    image: kubedb/pg-coordinator:v0.8.0
  db:
    image: timescale/timescaledb:2.1.0-pg13-oss
  distribution: TimescaleDB
  exporter:
    image: prometheuscommunity/postgres-exporter:v0.9.0
  initContainer:
    image: kubedb/postgres-init:0.4.0
  podSecurityPolicies:
    databasePolicyName: postgres-db
  securityContext:
    runAsAnyNonRoot: true
    runAsUser: 70
  stash:
    addon:
      backupTask:
        name: postgres-backup-13.1
      restoreTask:
        name: postgres-restore-13.1
  version: "13.13"
```

Once we add this PostgresVersion we can use it in a new Postgres like:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: timescale-postgres
  namespace: demo
spec:
  version: "timescaledb-2.1.0-pg13" # points to the name of our custom PostgresVersion
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```
