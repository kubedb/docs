---
title: Branch Neo4j into Another Namespace
menu:
  docs_{{ .version }}:
    identifier: neo4j-branch-same-cluster
    name: Same-Cluster
    parent: guides-neo4j-branch
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch Neo4j into Another Namespace

Same-cluster branching creates an independent Neo4j deployment from a point-in-time snapshot within the same Kubernetes cluster. Place the branch in another namespace to prepare development, testing, or data analysis environments using existing data.

The branch accepts independent reads and writes without changing the source. Source changes are not continuously replicated; scheduled refreshes can replace branch data with fresh source snapshots.

## Before You Begin

- Read the [branching overview](/docs/operatormanual/branching/overview.md).
- Complete [Courier setup](/docs/operatormanual/branching/setup.md#single-cluster-installation), including KubeDB and the Branch CRD. Use branching-capable Courier, provisioner, and database images.
- Ensure `Neo4jVersion` `2025.12.1` is available in your installed catalog.
- Use Ceph RBD storage with copy-on-write cloning, compatible VolumeSnapshot and VolumeGroupSnapshot classes, and the corresponding CSI APIs, controllers, and driver support.
- Review the [snapshot requirements](/docs/operatormanual/branching/snapshots.md). Your account must be able to create databases and Branches, read auth Secrets, inspect snapshots and Jobs, and open pod shells.
- Allow capacity for three source and three target database pods, each requesting 2Gi of memory, plus operators and action Jobs.

The source uses `demo`; the target uses `dev`. Replace `ceph-rbd` consistently if your Ceph StorageClass has another name. Reserve the tutorial names and create namespaces if they do not already exist:

```shell
kubectl create namespace demo
kubectl create namespace dev
```

## Deploy the source database

Save the following as `neo4j-source.yaml`. It creates a three-server cluster:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Neo4j
metadata:
  name: neo4j-source
  namespace: demo
spec:
  version: 2025.12.1
  replicas: 3
  podTemplate:
    spec:
      containers:
      - name: neo4j
        resources:
          requests:
            cpu: 500m
            memory: 2Gi
          limits:
            memory: 2Gi
  storageType: Durable
  storage:
    storageClassName: ceph-rbd
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 4Gi
  deletionPolicy: WipeOut
```

```shell
kubectl -n demo apply -f neo4j-source.yaml
kubectl -n demo wait neo4j/neo4j-source --for=jsonpath='{.status.phase}'=Ready --timeout=30m
```

Use the Enterprise server catalog entry for this cluster. The client examples select database `neo4j` and use routing so writes reach its leader.

## Seed and inspect the source data

Discover the auth Secret referenced by `neo4j-source`:

```shell
kubectl -n demo get neo4j neo4j-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

Create five customers. Expected results below omit connection banners and are formatted for readability.

```cypher
CREATE (c1:Customer {id: 1, full_name: "Customer 1", email: "customer1@example.com"}),
       (c2:Customer {id: 2, full_name: "Customer 2", email: "customer2@example.com"}),
       (c3:Customer {id: 3, full_name: "Customer 3", email: "customer3@example.com"}),
       (c4:Customer {id: 4, full_name: "Customer 4", email: "customer4@example.com"}),
       (c5:Customer {id: 5, full_name: "Customer 5", email: "customer5@example.com"});
CALL db.checkpoint();
```

`db.checkpoint()` is provided by the Enterprise server used here. Committed transaction logs still support recovery when the cloned database starts.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

## Create the Branch

Save the following as `neo4j-dev.yaml`:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: neo4j-dev
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Neo4j
      name: neo4j-source
    namespace: demo
  target:
    namespace: dev
    name: neo4j-dev
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        memory: 2Gi
  deletionPolicy: Delete
```

```shell
kubectl -n demo apply -f neo4j-dev.yaml
kubectl -n demo wait branch/neo4j-dev --for=condition=Ready --timeout=60m
kubectl -n demo get branch neo4j-dev
```

Courier creates `neo4j/neo4j-dev` and its cloned PVCs in `dev`; KubeDB provisions the target database.

## Verify the copied data

Discover the auth Secret referenced by `neo4j-dev`:

```shell
kubectl -n dev get neo4j neo4j-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-dev-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-dev-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-dev.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

### Inspect target resources

```shell
kubectl -n dev get pods,pvc -l app.kubernetes.io/instance=neo4j-dev
```

Discover the auth Secret referenced by `neo4j-dev`:

```shell
kubectl -n dev get neo4j neo4j-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-dev-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-dev-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-dev.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
SHOW SERVERS YIELD address, state, health RETURN address, state, health;
SHOW DATABASES YIELD name, address, writer, currentStatus RETURN name, address, writer, currentStatus;
```

Expect three branch server addresses, available servers, and online databases with a writer for `neo4j`.

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

## Verify independent writes

### Change the branch

Discover the auth Secret referenced by `neo4j-dev`:

```shell
kubectl -n dev get neo4j neo4j-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-dev-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-dev-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-dev.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer {id: 5}) SET c.full_name = "Branch-only Customer 5";
CALL db.checkpoint();
```

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Branch-only Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

### Change and inspect the source

Discover the auth Secret referenced by `neo4j-source`:

```shell
kubectl -n demo get neo4j neo4j-source -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-source-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n demo get secret neo4j-source-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n demo exec -it neo4j-source-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-source.demo.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer {id: 1}) SET c.full_name = "Updated Customer 1";
CALL db.checkpoint();
```

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Updated Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

The source retains the original customer `5`; the branch edit did not reach it.

### Check the branch again

Discover the auth Secret referenced by `neo4j-dev`:

```shell
kubectl -n dev get neo4j neo4j-dev -o jsonpath='{.spec.authSecret.name}{"\n"}'
```

```text
neo4j-dev-auth
```

Use the returned name below if it differs. Read its password:

```shell
kubectl -n dev get secret neo4j-dev-auth -o jsonpath='{.data.password}' | base64 -d
```

Enter this password when the client prompts.

From your terminal, open a pod shell:

```shell
kubectl -n dev exec -it neo4j-dev-0 -c neo4j -- bash
```

Inside the pod, start the client through the writable database Service:

```shell
cypher-shell -a neo4j://neo4j-dev.dev.svc:7687 -u neo4j -d neo4j --format plain
```

Enter Cypher at the client prompt. `neo4j://` discovers the cluster’s writer; a direct `bolt://` connection to an arbitrary member can reject writes.

```cypher
MATCH (c:Customer) RETURN c.id AS id, c.full_name AS full_name, c.email AS email ORDER BY id;
```

```text
id, full_name, email
1, "Customer 1", "customer1@example.com"
2, "Customer 2", "customer2@example.com"
3, "Customer 3", "customer3@example.com"
4, "Customer 4", "customer4@example.com"
5, "Branch-only Customer 5", "customer5@example.com"
```

Run `:exit` to leave `cypher-shell`, then `exit` to leave the pod shell.

The branch retains its own customer `5` and the original customer `1`. The source update did not reach it.

## Cleanup

Delete the Branch on the source cluster first. Let Courier remove its target and snapshots before deleting the source:

```shell
kubectl -n demo delete branch neo4j-dev --wait=true --timeout=30m
kubectl -n demo delete neo4j neo4j-source --wait=true --timeout=30m
```

Keep Courier, CSI, the shared backend, and OCM when used available until cleanup completes. These commands leave the namespaces in place.

## Next Steps

- [Refresh the branch](/docs/guides/neo4j/branch/refresh.md).
- [Reset the branch root password](/docs/guides/neo4j/branch/reset-root-password.md).
- [Prepare branch data with post-actions](/docs/guides/neo4j/branch/post-actions.md).
- [Configure snapshots, TLS, and deletion policies](/docs/guides/neo4j/branch/customization/index.md).
