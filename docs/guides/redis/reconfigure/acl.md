---
title: Reconfiguring Redis ACL
menu:
  docs_{{ .version }}:
    identifier: rd-reconfigure-acl
    name: Overview
    parent: rd-reconfigure
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring Redis ACL

This guide explains how to configure Redis Access Control Lists (ACL) when deploying Redis with KubeDB, and how to modify ACL users later using a RedisOpsRequest (Reconfigure).

## Overview

- Define ACL rules in the Redis CR under `spec.acl.rules`.
- Provide passwords via a Kubernetes Secret referenced by `spec.acl.secretRef`.
- To change or add users later, create a new Secret and a `RedisOpsRequest` of type `Reconfigure` that references the new Secret and uses `configuration.auth` fields.

`Note:` The way described bellow can be applied to valkey as well.

## 1. Deploy Redis with ACL

Example Redis CR that defines users via ACL rules and references a Secret for passwords:

```yaml
# example: Redis CR with ACL (shortened)
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-instance
  namespace: demo
spec:
  version: 8.2.2
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    resources:
      requests:
        storage: 20M
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
  deletionPolicy: WipeOut
  acl:
    secretRef:
      name: old-acl-secret         # Secret that holds passwords referenced by variables like ${k1}
    rules:
      - userName1 ${k1} allkeys +@string +@set -SADD
      - userName2 ${k2} allkeys +@string +@set -SADD
      - userName3 ${k3} allkeys +@string +@set -SADD
      - userName4 ${k4} allkeys +@string +@set -SADD
```

## 2. Create the Secret with passwords

Store the passwords as stringData keys that match the variable names used in `rules`:

```yaml
# example: Secret referenced by spec.acl.secretRef (old-acl-secret)
apiVersion: v1
kind: Secret
metadata:
  name: old-acl-secret
  namespace: demo
type: Opaque
stringData:
  k1: "pass1"
  k2: "pass2"
  k3: "pass3"
  k4: "pass4"
```

Apply both the Secret and the Redis CR. Wait until the Redis resource becomes Ready:

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/acl/old-acl-secret.yaml
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/acl/redis.yaml
kubectl get rd -n demo
# expect: redis-instance  8.2.2  Ready
```

## 3. Verify loaded ACL users

Use redis-cli inside a Redis Pod to list users:

```bash
kubectl exec -it -n demo redis-instance-shard0-0 -c redis -- redis-cli acl list
```

You should see entries for `userName1`, `userName2`, `userName3`, `userName4` and `default`.

## 4. Modify ACLs using RedisOpsRequest (Reconfigure)

To add or update users and/or delete existing users, create:
1) a new Secret with the new/updated password keys, and
2) a `RedisOpsRequest` of type `Reconfigure` referencing that Secret.

Example Secret (new-acl-secret):

```yaml
# example: new secret with updated/extra credentials
apiVersion: v1
kind: Secret
metadata:
  name: new-acl-secret
  namespace: demo
type: Opaque
stringData:
  k1: "updatedPass1"    # existing user password
  k10: "pass10"  # new user password for userName10
```

Example RedisOpsRequest that:
- syncs ACL entries from the given values,
- deletes an existing user `userName2`,
- references the `new-acl-secret` Secret.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: RedisOpsRequest
metadata:
  name: rdops
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: redis-instance
  configuration:
    auth:
      syncACL:
        - userName1 ${k1} +get ~mykeys:*     # update userName1's rule (uses ${k1} from new-acl-secret or the previous secret)
        - userName10 ${k10} +get ~mykeys:*   # add new user userName10
      deleteUsers:
        - userName2                           # remove userName2 from ACL
      secretRef:
        name: new-acl-secret                   # secret containing referenced keys
```

Apply the Secret and the RedisOpsRequest:

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/acl/new-acl-secret.yaml
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/acl/rd-ops.yaml
kubectl get redisopsrequest -n demo
# expect: rdops  Reconfigure  Successful
```

## 5. Verify reconfiguration

- Confirm the Redis CR `spec.acl` has been updated (operator patches spec to refer to the new secret and merged rules).
- Verify ACLs inside Redis:

### verify Redis CR spec.acl
```bash
kubectl get rd -n demo redis-instance -o yaml | yq '.spec.acl'
{
  "rules": [
    "userName3 ${k3} allkeys +@string +@set -SADD",
    "userName4 ${k4} allkeys +@string +@set -SADD",
    "userName10 ${k10} +get ~mykeys:*",
    "userName1 ${k1} +get ~mykeys:*"
  ],
  "secretRef": {
    "name": "new-acl-secret"
  }
}
```

```bash
kubectl exec -it -n demo redis-instance-shard0-0 -c redis -- redis-cli acl list
# expect entries for userName1 (with new rule), userName10, userName3, userName4, and default
```

## Notes and tips

- Variable substitution: ACL rules in the Redis CR and RedisOpsRequest can refer to Secret keys using `${key}`; the operator substitutes these values from the referenced Secret.
- Order: The operator typically applies the new secret and patches the Redis `spec.acl` before starting pods, ensuring the running Redis instances load the new ACL configuration.
- Safe updates: Use `RedisOpsRequest` Reconfigure to change ACLs without manually editing the primary Redis CR; Ops-manager will pause/resume the database during the operation.
- Deleting users: Use `deleteUsers` in RedisOpsRequest when you want to remove users rather than just overwrite rules.

## Cleanup

Remove resources created for testing:

```bash
kubectl delete redisopsrequest -n demo rdops
kubectl delete rd -n demo redis-instance
kubectl delete secret -n demo old-acl-secret new-acl-secret
```
