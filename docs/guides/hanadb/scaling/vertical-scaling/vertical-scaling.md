---
title: Vertical Scaling HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-scaling-vertical-scaling
    name: Vertical Scaling
    parent: guides-hanadb-scaling-vertical
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Vertical Scale HanaDB

This guide shows how to change the CPU and memory of a HanaDB database using a `HanaDBOpsRequest` of type
`VerticalScaling`. KubeDB updates the PetSet resources and performs a rolling restart (primary last for a
cluster).

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/scaling](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/scaling) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- Deploy a `hanadb-cluster` System Replication database (see [Restart](/docs/guides/hanadb/restart/restart.md#deploy-a-hanadb-system-replication-cluster)) in namespace `demo`.

## Check Resources Before Scaling

```bash
$ kubectl get pod -n demo hanadb-cluster-0 -o json | jq '.spec.containers[] | select(.name=="hanadb") | .resources'
{
  "limits": {
    "cpu": "4",
    "memory": "14Gi"
  },
  "requests": {
    "cpu": "1500m",
    "memory": "8Gi"
  }
}
```

## Create a VerticalScaling HanaDBOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-vscale
  namespace: demo
spec:
  type: VerticalScaling
  databaseRef:
    name: hanadb-cluster
  verticalScaling:
    hanadb:
      resources:
        requests:
          cpu: "2100m"
          memory: "8448Mi"
        limits:
          cpu: "4"
          memory: "14Gi"
  timeout: 30m
  apply: IfReady
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/scaling/system-replication-vertical-scaling.yaml
hanadbopsrequest.ops.kubedb.com/hdbops-vscale created
```

Here,

- `spec.verticalScaling.hanadb.resources` are the desired resources for the `hanadb` container. You can
  also scale the `coordinator` and `exporter` containers via `spec.verticalScaling.coordinator` and
  `spec.verticalScaling.exporter`.

## Verify the Scaling

```bash
$ kubectl get hdbops -n demo hdbops-vscale
NAME            TYPE              STATUS       AGE
hdbops-vscale   VerticalScaling   Successful   4m22s
```

```bash
$ kubectl describe hdbops -n demo hdbops-vscale
...
Status:
  Conditions:
    Message:  HanaDB ops-request has started to vertically scaling the HanaDB nodes
    Reason:   VerticalScaling
    Status:   True
    Type:     VerticalScaling
    Message:  Successfully updated PetSets Resources
    Reason:   UpdatePetSets
    Status:   True
    Type:     UpdatePetSets
    Message:  Successfully Restarted Pods With Resources
    Reason:   RestartPods
    Status:   True
    Type:     RestartPods
    Message:  Successfully completed the vertical scaling for HanaDB
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful
```

Confirm the new resources are in effect on the PetSet and pods:

```bash
$ kubectl get petset -n demo hanadb-cluster \
  -o jsonpath='{range .spec.template.spec.containers[?(@.name=="hanadb")]}{.name}{": "}{.resources}{"\n"}{end}'
hanadb: {"limits":{"cpu":"4","memory":"14Gi"},"requests":{"cpu":"2100m","memory":"8448Mi"}}

$ kubectl get pod -n demo hanadb-cluster-0 -o json | jq '.spec.containers[] | select(.name=="hanadb") | .resources'
{
  "limits": {
    "cpu": "4",
    "memory": "14Gi"
  },
  "requests": {
    "cpu": "2100m",
    "memory": "8448Mi"
  }
}
```

## Cleaning Up

```bash
$ kubectl delete hdbops -n demo hdbops-vscale
$ kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
$ kubectl delete ns demo
```

## Next Steps

- [Expand the volume](/docs/guides/hanadb/volume-expansion/volume-expansion.md) of a HanaDB.
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
