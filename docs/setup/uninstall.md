---
title: KubeDB Uninstall
menu:
  docs_{{ .version }}:
    identifier: uninstall-kubedb
    name: Uninstall
    parent: setup
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: setup
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).

# Uninstall KubeDB

To uninstall KubeDB operator, run the following command:

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm2-tab" data-toggle="tab" href="#helm2" role="tab" aria-controls="helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

In Helm 3, release names are [scoped to a namespace](https://v3.helm.sh/docs/faq/#release-names-are-now-scoped-to-the-namespace). So, provide the namespace you used to install the operator when installing.

```console
$ helm uninstall kubedb-operator --namespace kube-system
```

</div>
<div class="tab-pane fade" id="helm2" role="tabpanel" aria-labelledby="helm2-tab">

## Using Helm 2

```console
$ helm delete kubedb-operator
```

</div>
<div class="tab-pane fade show active" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML (with helm 3)

If you prefer to not use Helm, you can generate YAMLs from KubeDB operator chart and uninstall using `kubectl`.

```console
$ helm template kubedb-operator appscode/kubedb --namespace kube-system | kubectl delete -f -
```

</div>
</div>

## Purging KubeDB Custom Resources

The above command will leave the KubeDB crd objects as-is. Follow the setps below to keep a copy of the custom resources in your current directory and delete the CRDs.

- Now, wait several seconds for KubeDB to stop running. To confirm that KubeDB operator pod(s) have stopped running, run:

    ```console
    $ kubectl get pods --all-namespaces -l app=kubedb
    ```

- To keep a copy of your existing KubeDB objects, run:

    ```console
    kubectl get postgres.kubedb.com --all-namespaces -o yaml > postgres.yaml
    kubectl get elasticsearch.kubedb.com --all-namespaces -o yaml > elasticsearch.yaml
    kubectl get memcached.kubedb.com --all-namespaces -o yaml > memcached.yaml
    kubectl get mongodb.kubedb.com --all-namespaces -o yaml > mongodb.yaml
    kubectl get mysql.kubedb.com --all-namespaces -o yaml > mysql.yaml
    kubectl get redis.kubedb.com --all-namespaces -o yaml > redis.yaml
    kubectl get snapshot.kubedb.com --all-namespaces -o yaml > snapshot.yaml
    kubectl get dormant-database.kubedb.com --all-namespaces -o yaml > data.yaml
    ```

- To delete existing KubeDB objects from all namespaces, run the following command in each namespace one by one.

    ```console
    kubectl delete postgres.kubedb.com --all --cascade=false
    kubectl delete elasticsearch.kubedb.com --all --cascade=false
    kubectl delete memcached.kubedb.com --all --cascade=false
    kubectl delete mongodb.kubedb.com --all --cascade=false
    kubectl delete mysql.kubedb.com --all --cascade=false
    kubectl delete redis.kubedb.com --all --cascade=false
    kubectl delete snapshot.kubedb.com --all --cascade=false
    kubectl delete dormant-database.kubedb.com --all --cascade=false
    ```

- Delete the old CRD-registration.

    ```console
    kubectl delete crd -l app=kubedb
    ```
