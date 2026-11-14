---
title: Restart YugabyteDB
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-restart-guide
    name: Restart YugabyteDB
    parent: yugabytedb-restart
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Restart YugabyteDB

This guide shows how to restart all pods of a YugabyteDB universe with a `YugabyteDBOpsRequest`. The restart is rolling: masters first (the current leader last), then the tservers, one pod at a time. Before each tserver is stopped, KubeDB moves its tablet leaders away and checks that it is safe to take down; after each pod it waits until the pod is ready and caught up.

> Note: YAML files used in this tutorial are stored in [docs/examples/yugabytedb/restart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/yugabytedb/restart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Deploy the `yb-demo` universe from the [quickstart](/docs/guides/yugabytedb/quickstart/quickstart.md) and wait until it is `Ready`.
- Install the KubeDB Ops-manager.

## Apply the Restart OpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-demo-restart
  namespace: demo
spec:
  type: Restart
  databaseRef:
    name: yb-demo
  apply: IfReady
  rolling:
    masterWait: 20s
    tserverWait: 20s
```

- `apply: IfReady` runs the operation only while the database is `Ready`.
- `rolling.masterWait` and `rolling.tserverWait` are the pauses after each pod is back (the defaults are 60s; 20s was used in the test).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/restart/ops-request.yaml
yugabytedbopsrequest.ops.kubedb.com/yb-demo-restart created

$ kubectl get ybops -n demo yb-demo-restart -w
```

The conditions of a successful restart are `Restart`, `RollMasters`, `RollTServers` and `Successful`.

## What was measured

On the three-node kind cluster with RF3, while a load generator ran 4 workers of insert plus select through the client Service (about 1.2k calls per second):

| Measure | Result |
|---------|--------|
| Duration | 412 s (20 s waits) |
| Successful calls | 724,518 |
| Failed calls | 6, all `FATAL: terminating connection due to administrator command (57P01)`, i.e. connections that were open to the pod being stopped |
| Longest period without any success | 3.7 s |

A second run gave 1,354,494 successful calls, 8 dropped connections, no other failure and a longest gap of 4.2 s.

"Zero downtime" here means no failed calls except the connections dropped to the restarted pod. YSQL runs inside the tserver pod, so clients with a connection to that pod must reconnect. Use a client or pool that reconnects. With replication factor 1 there is nothing to fail over to and a restart is not zero downtime.

If the tablet leaders do not leave a tserver within `rolling.leaderDrainTimeout` (120 s by default), the restart goes on and records a `LeaderDrainTimedOut` condition.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-demo-restart
```

## Next Steps

- [Vertical Scaling](/docs/guides/yugabytedb/scaling/vertical-scaling/scale-vertically/)
- [YugabyteDBOpsRequest concepts](/docs/guides/yugabytedb/concepts/opsrequest.md)
