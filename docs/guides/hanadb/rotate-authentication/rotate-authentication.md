---
title: Rotate Authentication HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-rotate-authentication-guide
    name: Rotate Authentication
    parent: guides-hanadb-rotate-authentication
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication of HanaDB

This guide shows how to rotate the `SYSTEM` password of a HanaDB using a `HanaDBOpsRequest` of type
`RotateAuth`. You can let KubeDB generate a new password, or supply your own through a `Secret`.

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/rotate-authentication](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/rotate-authentication) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Create a namespace `demo` and deploy a standalone `hanadb-standalone` (see
  [Reconfigure](/docs/guides/hanadb/reconfigure/reconfigure.md#deploy-a-hanadb)).

## Password requirements

The HANA username always remains `SYSTEM`. A user-provided password must be HANA-compatible:

- ASCII letters and digits only,
- at least 8 characters,
- starts with a letter, and
- contains an uppercase letter, a lowercase letter, and a digit.

KubeDB-generated passwords already satisfy these rules.

## Check the Current Credentials

```bash
kubectl get secret hanadb-standalone-auth -n demo -o go-template='{{range $k,$v := .data}}{{$k}}{{"\n"}}{{end}}'
```
password
password.json
username

```bash
kubectl get secret hanadb-standalone-auth -n demo -o jsonpath='{.data.username}' | base64 -d; echo
```
SYSTEM

## Option A — Rotate with a KubeDB-generated password

Apply a `RotateAuth` ops request with no `authentication` block:

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/rotate-authentication/rotate-auth-generated.yaml
```
hanadbopsrequest.ops.kubedb.com/hdbops-rotate-auth-generated created

Wait for the ops request to succeed:

```bash
kubectl get hdbops -n demo hdbops-rotate-auth-generated
```
NAME                           TYPE         STATUS       AGE
hdbops-rotate-auth-generated   RotateAuth   Successful   4m22s

```bash
kubectl describe hdbops -n demo hdbops-rotate-auth-generated
```
...
Status:
  Conditions:
    Message:  HanaDBOpsRequest has started to rotate auth for HanaDB nodes
    Reason:   RotateAuth
    Status:   True
    Type:     RotateAuth
    Message:  Successfully generated new credentials
    Reason:   UpdateCredential
    Status:   True
    Type:     UpdateCredential
    Message:  Successfully reconciled HanaDB with new auth credentials
    Reason:   UpdatePetSets
    Status:   True
    Type:     UpdatePetSets
    Message:  Successfully restarted HanaDB nodes
    Reason:   RestartNodes
    Status:   True
    Type:     RestartNodes
    Message:  Successfully completed RotateAuth for HanaDB.
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful

KubeDB updates the `hanadb-standalone-auth` secret with the new password (keeping the previous one under
`.prev` keys) and verifies connectivity with the new credentials:

```bash
kubectl get secret hanadb-standalone-auth -n demo -o go-template='{{range $k,$v := .data}}{{$k}}{{"\n"}}{{end}}'
```
password
password.json
password.prev
username
username.prev

```bash
NEW_PASSWORD="$(kubectl get secret hanadb-standalone-auth -n demo -o jsonpath='{.data.password}' | base64 -d)"
```

```bash
kubectl exec -n demo hanadb-standalone-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p '$NEW_PASSWORD' 'SELECT 1 AS OK FROM DUMMY'"
```
OK
1
1 row selected

## Option B — Rotate with a user-provided password

First create a `Secret` with the new credentials (username `SYSTEM`):

```bash
kubectl create secret generic hanadb-new-auth -n demo \
  --from-literal=username=SYSTEM \
  --from-literal=password='NewHanaPass1'
```
secret/hanadb-new-auth created

Then reference it from the ops request:

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
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/rotate-authentication/rotate-auth-user.yaml
```
hanadbopsrequest.ops.kubedb.com/hdbops-rotate-auth-user created

```bash
kubectl get hdbops -n demo hdbops-rotate-auth-user
```
NAME                      TYPE         STATUS       AGE
hdbops-rotate-auth-user   RotateAuth   Successful   2m33s

After the request succeeds, KubeDB pins `spec.authSecret` to your secret (`externallyManaged: true`):

```bash
kubectl get hanadb.kubedb.com hanadb-standalone -n demo -o jsonpath='{.spec.authSecret}'
```
{"activeFrom":"...","externallyManaged":true,"name":"hanadb-new-auth"}

You can now connect with the password you supplied:

```bash
kubectl exec -n demo hanadb-standalone-0 -c hanadb -- /bin/sh -lc \
  "source /usr/sap/HXE/HDB90/HDBSettings.sh; hdbsql -i 90 -d SYSTEMDB -u SYSTEM -p 'NewHanaPass1' 'SELECT 1 AS OK FROM DUMMY'"
```
OK
1
1 row selected

## Cleaning Up

```bash
kubectl delete hdbops -n demo hdbops-rotate-auth-generated hdbops-rotate-auth-user
```

```bash
kubectl delete secret -n demo hanadb-new-auth
```

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-standalone
```

```bash
kubectl delete ns demo
```

## Next Steps

- Review the [HanaDB CRD](/docs/guides/hanadb/concepts/hanadb.md).
- Encrypt connections with [TLS/SSL](/docs/guides/hanadb/tls/overview.md).
