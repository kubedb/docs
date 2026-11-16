---
title: Rotate YugabyteDB Authentication
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-rotate-auth-guide
    name: Rotate Auth
    parent: yugabytedb-rotate-auth
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate YugabyteDB Authentication

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/rotate-auth) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

Deploy `yb-demo` from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md).

## Create the new credentials Secret

The operation used a user provided Secret named `yb-new-auth`. It needs the keys `username` (must be `yugabyte`) and `password`. The exact command used in the test was not recorded; a Secret of this shape works:

```bash
$ kubectl create secret generic yb-new-auth -n demo --from-literal=username=yugabyte --from-literal=password='<new password>'
```

(If you do not provide `secretRef`, KubeDB generates a random password. That variant was not run.)

## Apply the RotateAuth OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-rotate-auth
  namespace: demo
spec:
  type: RotateAuth
  databaseRef: {name: yb-demo}
  authentication:
    secretRef:
      kind: Secret
      name: yb-new-auth
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/rotate-auth/ops-request.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-rotate-auth created
```

## What was measured

Successful in **20 s** (conditions `UpdateCredential`, `VerifyCredential`). The new password was accepted and the old one rejected. `spec.authSecret` and the AppBinding secret reference pointed to `yb-new-auth` afterwards. Under load: 46,972 successful calls, **0 failed calls**; open sessions keep working (longest gap 13 ms). No pod restarted.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-rotate-auth
$ kubectl delete secret -n demo yb-new-auth
```
