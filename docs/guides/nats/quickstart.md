---
title: NATS Quickstart
menu:
  docs_{{ .version }}:
    identifier: nats-quickstart
    name: Quickstart
    parent: nats-guides
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# NATS QuickStart

This tutorial shows how to run a three-server NATS cluster with JetStream using KubeDB and connect to it.

> Note: The yaml files used in this tutorial are stored in [docs/examples/nats/quickstart](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/nats/quickstart) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- You need a Kubernetes cluster with `kubectl` configured. If you do not have one, create it with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install KubeDB following the steps [here](/docs/setup/README.md), with `--set global.featureGates.NATS=true` so the NATS CRDs are installed.
- This tutorial uses the `demo` namespace:

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Find Available NATSVersion

KubeDB installs a `NATSVersion` object for every supported release:

```bash
$ kubectl get natsversion
NAME     VERSION   DB_IMAGE                         DEPRECATED   AGE
2.14.8   2.14.8    ghcr.io/appscode-images/nats:2.14.8            2m
2.15.1   2.15.1    ghcr.io/appscode-images/nats:2.15.1            2m
```

## Create a NATS cluster

```yaml
apiVersion: kubedb.com/v1alpha2
kind: NATS
metadata:
  name: nats-quickstart
  namespace: demo
spec:
  version: "2.15.1"
  replicas: 3
  jetStream:
    storage:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
      storageClassName: standard
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/nats/quickstart/quickstart.yaml
nats.kubedb.com/nats-quickstart created
```

Here,

- `spec.version` is the name of the `NATSVersion` object that holds the images.
- `spec.replicas` is the number of `nats-server` pods. With JetStream on, use 1 or an odd number `>= 3`.
- `spec.jetStream.storage` is the PVC that holds the JetStream file store. The default is `10Gi`. Do not set limits here, PVCs are not resized automatically; use a [VolumeExpansion](/docs/guides/nats/ops-requests.md#volumeexpansion) ops request.
- `spec.deletionPolicy` controls what is kept when the object is deleted: `DoNotTerminate` blocks the delete, `Halt` keeps PVCs and Secrets, `Delete` keeps Secrets, `WipeOut` removes everything.

Wait for the object to become `Ready`:

```bash
$ kubectl get nats -n demo nats-quickstart
NAME              VERSION   STATUS   AGE
nats-quickstart   2.15.1    Ready    2m

$ kubectl get petset,svc,pvc,appbinding -n demo -l app.kubernetes.io/instance=nats-quickstart
```

KubeDB creates a PetSet, a client Service `nats-quickstart` (port 4222), a headless Service `nats-quickstart-pods` (routes, monitor), PVCs, and an AppBinding.

## Connect

KubeDB generates the `admin` user for the application account and stores it in `nats-quickstart-auth`:

```bash
$ kubectl get secret -n demo nats-quickstart-auth -o jsonpath='{.data.username}' | base64 -d
admin
$ kubectl get secret -n demo nats-quickstart-auth -o jsonpath='{.data.password}' | base64 -d
```

Port-forward and use the [`nats` CLI](https://github.com/nats-io/natscli):

```bash
$ kubectl port-forward -n demo svc/nats-quickstart 4222
$ nats --server nats://127.0.0.1:4222 --user admin --password "$PASS" server check connection
$ nats --server nats://127.0.0.1:4222 --user admin --password "$PASS" stream add ORDERS --subjects 'orders.*' --replicas 3 --storage file --defaults
$ nats --server nats://127.0.0.1:4222 --user admin --password "$PASS" pub orders.new hello
$ nats --server nats://127.0.0.1:4222 --user admin --password "$PASS" stream info ORDERS
```

Because the stream has `--replicas 3` and the cluster has three servers, every message is stored on all three pods.

## Cleanup

```bash
$ kubectl patch -n demo nats/nats-quickstart -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete -n demo nats/nats-quickstart
$ kubectl delete ns demo
```

## Next Steps

- Learn about [clustering and configuration](/docs/guides/nats/clustering.md).
- Secure it with [TLS](/docs/guides/nats/tls.md).
