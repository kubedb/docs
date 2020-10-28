---
title: Install KubeDB Enterprise Edition
description: Installation guide for KubeDB Enterprise edition
menu:
  docs_{{ .version }}:
    identifier: install-kubedb-enterprise
    name: Enterprise Edition
    parent: installation-guide
    weight: 20
product_name: kubedb
menu_name: docs_{{ .version }}
section_menu_id: setup
---

# Install KubeDB Enterprise Edition

KubeDB Enterprise edition is the open core version of [KubeDB](https://github.com/kubedb/operator). It includes all the features (clustering, etc.) of KubeDB Community Edition and extends it by automating Day 2 operations, improving security and productivity.

- Back and recovery - KubeDB will provide backup & recovery of databases using Stash.
- Upgrade and Scaling - KubeDB will provide operator managed human-in-the-loop patch and minor upgrade, downgrade and scaling operations
- SSL Support - KubeDB Enterprise operator supports SSL certificate management for supported database types via Jetstack’s [cert-manager](https://cert-manager.io/) project.
- User Management with HashiCorp Vault - KubeDB supports user management using HashiCorp Vault via [KubeVault](https://kubevault.com/) project.
- Web Dashboard - KubeDB Enterprise offers a web based management console with Prometheus and Grafana integration for monitoring.
- Connection Pooling - KubeDB Enterprise edition offers PgBouncer support for PostgreSQL and ProxySQL support for MySQL and Percona XtraDB.

A full features comparison between KubeDB Enterprise Edition and community version can be found [here](/docs/overview/README.md).

If you are willing to try KubeDB Enterprise Edition, you can grab a **14 days trial** license from [here](https://license-issuer.appscode.com/).

## Get a Trial License

In this section, we are going to show you how you can get a **14 days trial** license for KubeDB Enterprise edition. You can get a license for your Kubernetes cluster by going through the following steps:

- At first, go to [AppsCode License Server](https://license-issuer.appscode.com/) and fill up the form. It will ask for your Name, Email, the product you want to install, and your cluster ID (UID of the `kube-system` namespace).
- Provide your name and email address. **You must provide your work email address**.
- Then, select `KubeDB Enterprise Edition` in the product field.
- Now, provide your cluster ID. You can get your cluster ID easily by running the following command:

```bash
kubectl get ns kube-system -o=jsonpath='{.metadata.uid}'
```

- Then, you have to agree with the terms and conditions. We recommend reading it before checking the box.
- Now, you can submit the form. After you submit the form, the AppsCode License server will send an email to the provided email address with a link to your license file.
- Navigate to the provided link and save the license into a file. Here, we save the license to a `license.txt` file.

Here is a screenshot of the license form.

<figure align="center">
  <img alt="KubeDB Backend Overview" src="/docs/images/setup/enterprise_license_form.png">
  <figcaption align="center">Fig: KubeDB License Form</figcaption>
</figure>

You can create licenses for as many clusters as you want. You can upgrade your license any time without re-installing KubeDB by following the upgrading guide from [here](/docs/setup/upgrade.md#upgrading-license).

> KubeDB licensing process has been designed to work with CI/CD workflow. You can automatically obtain a license from your CI/CD pipeline by following the guide from [here](https://github.com/appscode/offline-license-server#offline-license-server).

## Get an Enterprise License

If you are interested in purchasing Enterprise license, please contact us via sales@appscode.com for further discussion. You can also set up a meeting via our [calendly link](https://calendly.com/appscode/30min).

If you are willing to purchasing Enterprise license but need more time to test in your dev cluster, feel free to contact sales@appscode.com. We will be happy to extend your trial period.

## Install

To activate the Enterprise features, you need to install both KubeDB Community operator and Enterprise operator chart. These operators can be installed as a Helm chart or simply as Kubernetes manifests. If you have already installed the Community operator, only install the Enterprise operator (step 4 in the following secttion).

<ul class="nav nav-tabs" id="installerTab" role="tablist">
  <li class="nav-item">
    <a class="nav-link active" id="helm3-tab" data-toggle="tab" href="#helm3" role="tab" aria-controls="helm3" aria-selected="true">Helm 3 (Recommended)</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="helm2-tab" data-toggle="tab" href="#helm2" role="tab" aria-controls="helm2" aria-selected="false">Helm 2</a>
  </li>
  <li class="nav-item">
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

KubeDB can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-enterprise) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install, follow the steps below:

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/kubedb --version {{< param "info.community" >}}
NAME                        CHART VERSION APP VERSION   DESCRIPTION
appscode/kubedb             {{< param "info.community" >}}  {{< param "info.community" >}}  KubeDB by AppsCode - Production ready databases on Kubernetes
appscode/kubedb-catalog     {{< param "info.community" >}}  {{< param "info.community" >}}  KubeDB Catalog by AppsCode - Catalog for database versions
appscode/kubedb-enterprise  {{< param "info.enterprise" >}}  {{< param "info.enterprise" >}}  KubeDB Enterprise by AppsCode - Enterprise features for KubeDB

# Step 1: Install KubeDB Community operator chart
$ helm install kubedb-community kubedb appscode/kubedb \
  --version {{< param "info.community" >}}      \
  --namespace kube-system                       \
  --set-file license=/path/to/the/license.txt

# Step 2: wait until crds are registered
$ kubectl get crds -l app=kubedb -w
NAME                               AGE
elasticsearches.kubedb.com         12s
elasticsearchversions.kubedb.com   8s
etcds.kubedb.com                   8s
etcdversions.kubedb.com            8s
memcacheds.kubedb.com              6s
memcachedversions.kubedb.com       6s
mongodbs.kubedb.com                7s
mongodbversions.kubedb.com         6s
mysqls.kubedb.com                  7s
mysqlversions.kubedb.com           7s
postgreses.kubedb.com              8s
postgresversions.kubedb.com        7s
redises.kubedb.com                 6s
redisversions.kubedb.com           6s

# Step 3(a): Install KubeDB catalog of database versions
$ helm install kubedb-catalog appscode/kubedb-catalog   \
  --version {{< param "info.community" >}}              \
  --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade KubeDB catalog of database versions
$ helm upgrade kubedb-catalog appscode/kubedb-catalog   \
  --version {{< param "info.community" >}}              \
  --namespace kube-system

# Step 4: Install KubeDB Enterprise operator chart
$ helm install kubedb-enterprise appscode/kubedb-enterprise  \
  --version {{< param "info.enterprise" >}}                  \
  --namespace kube-system                                    \
  --set-file license=/path/to/the/license.txt
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-enterprise).

</div>
<div class="tab-pane fade" id="helm2" role="tabpanel" aria-labelledby="helm2-tab">

## Using Helm 2

KubeDB can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-enterprise) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install the chart with the release name `kubedb`:

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search appscode/kubedb --version {{< param "info.community" >}}
NAME                        CHART VERSION APP VERSION   DESCRIPTION
appscode/kubedb             {{< param "info.community" >}}  {{< param "info.community" >}}  KubeDB by AppsCode - Production ready databases on Kubernetes
appscode/kubedb-catalog     {{< param "info.community" >}}  {{< param "info.community" >}}  KubeDB Catalog by AppsCode - Catalog for database versions
appscode/kubedb-enterprise  {{< param "info.enterprise" >}}  {{< param "info.enterprise" >}}  KubeDB Enterprise by AppsCode - Enterprise features for KubeDB

# Step 1: Install KubeDB Community operator chart
$ helm install appscode/kubedb --name kubedb-community \
  --version {{< param "info.community" >}}     \
  --namespace kube-system                      \
  --set-file license=/path/to/the/license.txt

# Step 2: wait until crds are registered
$ kubectl get crds -l app=kubedb -w
NAME                               AGE
elasticsearches.kubedb.com         12s
elasticsearchversions.kubedb.com   8s
etcds.kubedb.com                   8s
etcdversions.kubedb.com            8s
memcacheds.kubedb.com              6s
memcachedversions.kubedb.com       6s
mongodbs.kubedb.com                7s
mongodbversions.kubedb.com         6s
mysqls.kubedb.com                  7s
mysqlversions.kubedb.com           7s
postgreses.kubedb.com              8s
postgresversions.kubedb.com        7s
redises.kubedb.com                 6s
redisversions.kubedb.com           6s

# Step 3(a): Install KubeDB catalog of database versions
$ helm install appscode/kubedb-catalog --name kubedb-catalog   \
  --version {{< param "info.community" >}}                     \
  --namespace kube-system

# Step 3(b): Or, if previously installed, upgrade KubeDB catalog of database versions
$ helm upgrade kubedb-catalog appscode/kubedb-catalog          \
  --version {{< param "info.community" >}}                     \
  --namespace kube-system

# Step 4: Install KubeDB Enterprise operator chart
$ helm install appscode/kubedb-enterprise --name kubedb-enterprise   \
  --version {{< param "info.enterprise" >}}                          \
  --namespace kube-system                                            \
  --set-file license=/path/to/the/license.txt
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-enterprise).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from KubeDB chart and deploy using `kubectl`. Here we are going to show the prodecure using Helm 3.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update
$ helm search repo appscode/kubedb --version {{< param "info.community" >}}
NAME                        CHART VERSION APP VERSION   DESCRIPTION
appscode/kubedb             {{< param "info.community" >}}  {{< param "info.community" >}}  KubeDB by AppsCode - Production ready databases on Kubernetes
appscode/kubedb-catalog     {{< param "info.community" >}}  {{< param "info.community" >}}  KubeDB Catalog by AppsCode - Catalog for database versions
appscode/kubedb-enterprise  {{< param "info.enterprise" >}}  {{< param "info.enterprise" >}}  KubeDB Enterprise by AppsCode - Enterprise features for KubeDB

# Step 1: Install KubeDB Community operator chart
$ helm template kubedb-community appscode/kubedb \
  --version {{< param "info.community" >}}      \
  --namespace kube-system                       \
  --set-file license=/path/to/the/license.txt   \
  --no-hooks | kubectl apply -f -

# Step 2: wait until crds are registered
$ kubectl get crds -l app=kubedb -w
NAME                               AGE
elasticsearches.kubedb.com         12s
elasticsearchversions.kubedb.com   8s
memcacheds.kubedb.com              6s
memcachedversions.kubedb.com       6s
mongodbs.kubedb.com                7s
mongodbversions.kubedb.com         6s
mysqls.kubedb.com                  7s
mysqlversions.kubedb.com           7s
postgreses.kubedb.com              8s
postgresversions.kubedb.com        7s
redises.kubedb.com                 6s
redisversions.kubedb.com           6s

# Step 3: Install/Upgrade KubeDB catalog of database versions
$ helm template kubedb-catalog appscode/kubedb-catalog   \
  --version {{< param "info.community" >}}               \
  --namespace kube-system                                \
  --no-hooks | kubectl apply -f -

# Step 1: Install KubeDB Enterprise operator chart
$ helm template kubedb-enterprise appscode/kubedb-enterprise   \
  --version {{< param "info.enterprise" >}}                    \
  --namespace kube-system                                      \
  --set-file license=/path/to/the/license.txt                  \
  --no-hooks | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb-enterprise).

</div>
</div>

## Verify installation

To check if KubeDB operator pods have started, run the following command:

```bash
$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=kubedb --watch

NAMESPACE     NAME                     READY     STATUS    RESTARTS   AGE
kube-system   kubedb-859d6bdb56-m9br5  1/1       Running   2          5s

$ kubectl get pods --all-namespaces -l app.kubernetes.io/name=kubedb-enterprise --watch

NAMESPACE     NAME                                READY   STATUS    RESTARTS   AGE
kube-system   kubedb-enterprise-568c884795-hzbgg  2/2     Running   0          5h35m
```

Once the operator pod is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
```

Now, you are ready to [create your first database](/docs/guides/README.md) using KubeDB.
