---
title: Access Control Lists (ACL)
menu:
  docs_{{ .version }}:
    identifier: rd-configuration-acl
    name: ACL
    parent: rd-configuration
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

## Before You Begin

- [Custom Configuration](docs/guides/redis/configuration/redis.md)

> This guide shows how to configure Redis Access Control Lists (ACL) when creating a Redis resource with KubeDB.

# Deploy Redis with ACL

This document explains how to configure ACL users during initial deployment so the operator provisions Redis/Valkey with the desired ACL entries.

## Overview

- Define ACL rules in the Redis CR under `spec.acl.rules`.
- Provide passwords via a Kubernetes Secret referenced by `spec.acl.secretRef`.
- The operator substitutes `${key}` placeholders from the referenced Secret and applies ACLs when provisioning Redis.

## Example Redis CR

Define users and reference a Secret that contains passwords:

```yaml
# Redis CR: deploy with ACL
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
      name: rd-user         # Secret with password keys referenced below
    rules:
      - userName1 ${k1} allkeys +@string +@set -SADD
      - userName2 ${k2} allkeys +@string +@set -SADD
      - userName3 ${k3} allkeys +@string +@set -SADD
      - userName4 ${k4} allkeys +@string +@set -SADD
```

## Secret with passwords

Create a Secret whose `stringData` keys match the variable names used in the ACL rules:

```yaml
# Secret: rd-user
apiVersion: v1
kind: Secret
metadata:
  name: rd-user
  namespace: demo
type: Opaque
stringData:
  k1: "pass1"
  k2: "pass2"
  k3: "pass3"
  k4: "pass4"
```

Apply both the Secret and the Redis CR:

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/acl/old-acl-secret.yaml
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/redis/reconfigure/acl/redis.yaml
```

Wait until the Redis resource is Ready:

```bash
kubectl get rd -n demo
# expect: redis-instance  8.2.2  Ready
```

## Verify ACLs inside Redis

List ACL users from a Redis pod:

```bash
kubectl exec -it -n demo redis-instance-shard0-0 -c redis -- redis-cli acl list
1) "user userName1 on sanitize-payload #e6c3da5b206634d7f3f3586d747ffdb36b5c675757b380c6a5fe5c570c714349 ~* resetchannels -@all +@string +@set -sadd"
2) "user userName2 on sanitize-payload #1ba3d16e9881959f8c9a9762854f72c6e6321cdd44358a10a4e939033117eab9 ~* resetchannels -@all +@string +@set -sadd"
3) "user userName3 on sanitize-payload #3acb59306ef6e660cf832d1d34c4fba3d88d616f0bb5c2a9e0f82d18ef6fc167 ~* resetchannels -@all +@string +@set -sadd"
4) "user userName4 on sanitize-payload #a417b5dc3d06d15d91c6687e27fc1705ebc56b3b2d813abe03066e5643fe4e74 ~* resetchannels -@all +@string +@set -sadd"
5) "user default on sanitize-payload #b23f6deded0b32c4cac29a8846d8a21e3403a04961436bc686d9e59e3925371c ~* &* +@all"
```

You should see entries for userName1..userName4 and the default user. Each user line includes the hashed password, command categories and key patterns.

## Notes and tips

- Variable substitution: Use `${key}` in rules; the operator replaces these from the `spec.acl.secretRef` Secret.
- Secrets: Store passwords in a Secret referenced by `spec.acl.secretRef`. Use `stringData` during creation for convenience.
- Safety: For updates after deployment, use a `RedisOpsRequest` (Reconfigure) to sync ACL changes or delete users; this avoids manual pod edits.

## Next steps

If you need to modify or rotate ACL users later, see the reconfigure guide:
- [reconfigure acl](docs/guides/redis/reconfigure/acl.md)
