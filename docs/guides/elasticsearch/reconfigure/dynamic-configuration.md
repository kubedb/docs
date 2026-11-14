---
title: Dynamic Elasticsearch Configuration
menu:
  docs_{{ .version }}:
    identifier: es-reconfigure-dynamic
    name: Dynamic Configuration
    parent: es-reconfigure
    weight: 25
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Dynamic Elasticsearch Configuration

This guide applies a dynamic setting through a `Reconfigure` OpsRequest and verifies that the database Pods are not replaced.

## Before You Begin

- Install KubeDB Provisioner and Ops-manager versions that support dynamic reconfiguration, following the [setup guide](/docs/setup/README.md).
- Start with a Ready `Elasticsearch` database named `es-combined` in namespace `demo`. You can use the database from the [combined cluster guide](/docs/guides/elasticsearch/reconfigure/elasticsearch-combined.md).
- Read the [reconfiguration overview](/docs/guides/elasticsearch/reconfigure/overview.md) for restart rules.
- Ensure no other operation is changing the database during this example.

## Record the Current Pods

Save the Pod names and UIDs before applying the request. The selector below is shared by the database Pods managed by KubeDB:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=es-combined \
  -o custom-columns=NAME:.metadata.name,UID:.metadata.uid --no-headers | sort > /tmp/elasticsearch-pods-before.txt
cat /tmp/elasticsearch-pods-before.txt
```

## Choose a Cluster-Wide Setting

This example sets `cluster.max_shards_per_node` to `2000`. Use a different value if it is already `2000` in the effective user configuration. Ensure this key is not also defined in a role-specific file such as `data-elasticsearch.yml`.

For OpenSearch, use the same `ElasticsearchOpsRequest` kind and replace the `elasticsearch.yml` key below with `opensearch.yml`. The live cluster settings API applies changes across the cluster, including topology deployments.

## Apply the Dynamic Configuration

Save the following manifest as `es-reconfigure-dynamic.yaml`. It is also available in [the examples directory](/docs/examples/elasticsearch/reconfigure/es-reconfigure-dynamic.yaml).

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: elasticsearch-reconfigure-dynamic
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: es-combined
  configuration:
    restart: "false"
    applyConfig:
      elasticsearch.yml: |
        cluster.max_shards_per_node: 2000
  timeout: 5m
  apply: IfReady
```

`restart: "false"` makes this request fail if the change requires a restart. Omit it or use `restart: "auto"` to allow the operator to restart Pods when a change cannot be applied live. Use `restart: "true"` to request a restart explicitly.

```bash
kubectl apply -f es-reconfigure-dynamic.yaml
kubectl wait -n demo elasticsearchopsrequest/elasticsearch-reconfigure-dynamic \
  --for=jsonpath='{.status.phase}'=Successful --timeout=600s
```

If the wait times out, inspect the request before proceeding:

```bash
kubectl describe elasticsearchopsrequest -n demo elasticsearch-reconfigure-dynamic
```

## Verify Live Application

Check the mode selected by the operator:

```bash
kubectl get elasticsearchopsrequest -n demo elasticsearch-reconfigure-dynamic \
  -o jsonpath='{range .status.conditions[?(@.type=="ReconfigureMode")]}{.reason}{"\n"}{.message}{"\n"}{end}'
```

For a successful dynamic update, the reason is `Live`. A no-change request also uses `Live`, with a message explaining that no setting changed.

Query the persistent cluster settings. The following command assumes the TLS-enabled X-Pack database and default auth Secret from the combined cluster guide:

```bash
ES_USER=$(kubectl get secret -n demo es-combined-auth -o jsonpath='{.data.username}' | base64 -d)
ES_PASS=$(kubectl get secret -n demo es-combined-auth -o jsonpath='{.data.password}' | base64 -d)
kubectl exec -n demo es-combined-0 -c elasticsearch -- \
  curl -ksS --user "$ES_USER:$ES_PASS" \
  'https://localhost:9200/_cluster/settings?flat_settings=true&filter_path=persistent.cluster.max_shards_per_node&pretty'
```

Confirm that the persistent value is `2000`. Use your configured credentials, protocol, and container name for a different deployment. Live updates are stored in persistent cluster settings; `/_nodes/settings` reports settings loaded at startup and does not verify these API overrides.

The operator also stores the desired configuration for future Pods. If a later reconfiguration restarts the nodes, it clears persistent overrides for settings present in the old or new user configuration after the new files have been loaded.

Compare the Pod UIDs with the saved values:

```bash
kubectl get pods -n demo -l app.kubernetes.io/instance=es-combined \
  -o custom-columns=NAME:.metadata.name,UID:.metadata.uid --no-headers | sort > /tmp/elasticsearch-pods-after.txt
diff -u /tmp/elasticsearch-pods-before.txt /tmp/elasticsearch-pods-after.txt
```

An empty diff confirms that the Pods were not replaced during the operation. You can also inspect container restart counts with `kubectl get pods -n demo`.

## Clean Up

```bash
kubectl delete elasticsearchopsrequest -n demo elasticsearch-reconfigure-dynamic
```

Deleting the OpsRequest does not undo the configuration change. To remove a setting, follow the [setting-removal rules](/docs/guides/elasticsearch/reconfigure/overview.md#removing-settings); removal may require a restart.
