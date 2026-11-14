---
title: Elasticsearch License Guide
description: Elasticsearch License Guide
menu:
  docs_{{ .version }}:
    identifier: es-license-guide
    name: Guide
    parent: es-license-elasticsearch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Activate an Elasticsearch License

This guide shows how to activate an Elastic subscription license (or the built-in trial) on a KubeDB
`Elasticsearch` cluster, both declaratively via `spec.license` and on demand via a `RotateLicense`
`ElasticsearchOpsRequest`.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Elasticsearch](/docs/guides/elasticsearch/concepts/elasticsearch/index.md)
  - [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md)
  - [Elasticsearch License Overview](/docs/guides/elasticsearch/license/overview.md)

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured
  to communicate with your cluster. If you do not already have a cluster, you can create one using
  [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps
  [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout.

  ```bash
  $ kubectl create ns demo
  namespace/demo created
  ```

## Create an Elasticsearch Database

The Elasticsearch instance used for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: sample-es
  namespace: demo
spec:
  version: xpack-9.2.3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Let's create the above `Elasticsearch` object:

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/license/yamls/sample-es.yaml
elasticsearch.kubedb.com/sample-es created
```

Wait until `sample-es` has status `Ready`:

```shell
$ kubectl get elasticsearch -n demo sample-es -w
NAME        VERSION        STATUS   AGE
sample-es   xpack-9.2.3    Ready    3m12s
```

At this point the cluster is running on the free Basic tier -- no `spec.license` has been set yet.

## Option 1: Bring Your Own License

First, create a Secret holding the signed license file you downloaded from Elastic, under the key
`license.json`:

```shell
$ kubectl create secret generic sample-es-license -n demo \
   --from-file=license.json=./license.json
secret/sample-es-license created
```

Now set `spec.license.secretRef` on the `Elasticsearch` object to point at that Secret:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: sample-es
  namespace: demo
spec:
  version: xpack-9.2.3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  license:
    secretRef:
      name: sample-es-license
  deletionPolicy: WipeOut
```

Here, `spec.license.secretRef.name` references the Secret we just created. Apply the change:

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/license/yamls/es-with-license-secret.yaml
elasticsearch.kubedb.com/sample-es configured
```

On its next health-check pass, the operator reads the Secret and calls `PUT _license` on the cluster.
Watch the `LicenseActive` condition turn `True`:

```shell
$ kubectl get elasticsearch -n demo sample-es -o json | jq '.status.conditions[] | select(.type | startswith("License"))'
{
  "type": "LicenseActive",
  "status": "True",
  "reason": "LicenseActive",
  "message": "Elasticsearch license type: \"platinum\", status: \"active\". license expires at 2027-01-15T00:00:00Z",
  "observedGeneration": 2,
  "lastTransitionTime": "2026-09-24T04:12:03Z"
}
```

If the referenced Secret is missing the `license.json` key, or the `Elasticsearch` doesn't use the
`ElasticStack` distribution, the webhook rejects the change up front with a message explaining why.

## Option 2: Activate the Trial License

Instead of a BYO license, you can ask the operator to activate Elastic's built-in one-time 30-day
trial:

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: sample-es
  namespace: demo
spec:
  version: xpack-9.2.3
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  license:
    trial: true
  deletionPolicy: WipeOut
```

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/license/yamls/es-with-license-trial.yaml
elasticsearch.kubedb.com/sample-es configured
```

> Note: Elastic only allows the trial to be started once per cluster. If it was already used (for
> example, activated manually before KubeDB took it over), the operator sets `LicenseSyncFailed` with
> the underlying error instead of retrying every reconcile.

## Rotate a License with `RotateLicense`

Both options above are reconciled continuously, but that means waiting for the next health-check pass.
To apply a new license (or request the trial) immediately, use a `RotateLicense`
`ElasticsearchOpsRequest`:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: esops-rotate-license
  namespace: demo
spec:
  type: RotateLicense
  databaseRef:
    name: sample-es
  license:
    secretRef:
      name: sample-es-license-renewed
  timeout: 5m
  apply: IfReady
```

Here,

- `spec.databaseRef.name` specifies that we are rotating the license on the `sample-es` cluster.
- `spec.type` specifies that we are performing `RotateLicense` on Elasticsearch.
- `spec.license.secretRef.name` references a Secret holding the renewed license, the same
  `license.json`-keyed shape as `spec.license.secretRef` on the `Elasticsearch` object itself.
  Use `spec.license.trial: true` instead to request the trial through the ops request.

Let's create the `ElasticsearchOpsRequest`:

```shell
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/license/yamls/rotate-license-secret.yaml
elasticsearchopsrequest.ops.kubedb.com/esops-rotate-license created
```

Wait for it to become `Successful`:

```shell
$ kubectl get elasticsearchopsrequest -n demo
NAME                    TYPE            STATUS       AGE
esops-rotate-license    RotateLicense   Successful   45s
```

Unlike `RotateAuth` or `Reconfigure`, `RotateLicense` talks to the already-running cluster's REST API
directly -- it does not pause the database or restart any pods. Once successful, the operator patches
`spec.license` on the `Elasticsearch` object to match what was applied, so the CR remains the source of
truth for future reconciles:

```shell
$ kubectl get elasticsearch -n demo sample-es -o json | jq .spec.license
{
  "secretRef": {
    "name": "sample-es-license-renewed"
  }
}
```

## Cleaning Up

To clean up the Kubernetes resources created in this tutorial, run:

```shell
$ kubectl delete elasticsearchopsrequest -n demo esops-rotate-license
elasticsearchopsrequest.ops.kubedb.com "esops-rotate-license" deleted
$ kubectl delete elasticsearch -n demo sample-es
elasticsearch.kubedb.com "sample-es" deleted
$ kubectl delete secret -n demo sample-es-license sample-es-license-renewed
secret "sample-es-license" deleted
secret "sample-es-license-renewed" deleted
```

## Next Steps

- Learn about [Rotate Authentication](/docs/guides/elasticsearch/rotateauth/overview.md) for Elasticsearch.
- Learn about [Reconfiguring](/docs/guides/elasticsearch/reconfigure/overview.md) Elasticsearch settings.
- Detail concepts of the [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Detail concepts of [ElasticsearchOpsRequest](/docs/guides/elasticsearch/concepts/elasticsearch-ops-request/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
