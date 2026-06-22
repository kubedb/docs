---
title: Rotate Authentication HanaDB
menu:
  docs_{{ .version }}:
    identifier: hanadb-rotate-auth-guide
    name: Rotate Authentication
    parent: hanadb-rotate-auth
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication for HanaDB

This guide shows how to rotate HanaDB authentication credentials using a `HanaDBOpsRequest`.

## Before You Begin

- Prepare a Kubernetes cluster and configure `kubectl`.
- Install KubeDB following the steps [here](/docs/setup/README.md).
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

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hanadb-new-auth
  namespace: demo
type: kubernetes.io/basic-auth
stringData:
  username: SYSTEM
  password: Hana5678
---
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
