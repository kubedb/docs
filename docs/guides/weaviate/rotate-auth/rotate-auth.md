---
title: Rotate Authentication for Weaviate
menu:
  docs_{{ .version }}:
    identifier: weaviate-rotate-auth-description
    name: Rotate Auth
    parent: weaviate-rotate-auth
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Rotate Authentication for Weaviate

This guide will show you how to use the `KubeDB` Ops Manager to rotate the API-key authentication of a Weaviate cluster.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Rotate Authentication Overview](/docs/guides/weaviate/rotate-auth/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/rotate-auth](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/rotate-auth) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

Deploy a Weaviate cluster and wait for it to become `Ready`. By default, KubeDB generates an API key and stores it in the `weaviate-sample-auth` Secret:

```bash
$ kubectl get secret -n demo weaviate-sample-auth -o jsonpath='{.data.AUTHENTICATION_APIKEY_ALLOWED_KEYS}' | base64 -d
vzWSjiRGNNEZEytR
```

You can confirm this key works through a port-forward:

```bash
$ kubectl port-forward -n demo svc/weaviate-sample 8080:8080
# in another terminal
$ curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/v1/schema \
  -H "Authorization: Bearer vzWSjiRGNNEZEytR"
200
```

## Rotate Auth with a User-provided Secret

You can rotate the API key to a value you control by providing a Secret with the new key under the `AUTHENTICATION_APIKEY_ALLOWED_KEYS` data field.

First, create the Secret holding the new key:

```yaml
apiVersion: v1
data:
  AUTHENTICATION_APIKEY_ALLOWED_KEYS: VTFVenZyVHZuejVNdzljNA==
kind: Secret
metadata:
  name: weaviate-rotate-auth
  namespace: demo
type: Opaque
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/rotate-auth/weaviate-rotate-auth.yaml
secret/weaviate-rotate-auth created
```

Now, create the `RotateAuth` OpsRequest referencing that Secret:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: weaviate-rotate-auth-generated
  namespace: demo
spec:
  type: RotateAuth
  databaseRef:
    name: weaviate-sample
  authentication:
    secretRef:
      kind: Secret
      name: weaviate-rotate-auth
  timeout: 5m
  apply: IfReady
```

- `spec.type` specifies that this is a `RotateAuth` operation.
- `spec.authentication.secretRef.name` references the Secret holding the new API key. If you omit this field, the Ops Manager generates a brand-new random key instead.

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/rotate-auth/ops-request.yaml
weaviateopsrequest.ops.kubedb.com/weaviate-rotate-auth-generated created
```

The Ops Manager updates the credentials and restarts the pods one by one.

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-rotate-auth-generated
NAME                             TYPE         STATUS       AGE
weaviate-rotate-auth-generated   RotateAuth   Successful   70s
```

Let's check the `status.conditions` of the `WeaviateOpsRequest`:

```bash
$ kubectl get weaviateopsrequest -n demo weaviate-rotate-auth-generated -o yaml
...
status:
  conditions:
  - message: Weaviate ops-request has started to rotate auth for Weaviate nodes
    reason: RotateAuth
    status: "True"
    type: RotateAuth
  - message: Successfully referenced the user provided authSecret
    reason: UpdateCredential
    status: "True"
    type: UpdateCredential
  - message: successfully reconciled the Weaviate with new configuration
    reason: UpdatePetSets
    status: "True"
    type: UpdatePetSets
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: GetPod--weaviate-sample-0
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: EvictPod--weaviate-sample-0
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-0
    status: "True"
    type: RunningPod--weaviate-sample-0
  - message: get pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: GetPod--weaviate-sample-1
  - message: evict pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: EvictPod--weaviate-sample-1
  - message: running pod; ConditionStatus:True; PodName:weaviate-sample-1
    status: "True"
    type: RunningPod--weaviate-sample-1
  - message: Successfully Restarted Pods after rotating auth
    reason: RestartNodes
    status: "True"
    type: RestartNodes
  - message: Successfully completed rotating auth for Weaviate
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful
```

## Verify Authentication Rotated

After the rotation, the `Weaviate` object now references the provided Secret, and the Ops Manager has enriched it with the previous key (under `*-PREV`), the enabled flag, and the bound user:

```bash
$ kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.authSecret}'
{"activeFrom":"2026-06-30T17:47:20Z","apiGroup":"","externallyManaged":true,"kind":"","name":"weaviate-rotate-auth"}

$ kubectl get secret -n demo weaviate-rotate-auth -o go-template='{{ range $k, $v := .data }}{{ $k }}: {{ $v | base64decode }}{{ "\n" }}{{ end }}'
AUTHENTICATION_APIKEY_ALLOWED_KEYS: U1UzvrTvnz5Mw9c4
AUTHENTICATION_APIKEY_ALLOWED_KEYS-PREV: vzWSjiRGNNEZEytR
AUTHENTICATION_APIKEY_ENABLED: true
AUTHENTICATION_APIKEY_USERS: admin
```

Let's confirm that the new key works and the old key is rejected:

```bash
$ kubectl port-forward -n demo svc/weaviate-sample 8080:8080
# in another terminal
$ curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/v1/schema \
  -H "Authorization: Bearer U1UzvrTvnz5Mw9c4"
200

$ curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/v1/schema \
  -H "Authorization: Bearer vzWSjiRGNNEZEytR"
401
```

The new key returns `200` while the old key now returns `401` — the authentication has been rotated successfully.

> **Tip:** To let the operator generate a fresh random key instead of supplying your own, omit `spec.authentication` from the OpsRequest. KubeDB will generate a new key and update the `<database-name>-auth` Secret in place.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Encrypt traffic with [TLS](/docs/guides/weaviate/tls/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete weaviateopsrequest -n demo weaviate-rotate-auth-generated
$ kubectl delete weaviate -n demo weaviate-sample
$ kubectl delete ns demo
```
