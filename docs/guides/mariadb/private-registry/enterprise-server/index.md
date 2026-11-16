---
title: Run MariaDB Enterprise Server
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-privateregistry-enterprise
    name: Enterprise Server
    parent: guides-mariadb-privateregistry
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Deploy MariaDB Enterprise Server with KubeDB

[MariaDB Enterprise Server](https://mariadb.com/docs/server/deploy/deployment-methods/docker/enterprise-server/) is MariaDB Corporation's commercially-supported build of MariaDB. This guide shows how to point a KubeDB `MariaDB` database at a MariaDB Enterprise Server image instead of the community build.

This is a specialized case of the generic [private Docker registry](/docs/guides/mariadb/private-registry/quickstart/index.md) mechanism; read that guide first for the general concepts (`MariaDBVersion` catalog, `imagePullSecrets`). This page only covers what's specific to MariaDB Enterprise Server.

## Before You Begin

- Read [concept of MariaDB Version Catalog](/docs/guides/mariadb/concepts/mariadb-version/index.md) to learn the details of the `MariaDBVersion` object, including the `spec.distribution` field.
- You need a Kubernetes cluster, with `kubectl` configured to communicate with it. If you don't already have one, create one with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- You need an active MariaDB Enterprise Server subscription and a **Customer Download Token**, available from the [MariaDB Customer Portal](https://customers.mariadb.com/) under your account's download token page. This token is what grants pull access to `docker.mariadb.com`; it is not a runtime license and is never mounted into the database pod.

## No in-cluster license is required

Unlike some other commercial database images, **MariaDB Enterprise Server does not check any license key at runtime.** `mariadbd` boots and runs exactly like the community edition once the image is pulled. Access control is entirely at the container registry level: MariaDB Corporation gates pulls from `docker.mariadb.com` behind your Customer Download Token, the same way any other private registry works. There is no separate license Secret, ConfigMap, or environment variable that `mariadbd` itself reads, beyond the standard `imagePullSecrets`.

**KubeDB itself does ask for one more thing at the API level.** The `11.4.10-enterprise` `MariaDBVersion` catalog entry sets `spec.license.required: true`, mirroring the same `license.required` pattern KubeDB uses for other licensed distributions (e.g. AppsCode's Postgres Enterprise build and Oracle's MySQL Enterprise Edition). This does **not** mean a certificate gets mounted into the container or checked by `mariadbd` &mdash; it means KubeDB's admission webhook requires you to set `spec.license.secretRef` on the `MariaDB` CR, pointing at a Secret you provide, as an explicit administrative record that this deployment is backed by a real MariaDB Enterprise subscription. Any Secret works; nothing reads its contents. If you omit `spec.license`, the webhook rejects the `MariaDB` object with an error naming the `MariaDBVersion` that requires it.

## 1. Create an ImagePullSecret

Use your MariaDB ID as the username and your Customer Download Token as the password:

```bash
$ kubectl create secret docker-registry -n demo mariadb-enterprise-registry \
  --docker-server=docker.mariadb.com \
  --docker-username=<your-mariadb-id> \
  --docker-password=<your-customer-download-token> \
  --docker-email=<your-email>
secret/mariadb-enterprise-registry created
```

(See [Customer access to docker.mariadb.com](https://mariadb.com/docs/tools/mariadb-enterprise-operator/customer-access-to-docker-mariadb-com) for how to retrieve your token.)

## 2. Add a MariaDBVersion catalog entry for Enterprise Server

KubeDB ships a ready-made `11.4.10-enterprise` `MariaDBVersion` pointing at `docker.mariadb.com/enterprise-server:11.4.10` (added via the `kubedb-catalog` chart). If you need a different Enterprise Server version, create your own entry the same way:

```yaml
apiVersion: catalog.kubedb.com/v1alpha1
kind: MariaDBVersion
metadata:
  name: 11.4.10-enterprise
spec:
  coordinator:
    image: ghcr.io/kubedb/mariadb-coordinator:v0.47.0-rc.2
  db:
    image: docker.mariadb.com/enterprise-server:11.4.10
  distribution: Enterprise
  exporter:
    image: docker.io/prom/mysqld-exporter:v0.18.0
  initContainer:
    image: ghcr.io/kubedb/mariadb-init:0.9.0
  license:
    required: true
  podSecurityPolicies:
    databasePolicyName: maria-db
  version: 11.4.10
```

`spec.distribution: Enterprise` is descriptive metadata; it does not change how the operator reconciles the database. `spec.license.required: true` is what makes `spec.license` mandatory on any `MariaDB` using this version (see the callout above). Only `spec.db.image` and the pull secret determine which image actually runs.

## 3. Create a license Secret

Because the catalog entry sets `license.required: true`, every `MariaDB` using it must reference a Secret via `spec.license.secretRef`. Nothing inside KubeDB reads this Secret's contents at runtime; it exists purely so the deployment carries an explicit record of the MariaDB Enterprise subscription backing it. Any non-empty value works, for example your MariaDB customer account ID:

```bash
$ kubectl create secret generic -n demo mariadb-enterprise-license \
  --from-literal=license=<your MariaDB customer account ID or subscription reference>
secret/mariadb-enterprise-license created
```

## 4. Deploy a MariaDB using the Enterprise Server image

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb-enterprise
  namespace: demo
spec:
  version: "11.4.10-enterprise"
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  license:
    secretRef:
      name: mariadb-enterprise-license
  podTemplate:
    spec:
      imagePullSecrets:
      - name: mariadb-enterprise-registry
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f mariadb-enterprise.yaml
mariadb.kubedb.com/mariadb-enterprise created
```

Check that the pod comes up and is pulling from `docker.mariadb.com`:

```bash
$ kubectl get pods -n demo
NAME                    READY   STATUS    RESTARTS   AGE
mariadb-enterprise-0    1/1     Running   0          64s
```

## Cleaning up

```bash
$ kubectl delete mariadb -n demo mariadb-enterprise
mariadb.kubedb.com "mariadb-enterprise" deleted
$ kubectl delete secret -n demo mariadb-enterprise-registry
```

## Known limitations

The exact `docker.mariadb.com/enterprise-server` image path and tag naming in this guide have not been verified against a live pull with a real Customer Download Token. Confirm the path in the [MariaDB Enterprise Docker Registry documentation](https://mariadb.com/docs/server/server-management/automated-mariadb-deployment-and-administration/docker-and-mariadb/mariadb-enterprise-docker-registry-for-mariadb-enterprise-server) before relying on this in production.

## Next Steps

- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/mariadb-version/index.md).
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
