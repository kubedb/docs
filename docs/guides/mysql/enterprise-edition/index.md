---
title: Run MySQL Enterprise Edition
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-enterprise-edition
    name: Enterprise Edition
    parent: guides-mysql
    weight: 71
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy MySQL Enterprise Edition with KubeDB

Oracle's MySQL Enterprise Edition ships the same `mysqld` server as MySQL Community
Edition, built with additional commercial plugins (Enterprise Audit, Enterprise
Firewall, Enterprise Encryption/Keyring, Enterprise Thread Pool, Enterprise Backup).
KubeDB's MySQL operator runs it exactly like any other MySQL image; the only
difference is where the image comes from and how you're allowed to pull it.

## Does Enterprise Edition need a license key or Secret at runtime?

**No.** `mysqld` itself does not check a license key at startup, and there is no
separate license activation API to call. Licensing is enforced entirely at the
**image registry** level:

- MySQL Enterprise Edition is distributed only through the **Oracle Container
  Registry** (`container-registry.oracle.com`), under a Bring-Your-Own-License (BYOL)
  model tied to an Oracle account with an active MySQL Enterprise support
  subscription.
- Once you can pull the image, it runs and behaves like any other MySQL image.
  Enterprise plugins are enabled the normal MySQL way (`plugin-load-add`,
  `INSTALL PLUGIN`, or the relevant `my.cnf` settings) after the process is up.

So the only new requirement is a standard Kubernetes `imagePullSecret`, the same
mechanism KubeDB already documents for any [private registry](/docs/guides/mysql/private-registry/index.md).

## Before You Begin

- Read [concept of MySQL Version Catalog](/docs/guides/mysql/concepts/catalog/index.md) to learn the details of the `MySQLVersion` object.
- You need a Kubernetes cluster, with `kubectl` configured to talk to it.
- You need an Oracle account with a MySQL Enterprise Edition subscription. Visit
  [container-registry.oracle.com](https://container-registry.oracle.com/), choose
  **MySQL**, select **enterprise-server**, and accept the license agreement before
  you can pull the image.

## 1. Authenticate to the Oracle Container Registry

```bash
$ docker login container-registry.oracle.com
Username: <your Oracle account email>
Password: <your Oracle account password>
```

## 2. Create an ImagePullSecret

Create a namespace and an `imagePullSecret` from the same Oracle credentials,
following the exact same pattern as KubeDB's
[private registry guide](/docs/guides/mysql/private-registry/index.md#create-imagepullsecret):

```bash
$ kubectl create ns demo

$ kubectl create secret docker-registry -n demo oracle-ocr \
  --docker-server=container-registry.oracle.com \
  --docker-username=<your Oracle account email> \
  --docker-email=<your Oracle account email> \
  --docker-password=<your Oracle account password>
secret/oracle-ocr created
```

## 3. Add a MySQLVersion catalog entry

KubeDB ships a catalog entry for MySQL Enterprise Edition (`8.4.8-oracle`), sourced
from `container-registry.oracle.com/mysql/enterprise-server`. If your KubeDB
installation doesn't already have it (check with
`kubectl get mysqlversion 8.4.8-oracle`), create it:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MySQLVersion
metadata:
  name: 8.4.8-oracle
spec:
  version: 8.4.8
  distribution: MySQL
  db:
    image: container-registry.oracle.com/mysql/enterprise-server:8.4.8
  # coordinator, exporter, initContainer, router, archiver, stash, etc. are the
  # same sidecar images KubeDB uses for every other MySQL 8.4.8 catalog entry.
  securityContext:
    runAsUser: 999
```

> The `distribution` field uses the same `MySQL` value KubeDB already uses for other
> catalog entries; it does not need a separate Enterprise-specific value. What makes
> this an Enterprise Edition deployment is only the `db.image` pointing at the OCR
> image.

## 4. Deploy MySQL using the Enterprise image

Reference the `oracle-ocr` secret via `spec.podTemplate.spec.imagePullSecrets`:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql-enterprise
  namespace: demo
spec:
  version: "8.4.8-oracle"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      imagePullSecrets:
      - name: oracle-ocr
```

```bash
$ kubectl apply -f mysql-enterprise.yaml
mysql.kubedb.com/mysql-enterprise created
```

Confirm the pod pulled the Enterprise image and is running:

```bash
$ kubectl get pods -n demo
NAME                 READY   STATUS    RESTARTS   AGE
mysql-enterprise-0   1/1     Running   0          64s
```

## Cleaning up

```bash
kubectl patch -n demo mysql/mysql-enterprise -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete -n demo mysql/mysql-enterprise
kubectl delete ns demo
```

## Known limitations

The exact Oracle Container Registry image path and tag
(`container-registry.oracle.com/mysql/enterprise-server:8.4.8`) documented here
follows Oracle's published registry convention but has not been pulled and
tested end-to-end against a live cluster as part of this guide. Verify the path
and the image's `runAsUser` against your own OCR access before relying on it in
production.

## Next Steps

- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Learn about running MySQL from any [private Docker registry](/docs/guides/mysql/private-registry/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
