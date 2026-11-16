---
title: JWT Authorization
menu:
  docs_{{ .version }}:
    identifier: dw-authorization-guide
    name: JWT Authorization
    parent: dw-authorization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# JWT Authorization

`spec.authorization` turns on JWT authorization of the API calls to the frontend for both engines. Tokens are signed by your identity provider (RS256). The cluster only verifies them.

Verified: for both engines, a call without a token and a call with a reader token are denied, and a call with an admin token is allowed.

## Before You Begin

- Complete [Prepare Dependencies](/docs/guides/durableworkflow/quickstart/prerequisites.md).
- An identity provider that serves a JWKS document, or the JWKS document itself.
- You should be familiar with [DurableWorkflow](/docs/guides/durableworkflow/concepts/durableworkflow/index.md).

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## The key source

Give exactly one of:

- `jwksURL`: the URL of the key set that verifies the tokens (both engines).
- `publicKeySecret`: a Secret with a JWKS document under the key `jwks.json` (**Temporal only**).

The webhook rejects both set, and none set.

## Temporal

Temporal's default claim mapper and authorizer are used. They read the permissions of the caller from a claim of the token, in the form `<namespace>:<role>`. Fields:

| Field                  | Description                                                              |
|------------------------|--------------------------------------------------------------------------|
| `permissionsClaimName` | The claim with the permissions. Default `permissions`.                    |
| `permissionsRegex`     | A regex that parses a permission into the groups `namespace` and `role`.   |
| `audience`             | The audience that a token must carry.                                    |

**Temporal needs an `internalFrontend` for authorization.** The system worker of the cluster talks to the internal frontend, which is not authorized. Without one, it would call the authorized frontend without a token, so the webhook requires `spec.topology.internalFrontend`.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-authz
  namespace: demo
spec:
  version: temporal-1.32.1
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
    internalFrontend:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  authorization:
    jwksURL: https://idp.example.com/.well-known/jwks.json
    permissionsClaimName: permissions
    audience: temporal
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/authorization/temporal-jwks-url.yaml
```

To verify the tokens with a static key set instead, put the JWKS document in a Secret and refer to it with `publicKeySecret` ([temporal-public-key.yaml](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/authorization/temporal-public-key.yaml)). Replace the modulus in the sample.

## Cadence

Cadence verifies external tokens with `jwksURL` only. `publicKeySecret` is rejected. Fields:

| Field                 | Description                                                        |
|-----------------------|--------------------------------------------------------------------|
| `maxTokenTTL`         | The longest token lifetime accepted, in seconds. Default `3600`.   |
| `groupsAttributePath` | The claim with the groups of the caller. Default `groups`.          |
| `adminAttributePath`  | The claim that marks an administrator. Default `admin`.             |

The `permissionsClaimName`, `permissionsRegex` and `audience` fields are Temporal only, and the three above are Cadence only. The webhook rejects the wrong ones.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: cadence-authz
  namespace: demo
spec:
  version: cadence-1.4.2
  topology:
    frontend:
      replicas: 2
    history:
      replicas: 2
    matching:
      replicas: 2
    worker:
      replicas: 2
  persistence:
    default:
      name: pg
      namespace: demo
  authorization:
    jwksURL: https://idp.example.com/.well-known/jwks.json
    maxTokenTTL: 3600
    groupsAttributePath: groups
    adminAttributePath: admin
  ui:
    enabled: true
    auth: JWT
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/authorization/cadence.yaml
```

- **Internal tokens.** The system worker of Cadence signs its own tokens. The operator generates an RSA key pair in the Secret `<name>-auth-keys` (`private.pem` and `public.pem`) and mounts it. That is why external tokens must come from a JWKS URL: the single static key slot of Cadence is taken by this key. You cannot provide your own key pair.
- **The UI.** `spec.ui.auth: JWT` makes cadence-web forward the token from the `cadence-authorization` cookie. Cadence's own checks decide what the caller may do. cadence-web has no login of its own, so put an authenticating proxy in front of it or paste a token on its `/login` page.
- Authorization covers both the gRPC and the TChannel listeners of the frontend.

## Cleaning up

```bash
kubectl delete dw -n demo temporal-authz temporal-authz-key cadence-authz
```

## Next Steps

- Run with [TLS](/docs/guides/durableworkflow/tls/tls.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
