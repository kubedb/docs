---
title: Running PostgreSQL Branch Post-Actions
menu:
  docs_{{ .version }}:
    identifier: pg-branch-post-actions
    name: Post-Actions
    parent: pg-branch
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Running PostgreSQL Branch Post-Actions

This tutorial anonymizes email addresses in a PostgreSQL branch before its Branch resource becomes `Ready`. The source data remains unchanged. Read the common [post-action guide](/docs/operatormanual/branching/post-actions.md) for execution order and Job settings.

## Before You Begin

Prepare and seed the HA source using a [creation tutorial](/docs/guides/postgres/branch/overview.md). The `branch_demo` table must exist in database `postgres`.


These commands use explicit contexts. For the same-cluster environment, initialize:

```bash
SOURCE_CONTEXT=$(kubectl config current-context)
TARGET_CONTEXT="$SOURCE_CONTEXT"
```

For the cross-cluster environment, set `SOURCE_CONTEXT=source` and `TARGET_CONTEXT=target`, and add `spec.target.clusterName: target` to the downloaded Branch manifest **before creating it**. Apply it on the source cluster. The `demo/prod-pg` source and `dev` target namespace must already exist as described in the creation tutorials.


## Create a Branch with an action

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: pg-actions
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: prod-pg
    namespace: demo
  target:
    namespace: dev
    name: pg-actions
    storageClassName: ceph-rbd
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
  resetRootPassword: true
  postActions:
    - image: postgres:16-alpine
      jobDefaults:
        imagePullPolicy: IfNotPresent
        backoffLimit: 3
        activeDeadlineSeconds: 900
        ttlSecondsAfterFinished: 3600
      jobTemplate:
        metadata:
          labels:
            branch-action: anonymize
        spec:
          args:
            - sh
            - -ec
            - |
              psql -v ON_ERROR_STOP=1 -d postgres -c \
                "UPDATE branch_demo SET email = 'user-' || id || '@example.invalid';"
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
  deletionPolicy: Delete
```

Courier supplies `PGHOST`, `PGPORT`, `PGUSER`, and `PGPASSWORD` for the target. The image contains `psql`; its arguments run SQL against database `postgres`. `ON_ERROR_STOP=1` and the shell's `-e` make SQL failure fail the Job.

The action sets each address to a deterministic value derived from its ID, so retrying the update is safe. Customize SQL for your data model. With multiple actions, Courier runs them in list order. Set container options such as arguments and resources directly under `jobTemplate.spec`.

```bash
curl -fsSL -o branch-post-actions.yaml https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/branch/branch-post-actions.yaml
# For cross-cluster use, add spec.target.clusterName: target to this file first.
kubectl --context "$SOURCE_CONTEXT" apply -f branch-post-actions.yaml
kubectl --context "$SOURCE_CONTEXT" -n demo wait branch/pg-actions \
  --for=jsonpath='{.status.phase}'=Ready --timeout=20m
```

The Branch passes through `ActionsRunning` and reaches `Ready` only after the action Job succeeds. Wait for Branch readiness before allowing applications to consume the prepared data.

## Inspect the action Job

Jobs run in `dev` on the target cluster. List the Jobs, then use the action's pod label to inspect its pods and logs:

```bash
kubectl --context "$TARGET_CONTEXT" -n dev get jobs
kubectl --context "$TARGET_CONTEXT" -n dev get pods -l branch-action=anonymize
kubectl --context "$TARGET_CONTEXT" -n dev logs \
  -l branch-action=anonymize -c actions --tail=100
```

For a single action, its Job name is also exposed in the executing Branch's `status.resources.postActionJob`. A successful SQL update completes the Job; terminal Job failure leaves the Branch `Failed` until its action specification is corrected.

## Verify branch-only anonymization

```bash
BRANCH_POD=$(kubectl --context "$TARGET_CONTEXT" -n dev get pods \
  -l app.kubernetes.io/instance=pg-actions,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
BRANCH_SECRET=$(kubectl --context "$TARGET_CONTEXT" -n dev get postgres pg-actions -o jsonpath='{.spec.authSecret.name}')
BRANCH_USER=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.username}' | base64 -d)
BRANCH_PASSWORD=$(kubectl --context "$TARGET_CONTEXT" -n dev get secret "$BRANCH_SECRET" -o jsonpath='{.data.password}' | base64 -d)
kubectl --context "$TARGET_CONTEXT" -n dev exec "$BRANCH_POD" -c postgres -- \
  env PGPASSWORD="$BRANCH_PASSWORD" psql -h pg-actions -U "$BRANCH_USER" -d postgres \
  -c 'SELECT id, email FROM branch_demo WHERE id IN (1, 2) ORDER BY id;'
SOURCE_POD=$(kubectl --context "$SOURCE_CONTEXT" -n demo get pods \
  -l app.kubernetes.io/instance=prod-pg,kubedb.com/role=primary \
  -o jsonpath='{.items[0].metadata.name}')
kubectl --context "$SOURCE_CONTEXT" -n demo exec "$SOURCE_POD" -c postgres -- \
  psql -U postgres -d postgres \
  -c 'SELECT id, email FROM branch_demo WHERE id IN (1, 2) ORDER BY id;'
unset BRANCH_PASSWORD
```

The branch returns `user-1@example.invalid` and `user-2@example.invalid`; the source retains `alice@example.com` and `bob@example.com`.

## Combine post-actions with refresh

Add a `spec.schedule.cron` to the Branch before creating it to run the same actions after each refresh. Newly copied source rows are anonymized before the branch returns to `Ready`. Each generation gets its own action Jobs, and Courier removes stale action Jobs after successful replacement. Inspect current pods instead of relying on a fixed Job name across refreshes.

## Cleanup

```bash
kubectl --context "$SOURCE_CONTEXT" -n demo delete branch pg-actions --wait=true --timeout=15m
```
