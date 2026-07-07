---
title: Initialize Oracle using Script Source
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-initialization-script-source
    name: Using Script
    parent: guides-oracle-initialization
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Initialize Oracle with Script

KubeDB supports initializing an Oracle database with a user provided SQL script. When the database boots for the **first time**, KubeDB runs your script so you can pre-create schemas, tables, and seed data. This tutorial shows how.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/initialization](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/initialization) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Create an Oracle Container Registry token, if you haven't created one already, by following the instructions in the guide below: [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important)

## Prepare Initialization Script

The initialization script can be supplied through a `ConfigMap` or a `Secret`. The script file **must** be named `setup.sql`.

Below is a `ConfigMap` that creates a table named `emp`, inserts a row, and selects it back:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oracle-init-script
  namespace: demo
data:
  setup.sql: |
    CREATE TABLE emp (
                         id   INT,
                         name CHAR(10)
    );

    INSERT INTO emp VALUES (1, 'John');

    SELECT * FROM emp;
```

Let's create the `ConfigMap`,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/initialization/oracle-init-script-config-map.yaml
```
configmap/oracle-init-script created

> Note: the key inside the ConfigMap (the file name) must be `setup.sql`.

## Create Oracle with script source

Now, create an `Oracle` CR that references the ConfigMap through `spec.init.script.configMap`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: init-config
  namespace: demo
spec:
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  storageType: Durable
  replicas: 1
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  init:
    script:
      configMap:
        name: oracle-init-script
  deletionPolicy: WipeOut
```

Here,

- `spec.init.script.configMap.name` refers to the `ConfigMap` (`oracle-init-script`) holding the `setup.sql` script. You can use `spec.init.script.secret` to source the script from a `Secret` instead.

Let's create the `Oracle` CR,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/initialization/oracle-init-script.yaml
```
oracle.kubedb.com/init-config created

Now, wait until `init-config` has status `Ready` and the pod prints the `DATABASE IS READY TO USE!!!` banner. The initialization script runs once during this first boot.

> Note: Initialization scripts run **only on the first boot** of a fresh database. They do not re-run when an existing database is restarted.

## Verify Initialization

Once the database is ready, let's connect to it and confirm the `emp` table created by our script exists and contains the seeded row,

```bash
kubectl get secret -n demo init-config-auth -o jsonpath='{.data.password}' | base64 -d
```
# (use the printed password below)

```bash
kubectl exec -n demo init-config-0 -c oracle -- bash -lc \
    "echo -e 'SELECT * FROM emp;\nexit;' | sqlplus -s sys/<password>@localhost:1521/ORCL as sysdba"
```
        ID NAME
---------- ----------
         1 John

The `emp` table and its row are present, confirming the initialization script ran successfully.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl patch -n demo oracle/init-config -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo init-config
kubectl delete configmap -n demo oracle-init-script
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Run Oracle with a [custom configuration file](/docs/guides/oracle/configuration/using-config-file.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

> ## ⚠️ Legal Notice
>
> Oracle® and Oracle Database® are registered trademarks of Oracle Corporation.
> KubeDB is not affiliated with, endorsed by, or sponsored by Oracle Corporation.
>
> KubeDB provides only orchestration and management tooling for Kubernetes.
> It does not distribute, bundle, ship, or include any Oracle Database software or binaries.
>
> Users must provide their own Oracle container images and hold valid Oracle licenses.
> Users are solely responsible for compliance with Oracle’s licensing terms, including all rules regarding containers, Docker, and Kubernetes environments.
>
> KubeDB makes no representations or warranties regarding Oracle licensing compliance.
