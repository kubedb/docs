---
title: Rotate Authentication HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-rotate-auth-guide
    name: Guide
    parent: hanadb-rotate-auth
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Start with the [KubeDB documentation overview](/docs/README.md).

# Rotate Authentication for HanaDB

This guide shows how to rotate HanaDB authentication credentials using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB by following the [setup guide](/docs/setup/README.md).
- Create a namespace:

```bash
kubectl create ns demo
```

> Note: YAML files used in this tutorial are stored in [docs/examples/hanadb/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/rotate-auth).

## Deploy HanaDB

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/rotate-auth/hanadb-standalone.yaml
kubectl wait -n demo hanadb/hanadb-standalone --for=jsonpath='{.status.phase}'=Ready --timeout=1800s
```

## Rotate with Operator-Generated Password

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: hanadb-standalone
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/rotate-auth/rotate-auth-generated.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-rotate-auth-generated --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
```

## Rotate with User-Provided Password

Use a `kubernetes.io/basic-auth` Secret. The username must remain `SYSTEM`.

```bash
export NEW_HANA_PASSWORD='<set-a-hana-password>'
kubectl create secret generic hanadb-new-auth -n demo \
  --type=kubernetes.io/basic-auth \
  --from-literal=username=SYSTEM \
  --from-literal=password="${NEW_HANA_PASSWORD}"
```

Now create a `HanaDBOpsRequest` with `RotateAuth` type.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-rotate-auth-user
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: hanadb-standalone
  authentication:
    secretRef:
      kind: Secret
      name: hanadb-new-auth
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/rotate-auth/rotate-auth-user.yaml
kubectl wait -n demo hanadbopsrequest/hdbops-rotate-auth-user --for=jsonpath='{.status.phase}'=Successful --timeout=1800s
```

## Verify

```bash
kubectl get hdbops -n demo
kubectl get secret -n demo hanadb-standalone-auth -o jsonpath='{.data.password}' | base64 -d
```

## Cleanup

```bash
kubectl delete hdbops -n demo --all
kubectl delete hanadb -n demo hanadb-standalone
kubectl delete ns demo
```
