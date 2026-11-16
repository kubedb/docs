---
title: Reconfigure Overview
menu:
  docs_{{ .version }}:
    identifier: dw-reconfigure-overview
    name: Overview
    parent: dw-reconfigure-durableworkflow
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Overview

A `Reconfigure` ops request changes `spec.configuration` of a running DurableWorkflow. It has two very different effects, depending on which key changes.

| You change             | What happens                                                                                                                                      | Downtime |
|------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------|:--------:|
| `dynamic_config.yaml`  | The ConfigMap with the dynamic configuration is updated. **No pod is replaced.** The servers poll the file and pick the change up.                  | none     |
| `server.yaml`          | The static configuration changes, which changes the config checksum of the pod template. The pods are [rolled one at a time](/docs/guides/durableworkflow/update-version/overview.md#how-the-pods-are-rolled). | none on a highly available cluster |

Both were verified with 0 failed calls on a highly available cluster.

## The request

Under `spec.configuration`:

- `applyConfig` is a map from `server.yaml` or `dynamic_config.yaml` to the content. Its keys replace the same keys of `spec.configuration.inline`.
- `configSecret` replaces `spec.configuration.secretName`.
- `removeCustomConfig: true` first drops the current custom configuration, and then applies what the request carries.
- `restart` is `auto` (the default), `true` or `false`. A change of the static configuration rolls the pods unless it is `false`; with `false`, it takes effect when the pods are restarted next. A change of the dynamic configuration never needs a restart.

`numHistoryShards` cannot be changed, and the webhook rejects a request that tries it.

## Next Steps

- Run a [Reconfigure](/docs/guides/durableworkflow/reconfigure/durableworkflow.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
