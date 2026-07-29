---
title: Reconfigure Oracle
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-reconfigure-details
    name: Reconfigure
    parent: guides-oracle-reconfigure
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfigure Oracle Database

This guide will show you how to use `KubeDB` Ops-manager operator to reconfigure an Oracle database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Install `KubeDB` Provisioner and Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/reconfigure](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/reconfigure) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. Create an Oracle Container Registry token, if you haven't created one already, by following the instructions in the guide below: [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important)

## Deploy Oracle

In this section, we are going to deploy an Oracle standalone database. Below is the YAML of the `Oracle` CR that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: oracle-sa-sample
  namespace: demo
spec:
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
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
        storage: 10Gi
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR we have shown above and wait until it becomes `Ready`,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/reconfigure/standalone-minimal.yaml
oracle.kubedb.com/oracle-sa-sample created
```

## Reconfigure using a config Secret

Now we will reconfigure this database to set `PROCESSES = 800`.

At first, create a Secret with the new configuration. The Secret key must be `oracle.cnf`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: oracle-custom
  namespace: demo
type: Opaque
stringData:
  oracle.cnf: |
    PROCESSES = 800
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/reconfigure/oracle-custom-config-secret.yaml
secret/oracle-custom created
```

### Create OracleOpsRequest

Now, we will use this Secret to reconfigure the database via an `OracleOpsRequest`. Below is the YAML of the `OracleOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: standalone-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: oracle-sa-sample
  configuration:
    configSecret:
      name: oracle-custom
```

Here,

- `spec.type` specifies that we are performing a `Reconfigure` operation.
- `spec.databaseRef.name` specifies that we are reconfiguring the `oracle-sa-sample` database.
- `spec.configuration.configSecret.name` points to the Secret (with key `oracle.cnf`) that holds the new configuration. You can alternatively set `spec.configuration.applyConfig` to provide the configuration inline, or `spec.configuration.removeCustomConfig: true` to drop a previously applied custom configuration.

Let's create the `OracleOpsRequest`,

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/reconfigure/standalone-reconfigure.yaml
oracleopsrequest.ops.kubedb.com/standalone-reconfigure created
```

### Verify the new configuration

Let's wait for the `OracleOpsRequest` to become `Successful`,

```bash
$ kubectl get oracleopsrequest -n demo standalone-reconfigure
NAME                     TYPE          STATUS       AGE
standalone-reconfigure   Reconfigure   Successful   118s
```

We can see the full progress in the `kubectl describe` output,

```bash
$ kubectl describe oracleopsrequest -n demo standalone-reconfigure
Name:         standalone-reconfigure
Namespace:    demo
...
Status:
  Conditions:
    Last Transition Time:  2026-06-22T18:56:21Z
    Message:               Oracle ops-request has started to reconfigure Oracle nodes
    Reason:                Reconfigure
    Status:                True
    Type:                  Reconfigure
    Last Transition Time:  2026-06-22T18:56:37Z
    Message:               successfully reconciled the Oracle with new configuration
    Reason:                UpdatePetSets
    Status:                True
    Type:                  UpdatePetSets
    Last Transition Time:  2026-06-22T18:56:38Z
    Message:               successfully setup wallet from secret
    Reason:                WalletSetup
    Status:                True
    Type:                  WalletSetup
    Last Transition Time:  2026-06-22T18:57:29Z
    Message:               Successfully Restarted Oracle nodes
    Reason:                RestartNodes
    Status:                True
    Type:                  RestartNodes
    Last Transition Time:  2026-06-22T18:57:24Z
    Message:               Pod oracle-sa-sample-0 restarted and healthy
    Status:                True
    Type:                  RestartedPod--oracle-sa-sample-0
    Last Transition Time:  2026-06-22T18:57:29Z
    Message:               Successfully completed reconfiguring for Oracle
    Reason:                Successful
    Status:                True
    Type:                  Successful
  Phase:                   Successful
```

Finally, let's connect to the database and confirm that `PROCESSES` is now `800`,

```bash
$ kubectl get secret -n demo oracle-sa-sample-auth -o jsonpath='{.data.password}' | base64 -d
# (use the printed password below)

$ kubectl exec -n demo oracle-sa-sample-0 -c oracle -- bash -lc \
    "echo -e 'SHOW PARAMETER processes;\nexit;' | sqlplus -s sys/<password>@localhost:1521/ORCL as sysdba"

NAME                                 TYPE        VALUE
------------------------------------ ----------- -----
processes                            integer     800
```

The `processes` parameter has been updated to `800`, confirming the reconfiguration was applied successfully.

## Reconfiguring a DataGuard cluster

The same `OracleOpsRequest` works for a DataGuard cluster — point `spec.databaseRef.name` at the DataGuard database:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: OracleOpsRequest
metadata:
  name: dataguard-reconfigure
  namespace: demo
spec:
  type: Reconfigure
  databaseRef:
    name: oracle-dg-sample
  configuration:
    configSecret:
      name: oracle-custom
```

The new configuration is merged, the PetSets are reconciled to mount it, and the operator performs a rolling restart of all DataGuard pods so each re-applies the `ALTER SYSTEM` settings.

## Cleaning up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracleopsrequest -n demo standalone-reconfigure
kubectl delete secret -n demo oracle-custom
kubectl patch -n demo oracle/oracle-sa-sample -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
kubectl delete oracle -n demo oracle-sa-sample
kubectl delete ns demo
```

## Next Steps

- Detail concepts of [Oracle object](/docs/guides/oracle/concepts/oracle.md).
- Learn how to provision a database with a [custom configuration file](/docs/guides/oracle/configuration/using-config-file.md) at creation time.
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
