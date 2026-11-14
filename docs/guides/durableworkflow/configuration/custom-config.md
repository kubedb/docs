---
title: Custom Configuration
menu:
  docs_{{ .version }}:
    identifier: dw-custom-config-configuration
    name: Custom Configuration
    parent: dw-configuration
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Custom Configuration

KubeDB renders the server configuration of the engine for you: the persistence stores, the ports, the membership, the metrics and, when enabled, TLS and authorization. Passwords are never written to the configuration. They are read from environment variables that come from Secrets.

`spec.configuration` lets you change it. It has the usual KubeDB shape: a Secret (`secretName`), inline entries (`inline`), or both. Inline entries win over the same key in the Secret. Two keys are meaningful:

| Key                   | Effect                                                                                                                              | Pods restart? |
|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------|:-------------:|
| `dynamic_config.yaml` | The dynamic configuration of the engine. It is merged over the operator's drain settings and mounted from its own ConfigMap, which the servers poll. | no            |
| `server.yaml`         | Deep-merged over the rendered static configuration of the engine.                                                                    | yes, rolled   |

Other keys are ignored.

## Dynamic configuration

The syntax is the one of the engine: a key with a list of values.

```yaml
history.persistenceMaxQPS:
- value: 3000
```

The operator always sets the drain settings of the roles (for example `history.shutdownDrainDuration`), because the engines default them to 0, which would drop requests whenever a pod is replaced. A key of yours overrides them, so keep them unless you know why.

## Static configuration

`server.yaml` is a fragment of the static configuration of the engine and is merged key by key. Do not set `numHistoryShards` or change the persistence engine here. The usual use is to lower or raise the connection pools:

```yaml
persistence:
  datastores:
    default:
      sql:
        maxConns: 10
        maxIdleConns: 10
```

The pool of every pod is limited to 5 connections per store by default. If you raise it, raise `max_connections` of the database too, see the [connection budget](/docs/guides/durableworkflow/quickstart/prerequisites.md#the-connection-budget).

## Inline

```yaml
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-config
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
  persistence:
    default:
      name: pg
      namespace: demo
  configuration:
    inline:
      dynamic_config.yaml: |
        history.persistenceMaxQPS:
        - value: 3000
      server.yaml: |
        persistence:
          datastores:
            default:
              sql:
                maxConns: 10
                maxIdleConns: 10
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/configuration/temporal-config.yaml
durableworkflow.kubedb.com/temporal-config created
```

The same `spec.configuration` works for Cadence, see [cadence-config.yaml](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/configuration/cadence-config.yaml).

## From a Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dw-custom-config
  namespace: demo
stringData:
  dynamic_config.yaml: |
    history.persistenceMaxQPS:
    - value: 3000
---
apiVersion: kubedb.com/v1alpha2
kind: DurableWorkflow
metadata:
  name: temporal-secret-config
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
  persistence:
    default:
      name: pg
      namespace: demo
  configuration:
    secretName: dw-custom-config
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/durableworkflow/configuration/config-secret.yaml
```

> Note: YAML files used in this tutorial are stored in [docs/examples/durableworkflow](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/durableworkflow) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

To change the configuration of a running cluster, use a [Reconfigure](/docs/guides/durableworkflow/reconfigure/overview.md) ops request.

## Next Steps

- [Reconfigure](/docs/guides/durableworkflow/reconfigure/overview.md) a running cluster.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
