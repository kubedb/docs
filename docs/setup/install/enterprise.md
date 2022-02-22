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

KubeDB Enterprise edition is the open core version of [KubeDB](https://github.com/kubedb/operator). It includes all the features (clustering, etc.) of KubeDB Community Edition and extends it by automating Day 2 operations, improving security and productivity. `Enterprise Edition` can be used to manage KubeDB custom resources in any Kubernetes namespace.

- Back and recovery - KubeDB will provide backup & recovery of databases using Stash.
- Upgrade and Scaling - KubeDB will provide operator managed human-in-the-loop patch and minor upgrade, downgrade and scaling operations
- SSL Support - KubeDB Enterprise operator supports SSL certificate management for supported database types via Jetstack’s [cert-manager](https://cert-manager.io/) project.
- User Management with HashiCorp Vault - KubeDB supports user management using HashiCorp Vault via [KubeVault](https://kubevault.com/) project.
- Web Dashboard - KubeDB Enterprise offers a web based management console with Prometheus and Grafana integration for monitoring.
- Connection Pooling - KubeDB Enterprise edition offers PgBouncer support for PostgreSQL and ProxySQL support for MySQL and Percona XtraDB.

A full features comparison between KubeDB Enterprise Edition and community version can be found [here](/docs/overview/README.md).

If you are willing to try KubeDB Enterprise Edition, you can grab a **30 days trial** license from [here](https://license-issuer.appscode.com/?p=kubedb-enterprise).

## Get a Trial License

In this section, we are going to show you how you can get a **30 days trial** license for KubeDB Enterprise edition. You can get a license for your Kubernetes cluster by going through the following steps:

- At first, go to [AppsCode License Server](https://license-issuer.appscode.com/?p=kubedb-enterprise) and fill up the form. It will ask for your Name, Email, the product you want to install, and your cluster ID (UID of the `kube-system` namespace).
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

You can create licenses for as many clusters as you want. You can upgrade your license any time without re-installing KubeDB by following the upgrading guide from [here](/docs/setup/upgrade/index.md#updating-license).

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
    <a class="nav-link" id="script-tab" data-toggle="tab" href="#script" role="tab" aria-controls="script" aria-selected="false">YAML</a>
  </li>
</ul>
<div class="tab-content" id="installerTabContent">
  <div class="tab-pane fade show active" id="helm3" role="tabpanel" aria-labelledby="helm3-tab">

## Using Helm 3

KubeDB can be installed via [Helm](https://helm.sh/) using the [chart](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb) from [AppsCode Charts Repository](https://github.com/appscode/charts). To install, follow the steps below:

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm search repo appscode/kubedb
NAME                        CHART VERSION APP VERSION DESCRIPTION
appscode/kubedb                     {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler          {{< param "info.autoscaler" >}}       {{< param "info.autoscaler" >}}     KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog             {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-crds                {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB Custom Resource Definitions
appscode/kubedb-dashboard           {{< param "info.dashboard" >}}        {{< param "info.dashboard" >}}      KubeDB Dashboard by AppsCode
appscode/kubedb-grafana-dashboards  {{< param "info.version" >}}   {{< param "info.version" >}} A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics             {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB State Metrics
appscode/kubedb-ops-manager         {{< param "info.enterprise" >}}       {{< param "info.enterprise" >}}     KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter           {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB Opscenter by AppsCode
appscode/kubedb-provisioner         {{< param "info.community" >}}       {{< param "info.community" >}}     KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager      {{< param "info.schema-manager" >}}        {{< param "info.schema-manager" >}}      KubeDB Schema Manager by AppsCode
appscode/kubedb-ui-server           {{< param "info.ui-server" >}}   {{< param "info.ui-server" >}} A Helm chart for kubedb-ui-server by AppsCode
appscode/kubedb-webhook-server      {{< param "info.webhook-server" >}}        {{< param "info.webhook-server" >}}      KubeDB Webhook Server by AppsCode

# Install KubeDB Enterprise edition
$ helm install kubedb appscode/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set kubedb-provisioner.enabled=true \
  --set kubedb-ops-manager.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-dashboard.enabled=true \
  --set kubedb-schema-manager.enabled=true \
  --set-file global.license=/path/to/the/license.txt
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

</div>
<div class="tab-pane fade" id="script" role="tabpanel" aria-labelledby="script-tab">

## Using YAML

If you prefer to not use Helm, you can generate YAMLs from KubeDB chart and deploy using `kubectl`. Here we are going to show the procedure using Helm 3.

```bash
$ helm repo add appscode https://charts.appscode.com/stable/
$ helm repo update

$ helm search repo appscode/kubedb
NAME                        CHART VERSION APP VERSION DESCRIPTION
appscode/kubedb                     {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB by AppsCode - Production ready databases...
appscode/kubedb-autoscaler          {{< param "info.autoscaler" >}}       {{< param "info.autoscaler" >}}     KubeDB Autoscaler by AppsCode - Autoscale KubeD...
appscode/kubedb-catalog             {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB Catalog by AppsCode - Catalog for databa...
appscode/kubedb-crds                {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB Custom Resource Definitions
appscode/kubedb-dashboard           {{< param "info.dashboard" >}}        {{< param "info.dashboard" >}}      KubeDB Dashboard by AppsCode
appscode/kubedb-grafana-dashboards  {{< param "info.version" >}}   {{< param "info.version" >}} A Helm chart for kubedb-grafana-dashboards by A...
appscode/kubedb-metrics             {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB State Metrics
appscode/kubedb-ops-manager         {{< param "info.enterprise" >}}       {{< param "info.enterprise" >}}     KubeDB Ops Manager by AppsCode - Enterprise fea...
appscode/kubedb-opscenter           {{< param "info.version" >}}   {{< param "info.version" >}} KubeDB Opscenter by AppsCode
appscode/kubedb-provisioner         {{< param "info.community" >}}       {{< param "info.community" >}}     KubeDB Provisioner by AppsCode - Community feat...
appscode/kubedb-schema-manager      {{< param "info.schema-manager" >}}        {{< param "info.schema-manager" >}}      KubeDB Schema Manager by AppsCode
appscode/kubedb-ui-server           {{< param "info.ui-server" >}}   {{< param "info.ui-server" >}} A Helm chart for kubedb-ui-server by AppsCode
appscode/kubedb-webhook-server      {{< param "info.webhook-server" >}}        {{< param "info.webhook-server" >}}      KubeDB Webhook Server by AppsCode

# Install KubeDB Enterprise operator chart
$ helm template kubedb appscode/kubedb \
  --version {{< param "info.version" >}} \
  --namespace kubedb --create-namespace \
  --set-file global.license=/path/to/the/license.txt  \
  --set kubedb-provisioner.enabled=true \
  --set kubedb-ops-manager.enabled=true \
  --set kubedb-autoscaler.enabled=true \
  --set kubedb-dashboard.enabled=true \
  --set kubedb-schema-manager.enabled=true \
  --set global.skipCleaner=true | kubectl apply -f -
```

To see the detailed configuration options, visit [here](https://github.com/kubedb/installer/tree/{{< param "info.installer" >}}/charts/kubedb).

</div>
</div>

## Verify installation

To check if KubeDB operator pods have started, run the following command:

```bash
$ watch kubectl get pods --all-namespaces -l "app.kubernetes.io/instance=kubedb"

NAMESPACE   NAME                                        READY   STATUS    RESTARTS   AGE
kubedb      kubedb-kubedb-autoscaler-5845d7b8ff-ts82g   1/1     Running   0          56s
kubedb      kubedb-kubedb-community-595c96d499-srl7c    1/1     Running   0          56s
kubedb      kubedb-kubedb-enterprise-9cf54ccd-9hz96     1/1     Running   0          56s
```

Once the operator pod is running, you can cancel the above command by typing `Ctrl+C`.

Now, to confirm CRD groups have been registered by the operator, run the following command:

```bash
$ kubectl get crd -l app.kubernetes.io/name=kubedb
```

Now, you are ready to [create your first database](/docs/guides/README.md) using KubeDB.
